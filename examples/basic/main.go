// A somewhat comprehensive example of a typical combinatoric `bpaf` usage in combinatoric mode
package main

import (
	"fmt"
	"os"
	"strconv"

	"github.com/rprtr258/cli"
	"github.com/rprtr258/fun"
)

type PathBuf string

type Out struct {
	debug            bool
	verbose          uint
	speed            float64
	output           PathBuf
	nb_cars          uint32
	files_to_process []PathBuf
}

func opts() *cli.OptionParser[Out] {
	// A flag, true if used in the command line. Can be required, this one is optional
	debug := cli.Short('d'). // start with a short name
					Long("debug").               // also add a long name
					Help("Activate debug mode"). // and a help message to use
					Switch()                     // turn this into a switch

	// number of occurrences of the v/verbose flag capped at 3 with an error here but you can also
	// use `max` inside `map`
	verbose := cli.Map(cli.Many(cli.ReqFlag(cli.Short('v').
		Long("verbose").
		Help("Increase the verbosity\n You can specify it up to 3 times\n either as -v -v -v or as -vvv"), struct{}{}).Parser(), false).Parser(),
		func(xs []struct{}) uint { return uint(len(xs)) }).
		Guard(func(x uint) bool { return x <= 3 }, "It doesn't get any more verbose than this")

	// an argument, parsed and with default value
	speed := cli.Argument[float64](cli.Short('s').
		Long("speed").
		Help("Set speed"), "SPEED", func(s string) (float64, error) {
		return strconv.ParseFloat(s, 64)
	}).Parser(). // you can specify a type to parse
			Fallback(42.0).
			DisplayFallback()

	output := cli.Argument[PathBuf](cli.Short('o').
		Long("output").
		Help("output file"),
		"OUTPUT", func(s string) (PathBuf, error) {
			return PathBuf(s), nil
		}). // but it's optional when rustc can derive it
		Parser().
		CompleteShell(cli.ShellCompFile(cli.Invalid[string]()))

	// no magical name transmogrifications in combinatoric API,
	nb_cars := cli.Argument[uint32](cli.Short('n').
		Long("nb-cars").
		Help("Number of items to process"),
		"N", func(s string) (uint32, error) {
			u, err := strconv.ParseUint(s, 10, 32)
			return uint32(u), err
		}).
		Parser().
		Fallback(1).
		DisplayFallback()

	// a parser that consumes one argument
	// you can build the inner parser in one go or as multiple steps giving each step a name
	// you can also add some static shell completion functionality
	file_to_proces := cli.Argument[PathBuf](cli.Short('f').
		Long("file").
		Help("File to process"),
		"FILE", func(s string) (PathBuf, error) {
			return PathBuf(s), nil
		}).
		Parser().
		CompleteShell(cli.ShellCompFile(cli.Valid("*.rs")))

	files_to_process := cli.Many(file_to_proces.Parser(), false)

	// packing things in a struct assumes parser for each field is in scope.
	// TODO: codgen on Out struct
	return cli.Map(
		cli.Pair(
			cli.Pair(
				cli.Pair(
					debug,
					verbose.Parser(),
				).Parser(),
				cli.Pair(
					speed.Parser(),
					output.Parser(),
				).Parser(),
			).Parser(),
			cli.Pair(
				nb_cars.Parser(),
				files_to_process,
			).Parser(),
		).Parser(), func(x fun.Pair[fun.Pair[fun.Pair[bool, uint], fun.Pair[float64, PathBuf]], fun.Pair[uint32, []PathBuf]]) Out {
			return Out{
				debug:            x.K.K.K,
				verbose:          x.K.K.V,
				speed:            x.K.V.K,
				output:           x.K.V.V,
				nb_cars:          x.V.K,
				files_to_process: x.V.V,
			}
		}).
		ToOptions().
		Descr("This is a description")
}

func Main() {
	opts := opts().FallbackToUsage().Run(os.Args)
	fmt.Println(opts)
}
