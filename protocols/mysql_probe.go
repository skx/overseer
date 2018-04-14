package protocols

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"database/sql"

	"github.com/go-sql-driver/mysql"
)

//
// Our structure.
//
// We store state in the `input` field.
//
type MYSQLTest struct {
	input   string
	options TestOptions
}

//
// Run the test against the specified target.
//
func (s *MYSQLTest) RunTest(target string) error {
	var err error

	//
	// Parse the options so we can find username, password, port, etc.
	//
	options := ParseArguments(s.input)

	//
	// The password might be blank, but the username is required.
	//
	if options["username"] == "" {
		return errors.New("No username specified.")
	}

	//
	// The default port to connect to.
	//
	port := 3306

	if options["port"] != "" {
		port, err = strconv.Atoi(options["port"])
		if err != nil {
			return err
		}
	}

	//
	// Create a default configuration structure for MySQL.
	//
	config := mysql.NewConfig()

	//
	// Setup the connection timeout
	//
	config.Timeout = s.options.Timeout

	//
	// Populate the username & password fields.
	//
	config.User = options["username"]
	config.Passwd = options["password"]

	//
	// Default to connecting to an IPv4-address
	//
	address := fmt.Sprintf("%s:%d", target, port)

	//
	// If we find a ":" we know it is an IPv6 address though
	//
	if strings.Contains(target, ":") {
		address = fmt.Sprintf("[%s]:%d", target, port)
	}

	//
	// Setup the address in the configuration structure
	//
	config.Net = "tcp"
	config.Addr = address

	//
	// Now convert the connection-string to a DSN, which
	// is used to connect to the database.
	//
	dsn := config.FormatDSN()

	//
	// Show the DSN, if appropriate.
	//
	if s.options.Verbose {
		fmt.Printf("\tMySQL DSN is %s\n", dsn)
	}

	//
	// Connect to the database
	//
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return err
	}
	defer db.Close()

	//
	// And test that the connection actually worked.
	//
	err = db.Ping()
	return err
}

//
// Store the complete line from the parser in our private
// field; this could be used if there are protocol-specific options
// to be understood.
//
func (s *MYSQLTest) SetLine(input string) {
	s.input = input
}

//
// Store the options for this test
//
func (s *MYSQLTest) SetOptions(opts TestOptions) {
	s.options = opts
}

//
// Register our protocol-tester.
//
func init() {
	Register("mysql", func() ProtocolTest {
		return &MYSQLTest{}
	})
}
