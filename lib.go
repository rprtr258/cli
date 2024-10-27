package cli

import (
	"github.com/rprtr258/fun"
)

// #![warn(missing_docs)]
// #![allow(clippyneedless_doctest_main)]
// #![allow(clippyredundant_else)] // not useful
// #![allow(rustdocredundant_explicit_links)] // two random markdown parsers I tried only supports explicit links
// #![cfg_attr(docsrs, feature(doc_auto_cfg))]

//! Lightweight and flexible command line argument parser with derive and combinatoric style API

//! # Quick links
//! - [Introduction](_documentation_0_intro) - features, design goals, restrictions
//! - [Tutorials](_documentation_1_tutorials) - practical learning oriented information and
//!   examples to get you started
//!   + [Types of arguments](_documentation_1_tutorials_0_types_of_arguments) -
//!     common types of line fun.Options and conventions (optional)
//!   + [Combinatoric API](_documentation_1_tutorials_1_combinatoric_api)  -
//!     Parse arguments without using proc macros
//!   + [Derive API](_documentation_1_tutorials_2_derive_api) -
//!     Create a parser by defining a structure
//! - [How-to and guides](_documentation_2_howto) - assumes familiarity with the basics and
//!   explains how to concrete tasks
//! - [Explanations](_documentation_4_explanation) - theoretical information about abstractions
//!   used by the library, oriented for understanding
//! - [FAQ](https://github.com/pacak/bpaf/discussions) - questions from library users

//! # A quick start
//!
//! Add `bpaf`, fun.Optionally with derive enabled
//!
//! ```text
//! $ cargo add bpaf -F derive,dull_color
//! ```
//!
//! Use either derive or combinatoric API and try running it
//!
// #![cfg_attr(not(doctest), doc = include_str!("docs2/intro.md"))]

