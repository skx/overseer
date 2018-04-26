// HTTP Tester
//
// The HTTP tester allows you to confirm that a remote HTTP-server is
// responding correctly.  You may test the response of a HTTP GET or
// POST request.
//
// This test is invoked via input like so:
//
//    http://example.com/ must run http
//
// By default a remote HTTP-server is considered working if it responds
// with a HTTP status-code of 200, but you can change this via:
//
//    with status 301
//
// Or if you do not care about the specific status-code at all, but you
// wish to see an alert when a connection-refused/failed/timeout condition
// occurs you could say:
//
//    with status any
//
// It is also possible to regard a fetch as a failure if the response body
// does not contain a particular piece of test.  For example the following
// would be regarded as a failure if my website did not contain my name
// in the body of the response:
//
//    https://steve.fi/ must run http with content 'Steve Kemp'
//
// If your URL requires the use of HTTP basic authentication this is
// supported by adding a username and password parameter to your test,
// for example:
//
//    https://jigsaw.w3.org/HTTP/Basic/ must run http with username 'guest' with password 'guest' with content "Your browser made it"
//
// If you need to disable failures due to expired, broken, or
// otherwise bogus SSL certificates you can do so via the tls setting:
//
//    https://expired.badssl.com/ must run http with tls insecure
//
// Finally if you submit a "data" argument, like in this next example
// the request made will be a HTTP POST:
//
//    https://steve.fi/Security/XSS/Tutorial/filter.cgi must run http with data "text=test%20me" with content "test me"
//
//
// NOTE: This test deliberately does not follow redirections, to allow
// enhanced testing.
//
package protocols

import (
	"bytes"
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"strconv"
	"strings"

	"github.com/skx/overseer/test"
)

//
// Our structure.
//
type HTTPTest struct {
}

// Return the arguments which this protocol-test understands.
func (s *HTTPTest) Arguments() []string {
	known := []string{"tls", "data", "username", "password", "status", "content"}
	return known
}

//
// Make a HTTP-test against the given URL.
//
//
// For the purposes of clarity this test makes a HTTP-fetch.  The `test.Test`
// structure contains are raw test, and the `target` variable contains the
// IP address to make the request to.
//
// So:
//
//    tst.Target => "https://steve.kemp.fi/
//
//    target => "176.9.183.100"
//
func (s *HTTPTest) RunTest(tst test.Test, target string, opts test.TestOptions) error {

	//
	// Determine the port to connect to, via the protocol
	// in the URI.
	//
	port := 80
	if strings.HasPrefix(tst.Target, "https:") {
		port = 443
	}

	//
	// Be clear about the IP vs. the hostname.
	//
	address := target
	target = tst.Target

	//
	// Setup a dialer which will be dual-stack
	//
	dialer := &net.Dialer{
		DualStack: true,
	}

	//
	// This is where some magic happens, we want to connect and do
	// a http check on http://example.com/, but we want to do that
	// via the IP address.
	//
	// We could do that manually by connecting to http://1.2.3.4,
	// and sending the appropriate HTTP Host: header but that risks
	// a bit of complexity with SSL in particular.
	//
	// So instead we fake the address in the dialer object, so that
	// we don't rewrite anything, don't do anything manually, and
	// instead just connect to the right IP by magic.
	//
	dial := func(ctx context.Context, network, addr string) (net.Conn, error) {
		//
		// Assume an IPv4 address by default.
		//
		addr = fmt.Sprintf("%s:%d", address, port)

		//
		// If we find a ":" we know it is an IPv6 address though
		//
		if strings.Contains(address, ":") {
			addr = fmt.Sprintf("[%s]:%d", address, port)
		}

		//
		// Use the replaced/updated address in our connection.
		//
		return dialer.DialContext(ctx, network, addr)
	}

	//
	// Create a context which uses the dial-context
	//
	// The dial-context is where the magic happens.
	//
	tr := &http.Transport{
		DialContext: dial,
	}

	//
	// If we're running insecurely then ignore SSL errors
	//
	if tst.Arguments["tls"] == "insecure" {
		tr.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	}

	//
	// Create a client with a timeout, disabled redirection, and
	// the magical transport we've just created.
	//
	var netClient = &http.Client{
		Timeout: opts.Timeout,

		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
		Transport: tr,
	}

	//
	// Now we can make a request-object
	//
	var req *http.Request
	var err error

	//
	// If we have no data then make a GET request
	//
	if tst.Arguments["data"] == "" {
		req, err = http.NewRequest("GET", target, nil)
	} else {

		//
		// Otherwise make a HTTP POST request, with
		// the specified data.
		//
		req, err = http.NewRequest("POST", target,
			bytes.NewBuffer([]byte(tst.Arguments["data"])))
	}
	if err != nil {
		return err
	}

	//
	// Are we using basic-auth?
	//
	if tst.Arguments["username"] != "" {
		req.SetBasicAuth(tst.Arguments["username"],
			tst.Arguments["password"])
	}

	//
	// Perform the request
	//
	response, err := netClient.Do(req)
	if err != nil {
		return err
	}

	//
	// Get the body and status-code.
	//
	defer response.Body.Close()
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return err
	}
	status := response.StatusCode

	//
	// The default status-code we accept as OK
	//
	ok := 200

	//
	// Did the user want to look for a specific status-code?
	//
	if tst.Arguments["status"] != "" {
		// Status code might be "any"
		if tst.Arguments["status"] != "any" {
			ok, err = strconv.Atoi(tst.Arguments["status"])
			if err != nil {
				return err
			}
		}
	}

	//
	// See if the status-code matched our expectation(s).
	//
	// If they mis-match that means the test failed, unless the user
	// said "with status any".
	//
	if ok != status && (tst.Arguments["status"] != "any") {
		return errors.New(fmt.Sprintf("Status code was %d not %d", status, ok))
	}

	//
	// Looking for a body-match?
	//
	if tst.Arguments["content"] != "" {
		if !strings.Contains(string(body), tst.Arguments["content"]) {
			return errors.New(
				fmt.Sprintf("Body didn't contain '%s'", tst.Arguments["content"]))
		}

	}

	//
	// If we reached here all is OK
	//
	return nil
}

//
// Register our protocol-tester.
//
func init() {
	Register("http", func() ProtocolTest {
		return &HTTPTest{}
	})
}
