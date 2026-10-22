// When implementing a cargo subcommand parser needs to be able to skip the first argument which
// is always the same as the executable name. For this example executable name is `cargo-cmd` so
// first argument would be `cmd`. A way to support both cases - when it's present and it's absent
// would be to use a combination of `literal`, `optional` and `hide`.
// The library also provides a helper `cargo_helper` that does exactly this.
package examples

import (
	"fmt"
	"os"

	"github.com/rprtr258/cli"
	"github.com/rprtr258/fun"
)

func CargoCmd() {
	type Opts struct {
		width  uint
		height uint
	}

	// defining a parser in a usual way
	width := cli.Argument(cli.Short('w'), "WIDTH", parseUint).Fallback(10)
	height := cli.Argument(cli.Short('h'), "HEIGHT", parseUint).Fallback(10)
	parser := cli.Pair(width, height).Map(func(x fun.Pair[uint, uint]) Opts { return Opts{x.K, x.V} })

	cmd := cli.Literal("cmd").Optional().Hide()
	combined_parser := cli.Pair(cmd, parser).Map(func(x fun.Pair[cli.Opt[struct{}], Opts]) Opts { return x.V })

	opts := combined_parser.ToOptions().Run(os.Args)

	fmt.Println(opts)
}
