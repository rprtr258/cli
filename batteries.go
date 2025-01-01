// Helpful parsers that use only public API.
//
// `bpaf` comes with a few extra functions that use only public API in their implementation. You
// might find them useful either for your code or as an inspiration source
//
// **To use anything in this module you need to enable `batteries` cargo feature.**
//
// Examples contain combinatoric usage, for derive usage you should create a parser function and
// use `external` annotation.

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
//	use bpaf*;
//	use bpafbatteries*;
//
//	func verbose()impl Parser[uint] {
//	    verbose_and_quiet_by_number(2, 0, 5).map(|v| v as uint)
//	}
func verbose_and_quiet_by_number(offset, min, max int) Parser[int] {
	vp := Many(ReqFlag(Short('v').
		Long("verbose").
		Help("Increase output verbosity, can be used several times"),
		struct{}{}), false).Parser()
	verbose := Map(Parser[[]struct{}]{vp.eval, vp.meta},
		func(v []struct{}) int { return len(v) })

	qp := Many(ReqFlag(Short('q').
		Long("quiet").
		Help("Decrease output verbosity, can be used several times"),
		struct{}{}), false).Parser()
	quiet := Map(Parser[[]struct{}]{qp.eval, qp.meta},
		func(v []struct{}) int { return len(v) })

	return Map(Pair(verbose, quiet).Parser(), func(vq fun.Pair[int, int]) int { v, q := vq.K, vq.V; return fun.Clamp(v-q+offset, min, max) })
}

// `--verbose` and `--quiet` flags with results choosen from a slice of values
//
// # Parameters specify an array of possible values and a default index
//
// # Usage
// ```rust
// # use bpaf*;
// use bpafbatteries*;
//
// #[derive(Debug, Copy, Clone, Eq, PartialEq, Ord, PartialOrd)]
//
//	enum Level {
//	   Error,
//	   Warning,
//	   Info,
//	   Debug,
//	   Trace,
//	}
//
//	func verbose()impl Parser<Level> {
//	    use Level*;
//	    verbose_by_slice(2, [Error, Warning, Info, Debug, Trace])
//	}
//
// # parser := verbose().to_options();
// # res := parser.run_inner(newArgs()).unwrap();
// # require.Equal(t,LevelInfo, res);
// # res := parser.run_inner(newArgs("-q")).unwrap();
// # require.Equal(t,LevelWarning, res);
// # res := parser.run_inner(newArgs("-qqq")).unwrap();
// # require.Equal(t,LevelError, res);
// # res := parser.run_inner(newArgs("-qqqq")).unwrap();
// # require.Equal(t,LevelError, res);
// # res := parser.run_inner(newArgs("-vvvvq")).unwrap();
// # require.Equal(t,LevelTrace, res);
// ```
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
// Usually `bpaf` only allows to parse a single instance for every invocation unless
// you specify [many](Parsermany) or [some](Parsersome). `toggle_flag` would consume
// multiple instances of two different flags and returns last specified value.
//
// This function relies on a fact that selection between two different parsers prefers left most
// value. This helps to preserve relative order of parsrs.
// You can use similar approach to combine multiple flags accounting for their relative order.
//
// Parser returns `Optional[T]` value, you can add a fallback with [map](Parsermap) or turn
// missing value info failure with a custom error message with [parse](Parserparse).
//
// # Example
// ```console
// $ app --banana --no-banana --banana --banana
// Valid(Banana)
// $ app
// Invalid()
// ```
//
// # Usage
//
//	use bpaf*;
//	use bpafbatteriestoggle_flag;
//
//	type Select bool
//	const (
//	    SelectNoBanana Select = false
//	    SelectBanana Select = true
//	)
//
//	func pick() Parser[Option[Select]] {
//	    return toggle_flag(
//			Long("banana"), SelectBanana,
//			Long("no-banana"), SelectNoBanana,
//		)
//	}
func toggle_flag[T interface{ ~bool }](
	a NamedArg, val_a T,
	b NamedArg, val_b T,
) Parser[Option[T]] {
	pa := ReqFlag(a, val_a)
	pb := ReqFlag(b, val_b)
	return Map(
		Many(Choice(pa, pb), false).Parser(),
		last)
}

// Get usage for a parser
//
// In some cases you might want to print usage if user gave no command line fun.Options, in this case
// you should add an enum variant to a top level enum, make it hidden with `#[bpaf(hide)]`, make
// it default for the top level parser with something like `#[bpaf(fallback(ArgHelp))]`.
//
// When handling cases you can do something like this for `Help` variant:
//
// ```ignore
//
//	...
//	ArgHelp{
//	    fmt.Println(get_usage(parser()));
//	    os.Exit(0);
//	}
//	...
//
// ```
func get_usage[T any](parser OptionParser[fmt.Formatter]) string {
	x, e := parser.run_inner(newArgs("--help"))
	if e == nil {
		panic(x)
	}
	return unwrap_stdout(e)
}
