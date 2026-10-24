// Positional arguments built with the combinator API.
package examples

import (
	"fmt"
	"os"
	"strconv"

	"github.com/rprtr258/cli"
)

func ExPositional() {
	type Options struct {
		value uint32
		files []string
	}

	value := cli.Long("value").
		Help("Mysterious value").Argument("VAL", func(s string) (uint32, error) {
		u, err := strconv.ParseUint(s, 10, 32)
		return uint32(u), err
	}).
		Parser().
		Fallback(42).
		Parser()
	files := cli.Positional("FILE", func(s string) (string, error) {
		return s, nil
	}).Parser().Many()
	opts := cli.Map2(value, files, func(value uint32, files []string) Options { return Options{value, files} }).ToOptions().Run(os.Args)

	fmt.Println(opts)
}
