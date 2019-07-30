// Package test contains details about a single parsed test which should be
// executed against a remote host.
//
// Tests are parsed via the parser-module, and have the general form:
//
//    HOST must run PROTOCOL with ARG_NAME1 ARG_VALUE1 ..
//
// For example a simple test might read:
//
//      1.2.3.4 must run ftp
//
// To change the port from the default the `port` argument could be
// given:
//
//      1.2.3.4 must run ftp with port 2121
//
//
package test

import (
	"fmt"
	"sort"
	"time"
)

// Test contains a single test definition as identified by the parser.
type Test struct {
	// Target of the test.
	//
	// In the example above this would be `1.2.3.4`.
	Target string

	// Type contains the type of the test.
	//
	// In the example above this would be `ftp`.
	Type string

	// Input contains a copy of the complete input-line the parser case.
	//
	// In the example above this would be `1.2.3.4 must run ftp`.
	Input string

	// MaxRetries overrides the global overseer setting for max test retries, if >= 0
	MaxRetries int

	// Arguments contains a map of any optional arguments supplied to
	// test test.
	//
	// In the example above the map would contain one key `port`,
	// with the value `2121` (as a string).
	//
	Arguments map[string]string
}

// Sanitize returns a copy of the input string, but with any password
// removed
func (obj *Test) Sanitize() string {

	// The basic test
	res := fmt.Sprintf("%s must run %s", obj.Target, obj.Type)

	// Arguments, sorted
	var keys []string
	for k := range obj.Arguments {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// Now append the arguments and their values.
	for _, k := range keys {
		tmp := ""

		// Censor passwords
		if k == "password" {
			tmp = " with password 'CENSORED'"
		} else {

			// Otherwise leave alone.
			tmp = fmt.Sprintf(" with %s '%s'", k, obj.Arguments[k])
		}
		res += tmp
	}

	return res
}

// Options are options which are passed to every test-handler.
//
// The options might change the way the test operates.
type Options struct {
	// Timeout for the single test, in seconds.
	Timeout time.Duration

	// Should the protocol-tests run verbosely?
	Verbose bool
}
