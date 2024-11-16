// A way to represent xorg like flags, not a typical usage
package main

import (
	"fmt"
	"strings"

	"github.com/rprtr258/cli"
	"github.com/rprtr258/fun"
)

type Options struct {
	turbo, backing, xinerama bool
	extensions               []fun.Pair[string, bool]
}

// matches literal name prefixed with - or +
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
		}).
		Help(help).
		Anywhere().
		Parser()
}

// matches literal +ext and -ext followed by extension name
var extension cli.Parser[fun.Pair[string, bool]] = cli.
	Map(cli.Pair(cli.Any("(+|-)ext", func(s string) (bool, bool) {
		switch s {
		case "-ext":
			return false, true
		case "+ext":
			return true, true
		default:
			return false, false
		}
	}).
		Anywhere().Parser(), cli.Positional("EXT", func(s string) (string, error) {
		return s, nil
	}).
		Help("Extension to enable or disable, see documentation for the full list").Parser()).Adjacent().Parser(),
		func(ab fun.Pair[bool, string]) fun.Pair[string, bool] { return fun.Pair[string, bool]{ab.V, ab.K} })

var options cli.Parser[Options] = cli.Map4(
	cli.Short('t').
		Long("turbo").
		Help("Engage the turbo mode").
		Switch(),
	toggleOptions("(+|-)backing", "backing", "Set backing status").Fallback(false).Parser(),
	toggleOptions("(+|-)xinerama", "xinerama", "Set Xinerama status").Fallback(true).Parser(),
	cli.Many(extension).Parser(),
	func(turbo bool, backing bool, xinerama bool, extensions []fun.Pair[string, bool]) Options {
		return Options{turbo, backing, xinerama, extensions}
	},
)

func main() {
	fmt.Println(options.Run())
}
