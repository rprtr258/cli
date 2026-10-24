// To parse comma separated values it's easier to treat them as strings
package examples

import (
	"fmt"
	"os"

	"github.com/rprtr258/cli"
	"github.com/rprtr258/fun"
)

func ManyCommaSeparatedArgs() {
	// --ports 1,2,3 --ports 4,5 => [1,2,3,4,5]
	args := cli.Long("ports").
		Help("Comma separated list of ports").
		Argument("PORTS", splitAndParseU16).
		Many().
		Map(func(nested [][]uint16) []uint16 { return fun.Flatten(nested...) }).
		ToOptions().
		Run(os.Args)

	fmt.Println(args)
}
