// You don't need to import this library in order to use it via struct tags, it will look ugly though
package examples

import (
	"fmt"
	"os"

	"github.com/rprtr258/cli"
)

func NoImport() {
	type Out struct {
		debug bool
		speed float64
	}

	// A flag, true if used in the command line. Can be required, this one is optional
	debug := cli.
		Short('d').
		Long("debug").
		Help("Activate debug mode").
		Switch()

	// an argument, parsed and with default value
	speed := cli.Short('s').
		Long("speed").
		Help("Set speed").
		Argument("SPEED", cli.ParseFloat).
		Fallback(42.0)

	// packing things in a struct assumes parser for each field is in scope.
	opt := cli.Map2(debug, speed, func(debug bool, speed float64) Out { return Out{debug, speed} }).ToOptions().Run(os.Args)

	fmt.Println(opt)
}
