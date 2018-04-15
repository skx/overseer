//
// This is our HTTP/HTTPS protocol-test.
//
// It allows fetching remote URLs and testing the status-code and body
// response.
//
// NOTE: This deliberately does not follow redirections, to allow enhanced
// testing.
//
package protocols

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/skx/overseer/test"
)

//
// Our structure.
//
type HTTPTest struct {
}

//
// Make a HTTP-test against the given URL.
//
func (s *HTTPTest) RunTest(tst test.Test, target string, opts TestOptions) error {

	//
	// We want to turn the target URL into an IPv4 and IPv6
	// address so that we can test each of them.
	//
	var ipv4 []string
	var ipv6 []string

	//
	// Find the hostname we should connect to.
	//
	u, err := url.Parse(target)
	if err != nil {
		return nil
	}

	//
	// The port we connect to, on that host
	//
	port := 80
	if u.Scheme == "https" {
		port = 443
	}

	//
	// Lookup the IP addresses of the host.
	//
	ips, err := net.LookupIP(u.Host)
	if err != nil {
		return errors.New(fmt.Sprintf("Failed to resolve %s\n", u.Host))
	}

	//
	// Process each of the resolved results
	//
	for _, ip := range ips {

		//
		// IPv4 address
		//
		if ip.To4() != nil {
			ipv4 = append(ipv4, fmt.Sprintf("%s:%d", ip, port))
		}

		//
		// IPv6 address
		//
		if ip.To16() != nil && ip.To4() == nil {
			ipv6 = append(ipv6, fmt.Sprintf("[%s]:%d", ip, port))
		}
	}

	//
	// Now we're going to run the testing
	//

	//
	// IPv4 only?
	//
	if (opts.IPv4 == true) && (opts.IPv6 == false) {
		if opts.Verbose {
			fmt.Printf("\tIPv4-only testing enabled for HTTP\n")
		}

		if len(ipv4) > 0 {
			err := s.RunHTTPTest(target, ipv4[0], tst, opts)
			if opts.Verbose {
				fmt.Printf("\tRunning against %s\n", ipv4[0])
			}
			return err
		} else {
			return errors.New(fmt.Sprintf("Failed to resolve %s to IPv4 address", target))
		}
	}

	//
	// IPv6 only?
	//
	if (opts.IPv6 == true) && (opts.IPv4 == false) {
		if opts.Verbose {
			fmt.Printf("\tIPv6-only testing enabled for HTTP\n")
		}

		if len(ipv6) > 0 {
			if opts.Verbose {
				fmt.Printf("\tRunning against %s\n", ipv6[0])
			}
			err := s.RunHTTPTest(target, ipv6[0], tst, opts)
			return err
		} else {
			return errors.New(fmt.Sprintf("Failed to resolve %s to IPv6 address", target))
		}
	}

	//
	// Both?
	//
	if (opts.IPv6 == true) && (opts.IPv4 == true) {
		if opts.Verbose {
			fmt.Printf("\tIPv4 & IPv6 testing enabled for HTTP\n")
		}

		executed := 0

		// ipv4
		if len(ipv4) > 0 {
			if opts.Verbose {
				fmt.Printf("\tRunning against %s\n", ipv4[0])
			}
			err := s.RunHTTPTest(target, ipv4[0], tst, opts)
			if err != nil {
				return err
			}
			executed += 1
		}

		// ipv6
		if len(ipv6) > 0 {
			if opts.Verbose {
				fmt.Printf("\tRunning against %s\n", ipv6[0])
			}
			err := s.RunHTTPTest(target, ipv6[0], tst, opts)
			if err != nil {
				return err
			}
			executed += 1
		}
		if executed < 1 {
			return errors.New(fmt.Sprintf("Failed to perform HTTP test of target %s", target))
		}
	}

	if (opts.IPv6 == false) && (opts.IPv4 == false) {
		return errors.New("IPv4 AND IPv6 are disabled, no HTTP test is possible.")
	}
	return nil
}

func (s *HTTPTest) RunHTTPTest(target string, address string, tst test.Test, opts TestOptions) error {

	//
	// Setup a dialer which will be dual-stack
	//
	dialer := &net.Dialer{
		DualStack: true,
	}

	//
	// Magic happens.
	//
	dial := func(ctx context.Context, network, addr string) (net.Conn, error) {
		addr = address
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
	// Now we can make the request and get a response.
	//
	response, err := netClient.Get(target)
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
