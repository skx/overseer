//
// Execute the tests locally, by pulling them from the queue.
//

package main

import (
	"context"
	"flag"
	"fmt"
	"time"

	"github.com/skx/overseer/protocols"

	"github.com/go-redis/redis"
	"github.com/google/subcommands"
)

type workerCmd struct {
	IPv4          bool
	IPv6          bool
	Purppura      string
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
	f.BoolVar(&p.IPv4, "4", true, "Enable IPv4 tests.")
	f.BoolVar(&p.IPv6, "6", true, "Enable IPv6 tests.")
	f.IntVar(&p.Timeout, "timeout", 10, "The global timeout for all tests, in seconds.")
	f.StringVar(&p.RedisHost, "redis-host", "localhost:6379", "Specify the address of the redis queue.")
	f.StringVar(&p.RedisPassword, "redis-pass", "", "Specify the password for the redis queue.")
	f.StringVar(&p.Purppura, "purppura", "", "Specify the URL of the purppura end-point.")
}

//
// Entry-point.
//
func (p *workerCmd) Execute(_ context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {

	//
	// Set the global purppura end-point
	//
	ConfigOptions.Purppura = p.Purppura

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
	// Options for our tests
	//
	var opts protocols.TestOptions
	opts.Verbose = p.Verbose
	opts.IPv4 = p.IPv4
	opts.IPv6 = p.IPv6
	opts.Timeout = time.Duration(p.Timeout) * time.Second

	//
	// Wait for the members
	//
	for true {

		test, _ := p._r.LPop("overseer.jobs").Result()

		if test != "" {
			run_test_string(test, opts)
		}

	}

	return subcommands.ExitSuccess
}
