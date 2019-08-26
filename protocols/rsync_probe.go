// Rsync Tester
//
// The Rsync tester connects to a remote host and ensures that a response
// is received that looks like an rsync-server banner.
//
// This test is invoked via input like so:
//
//    host.example.com must run rsync [with port 873]
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

// RSYNCTest is our object.
type RSYNCTest struct {
}

// Arguments returns the names of arguments which this protocol-test
// understands, along with corresponding regular-expressions to validate
// their values.
func (s *RSYNCTest) Arguments() map[string]string {
	known := map[string]string{
		"port": "^[0-9]+$",
	}
	return known
}

// Example returns sample usage-instructions for self-documentation purposes.
func (s *RSYNCTest) Example() string {
	str := `
Rsync Tester
------------
 The rsync tester connects to a remote host and ensures that a response
 is received that looks like an rsync-server banner.

 This test is invoked via input like so:

    host.example.com must run rsync
`
	return str
}

// RunTest is the part of our API which is invoked to actually execute a
// test against the given target.
//
// In this case we make a TCP connection, defaulting to port 873, and
// look for a response which appears to be an rsync-server.
func (s *RSYNCTest) RunTest(tst test.Test, target string, opts test.Options) error {
	var err error

	//
	// The default port to connect to.
	//
	port := 873

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

	if !strings.Contains(banner, "RSYNC") {
		return errors.New("banner doesn't look like a rsync-banner")
	}

	return nil
}

//
// Register our protocol-tester.
//
func init() {
	Register("rsync", func() ProtocolTest {
		return &RSYNCTest{}
	})
}
