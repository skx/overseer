package protocols

import (
	"testing"
)

// Test parsing things that should return no options
func TestNoOptions(t *testing.T) {

	tests := []string{
		"",
		"this is just a string",
		"foo must run http",
		"bar must run ftp",
		"rsync must run ftp"}

	for _, input := range tests {
		//
		//
		out := ParseArguments(input)
		if len(out) != 0 {
			t.Errorf("Surprising output")
		}
	}
}

// Test parsing the some standard HTTP options
func TestHTTPOptions(t *testing.T) {

	tests := []string{
		"http://example.com/ must run http with content 'moi' and ..",
		"http://example.com/ must run http with content moi",
		"http://example.com/ must run http with status '200'",
		"http://example.com/ must run http with status 200",
	}

	for _, input := range tests {
		out := ParseArguments(input)
		if len(out) != 1 {
			t.Errorf("Surprising output - we expected 1 option but found %d", len(out))
		}
	}
}

// Test quotation-removal
func TestQuoteRemoval(t *testing.T) {

	tests := []string{
		"must run http with content 'moi' and ..",
		"must run http with content \"moi\"",
		"must run http with content moi",
	}

	for _, input := range tests {
		out := ParseArguments(input)

		// We expect one parameter: content
		if len(out) != 1 {
			t.Errorf("Surprising output - we expected 1 option but found %d", len(out))
		}

		// The value should be 'moi'
		if out["content"] != "moi" {
			t.Errorf("We expected the key 'content' to have the value 'moi', but found %s", out["content"])
		}
	}
}

// Test quotation-removal doesn't modify the content of a string
func TestQuoteRemovalSanity(t *testing.T) {

	tests := []string{
		"must run http with content 'm\"'oi' and ..",
		"must run http with content \"m\"'oi\"",
		"must run http with content m\"'oi",
	}

	for _, input := range tests {
		out := ParseArguments(input)

		// We expect one parameter: content
		if len(out) != 1 {
			t.Errorf("Surprising output - we expected 1 option but found %d", len(out))
		}

		// The value should have a single quote and double-quote
		single := 0
		double := 0
		for _, c := range out["content"] {
			if c == '"' {
				double += 1
			}
			if c == '\'' {
				single += 1
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

func TestReal(t *testing.T) {
	in := "http://steve.fi/ must run http with status 301 with content 'Steve Kemp'"
	out := ParseArguments(in)

	// We expect two parameter: content + status
	if len(out) != 2 {
		t.Errorf("Received the wrong number of parameters")
	}
	if out["status"] != "301" {
		t.Errorf("Failed to get the correct status-value")
	}
	if out["content"] != "Steve Kemp" {
		t.Errorf("Failed to get the correct content-value")
	}

}
