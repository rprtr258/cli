// `app_name.go` shows using the application name (taken from argv[0]) as part of
// the parsed structure, alongside regular arguments.
package examples

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/rprtr258/cli"
)

func AppName() {
	type Opts struct {
		user    string
		appname string
	}

	user := cli.Short('u').Long("user").
		Help("Specify user name").
		// you can specify exact type argument should produce
		// for as long as it implements FromStr
		Argument("NAME", cli.ParseString)

	// the counterpart of bpaf's `appname()` battery: a pure parser that
	// evaluates to the application name at run time
	appname := cli.PureWith(func() (string, error) {
		return filepath.Base(os.Args[0]), nil
	})

	opts := cli.Map2(user, appname, func(user, appname string) Opts {
		return Opts{user, appname}
	}).
		ToOptions().
		Run(os.Args)

	fmt.Println(opts)
}
