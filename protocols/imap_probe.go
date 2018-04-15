package protocols

import (
	"fmt"
	"net"
	"strconv"
	"strings"

	client "github.com/emersion/go-imap/client"
)

//
// Our structure.
//
// We store state in the `input` field.
//
type IMAPTest struct {
	input   string
	options TestOptions
}

//
// Run the test against the specified target.
//
func (s *IMAPTest) RunTest(target string) error {

	var err error

	//
	// The default port to connect to.
	//
	port := 143

	//
	// If the user specified a different port update to use it.
	//
	out := ParseArguments(s.input)
	if out["port"] != "" {
		port, err = strconv.Atoi(out["port"])
		if err != nil {
			return err
		}
	}

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

	var dial = &net.Dialer{
		Timeout: s.options.Timeout,
	}

	//
	// Connect.
	//
	conn, err := client.DialWithDialer(dial, address)
	if err != nil {
		return err
	}

	//
	// If we got username/password then use them
	//
	if (out["username"] != "") && (out["password"] != "") {
		err = con.Login(out["username"], out["password"])
		if err != nil {
			return err
		}
	}

	return nil
}

//
// Store the complete line from the parser in our private
// field; this could be used if there are protocol-specific options
// to be understood.
//
func (s *IMAPTest) SetLine(input string) {
	s.input = input
}

//
// Store the options for this test
//
func (s *IMAPTest) SetOptions(opts TestOptions) {
	s.options = opts
}

//
// Register our protocol-tester.
//
func init() {
	Register("imap", func() ProtocolTest {
		return &IMAPTest{}
	})
}
