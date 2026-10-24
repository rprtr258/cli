// `derive-smart-pointer.rs` - with minimal changes you can parse directly into
// smart pointers.
//
// While the parser knows nothing about Box or Arc it lets you parse into them by
// overriding the parsing sequence: here `name` is first parsed into a regular
// `string` then converted into a pointer using Map. Similarly a set of coins is
// parsed into a regular slice and later converted into a pointer still inside the
// parser.
package examples

import (
	"fmt"
	"os"

	"github.com/rprtr258/cli"
)

type smartPointerOptions struct {
	// Adventurer's name
	name *string
	// A set of coins
	coins *[]uint
}

func DeriveSmartPointer() {
	name := cli.Long("name").
		Help("Adventurer's name").
		Argument("NAME", cli.ParseString).
		Parser().
		Map(func(s string) *string { return &s })

	coins := cli.Positional("COIN", cli.ParseUint).
		Help("A set of coins").
		Parser().
		Many().
		Map(func(xs []uint) *[]uint { return &xs })

	opts := cli.Map2(name, coins, func(name *string, coins *[]uint) smartPointerOptions {
		return smartPointerOptions{name: name, coins: coins}
	}).
		Parser().
		ToOptions().
		FallbackToUsage().
		Run(os.Args)

	fmt.Println(opts)
}
