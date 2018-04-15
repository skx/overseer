// The test package holds the structure of a successfully
// parsed test.
//
// Tests are parsed via the parser-module, and have several
// fields which are documented.
//
// For example a test might read:
//
//      1.2.3.4 must run ftp with port 21
//

package test

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
