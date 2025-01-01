// You don't need to import bpaf in order to use it, it will look ugly though
package main

import (
	"fmt"
	"strconv"

	"github.com/rprtr258/cli"
	"github.com/rprtr258/fun"
)

type Out struct {
	debug bool
	speed float64
}

func main() {
	// A flag, true if used in the command line. Can be required, this one is optional
	debug := cli.
		Short('d').
		Long("debug").
		Help("Activate debug mode").
		Switch()

	// an argument, parsed and with default value
	speed := cli.Argument(cli.Short('s').
		Long("speed").
		Help("Set speed"),
		"SPEED", func(s string) (float64, error) {
			return strconv.ParseFloat(s, 64)
		}).
		Parser().
		Fallback(42.0).
		Parser()

	// packing things in a struct assumes parser for each field is in scope.
	opt := cli.Map(cli.Pair(debug, speed).Parser(), func(x fun.Pair[bool, float64]) Out { return Out{x.K, x.V} }).ToOptions().Run()

	fmt.Println(opt)
}
