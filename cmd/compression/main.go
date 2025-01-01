// Numeric flags similar to zip compression levels, accepts -1 to -9, produces usize
package main

import (
	"fmt"
	"strconv"

	"github.com/rprtr258/cli"
	"github.com/rprtr258/fun"
)

var compression cli.Parser[uint] = cli.
	Any("COMP", func(s string) (uint, bool) {
		x, err := strconv.ParseInt(s, 10, 64)
		return uint(-x), err == nil && -9 <= x && x <= -1
	}, true).
	Metavar(cli.NewDocFromPairs([]fun.Pair[string, cli.Style]{
		{"-1", cli.StyleLiteral},
		{" to ", cli.StyleText},
		{"-9", cli.StyleLiteral},
	})).
	Help("Compression level").
	Parser()

func main() {
	opts := compression.ToOptions().Run()
	fmt.Println(opts)
}
