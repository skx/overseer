// SMTP Tester
//
// The SMTP tester checks on the status of a remote SMTP-server.
//
// This test is invoked via input like so:
//
//    host.example.com must run smtp [with port 25]
//
// By default a connection will be attempted and nothing else.  A more
// complete test would be to specify a username & password and test that
// authentication succeeds.
//
// Note that performing an authentication-request requires the use of
// `STARTTLS`.  If the TLS certificate is self-signed or otherwise
// non-trusted you'll need to disable the validity checking by appending
// `with tls insecure`.
//
// A complete example, testing a login, will look like this:
//
//    host.example.com must run smtp [with port 587] with username 'steve@example.com' with password 'secret'  [with tls insecure]
//
//

package protocols

import (
	"crypto/tls"
	"errors"
	"fmt"
	"net"
	"net/smtp"
	"strconv"
	"strings"

	"github.com/skx/overseer/test"
)

// SMTPTest is our object
type SMTPTest struct {
}

// Arguments returns the names of arguments which this protocol-test
// understands, along with corresponding regular-expressions to validate
// their values.
func (s *SMTPTest) Arguments() map[string]string {
	known := map[string]string{
		"port":     "^[0-9]+$",
		"username": ".*",
		"password": ".*",
		"tls":      "insecure",
	}
	return known
}

// Example returns sample usage-instructions for self-documentation purposes.
func (s *SMTPTest) Example() string {
	str := `
SMTP Tester
-----------
 The SMTP tester checks on the status of a remote SMTP-server.

 This test is invoked via input like so:

    host.example.com must run smtp [with port 25]

 By default a connection will be attempted and nothing else.  A more
 complete test would be to specify a username & password and test that
 authentication succeeds.

 Note that performing an authentication-request requires the use of
 STARTTLS.  If the TLS certificate is self-signed or otherwise
 non-trusted you'll need to disable the validity checking by appending
 'with tls insecure'.

 A complete example, testing a login, will look like this:

    host.example.com must run smtp [with port 587] with username 'steve@example.com' with password 's3cr3t'  [with tls insecure]
`
	return str
}

// RunTest is the part of our API which is invoked to actually execute a
// test against the given target.
//
// In this case we make a TCP connection, defaulting to port 25, and
// look for a response which appears to be an SMTP-server.
func (s *SMTPTest) RunTest(tst test.Test, target string, opts test.TestOptions) error {
	var err error

	//
	// The default port to connect to.
	//
	port := 25

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

	// The default TLS configuration verifies the certificate
	// matches the hostname of our target.
	tlsconfig := &tls.Config{
		ServerName: tst.Target,
	}

	// However if the user is being insecure then we'll validate
	// nothing - allowing self-signed certificates, and hostname
	// mismatches.
	if tst.Arguments["tls"] == "insecure" {
		tlsconfig = &tls.Config{
			InsecureSkipVerify: true,
		}
	}

	// Create the SMTP-client
	client, err := smtp.NewClient(conn, tst.Target)
	if err != nil {
		return err
	}

	defer client.Close()

	if err = client.Hello(tst.Target); err != nil {
		return err
	}

	//
	// If we have a username & password then we have to
	// try them - but this will require TLS so we'll start
	// that first.
	//
	if tst.Arguments["username"] != "" &&
		tst.Arguments["password"] != "" {

		hasStartTLS, _ := client.Extension("STARTTLS")
		if !hasStartTLS {
			return errors.New("We cannot login without STARTTLS, and that was not advertised.")
		}

		if err = client.StartTLS(tlsconfig); err != nil {
			return err
		}

		//
		// In the future we might try more options
		//
		// CRAM MD5 is available in the net/smtp client at least.
		//
		auth := smtp.PlainAuth("", tst.Arguments["username"],
			tst.Arguments["password"], tst.Target)

		//
		// If auth failed then report that.
		//
		if err = client.Auth(auth); err != nil {
			return err
		}
	}

	// All done
	return nil
}

//
// Register our protocol-tester.
//
func init() {
	Register("smtp", func() ProtocolTest {
		return &SMTPTest{}
	})
}
