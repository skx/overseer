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
// does not contain a particular piece of text.  For example the following
// would be regarded as a failure if my website did not contain my name
// in the body of the response:
//
//    https://steve.fi/ must run http with content 'Steve Kemp'
//
// The 'content' setting looks for a literal match in the response-body,
// if you're looking for something more flexible you can instead test that
// the response-body matches a given regular-expression:
//
//   https://steve.fi/ must run http with pattern 'Steve\s+Kemp'
//
// (The regular expression will be assumed to be multi-line, and
// will also allow newlines to be matched with ".".)
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
// By default tests will fail if you're probing an SSL-site which has
// a certificate which will expire within the next 14 days.  To change
// the time-period specify it explicitly like so, if not stated the
// expiration period is assumed to be days:
//
//    # seven days
//    https://steve.fi/ must run http with expiration 7d
//
//    # 12 hours (!)
//    https://steve.fi/ must run http with expiration 12h
//
// To disable the SSL-expiration checking entirely specify "any":
//
//    https://steve.fi/ must run http with expiration any
//
// Finally if you submit a "data" argument, like in this next example
// the request made will be a HTTP POST:
//
//    https://steve.fi/Security/XSS/Tutorial/filter.cgi must run http with data "text=test%20me" with content "test me"
//
// But note that you can override the HTTP-verb via:
//
//    https://example.com/ must run http with method HEAD
//
// Combining these you can submit data with a PUT method:
//
//    https://steve.fi/Security/XSS/Tutorial/filter.cgi must run http with method PUT with data "text=test%20me" with content "test me"
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
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/skx/overseer/test"
)

// HTTPTest is our object.
type HTTPTest struct {
}

// Arguments returns the names of arguments which this protocol-test
// understands, along with corresponding regular-expressions to validate
// their values.
func (s *HTTPTest) Arguments() map[string]string {
	known := map[string]string{
		"user-agent": ".*",
		"content":    ".*",
		"data":       ".*",
		"expiration": "^(any|[0-9]+[hd]?)$",
		"method":     "^(GET|HEAD|POST|PUT|PATCH|DELETE)$",
		"password":   ".*",
		"pattern":    ".*",
		"status":     "^(any|[0-9]+)$",
		"tls":        "insecure",
		"username":   ".*",
	}
	return known
}

// Example returns sample usage-instructions for self-documentation purposes.
func (s *HTTPTest) Example() string {
	str := `
HTTP Tester
-----------
 The HTTP tester allows you to confirm that a remote HTTP-server is
 responding correctly.  You may test the response of a HTTP GET or
 POST request.

 This test is invoked via input like so:

   http://example.com/ must run http

 By default a remote HTTP-server is considered working if it responds
 with a HTTP status-code of 200, but you can change this via:

   with status 301

 Or if you do not care about the specific status-code at all, but you
 wish to see an alert when a connection-refused/failed/timeout condition
 occurs you could say:

   with status any

 It is also possible to regard a fetch as a failure if the response body
 does not contain a particular piece of text.  For example the following
 would be regarded as a failure if my website did not contain my name
 in the body of the response:

   https://steve.fi/ must run http with content 'Steve Kemp'

 The 'content' setting looks for a literal match in the response-body,
 if you're looking for something more flexible you can instead test that
 the response-body matches a given regular-expression:

   https://steve.fi/ must run http with pattern 'Steve\s+Kemp'

 (The regular expression will be assumed to be multi-line, and
 will also allow newlines to be matched with ".".)

 If your URL requires the use of HTTP basic authentication this is
 supported by adding a username and password parameter to your test,
 for example:

   https://jigsaw.w3.org/HTTP/Basic/ must run http with username 'guest' with password 'guest' with content "Your browser made it"

 If you need to disable failures due to expired, broken, or
 otherwise bogus SSL certificates you can do so via the tls setting:

   https://expired.badssl.com/ must run http with tls insecure

 By default tests will fail if you're probing an SSL-site which has
 a certificate which will expire within the next 14 days.  To change
 the time-period specify it explicitly like so, if not stated the
 expiration period is assumed to be days:

   # seven days
   https://steve.fi/ must run http with expiration 7d

   # 12 hours (!)
   https://steve.fi/ must run http with expiration 12h

 To disable the SSL-expiration checking entirely specify "any":

   https://steve.fi/ must run http with expiration any

 Finally if you submit a "data" argument, like in this next example
 the request made will be a HTTP POST:

   https://steve.fi/Security/XSS/Tutorial/filter.cgi must run http with data "text=test%20me" with content "test me"

 But note that you can override the HTTP-verb via:

    https://example.com/ must run http with method HEAD

 Combining these you can submit data with a PUT method:

    https://steve.fi/Security/XSS/Tutorial/filter.cgi must run http with method PUT with data "text=test%20me" with content "test me"

 Do note that the HTTP-probe never follow redirections, to allow enhanced
 testing.
`
	return str
}

