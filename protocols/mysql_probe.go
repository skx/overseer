// MySQL Tester
//
// The MySQL tester connects to a remote MySQL database  and ensures that this
// succeeds. This test is invoked via input like so:
//
//    host.example.com must run mysql with username 'root' with password 'test'
//
// Specifying a username and password is mandatory, because otherwise we
// cannot connect to the database.
//

package protocols

import (
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/go-sql-driver/mysql"
	"github.com/skx/overseer/test"
)

// MYSQLTest is our object
type MYSQLTest struct {
}

// Arguments returns the names of arguments which this protocol-test
// understands, along with corresponding regular-expressions to validate
// their values.
func (s *MYSQLTest) Arguments() map[string]string {
	known := map[string]string{
		"port":     "^[0-9]+$",
		"username": ".*",
		"password": ".*",
	}
	return known
}

// Example returns sample usage-instructions for self-documentation purposes.
func (s *MYSQLTest) Example() string {
	str := `
MySQL Tester
------------
 The MySQL tester connects to a remote MySQL database  and ensures that this
 succeeds. This test is invoked via input like so:

    host.example.com must run mysql [with username 'root' with password 'test'

 Specifying a username and password is mandatory, because otherwise we
 cannot connect to the database.
`
	return str
}

// RunTest is the part of our API which is invoked to actually execute a
// test against the given target.
//
// In this case we make a TCP connection to the host and attempt to login
// with the specified username & password.
func (s *MYSQLTest) RunTest(tst test.Test, target string, opts test.TestOptions) error {
	var err error

	//
	// The password might be blank, but the username is required.
	//
	if tst.Arguments["username"] == "" {
		return errors.New("No username specified.")
	}

	//
	// The default port to connect to.
	//
	port := 3306

	if tst.Arguments["port"] != "" {
		port, err = strconv.Atoi(tst.Arguments["port"])
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
	config.Timeout = opts.Timeout

	//
	// Populate the username & password fields.
	//
	config.User = tst.Arguments["username"]
	config.Passwd = tst.Arguments["password"]

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
	if opts.Verbose {
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
// Register our protocol-tester.
//
func init() {
	Register("mysql", func() ProtocolTest {
		return &MYSQLTest{}
	})
}
