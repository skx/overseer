package protocols

import (
	"crypto/tls"
	"fmt"
	"strconv"
	"strings"

	"github.com/simia-tech/go-pop3"
)

//
// Our structure.
//
// We store state in the `input` field.
//
type POP3STest struct {
	input   string
	options TestOptions
}

//
// Run the test against the specified target.
//
func (s *POP3STest) RunTest(target string) error {
	var err error

	//
	// The default port to connect to.
	//
	port := 995

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
	// Should we skip validation of the SSL certificate?
	//
	insecure := false
	if out["tls"] == "insecure" {
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

	//
	// Setup the default TLS config
	//
	tls_setup := &tls.Config{}

	//
	// If we're being insecure then remove the verification
	//
	if insecure {
		tls_setup = &tls.Config{
			InsecureSkipVerify: true,
		}
	}

	//
	// Connect
	//
	c, err := pop3.Dial(address, pop3.UseTLS(tls_setup), pop3.UseTimeout(s.options.Timeout))

	if err != nil {
		return err
	}

	c.Quit()

	return nil
}

//
// Store the complete line from the parser in our private
// field; this could be used if there are protocol-specific options
// to be understood.
//
func (s *POP3STest) SetLine(input string) {
	s.input = input
}

//
// Store the options for this test
//
func (s *POP3STest) SetOptions(opts TestOptions) {
	s.options = opts
}

//
// Register our protocol-tester.
//
func init() {
	Register("pop3s", func() ProtocolTest {
		return &POP3STest{}
	})
}
