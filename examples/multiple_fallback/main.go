// Multi level fallback example:
//
// Fallback to one of several values
// - the command line argument
// - the environmental variable
// - the config file
// - the hard-coded default
package multiple_fallback

import (
	"fmt"
	"os"
	"strconv"

	"github.com/rprtr258/cli"
	"github.com/rprtr258/fun"
)

type Config struct {
	field1 uint32
	field2 uint64
}

// Here this is a constant but it can be OnceCell from `once_cell`
// that reads the config and deals with substituting missing falues with defaults
var DEFAULT_CONFIG = Config{
	field1: 42,
	field2: 10,
}

func Main() {
	field1 := cli.Argument[uint32](cli.Long("field1").
		Env("FIELD1").
		Help("Field 1"),
		"ARG", func(s string) (uint32, error) {
			u, err := strconv.ParseUint(s, 10, 32)
			return uint32(u), err
		}).
		Parser().
		Fallback(DEFAULT_CONFIG.field1).
		Parser()
	field2 := cli.Argument[uint64](cli.Long("field2").
		Env("FIELD2").
		Help("Field 2"),
		"ARG", func(s string) (uint64, error) {
			return strconv.ParseUint(s, 10, 64)
		}).
		Parser().
		Fallback(DEFAULT_CONFIG.field2).
		Parser()

	opts := cli.Map(cli.Pair(field1, field2).Parser(), func(x fun.Pair[uint32, uint64]) Config { return Config{x.K, x.V} }).ToOptions().Run(os.Args)

	// At this point if you get opts - it should be taken from one of
	// - the command line argument
	// - the environmental variable
	// - the config file
	// - the hard-coded default (from config parser)
	fmt.Println(opts)
}
