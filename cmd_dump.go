// Dump
//
// The dump sub-command dumps the (parsed) configuration file(s)
package main

import (
	"context"
	"flag"
	"fmt"

	"github.com/google/subcommands"
	"github.com/skx/overseer/parser"
	"github.com/skx/overseer/test"
)

type dumpCmd struct {
}

//
// Glue
//
func (*dumpCmd) Name() string     { return "dump" }
func (*dumpCmd) Synopsis() string { return "Dump a parsed configuration file" }
func (*dumpCmd) Usage() string {
	return `dump :
  Dump a parsed configuration file.

  This is particularly useful to show the result of macro-expansion.
`
}

//
// Flag setup.
//
func (p *dumpCmd) SetFlags(f *flag.FlagSet) {
}

//
// This is a callback invoked by the parser when a job
// has been successfully parsed.
//
func dumpTest(tst test.Test) error {
	fmt.Printf("%s\n", tst.Input)
	return nil
}

//
// Entry-point.
//
func (p *dumpCmd) Execute(_ context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {

	for _, file := range f.Args() {

		//
		// Create an object to parse our file.
		//
		helper := parser.New()

		//
		// For each parsed job call `dump_test` to show it
		//
		err := helper.ParseFile(file, dumpTest)
		if err != nil {
			fmt.Printf("Error parsing file: %s\n", err.Error())
		}
	}

	return subcommands.ExitSuccess
}
