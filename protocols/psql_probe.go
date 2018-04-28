// PSQL Tester
//
// The PSQL tester connects to a remote database and ensures that this
// succeeds.
//
// This test is invoked via input like so:
//
//    host.example.com must run psql with username 'postgres' with password 'mysecretpassword' [with port 5432] [with tls disable]
//
// The `tls` setting may be used to configure how TLS is used, valid values
// are "disable", "require", "verify-ca", or "verify-full".
//
// Specifying a username and password is required, because otherwise we
// cannot connect to the database.
//

package protocols

import (
	"database/sql"
	"errors"
	"fmt"
	"strconv"

	_ "github.com/lib/pq"
	"github.com/skx/overseer/test"
)

// PSQLTest is our object
type PSQLTest struct {
}

// Arguments returns the names of arguments which this protocol-test
// understands, along with corresponding regular-expressions to validate
// their values.
func (s *PSQLTest) Arguments() map[string]string {
	known := map[string]string{
		"port":     "^[0-9]+$",
		"username": ".*",
		"password": ".*",
		"tls":      "^(disable|require|verify-ca|verify-full)$",
	}
	return known
}

// RunTest is the part of our API which is invoked to actually execute a
// test against the given target.
//
// In this case we make a TCP connection to the database host and attempt
// to login with the specified username & password.
func (s *PSQLTest) RunTest(tst test.Test, target string, opts test.TestOptions) error {
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
	port := 5432
	if tst.Arguments["port"] != "" {
		port, err = strconv.Atoi(tst.Arguments["port"])
		if err != nil {
			return err
		}
	}

	//
	// The default SSL mode
	//
	ssl := "disable"
	if tst.Arguments["tsl"] != "" {
		ssl = tst.Arguments["tsl"]
	}

	//
	// This is the string we'll use for the database connection.
	//
	connect := fmt.Sprintf("host=%s port='%d' user='%s' password='%s' connect_timeout='%d' sslmode='%s'", target, port, tst.Arguments["username"], tst.Arguments["password"], opts.Timeout, ssl)

	//
	// Show the config, if appropriate.
	//
	if opts.Verbose {
		fmt.Printf("\tPSQL connection string is %s\n", connect)
	}

	//
	// Connect to the database
	//
	db, err := sql.Open("postgres", connect)
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
	Register("psql", func() ProtocolTest {
		return &PSQLTest{}
	})
}
