//
// Parse the named configuration-files, which are assumed to contain
// test-cases, then execute each in turn.
//

package main

import (
	"bytes"
	"crypto/sha1"
	"encoding/base64"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"net"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/skx/overseer/parser"
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
	// complete input-line and the target we executed against.
	//
	hasher := sha1.New()
	hasher.Write([]byte(test_target))
	hasher.Write([]byte(input))
	hash := base64.URLEncoding.EncodeToString(hasher.Sum(nil))

	//
	// All alerts will have an ID + Subject field.
	//
	values := map[string]string{
		"id":      hash,
		"subject": input,
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
// run_test is a callback invoked by the parser when a job
// has been successfully parsed
//
// It runs the test.
//
func run_test(tst parser.Test) error {

	//
	// Setup our local state.
	//
	test_type := tst.Type
	test_target := tst.Target
	input := tst.Input

	//
	// Look for a suitable protocol handler
	//
	tmp := ProtocolHandler(test_type)
	if tmp == nil {
		fmt.Printf("WARNING: Unknown protocol handler '%s'\n", test_type)
		postPurple(test_type, test_target, input, errors.New(fmt.Sprintf("Uknown protocol-handler %s", test_type)))
		return nil
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
	if strings.HasPrefix(test_target, "http") {
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

			postPurple(test_type, test_target, input, errors.New(fmt.Sprintf("Failed to resolve name %s", test_target)))
			fmt.Printf("WARNING: Failed to resolve %s\n", test_target)
			return nil
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
			fmt.Printf("Running %s test against %s (%s)\n", test_type, test_target, target)
		}

		//
		// Run the test.
		//
		result := tmp.runTest(target)

		if result == nil {
			if ConfigOptions.Verbose {
				fmt.Printf("\tTest passed.\n")
			}

		} else {
			if ConfigOptions.Verbose {
				fmt.Printf("\tTest failed: %s\n", result.Error())
			}
		}

		//
		// Post the result to purple
		//
		postPurple(test_type, target, input, result)
	}

	return nil
}

//
// Open the named configuration file, and parse it
//
func main() {

	//
	// Our command-line options
	//
	flag.BoolVar(&ConfigOptions.IPv4, "4", true, "Should we run IPv4 tests?")
	flag.BoolVar(&ConfigOptions.IPv6, "6", true, "Should we run IPv6 tests?")
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

		//
		// Create an object to parse the given file.
		//
		helper := parser.New(file)

		//
		// For each parsed job call `run_test` to invoke it
		//
		err := helper.Parse(run_test)
		if ( err != nil ) {
			fmt.Printf("Error parsing file: %s\n", err.Error())
		}
	}

}
