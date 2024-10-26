// `git.rs` serves as a demonstration of how to use subcommands,
// as well as a demonstration of adding documentation to subcommands.
//
// Note, this "fetch" command uses fallback to inner description to get the help message, "add"
// uses explicit override with the same value.
package main

import (
	"fmt"

	"github.com/rprtr258/cli"
	"github.com/rprtr258/fun"
)

type PathBuf string

type Opt func(
	Fetch func(
		dry_run bool,
		all bool,
		repository string,
	),
	Add func(
		interactive bool,
		all bool,
		files []PathBuf,
	),
)

func OptFetch(dry_run bool, all bool, repository string) Opt {
	return func(
		f func(dry_run bool, all bool, repository string),
		_ func(interactive bool, all bool, files []PathBuf),
	) {
		if f != nil {
			f(dry_run, all, repository)
		}
	}
}
func OptAdd(interactive bool, all bool, files []PathBuf) Opt {
	return func(
		_ func(dry_run bool, all bool, repository string),
		f func(interactive bool, all bool, files []PathBuf),
	) {
		if f != nil {
			f(interactive, all, files)
		}
	}
}

func main() {
	dry_run := cli.Long("dry_run").Switch()
	all := cli.Long("all").Switch()
	repository := cli.Positional("SRC", func(s string) (string, error) {
		return s, nil
	}).Parser().Fallback("origin").Parser()
	fetch := cli.Map(cli.Pair(cli.Pair(dry_run, all).Parser(), repository).Parser(), func(x fun.Pair[fun.Pair[bool, bool], string]) Opt { return OptFetch(x.K.K, x.K.V, x.V) }).
		ToOptions().
		Descr("fetches branches from remote repository")

	fetch_cmd := fetch.Command("fetch")

	interactive := cli.Short('i').Switch()
	all2 := cli.Long("all").Switch()
	files := cli.Many(cli.Positional("FILE", func(s string) (PathBuf, error) {
		return PathBuf(s), nil
	}).Parser())
	add := cli.Map(cli.Pair(cli.Pair(interactive, all2).Parser(), files).Parser(), func(x fun.Pair[fun.Pair[bool, bool], []PathBuf]) Opt { return OptAdd(x.K.K, x.K.V, x.V) }).
		ToOptions().
		Descr("add files to the staging area")

	add_cmd := add.Command("add").Help("add files to the staging area")

	opt := cli.Choice(fetch_cmd.Parser(), add_cmd.Parser()).
		ToOptions().
		Descr("The stupid content tracker").
		Run()

	fmt.Println(opt)
}
