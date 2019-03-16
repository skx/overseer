// Telnet Tester
//
// The telnet tester connects to a remote host and does nothing else.
//
// This test is invoked via input like so:
//
//    host.example.com must run telnet
//

package protocols

import (
	"fmt"
	"net"
	"strconv"
	"strings"

	"github.com/skx/overseer/test"
)

// TELNETTest is our object
type TELNETTest struct {
}

// Arguments returns the names of arguments which this protocol-test
// understands, along with corresponding regular-expressions to validate
// their values.
func (s *TELNETTest) Arguments() map[string]string {
	known := map[string]string{
		"port": "^[0-9]+$",
	}
	return known
}

// Example returns sample usage-instructions for self-documentation purposes.
func (s *TELNETTest) Example() string {
	str := `
Telnet Tester
-------------
 The telnet tester connects to a remote host and does nothing else.

 This test is invoked via input like so:

    host.example.com must run telnet
`
	return str
}

// RunTest is the part of our API which is invoked to actually execute a
// test against the given target.
//
// In this case we make a TCP connection to the specified port, and assume
// that everything is OK if that succeeded.
func (s *TELNETTest) RunTest(tst test.Test, target string, opts test.Options) error {
	var err error

	//
	// The default port to connect to.
	//
	port := 23

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
	// Set an explicit timeout
	//
	d := net.Dialer{Timeout: opts.Timeout}

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
	// Make the TCP connection.
	//
	conn, err := d.Dial("tcp", address)
	if err != nil {
		return err
	}
	conn.Close()

	return nil
}

//
// Register our protocol-tester.
//
func init() {
	Register("telnet", func() ProtocolTest {
		return &TELNETTest{}
	})
}
