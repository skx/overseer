package parser

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/skx/overseer/protocols"
	"github.com/skx/overseer/test"
)

// Parser holds our parser-state.
type Parser struct {
	MACROS map[string][]string
}

// A function that can be invoked
type ParsedTest func(x test.Test) error

// New is the constructor.
func New() *Parser {
	m := new(Parser)
	m.MACROS = make(map[string][]string)
	return m
}

// Parse processes the file passed in the constructor,
// for each line ParseLine is invoked
func (s *Parser) ParseFile(filename string, cb ParsedTest) error {

	// Open the given file.
	file, err := os.Open(filename)
	if err != nil {
		return errors.New(fmt.Sprintf("Error opening %s - %s\n", filename, err.Error()))
	}
	defer file.Close()

	//
	// We'll process it line by line
	//
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {

		//
		// Get a line and trim leading/trailing whitespace
		//
		line := scanner.Text()
		line = strings.TrimSpace(line)

		//
		// If the line wasn't empty, and didn't start with
		// a comment then process it.
		//
		if (line != "") && (!strings.HasPrefix(line, "#")) {
			_, err = s.ParseLine(line, cb)
			if err != nil {
				return err
			}
		}
	}

	//
	// Was there an error with the scanner?  If so catch it
	// here.  To be honest I'm not sure if anything needs to
	// happen here
	//
	if err := scanner.Err(); err != nil {
		return err
	}

	// No error
	return nil
}

// parseLine parses a single line, and returns an error if
// one was found.
func (s *Parser) ParseLine(input string, cb ParsedTest) (test.Test, error) {

	//
	// The result for the caller
	//
	var result test.Test

	//
	// Our input will contain lines of two forms:
	//
	//  MACRO are host1, host2, host3
	//
	// NOTE: Macro-names are UPPERCASE, and redefinining a macro
	//       is an error - because it would be too confusing otherwise.
	//
	//
	//  TARGET must run PROTOCOL [OPTIONAL EXTRA ARGS]
	//

	//
	// Is this a macro-definition?
	//
	macro := regexp.MustCompile("^([A-Z0-9]+)\\s+are\\s+(.*)$")
	match := macro.FindStringSubmatch(input)
	if len(match) == 3 {

		name := match[1]
		vals := match[2]

		//
		// If this macro-exists that is a fatal error
		//
		if s.MACROS[name] != nil {
			return result, errors.New(fmt.Sprintf("Redeclaring an existing macro is a fatal-error, %s exists already.\n", name))
		}

		//
		// The macro-value is a comma-separated list of hosts
		//
		hosts := strings.Split(vals, ",")

		//
		// Save each host away, under the name of the macro.
		//
		for _, ent := range hosts {
			s.MACROS[name] = append(s.MACROS[name], strings.TrimSpace(ent))
		}
		return result, nil
	}

	//
	// Look to see if this line matches the testing line
	//
	re := regexp.MustCompile("^([^ \t]+)\\s+must\\s+run\\s+([^\\s]+)")
	out := re.FindStringSubmatch(input)

	//
	// If it didn't then we have a malformed line
	//
	if len(out) != 3 {
		return result, errors.New(fmt.Sprintf("WARNING: Unrecognized line - '%s'\n", input))
	}

	//
	// Save the type + target away
	//
	test_target := out[1]
	test_type := out[2]

	handler := protocols.ProtocolHandler(test_type)
	if handler == nil {
		return result, errors.New(fmt.Sprintf("Unknown test-type '%s' in input '%s'", test_type, input))
	}

	//
	// Is this target a macro?
	//
	// If so we expand for each host in the macro-definition and
	// execute those expanded versions in turn.
	//
	hosts := s.MACROS[test_target]
	if len(hosts) > 0 {

		//
		// So we have a bunch of hosts that this macro-name
		// should be replaced with.
		//
		for _, i := range hosts {

			//
			// Reparse the line for each host by taking advantage
			// of the fact the first entry in the line is the
			// target.
			//
			// So we change:
			//
			//  HOSTS must run xxx..
			//
			// Into:
			//
			//   host1 must run xxx.
			//   host2 must run xxx.
			//   ..
			//   hostN must run xxx.
			//
			split := regexp.MustCompile("^([^\\s]+)\\s+(.*)$")
			line := split.FindStringSubmatch(input)

			//
			// Create a new test, with the macro-host
			// in-place of the original target.
			//
			new := fmt.Sprintf("%s %s", i, line[2])

			//
			// Call ourselves to run the test.
			//
			s.ParseLine(new, cb)
		}

		//
		// We've called ourself (processLine) with the updated
		// line for each host in the macro-definition.
		//
		// So we can return here.
		//
		return result, nil
	}

	//
	// Create a temporary structure to hold our test
	//
	result.Target = test_target
	result.Type = test_type
	result.Input = input
	result.Arguments = s.ParseArguments(input)

	//
	// Invoke the user-supplied callback on this parsed test.
	//
	//
	// Ensure that we have a callback.
	//
	if cb != nil {
		cb(result)
	}

	return result, nil
}

// TrimQuotes removes the matching quotes from around a string, if they
// are present.
//
// For example 'steve' becomes steve, but 'steve stays unchanged, as there
// are not matching single-quotes around the string.
func (s *Parser) TrimQuotes(in string, c byte) string {
	if len(in) >= 2 {
		if in[0] == c && in[len(in)-1] == c {
			return in[1 : len(in)-1]
		}
	}
	return in
}

// ParseArguments takes a string such as this:
//
//   foo must run http with username 'steve' with password 'bob'
//
// And extracts the values of the named options.
//
// Any option that is wrapped in matching quotes has them removed.
//
func (s *Parser) ParseArguments(input string) map[string]string {
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
		value = s.TrimQuotes(value, '\'')
		value = s.TrimQuotes(value, '"')

		// Store the value in our map
		res[name] = value

		// Continue matching the tail of the string.
		input = prefix
		match = expr.FindStringSubmatch(input)
	}
	return res
}
