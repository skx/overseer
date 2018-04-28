// Helper function to run tests and issue notifications

package main

import (
	"fmt"
	"net"
	"net/url"
	"strings"

	"github.com/skx/overseer/notifiers"

	"github.com/skx/overseer/protocols"
	"github.com/skx/overseer/test"
)

// run_test is the core of our application.
//
// Given a test to be executed this function is responsible for invoking
// it, and handling the result.
//
// The test result will be passed to the specified notifier instance upon
// completion.
//
func run_test(tst test.Test, opts test.TestOptions, notifier notifiers.Notifier) error {

	//
	// Setup our local state.
	//
	testType := tst.Type
	testTarget := tst.Target

	//
	// Look for a suitable protocol handler
	//
	tmp := protocols.ProtocolHandler(testType)

	//
	// Each test will be executed for each address-family, unless it is
	// a HTTP-test.
	//
	var targets []string

	//
	// If the first argument looks like an URI then get the host
	// out of it.
	//
	if strings.Contains(testTarget, "://") {
		u, err := url.Parse(testTarget)
		if err != nil {
			return err
		}
		testTarget = u.Host
	}

	//
	// Now resolve the target to IPv4 & IPv6 addresses.
	//
	ips, err := net.LookupIP(testTarget)
	if err != nil {

		//
		// If we have a notifier tell it that we failed.
		//
		if notifier != nil {
			notifier.Notify(tst, fmt.Errorf("Failed to resolve name %s", testTarget))
		}

		//
		// Otherwise we're done.
		//
		fmt.Printf("WARNING: Failed to resolve %s for %s test!\n", testTarget, testType)
		return err
	}

	//
	// We'll now run the test against each of the resulting IPv4 and
	// IPv6 addresess - ignoring any IP-protocol which is disabled.
	//
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

	//
	// Now for each target, run the test.
	//
	for _, target := range targets {

		//
		// Show what we're doing.
		//
		if opts.Verbose {
			fmt.Printf("Running '%s' test against %s (%s)\n", testType, testTarget, target)
		}

		//
		// We'll repeat failing tests up to five times by default
		//
		attempt := 0
		max_attempts := 5

		//
		// If retrying is disabled then don't retry.
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
			result = tmp.RunTest(tst, target, opts)

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
		// Post the result of the test to the notifier.
		//
		// Before we trigger the notification we need to
		// update the target to the thing we probed, which might
		// not necessarily be that which was originally submitted.
		//
		//  i.e. "mail.steve.org.uk must run ssh" might become
		// "1.2.3.4 must run ssh" as a result of the DNS lookup.
		//
		// However because we might run the same test against
		// multiple hosts we need to do this with a copy so that
		// we don't lose the original target.
		//
		copy := tst
		copy.Target = target
		if notifier != nil {
			notifier.Notify(copy, result)
		}
	}

	return nil
}
