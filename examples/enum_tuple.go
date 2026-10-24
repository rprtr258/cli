// How to extract subcommands' args into external structs.
package examples

import (
	"fmt"
	"os"

	"github.com/rprtr258/cli"
)

func EnumTuple() {
	type Foo struct {
		bar cli.Opt[string]
	}

	type Command func(
		Foo func(Foo),
	)

	CommandFoo := func(foo Foo) Command {
		return func(f func(Foo)) {
			if f != nil {
				f(foo)
			}
		}
	}

	bar := cli.Argument[string](cli.Short('b').
		Long("bar").
		Help("some bar command"),
		"BAR", func(s string) (string, error) {
			return s, nil
		}).
		Parser().
		Optional()

	bar_cmd := bar.Map(func(bar cli.Opt[string]) Foo { return Foo{bar} }).
		ToOptions().
		Description("This command will try to do foo given a bar argument")

	opt := bar_cmd.
		Command("foo").
		Help("command for doing foo").
		Map(CommandFoo).
		ToOptions().Run(os.Args)

	fmt.Println(opt)
}
