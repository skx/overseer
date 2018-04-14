package protocols

import (
	"regexp"
)

// TrimQuotes removes the matching quotes from around a string, if they
// are present.
//
// For example 'steve' becomes steve, but 'steve stays unchanged, as there
// are not matching single-quotes around the string.
func TrimQuotes(s string, c byte) string {
	if len(s) >= 2 {
		if s[0] == c && s[len(s)-1] == c {
			return s[1 : len(s)-1]
		}
	}
	return s
}

// ParseArguments takes a string such as this:
//
//   foo must run http with username 'steve' with password 'bob'
//
// And extracts the values of the named options.
//
// Any option that is wrapped in matching quotes has them removed.
//
func ParseArguments(input string) map[string]string {
	res := make(map[string]string)

	//
	// Look for each option
	//
	expr := regexp.MustCompile("^(.*)\\s+with\\s+([^\\s]+)\\s+('.+'|\".+\"|\\S+)")
	match := expr.FindStringSubmatch(input)

	for len(match) > 1 {
		prefix := match[1]
		name := match[2]
		value := match[3]

		// Strip quotes
		value = TrimQuotes(value, '\'')
		value = TrimQuotes(value, '"')

		// Store the value in our map
		res[name] = value

		// Continue matching the tail of the string.
		input = prefix
		match = expr.FindStringSubmatch(input)
	}
	return res
}
