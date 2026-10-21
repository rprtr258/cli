// How to require presence of at least N values,
//
// This program accepts "-f -f -f" or "-fffff" but not "-f"
package examples

import (
	"fmt"
	"os"

	"github.com/rprtr258/cli"
)

func AtLeastTwo() {
	opt := cli.Many(cli.ReqFlag(cli.Short('f'), struct{}{}).Parser()).Parser().
		Guard(func(x []struct{}) bool { return len(x) >= 2 }, "at least two arguments are required").
		Parser().
		ToOptions().Run(os.Args)

	fmt.Println(opt)
}
