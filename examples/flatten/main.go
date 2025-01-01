// All the flags don't have to live in the same structure, this example uses non derive version.
// with derive API you would use `external` annotation
package flatten

import (
	"fmt"
	"os"

	"github.com/rprtr258/cli"
	"github.com/rprtr258/fun"
)

type DaemonOpts struct {
	// daemon user
	user string

	// daemon group
	group string
}

type Cmdline struct {
	// switch verbosity on
	verbose     bool
	daemon_opts DaemonOpts
}

func Main() {
	verbose := cli.Short('v').Help("switch verbosity on").Switch()
	user := cli.Argument(cli.Short('u').Help("daemon user"), "USER", func(s string) (string, error) { return s, nil }).Parser()
	group := cli.Argument(cli.Short('g').Help("daemon group"), "GROUP", func(s string) (string, error) { return s, nil }).Parser()
	daemon_opts := cli.Map(cli.Pair(user, group).Parser(), func(x fun.Pair[string, string]) DaemonOpts { return DaemonOpts{x.K, x.V} })
	opt := cli.Map(cli.Pair(verbose, daemon_opts).Parser(), func(x fun.Pair[bool, DaemonOpts]) Cmdline { return Cmdline{x.K, x.V} }).ToOptions().Run(os.Args)
	fmt.Println(opt)
}
