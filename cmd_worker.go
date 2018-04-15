// The worker sub-command executes tests pulled from a central redis queue.
package main

import (
	"context"
	"flag"
	"fmt"
	"time"

	"github.com/skx/overseer/notifiers"
	"github.com/skx/overseer/parser"
	"github.com/skx/overseer/protocols"

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
	f.BoolVar(&p.Verbose, "verbose", false, "Show more output.")
	f.BoolVar(&p.Retry, "retry", true, "Should failing tests be retried a few times before raising a notification.")
	f.BoolVar(&p.IPv4, "4", true, "Enable IPv4 tests.")
	f.BoolVar(&p.IPv6, "6", true, "Enable IPv6 tests.")
	f.IntVar(&p.Timeout, "timeout", 10, "The global timeout for all tests, in seconds.")
	f.StringVar(&p.RedisHost, "redis-host", "localhost:6379", "Specify the address of the redis queue.")
	f.StringVar(&p.RedisPassword, "redis-pass", "", "Specify the password for the redis queue.")

	// Notifier setup
	f.StringVar(&p.Notifier, "notifier", "", "Specify the notifier object to use.")
	f.StringVar(&p.NotifierData, "", "", "Specify the notifier data to use.")
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
		notifier = notifiers.NotifierType(p.Notifier)

		//
		// Set the notifier options
		//
		if p.NotifierData != "" {
			var nopt notifiers.Options
			nopt.Data = p.NotifierData
			notifier.SetOptions(nopt)
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
	var opts protocols.TestOptions
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
