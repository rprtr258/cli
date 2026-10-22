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
	"strconv"

	"github.com/rprtr258/cli"
	"github.com/rprtr258/fun"
)

func NumericPrefix() {
	type Options struct {
		prefix  cli.Opt[uint]
		command string
	}

	options := func() cli.Parser[Options] {
		prefix := cli.Positional("PREFIX", func(s string) (uint, error) {
			u, err := strconv.ParseUint(s, 10, 64)
			return uint(u), err
		}).
			Help("Optional numeric command prefix").
			Parser().
			OptionalCatch()
		command := cli.Positional("COMMAND", func(s string) (string, error) {
			return s, nil
		}).Help("Required command name")

		return cli.Pair(prefix, command).Map(func(x fun.Pair[cli.Opt[uint], string]) Options { return Options{x.K, x.V} })
	}()

	fmt.Println(options.ToOptions().Run(os.Args))
}
