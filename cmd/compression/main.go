// Numeric flags similar to zip compression levels, accepts -1 to -9, produces usize
package main

import (
	"fmt"
	"strconv"

	"github.com/rprtr258/cli"
	"github.com/rprtr258/fun"
)

func compression() cli.Parser[uint] {
	return cli.Any("COMP", func(s string) cli.Option[uint] {
		x, err := strconv.ParseInt(s, 10, 64)
		if err == nil && -9 <= x && x <= -1 {
			return cli.Valid(uint(-x))
		} else {
			return cli.Invalid[uint]()
		}
	}).
		Metavar(cli.NewDocFromPairs([]fun.Pair[string, cli.Style]{
			{"-1", cli.StyleLiteral},
			{" to ", cli.StyleText},
			{"-9", cli.StyleLiteral},
		})).
		Help("Compression level").
		Anywhere().
		Parser()
}

func main() {
	opts := compression().ToOptions().Run()
	fmt.Println(opts)
}
