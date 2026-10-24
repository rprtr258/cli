// wtf.go - a kitchen-sink example ported from the old struct-based API:
// top-level subcommands with a nested `docker image` command tree, command
// aliases, switches, required and optional arguments, repeated arguments,
// default values, and a custom-parsed type.
package examples

import (
	"fmt"
	"os"

	"github.com/rprtr258/cli"
)

// wtfPoint is a custom type parsed from an "x,y" argument, mirroring the
// MarshalFlag/UnmarshalFlag pair of the original struct-based example.
type wtfPoint struct {
	x, y int
}

func wtfParsePoint(s string) (wtfPoint, error) {
	var p wtfPoint
	if _, err := fmt.Sscanf(s, "%d,%d", &p.x, &p.y); err != nil {
		return wtfPoint{}, fmt.Errorf(`expected two numbers separated by a comma, e.g. "10,20", but got %s`, s)
	}

	return p, nil
}

// wtfExampleExtra groups the non-boolean options of the example command, so
// they and the positionals fit into the Map3/Map6 combinators.
type wtfExampleExtra struct {
	num  cli.Opt[int]
	strs []string
	file string
}

func Wtf() {
	// $ app add [-a] FILE...
	add := cli.Map2(
		cli.Short('a').Long("all").Help("Add all files").Switch(),
		cli.Positional("FILE", cli.ParseString).Many(),
		func(all bool, files []string) string {
			return fmt.Sprintf("Adding (all=%v): %#v", all, files)
		},
	).
		ToOptions().
		Description("Add a file").
		Command("add")

	// $ app rm [-f] FILE...
	rm := cli.Map2(
		cli.Short('f').Long("force").Help("Force removal of files").Switch(),
		cli.Positional("FILE", cli.ParseString).Many(),
		func(force bool, files []string) string {
			return fmt.Sprintf("Removing (force=%v): %#v", force, files)
		},
	).
		ToOptions().
		Description("Remove a file").
		Command("rm")

	// $ app example -n NAME [flags] FILE...
	// exercises every kind of named option: a switch, an argument with a
	// default, a required argument, an optional argument, a repeated
	// argument, and an argument of a custom-parsed type
	extra := cli.Map3(
		cli.Short('p').
			Help("A pointer to an integer").
			Argument("NUM", cli.ParseInt).
			Optional(),
		cli.Short('s').
			Help("A slice of strings").
			Argument("STR", cli.ParseString).
			Many(),
		cli.Short('f').Long("file").Help("A file").Argument("FILE", cli.ParseString).FallbackDisplay("-"),
		func(num cli.Opt[int], strs []string, file string) wtfExampleExtra {
			return wtfExampleExtra{num: num, strs: strs, file: file}
		},
	)
	example := cli.Map6(
		cli.Short('v').Long("verbose").Help("Show verbose debug information").Switch(),
		cli.Long("offset").Help("Offset").Argument("OFFSET", cli.ParseUint).FallbackDisplay(0),
		cli.Short('n').Long("name").Help("A name").Argument("NAME", cli.ParseString),
		cli.Long("point").Help("A x,y point").Argument("POINT", wtfParsePoint).FallbackDisplay(wtfPoint{1, 2}),
		extra,
		cli.Positional("FILE", cli.ParseString).Many(),
		func(verbose bool, offset uint, name string, p wtfPoint, extra wtfExampleExtra, files []string) string {
			return fmt.Sprintf("Example(verbose=%v, offset=%d, name=%q, point=%d,%d, file=%q, num=%v, strs=%#v, files=%#v)",
				verbose, offset, name, p.x, p.y, extra.file, extra.num, extra.strs, files)
		},
	).
		ToOptions().
		Description("Command example with various flag types").
		Command("example")

	// $ app docker image list [-v] [-o FORMAT] NAME...
	// nested commands with a hidden long alias `ls` for the list command
	list := cli.Map3(
		cli.Short('v').Long("verbose").Help("Verbose output").Switch(),
		cli.Short('o').Long("output").Help("Output format").Argument("FORMAT", cli.ParseString).FallbackDisplay("json"),
		cli.Positional("NAME", cli.ParseString).Many(),
		func(verbose bool, output string, names []string) string {
			return fmt.Sprintf("Listing images (verbose=%v, output=%s): %#v", verbose, output, names)
		},
	).
		ToOptions().
		Description("List images").
		Command("list").
		Long("ls")
	image := list.
		ToOptions().
		Description("Manage images").
		Command("image")
	docker := image.
		ToOptions().
		Description("Manage Docker").
		Command("docker")

	opts := cli.Choice(add, rm, example, docker).
		ToOptions().
		Description("A kitchen-sink example").
		Run(os.Args)
	fmt.Println(opts)
}
