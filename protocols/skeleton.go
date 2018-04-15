// Skeleton Test
//
// This is a skeleton test, for documentation purposes.
//
// Each protocol-test must conform to the ProtocolTest interface, which
// means implementing the RunTest method.
//
// Assuming you have an object which implements this method you can register
// your handler via the Register function of the protocol_api.go class.
//
package protocols

import (
	"fmt"
	"net"
	"strconv"
	"strings"

	"github.com/skx/overseer/test"
)

// SKELETONTest is our object.
type SKELETONTest struct {
}

// RunTest is the method that is invoked to perform the test.
//
// There are three arguments:
//
// 1. The test object which is to be executed, this contains the target hostname, etc.
//
// 2. The target IP address to run the test against, this might be IPv4 or IPv6.
//
// 3. An instance of TestOptions, which could be used to modify behaviour.
//
// The function should return a suitably descriptive error when the
// test fails, otherwise it should return `nil` to indicate that the test
// passed.
//
func (s *SKELETONTest) RunTest(tst test.Test, target string, opts TestOptions) error {
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
	// Set an explicit timeout, via the period in our
	// options.
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
	_, err = d.Dial("tcp", address)
	if err != nil {
		return err
	}

	//
	// Because the connection passed we can return nil here,
	// which indicates that there was no problem probing our
	// host.
	//
	return nil
}

// Register our protocol-tester.
//
// Once this has been done our function will be invoked whenever a
// test is found that is of the form:
//
//    target.example.com must run skeleton
//
func init() {
	Register("skeleton", func() ProtocolTest {
		return &SKELETONTest{}
	})
}
