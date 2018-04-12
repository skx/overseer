//
// This is our HTTP/HTTPS protocol-test.
//
// It allows fetching remote URLs and testing the status-code and body
// response.
//
// NOTE: This deliberately does not follow redirections, to allow enhanced
// testing.
//
package main

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"regexp"
	"strconv"
	"strings"
)

//
// Our structure.
//
// We store state in the `input` field.
//
type HTTPTest struct {
	input string
}

//
// Make a HTTP-test against the given URL.
//
func (s *HTTPTest) runTest(target string) error {

	//
	// We want to turn the target URL into an IPv4 and IPv6
	// address so that we can test each of them.
	//
	var ipv4 []string
	var ipv6 []string

	//
	// Find the hostname we should connect to, by parsing
	// the URL with a regular expression.
	//
	rehost := regexp.MustCompile("^(https?)://([^/]+)")
	match := rehost.FindStringSubmatch(target)
	if len(match) != 0 {

		//
		// Protocol + Host
		//
		proto := match[1]
		host := match[2]

		//
		// Lookup the IP addresses of the host.
		//
		ips, err := net.LookupIP(host)
		if err != nil {
			return errors.New(fmt.Sprintf("WARNING: Failed to resolve %s\n", host))
		}

		//
		// Process each of the resolved results
		//
		for _, ip := range ips {

			//
			// IPv4 address
			//
			if ip.To4() != nil {
				if proto == "http" {
					ipv4 = append(ipv4, fmt.Sprintf("%s:%d", ip, 80))
				} else {
					ipv4 = append(ipv4, fmt.Sprintf("%s:%d", ip, 443))
				}

			}

			//
			// IPv6 address
			//
			if ip.To16() != nil && ip.To4() == nil {
				if proto == "http" {
					ipv6 = append(ipv6, fmt.Sprintf("[%s]:%d", ip, 80))
				} else {
					ipv6 = append(ipv6, fmt.Sprintf("[%s]:%d", ip, 443))
				}
			}
		}
	}

	//
	// Now we're going to run the testing
	//

	//
	// IPv4 only?
	//
	if (ConfigOptions.IPv4 == true) && (ConfigOptions.IPv6 == false) {
		if len(ipv4) > 0 {
			err := s.RunHTTPTest(target, ipv4[0])
			return err
		} else {
			return errors.New(fmt.Sprintf("Failed to resolve %s to IPv4 address", target))
		}
	}

	//
	// IPv6 only?
	//
	if (ConfigOptions.IPv6 == true) && (ConfigOptions.IPv4 == false) {
		if len(ipv6) > 0 {
			err := s.RunHTTPTest(target, ipv6[0])
			return err
		} else {
			return errors.New(fmt.Sprintf("Failed to resolve %s to IPv6 address", target))
		}
	}

	//
	// Both?
	//
	if (ConfigOptions.IPv6 == true) && (ConfigOptions.IPv4 == true) {
		executed := 0

		// ipv4
		if len(ipv4) > 0 {
			err := s.RunHTTPTest(target, ipv4[0])
			if err != nil {
				return err
			}
			executed += 1
		}

		// ipv6
		if len(ipv6) > 0 {
			err := s.RunHTTPTest(target, ipv6[0])
			if err != nil {
				return err
			}
			executed += 1
		}

		if executed < 1 {
			return errors.New(fmt.Sprintf("Failed to perform HTTP test of target %s", target))
		}
	}

	if (ConfigOptions.IPv6 == false) && (ConfigOptions.IPv4 == false) {
		return errors.New("IPv4 AND IPv6 are disabled, no HTTP test is possible.")
	}
	return nil
}

func (s *HTTPTest) RunHTTPTest(target string, address string) error {

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
		Timeout: TIMEOUT,

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
	// If the user specified a different status-code update it.
	//
	re := regexp.MustCompile("with\\s+status\\s+([0-9]+)")
	out := re.FindStringSubmatch(s.input)
	if len(out) == 2 {
		ok, err = strconv.Atoi(out[1])
		if err != nil {
			return err
		}
	}

	//
	// See if the status-code matched our expectation(s).
	//
	// If there is a configured status-code of `999` that means the
	// client doesn't care what the response was.  This is useful because
	// you can find that sites present a different status-code over
	// IPv4 and IPv6 making a single test useless.
	//
	if ok != status && (ok != 999) {
		return errors.New(fmt.Sprintf("Status code was %s not %s", status, ok))
	}

	//
	// Looking for a body-match?
	//
	rebody := regexp.MustCompile("with\\s+content\\s+'([^']+)'")
	out = rebody.FindStringSubmatch(s.input)
	if len(out) == 2 {
		if !strings.Contains(string(body), out[1]) {
			return errors.New(
				fmt.Sprintf("Body didn't contain '%s'", out[1]))
		}

	}

	//
	// If we reached here all is OK
	//
	return nil
}

//
// Store the complete line from the parser in our private
// field; this could be used if there are protocol-specific
// options to be understood.
//
func (s *HTTPTest) setLine(input string) {
	s.input = input
}

//
// Register our protocol-tester.
//
func init() {
	Register("http", func() ProtocolTest {
		return &HTTPTest{}
	})
}
