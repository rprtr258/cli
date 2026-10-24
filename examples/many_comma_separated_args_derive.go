// `many_comma_separated_args_derive.go` - to parse comma separated values it's
// easier to treat them as strings.
//
// The Rust file is the derive-syntax twin of `many_comma_separated_args.go` with
// `parse(split_and_parse), many, map(flatten)` attributes. The Go struct-tag API
// has no per-word custom converter for slice fields, so this is the same parser
// expressed with the combinatoric API.
package examples

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/rprtr258/cli"
	"github.com/rprtr258/fun"
)

func splitAndParseU16(s string) ([]uint16, error) {
	return fun.MapErr[uint16](func(part string) (uint16, error) {
		u, err := strconv.ParseUint(part, 10, 16)
		return uint16(u), err
	}, strings.Split(s, ",")...)
}

type commaSeparatedOpts struct {
	// Comma separated list of ports
	Ports []uint16
}

func ManyCommaSeparatedArgsDerive() {
	// --ports 1,2,3 --ports 4,5 => [1,2,3,4,5]
	opts := cli.Long("ports").
		Help("Comma separated list of ports").
		Argument("PORTS", splitAndParseU16).
		Many().
		Map(func(nested [][]uint16) []uint16 { return fun.Flatten(nested...) }).
		Map(func(ports []uint16) commaSeparatedOpts {
			return commaSeparatedOpts{Ports: ports}
		}).
		ToOptions().
		Run(os.Args)

	fmt.Println(opts)
}
