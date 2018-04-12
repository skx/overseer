//
// This is a stub protocol-test which might one day make FTP-connections.
//

package main

import (
	"errors"
	"fmt"
)

//
// Our structure.
//
// We store state in the `input` field.
//
type FTPTest struct {
	input string
}

//
// Run the test against the specified target.
//
func (s *FTPTest) runTest(target string) error {
	fmt.Printf("Making ftp test against %s\n", target)
	fmt.Printf("\tInput was %s\n", s.input)
	return errors.New("Not implemented")
}

//
// Store the complete line from the parser in our private
// field; this could be used if there are protocol-specific options
// to be understood.
//
func (s *FTPTest) setLine(input string) {
	s.input = input
}

//
// Register our protocol-tester.
//
func init() {
	Register("ftp", func() ProtocolTest {
		return &FTPTest{}
	})
}
