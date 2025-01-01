// A simple combinatoric parser with extra metainformation attached, see --help
package rectangle

import (
	"fmt"
	"os"
	"strconv"

	"github.com/rprtr258/cli"
	"github.com/rprtr258/fun"
)

type Out struct {
	rect    cli.Option[Rect]
	verbose bool
}

type Rect struct {
	width  uint
	height uint
}

func parseUint(s string) (uint, error) {
	u, err := strconv.ParseUint(s, 10, 64)
	return uint(u), err
}

func Main() {
	width := cli.Argument[uint](cli.Short('w').
		Long("width").
		Help("Width of the rectangle"),
		"PX", parseUint).Parser()

	height := cli.Argument[uint](cli.Short('h').
		Long("height").
		Help("Height of the rectangle"),
		"PX", parseUint).Parser()

	rect := cli.Optional(cli.Map(cli.Pair(width, height).Parser(), func(x fun.Pair[uint, uint]) Rect { return Rect{x.K, x.V} }).
		GroupHelp("Rectangle is defined by width and height in meters").Parser(), false).Parser()

	verbose := cli.Short('v').
		Long("verbose").
		Help("Print computation steps").
		Switch()

	opt := cli.Map(cli.Pair(verbose, rect).Parser(), func(x fun.Pair[bool, cli.Option[Rect]]) Out { return Out{x.V, x.K} }).
		ToOptions().
		Descr("This program calculates rectangle's area").
		Header("vvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvv").
		Footer("^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^").
		Run(os.Args)
	fmt.Println(opt)
}
