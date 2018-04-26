// TCP Tester
//
// The TCP tester connects to a remote host and does nothign else.
//
// In short it determines whether a TCP-based service is reachable,
// by excluding errors such as "host not found", or "connection refused".
//
// This test is invoked via input like so:
//
//    host.example.com must run tcp [with port 873]
//
package protocols

import (
	"errors"
	"fmt"
	"net"
	"strconv"
	"strings"

	"github.com/skx/overseer/test"
)

//
// Our structure.
//
type TCPTest struct {
}

// Return the arguments which this protocol-test understands.
func (s *TCPTest) Arguments() []string {
	known := []string{"port"}
	return known
}

//
// Run the test against the specified target.
//
func (s *TCPTest) RunTest(tst test.Test, target string, opts test.TestOptions) error {
	var err error

	//
	// The default port to connect to.
	//
	port := -1

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
	// If there was no port that's an error
	//
	if port == -1 {
		return errors.New("You must specify the port for TCP-tests")
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

	//
	// And close it: without having read or written to it
	//
	conn.Close()

	return nil
}

//
// Register our protocol-tester.
//
func init() {
	Register("tcp", func() ProtocolTest {
		return &TCPTest{}
	})
}
