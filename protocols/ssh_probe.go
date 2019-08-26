// SSH Tester
//
// The SSH tester connects to a remote host and ensures that a response
// is received that looks like an SSH-server banner.
//
// This test is invoked via input like so:
//
//    host.example.com must run ssh [with port 22]
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

// SSHTest is our object.
type SSHTest struct {
}

// Arguments returns the names of arguments which this protocol-test
// understands, along with corresponding regular-expressions to validate
// their values.
func (s *SSHTest) Arguments() map[string]string {
	known := map[string]string{
		"port": "^[0-9]+$",
	}
	return known
}

// Example returns sample usage-instructions for self-documentation purposes.
func (s *SSHTest) Example() string {
	str := `
SSH Tester
----------
 The ssh tester connects to a remote host and ensures that a response
 is received that looks like an ssh-server banner.

 This test is invoked via input like so:

    host.example.com must run ssh
`
	return str
}

// RunTest is the part of our API which is invoked to actually execute a
// test against the given target.
//
// In this case we make a TCP connection, defaulting to port 22, and
// look for a response which appears to be an SSH-server.
func (s *SSHTest) RunTest(tst test.Test, target string, opts test.Options) error {
	var err error

	//
	// The default port to connect to.
	//
	port := 22

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

	if !strings.Contains(banner, "SSH-") {
		return errors.New("banner doesn't look like an SSH server")
	}

	return nil
}

//
// Register our protocol-tester.
//
func init() {
	Register("ssh", func() ProtocolTest {
		return &SSHTest{}
	})
}
