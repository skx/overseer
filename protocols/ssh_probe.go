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

//
// Our structure.
//
type SSHTest struct {
}

// Return the arguments which this protocol-test understands.
func (s *SSHTest) Arguments() []string {
	known := []string{"port"}
	return known
}

//
// Run the test against the specified target.
//
func (s *SSHTest) RunTest(tst test.Test, target string, opts test.TestOptions) error {
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

	if !strings.Contains(banner, "OpenSSH") {
		return errors.New("Banner doesn't look like OpenSSH")
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
