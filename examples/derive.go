// A basic [Construct] example: the whole command line interface declared as one
// typed struct with `cli` tags.
package examples

import (
	"fmt"
	"os"

	"github.com/rprtr258/cli"
)

type Opts struct {
	// Activate debug mode
	Debug bool `cli:"short=d,long=debug,help='Activate debug mode'"`

	// number of occurrences of the v/verbose flag capped at 3
	Verbose cli.Count `cli:"short=v,long=verbose,help='Increase the verbosity; you can specify it up to 3 times either as -v -v -v or as -vvv'"`

	// Set speed
	Speed float64 `cli:"long=speed,metavar=SPEED,default=42"`

	// Output file
	Output string

	Files []string
}

func Main() {
	parser := cli.Construct[Opts]().ToOptions()
	opts := parser.Run(os.Args)
	fmt.Println(opts)
}
