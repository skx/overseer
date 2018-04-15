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

//
// Our structure.
//
type FTPTest struct {
}

//
// Run the test against the specified target.
//
func (s *FTPTest) RunTest(tst test.Test, target string, opts TestOptions) error {
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