//!
//! ## Consuming items - making `Parser`
//!
//! `bpaf` allows you to describe the parsers using a mix of two APIs: combinatoric and derive.
//! Both APIs can achieve the same results, you can use one that better suits your needs. You can
//! find documentation with more examples following those links.
//!
//! - For an argument with a name you define [NamedArg] using a combination of [short],
//!   [long] and [env](crateenv()). At the same time you can attach
//!   [help](NamedArghelp).
//! - [NamedArgswitch] - simple switch that returns `true` if it's present on a command
//!   line and `false` otherwise.
//! - [NamedArgflag] - a variant of `switch` that lets you return one of two custom
//!   values, for example `ColorOn` and `ColorOff`.
//! - [NamedArgreq_flag] - a variant of `switch` that only only succeeds when it's name
//!   is present on a command line
//! - [NamedArgargument] - named argument containing a value, you can further
//!   customize it with [adjacent](crateparsersParseArgumentadjacent)
//! - [positional] - positional argument, you can further customize it with
//!   [strict](ParsePositionalstrict)
//! - [OptionParsercommand] - subcommand parser.
//! - [any] and its specialized version [literal] are escape hatches that can parse anything
//!   not fitting into usual classification.
//! - [pure] and [pure_with] - a way to generate a value that can be composed without parsing
//!   it from the command line.
//!
//! ## Transforming and changing parsers
//!
//! By default primitive parsers gives you back a single `bool`, a single `PathBuf` or a single
//! value produced by [FromStr] trait, etc. You can further transform it by chaining methods from
//! [Parser] trait, some of those methods are applied automagically if you are using derive API.
//!
//! `bpaf` distinguishes two types of parse failures - "value is absent" and
//! "value is present but invalid", most parsers listed in this section only handle the first
//! type of failure by default, but you can use their respective `catch` method to handle the later
//! one.
//!
//! - [fallback](Parserfallback) and [fallback_with](Parserfallback_with) - return a
//!   different value if parser fails to find what it is looking for. Generated help for former
//!   can be updated to include default value using
//!   [display_fallback](ParseFallbackdisplay_fallback) and
//!   [debug_fallback](ParseFallbackdebug_fallback) .
//! - [optional](Parseroptional) - return `Invalid()` if value is missing instead of failing, see
//!   also [catch](ParseOptionalcatch) .
//! - [many](Parsermany), [some](Parsersome) and [collect](Parsercollect) - collect
//!   multiple values into a collection, usually a vector, see their respective
//!   [catch](ParseManycatch), [catch](ParseSomecatch) and [catch](ParseCollectcatch).
//! - [map](Parsermap), [parse](Parserparse) and [guard](Parserguard) - transform
//!   and/or validate value produced by a parser
//! - [to_options](Parserto_options) - finalize the parser and prepare to run it
//!
//! ## Combining multiple parsers together
//!
//! Once you have parsers for all the primitive fields figured out you can start combining them
//! together to produce a parser for a final result - data type you designed in the step one.
//! For derive API you apply annotations to data types with `#[derive(Bpaf)`] and `#[bpaf(..)]`,
//! with combinatoric API you use [construct](crateconstruct) macro.
//!
//! All fields in a struct needs to be successfully parsed in order for the parser to succeed
//! and only one variant from enum will consume its values at a time.
//!
//! You can use [adjacent](ParseConadjacent) annotation to parse multiple flags as an adjacent
//! group allowing for more unusual scenarios such as multiple value arguments or chained commands.
//!
//! ## Improving user experience
//!
//! `bpaf` would use doc comments on fields and structures in derive mode and and values passed
//! in various `help` methods to generate `--help` documentation, you can further improve it
//! using those methods:
//!
//! - [hide_usage](Parserhide_usage) and [hide](Parserhide) - hide the parser from
//!   generated *Usage* line or whole generated help
//! - [group_help](Parsergroup_help) and [with_group_help](Parserwith_group_help) -
//!   add a common description shared by several parsers
//! - [custom_usage](Parsercustom_usage) - customize usage for a primitive or composite parser
//! - [usage](OptionParserusage) and [with_usage](OptionParserwith_usage) lets you to
//!   customize whole usage line as a whole either by completely overriding it or by building around it.
//!
//! By default with completion enabled `bpaf` would complete names for flags, arguments and
//! commands. You can also generate completion for argument values, possible positionals, etc.
//! This requires enabling **autocomplete** cargo feature.
//!
//! - [complete](Parsercomplete) and [complete_shell](Parsercomplete_shell)
//!
//! And finally you can generate documentation for command line in markdown, html and manpage
//! formats using [render_markdown](OptionParserrender_markdown),
//! [render_html](OptionParserrender_html) and [render_manpage](OptionParserrender_manpage),
//! for more detailed info see [doc] module
//!
//! ## Testing your parsers and running them
//! - You can [OptionParserrun] the parser on the arguments passed on the command line
//! - [check_invariants](OptionParsercheck_invariants) checks for a few invariants in the
//!   parser `bpaf` relies on
//! - [run_inner](OptionParserrun_inner) runs the parser with custom [Args] you can create
//!   either explicitly or implicitly using one of the [From] implementations, `Args` can be
//!   customized with [set_comp](Argsset_comp) and [set_name](Argsset_name).
//! - [ParseFailure] contains the parse outcome, you can consume it either by hands or using one
//!   of [exit_code](ParseFailureexit_code), [unwrap_stdout](ParseFailureunwrap_stdout) and
//!   [unwrap_stderr](ParseFailureunwrap_stderr)
//!
//! ## Cargo features
//!
//!  - `derive`: adds a dependency on `bpaf_derive` crate and reexport `Bpaf` derive macro. You
//!     need to enable it to use derive API. Disabled by default.
//!
//!  - `batteries`: helpers implemented with public `bpaf` API. Disabled by default.
//!
//!  - `autocomplete`: enables support for shell autocompletion. Disabled by default.
//!
//!
//! - `bright-color`, `dull-color`: use more colors when printing `--help` and such. Enabling
//!   either color feature adds some extra dependencies and might raise MRSV. If you are planning
//!   to use this feature in a published app - it’s best to expose them as feature flags:
//!
//!    ```toml
//!    [features]
//!    bright-color = ["bpaf/bright-color"]
//!    dull-color = ["bpaf/dull-color"]
//!    ```
//!    Disabled by default.
//!
//!  - `docgen`: generate documentation from help declaration, see [OptionParserrender_markdown] and [doc](cratedoc). Disabled by default.

