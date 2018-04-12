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
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"
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
	// Setup an explicit timeout
	//
	var netClient = &http.Client{
		Timeout: time.Second * 10,

		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	//
	// Make the request and get a response.
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
	if ok != status {
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
