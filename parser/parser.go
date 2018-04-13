package parser

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"regexp"
	"strings"
)

// Parser holds our parser-state.
type Parser struct {
	Filename string
	MACROS   map[string][]string
}

// A single test definition
type Test struct {
	Target string
	Type   string
	Input  string
}

// A function that can be invoked
type ParsedTest func(x Test) error

// New is the constructor.
func New(filename string) *Parser {
	m := new(Parser)
	m.Filename = filename
	m.MACROS = make(map[string][]string)
	return m
}

// Parse processes the file passed in the constructor,
// for each line ParseLine is invoked
func (s *Parser) Parse(cb ParsedTest) error {

	// Open the given file.
	file, err := os.Open(s.Filename)
	if err != nil {
		return errors.New(fmt.Sprintf("Error opening %s - %s\n", s.Filename, err.Error()))
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
			err = s.parseLine(line, cb)
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
func (s *Parser) parseLine(input string, cb ParsedTest) error {

	//
	// Ensure that we have a callback.
	//
	if cb == nil {
		return errors.New( "nil callback submitted to parseLine")
	}

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
			fmt.Printf("Redefinining a macro is a fatal error!\n")
			fmt.Printf("A macro named '%s' already exists.\n", name)
			os.Exit(1)
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
		return nil
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
		return errors.New(fmt.Sprintf("WARNING: Unrecognized line - '%s'\n", input))
	}

	//
	// Save the type + target away
	//
	test_target := out[1]
	test_type := out[2]

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
			s.parseLine(new, cb)
		}

		//
		// We've called ourself (processLine) with the updated
		// line for each host in the macro-definition.
		//
		// So we can return here.
		//
		return nil
	}

	//
	// Create a temporary structure to hold our test
	//
	var tmp Test
	tmp.Target = test_target
	tmp.Type = test_type
	tmp.Input = input

	//
	// Invoke the user-supplied callback on this parsed test.
	//
	cb(tmp)
	return nil
}
