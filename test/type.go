// Test is the package which contains details about a single parsed
// test which should be executed against a remote host.
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

	// Arguments contains a map of any optional arguments supplied to
	// test test.
	//
	// In the example above the map would contain one key `port`,
	// with the value `2121` (as a string).
	//
	Arguments map[string]string
}

// TestOptions are options which are passed to every test-handler.
//
// The options might change the way the test operates.
type TestOptions struct {
	// Retry controls whether failing tests should be retried.
	Retry bool

	// Timeout for the single test, in seconds.
	Timeout time.Duration

	// Verbose controls the level of disagnosting printing the
	// tests & drivers produce.
	Verbose bool

	// Should the test probe IPv4 addresses?
	IPv4 bool

	// Should the test probe IPv6 addresses?
	IPv6 bool
}
