// How to extract subcommands' args into external structs.
package enum_tuple

import (
	"fmt"
	"os"

	"github.com/rprtr258/cli"
)

type Foo struct {
	bar cli.Option[string]
}

type Command func(
	Foo func(Foo),
)

func CommandFoo(foo Foo) Command {
	return func(f func(Foo)) {
		if f != nil {
			f(foo)
		}
	}
}

func Main() {
	bar := cli.Optional(cli.Argument[string](cli.Short('b').
		Long("bar").
		Help("some bar command"),
		"BAR", func(s string) (string, error) {
			return s, nil
		}).Parser(), false).Parser()

	bar_cmd := cli.Map(bar, func(bar cli.Option[string]) Foo { return Foo{bar} }).
		ToOptions().
		Descr("This command will try to do foo given a bar argument")

	opt := cli.Map(bar_cmd.
		Command("foo").
		Help("command for doing foo").Parser(),
		CommandFoo).
		ToOptions().Run(os.Args)

	fmt.Println(opt)
}
