package cli

import (
	"fmt"

	"github.com/rprtr258/fun"
)

// Parser that substitutes missing value with a function results but not parser
// failure, created with [`fallback_with`](Parserfallback_with).
type ParseFallbackWith[T any] struct {
	inner     Parser[T]
	fallback  func() Result[T]
	value_str string
}

// impl<T, P, F, E> Parser[T] for ParseFallbackWith<T, P, F, E>
func (self ParseFallbackWith[T]) eval(args *State) Result[T] {
	clone := args
	ok, e := self.inner.eval(clone).unpack()
	if e == nil {
		stdmemswap(args, clone)
		return Ok(ok)
	} else {
		args.swap_comps(clone)
		if e.can_catch() {
			ok, e := self.fallback().unpack()
			if e == nil {
				return Ok(ok)
			} else {
				return Err[T](Error(MessagePureFailed(e.to_string())))
			}
		} else {
			return Err[T](Error(e))
		}
	}
}

func (self ParseFallbackWith[T]) meta() Meta {
	m := MetaOptional(self.inner.meta())
	if self.value_str == "" {
		return m
	} else {
		buf := Docfrom(self.value_str)
		return MetaSuffix(m, buf)
	}
}

// Parser with attached message to several fields, created with [`group_help`](Parsergroup_help).
type ParseGroupHelp[T any] struct {
	inner   Parser[T]
	message Doc
}

// impl<T, P> Parser[T] for ParseGroupHelp<P>
func (self ParseGroupHelp[T]) eval(args *State) Result[T] {
	comp_items := []Comp(nil)
	args.swap_comps_with(&comp_items)
	res := self.inner.eval(args)
	args.swap_comps_with(&comp_items)
	args.push_with_group(&self.message.to_completion(), &comp_items)
	return res
}

func (self ParseGroupHelp[T]) meta() Meta {
	meta := self.inner.meta()
	return MetaSubsection(meta, self.message)
}

// Parser with attached message to several fields, created with [`group_help`](Parsergroup_help).
type ParseWithGroupHelp[T any] struct {
	inner Parser[T]
	f     func(MetaInfo) Doc
}

// impl<T, P, F> Parser[T] for ParseWithGroupHelp<P, F>
func (self ParseWithGroupHelp[T]) eval(args *State) Result[T] {
	return self.inner.eval(args)
}

func (self ParseWithGroupHelp[T]) meta() Meta {
	meta := self.inner.meta()
	buf := (self.f)(MetaInfo(meta))

	return MetaSubsection(meta, buf)
}

// Apply inner parser several times and collect results into `Vec`, created with
// [`some`](Parsersome), requires for at least one item to be available to succeed.
// Implements [`catch`](ParseManycatch)
type ParseSome[T any] struct {
	inner   Parser[T]
	message string
	catch   bool
}

// Handle parse failures
//
// Can be useful to decide to skip parsing of some items on a command line
// When parser succeeds - `catch` version would return a value as usual
// if it fails - `catch` would restore all the consumed values and return fun.Invalid().
//
// There's several structures that implement this attribute: [`ParseOptional`], [`ParseMany`]
// and [`ParseSome`], behavior should be identical for all of them.
// #[cfg_attr(not(doctest), doc = include_str!("docs2/some_catch.md"))]
func (self *ParseSome[T]) Catch() *ParseSome[T] {
	self.catch = true
	return self
}

// impl<T, P> Parser[[]T] for ParseSome<P>
func (self ParseSome[T]) eval(args *State) Result[[]T] {
	res := nil
	len := uintMAX

	for {
		val, ok := parse_option(&self.inner, &len, args, self.catch)
		if !ok {
			break
		}
		res.push(val)
	}

	if res.is_empty() {
		return Err(Error(MessageParseValid(self.message)))
	} else {
		return Ok(res)
	}
}

func (self ParseSome[P]) meta() Meta {
	return MetaMany(MetaRequired(self.inner.meta()))
}

