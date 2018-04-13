package main

import (
	"bytes"
	"crypto/sha1"
	"encoding/base64"
	"encoding/json"

	"errors"
	"fmt"
	"net"
	"net/http"
	"regexp"
	"strings"

	"github.com/skx/overseer/parser"
	"github.com/skx/overseer/protocols"
)

//
// ConfigOptions is the globally visible structure which is designed to
// hold our configuration-options - as set by the command-line flags.
//
// It is perhaps poor practice to do things this way, but it eases coding.
//
var ConfigOptions struct {
	Purppura string
}

//
// Regardless of whether a test fails/passes we must pass the result
// on to purppura.
//
func postPurple(tst parser.Test, result error) {

	//
	// If we don't have a server configured then
	// return without sending
	//
	if ConfigOptions.Purppura == "" {
		return
	}

	test_type := tst.Type
	test_target := tst.Target
	input := tst.Input

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
// Run the given test, with the options
//
// This is horrid because it exposes our internal state - it should
// be moved to the parser-object.
//
// For the moment it remains because it is used to parse the string
// fetched from redis.
//
func run_test_string(tst string, opts protocols.TestOptions) error {

	var obj parser.Test

	re := regexp.MustCompile("^([^ \t]+)\\s+must\\s+run\\s+([a-z]+)")
	out := re.FindStringSubmatch(tst)

	//
	// If it didn't then we have a malformed line
	//
	if len(out) != 3 {
		return errors.New(fmt.Sprintf("WARNING: Unrecognized line - '%s'\n", tst))
	}

	//
	// Save the type + target away
	//
	obj.Target = out[1]
	obj.Type = out[2]
	obj.Input = tst

	return (run_test(obj, opts))

}

//
// Run the given test, with the options.
//
func run_test(tst parser.Test, opts protocols.TestOptions) error {

	//
	// Setup our local state.
	//
	test_type := tst.Type
	test_target := tst.Target
	input := tst.Input

	//
	// Look for a suitable protocol handler
	//
	tmp := protocols.ProtocolHandler(test_type)
	if tmp == nil {
		fmt.Printf("WARNING: Unknown protocol handler '%s'\n", test_type)
		postPurple(tst, errors.New(fmt.Sprintf("Unknown protocol-handler %s", test_type)))
		return nil
	}

	//
	// Pass the full input-line to our protocol tester
	// to allow any extra options/flags to be parsed
	//
	tmp.SetLine(input)

	//
	// Set our options
	//
	tmp.SetOptions(opts)

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

			postPurple(tst, errors.New(fmt.Sprintf("Failed to resolve name %s", test_target)))
			fmt.Printf("WARNING: Failed to resolve %s\n", test_target)
			return nil
		}
		for _, ip := range ips {
			if ip.To4() != nil {
				if opts.IPv4 {
					targets = append(targets, fmt.Sprintf("%s", ip))
				}
			}
			if ip.To16() != nil && ip.To4() == nil {
				if opts.IPv6 {
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
		if opts.Verbose {
			fmt.Printf("Running %s test against %s (%s)\n", test_type, test_target, target)
		}

		//
		// Run the test.
		//
		result := tmp.RunTest(target)

		if result == nil {
			if opts.Verbose {
				fmt.Printf("\tTest passed.\n")
			}

		} else {
			if opts.Verbose {
				fmt.Printf("\tTest failed: %s\n", result.Error())
			}
		}

		//
		// Post the result to purple
		//
		// First update the target to the thing we probed,
		// which might not necessarily be that which was
		// written.
		//
		//  i.e. "mail.steve.org.uk must run ssh"
		// might become "1.2.3.4 must run ssh"
		//
		tst.Target = target
		postPurple(tst, result)
	}

	return nil
}
