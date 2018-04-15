package protocols

import (
	"crypto/tls"
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
	// Setup a dialer so we can have a suitable timeout
	//
	var dial = &net.Dialer{
		Timeout: s.options.Timeout,
	}

	//
	// Setup the default TLS config.
	//
	// We need to setup the hostname that the TLS certificate
	// will verify upon, from our input-line.
	//
	data := strings.Fields(s.input)
	tls_setup := &tls.Config{ServerName: data[0]}

	//
	// Disable verification if we're being insecure.
	//
	if insecure {
		tls_setup = &tls.Config{
			InsecureSkipVerify: true,
		}
	}

	//
	// Connect.
	//
	con, err := client.DialWithDialerTLS(dial, address, tls_setup)
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
