// A way to represent xorg like flags, not a typical usage
package examples

import (
	"fmt"
	"os"
	"strings"

	"github.com/rprtr258/cli"
	"github.com/rprtr258/fun"
)

// toggleOptions matches a literal name prefixed with - or +.
func toggleOptions(meta, name, help string) cli.Parser[bool] {
	return cli.
		Any(meta, func(s string) (bool, bool) {
			if suf, ok := strings.CutPrefix(s, "+"); ok {
				return true, suf == name
			} else if suf, ok := strings.CutPrefix(s, "-"); ok {
				return false, suf == name
			} else {
				return false, false
			}
		}, true).
		Help(help).
		Parser()
}

func Xorg() {
	type Options struct {
		turbo, backing, xinerama bool
		extensions               []fun.Pair[string, bool]
	}

	// matches literal +ext and -ext followed by extension name
	extension := cli.Pair(
		cli.Any("(+|-)ext", func(s string) (bool, bool) {
			return s == "+ext", s == "-ext" || s == "+ext"
		}, true),
		cli.Positional("EXT", func(s string) (string, error) { return s, nil }).
			Help("Extension to enable or disable, see documentation for the full list")).
		Adjacent().
		Map(func(ab fun.Pair[bool, string]) fun.Pair[string, bool] {
			return fun.Pair[string, bool]{K: ab.V, V: ab.K}
		}).
		Many()

	options := cli.Map4(
		cli.Short('t').
			Long("turbo").
			Help("Engage the turbo mode").
			Switch(),
		toggleOptions("(+|-)backing", "backing", "Set backing status").Fallback(false),
		toggleOptions("(+|-)xinerama", "xinerama", "Set Xinerama status").Fallback(true),
		extension,
		func(turbo bool, backing bool, xinerama bool, extensions []fun.Pair[string, bool]) Options {
			return Options{turbo, backing, xinerama, extensions}
		},
	).
		ToOptions().
		Run(os.Args)

	fmt.Println(options)
}
