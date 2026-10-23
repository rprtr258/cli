// Positional arguments built with the combinator API.
package examples

import (
	"fmt"
	"os"
	"strconv"

	"github.com/rprtr258/cli"
	"github.com/rprtr258/fun"
)

func ExPositional() {
	type Options struct {
		value uint32
		files []string
	}

	value := cli.Argument[uint32](cli.Long("value").
		Help("Mysterious value"),
		"VAL", func(s string) (uint32, error) {
			u, err := strconv.ParseUint(s, 10, 32)
			return uint32(u), err
		}).
		Parser().
		Fallback(42).
		Parser()
	files := cli.Positional("FILE", func(s string) (string, error) {
		return s, nil
	}).Parser().Many()
	opts := cli.Pair(value, files).Map(func(x fun.Pair[uint32, []string]) Options { return Options{x.K, x.V} }).ToOptions().Run(os.Args)

	fmt.Println(opts)
}
