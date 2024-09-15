//! # Batteries included - helpful parsers that use only public API
//!
//! `bpaf` comes with a few extra functions that use only public API in their implementation. You
//! might find them useful either for your code or as an inspiration source
//!
//! **To use anything in this module you need to enable `batteries` cargo feature.**
//!
//! Examples contain combinatoric usage, for derive usage you should create a parser function and
//! use `external` annotation.

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
// ```rust
// # use bpaf*;
// use bpafbatteries*;
//
//	func verbose()impl Parser[uint] {
//	    verbose_and_quiet_by_number(2, 0, 5).map(|v| v as uint)
//	}
//
// ```
func verbose_and_quiet_by_number(offset int, min int, max int) Parser[int] {
	verbose := short('v').
		long("verbose").
		help("Increase output verbosity, can be used several times").
		req_flag(struct{}{}).
		many().
		Map((func(v) v.len)())

	quiet := short('q').
		long("quiet").
		help("Decrease output verbosity, can be used several times").
		req_flag(struct{}{}).
		many().
		Map((func(v) v.len)())

	construct(verbose, quiet).Map(func(v, q) { return (v - q + offset).clamp(min, max) })
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
// # res := parser.run_inner(&[]).unwrap();
// # assert_eq!(LevelInfo, res);
// # res := parser.run_inner(&["-q"]).unwrap();
// # assert_eq!(LevelWarning, res);
// # res := parser.run_inner(&["-qqq"]).unwrap();
// # assert_eq!(LevelError, res);
// # res := parser.run_inner(&["-qqqq"]).unwrap();
// # assert_eq!(LevelError, res);
// # res := parser.run_inner(&["-vvvvq"]).unwrap();
// # assert_eq!(LevelTrace, res);
// ```
func verbose_by_slice(
	offset uint,
	items []T,
) Parser[T] {
	verbose_and_quiet_by_number(offset, 0, items.len()-1).
		Map(func(i) items[i])
}

// Pick last passed value between two different flags
//
// Usually `bpaf` only allows to parse a single instance for every invocation unless
// you specify [`many`](Parsermany) or [`some`](Parsersome). `toggle_flag` would consume
// multiple instances of two different flags and returns last specified value.
//
// This function relies on a fact that selection between two different parsers prefers left most
// value. This helps to preserve relative order of parsrs.
// You can use similar approach to combine multiple flags accounting for their relative order.
//
// Parser returns `Optional<T>` value, you can add a fallback with [`map`](Parsermap) or turn
// missing value info failure with a custom error message with [`parse`](Parserparse).
//
// # Example
// ```console
// $ app --banana --no-banana --banana --banana
// fun.Valid(Banana)
// $ app
// fun.Invalid()
// ```
//
// # Usage
// ```rust
// # use bpaf*;
// use bpafbatteriestoggle_flag;
//
// #[derive(Copy, Clone, Debug, PartialEq, Eq)]
//
//	enum Select {
//	    Banana,
//	    NoBanana,
//	}
//
//	func pick()impl Parser<fun.Option[Select>] {
//	    toggle_flag(long("banana"), SelectBanana, long("no-banana"), SelectNoBanana)
//	}
//
// ```
func toggle_flag[T any](
	a NamedArg,
	val_a T,
	b NamedArg,
	val_b T,
) Parser[fun.Option[T]] {
	a := a.req_flag(val_a)
	b := b.req_flag(val_b)
	construct(a, b).many().Map((func(xs) xs.into_iter)().last())
}

// Strip a command name if present at the front when used as a `cargo` command
//
// When implementing a cargo subcommand parser needs to be able to skip the first argument which
// is always the same as the executable name without `cargo-` prefix. For example if executable name is
// `cargo-cmd` so first argument would be `cmd`. `cargo_helper` helps to support both invocations:
// with name present when used via cargo and without it when used locally.
//
// You can read the code of this function as this approximate sequence of statements:
// 1. Try to parse a string literal that corresponds to a command name
// 2. It's okay if it's missing
// 3. And don't show anything to the user in `--help` or completion
// 4. Parse this word and then everything else as a tuple, return that second item.
func cargo_helper[T any](cmd string, parser Parser[T]) Parser[T] {
	skip := literal(cmd).optional().hide()
	construct(skip, parser).Map(func(x) { return x[1] })
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
//	    println!("{}", get_usage(parser()));
//	    os.Exit(0);
//	}
//	...
//
// ```
func get_usage[T any](parser OptionParser[fmt.Formatter]) string {
	return parser.run_inner("--help").unwrap_err().unwrap_stdout()
}
