// IMAP Tester
//
// The IMAP tester connects to a remote host and ensures that this
// succeeds.  If you supply a username & password a login will be
// made, and the test will fail if this login fails.
//
// This test is invoked via input like so:
//
//    host.example.com must run imap [with username 'steve@steve' with password 'secret']
//
package protocols

import (
	"fmt"
	"net"
	"strconv"
	"strings"

	client "github.com/emersion/go-imap/client"
	"github.com/skx/overseer/test"
)

//
// Our structure.
//
type IMAPTest struct {
}

//
// Run the test against the specified target.
//
func (s *IMAPTest) RunTest(tst test.Test, target string, opts TestOptions) error {

	var err error

	fmt.Printf("target:%s test.target:%s\n", target, tst.Target)

	//
	// The default port to connect to.
	//
	port := 143

	//
	// If the user specified a different port update to use it.
	//
	if tst.Arguments["port"] != "" {
		port, err = strconv.Atoi(tst.Arguments["port"])
		if err != nil {
			return err
		}
	}

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

	var dial = &net.Dialer{
		Timeout: opts.Timeout,
	}

	//
	// Connect.
	//
	con, err := client.DialWithDialer(dial, address)
	if err != nil {
		return err
	}

	//
	// If we got username/password then use them
	//
	if (tst.Arguments["username"] != "") && (tst.Arguments["password"] != "") {
		err = con.Login(tst.Arguments["username"], tst.Arguments["password"])
		if err != nil {
			return err
		}
	}

	return nil
}

//
// Register our protocol-tester.
//
func init() {
	Register("imap", func() ProtocolTest {
		return &IMAPTest{}
	})
}
