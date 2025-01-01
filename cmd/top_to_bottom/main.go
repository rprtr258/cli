// A somewhat comprehensive example of a typical `bpaf` usage.
// Since those can be functions - order doesn't really matter
package main

import (
	"fmt"
	"os"
	"strconv"

	"github.com/rprtr258/cli"
)

type PathBuf string

func parsePathBuf(s string) (PathBuf, error) {
	return PathBuf(s), nil
}

type Out struct {
	debug            bool
	verbose          uint
	speed            float64
	output           PathBuf
	nb_cars          uint32
	files_to_process []PathBuf
}

// A flag, true if used in the command line. Can be required, this one is optional
var debug cli.Parser[bool] = cli.
	Short('d').
	Long("debug").
	Help("Activate debug mode").
	Switch().Parser()

// number of occurrences of the v/verbose flag capped at 3
var verbose cli.Parser[uint] = cli.
	Map(cli.Many(cli.ReqFlag(cli.Short('v').
		Long("verbose").
		Help("Increase the verbosity\nYou can specify it up to 3 times\neither as -v -v -v or as -vvv"),
		struct{}{}).Parser(), false).Parser(), func(xs []struct{}) uint { return uint(len(xs)) }).
	Guard(func(x uint) bool { return x <= 3 }, "It doesn't get any more verbose than this").
	Parser()

// an argument, parsed and with default value
var speed cli.Parser[float64] = cli.
	Argument(cli.Short('s').
		Long("speed").
		Help("Set speed"), "SPEED", func(s string) (float64, error) {
		return strconv.ParseFloat(s, 64)
	}).
	Parser().
	Fallback(42.0).
	Parser()

var output cli.Parser[PathBuf] = cli.
	Argument(cli.Short('o').
		Long("output").
		Help("output file"),
		"OUTPUT", parsePathBuf).
	Parser()

// no magical name transmogrifications.
var nb_cars cli.Parser[uint32] = cli.
	Argument(cli.Short('n').
		Long("nb-cars"),
		"N", func(s string) (uint32, error) {
			u, err := strconv.ParseUint(s, 10, 32)
			return uint32(u), err
		}).
	Parser()

var files_to_process cli.Parser[[]PathBuf] = cli.
	Many(cli.Argument(cli.Short('f').
		Long("file").
		Help("File to process"), "FILE", parsePathBuf).Parser(), false).Parser()

func main() {
	// packing things in a struct assumes parser for each field is in scope.
	opt := (cli.Map6(
		debug,
		verbose,
		speed,
		output,
		nb_cars,
		files_to_process,
		func(debug bool, verbose uint, speed float64, output PathBuf, nb_cars uint32, files_to_process []PathBuf) Out {
			return Out{debug, verbose, speed, output, nb_cars, files_to_process}
		})).ToOptions().Run(os.Args)
	fmt.Println(opt)
}
