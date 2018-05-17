// NNTP Tester
//
// The NNTP tester connects to a remote host and ensures that a response
// is received that looks like an news-server banner.
//
// This test is invoked via input like so:
//
//    host.example.com must run nntp [with port 119]
//
// For a more complete test you can also validate the existance of a
// specific newsgroup:
//
//    blaine.gmane.org must run nntp with group 'gmane.org.wikimedia.foundation.uk'
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

// NNTPTest is our object.
type NNTPTest struct {
}

// Arguments returns the names of arguments which this protocol-test
// understands, along with corresponding regular-expressions to validate
// their values.
func (s *NNTPTest) Arguments() map[string]string {
	known := map[string]string{
		"port":  "^[0-9]+$",
		"group": ".*",
	}
	return known
}

// Example returns sample usage-instructions for self-documentation purposes.
func (s *NNTPTest) Example() string {
	str := `
NNTP Tester
-----------
 The NNTP tester connects to a remote host and ensures that a response
 is received that looks like an news-server banner.

 This test is invoked via input like so:

    host.example.com must run nntp [with port 119]

 For a more complete test you can also validate the existance of a
 specific newsgroup:

    blaine.gmane.org must run nntp with group 'gmane.org.wikimedia.foundation.uk'
`
	return str
}

// RunTest is the part of our API which is invoked to actually execute a
// test against the given target.
//
// In this case we make a TCP connection, defaulting to port 119, and
// look for a response which appears to be an NNTP-server.
func (s *NNTPTest) RunTest(tst test.Test, target string, opts test.TestOptions) error {
	var err error

	//
	// The default port to connect to.
	//
	port := 119

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
	defer conn.Close()

	if !strings.Contains(banner, "200") {
		return errors.New("Banner doesn't look like a news-server")
	}

	//
	// If we have a group try to select it
	//
	if tst.Arguments["group"] != "" {

		//
		// Select the group
		//
		_, err = fmt.Fprintf(conn, "GROUP "+tst.Arguments["group"]+"\n")
		if err != nil {
			return err
		}

		//
		// Read the response
		//
		resp, err := bufio.NewReader(conn).ReadString('\n')
		if err != nil {
			return err
		}

		if strings.HasPrefix(resp, "211") {
			return nil
		} else {
			return fmt.Errorf("Selecting group %s failed - %s", tst.Arguments["group"], resp)
		}
	}

	return nil
}

//
// Register our protocol-tester.
//
func init() {
	Register("nntp", func() ProtocolTest {
		return &NNTPTest{}
	})
}
