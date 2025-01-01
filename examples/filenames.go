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
	parser := cli.Many(cli.Positional("FILE", parseString).Parser().
		CompleteShell(cli.ShellCompFile(cli.Valid("*.(md|toml)"))).Parser(), false)

	r := parser.Parser().ToOptions().Run(os.Args)
	fmt.Println(r)
}
