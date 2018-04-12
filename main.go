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
}

var (
	//
	// Our version number.
	//
	version = "master/latest"

	//
	// The default timeout period
	//
	TIMEOUT = time.Second * 10
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
	// Our tests are all of the form:
	//
	//  TARGET must run PROTOCOL [OPTIONA EXTRA ARGS]
	//
	// Look to see if this line matches
	//
	re := regexp.MustCompile("^([^ \t]+)\\s+must\\s+run\\s+([a-z]+)")
	out := re.FindStringSubmatch(input)

	//
	// If it didn't then we have a malformed line
	//
	if len(out) != 3 {
		return
	}

	//
	// Save the type + target away
	//
	test_target := out[1]
	test_type := out[2]

	//
	// Look for a suitable protocol handler
	//
	tmp := ProtocolHandler(test_type)
	if tmp == nil {
		fmt.Printf("Uknown protocol handler invoked '%s'\n", test_type)
		return
	}

	//
	// Pass the full input-line to our protocol tester
	// to allow any extra options/flags to be parsed
	//
	tmp.setLine(input)

	//
	// Run the damn test :)
	//
	if ConfigOptions.Verbose {
		fmt.Printf("Running %s test against %s\n", test_type, test_target)
	}

	result := tmp.runTest(test_target)

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