// Apply inner parser several times and collect results into `FromIterator`, created with
// [`collect`](Parsercollect),
// Implements [`catch`](ParseCollectcatch)
type ParseCollect[T any] struct {
	inner Parser[T]
	catch bool
}

// Handle parse failures
//
// Can be useful to decide to skip parsing of some items on a command line
// When parser succeeds - `catch` version would return a value as usual
// if it fails - `catch` would restore all the consumed values and return fun.Invalid().
//
// There's several structures that implement this attribute: [`ParseOptional`], [`ParseMany`]
// and [`ParseSome`], behavior should be identical for all of them.
// #[cfg_attr(not(doctest), doc = include_str!("docs2/some_catch.md"))]
func (self *ParseCollect[T]) Catch() *ParseCollect[T] {
	self.catch = true
	return self
}

// impl<T, C, P> Parser<C> for ParseCollect<P, C, T>
func (self ParseCollect[T]) eval(args *State) Result[[]T] {
	len := uintMAX
	stditerfrom_fn((func() parse_option)(&self.inner, &len, args, self.catch).transpose()).collect[Result[[]T]]()
}

func (self ParseCollect[T]) meta() Meta {
	return MetaMany(MetaRequired(self.inner.meta()))
}

// Parser that returns results as usual but not shown in `--help` output, created with
// [`Parserhide`]
type ParseHide[P any] struct {
	inner P
}

// impl<T, P> Parser[T] for ParseHide<P>
//
//	P: Parser[T],
func (self ParseHide[P]) eval(args *State) Result[T] {
	comps := nil

	args.swap_comps_with(&comps)

	res := self.inner.eval(args)

	args.swap_comps_with(&comps)
	// if Err(Error(MessageMissing(_))) := res {
	//     Err(Error(MessageMissing(nil)))
	// } else {
	//     res
	// }
}

func (self ParseHide[P]) meta() Meta {
	return MetaSkip
}

// Parser that hides inner parser from usage line
//
// No other changes to the inner parser
type ParseUsage[P any] struct {
	inner P
	usage Doc
}

// impl<T, P> Parser[T] for ParseUsage<P>
//
//	P: Parser[T],
func (self ParseUsage[P]) eval(args *State) Result[T] {
	return self.inner.eval(args)
}

func (self ParseUsage[P]) meta() Meta {
	return MetaCustomUsage(self.inner.meta(), self.usage)
}

// Parser that tries to either of two parsers and uses one that succeeeds, created with
// [`Parseror_else`].
type ParseOrElse[T any] struct {
	this Parser[T]
	that Parser[T]
}

// impl[T] Parser[T] for ParseOrElse[T] {
func (self ParseOrElse[T]) eval(args *State) Result[T] {
	comp_items := nil

	args.swap_comps_with(&comp_items)

	// create forks for both branches
	// if they both fail - fallback to the original arguments
	// if they both succed - pick the one that consumes left, remember the second one
	// if one succeeds - pick that, forget the remaining one unless we are doing completion
	args_a := args
	args_b := args

	// run both parsers, expand Result[T, Error> into Option[T] + fun.Option<Error]
	// so that code that does a bunch of comparing logic can be shared across
	// all invocations of parsers rather than being inlined into each one.

	// let (res_a, err_a) = switch self.this.eval(&mut args_a) {
	//     Ok(ok)(Valid(ok), fun.Invalid()),
	//     Err(err)(fun.Invalid(), Valid(err)),
	// };

	// let (res_b, err_b) = switch self.that.eval(&mut args_b) {
	//     Ok(ok)(Valid(ok), fun.Invalid()),
	//     Err(err)(fun.Invalid(), Valid(err)),
	// };

	// if this_or_that_picks_first(
	//     err_a,
	//     err_b,
	//     args,
	//     &args_a,
	//     &args_b,
	//     comp_items,
	// )? {
	//     Ok(res_a.unwrap())
	// } else {
	//     Ok(res_b.unwrap())
	// }
}

func (self ParseOrElse[T]) meta() Meta {
	self.this.meta().or(self.that.meta())
}

