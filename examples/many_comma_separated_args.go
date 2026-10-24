// To parse comma separated values it's easier to treat them as strings
package examples

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/rprtr258/cli"
	"github.com/rprtr258/fun"
)

func ManyCommaSeparatedArgs() {
	// --ports 1,2,3 --ports 4,5 => [1,2,3,4,5]
	var args cli.Parser[[]uint16] = cli.Long("ports").
		Help("Comma separated list of ports").Argument("PORTS", func(s string) ([]uint16, error) {
		parts := strings.Split(s, ",")
		res := make([]uint16, len(parts))
		for i, part := range parts {
			u, err := strconv.ParseUint(part, 10, 16)
			if err != nil {
				return nil, err
			}
			res[i] = uint16(u)
		}
		return res, nil
	}).Parser().Many().Map(func(nested [][]uint16) []uint16 {
		return fun.ConcatMap(func(x []uint16) []uint16 { return x }, nested...)
	})

	fmt.Println(args.ToOptions().Run(os.Args))
}
