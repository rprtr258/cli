// All the flags don't have to live in the same structure: this example composes
// parsers from separate pieces and combines them with Map2.
package examples

import (
	"fmt"
	"os"

	"github.com/rprtr258/cli"
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
	user := cli.Short('u').Help("daemon user").Argument("USER", func(s string) (string, error) { return s, nil })
	group := cli.Short('g').Help("daemon group").Argument("GROUP", func(s string) (string, error) { return s, nil })
	daemon_opts := cli.Map2(user, group, func(user, group string) DaemonOpts { return DaemonOpts{user, group} })
	opt := cli.Map2(verbose, daemon_opts, func(verbose bool, daemon_opts DaemonOpts) Cmdline { return Cmdline{verbose, daemon_opts} }).ToOptions().Run(os.Args)
	fmt.Println(opt)
}
