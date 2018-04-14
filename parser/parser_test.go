//
//  Basic testing of our DB primitives
//

package parser

import (
	"io/ioutil"
	"os"
	"strings"
	"testing"
)

// Test that parsing a missing file returns an error
func TestMissingFile(t *testing.T) {

	p := New()
	err := p.ParseFile("/path/is/not/found", nil)

	if err == nil {
		t.Errorf("Parsing a missing file didn't raise an error!")
	}
}

// Test reading samples from a file
func TestFile(t *testing.T) {
	file, err := ioutil.TempFile(os.TempDir(), "prefix")
	defer os.Remove(file.Name())

	// Write to the file
	lines := `
http://example.com/ must run http
# This is fine
http://example.com/ must run http with content 'moi'

# The content-type here will not match
http://example.com/ must run http with content "moi"
`
	//
	err = ioutil.WriteFile(file.Name(), []byte(lines), 0644)
	if err != nil {
		t.Errorf("Error writing our test-case")
	}

	//
	// Now parse the file
	//
	p := New()
	err = p.ParseFile(file.Name(), nil)

	if err != nil {
		t.Errorf("Error parsing our valid file")
	}
}

// Test reading macro-based samples from a file
func TestFileMacro(t *testing.T) {
	file, err := ioutil.TempFile(os.TempDir(), "prefix")
	defer os.Remove(file.Name())

	// Write to the file
	lines := `
FOO are host1.example.com, host2.example.com
FOO must run ssh
`
	//
	err = ioutil.WriteFile(file.Name(), []byte(lines), 0644)
	if err != nil {
		t.Errorf("Error writing our test-case")
	}

	//
	// Now parse the file
	//
	p := New()
	err = p.ParseFile(file.Name(), nil)

	if err != nil {
		t.Errorf("Error parsing our valid file")
	}
}

// Test redefinining macros is a bug.
func TestFileMacroRedefined(t *testing.T) {
	file, err := ioutil.TempFile(os.TempDir(), "prefix")
	defer os.Remove(file.Name())

	// Write to the file
	lines := `
FOO are host1.example.com, host2.example.com
FOO must run ssh
FOO are host3.example.com, host4.example.com
FOO must run ftp
`
	//
	err = ioutil.WriteFile(file.Name(), []byte(lines), 0644)
	if err != nil {
		t.Errorf("Error writing our test-case")
	}

	//
	// Now parse the file
	//
	p := New()
	err = p.ParseFile(file.Name(), nil)

	if err == nil {
		t.Errorf("Expected error parsing file, didn't see one!")
	}
	if !strings.Contains(err.Error(), "Redeclaring an existing macro") {
		t.Errorf("The expected error differed from what we received")
	}
}

// Test some valid input
func TestValidLines(t *testing.T) {

	var inputs = []string{
		"foo must run http",
		"bar must run http",
		"baz must run ftp"}

	for _, line := range inputs {

		p := New()
		_, err := p.ParseLine(line, nil)

		if err != nil {
			t.Errorf("Found error parsing valid line: %s\n", err.Error())
		}
	}
}

// Test some malformed lines
func TestUnknownInput(t *testing.T) {

	var inputs = []string{
		"foo must RAN blah",
		"bar mustn't exist",
		"baz must ping"}

	for _, line := range inputs {

		p := New()
		_, err := p.ParseLine(line, nil)

		if err == nil {
			t.Errorf("Should have found error parsing line: %s\n", err.Error())
		}
		if !strings.Contains(err.Error(), "Unrecognized line") {
			t.Errorf("Received unexpected error: %s\n", err.Error())
		}
	}
}

// Test some invalid inputs
func TestUnknownProtocols(t *testing.T) {

	var inputs = []string{
		"foo must run blah",
		"bar must run moi",
		"baz must run kiss"}

	for _, line := range inputs {

		p := New()
		_, err := p.ParseLine(line, nil)

		if err == nil {
			t.Errorf("Should have found error parsing line: %s\n", err.Error())
		}
		if !strings.Contains(err.Error(), "Unknown test-type") {
			t.Errorf("Received unexpected error: %s\n", err.Error())
		}
	}
}