// RunTest is the part of our API which is invoked to actually execute a
// HTTP-test against the given URL.
//
// For the purposes of clarity this test makes a HTTP-fetch.  The `test.Test`
// structure contains our raw test, and the `target` variable contains the
// IP address against which to make the request.
//
// So:
//
//    tst.Target => "https://steve.kemp.fi/
//
//    target => "176.9.183.100"
//
func (s *HTTPTest) RunTest(tst test.Test, target string, opts test.Options) error {

	//
	// Determine the port to connect to, initially via the protocol
	// in the string, but allow the URI to override that.
	//
	// e.g: We expect
	//
	//  http://example.com/      -> 80
	//  https://example.com/     -> 443
	//  http://example.com:8080/ -> 8080
	//
	port := "80"
	u, err := url.Parse(tst.Target)
	if err != nil {
		return err
	}
	if u.Scheme == "http" {
		port = "80"
	}
	if u.Scheme == "https" {
		port = "443"
	}
	if u.Port() != "" {
		port = u.Port()
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
		addr = fmt.Sprintf("%s:%s", address, port)

		//
		// If we find a ":" we know it is an IPv6 address though
		//
		if strings.Contains(address, ":") {
			addr = fmt.Sprintf("[%s]:%s", address, port)
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

	//
	// The default method is "GET"
	//
	method := "GET"

	//
	// That can be changed
	//
	if tst.Arguments["method"] != "" {
		method = tst.Arguments["method"]
	}

	//
	// If we have no data then make a GET request
	//
	if tst.Arguments["data"] == "" {
		req, err = http.NewRequest(method, target, nil)
	} else {

		//
		// Otherwise make a HTTP POST request, with
		// the specified data.
		//
		req, err = http.NewRequest(method, target,
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
	// Set a suitable user-agent
	//
	if tst.Arguments["user-agent"] != "" {
		req.Header.Set("User-Agent", tst.Arguments["user-agent"])
	} else {
		req.Header.Set("User-Agent", "overseer/probe")
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
		return fmt.Errorf("status code was %d not %d", status, ok)
	}

	//
	// Is the user looking for a literal body-match?
	//
	if tst.Arguments["content"] != "" {
		if !strings.Contains(string(body), tst.Arguments["content"]) {
			return fmt.Errorf("body didn't contain '%s'", tst.Arguments["content"])
		}
	}

	//
	// Is the user expecting a regular expression to match the content?
	//
	if tst.Arguments["pattern"] != "" {
		re, error := regexp.Compile("(?ms)" + tst.Arguments["pattern"])
		if error != nil {
			return error
		}

		// Skip unless this handler matches the filter.
		match := re.FindAllStringSubmatch(string(body), -1)
		if len(match) < 1 {
			return fmt.Errorf("body didn't match the regular expression '%s'", tst.Arguments["pattern"])
		}
	}

	//
	// If we reached here then our actual test was fine.
	//
	// However as a special extension we're going to test the
	// certificate expiration date for any SSL sites.  We'll
	// do that now.
	//
	if strings.HasPrefix(tst.Target, "https:") {

		//
		// The default expiration-time 14 days.
		//
		period := 14 * 24

		//
		// If the validity was set to `any` that means we just
		// don't care, so we don't even need to test the result.
		//
		if tst.Arguments["expiration"] == "any" {
			return nil
		}

		//
		// The user might have specified a different period
		// in hours / days.
		//
		expire := tst.Arguments["expiration"]
		if expire != "" {

			//
			// How much to scale the given figure by
			//
			// By default if no units are specified we'll
			// assume the figure is in days, so no scaling
			// is required.
			//
			mul := 1

			// Days?
			if strings.HasSuffix(expire, "d") {
				expire = strings.Replace(expire, "d", "", -1)
				mul = 24
			}

			// Hours?
			if strings.HasSuffix(expire, "h") {
				expire = strings.Replace(expire, "h", "", -1)
				mul = 1
			}

			// Get the period.
			period, err = strconv.Atoi(expire)
			if err != nil {
				return err
			}

			//
			// Multiply by our multiplier.
			//
			period *= mul
		}

		//
		// Check the expiration
		//
		hours, err := s.SSLExpiration(tst.Target, opts.Verbose)

		if err == nil {
			// Is the age too short?
			if int64(hours) < int64(period) {

				return fmt.Errorf("SSL certificate will expire in %d hours (%d days)", hours, int(hours/24))
			}
		}

	}

	//
	// If we reached here all is OK
	//
	return nil
}

// SSLExpiration returns the number of hours remaining for a given
// SSL certificate chain.
func (s *HTTPTest) SSLExpiration(host string, verbose bool) (int64, error) {

	// Expiry time, in hours
	var hours int64
	hours = -1

	//
	// If the string matches https://, then strip it off
	//
	re, err := regexp.Compile(`^https:\/\/([^\/]+)`)
	if err != nil {
		return 0, err
	}
	res := re.FindAllStringSubmatch(host, -1)
	for _, v := range res {
		host = v[1]
	}

	//
	// If no port is specified default to :443
	//
	p := strings.Index(host, ":")
	if p == -1 {
		host += ":443"
	}

	//
	// Show what we're doing.
	//
	if verbose {
		fmt.Printf("SSLExpiration testing: %s\n", host)
	}

	conn, err := tls.Dial("tcp", host, nil)
	if err != nil {
		return 0, err
	}
	defer conn.Close()

	timeNow := time.Now()
	for _, chain := range conn.ConnectionState().VerifiedChains {
		for _, cert := range chain {

			// Get the expiration time, in hours.
			expiresIn := int64(cert.NotAfter.Sub(timeNow).Hours())

			if verbose {
				fmt.Printf("SSLExpiration - certificate: %s expires in %d hours (%d days)\n", cert.Subject.CommonName, expiresIn, expiresIn/24)
			}

			// If we've not checked anything this is the benchmark
			if hours == -1 {
				hours = expiresIn
			} else {
				// Otherwise replace our result if the
				// certificate is going to expire more
				// recently than the current "winner".
				if expiresIn < hours {
					hours = expiresIn
				}
			}
		}
	}

	return hours, nil
}

// init is used to dynamically register our protocol-tester.
func init() {
	Register("http", func() ProtocolTest {
		return &HTTPTest{}
	})
}
