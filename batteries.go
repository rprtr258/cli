// Helpful parsers built from the public API.
//
// These helpers use only the public API in their implementation. You
// might find them useful for your code or as an inspiration source.

package cli

import (
	"fmt"

	"github.com/rprtr258/fun"
)

// `--verbose` and `--quiet` flags with results encoded as number
//
// Parameters specify the offset and minimal/maximal values. Parser accepts many `-v | --verbose` and
// `-q | --quiet` to increase and decrease verbosity respectively
//
// # Usage
//
//	verbosity := verboseAndQuietByNumber(2, 0, 5)
func verbose_and_quiet_by_number(offset, min, max int) Parser[int] {
	vp := Many(ReqFlag(Short('v').
		Long("verbose").
		Help("Increase output verbosity, can be used several times"),
		struct{}{})).Parser()
	verbose := Map(Parser[[]struct{}]{vp.eval, vp.meta}, lenOf)

	qp := Many(ReqFlag(Short('q').
		Long("quiet").
		Help("Decrease output verbosity, can be used several times"),
		struct{}{})).Parser()
	quiet := Map(Parser[[]struct{}]{qp.eval, qp.meta}, lenOf)

	return Map(Pair(verbose, quiet).Parser(), func(vq fun.Pair[int, int]) int { v, q := vq.K, vq.V; return fun.Clamp(v-q+offset, min, max) })
}

// `--verbose` and `--quiet` flags with results chosen from a slice of values
//
// The parameters specify the possible values and the offset: the number of `-v` minus `-q`
// selects the value by index.
//
//	level := verboseBySlice(2, []Level{LevelError, LevelWarning, LevelInfo, LevelDebug, LevelTrace})
func verbose_by_slice[T any](
	offset int,
	items []T,
) Parser[T] {
	pp := Map(verbose_and_quiet_by_number(offset, 0, len(items)-1),
		func(i int) T { return items[i] })
	return Parser[T]{pp.eval, pp.meta}
}

// Pick last passed value between two different flags
//
// Usually the parser accepts only a single instance of each item unless
// you use [Many] or [Some]. toggle_flag consumes multiple instances of two different
// flags and returns the last specified value.
//
// This function relies on the fact that selection between two parsers prefers the left most
// value, which preserves the relative order of the parsers.
// You can use a similar approach to combine multiple flags accounting for their relative order.
//
// The parser returns an Opt[T] value: add a fallback with [Parser.Fallback] or turn
// the missing value into a failure with a custom error message with [Parse].
//
// # Example
//
//	$ app --banana --no-banana --banana --banana
//	Valid(Banana)
//	$ app
//	Invalid()
//
// # Usage
//
//	type Select bool
//	const (
//	    SelectNoBanana Select = false
//	    SelectBanana   Select = true
//	)
//
//	func pick() cli.Parser[cli.Opt[Select]] {
//	    return cli.ToggleFlag(
//	        cli.Long("banana"), SelectBanana,
//	        cli.Long("no-banana"), SelectNoBanana,
//	    )
//	}
func toggle_flag[T interface{ ~bool }](
	a NamedArg, val_a T,
	b NamedArg, val_b T,
) Parser[Opt[T]] {
	pa := ReqFlag(a, val_a)
	pb := ReqFlag(b, val_b)
	return Map(
		Many(Choice(pa, pb)).Parser(),
		last)
}

// Get usage for a parser
//
// In some cases you might want to print usage if the user gave no command line options.
// To do that, add a hidden fallback parser for the help variant and print the result
// of get_usage:
//
//	case argHelp:
//	    fmt.Println(GetUsage(parser()))
//	    os.Exit(0)
func get_usage[T any](parser OptionParser[fmt.Formatter]) string {
	x, e := parser.run(newArgList("--help"))
	if e == nil {
		panic(x)
	}
	return unwrap_stdout(e)
}
