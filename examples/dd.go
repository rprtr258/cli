// `dd.rs` - this is not a typical usage, but you should be able to replicate the
// command line interface used by dd.
package examples

import (
	"fmt"
	"os"
	"strings"

	"github.com/rprtr258/cli"
	"github.com/rprtr258/fun"
)

type ddOptions struct {
	magic     bool
	inFile    string
	outFile   string
	blockSize uint
}

// Parses a string that starts with `name`, returns the suffix parsed in a usual way.
func ddTag[T any](name, meta, help string, parse func(string) (T, error)) cli.Parser[T] {
	// it is possible to parse raw command line items here and strip the prefix manually
	return cli.Parse(
		cli.Any("", func(s string) (string, bool) {
			return strings.CutPrefix(s, name)
		}, true).
			// this defines a custom metavar for the help message
			Metavar(cli.NewDocFromPairs([]fun.Pair[string, cli.Style]{
				{K: name, V: cli.StyleLiteral},
				{K: meta, V: cli.StyleMetavar},
			})).
			Help(help).
			Parser(),
		parse)
}

func ddInFile() cli.Parser[string] {
	return ddTag("if=", "FILE", "read from FILE", cli.ParseString).
		Fallback("-").
		FallbackDisplay("-")
}

func ddOutFile() cli.Parser[string] {
	return ddTag("of=", "FILE", "write to FILE", cli.ParseString).
		Fallback("-").
		FallbackDisplay("-")
}

func ddBlockSize() cli.Parser[uint] {
	// it is possible to parse the notation used by dd itself as well,
	// using uint only for simplicity
	return ddTag("bs=", "SIZE", "read/write SIZE blocks at once", cli.ParseUint).
		Fallback(512).
		FallbackDisplay(512)
}

func Dd() {
	magic := cli.Short('m').Long("magic").
		Help("a usual switch still works").Switch()

	opts := cli.Map4(
		magic, ddInFile(), ddOutFile(), ddBlockSize(),
		func(magic bool, inFile, outFile string, blockSize uint) ddOptions {
			return ddOptions{magic: magic, inFile: inFile, outFile: outFile, blockSize: blockSize}
		},
	).ToOptions()

	fmt.Println(opts.Run(os.Args))
}