// Given two possible errors along with to sets of arguments produce a new error or an instruction
// to pick between two answers. Updates arguments state to switch the results
func this_or_that_picks_first(
	err_a Option[Error],
	err_b Option[Error],
	args *State,
	args_a *State,
	args_b *State,
	comp_stash *[]Comp,
) Result[bool] {
	// if higher depth parser succeeds - it takes a priority
	// completion from different depths should never mix either
	// switch Ordcmp(&args_a.depth(), &args_b.depth()) {
	//     stdcmpOrderingLess{
	//         stdmemswap(args, args_b);

	//         if Valid(comp) := args.comp_mut() {
	//             comp.extend_comps(comp_stash);
	//         }
	//         return switch err_b {
	//             Valid(err)Err(err),
	//             fun.Invalid()Ok(false),
	//         };
	//     }
	//     stdcmpOrderingEqual{}
	//     stdcmpOrderingGreater{
	//         stdmemswap(args, args_a);

	//         if Valid(comp) := args.comp_mut() {
	//             comp.extend_comps(comp_stash);
	//         }
	//         return switch err_a {
	//             Valid(err)Err(err),
	//             fun.Invalid()Ok(true),
	//         };
	//     }
	// }

	// otherwise pick based on the left most or successful one
	// res := switch (err_a, err_b) {
	//     (fun.Invalid(), fun.Invalid()){
	//         if args.len() == args_a.len() && args.len() == args_b.len() {
	//             Ok((true, fun.Invalid()))
	//         } else {
	//             Ok(args_a.pick_winner(args_b))
	//         }
	//     }
	//     (Valid(e1), Valid(e2))Err(e1.combine_with(e2)),
	//     // otherwise either a or b are success, true means a is success
	//     (a_ok, _)Ok((a_ok.is_fun.Invalid()(), fun.Invalid())),
	// };

	{
		keep_a := true
		keep_b := true
		if args_a.len() != args_b.len() {
			// If neither parser consumed anything - both can produce valid completions, otherwise
			// look for the first "significant" consume and keep that parser
			//
			// This is needed to preserve completion from a choice between a positional and a flag
			// See https://github.com/pacak/bpaf/issues/303 for more details
			// if let (Valid(_), Valid(_)) = (args_a.comp_mut(), args_b.comp_mut()) {
			//     CHECK: for (ix, arg) in args_a.items.iter().enumerate() {
			//         // During completion process named and unnamed arguments behave
			//         // different - `-` and `--` are positional arguments, but we want to produce
			//         // named items too. An empty string is also a special type of item that
			//         // gets passed when user starts completion without passing any actual data.
			//         //
			//         // All other strings are either valid named items or valid positional items
			//         // those are hopefully follow the right logic for being parsed/not parsed
			//         if ix + 1 == args_a.len(items) {
			//             os := arg.os_str();
			//             if os.is_empty() || os == "-" || os == "--" {
			//                 break CHECK;
			//             }
			//         }
			//         if let (Valid(a), Valid(b)) = (args_a.present(ix), args_b.present(ix)) {
			//             switch (a, b) {
			//                 (false, true){
			//                     keep_b = false;
			//                     break CHECK;
			//                 }
			//                 (true, false){
			//                     keep_a = false;
			//                     break CHECK;
			//                 }
			//                 _{}
			//             }
			//         }
			//     }
			// }
		}

		// if let (Valid(a), Valid(b)) = (args_a.comp_mut(), args_b.comp_mut()) {
		//     if keep_a {
		//         comp_stash.extend(a.drain_comps());
		//     }
		//     if keep_b {
		//         comp_stash.extend(b.drain_comps());
		//     }
		// }
	}

	// switch res {
	//     Ok((true, ix)){
	//         if Valid(win) := ix {
	//             args_a.save_conflicts(args_b, win);
	//         }
	//         stdmemswap(args, args_a);
	//     }
	//     Ok((false, ix)){
	//         if Valid(win) := ix {
	//             args_b.save_conflicts(args_a, win);
	//         }
	//         stdmemswap(args, args_b);
	//     }
	//     // no winner, keep the completions but don't touch args otherwise
	//     Err(_){}
	// }

	if comp, ok := args.comp_mut().Unpack(); ok {
		comp.extend_comps(comp_stash)
	}

	// Ok(res?.0)
}

