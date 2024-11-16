// This example shows how to use shell completion to ask for
// a file with one of two extensions. If you want to specify just one
// extension having it as something like "*.rs" is good enough
package main

import (
	"fmt"

	"github.com/rprtr258/cli"
)

type PathBuf string

func main() {
	parser := cli.Many(cli.Positional("FILE", func(s string) (PathBuf, error) {
		return PathBuf(s), nil
	}).Parser().
		CompleteShell(cli.ShellCompFile(cli.Valid("*.(md|toml)"))).Parser())

	r := parser.Parser().Run()
	fmt.Println(r)
}
