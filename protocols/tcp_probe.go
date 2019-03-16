// TCP Tester
//
// The TCP tester connects to a remote host and does nothing else.
//
// In short it determines whether a TCP-based service is reachable,
// by excluding errors such as "host not found", or "connection refused".
//
// This test is invoked via input like so:
//
//    host.example.com must run tcp with port 123
//
//  The port-setting is mandatory, such that the tests knows what to connect to.
//
// Optionally you may specify a regular expression to match against a
// banner the remote host sends on connection:
//
//    host.example.com must run tcp with port 655 with banner '0 \S+ 17'
//

package protocols

import (
	"bufio"
	"errors"
	"fmt"
	"net"
	"regexp"
	"strconv"
	"strings"

	"github.com/skx/overseer/test"
)

// TCPTest is our object
type TCPTest struct {
}

// Arguments returns the names of arguments which this protocol-test
// understands, along with corresponding regular-expressions to validate
// their values.
func (s *TCPTest) Arguments() map[string]string {
	known := map[string]string{
		"port":   "^[0-9]+$",
		"banner": ".*",
	}
	return known
}

// Example returns sample usage-instructions for self-documentation purposes.
func (s *TCPTest) Example() string {
	str := `
TCP Tester
----------
 The TCP tester connects to a remote host and does nothing else.

 In short it determines whether a TCP-based service is reachable,
 by excluding errors such as "host not found", or "connection refused".

 This test is invoked via input like so:

    host.example.com must run tcp with port 123

 The port-setting is mandatory, such that the tests knows what to connect to.

 Optionally you may specify a regular expression to match against a
 banner the remote host sends on connection:

    host.example.com must run tcp with port 655 with banner '0 \S+ 17'
`
	return str
}

// RunTest is the part of our API which is invoked to actually execute a
// test against the given target.
//
// In this case we make a TCP connection to the specified port, and assume
// that everything is OK if that succeeded.
func (s *TCPTest) RunTest(tst test.Test, target string, opts test.Options) error {
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

	defer conn.Close()

	//
	// If we're going to do a banner match then we should read a line
	// from the host
	//
	if tst.Arguments["banner"] != "" {

		// Compile the regular expression
		re, error := regexp.Compile("(?ms)" + tst.Arguments["banner"])
		if error != nil {
			return error
		}

		// Read a single line of input
		banner, err := bufio.NewReader(conn).ReadString('\n')
		if err != nil {
			return err
		}

		//
		// If the regexp doesn't match that's an error.
		//
		match := re.FindAllStringSubmatch(string(banner), -1)
		if len(match) < 1 {
			return fmt.Errorf("Remote banner '%s' didn't match the regular expression '%s'", banner, tst.Arguments["banner"])
		}
	}

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