// Parser that transforms parsed value with a failing function, created with
// [`parse`](Parserparse)
type ParseWith[T, P, F, E, R any] struct {
	inner    P
	parse_fn F
}

// impl<T, P, F, E, R> Parser<R> for ParseWith<T, P, F, E, R>
//
//	P: Parser[T],
//	F: func(T)Result[R, E],
//	E: ToString,
func (self ParseWith[T, P, F, E, R]) eval(args *State) Result[R, Error] {
	// t := self.inner.eval(args)?;
	// switch (self.parse_fn)(t) {
	//     Ok(r)Ok(r),
	//     Err(e)Err(Error(MessageParseFailed(args.current, e.to_string()))),
	// }
}

func (self ParseWith[T, P, F, E, R]) meta() Meta {
	self.inner.meta()
}

// Parser that substitutes missing value but not parse failure, created with
// [`fallback`](Parserfallback).
type ParseFallback[P, T any] struct {
	inner     P
	value     T
	value_str string
}

// impl<P, T> Parser[T] for ParseFallback<P, T>
//
//	P: Parser[T],
//	T: Clone,
func (self ParseFallback[P, T]) eval(args *State) Result[T] {
	clone := args
	switch self.inner.eval(&clone) {
	// Ok(ok){
	//     stdmemswap(args, &mut clone);
	//     Ok(ok)
	// }
	// Err(Error(e)){

	//     args.swap_comps(&mut clone);
	//     if e.can_catch() {
	//         Ok(self.value)
	//     } else {
	//         Err(Error(e))
	//     }
	// }
	}
}

func (self ParseFallback[P, T]) meta() Meta {
	m := MetaOptional(self.inner.meta())
	if self.value_str.is_empty() {
		m
	} else {
		buf := Docfrom(self.value_str)
		MetaSuffix(m, buf)
	}
}

// impl<P, T: stdfmtDisplay> ParseFallback<P, T> {
// Show [`fallback`](Parserfallback) value in `--help` using [`Display`](stdfmtDisplay)
// representation
//
// #[cfg_attr(not(doctest), doc = include_str!("docs2/dis_fallback.md"))]
func (self *ParseFallback[P, T]) display_fallback() ParseFallback[P, T] {
	self.value_str = fmt.Sprintf("[default: {}]", self.value)
	return self
}

// Show [`fallback`](Parserfallback) value in `--help` using [`Debug`](stdfmtDebug)
// representation
//
// #[cfg_attr(not(doctest), doc = include_str!("docs2/deb_fallback_with.md"))]
func (self *ParseFallback[P, T]) debug_fallback() ParseFallback[P, T] {
	self.value_str = fmt.Sprintf("[default: %#v]", self.value)
	return self
}

// impl<P, T: stdfmtDisplay, F, E>
// F: func()Result[T, E],
// Show [`fallback_with`](Parserfallback_with) value in `--help` using [`Display`](stdfmtDisplay)
// representation
//
// # If fallback function fails - no value will show up
//
// #[cfg_attr(not(doctest), doc = include_str!("docs2/dis_fallback_with.md"))]
func (self *ParseFallbackWith[T, P, F, E]) display_fallback() ParseFallbackWith[T, P, F, E] {
	// if Ok(val) := (self.fallback)() {
	//     self.value_str = fmt.Sprintf("[default: {}]", val);
	// }
	return self
}

// impl<P, T: stdfmtDebug, F, E>
//
//	F: func()Result[T, E],
//
// Show [`fallback_with`](Parserfallback_with) value in `--help` using [`Debug`](stdfmtDebug)
// representation
//
// # If fallback function fails - no value will show up
//
// #[cfg_attr(not(doctest), doc = include_str!("docs2/deb_fallback.md"))]
func (self *ParseFallbackWith[T, P, F, E]) debug_fallback() ParseFallbackWith[T, P, F, E] {
	// if Ok(val) := (self.fallback)() {
	//     self.value_str = fmt.Sprintf("[default: %#v]", val);
	// }
	self
}

