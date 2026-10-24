// The same command line interface as basic.go, defined with a struct and tags
// instead of hand-composed parsers.
package examples

import (
	"fmt"
	"os"

	"github.com/rprtr258/cli"
)

func ConstructStruct() {
	type Out struct {
		Debug            bool         `cli:"short=d,long=debug,help='Activate debug mode'"`
		Verbose          cli.Count    `cli:"short=v,long=verbose,help='Increase the verbosity, up to 3 times'"`
		Speed            float64      `cli:"short=s,long=speed,metavar=SPEED,default=42"`
		Output           string       `cli:"short=o,long=output,metavar=OUTPUT,help='output file'"`
		NbCars           uint32       `cli:"short=n,long=nb-cars,metavar=N,help='Number of items to process'"`
		FilesToProcess   []string     // positional, zero or more
		MaxFileSizeBytes cli.Opt[int] `cli:"long=max-file-size"`
	}

	out := cli.Construct[Out]().
		Parser().
		ToOptions().
		Description("This is a description").
		Run(os.Args)
	fmt.Println(out)
}
