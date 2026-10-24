// A somewhat comprehensive example of a typical usage of this library in combinatoric mode
package examples

import (
	"fmt"
	"os"
	"strconv"

	"github.com/rprtr258/cli"
)

func Basic() {
	type Out struct {
		debug            bool
		verbose          uint
		speed            float64
		output           string
		nb_cars          uint32
		files_to_process []string
	}

	newOpts := func() *cli.OptionParser[Out] {
		// A flag, true if used in the command line. Can be required, this one is optional
		debug := cli.Short('d'). // start with a short name
						Long("debug").               // also add a long name
						Help("Activate debug mode"). // and a help message to use
						Switch()                     // turn this into a switch

		// number of occurrences of the v/verbose flag capped at 3 with an error here but you can also
		// use `max` inside `map`
		verbose := cli.Short('v').
			Long("verbose").
			Help("Increase the verbosity\n You can specify it up to 3 times\n either as -v -v -v or as -vvv").
			ReqFlag(struct{}{}).
			Many().
			Map(func(xs []struct{}) uint { return uint(len(xs)) }).
			Guard(func(x uint) bool { return x <= 3 }, "It doesn't get any more verbose than this")

		// an argument, parsed and with default value
		speed := cli.Short('s').
			Long("speed").
			Help("Set speed").Argument("SPEED", func(s string) (float64, error) {
			return strconv.ParseFloat(s, 64)
		}). // you can specify a type to parse
			FallbackDisplay(42.0)

		output := cli.Short('o').
			Long("output").
			Help("output file").Argument("OUTPUT", cli.ParseString). // no fallback means the value is required
			CompleteShell(cli.ShellCompFile(cli.Invalid[string]()))

		// no magical name transmogrifications in combinatoric API,
		nb_cars := cli.Short('n').
			Long("nb-cars").
			Help("Number of items to process").Argument("N", func(s string) (uint32, error) {
			u, err := strconv.ParseUint(s, 10, 32)
			return uint32(u), err
		}).
			FallbackDisplay(1)

		// a parser that consumes one argument
		// you can build the inner parser in one go or as multiple steps giving each step a name
		// you can also add some static shell completion functionality
		file_to_proces := cli.Short('f').
			Long("file").
			Help("File to process").Argument("FILE", cli.ParseString).
			CompleteShell(cli.ShellCompFile(cli.Valid("*.go")))

		files_to_process := file_to_proces.Many()

		// packing things in a struct assumes parser for each field is in scope.
		// TODO: codgen on Out struct
		return cli.Map6(
			debug,
			verbose,
			speed,
			output,
			nb_cars,
			files_to_process,
			func(debug bool, verbose uint, speed float64, output string, nb_cars uint32, files_to_process []string) Out {
				return Out{
					debug,
					verbose,
					speed,
					output,
					nb_cars,
					files_to_process,
				}
			}).
			ToOptions().
			Description("This is a description")
	}
	opts := newOpts().FallbackToUsage().Run(os.Args)
	fmt.Println(opts)
}
