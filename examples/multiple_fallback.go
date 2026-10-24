// Multi level fallback example:
//
// Fallback to one of several values
// - the command line argument
// - the environmental variable
// - the config file
// - the hard-coded default
package examples

import (
	"fmt"
	"os"

	"github.com/rprtr258/cli"
)

func MultipleFallback() {
	type Config struct {
		field1 uint32
		field2 uint64
	}

	// Here this is a constant but it can be OnceCell from `once_cell`
	// that reads the config and deals with substituting missing falues with defaults
	DEFAULT_CONFIG := Config{
		field1: 42,
		field2: 10,
	}

	field1 := cli.Long("field1").
		Env("FIELD1").
		Help("Field 1").
		Argument("ARG", cli.ParseUint32).
		Fallback(DEFAULT_CONFIG.field1)
	field2 := cli.Long("field2").
		Env("FIELD2").
		Help("Field 2").
		Argument("ARG", cli.ParseUint64).
		Fallback(DEFAULT_CONFIG.field2)

	opts := cli.Map2(field1, field2, func(field1 uint32, field2 uint64) Config { return Config{field1, field2} }).ToOptions().Run(os.Args)

	// At this point if you get opts - it should be taken from one of
	// - the command line argument
	// - the environmental variable
	// - the config file
	// - the hard-coded default (from config parser)
	fmt.Println(opts)
}
