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
	"os"
	"time"

	"github.com/skx/overseer/notifiers"
	"github.com/skx/overseer/parser"
	"github.com/skx/overseer/test"

	"github.com/go-redis/redis"
	"github.com/google/subcommands"
)

type workerCmd struct {
	IPv4          bool
	IPv6          bool
	Retry         bool
	Notifier      string
	NotifierData  string
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
func (*workerCmd) Synopsis() string { return "Execute tests via the queue." }
func (*workerCmd) Usage() string {
	return `worker :
  Execute tests from the redis queue, until terminated.
`
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
	defaults.Notifier = ""
	defaults.NotifierData = ""
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

	// Notifier setup
	f.StringVar(&p.Notifier, "notifier", defaults.Notifier, "Specify the notifier object to use.")
	f.StringVar(&p.NotifierData, "", defaults.NotifierData, "Specify the notifier data to use.")
}

//
// Entry-point.
//
func (p *workerCmd) Execute(_ context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {

	//
	// The notifier we're going to use, if any.
	//
	var notifier notifiers.Notifier

	//
	// If the notifier is set then create it.
	//
	if p.Notifier != "" {
		notifier = notifiers.NotifierType(p.Notifier, p.NotifierData)

		if notifier == nil {
			fmt.Printf("Unknown notifier: %s\n", p.Notifier)
			os.Exit(1)
		}
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
	_, err := p._r.Ping().Result()
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
		test, _ := p._r.LPop("overseer.jobs").Result()

		//
		// Parse it
		//
		if test != "" {
			job, err := parse.ParseLine(test, nil)

			if err == nil {
				run_test(job, opts, notifier)
			} else {
				fmt.Printf("Error parsing job from queue: %s - %s\n", test, err.Error())
			}
		}

	}

	return subcommands.ExitSuccess
}
