//
// Execute the tests locally, with no queue-usage.
//

package main

import (
	"context"
	"flag"
	"fmt"
	"time"

	"github.com/google/subcommands"
	"github.com/skx/overseer/parser"
	"github.com/skx/overseer/protocols"
)

type localCmd struct {
	Verbose bool
	IPv4    bool
	IPv6    bool
}

//
// Glue
//
func (*localCmd) Name() string     { return "local" }
func (*localCmd) Synopsis() string { return "Execute tests locally." }
func (*localCmd) Usage() string {
	return `local :
  Execute the tests in the given files locally, without the use of a queue.
`
}

//
// Flag setup.
//
func (p *localCmd) SetFlags(f *flag.FlagSet) {
	f.BoolVar(&p.Verbose, "verbose", false, "Show more output.")
	f.BoolVar(&p.IPv4, "4", true, "Enable IPv4 tests.")
	f.BoolVar(&p.IPv6, "6", true, "Enable IPv6 tests.")
}

//
// This is a callback invoked by the parser when a job
// has been successfully parsed.
//
// Run the test, via our helper
//
func (p *localCmd) run_test(tst parser.Test) error {

	var opts protocols.TestOptions
	opts.Verbose = p.Verbose
	opts.IPv4 = p.IPv4
	opts.IPv6 = p.IPv6
	opts.Timeout = 10 * time.Second

	return run_test(tst, opts)
}

//
// Entry-point.
//
func (p *localCmd) Execute(_ context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {

	for _, file := range f.Args() {

		//
		// Create an object to parse the given file.
		//
		helper := parser.New(file)

		//
		// Invoke the run_test callback to execute each test.
		//
		err := helper.Parse(p.run_test)
		if err != nil {
			fmt.Printf("Error parsing file: %s\n", err.Error())
		}
	}

	return subcommands.ExitSuccess
}
