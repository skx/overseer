package test

// A single test definition as identified by the parser.
type Test struct {
	// The target of the test
	Target string

	// The type of the test.
	Type string

	// The complete line of input the parser saw
	Input string

	// Any optional arguments supplied to the parser.
	Arguments map[string]string
}
