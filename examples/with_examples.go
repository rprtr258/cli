// Documented usage examples
//
// Examples are shown in the "Examples:" section of the `--help` output and in
// the EXAMPLES section of the manpage. Every example is validated against the
// parser: if an example stops parsing after a parser change, the program
// panics with a "cli usage BUG" message instead of shipping stale docs.
package examples

import (
	"fmt"
	"os"

	"github.com/rprtr258/cli"
	"github.com/rprtr258/fun"
)

func WithExamples() {
	opt := cli.Map2(
		cli.Short('r').
			Long("recursive").
			Help("Copy directories recursively").
			Switch(),
		cli.Map2(
			cli.Positional("SRC", cli.ParseString).Help("Source path"),
			cli.Positional("DST", cli.ParseString).Help("Destination path"),
			func(src, dst string) string { return src + " -> " + dst },
		),
		func(recursive bool, transfer string) string { return transfer + fun.IF(recursive, " recursively", "") },
	).ToOptions().
		Description("Copy files and directories").
		Example("Copy a file to a directory", "file.txt backup/").
		Example("Copy a directory recursively", "--recursive src/ dst/").
		Run(os.Args)

	fmt.Println(opt)
}
