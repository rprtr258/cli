// `many_comma_separated_args_derive.rs` - to parse comma separated values it's
// easier to treat them as strings.
//
// The Rust file is the derive-syntax twin of `many_comma_separated_args.rs` with
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
	parts := strings.Split(s, ",")
	res := make([]uint16, 0, len(parts))
	for _, part := range parts {
		u, err := strconv.ParseUint(part, 10, 16)
		if err != nil {
			return nil, err
		}
		res = append(res, uint16(u))
	}
	return res, nil
}

type commaSeparatedOpts struct {
	// Comma separated list of ports
	Ports []uint16
}

func ManyCommaSeparatedArgsDerive() {
	// --ports 1,2,3 --ports 4,5 => [1,2,3,4,5]
	ports := cli.Long("ports").
		Help("Comma separated list of ports").
		Argument("PORTS", splitAndParseU16).
		Parser().
		Many().
		Map(func(nested [][]uint16) []uint16 {
			return fun.ConcatMap(func(x []uint16) []uint16 { return x }, nested...)
		})

	opts := ports.Map(func(ports []uint16) commaSeparatedOpts {
		return commaSeparatedOpts{Ports: ports}
	}).Parser().ToOptions().Run(os.Args)

	fmt.Println(opts)
}
