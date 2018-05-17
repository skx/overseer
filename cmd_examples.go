// Examples
//
// Show information about our protocols.
package main

import (
	"context"
	"flag"
	"fmt"
	"regexp"
	"sort"

	"github.com/google/subcommands"
	"github.com/skx/overseer/protocols"
)

type examplesCmd struct {
}

//
// Glue
//
func (*examplesCmd) Name() string     { return "examples" }
func (*examplesCmd) Synopsis() string { return "Show example protocol-tests." }
func (*examplesCmd) Usage() string {
	return `examples :
  Provide sample usage of each of our protocol-tests.
`
}

//
// Flag setup.
//
func (p *examplesCmd) SetFlags(f *flag.FlagSet) {
}

//
// Show example output for any protocol-handler matching the
// pattern specified.
//
// If the filter is empty then show all.
//
func showExamples(filter string) {

	re := regexp.MustCompile(filter)

	// For each (sorted) protocol-handler
	handlers := protocols.Handlers()
	sort.Strings(handlers)

	// Get the name
	for _, name := range handlers {

		// Skip unless this handler matches the filter.
		match := re.FindAllStringSubmatch(name, -1)
		if len(match) < 1 {
			continue
		}

		// Create an instance of it
		x := protocols.ProtocolHandler(name)

		// Show the output of that function
		out := x.Example()
		fmt.Printf("%s\n", out)

		fmt.Printf("Arguments which are supported are now shown:\n\n")

		fmt.Printf("  %10s|%s\n", "Name", "Valid Value")
		fmt.Printf("  ----------------------------------\n")

		//
		// The arguments this test supports
		//
		m := x.Arguments()

		//
		// Temporary structure to store the keys.
		//
		var keys []string
		for k := range m {
			keys = append(keys, k)
		}
		sort.Strings(keys)

		//
		// Now show the keys + values in sorted order
		//
		for _, k := range keys {
			fmt.Printf("  %10s|%s\n", k, m[k])
		}
		fmt.Printf("\n\n")

	}
}

//
// Entry-point.
//
func (p *examplesCmd) Execute(_ context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {

	if len(f.Args()) > 0 {
		for _, name := range f.Args() {
			showExamples(name)
		}
	} else {
		showExamples(".*")
	}
	return subcommands.ExitSuccess
}
