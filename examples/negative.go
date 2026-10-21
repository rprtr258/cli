// Parsing negative numbers from the command line. The library won't try to do anything special here,
// user will have to escape minus sign either with `=` or --
package examples

import (
	"fmt"
	"os"
	"strconv"

	"github.com/rprtr258/cli"
)

func Negative() {
	age := cli.Argument[int64](cli.Long("age"), "AGE", func(s string) (int64, error) {
		return strconv.ParseInt(s, 10, 64)
	})
	msg := `To pass a value that starts with a dash requres one one of two special syntaxes:

This will pass '-1' to '--age' handler and leave remaining arguments as is
    --age=-1
This will transform everything after '--' into non flags, '--age' will handle '-1'
and positional handlers will be able to handle the rest.
    --age -- -1`
	num := age.ToOptions().Description(msg).Run(os.Args)
	fmt.Printf("age: %d\n", num)
}
