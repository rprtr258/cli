// `positional_derive.go` - struct with named values that uses a positional argument.
package examples

import (
	"fmt"
	"os"

	"github.com/rprtr258/cli"
)

type positionalOpts struct {
	// Mysterious value
	Value uint32 `cli:"long=value,metavar=VAL,default=42,help='Mysterious value'"`
	Files []string
}

func PositionalDerive() {
	opts := cli.Construct[positionalOpts]().ToOptions().Run(os.Args)
	fmt.Println(opts)
}
