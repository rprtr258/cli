// A somewhat comprehensive example of a typical usage of this library.
// Since those can be functions - order doesn't really matter
package examples

import (
	"fmt"
	"os"

	"github.com/rprtr258/cli"
)

func TopToBottom() {
	type Out struct {
		debug            bool
		verbose          uint
		speed            float64
		output           string
		nb_cars          uint32
		files_to_process []string
	}

	// A flag, true if used in the command line. Can be required, this one is optional
	debug := cli.
		Short('d').
		Long("debug").
		Help("Activate debug mode").
		Switch()

	// number of occurrences of the v/verbose flag capped at 3
	verbose := cli.
		Short('v').
		Long("verbose").
		Help("Increase the verbosity\nYou can specify it up to 3 times\neither as -v -v -v or as -vvv").
		ReqFlag(struct{}{}).
		Many().
		Map(func(xs []struct{}) uint { return uint(len(xs)) }).
		Guard(func(x uint) bool { return x <= 3 }, "It doesn't get any more verbose than this")

	// an argument, parsed and with default value
	speed := cli.Short('s').
		Long("speed").
		Help("Set speed").
		Argument("SPEED", cli.ParseFloat).
		Fallback(42.0)

	output := cli.Short('o').
		Long("output").
		Help("output file").
		Argument("OUTPUT", cli.ParseString)

	// no magical name transmogrifications.
	nb_cars := cli.Short('n').
		Long("nb-cars").
		Argument("N", cli.ParseUint32)

	files_to_process := cli.Short('f').
		Long("file").
		Help("File to process").
		Argument("FILE", cli.ParseString).
		Many()
	// packing things in a struct assumes parser for each field is in scope.
	opt := cli.Map6(
		debug,
		verbose,
		speed,
		output,
		nb_cars,
		files_to_process,
		func(debug bool, verbose uint, speed float64, output string, nb_cars uint32, files_to_process []string) Out {
			return Out{debug, verbose, speed, output, nb_cars, files_to_process}
		}).ToOptions().Run(os.Args)
	fmt.Println(opt)
}
