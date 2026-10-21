// A simple combinatoric parser with extra metainformation attached, see --help
package examples

import (
	"fmt"
	"os"

	"github.com/rprtr258/cli"
	"github.com/rprtr258/fun"
)

func Rectangle() {
	type Rect struct {
		width  uint
		height uint
	}

	type Out struct {
		rect    cli.Opt[Rect]
		verbose bool
	}

	width := cli.Argument[uint](cli.Short('w').
		Long("width").
		Help("Width of the rectangle"), "PX", parseUint)

	height := cli.Argument[uint](cli.Short('h').
		Long("height").
		Help("Height of the rectangle"), "PX", parseUint)

	rect := cli.Optional(cli.Pair(width, height).
		Map(func(x fun.Pair[uint, uint]) Rect { return Rect{x.K, x.V} }).
		GroupHelp("Rectangle is defined by width and height in meters"))

	verbose := cli.
		Short('v').
		Long("verbose").
		Help("Print computation steps").
		Switch()

	opt := cli.Pair(verbose, rect).Map(func(x fun.Pair[bool, cli.Opt[Rect]]) Out { return Out{x.V, x.K} }).
		ToOptions().
		Description("This program calculates rectangle's area").
		Header("vvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvv").
		Footer("^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^").
		Run(os.Args)
	fmt.Println(opt)
}
