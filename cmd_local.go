// The local sub-command executes parsed tests locally.
package main

import (
	"context"
	"flag"
	"fmt"
	"time"

	"github.com/google/subcommands"
	"github.com/skx/overseer/notifiers"
	"github.com/skx/overseer/parser"
	"github.com/skx/overseer/protocols"
	"github.com/skx/overseer/test"
)

type localCmd struct {
	IPv4         bool
	IPv6         bool
	Retry        bool
	Notifier     string
	NotifierData string
	Timeout      int
	Verbose      bool
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
	f.BoolVar(&p.Retry, "retry", true, "Should failing tests be retried a few times before raising a notification.")
	f.IntVar(&p.Timeout, "timeout", 10, "The global timeout for all tests, in seconds.")

	// Notifier setup
	f.StringVar(&p.Notifier, "notifier", "", "Specify the notifier object to use.")
	f.StringVar(&p.NotifierData, "notifier-data", "", "Specify the notifier data to use.")
}

//
// This is a callback invoked by the parser when a job
// has been successfully parsed.
//
// Run the test, via our helper
//
func (p *localCmd) run_test(tst test.Test) error {

	//
	// Setup the options for the test.
	//
	var opts protocols.TestOptions
	opts.Verbose = p.Verbose
	opts.IPv4 = p.IPv4
	opts.IPv6 = p.IPv6
	opts.Retry = p.Retry
	opts.Timeout = time.Duration(p.Timeout) * time.Second

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
	// Now run the test.
	//
	return run_test(tst, opts, notifier)
}

//
// Entry-point.
//
func (p *localCmd) Execute(_ context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {

	//
	// We'll run tests from each file passed as an argument.
	//
	for _, file := range f.Args() {

		//
		// Create an object to parse our file.
		//
		helper := parser.New()

		//
		// Invoke the run_test callback to execute each test.
		//
		err := helper.ParseFile(file, p.run_test)
		if err != nil {
			fmt.Printf("Error parsing file: %s\n", err.Error())
		}
	}

	return subcommands.ExitSuccess
}
