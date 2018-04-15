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

//
// Run the test against the specified target.
//
func (s *IMAPSTest) RunTest(tst test.Test, target string, opts TestOptions) error {
	var err error

	fmt.Printf("target:%s test.target:%s\n", target, tst.Target)
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
