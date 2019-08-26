// VNC Tester
//
// The VNC tester connects to a remote host and ensures that a response
// is received that looks like an VNC banner.
//
// This test is invoked via input like so:
//
//    host.example.com must run vnc [with port 5900]
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

// VNCTest is our object
type VNCTest struct {
}

// Arguments returns the names of arguments which this protocol-test
// understands, along with corresponding regular-expressions to validate
// their values.
func (s *VNCTest) Arguments() map[string]string {
	known := map[string]string{
		"port": "^[0-9]+$",
	}
	return known
}

// Example returns sample usage-instructions for self-documentation purposes.
func (s *VNCTest) Example() string {
	str := `
VNC Tester
----------
 The VNC tester connects to a remote host and ensures that a response
 is received that looks like an VNC banner.

 This test is invoked via input like so:

    host.example.com must run vnc
`
	return str
}

// RunTest is the part of our API which is invoked to actually execute a
// test against the given target.
//
// In this case we make a TCP connection, defaulting to port 5900, and
// look for a response which appears to be an VNC-server.
func (s *VNCTest) RunTest(tst test.Test, target string, opts test.Options) error {
	var err error

	//
	// The default port to connect to.
	//
	port := 5900

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
	// Read the banner.
	//
	banner, err := bufio.NewReader(conn).ReadString('\n')
	if err != nil {
		return err
	}
	conn.Close()

	if !strings.Contains(banner, "RFB") {
		return errors.New("banner doesn't look like VNC")
	}

	return nil
}

//
// Register our protocol-tester.
//
func init() {
	Register("vnc", func() ProtocolTest {
		return &VNCTest{}
	})
}