// Parser fails with a message if check returns false, created with [`guard`](Parserguard).
type ParseGuard[P, F any] struct {
	inner   P
	check   F
	message string
}

// impl<T, P, F> Parser[T] for ParseGuard<P, F>
//
//	P: Parser[T],
//	F: func(&T)bool,
func (self Parse[P, F]) eval(args *State) Result[T] {
	// t := self.inner.eval(args)?;
	if (self.check)(&t) {
		Ok(t)
	} else {
		Err(Error(MessageGuardFailed(args.current, self.message)))
	}
}

func (self Parse[P, F]) meta() Meta {
	self.inner.meta()
}

// Apply inner parser as many times as it succeeds while consuming something and return this
// number
type ParseCount[P, T any] struct {
	inner P
}

// impl<T, P> Parser[uint] for ParseCount<P, T>
//
//	P: Parser[T],
func (self ParseCount[P, T]) eval(args *State) Result[uint] {
	res := 0
	current := args.len()
	len := uintMAX
	// while (parse_option(&self.inner, &len, args, false)?).is_Valid() {
	//     res += 1;
	//     if current == args.len() {
	//         break;
	//     }
	//     current = args.len();
	// }
	Ok(res)
}

func (self ParseCount[P, T]) meta() Meta {
	MetaMany(MetaOptional(self.inner.meta()))
}

// Apply inner parser as many times as it succeeds while consuming something and return this number
type ParseLast[T any] struct {
	inner Parser[T]
}

// impl<T, P> Parser[T] for ParseLast<P>
func (self ParseLast[T]) eval(args *State) Result[T] {
	last := fun.Invalid()
	current := args.len()
	len := uintMAX
	for {
		val, ok := parse_option(&self.inner, &len, args, false)
		if !ok {
			break
		}
		last = Valid(val)
		if current == args.len() {
			break
		}
		current = args.len()
	}
	if last, ok := last.Unpack(); ok {
		return Ok(last)
	} else {
		return self.inner.eval(args)
	}
}

func (self ParseLast[T]) meta() Meta {
	return MetaMany(MetaRequired(self.inner.meta()))
}

// Handle parse failures for fun.Optional parsers
//
// Can be useful to decide to skip parsing of some items on a command line.
// When parser succeeds - `catch` version would return a value as usual
// if it fails - `catch` would restore all the consumed values and return fun.Invalid().
//
// There's several structures that implement this attribute: [`ParseOptional`], [`ParseMany`]
// and [`ParseSome`], behavior should be identical for all of them.
//
// Those examples are very artificial and designed to show what difference `catch` makes, to
// actually parse arguments like in examples you should [`parse`](Parserparse) or construct
// enum with alternative branches
// #[cfg_attr(not(doctest), doc = include_str!("docs2/optional_catch.md"))]
func (self *ParseOptional[P]) Catch() *ParseOptional[P] {
	self.catch = true
	return self
}

// Apply inner parser several times and collect results into `Vec`, created with
// [`many`](Parsermany), implements [`catch`](ParseManycatch).
type ParseMany[T any] struct {
	inner Parser[T]
	catch bool
}

// Handle parse failures
//
// Can be useful to decide to skip parsing of some items on a command line
// When parser succeeds - `catch` version would return a value as usual
// if it fails - `catch` would restore all the consumed values and return fun.Invalid().
//
// There's several structures that implement this attribute: [`ParseOptional`], [`ParseMany`]
// and [`ParseSome`], behavior should be identical for all of them.
// #[cfg_attr(not(doctest), doc = include_str!("docs2/many_catch.md"))]
func (self *ParseMany[P]) Catch() *ParseMany[P] {
	self.catch = true
	return self
}

