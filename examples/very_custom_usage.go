// `very_custom_usage.rs` - a way to customize usage for a nested command.
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
	cmd := cli.Long("opt").Switch().
		Parser().
		ToOptions().
		Usage(cli.NewDoc("A very custom usage goes here")).
		Command("cmd").
		Map(func(opt bool) veryCustomUsageCmd { return veryCustomUsageCmd{opt: opt} }).
		Parser()

	fmt.Println(cmd.ToOptions().Run(os.Args))
}
