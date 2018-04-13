package protocols

import (
	"crypto/tls"
	"fmt"
	"net"
	"regexp"
	"strconv"
	"strings"

	client "github.com/emersion/go-imap/client"
)

//
// Our structure.
//
// We store state in the `input` field.
//
type IMAPSTest struct {
	input   string
	options TestOptions
}

//
// Run the test against the specified target.
//
func (s *IMAPSTest) RunTest(target string) error {

	var err error

	//
	// The default port to connect to.
	//
	port := 993

	//
	// If the user specified a different port update it.
	//
	re := regexp.MustCompile("on\\s+port\\s+([0-9]+)")
	out := re.FindStringSubmatch(s.input)
	if len(out) == 2 {
		port, err = strconv.Atoi(out[1])
		if err != nil {
			return err
		}
	}

	//
	// Insecure operation allows us to skip validation of
	// the SSL certificate
	//
	insecure := false
	if strings.Contains(s.input, " insecure") {
		insecure = true
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

	if insecure {
		tls := &tls.Config{
			InsecureSkipVerify: true,
		}
		_, err = client.DialWithDialerTLS(dial, address, tls)
	} else {
		_, err = client.DialWithDialerTLS(dial, address, nil)
	}

	if err != nil {
		return err
	}

	return nil
}

//
// Store the complete line from the parser in our private
// field; this could be used if there are protocol-specific options
// to be understood.
//
func (s *IMAPSTest) SetLine(input string) {
	s.input = input
}

//
// Store the options for this test
//
func (s *IMAPSTest) SetOptions(opts TestOptions) {
	s.options = opts
}

//
// Register our protocol-tester.
//
func init() {
	Register("imaps", func() ProtocolTest {
		return &IMAPSTest{}
	})
}
