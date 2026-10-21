// All the flags don't have to live in the same structure, this example uses non derive version.
// with derive API you would use `external` annotation
package examples

import (
	"fmt"
	"os"

	"github.com/rprtr258/cli"
	"github.com/rprtr258/fun"
)

func Flatten() {
	type DaemonOpts struct {
		user  string // daemon user
		group string // daemon group
	}

	type Cmdline struct {
		// switch verbosity on
		verbose     bool
		daemon_opts DaemonOpts
	}

	verbose := cli.Short('v').Help("switch verbosity on").Switch()
	user := cli.Argument(cli.Short('u').Help("daemon user"), "USER", func(s string) (string, error) { return s, nil })
	group := cli.Argument(cli.Short('g').Help("daemon group"), "GROUP", func(s string) (string, error) { return s, nil })
	daemon_opts := cli.Pair(user, group).Map(func(x fun.Pair[string, string]) DaemonOpts { return DaemonOpts{x.K, x.V} })
	opt := cli.Pair(verbose, daemon_opts).Map(func(x fun.Pair[bool, DaemonOpts]) Cmdline { return Cmdline{x.K, x.V} }).ToOptions().Run(os.Args)
	fmt.Println(opt)
}
