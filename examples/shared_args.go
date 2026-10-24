// `shared_args.rs` - it is possible to have shared args returned alongside a subcommand.
package examples

import (
	"fmt"
	"os"

	"github.com/rprtr258/cli"
	"github.com/rprtr258/fun"
)

type sharedAction struct {
	verbose bool
	number  uint
}

type sharedCommandKind uint8

const (
	sharedCommandAction sharedCommandKind = iota
	sharedCommandBuild
)

type sharedCommand struct {
	kind         sharedCommandKind
	action       sharedAction
	buildVerbose bool
}

type sharedResult struct {
	command sharedCommand
	args    []string
}

func SharedArgs() {
	shared := cli.Positional("ARG", cli.ParseString).Parser().Many()

	action := cli.Pair(
		cli.Map2(
			cli.Long("verbose").Switch(),
			cli.Long("number").Help("Number of things").Argument("NUM", cli.ParseUint).Parser(),
			func(verbose bool, number uint) sharedCommand {
				return sharedCommand{kind: sharedCommandAction, action: sharedAction{verbose: verbose, number: number}}
			},
		).Parser(),
		shared,
	).
		Parser().
		Map(func(p fun.Pair[sharedCommand, []string]) sharedResult {
			return sharedResult{command: p.K, args: p.V}
		}).
		ToOptions().
		Command("action").
		Parser()

	build := cli.Pair(
		cli.Long("verbose").Switch().
			Map(func(verbose bool) sharedCommand {
				return sharedCommand{kind: sharedCommandBuild, buildVerbose: verbose}
			}).Parser(),
		shared,
	).
		Parser().
		Map(func(p fun.Pair[sharedCommand, []string]) sharedResult {
			return sharedResult{command: p.K, args: p.V}
		}).
		ToOptions().
		Command("build").
		Parser()

	opts := cli.Choice(action, build).ToOptions().Run(os.Args)
	fmt.Println(opts)
}
