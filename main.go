//
// Parse the named configuration-files, which are assumed to contain
// test-cases, then execute each in turn.
//

package main

import (
	"bufio"
	"bytes"
	"crypto/sha1"
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"net"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"
)

//
// ConfigOptions is the globally visible structure which is designed to
// hold our configuration-options - as set by the command-line flags.
//
// It is perhaps poor practice to do things this way, but it eases coding.
//
var ConfigOptions struct {
	Purppura string
	Timeout  int
	Verbose  bool
	Version  bool
	IPv4     bool
	IPv6     bool
}

var (
	//
	// Our version number.
	//
	version = "master/latest"

	//
	// The timeout period to use for EVERY protocol-type.
	//
	TIMEOUT = time.Second * 10

	//
	// Macro-targets
	//
	MACROS = make(map[string][]string)
)

//
// Regardless of whether a test fails/passes we must pass the result
// on to purppura.
//
func postPurple(test_type string, test_target string, input string, result error) {

	//
	// If we don't have a server configured then
	// return without sending
	//
	if ConfigOptions.Purppura == "" {
		return
	}

	//
	// We need a stable ID for each test - get one by hashing the
	// complete input-line
	//
	hasher := sha1.New()
	hasher.Write([]byte(input))
	hash := base64.URLEncoding.EncodeToString(hasher.Sum(nil))

	//
	// All alerts will have an ID + Subject field.
	//
	values := map[string]string{
		"id":      hash,
		"subject": test_target,
	}

	//
	// If the test failed we'll set the detail + trigger a raise
	//
	if result != nil {
		values["detail"] =
			fmt.Sprintf("<p>The <code>%s</code> test against <code>%s</code> failed:</p><p><pre>%s</pre></p>",
				test_type, test_target, result.Error())
		values["raise"] = "now"
	} else {
		//
		// Otherwise the test passed and so all is OK
		//
		values["detail"] =
			fmt.Sprintf("<p>The <code>%s</code> test against <code>%s</code> passed.</p>",
				test_type, test_target)
		values["raise"] = "clear"
	}

	//
	// Export the fields to json to post.
	//
	jsonValue, _ := json.Marshal(values)

	//
	// Post to purppura
	//
	_, err := http.Post(ConfigOptions.Purppura,
		"application/json",
		bytes.NewBuffer(jsonValue))

	if err != nil {
		fmt.Printf("failed to submit test to purppura\n")
	}
}

//
// Process a single line from our configuration-file.
//
func processLine(input string) {

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
		if MACROS[name] != nil {
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
			MACROS[name] = append(MACROS[name], strings.TrimSpace(ent))
		}
		return
	}

	//
	// Look to see if this line matches the testing line
	//
	re := regexp.MustCompile("^([^ \t]+)\\s+must\\s+run\\s+([a-z]+)")
	out := re.FindStringSubmatch(input)

	//
	// If it didn't then we have a malformed line
	//
	if len(out) != 3 {
		fmt.Printf("WARNING: Unrecognized line - '%s'\n", input)
		return
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
	hosts := MACROS[test_target]
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
			new := fmt.Sprintf("%s %s\n", i, line[2])

			//
			// Call ourselves to run the test.
			//
			processLine(new)
		}

		//
		// We've called ourself (processLine) with the updated
		// line for each host in the macro-definition.
		//
		// So we can return here.
		//
		return
	}

	//
	// Look for a suitable protocol handler
	//
	tmp := ProtocolHandler(test_type)
	if tmp == nil {
		fmt.Printf("WARNING: Unknown protocol handler '%s'\n", test_type)
		return
	}

	//
	// Pass the full input-line to our protocol tester
	// to allow any extra options/flags to be parsed
	//
	tmp.setLine(input)

	//
	// Each test will be executed for each address-family, unless it is
	// a HTTP-test.
	//
	var targets []string

	//
	// If this is a http-test then just add our existing target
	//
	if strings.HasPrefix(test_target, "http://") {
		targets = append(targets, test_target)
	} else {

		//
		// Otherwise resolve the target as much
		// as we can.  This will mean that an SSH-test, for example,
		// will be carried out for each address-family which is
		// present in DNS.
		//
		ips, err := net.LookupIP(test_target)
		if err != nil {
			fmt.Printf("WARNING: Failed to resolve %s\n", test_target)
			return
		}
		for _, ip := range ips {
			if ip.To4() != nil {
				if ConfigOptions.IPv4 {
					targets = append(targets, fmt.Sprintf("%s", ip))
				}
			}
			if ip.To16() != nil && ip.To4() == nil {
				if ConfigOptions.IPv6 {
					targets = append(targets, fmt.Sprintf("%s", ip))
				}
			}
		}
	}

	//
	// Now for each target, run the test.
	//
	for _, target := range targets {

		//
		// Show what we're doing.
		//
		if ConfigOptions.Verbose {
			fmt.Printf("Running %s test against %s with address %s\n", test_type, test_target, target)
		}

		//
		// Run the test.
		//
		result := tmp.runTest(target)

		if result == nil {
			if ConfigOptions.Verbose {
				fmt.Printf("Test passed!\n")
			}

		} else {
			if ConfigOptions.Verbose {
				fmt.Printf("Test failed; %s\n", result.Error())
			}
		}

		//
		// Post the result to purple
		//
		postPurple(test_type, test_target, input, result)
	}
}

//
// Open the given configuration file, and parse it line-by-line.
//
func processFile(path string) {

	//
	// Open the given file.
	//
	file, err := os.Open(path)
	if err != nil {
		fmt.Printf("Error opening %s - %s\n", path, err.Error())
		return
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
			processLine(line)
		}
	}

	//
	// Was there an error with the scanner?  If so catch it
	// here.  To be honest I'm not sure if anything needs to
	// happen here
	//
	if err := scanner.Err(); err != nil {
		panic(err)
	}
}

//
// Open the named configuration file, and parse it
//
func main() {

	//
	// Our command-line options
	//
	flag.BoolVar(&ConfigOptions.IPv4, "4", true, "Should we run IPv4 tests?")
	flag.BoolVar(&ConfigOptions.IPv6, "6", false, "Should we run IPv6 tests?")
	flag.BoolVar(&ConfigOptions.Verbose, "verbose", true, "Should we be verbose?")
	flag.BoolVar(&ConfigOptions.Version, "version", false, "Show our version and exit.")
	flag.IntVar(&ConfigOptions.Timeout, "timeout", 0, "Set a timeout period, in seconds, for all tests.")
	flag.StringVar(&ConfigOptions.Purppura, "purppura", "", "Specify the purppura-endpoint.")
	flag.Parse()

	//
	// If we're to show our version, do so.
	//
	if ConfigOptions.Version {
		fmt.Printf("overseer %s\n", version)
		os.Exit(0)
	}

	//
	// If we got a timeout value then update our global timeout
	// period to be that number of seconds.
	//
	if ConfigOptions.Timeout != 0 {
		TIMEOUT = time.Second * time.Duration(ConfigOptions.Timeout)
	}

	//
	// Otherwise ensure we have at least one configuration-file.
	//

	if len(flag.Args()) < 1 {
		fmt.Printf("Usage %s file1 file2 .. fileN\n", os.Args[0])
		os.Exit(1)
	}

	//
	// Process each named file as a configuration file.
	//
	for _, file := range flag.Args() {
		processFile(file)
	}
}
