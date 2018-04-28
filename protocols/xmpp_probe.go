// XMPP Tester
//
// The XMPP tester connects to a remote host and ensures that a response
// is received that looks like an XMPP-server banner.
//
// This test is invoked via input like so:
//
//    host.example.com must run xmpp [with port 5222]
//

package protocols

import (
	"bufio"
	"fmt"
	"net"
	"strconv"
	"strings"

	"github.com/skx/overseer/test"
)

// XMPPTest is our object
type XMPPTest struct {
}

// Arguments returns the names of arguments which this protocol-test
// understands, along with corresponding regular-expressions to validate
// their values.
func (s *XMPPTest) Arguments() map[string]string {
	known := map[string]string{
		"port": "^[0-9]+$",
	}
	return known
}

// RunTest is the part of our API which is invoked to actually execute a
// test against the given target.
//
// In this case we make a TCP connection, defaulting to port 5222, and
// look for a response which appears to be an XMPP-server.
func (s *XMPPTest) RunTest(tst test.Test, target string, opts test.TestOptions) error {
	var err error

	//
	// The default port to connect to.
	//
	port := 5222

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
	// Send a (bogus) greeting
	//
	_, err = conn.Write([]byte("<>\n"))
	if err != nil {
		return err
	}

	//
	// Read the response.
	//
	banner, err := bufio.NewReader(conn).ReadString('>')
	if err != nil {
		return err
	}

	//
	// Now close the connection
	//
	err = conn.Close()
	if err != nil {
		return err
	}

	if !strings.Contains(banner, "<?xml") {
		return fmt.Errorf("Banner doesn't look like an XMPP-banner '%s'", banner)
	}

	return nil
}

//
// Register our protocol-tester.
//
func init() {
	Register("xmpp", func() ProtocolTest {
		return &XMPPTest{}
	})
}
