// A simple combinatoric parser with extra metainformation attached, see --help
package examples

import (
	"fmt"
	"os"

	"github.com/rprtr258/cli"
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
		Help("Width of the rectangle"), "PX", cli.ParseUint)

	height := cli.Argument[uint](cli.Short('h').
		Long("height").
		Help("Height of the rectangle"), "PX", cli.ParseUint)

	rect := cli.Map2(width, height, func(width, height uint) Rect { return Rect{width, height} }).
		GroupHelp("Rectangle is defined by width and height in meters").
		Optional()

	verbose := cli.
		Short('v').
		Long("verbose").
		Help("Print computation steps").
		Switch()

	opt := cli.Map2(verbose, rect, func(verbose bool, rect cli.Opt[Rect]) Out { return Out{rect, verbose} }).
		ToOptions().
		Description("This program calculates rectangle's area").
		Header("vvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvv").
		Footer("^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^").
		Run(os.Args)
	fmt.Println(opt)
}
