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
package main

import (
	"fmt"
	"strconv"

	"github.com/rprtr258/cli"
	"github.com/rprtr258/fun"
)

type Options struct {
	prefix  cli.Option[uint]
	command string
}

var options = func() cli.Parser[Options] {
	prefix := cli.Optional(cli.Positional("PREFIX", func(s string) (uint, error) {
		u, err := strconv.ParseUint(s, 10, 64)
		return uint(u), err
	}).
		Help("Optional numeric command prefix").Parser(), true).
		Parser()
	command := cli.Positional("COMMAND", func(s string) (string, error) {
		return s, nil
	}).Help("Required command name").Parser()

	return cli.Map(cli.Pair(prefix, command).Parser(), func(x fun.Pair[cli.Option[uint], string]) Options { return Options{x.K, x.V} })
}()

func main() {
	fmt.Println(options.ToOptions().Run())
}