// try to parse
func parse_Option[T any](
	parser Parser[T],
	len *uint,
	args *State,
	catch bool,
) Result[Option[T]] {
	orig_args := args
	// switch parser.eval(args) {
	//     // we keep including values for as long as we consume values from the argument
	//     // list or at least one value
	//     Ok(val)Ok(if args.len() < *len {
	//         *len = args.len();
	//         Valid(val)
	//     } else {
	//         fun.Invalid()
	//     }),
	//     Err(Error(err)){
	//         // this is safe to return Ok(fun.Invalid()) in following scenarios
	//         // when inner parser never consumed anything and
	//         // 1. produced ErrorMissing
	//         // 2. produced ErrorMessage(_, true)
	//         // 3. produced ErrorMessage and catch is enabled
	//         //
	//         // In all other scenarios we should return the original error
	//         //
	//         // When parser returns Ok(fun.Invalid()) we should return the original arguments so if there's
	//         // anything left unconsumed - this won't be lost.

	//         missing := matches(err, MessageMissing(_));

	//         if catch || (missing && orig_args.len() == args.len()) || (!missing && err.can_catch())
	//         {
	//             stdmemswap(&mut orig_args, args);

	//             if orig_args.comp_mut().is_Valid() {
	//                 args.swap_comps(&mut orig_args);
	//             }
	//             Ok(fun.Invalid())
	//         } else {
	//             Err(Error(err))
	//         }
	//     }
	// }
}

// impl<T, P> Parser[[]T] for ParseMany<P>
func (self ParseMany[T]) eval(args *State) Result[[]T] {
	len := uintMAX
	return stditerfrom_fn((func() parse_option)(&self.inner, &len, args, self.catch).transpose()).collect[Result[[]T]]()
}

func (self ParseMany[T]) meta() Meta {
	return MetaMany(MetaOptional(self.inner.meta()))
}

// Parser that returns a given value without consuming anything, created with
// [`pure`](cratepure).
type ParsePure[T any] struct {
	inner T
}

// impl<T: Clone > Parser[T] for ParsePure[T] {
func (self ParsePure[T]) eval(args *State) Result[T] {
	args.current = Invalid[int]()
	return Ok(self.inner)
}

func (self ParsePure[T]) meta() Meta {
	return MetaSkip
}

type ParsePureWith[T any] struct{ inner func() Result[T] }

// impl<T: Clone , F: func()Result[T, E>, E: ToString> Parser<T]
func (self ParsePureWith[T]) eval(_args *State) Result[T] {
	ok, e := self.inner().unpack()
	if e == nil {
		return Ok(ok)
	}
	return Err(Error(MessagePureFailed(e.to_string())))
}

func (self ParsePureWith[T]) meta() Meta {
	return MetaSkip
}

// Parser that fails without consuming any input, created with [`fail`](cratefail).
type ParseFail[T any] struct {
	field1 string
}

// impl[T] Parser[T] for ParseFail[T] {
func (self ParseFail[T]) eval(args *State) Result[T] {
	args.current = Invalid[int]()
	return Err[T](Error(MessageParseFail(self.field1)))
}

func (self ParseFail[T]) meta() Meta {
	return MetaSkip
}

// Parser that transforms parsed value with a function, created with [`map`](Parsermap).
type ParseMap[T, R any] struct {
	inner  Parser[T]
	map_fn func(T) R
}

func (self ParseMap[T, R]) eval(args *State) Result[R] {
	t, e := self.inner.eval(args).unpack()
	if e != nil {
		return Err[R](e)
	}
	return Ok(self.map_fn(t))
}

func (self ParseMap[T, R]) meta() Meta {
	return self.inner.meta()
}

// Create parser from a function, [`construct`](crateconstruct) uses it internally
type ParseCon[T any] struct {
	// inner parser closure
	inner func(bool, *State) Result[T]
	// metas for inner parsers
	meta Meta
	// To produce a better error messages while parsing constructed values
	// we want to look at all the items so values that can be consumed are consumed
	// autocomplete relies on the same logic
	//
	// However when dealing with adjacent restriction detecting the first item relies on failing
	// fast
	failfast bool
}

// impl<T, P> Parser[T] for ParseCon<P>
func (self ParseCon[T]) eval(args *State) Result[T] {
	res := self.inner(self.failfast, args)
	args.current = Invalid[int]()
	return res
}

