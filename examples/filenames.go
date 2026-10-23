// This example shows how to use shell completion to ask for
// a file with one of two extensions. If you want to specify just one
// extension having it as something like "*.rs" is good enough
package examples

import (
	"fmt"
	"os"

	"github.com/rprtr258/cli"
)

func Filenames() {
	parser := cli.Positional("FILE", cli.ParseString).
		Parser().
		CompleteShell(cli.ShellCompFile(cli.Valid("*.(md|toml)"))).
		Many()

	r := parser.ToOptions().Run(os.Args)
	fmt.Println(r)
}
