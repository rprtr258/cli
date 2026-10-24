// This example shows dynamic shell completion features
package examples

import (
	"fmt"
	"iter"
	"os"
	"strings"

	"github.com/rprtr258/cli"
)

func Csample() {
	complete_calculator := func(input string) iter.Seq2[string, cli.Opt[string]] {
		items := []string{"alpha", "beta", "banana", "cat", "durian"}
		return func(yield func(string, cli.Opt[string]) bool) {
			for _, item := range items {
				if strings.HasPrefix(item, input) && !yield(item, cli.Invalid[string]()) {
					return
				}
			}
		}
	}

	a := cli.Short('a').Long("avocado").Help("Use avocado").Switch()
	b := cli.Short('b').Long("banana").Help("Use banana").Switch()
	bb := cli.Long("bananananana").Help("I'm Batman").Switch()
	c := cli.Long("calculator").
		Help("calculator expression").Argument("EXPR", func(s string) (string, error) {
		return s, nil
	}).Complete(complete_calculator)
	parser := cli.Pair(cli.Pair(a, b), cli.Pair(bb, c)).
		ToOptions().
		Description("Dynamic autocomplete example").
		Footer(`This library supports dynamic autocompletion for a few shells, make sure your binary is in $PATH and try using one of those; this output should go into a file that depends on your shell:
$ csample --cli-complete-style-bash
$ csample --cli-complete-style-zsh
$ csample --cli-complete-style-fish
$ csample --cli-complete-style-elvish`)

	fmt.Println(parser.FallbackToUsage().Run(os.Args))
}
