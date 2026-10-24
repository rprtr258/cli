// `fallback_command.rs` - parse some commands manually or collect anything else as
// for manual parsing.
package examples

import (
	"fmt"
	"os"

	"github.com/rprtr258/cli"
	"github.com/rprtr258/fun"
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
		Parser().
		ToOptions().
		Command("build").
		Parser()

	fallback := cli.Pair(
		cli.Positional("COMMAND", cli.ParseString).Parser(),
		cli.Any("ARG", func(s string) (string, bool) { return s, true }, false).Parser().Many(),
	).
		Parser().
		Map(func(p fun.Pair[string, []string]) fallbackCommand {
			return fallbackCommand{kind: fallbackCommandFallback, name: p.K, args: p.V}
		}).
		Hide()

	opts := cli.Choice(build, fallback).ToOptions().Run(os.Args)
	fmt.Println(opts)
}
