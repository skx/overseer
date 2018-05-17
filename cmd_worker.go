// Worker
//
// The worker sub-command executes tests pulled from a central redis queue.
package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/go-redis/redis"
	"github.com/google/subcommands"
	"github.com/skx/overseer/notifier"
	"github.com/skx/overseer/parser"
	"github.com/skx/overseer/protocols"
	"github.com/skx/overseer/test"
)

// This is our structure, largely populated by command-line arguments
type workerCmd struct {
	IPv4          bool
	IPv6          bool
	Retry         bool
	RedisHost     string
	RedisPassword string
	Timeout       int
	Verbose       bool
	_r            *redis.Client
}

//
// Glue
//
func (*workerCmd) Name() string     { return "worker" }
func (*workerCmd) Synopsis() string { return "Fetch jobs from the central queue and execute them" }
func (*workerCmd) Usage() string {
	return `worker :
  Execute tests pulled from the central redis queue, until terminated.
`
}

// verbose shows a message only if we're running verbosely
func (p *workerCmd) verbose(txt string) {
	if p.Verbose {
		fmt.Sprintf("%s", txt)
	}
}

//
// Flag setup.
//
func (p *workerCmd) SetFlags(f *flag.FlagSet) {

	//
	// Create the default options here
	//
	// This is done so we can load defaults via a configuration-file
	// if present.
	//
	var defaults workerCmd
	defaults.IPv4 = true
	defaults.IPv6 = true
	defaults.Retry = true
	defaults.Timeout = 10
	defaults.Verbose = false
	defaults.RedisHost = "localhost:6379"
	defaults.RedisPassword = ""

	//
	// If we have a configuration file then load it
	//
	if len(os.Getenv("OVERSEER")) > 0 {
		cfg, err := ioutil.ReadFile(os.Getenv("OVERSEER"))
		if err == nil {
			err = json.Unmarshal(cfg, &defaults)
			if err != nil {
				fmt.Printf("WARNING: Error loading overseer.json - %s\n",
					err.Error())
			}
		} else {
			fmt.Printf("WARNING: Failed to read configuration-file - %s\n",
				err.Error())
		}
	}

	f.BoolVar(&p.Verbose, "verbose", defaults.Verbose, "Show more output.")
	f.BoolVar(&p.Retry, "retry", defaults.Retry, "Should failing tests be retried a few times before raising a notification.")
	f.BoolVar(&p.IPv4, "4", defaults.IPv4, "Enable IPv4 tests.")
	f.BoolVar(&p.IPv6, "6", defaults.IPv6, "Enable IPv6 tests.")
	f.IntVar(&p.Timeout, "timeout", defaults.Timeout, "The global timeout for all tests, in seconds.")
	f.StringVar(&p.RedisHost, "redis-host", defaults.RedisHost, "Specify the address of the redis queue.")
	f.StringVar(&p.RedisPassword, "redis-pass", defaults.RedisPassword, "Specify the password for the redis queue.")
}

// runTest is really the core of our application.
//
// Given a test to be executed this function is responsible for invoking
// it, and handling the result.
//
// The test result will be passed to the specified notifier instance upon
// completion.
//
func (p *workerCmd) runTest(tst test.Test, opts test.TestOptions, notify *notifier.Notifier) error {

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
	// Each test will be executed for each address-family, so we need to
	// keep track of the IPs of the real test-target.
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
		// Notify the world about our DNS-failure.
		//
		notify.Notify(tst, fmt.Errorf("Failed to resolve name %s", testTarget))

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
		p.verbose(fmt.Sprintf("Running '%s' test against %s (%s)\n", testType, testTarget, target))

		//
		// We'll repeat failing tests up to five times by default
		//
		attempt := 0
		maxAttempts := 5

		//
		// If retrying is disabled then don't retry.
		//
		if opts.Retry == false {
			maxAttempts = attempt + 1
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
		for attempt < maxAttempts {
			attempt += 1

			//
			// Run the test
			//
			result = tmp.RunTest(tst, target, opts)

			//
			// If the test passed then we're good.
			//
			if result == nil {
				p.verbose(fmt.Sprintf("\t[%d/%d] - Test passed.\n", attempt, maxAttempts))

				// break out of loop
				attempt = maxAttempts + 1

			} else {

				//
				// The test failed.
				//
				// It will be repeated before a notifier
				// is invoked.
				//
				p.verbose(fmt.Sprintf("\t[%d/%d] Test failed: %s\n", attempt, maxAttempts, result.Error()))

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

		//
		// We also want to filter out any password which was found
		// on the input-line.
		//
		copy.Input = tst.Sanitize()

		//
		// Now we can trigger the notification with our updated
		// copy of the test.
		//
		notify.Notify(copy, result)
	}

	return nil
}

//
// Entry-point.
//
func (p *workerCmd) Execute(_ context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {

	//
	// Create a notifier object, for posting our results.
	//
	notify, err := notifier.New(p.RedisHost, p.RedisPassword)

	if err != nil {
		fmt.Printf("Failed to connect to redis for publishing results: %s\n", err.Error())
		os.Exit(1)
	}

	//
	// Connect to the redis-host.
	//
	p._r = redis.NewClient(&redis.Options{
		Addr:     p.RedisHost,
		Password: p.RedisPassword,
		DB:       0, // use default DB
	})

	//
	// And run a ping, just to make sure it worked.
	//
	_, err = p._r.Ping().Result()
	if err != nil {
		fmt.Printf("Redis connection failed: %s\n", err.Error())
		return subcommands.ExitFailure
	}

	//
	// Setup the options for the tests.
	//
	var opts test.TestOptions
	opts.Verbose = p.Verbose
	opts.IPv4 = p.IPv4
	opts.IPv6 = p.IPv6
	opts.Retry = p.Retry
	opts.Timeout = time.Duration(p.Timeout) * time.Second

	//
	// Create a parser for our input
	//
	parse := parser.New()

	//
	// Wait for the members
	//
	for true {

		//
		// Get a job.
		//
		test, _ := p._r.BLPop(0, "overseer.jobs").Result()

		//
		// Parse it
		//
		//   test[0] will be "overseer.jobs"
		//
		//   test[1] will be the value removed from the list.
		//
		if len(test) >= 1 {
			job, err := parse.ParseLine(test[1], nil)

			if err == nil {
				p.runTest(job, opts, notify)
			} else {
				fmt.Printf("Error parsing job from queue: %s - %s\n", test[1], err.Error())
			}
		}

	}

	return subcommands.ExitSuccess
}