func (self ParseCon[P]) Meta() Meta {
	return self.meta
}

// Automagically restrict the inner parser scope to accept adjacent values only
//
// `adjacent` can solve surprisingly wide variety of problems: sequential command chaining,
// multi-value arguments, fun.Option-structs to name a few. If you want to run a parser on a
// sequential subset of arguments - `adjacent` might be able to help you. Check the examples
// for better intuition.
//
// Let's consider two examples with consumed items marked in bold and constructor containing
// parsers for `-c` and `-d`.
//
// - <code>**-a** -b **-c** -d</code>
// - <code>**-a** **-c** -b -d</code>
//
// In the first example `-b` breaks the adjacency for all the consumed items so parsing will fail,
// while here in the second one all the consumed items are adjacent to each other so
// parsing will succeed.
//
// # Multi-value arguments
//
// Parsing things like `--point X Y Z`
// #[cfg_attr(not(doctest), doc = include_str!("docs2/adjacent_struct_0.md"))]
//
// # Structure groups
//
// Parsing things like `--rect --width W --height H --rect --height H --width W`
// #[cfg_attr(not(doctest), doc = include_str!("docs2/adjacent_struct_1.md"))]
//
// # Chaining commands
// This example explains [`adjacent`](crateparamsParseCommandadjacent), but the same idea holds.
// Parsing things like `cmd1 --arg1 cmd2 --arg2 --arg3 cmd3 --flag`
//
// #[cfg_attr(not(doctest), doc = include_str!("docs2/adjacent_command.md"))]
//
// # Capturing everything between markers
//
// Parsing things like `find . --exec foo {} -bar ; --more`
//
// #[cfg_attr(not(doctest), doc = include_str!("docs2/adjacent_struct_3.md"))]
//
// # Multi-value arguments with fun.Optional flags
//
// Parsing things like `--foo ARG1 --flag --inner ARG2`
//
// So you can parse things while parsing things. Not sure why you might need this, but you can
// :)
//
// #[cfg_attr(not(doctest), doc = include_str!("docs2/adjacent_struct_4.md"))]
//
// # Performance and other considerations
//
// `bpaf` can run adjacently restricted parsers multiple times to refine the guesses. It's
// best not to have complex inter-fields verification since they might trip up the detection
// logic: instead of restricting, for example "sum of two fields to be 5 or greater" *inside* the
// `adjacent` parser, you can restrict it *outside*, once `adjacent` done the parsing.
//
// There's also similar method [`adjacent`](crateparsersParseArgument) that allows to restrict argument
// parser to work only for arguments where both key and a value are in the same shell word:
// `-f=bar` or `-fbar`, but not `-f bar`.
func (self ParseCon[T]) adjacent() ParseAdjacent[T] {
	self.failfast = true
	return ParseAdjacent[T]{inner: Parser[T]{self.eval, self.Meta}}
}

// Parser that replaces metavar placeholders with actual info in shell completion

type ParseComp[T any] struct {
	inner Parser[T]
	op    func(T) []fun.Pair[string, Option[string]]
	group Option[string]
}

// Attach group name to parsed values
func (self *ParseComp[T]) Group(group string) *ParseComp[T] {
	self.group = Valid(group)
	return self
}

// impl<P, T, F, M> Parser[T] for ParseComp<P, F>
func (self ParseComp[T]) eval(args *State) Result[T] {
	// stash old
	comp_items := []Comp(nil)
	args.swap_comps_with(&comp_items)

	res := self.inner.eval(args)

	// restore old, now metavars added by inner parser, if any, are in comp_items
	args.swap_comps_with(&comp_items)

	if comp, ok := args.comp_mut().Unpack(); ok {
		if res.Err != nil {
			comp.extend_comps(comp_items)
			return res
		}
	}

	// res := res?;

	// completion function generates suggestions based on the parsed inner value, for
	// that `res` must contain a parsed value
	depth := args.depth()
	if comp, ok := args.comp_mut().Unpack(); ok {
		for _, ci := range comp_items {
			is_meta := ci.is_metavar()
			if is_arg, ok := is_meta.Unpack(); ok {
				suggestions := self.op(res.Value)
				// strip metavar when completion makes a single good suggestion
				if len(suggestions) != 1 {
					comp.push_comp(ci)
				}
				for _, sugg := range suggestions {
					replacement, description := sugg.K, sugg.V
					group := self.group
					comp.push_value(
						replacement,
						description,
						group,
						depth,
						is_arg,
					)
				}
			} else {
				comp.push_comp(ci)
			}
		}
	}
	return res
}

