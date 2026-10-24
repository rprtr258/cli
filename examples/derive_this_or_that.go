// `derive_this_or_that.go` - deriving for a tri-state enabled/disabled/undecided
// switch; the combinatoric version uses a combination of `req_flag` and `Choice`.
package examples

import (
	"fmt"
	"os"

	"github.com/rprtr258/cli"
)

type thisOrThat uint8

const (
	// undecided
	thisOrThatUndecided thisOrThat = iota
	// enabled
	thisOrThatOn
	// disabled
	thisOrThatOff
)

func DeriveThisOrThat() {
	on := cli.Long("on").Help("enabled").ReqFlag(thisOrThatOn)
	off := cli.Long("off").Help("disabled").ReqFlag(thisOrThatOff)

	opts := cli.Choice(on, off).
		Fallback(thisOrThatUndecided).
		ToOptions().
		Run(os.Args)

	fmt.Println(opts)
}
