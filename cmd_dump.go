//
// Dump the (parsed) configuration file(s)
//

package main

import (
	"context"
	"flag"
	"fmt"

	"github.com/google/subcommands"
	"github.com/skx/overseer/parser"
)

type dumpCmd struct {
}

//
// Glue
//
func (*dumpCmd) Name() string     { return "dump" }
func (*dumpCmd) Synopsis() string { return "Dump a parsed configuration file." }
func (*dumpCmd) Usage() string {
	return `dump :
  Dump a parsed configuration file, showing expanded macros.
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
func dump_test(tst parser.Test) error {
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
		err := helper.ParseFile(file, dump_test)
		if err != nil {
			fmt.Printf("Error parsing file: %s\n", err.Error())
		}
	}

	return subcommands.ExitSuccess
}