// Simple or composed argument parser
//
// # Overview
//
// It's best to think of an object implementing [Parser] trait as a container with a value
// inside that is composable with other `Parser` containers using [construct] and the only
// way to extract this value is by transforming it to [OptionParser] with
// [to_options](Parserto_options) and running it with [run](OptionParserrun). At which
// point you either get your value out or `bpaf` would generate a message describing a problem
// (missing argument, validation failure, user requested help, etc) and the program would
// exit.
//
// Values inside can be of any type for as long as they implement `Debug`, `Clone` and
// there are no lifetimes other than static.
//
// When consuming the values you can jump straight to a value that implements
// [FromStr] trait and then transform it into something that your program would use. Alternatively,
// you can consume either `String` or `string` and parse that by hand. It's better to perform
// as much parsing and validation inside the `Parser` as possible so the program itself gets
// strictly typed and correct value while the user gets immediate feedback on what's wrong with the
// arguments they pass.
//
// Order of operations matters, each subsequent parser gets the output of the earlier one. Both
// parsers `a` and `b` would consume multiple numeric values, each less than 10, but `a`
// validates a single value and then consumes multiple of them already validated, while `b` first
// consumes and then performs validation. The former approach is usually more readable.
// ```rust
//
//	use bpaf*;
//
//	func simple() {
//	  a := short('a').argument[uint]("N")
//	    .guard(|&a| a < 10, "`a` must be below 10")
//	    .many();
//
//	  b := short('b').argument[uint]("N")
//	   .many()
//	   .guard(|bs| bs.iter().all(|&b| b < 10), "`b` must be below 10");
//	}
//
// ```
//
// The same logic applies to derive API - the current type depends on the order of annotations:
// ```rust
// # use bpaf*;
// # func less_than_10(a: &uint)bool { true }
// # func all_less_than_10(a: &Vec[uint])bool { true }
// #[derive(Bpaf, Debug, Clone)]
//
//	struct Simple {
//	    #[bpaf(argument("N"), guard(less_than_10, "`a` must be below 10"), many)]
//	    a: Vec[uint],
//	    #[bpaf(argument("N"), many, guard(all_less_than_10, "`b` must be below 10"))]
//	    b: Vec[uint],
//	}
//
// ```
//
// For example suppose your program needs the user to specify dimensions of a rectangle, with sides
// being 1..20 units long and the total area must not exceed 200 units square. A parser that
// consumes it might look like this:
//
// ```rust
// # use bpaf*;
// #[derive(Debug, Copy, Clone)]
//
//	struct Rectangle {
//	    width: u32,
//	    height: u32,
//	}
//
//	func rectangle()impl Parser<Rectangle> {
//	    invalid_size := "Sides of a rectangle must be 1..20 units long";
//	    invalid_area := "Area of a rectangle must not exceed 200 units square";
//	    width := long("width")
//	        .help("Width of the rectangle")
//	        .argument<u32>("PX")
//	        .guard(|&x| 1 <= x && x <= 10, invalid_size);
//	    height := long("height")
//	        .help("Height of the rectangle")
//	        .argument<u32>("PX")
//	        .guard(|&x| 1 <= x && x <= 10, invalid_size);
//	    construct(Rectangle { width, height })
//	        .guard(|&r| r.width * r.height <= 400, invalid_area)
//	}
//
// ```
//
// # Derive specific considerations
//
// Every method defined on this trait belongs to the `postprocessing` section of the field
// annotation. `bpaf` would try to figure out what chain to use for as long as there are no
// fun.Options changing the type: you can use [fallback](Parserfallback_with),
// [fallback_with](Parserfallback_with), [guard](Parserguard), [hide](Parserhide`) and
// [group_help](Parsergroup_help) but not the rest of them.
//
// ```rust
// # use bpaf*;
//
//	struct fun.Options {
//	    // no annotation at all - `bpaf` inserts implicit `argument` and gets the right type
//	    number_1: u32,
//
//	    // fallback isn't changing the type so `bpaf` still handles it
//	    #[bpaf(fallback(42))]
//	    number_2: u32,
//
//	    // `bpaf` inserts implicit `argument`, `optional` and the right type
//	    number_3: Option[u32],
//
//	    // fails to compile: you need to specify `argument`
//	    // #[bpaf(optional)]
//	    // number_4: Option[u32],
//
//	    #[bpaf(argument("N"), fun.Optional)]
//	    number_5: Option[u32],
//
//	    // explicit consumer and a full postprocessing chain
//	    #[bpaf(argument<u32>("N"), fun.Optional)]
//	    number_6: Option[u32],
//	}
//
// ```
type Parser[T any] struct {
	eval func(args *State) (T, Error) // Evaluate inner function
	meta Meta                         // Included information about the parser
}

