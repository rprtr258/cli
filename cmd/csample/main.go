// This example shows dynamic shell completion features
package main

import (
	"fmt"
	"iter"
	"strings"

	"github.com/rprtr258/cli"
)

func complete_calculator(input string) iter.Seq2[string, cli.Option[string]] {
	items := []string{"alpha", "beta", "banana", "cat", "durian"}
	return func(yield func(string, cli.Option[string]) bool) {
		for _, item := range items {
			if strings.HasPrefix(item, input) {
				yield(item, cli.Invalid[string]())
			}
		}
	}
}

func main() {
	a := cli.Short('a').Long("avocado").Help("Use avocado").Switch()
	b := cli.Short('b').Long("banana").Help("Use banana").Switch()
	bb := cli.Long("bananananana").Help("I'm Batman").Switch()
	c := cli.Argument[string](cli.Long("calculator").
		Help("calculator expression"), "EXPR", func(s string) (string, error) {
		return s, nil
	}).Parser().
		Complete(complete_calculator)
	parser := cli.Pair(cli.Pair(a, b).Parser(), cli.Pair(bb, c.Parser()).Parser()).Parser().
		ToOptions().
		Descr("Dynamic autocomplete example").
		Footer(`bpaf supports dynamic autocompletion for a few shells, make sure your binary is in $PATH and try using one of those this output should go into a file that depends on your shell:
$ csample --bpaf-complete-style-bash
$ csample --bpaf-complete-style-zsh
$ csample --bpaf-complete-style-fish
$ csample --bpaf-complete-style-elvish`)

	fmt.Println(parser.FallbackToUsage().Run())
}
