// `very_custom_usage.go` - a way to customize usage for a nested command.
//
// Usually you would go with generated usage or by overriding it using the `usage`
// annotation to the top level or command level annotation. By taking advantage of
// command being just a set of options with its own help message and a custom prefix
// you can override the usage with an arbitrary string, including one generated at
// runtime.
package examples

import (
	"fmt"
	"os"

	"github.com/rprtr258/cli"
)

type veryCustomUsageCmd struct {
	opt bool
}

func VeryCustomUsage() {
	opt := cli.Long("opt").
		Switch().
		ToOptions().
		Usage(cli.NewDoc("A very custom usage goes here")).
		Command("cmd").
		Map(func(opt bool) veryCustomUsageCmd { return veryCustomUsageCmd{opt} }).
		ToOptions().
		Run(os.Args)

	fmt.Println(opt)
}
