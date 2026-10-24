// `fallback_command.go` - parse some commands manually or collect anything else as for manual parsing.
package examples

import (
	"fmt"
	"os"

	"github.com/rprtr258/cli"
)

type fallbackCommandKind uint8

const (
	fallbackCommandBuild fallbackCommandKind = iota
	fallbackCommandFallback
)

type fallbackCommand struct {
	kind fallbackCommandKind
	opt  uint
	name string
	args []string
}

func FallbackCommand() {
	build := cli.Positional("OPT", cli.ParseUint).
		Help("Optimization level").
		Map(func(opt uint) fallbackCommand {
			return fallbackCommand{kind: fallbackCommandBuild, opt: opt}
		}).
		ToOptions().
		Command("build")

	fallback := cli.Map2(
		cli.Positional("COMMAND", cli.ParseString),
		cli.Any("ARG", func(s string) (string, bool) { return s, true }, false).Many(),
		func(name string, args []string) fallbackCommand {
			return fallbackCommand{kind: fallbackCommandFallback, name: name, args: args}
		}).
		Hide()

	opts := cli.Choice(build, fallback).ToOptions().Run(os.Args)
	fmt.Println(opts)
}
