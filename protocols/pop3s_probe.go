// POP3 Tester
//
// The POP3 tester connects to a remote host and ensures that this
// succeeds.  If you supply a username & password a login will be
// made, and the test will fail if this login fails.
//
// This test is invoked via input like so:
//
//    host.example.com must run pop3 [with username 'steve@steve' with password 'secret']
//
// Because POP3S uses TLS it will test the validity of the certificate as
// part of the test, if you wish to disable this add `with tls insecure`.
//
package protocols

import (
	"crypto/tls"
	"fmt"
	"strconv"
	"strings"

	"github.com/simia-tech/go-pop3"
	"github.com/skx/overseer/test"
)

//
// Our structure.
//
type POP3STest struct {
}

// Return the arguments which this protocol-test understands.
func (s *POP3STest) Arguments() map[string]string {
	known := map[string]string{
		"port":     "^[0-9]+$",
		"tls":      "insecure",
		"username": ".*",
		"password": ".*",
	}
	return known
}

//
// Run the test against the specified target.
//
func (s *POP3STest) RunTest(tst test.Test, target string, opts test.TestOptions) error {
	var err error

	//
	// The default port to connect to.
	//
	port := 995

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
	// Should we skip validation of the SSL certificate?
	//
	insecure := false
	if tst.Arguments["tls"] == "insecure" {
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
	// Setup the default TLS config.
	//
	// We need to setup the hostname that the TLS certificate
	// will verify upon, from our input-line.
	//
	data := strings.Fields(tst.Input)
	tlsSetup := &tls.Config{ServerName: data[0]}

	//
	// If we're being insecure then remove the verification
	//
	if insecure {
		tlsSetup = &tls.Config{
			InsecureSkipVerify: true,
		}
	}

	//
	// Connect
	//
	c, err := pop3.Dial(address, pop3.UseTLS(tlsSetup), pop3.UseTimeout(opts.Timeout))
	if err != nil {
		return err
	}

	//
	// Did we get a username/password?  If so try to authenticate
	// with them
	//
	if (tst.Arguments["username"] != "") && (tst.Arguments["password"] != "") {
		err = c.Auth(tst.Arguments["username"], tst.Arguments["password"])
		if err != nil {
			return err
		}
	}

	//
	// Quit and return
	//
	c.Quit()

	return nil
}

//
// Register our protocol-tester.
//
func init() {
	Register("pop3s", func() ProtocolTest {
		return &POP3STest{}
	})
}
