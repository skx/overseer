package parser

import (
	"io/ioutil"
	"os"
	"strings"
	"testing"

	"github.com/skx/overseer/test"
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
	if err != nil {
		t.Errorf("Error creating temporary-directory %s", err.Error())
	}
	defer os.Remove(file.Name())

	// Write to the file
	lines := `
http://example.com/ must run http
# This is fine
http://example.com/ must run http with content 'moi'
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
	if err != nil {
		t.Errorf("Error creating temporary-directory %s", err.Error())
	}

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
	if err != nil {
		t.Errorf("Error creating temporary-directory %s", err.Error())
	}
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
	if !strings.Contains(err.Error(), "redeclaring an existing macro") {
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

// Test parsing things that should return no options
func TestNoArguments(t *testing.T) {

	tests := []string{
		"127.0.0.1 must run ping",
		"127.0.0.1 must run ssh",
	}

	// Create a parser
	p := New()

	// Parse each line
	for _, input := range tests {

		out, err := p.ParseLine(input, nil)
		if err != nil {
			t.Errorf("Error parsing %s - %s", input, err.Error())
		}
		if len(out.Arguments) != 0 {
			t.Errorf("Surprising output")
		}
	}
}

// Test parsing a multi-line statement
func TestContinuation(t *testing.T) {

	file, err := ioutil.TempFile(os.TempDir(), "prefix")
	if err != nil {
		t.Errorf("Error creating temporary-directory %s", err.Error())
	}
	defer os.Remove(file.Name())

	// Write to the file
	lines := `
127.0.\
0.1 \
   must   \
     run redis


`
	//
	err = ioutil.WriteFile(file.Name(), []byte(lines), 0644)
	if err != nil {
		t.Errorf("Error writing our test-case")
	}

	//
	// Count of parsed lines.
	//
	count := 0

	//
	// Now parse the file
	//
	p := New()
	err = p.ParseFile(file.Name(), func(tst test.Test) error {
		count++
		if tst.Type != "redis" {
			t.Errorf("Our parser was broken!")
		}
		if tst.Target != "127.0.0.1" {
			t.Errorf("Our parser was broken!")
		}
		if tst.Sanitize() != "127.0.0.1 must run redis" {
			t.Errorf("Our parser resulted in a mismatched result!")
		}
		return nil
	})

	if err != nil {
		t.Errorf("Expected no error, but found %s", err.Error())
	}
	if count != 1 {
		t.Errorf("Expected a single valid line, found %d", count)
	}
}

// Test parsing a continued-comment.
func TestCommentContinuation(t *testing.T) {

	file, err := ioutil.TempFile(os.TempDir(), "prefix")
	if err != nil {
		t.Errorf("Error creating temporary-directory %s", err.Error())
	}
	defer os.Remove(file.Name())

	// Write to the file
	lines := `
# This is a comment \
comment must run http \
 This is still a comment.
`
	//
	err = ioutil.WriteFile(file.Name(), []byte(lines), 0644)
	if err != nil {
		t.Errorf("Error writing our test-case")
	}

	//
	// Count of parsed lines.
	//
	count := 0

	//
	// Now parse the file
	//
	p := New()
	err = p.ParseFile(file.Name(), func(tst test.Test) error {
		count++
		return nil
	})

	if err != nil {
		t.Errorf("Expected no error, but found %s", err.Error())
	}
	if count != 0 {
		t.Errorf("Expected zero valid lines, found %d", count)
	}
}

// Test parsing an argument that fails validation
func TestInvalidArgument(t *testing.T) {

	tests := []string{
		"http://example.com/ must run http with status moi",
		"http://example.com/ must run http with expiration 12s",
	}

	// Create a parser
	p := New()

	// Parse each line
	for _, input := range tests {

		_, err := p.ParseLine(input, nil)
		if err == nil {
			t.Errorf("Expected an error parsing input, received none.  Input was %s", input)
		}
		if !strings.Contains(err.Error(), "did not match pattern") {
			t.Errorf("Received unexpected error: %s\n", err.Error())
		}

	}
}

// Test parsing some common HTTP options
func TestHTTPOptions(t *testing.T) {

	tests := []string{
		"http://example.com/ must run http with content 'moi' and ..",
		"http://example.com/ must run http with content moi",
		"http://example.com/ must run http with status '200'",
		"http://example.com/ must run http with status 200",
	}

	// Create a parser
	p := New()

	// Parse each line
	for _, input := range tests {

		// Parse the line
		out, err := p.ParseLine(input, nil)
		if err != nil {
			t.Errorf("Error parsing %s - %s", input, err.Error())
		}

		// We should have a single argument in each case
		if len(out.Arguments) != 1 {
			t.Errorf("Surprising output - we expected 1 option but found %d", len(out.Arguments))
		}
	}
}

// Test quotation-removal
func TestQuoteRemoval(t *testing.T) {

	tests := []string{
		"http://example.com/ must run http with content 'moi' and ..",
		"http://example.com/ must run http with content \"moi\"",
		"http://example.com/ must run http with content moi",
	}

	// Create a parser
	p := New()

	// Parse each line
	for _, input := range tests {

		out, err := p.ParseLine(input, nil)
		if err != nil {
			t.Errorf("Error parsing %s - %s", input, err.Error())
		}

		// We expect one parameter: content
		if len(out.Arguments) != 1 {
			t.Errorf("Surprising output - we expected 1 option but found %d", len(out.Arguments))
		}

		// The value should be 'moi'
		if out.Arguments["content"] != "moi" {
			t.Errorf("We expected the key 'content' to have the value 'moi', but found %s", out.Arguments["content"])
		}
	}
}

// Test quotation-removal doesn't modify the content of a string
func TestQuoteRemovalSanity(t *testing.T) {

	tests := []string{
		"http://example.com/ must run http with content 'm\"'oi' and ..",
		"http://example.com/ must run http with content \"m\"'oi\"",
		"http://example.com/ must run http with content m\"'oi",
	}

	// Create a parser
	p := New()

	// Parse each line
	for _, input := range tests {

		out, err := p.ParseLine(input, nil)
		if err != nil {
			t.Errorf("Error parsing %s - %s", input, err.Error())
		}

		// We expect one parameter: content
		if len(out.Arguments) != 1 {
			t.Errorf("Surprising output - we expected 1 option but found %d", len(out.Arguments))
		}

		// The value should have a single quote and double-quote
		single := 0
		double := 0
		for _, c := range out.Arguments["content"] {
			if c == '"' {
				double++
			}
			if c == '\'' {
				single++
			}
		}

		if single != 1 {
			t.Errorf("We found the wrong number of single-quotes: %d != 1", single)

		}
		if double != 1 {
			t.Errorf("We found the wrong number of double-quotes: %d != 1", double)
		}
	}
}

// Test a real line
func TestReal(t *testing.T) {
	in := "http://steve.fi/ must run http with status 301 with content 'Steve Kemp'"

	// Create a parser
	p := New()

	out, err := p.ParseLine(in, nil)
	if err != nil {
		t.Errorf("Error parsing %s - %s", in, err.Error())
	}

	// We expect two parameter: content + status
	if len(out.Arguments) != 2 {
		t.Errorf("Received the wrong number of parameters")
	}
	if out.Arguments["status"] != "301" {
		t.Errorf("Failed to get the correct status-value")
	}
	if out.Arguments["content"] != "Steve Kemp" {
		t.Errorf("Failed to get the correct content-value")
	}

}

// Test that later arguments replace earlier ones.
func TestDuplicateArguments(t *testing.T) {
	in := "http://steve.fi/ must run http with status 301 with status 302 with status any"

	// Create a parser
	p := New()

	out, err := p.ParseLine(in, nil)
	if err != nil {
		t.Errorf("Error parsing %s - %s", in, err.Error())
	}

	// We expect one parameter: status
	if len(out.Arguments) != 1 {
		t.Errorf("Received the wrong number of parameters")
	}
	if out.Arguments["status"] != "any" {
		t.Errorf("Failed to get the correct status-value")
	}
}

// Test some invalid options
func TestInvalidOptions(t *testing.T) {
	tests := []string{
		"http://example.com/ must run http with CONTENT 'moi'",
		"http://example.com/ must run http with header 'foo: bar'",
		"http://example.com/ must run http with statsu 300 ",
	}

	// Create a parser
	p := New()

	// Parse each line
	for _, input := range tests {

		_, err := p.ParseLine(input, nil)
		if err == nil {
			t.Errorf("We expected an error parsing %s, but found none!", input)
		}

		if !strings.Contains(err.Error(), "Unsupported argument") {
			t.Errorf("The error we received was the wrong error: %s", err.Error())

		}
	}
}

func TestMaxRetries(t *testing.T) {
	tests := []string{
		"http://example.com/ must run http with maxRetries 0",
		"http://example.com/ must run http with maxRetries 1",
		"http://example.com/ must run http with maxRetries 2",
	}

	// Create a parser
	p := New()

	// Parse each line
	for idx, input := range tests {

		tst, err := p.ParseLine(input, nil)
		if err != nil {
			t.Errorf("We did not expect an error parsing %s!", input)
			continue
		}

		if tst.MaxRetries != idx {
			t.Errorf("Invalid maxRetries number. Expected %d, got %d", idx, tst.MaxRetries)

		}
	}
}

// Test invoking a callback.
func TestCallback(t *testing.T) {
	file, err := ioutil.TempFile(os.TempDir(), "prefix")
	if err != nil {
		t.Errorf("Error creating temporary-directory %s", err.Error())
	}
	defer os.Remove(file.Name())

	// Content to write to a file
	lines := `
http://example.com/ must run http
# This is fine
http://example.com/ must run http with content 'moi'
`
	// Write it out
	err = ioutil.WriteFile(file.Name(), []byte(lines), 0644)
	if err != nil {
		t.Errorf("Error writing our test-case")
	}

	//
	// Count of how many times we were calledback
	//
	i := 0

	// callback

	//
	// Now parse the file - using the callback
	//
	p := New()
	err = p.ParseFile(file.Name(), func(tst test.Test) error {
		i = i + 1
		return nil
	})

	//
	// We'll test that worked.
	//
	if err != nil {
		t.Errorf("Error parsing our valid file")
	}
	if i != 2 {
		t.Errorf("Callback invoked the wrong number of times: %d", i)
	}
}

// Test sanitise
func TestSanitize(t *testing.T) {
	file, err := ioutil.TempFile(os.TempDir(), "prefix")
	if err != nil {
		t.Errorf("Error creating temporary-directory %s", err.Error())
	}
	defer os.Remove(file.Name())

	// Content to write to a file
	lines := `
http://example.com/ must run http with username 'steve' with password 'ke'mp'
`
	// Write it out
	err = ioutil.WriteFile(file.Name(), []byte(lines), 0644)
	if err != nil {
		t.Errorf("Error writing our test-case")
	}

	//
	// The test
	//
	var tmp test.Test

	//
	// Now parse the file - using the callback
	//
	p := New()
	err = p.ParseFile(file.Name(), func(tst test.Test) error {
		tmp = tst
		return nil
	})

	//
	// We'll test that worked.
	//
	if err != nil {
		t.Errorf("Error parsing our valid file")
	}

	//
	// So now we have a parsed test.Test object.
	//
	// Check the fields
	//
	if tmp.Target != "http://example.com/" {
		t.Errorf("Parsed test had wrong target!")
	}
	if tmp.Arguments["username"] != "steve" {
		t.Errorf("Parsed test had wrong username!")
	}
	if tmp.Arguments["password"] != "ke'mp" {
		t.Errorf("Parsed test had wrong password!")
	}

	//
	// Sanitize
	//
	safe := tmp.Sanitize()

	if strings.Contains(safe, "ke'mp") {
		t.Errorf("Password is still visible")
	}
	if !strings.Contains(safe, "CENSOR") {
		t.Errorf("We see no evidence of censorship")
	}
}
