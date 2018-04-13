//
// Enqueue jobs to a central redis-queue.
//

package main

import (
	"context"
	"flag"
	"fmt"

	"github.com/go-redis/redis"
	"github.com/google/subcommands"
	"github.com/skx/overseer/parser"
)

type enqueueCmd struct {
	Redis string
	_r    *redis.Client
}

//
// Glue
//
func (*enqueueCmd) Name() string     { return "enqueue" }
func (*enqueueCmd) Synopsis() string { return "Enqueue a parsed configuration file." }
func (*enqueueCmd) Usage() string {
	return `enqueue :
  Enqueue a parsed configuration file to a central redis instance.
`
}

//
// Flag setup.
//
func (p *enqueueCmd) SetFlags(f *flag.FlagSet) {
	f.StringVar(&p.Redis, "redis", "localhost:6379", "Specify the address of the redis queue.")
}

//
// This is a callback invoked by the parser when a job
// has been successfully parsed.
//
func (p *enqueueCmd) enqueue_test(tst parser.Test) error {
	_, err := p._r.RPush("overseer.jobs", tst.Input).Result()
	return err
}

//
// Entry-point.
//
func (p *enqueueCmd) Execute(_ context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {

	//
	// Connect to the redis-host.
	//
	p._r = redis.NewClient(&redis.Options{
		Addr: p.Redis,
		DB:   0, // use default DB
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
	// For each file on the command-line we can now parse and
	// enqueue the jobs
	//
	for _, file := range f.Args() {

		//
		// Create an object to parse the given file.
		//
		helper := parser.New(file)

		//
		// For each parsed job call `enqueue_test`.
		//
		err := helper.Parse(p.enqueue_test)

		//
		// Did we see an error?
		//
		if err != nil {
			fmt.Printf("Error parsing file: %s\n", err.Error())
		}
	}

	return subcommands.ExitSuccess
}
