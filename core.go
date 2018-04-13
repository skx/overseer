package main

import (
	"errors"
	"fmt"
	"net"
	"regexp"
	"strings"

	"github.com/skx/overseer/notifiers"
	"github.com/skx/overseer/parser"
	"github.com/skx/overseer/protocols"
)

//
// Run the given test, with the options
//
// This is horrid because it exposes our internal state - it should
// be moved to the parser-object.
//
// For the moment it remains because it is used to parse the string
// fetched from redis.
//
func run_test_string(tst string, opts protocols.TestOptions, notifier notifiers.Notifier) error {

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

	return (run_test(obj, opts, notifier))

}

//
// Run the given test, with the options, and send notification with the
// given notifier.
//
func run_test(tst parser.Test, opts protocols.TestOptions, notifier notifiers.Notifier) error {

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
		if notifier != nil {
			notifier.Notify(tst, errors.New(fmt.Sprintf("Unknown protocol-handler %s", test_type)))
		}
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

			if notifier != nil {
				notifier.Notify(tst, errors.New(fmt.Sprintf("Failed to resolve name %s", test_target)))
			}
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
		if notifier != nil {
			notifier.Notify(tst, result)
		}
	}

	return nil
}
