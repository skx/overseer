// IMAPS Tester
//
// The IMAPS tester connects to a remote host and ensures that this
// succeeds.  If you supply a username & password a login will be
// made, and the test will fail if this login fails.
//
// This test is invoked via input like so:
//
//    host.example.com must run imap [with username 'steve@steve' with password 'secret']
//
// Because IMAPS uses TLS it will test the validity of the certificate as
// part of the test, if you wish to disable this add `with tls insecure`.
//
package protocols

import (
	"crypto/tls"
	"fmt"
	"net"
	"strconv"
	"strings"

	client "github.com/emersion/go-imap/client"
	"github.com/skx/overseer/test"
)

//
// Our structure.
//
type IMAPSTest struct {
}

// Return the arguments which this protocol-test understands.
func (s *IMAPSTest) Arguments() map[string]string {
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
func (s *IMAPSTest) RunTest(tst test.Test, target string, opts test.TestOptions) error {
	var err error

	//
	// The default port to connect to.
	//
	port := 993

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
	// Setup a dialer so we can have a suitable timeout
	//
	var dial = &net.Dialer{
		Timeout: opts.Timeout,
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
	// Disable verification if we're being insecure.
	//
	if insecure {
		tlsSetup = &tls.Config{
			InsecureSkipVerify: true,
		}
	}

	//
	// Connect.
	//
	con, err := client.DialWithDialerTLS(dial, address, tlsSetup)
	if err != nil {
		return err
	}

	//
	// If we got username/password then use them
	//
	if (tst.Arguments["username"] != "") && (tst.Arguments["password"] != "") {
		err = con.Login(tst.Arguments["username"], tst.Arguments["password"])
		if err != nil {
			return err
		}
	}

	return nil
}

//
// Register our protocol-tester.
//
func init() {
	Register("imaps", func() ProtocolTest {
		return &IMAPSTest{}
	})
}
