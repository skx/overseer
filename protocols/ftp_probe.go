// FTP Tester
//
// The FTP tester connects to a remote host and ensures that a response
// is received that looks like an FTP-server banner.
//
// This test is invoked via input like so:
//
//    host.example.com must run ftp [with port 21]
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

// FTPTest is our object.
type FTPTest struct {
}

// Arguments returns the names of arguments which this protocol-test
// understands, along with corresponding regular-expressions to validate
// their values.
func (s *FTPTest) Arguments() map[string]string {
	known := map[string]string{
		"port": "^[0-9]+$",
	}
	return known
}

// RunTest is the part of our API which is invoked to actually execute a
// test against the given target.
//
// In this case we make a TCP connection, defaulting to port 21, and
// look for a response which appears to be an FTP-server.
func (s *FTPTest) RunTest(tst test.Test, target string, opts test.TestOptions) error {
	var err error

	//
	// The default port to connect to.
	//
	port := 21

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

	if !strings.Contains(banner, "220") {
		return errors.New("Banner doesn't look like an FTP success")
	}

	return nil
}

//
// Register our protocol-tester.
//
func init() {
	Register("ftp", func() ProtocolTest {
		return &FTPTest{}
	})
}
