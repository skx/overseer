// Finger Tester
//
// The finger tester connects to a remote host and ensures that a response
// is received.
//
// This test is invoked via input like so:
//
//    host.example.com must run finger with user 'skx'
//
//  If you wish you can regard the fetch as a failure unless some specific
// text is returned too:
//
//    host.example.com must run finger with user 'skx' with content '2018'
//
//  NOTE: The user-argument is mandatory.
//

package protocols

import (
	"bufio"
	"errors"
	"fmt"
	"net"
	"strconv"
	"strings"

	"github.com/skx/overseer/test"
)

// FINGERTest is our object.
type FINGERTest struct {
}

// Arguments returns the names of arguments which this protocol-test
// understands, along with corresponding regular-expressions to validate
// their values.
func (s *FINGERTest) Arguments() map[string]string {
	known := map[string]string{
		"content": ".*",
		"port":    "^[0-9]+$",
		"user":    ".*",
	}
	return known
}

// Example returns sample usage-instructions for self-documentation purposes.
func (s *FINGERTest) Example() string {
	str := `
Finger Tester
-------------
 The finger tester connects to a remote host and ensures that a response
 is received.

 This test is invoked via input like so:

    host.example.com must run finger with user 'skx'

 If you wish you can regard the fetch as a failure unless some specific
 text is returned too:

    host.example.com must run finger with user 'skx' with content '2018'

 NOTE: The user-argument is mandatory.
`
	return str
}

// RunTest is the part of our API which is invoked to actually execute a
// test against the given target.
//
// In this case we make a TCP connection, defaulting to port 79, and
// look for a non-empty response.
func (s *FINGERTest) RunTest(tst test.Test, target string, opts test.TestOptions) error {
	var err error

	//
	// Ensure we have a username
	//
	if tst.Arguments["user"] == "" {
		return errors.New("A 'user' argument is mandatory.")
	}

	//
	// The default port to connect to.
	//
	port := 79

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

	//
	// Send the username
	//
	_, err = fmt.Fprintf(conn, tst.Arguments["user"]+"\r\n")
	if err != nil {
		return err
	}

	//
	// Read the response from the finger-server
	//
	var output string
	output, err = bufio.NewReader(conn).ReadString('\n')
	if err != nil {
		return err
	}
	conn.Close()

	//
	// If we didn't get a response of some kind, (i.e. "~/.plan" contents)
	// then the test failed.
	//
	if output == "" {
		return errors.New("The server didn't send a response.")
	}

	//
	// If we require some specific content in the response we should
	// test for that here.
	//
	content := tst.Arguments["content"]
	if content != "" {
		if !strings.Contains(output, content) {
			return fmt.Errorf("The finger-output did not contain the required text '%s'", content)
		}
	}

	return nil
}

//
// Register our protocol-tester.
//
func init() {
	Register("finger", func() ProtocolTest {
		return &FINGERTest{}
	})
}