// Parse
// Apply a failing transformation to a contained value
//
// Transformation preserves the present/absent state of the value: to Parse an fun.Optional value you
// can either first try to `Parse` it and then mark it as [optional](Parseroptional) or first
// deal with the fun.Optionality and then Parse a value wrapped in [Option]. In most cases
// the former approach is more concise.
//
// Similarly, it is possible to Parse multiple items with [many](Parsermany) or
// [some](Parsersome) by either parsing a single item first and then turning it into a [Vec]
// or collecting them into a [Vec] first and then parsing the whole vector. The former approach
// is more concise.
//
// This is a most general of transforming parsers and you can express
// [map](Parsermap) and [guard](Parserguard) in terms of it.
//
// Examples are a bit artificial, to Parse a value from a string you can specify
// the type directly in the `argument`'s turbofish and then apply `map`.
//
// # Derive usage:
// `Parse` takes a single parameter: function name to call. Function type should match
// parameter `F` used by `Parse` in combinatoric API.
//
// #[cfg_attr(not(doctest), doc = include_str!("docs2/Parse.md"))]
func Parse[T, R any](self Parser[T], f func(T) Result[R]) ParseWith[T, R] {
	return ParseWith[T, R]{
		inner:    self,
		parse_fn: f,
	}
}

// {{{ Guard
// Validate or fail with a message
//
// If the value doesn't satisfy the constraint - the parser fails with the specified error message.
//
// # Derive usage
// Derive variant of the `Guard` takes a function name instead of a closure, mostly to keep things
// clean. The second argument can be either a string literal or a constant name for a static [str].
//
// }}}
// #[cfg_attr(not(doctest), doc = include_str!("docs2/Guard.md"))]
func (self Parser[T]) Guard(check func(T) bool, message string) ParseGuard[T] {
	return ParseGuard[T]{
		inner:   self,
		check:   check,
		message: message,
	}
}

// combine
// {{{ Fallback
// Use this value as default if the value isn't present on a command line
//
// # Parser would still fail if the value is present but failure comes from some transformation
//
// #[cfg_attr(not(doctest), doc = include_str!("docs2/dis_fallback.md"))]
//
// # See also
// [fallback_with](Parserfallback_with) would allow to try to Fallback to a value that
// comes from a failing computation such as reading a file. By default Fallback value will
// not be shown in the `--help` output, you can change that by using
// [display_fallback](ParseFallbackdisplay_fallback) and
// [debug_fallback](ParseFallbackdebug_fallback).
// }}}
func (self Parser[T]) Fallback(value T) *ParseFallback[T] {
	return &ParseFallback[T]{
		inner:     self,
		value:     value,
		value_str: "",
	}
}

// {{{ fallback_with
// Use value produced by this function as default if the value isn't present
//
// # Would still fail if the value is present but failure comes from some earlier transformation
//
// #[cfg_attr(not(doctest), doc = include_str!("docs2/dis_fallback_with.md"))]
//
// # See also
// [fallback](Parserfallback) implements similar logic expect that failures aren't expected.
// By default fallback value will not be shown in the `--help` output, you can change that by using
// [display_fallback](ParseFallbackWithdisplay_fallback) and
// [debug_fallback](ParseFallbackWithdebug_fallback).
// }}}
func (self Parser[T]) fallback_with(fallback func() Result[T]) ParseFallbackWith[T] {
	return ParseFallbackWith[T]{
		inner:     self,
		fallback:  fallback,
		value_str: "",
	}
}

// {{{ or_else
// If first parser fails - try the second one
//
// For parser to succeed eiter of the components needs to succeed. If both succeed - `bpaf`
// would use output from one that consumed the left most value. The second flag on the command
// line remains unconsumed by `or_else`.
//
// # Combinatoric usage:
// There's two ways to write this combinator with identical results:
// ```rust
// # use bpaf*;
//
//	func a()impl Parser<u32> {
//	    short('a').argument<u32>("NUM")
//	}
//
//	func b()impl Parser<u32> {
//	    short('b').argument<u32>("NUM")
//	}
//
//	func a_or_b_comb()impl Parser<u32> {
//	    construct([a(), b()])
//	}
//
//	func a_or_b_comb2()impl Parser<u32> {
//	    a().or_else(b())
//	}
//
// ```
//
// # Example
// ```console
// $ app -a 12 -b 3
// // 12
// $ app -b 3 -a 12
// // 3
// $ app -b 13
// // 13
// $ app
// // fails asking for either -a NUM or -b NUM
// ```
//
// # Derive usage:
//
// `bpaf` translates enum into alternative combinations, different shapes of variants
// produce different results.
//
// ```bpaf
// # use bpaf*;
//
//	enum Flag {
//	    A { a: u32 }
//	    B { b: u32 }
//	}
//
// ```
//
// ```console
// $ app -a 12 -b 3
// // FlagA { a: 12 }
// $ app -b 3 -a 12
// // FlagB { b: 3 }
// $ app -b 3
// // FlagB { b: 3 }
// $ app
// // fails asking for either -a NUM or -b NUM
// ```
//
// # Performance
//
// `bpaf` tries to evaluate both branches regardless of the successes to produce a
// better error message for combinations of mutually exclusive parsers:
// Suppose program accepts one of two mutually exclusive switches `-a` and `-b`
// and both are present error message should point at the second flag
// Deprecated: instead of a.or_else(b) you should use construct([a, b])
// }}}
func (self Parser[T]) or_else(alt Parser[T]) ParseOrElse[T] {
	return ParseOrElse[T]{
		this: self,
		that: alt,
	}
}

