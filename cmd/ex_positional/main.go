// Non derive version for positional arguments
package main

import (
	"fmt"
	"strconv"

	"github.com/rprtr258/cli"
	"github.com/rprtr258/fun"
)

type PathBuf string

type Options struct {
	value uint32
	files []PathBuf
}

func main() {
	value := cli.Argument[uint32](cli.Long("value").
		Help("Mysterious value"),
		"VAL", func(s string) (uint32, error) {
			u, err := strconv.ParseUint(s, 10, 32)
			return uint32(u), err
		}).
		Parser().
		Fallback(42).
		Parser()
	files := cli.Many(cli.Positional[PathBuf]("FILE", func(s string) (PathBuf, error) {
		return PathBuf(s), nil
	}).Parser(), false)
	opts := cli.Map(cli.Pair(value, files).Parser(), func(x fun.Pair[uint32, []PathBuf]) Options { return Options{x.K, x.V} }).Run()

	fmt.Println(opts)
}
