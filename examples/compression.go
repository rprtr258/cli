// Numeric flags similar to zip compression levels, accepts -1 to -9, produces usize
package examples

import (
	"fmt"
	"os"
	"strconv"

	"github.com/rprtr258/cli"
	"github.com/rprtr258/fun"
)

func Compression() {
	opts := cli.
		Any("COMP", func(s string) (uint, bool) {
			x, err := strconv.ParseInt(s, 10, 64)
			return uint(-x), err == nil && -9 <= x && x <= -1
		}, true).
		Metavar(cli.NewDocFromPairs([]fun.Pair[string, cli.Style]{
			{K: "-1", V: cli.StyleLiteral},
			{K: " to ", V: cli.StyleText},
			{K: "-9", V: cli.StyleLiteral},
		})).
		Help("Compression level").
		ToOptions().
		Run(os.Args)

	fmt.Println(opts)
}
