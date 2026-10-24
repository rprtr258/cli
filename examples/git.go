// `git.rs` serves as a demonstration of how to use subcommands,
// as well as a demonstration of adding documentation to subcommands.
//
// Note, this "fetch" command uses fallback to inner description to get the help message, "add"
// uses explicit override with the same value.
package examples

import (
	"fmt"
	"os"

	"github.com/rprtr258/cli"
)

func Git() {
	type Opt func(
		Fetch func(
			dry_run bool,
			all bool,
			repository string,
		),
		Add func(
			interactive bool,
			all bool,
			files []string,
		),
	)

	OptFetch := func(dry_run bool, all bool, repository string) Opt {
		return func(
			f func(dry_run bool, all bool, repository string),
			_ func(interactive bool, all bool, files []string),
		) {
			if f != nil {
				f(dry_run, all, repository)
			}
		}
	}
	OptAdd := func(interactive bool, all bool, files []string) Opt {
		return func(
			_ func(dry_run bool, all bool, repository string),
			f func(interactive bool, all bool, files []string),
		) {
			if f != nil {
				f(interactive, all, files)
			}
		}
	}

	dry_run := cli.Long("dry_run").Switch()
	all := cli.Long("all").Switch()
	repository := cli.Positional("SRC", func(s string) (string, error) {
		return s, nil
	}).Fallback("origin")
	fetch_cmd := cli.Map3(dry_run, all, repository, func(dry_run, all bool, repository string) Opt { return OptFetch(dry_run, all, repository) }).
		ToOptions().
		Description("fetches branches from remote repository").
		Command("fetch")

	interactive := cli.Short('i').Switch()
	all2 := cli.Long("all").Switch()
	files := cli.Positional("FILE", cli.ParseString).Parser().Many()
	add_cmd := cli.Map3(interactive, all2, files, func(interactive, all bool, files []string) Opt { return OptAdd(interactive, all, files) }).
		ToOptions().
		Description("add files to the staging area").
		Command("add").
		Help("add files to the staging area")

	opt := cli.Choice(fetch_cmd, add_cmd).
		ToOptions().
		Description("The stupid content tracker").
		Run(os.Args)

	fmt.Println(opt)
}
