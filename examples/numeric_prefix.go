// You can parse multiple positional elements with earlier being optional as well
// This example takes two - optional numeric prefix and a command name:
//
// > numeric_prefix 8 work
// Options { prefix: Some(8), command: "work" }
//
// > numeric_prefix sleep
// Options { prefix: None, command: "sleep" }
//
// Generated usage reflects that:
// Usage: numeric_prefix [PREFIX] COMMAND
package examples

import (
	"fmt"
	"os"

	"github.com/rprtr258/cli"
)

func NumericPrefix() {
	type Options struct {
		prefix  cli.Opt[uint]
		command string
	}

	prefix := cli.Positional("PREFIX", cli.ParseUint).
		Help("Optional numeric command prefix").
		Parser().
		OptionalCatch()
	command := cli.Positional("COMMAND", cli.ParseString).Help("Required command name")
	options := cli.Map2(prefix, command, func(prefix cli.Opt[uint], command string) Options { return Options{prefix, command} })

	fmt.Println(options.ToOptions().Run(os.Args))
}
