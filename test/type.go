// The Test structure represents a single successfully parsed test.
//
// Tests are parsed via the parser-module, and have several
// fields which are documented.
//
// For example a test might read:
//
//      1.2.3.4 must run ftp with port 21
//
package test

import (
	"time"
)

// A single test definition as identified by the parser.
type Test struct {
	// The target of the test, in the example above this
	// would be `1.2.3.4`.
	Target string

	// The type of the test, in the example above this would
	// be `ftp`.
	Type string

	// The complete line of input the parser saw, in the
	// example above this would be `1.2.3.4 must run ftp with port 21`.
	Input string

	// Any optional arguments supplied to the parser.
	// In the example above the map would contain one key `port`,
	// with the value `21` (as a string).
	Arguments map[string]string
}

// TestOptions are options which are passed to every test-handler.
//
// The options might change the way the test operates.
type TestOptions struct {
	// Should failing tests be retried?
	Retry bool

	// Timeout for the single test, in seconds.
	Timeout time.Duration

	// Should the test consider itself to be running verbosely?
	Verbose bool

	// Should the test probe IPv4 addresses?
	IPv4 bool

	// Should the test probe IPv6 addresses?
	IPv6 bool
}