func (self ParseComp[T]) meta() Meta {
	return self.inner.meta()
}

/*

pub struct ParseCompStyle<P> {
    inner P
    style: CompleteDecor,
}


impl<P, T> Parser[T] for ParseCompStyle<P>
where
    P: Parser[T] + Sized,
{
    func eval(&self, args: &mut State)Result[T] {
        comp_items := nil;
        args.swap_comps_with(&mut comp_items);
        res := self.inner.eval(args);
        args.swap_comps_with(&mut comp_items);
        args.extend_with_style(self.style, &mut comp_items);
        res
    }

    func meta(&self)Meta {
        self.inner.meta()
    }
}*/

type ParseAdjacent[T any] struct {
	inner Parser[T]
}

// impl<P, T> Parser[T] for ParseAdjacent<P>
func (self ParseAdjacent[T]) eval(args *State) Result[T] {
	original_scope := args.scope

	var first_item Item
	inner_meta := self.inner.meta()

	item, ok := Metafirst_item(inner_meta).Unpack()
	if !ok {
		panic("bpaf usage BUG: adjacent should start with a required argument")
	}
	best_error := func() Message {
		first_item = item
		missing_item := MissingItem{
			item:     item,
			position: original_scope.start(),
			scope:    original_scope,
		}
		return MessageMissing(missing_item)
	}()
	best_args := args
	best_consumed := 0

	for start, width /*, mut this_arg*/ := range args.ranges(first_item) {
		var this_arg State
		// since we only want to parse things to the right of the first item we perform
		// parsing in two passes:
		// - try to run the parser showing only single argument available at all the indices
		// - try to run the parser showing starting at that argument and to the right of it
		// this means constructing argument parsers from req flag and positional works as
		// expected:
		// consider examples "42 -n" and "-n 42"
		// without multi step approach first command line also parses into 42
		scratch := this_arg
		scratch.set_scope(Range{start, start + width})
		before := scratch.len()

		// nothing to consume, might as well skip this segment right now
		// it will most likely fail, but it doesn't matter, we are only looking for the
		// left most match
		if before == 0 {
			continue
		}

		_ = self.inner.eval(&scratch)

		if before == scratch.len() {
			// failed to consume anything which means we don't start parsing at this point
			continue
		}

		this_arg.set_scope(Range{start, original_scope.end()})
		before = this_arg.len()

		// values consumed by adjacent must be actually adjacent - if a scope contains
		// already parsed values inside we need to trim it
		if original_scope.end-start > before {
			this_arg.set_scope(this_arg.adjacently_available_from(start))
		}

		for {
			res, err := self.inner.eval(&this_arg).unpack()
			if err == nil {
				// there's a smaller adjacent scope, we must try it before returning.
				if adj_scope, ok := this_arg.adjacent_scope(*args).Unpack(); ok {
					this_arg = *args
					this_arg.set_scope(adj_scope)
				} else {
					stdmemswap(args, &this_arg)
					args.set_scope(original_scope)
					return Ok(res)
				}
			} else {
				consumed := before - this_arg.len()
				if consumed > best_consumed {
					best_consumed = consumed
					stdmemswap(best_args, &this_arg)
					best_error = Message(err)
				}
				break
			}
		}
	}

	stdmemswap(args, best_args)
	return Err[T](Error(best_error))
}

func (self ParseAdjacent[T]) meta() Meta {
	meta := self.inner.meta()
	return MetaAdjacent(meta)
}
