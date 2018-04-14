package main

import (
	"errors"
	"fmt"
	"net"
	"strings"

	"github.com/skx/overseer/notifiers"
	"github.com/skx/overseer/parser"
	"github.com/skx/overseer/protocols"
)

// run_test is the core of our application.
//
// Given a test to be executed this function is responsible for invoking
// it, and handling the result.
//
// The test result will be passed to the specified notifier instance upon
// completion.
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
			fmt.Printf("Running '%s' test against %s (%s)\n", test_type, test_target, target)
		}

		//
		// We'll repeat failing tests up to five times by default
		//
		attempt := 0
		max_attempts := 5

		//
		// If retrying is disabled then don't do that
		//
		if opts.Retry == false {
			max_attempts = attempt + 1
		}

		//
		// The result of the test.
		//
		var result error

		//
		// Prepare to repeat the test.
		//
		// We only repeat tests that fail, if the test passes then
		// it will only be executed once.
		//
		// This is designed to cope with transient failures, at a
		// cost that flapping services might be missed.
		//
		for attempt < max_attempts {
			attempt += 1

			//
			// Run the test
			//
			result = tmp.RunTest(target)

			//
			// If the test passed then we're good.
			//
			if result == nil {
				if opts.Verbose {
					fmt.Printf("\t[%d/%d] - Test passed.\n", attempt, max_attempts)
				}

				// break out of loop
				attempt = max_attempts + 1

			} else {

				//
				// The test failed.
				//
				// It will be repeated before a notifier
				// is invoked.
				//
				if opts.Verbose {
					fmt.Printf("\t[%d/%d] Test failed: %s\n", attempt, max_attempts, result.Error())
				}

			}
		}

		//
		// Post the result to the specified notifier.
		//
		// Before we trigger the notification we need to
		// update the target to the thing we probed, which might
		// not necessarily be that which was originally submitted.
		//
		//  i.e. "mail.steve.org.uk must run ssh" might become
		// "1.2.3.4 must run ssh" as a result of the DNS lookup.
		//
		tst.Target = target
		if notifier != nil {
			notifier.Notify(tst, result)
		}
	}

	return nil
}
