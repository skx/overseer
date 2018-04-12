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
	"fmt"
	"net/http"
	"os"
	"regexp"
	"strings"
)

//
// Regardless of whether a test fails/passes we must pass the result
// on to purpurra.
//
func postPurple(test_type string, test_target string, input string, result error) {

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
		"subject": fmt.Sprintf("%s", test_target),
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
	// Post to purpurra
	//
	_, err := http.Post("http://localhost:8080/events",
		"application/json",
		bytes.NewBuffer(jsonValue))

	if err != nil {
		fmt.Printf("failed to submit test to purpurra\n")
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
		fmt.Printf("Uknonw protocol handler invoked %s\n", test_type)
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
	fmt.Printf("Running %s test against %s\n", test_type, test_target)

	result := tmp.runTest(test_target)

	if result == nil {
		fmt.Printf("Test passed!\n")

	} else {
		fmt.Printf("Test failed; %s\n", result.Error())
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
	// Ensure we have at least one configuration-file.
	//
	if len(os.Args) < 2 {
		fmt.Printf("Usage %s file1 file2 .. fileN\n", os.Args[0])
		os.Exit(1)
	}

	//
	// Process each named file as a configuration file.
	//
	for _, file := range os.Args[1:] {
		processFile(file)
	}
}
