// When implementing a cargo subcommand parser needs to be able to skip the first argument which
// is always the same as the executable name. For this example executable name is `cargo-cmd` so
// first argument would be `cmd`. A way to support both cases - when it's present and it's absent
// would be to use a combination of `literal`, `optional` and `hide`.
// `bpaf` also provides a helper `cargo_helper` that does exactly this.
package main

import (
	"fmt"
	"strconv"

	"github.com/rprtr258/cli"
	"github.com/rprtr258/fun"
)

type Opts struct {
	width  uint
	height uint
}

func parseUint(s string) (uint, error) {
	u, err := strconv.ParseUint(s, 10, 64)
	return uint(u), err
}

func main() {
	// defining a parser in a usual way
	width := cli.Argument(cli.Short('w'), "WIDTH", parseUint).Parser().Fallback(10)
	height := cli.Argument(cli.Short('h'), "HEIGHT", parseUint).Parser().Fallback(10)
	parser := cli.Map(cli.Pair(width.Parser(), height.Parser()).Parser(), func(x fun.Pair[uint, uint]) Opts { return Opts{x.K, x.V} })

	cmd := cli.Optional(cli.Literal("cmd"), false).Parser().Hide()
	combined_parser := cli.Map(cli.Pair(cmd, parser).Parser(), func(x fun.Pair[cli.Option[struct{}], Opts]) Opts { return x.V })

	opts := combined_parser.ToOptions().Run()

	fmt.Println(opts)
}
