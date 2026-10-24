// `shared_args.go` - it is possible to have shared args returned alongside a subcommand.
package examples

import (
	"fmt"
	"os"

	"github.com/rprtr258/cli"
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
	shared := cli.Positional("ARG", cli.ParseString).Many()

	action := cli.Map2(
		cli.Map2(
			cli.Long("verbose").Switch(),
			cli.Long("number").Help("Number of things").Argument("NUM", cli.ParseUint),
			func(verbose bool, number uint) sharedCommand {
				return sharedCommand{kind: sharedCommandAction, action: sharedAction{verbose: verbose, number: number}}
			},
		),
		shared,
		func(command sharedCommand, args []string) sharedResult {
			return sharedResult{command, args}
		}).
		ToOptions().
		Command("action")

	build := cli.Map2(
		cli.Long("verbose").Switch().Map(func(verbose bool) sharedCommand {
			return sharedCommand{kind: sharedCommandBuild, buildVerbose: verbose}
		}),
		shared,
		func(command sharedCommand, args []string) sharedResult {
			return sharedResult{command, args}
		}).
		ToOptions().
		Command("build")

	opts := cli.Choice(action, build).ToOptions().Run(os.Args)
	fmt.Println(opts)
}
