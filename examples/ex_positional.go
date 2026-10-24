// Positional arguments built with the combinator API.
package examples

import (
	"fmt"
	"os"

	"github.com/rprtr258/cli"
)

func ExPositional() {
	type Options struct {
		value uint32
		files []string
	}

	opts := cli.Map2(
		cli.Long("value").
			Help("Mysterious value").
			Argument("VAL", cli.ParseUint32).
			Fallback(42),
		cli.Positional("FILE", cli.ParseString).
			Many(),
		func(value uint32, files []string) Options { return Options{value, files} }).
		ToOptions().
		Run(os.Args)

	fmt.Println(opts)
}