// misc
// {{{ Hide
// Ignore this parser during any sort of help generation
//
// Best used for fun.Optional parsers or parsers with a defined fallback, usually for implementing
// backward compatibility or hidden aliases
//
// #[cfg_attr(not(doctest), doc = include_str!("docs2/Hide.md"))]
// }}}
func (self Parser[T]) Hide() Parser[T] {
	p := ParseHide[T]{inner: self}
	return Parser[T]{p.eval, MetaSkip}
}

// Ignore this parser when generating a usage line
//
// Parsers hidden from usage will still show up in the available arguments list. Best used on
// fun.Optional things that augment the main application functionality but not define it.
// Alternatively, you can use [custom_usage](Parsercustom_usage) to replace a single
// fun.Option or a group of them with some other text.
// #[cfg_attr(not(doctest), doc = include_str!("docs2/HideUsage.md"))]
func (self Parser[T]) HideUsage() ParseUsage[T] {
	return ParseUsage[T]{
		inner: self,
		usage: Doc{},
	}
}

// Customize how this parser looks like in the usage line
//
// #[cfg_attr(not(doctest), doc = include_str!("docs2/CustomUsage.md"))]
func (self Parser[T]) CustomUsage(usage Doc) ParseUsage[T] {
	return ParseUsage[T]{
		inner: self,
		usage: usage,
	}
}

// {{{ GroupHelp
// Attach a help message to a complex parser
//
// `bpaf` inserts the group help message before the block with all the fields
// from the inner parser and an empty line after the block.
//
// #[cfg_attr(not(doctest), doc = include_str!("docs2/GroupHelp.md"))]
// }}}
func (self Parser[T]) GroupHelp(message string) ParseGroupHelp[T] {
	return ParseGroupHelp[T]{
		inner:   self,
		message: newDocFromString(message),
	}
}

// Make a help message for a complex parser from its [MetaInfo]
//
// #[cfg_attr(not(doctest), doc = include_str!("docs2/with_group_help.md"))]
func (self Parser[T]) with_group_help(f func(MetaInfo) Doc) ParseWithGroupHelp[T] {
	return ParseWithGroupHelp[T]{inner: self, f: f}
}

// consume
// {{{ ToOptions
// Transform `Parser` into [OptionParser] to get ready to [run](OptionParserrun) it
//
// # Derive usage
// Add a top-level `options` annotation to generate [OptionParser] instead of default
// [Parser].
//
// In addition to `options` annotation, you can also specify either `version` or
// `version(value)` annotation. The former uses version from `cargo`, later uses the
// specified value which should be an expression of type `string`, see
// [version](OptionParserversion).
//
// #[cfg_attr(not(doctest), doc = include_str!("docs2/ToOptions.md"))]
//
// # See also
// There's some methods implemented on [OptionParser] directly to customize the appearance
// }}}
func (self Parser[T]) ToOptions() *OptionParser[T] {
	return &OptionParser[T]{
		info:  NewInfo(),
		inner: self,
	}
}

// Finalize and run the parser
//
// Generally, you'd want to use [Parserto_options] to finalize the parser and [OptionParserrun],
// but this also works for simple cases:
//
// ```rust
// # use bpaf*;
//
//	func main() {
//	    name := short('n').long("name").argument<String>("USER").run();
//	    // do things with name
//	}
//
// ```
func (self Parser[T]) Run() T {
	return self.ToOptions().Run()
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
	skip := Optional(Literal(cmd)).Parser().Hide()
	return Map(Pair(skip, parser).Parser(), func(x fun.Pair[Option[struct{}], T]) T { return x.V })
}
