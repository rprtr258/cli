// Helpful parsers built from the public API.
//
// These helpers use only the public API in their implementation. You
// might find them useful for your code or as an inspiration source.

package cli

import (
	"fmt"

	"github.com/rprtr258/fun"
)

// verbose_and_quiet_by_number returns `--verbose` and `--quiet` flags with
// results encoded as a number.
//
// Parameters specify the offset and minimal/maximal values. The parser accepts many `-v | --verbose` and
// `-q | --quiet` to increase and decrease verbosity respectively
//
// # Usage
//
//	verbosity := verboseAndQuietByNumber(2, 0, 5)
func verbose_and_quiet_by_number(offset, lo, hi int) Parser[int] {
	vp := ReqFlag(Short('v').
		Long("verbose").
		Help("Increase output verbosity, can be used several times"),
		struct{}{}).Many()
	verbose := vp.Map(lenOf)

	qp := ReqFlag(Short('q').
		Long("quiet").
		Help("Decrease output verbosity, can be used several times"),
		struct{}{}).Many()
	quiet := qp.Map(lenOf)

	return Map2(verbose, quiet, func(v, q int) int { return fun.Clamp(v-q+offset, lo, hi) })
}

// verbose_by_slice returns `--verbose` and `--quiet` flags with results
// chosen from a slice of values.
//
// The parameters specify the possible values and the offset: the number of `-v` minus `-q`
// selects the value by index.
//
//	level := verboseBySlice(2, []Level{LevelError, LevelWarning, LevelInfo, LevelDebug, LevelTrace})
func verbose_by_slice[T any](
	offset int,
	items []T,
) Parser[T] {
	pp := verbose_and_quiet_by_number(offset, 0, len(items)-1).
		Map(func(i int) T { return items[i] })
	return Parser[T]{pp.eval, pp.meta}
}

// toggle_flag picks the last passed value between two different flags.
//
// Usually the parser accepts only a single instance of each item unless
// you use [Parser.Many] or [Parser.Some]. toggle_flag consumes multiple instances of two different
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
	return Choice(pa, pb).Many().Map(last)
}

// get_usage returns the usage for a parser.
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
