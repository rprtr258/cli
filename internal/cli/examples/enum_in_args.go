// parsing argument value into enum. You can use crate `strum`'s `EnumString` for this purposes as well
package examples

import (
	"errors"
	"fmt"
	"os"

	"github.com/rprtr258/cli/internal/cli"
)

func EnumInArgs() {
	type Baz uint8

	const (
		BazFoo Baz = iota
		BazBar
		BazFooBar
	)

	parseBaz := func(s string) (Baz, error) {
		switch s {
		case "foo":
			return BazFoo, nil
		case "bar":
			return BazBar, nil
		case "foobar":
			return BazFooBar, nil
		default:
			return 0, errors.New("Expected foo|bar|foobar")
		}
	}

	opt := cli.Argument(cli.Long("baz").
		Short('b').
		Help("choose between foo, bar or foobar"), "CMD", parseBaz).
		Parser().
		ToOptions().Run(os.Args)

	fmt.Println(opt)
}
