package cli

import (
	"cmp"
	"fmt"
	"iter"

	"github.com/rprtr258/fun"
)

// Parser that substitutes missing value with a function results but not parser
// failure, created with [fallback_with](Parserfallback_with).
type ParseFallbackWith[T any] struct {
	inner     Parser[T]
	fallback  func() Result[T]
	value_str string
}

func (self ParseFallbackWith[T]) eval(args *State) Result[T] {
	clone := args
	ok, e := self.inner.eval(clone)
	if e == nil {
		stdmemswap(args, clone)
		return Ok(ok)
	} else {
		args.swap_comps(clone)
		if Message(e).can_catch() {
			ok, e := self.fallback().unpack()
			if e == nil {
				return Ok(ok)
			} else {
				return Err[T](Error(MessagePureFailed(e.String())))
			}
		} else {
			return Err[T](Error(e))
		}
	}
}

func (self ParseFallbackWith[T]) meta() Meta {
	m := MetaOptional(self.inner.meta)
	if self.value_str == "" {
		return m
	} else {
		buf := newDocFromString(self.value_str)
		return MetaSuffix(m, buf)
	}
}

// Parser with attached message to several fields, created with [group_help](Parsergroup_help).
type ParseGroupHelp[T any] struct {
	inner   Parser[T]
	message Doc
}

func (self ParseGroupHelp[T]) Parser() Parser[T] {
	return Parser[T]{self.eval, self.meta()}
}

func (self ParseGroupHelp[T]) eval(args *State) (T, Error) {
	comp_items := []Comp(nil)
	args.swap_comps_with(&comp_items)
	res := newResult(self.inner.eval(args))
	args.swap_comps_with(&comp_items)
	args.push_with_group(self.message.to_completion(), &comp_items)
	return res.unpack()
}

func (self ParseGroupHelp[T]) meta() Meta {
	meta := self.inner.meta
	return MetaSubsection(meta, self.message)
}

// Parser with attached message to several fields, created with [group_help](Parsergroup_help).
type ParseWithGroupHelp[T any] struct {
	inner Parser[T]
	f     func(MetaInfo) Doc
}

func (self ParseWithGroupHelp[T]) eval(args *State) (T, Error) {
	return self.inner.eval(args)
}

func (self ParseWithGroupHelp[T]) meta() Meta {
	meta := self.inner.meta
	buf := (self.f)(MetaInfo(meta))

	return MetaSubsection(meta, buf)
}

// Parser that returns results as usual but not shown in `--help` output, created with [Parserhide]
type ParseHide[T any] struct {
	inner Parser[T]
}

func (self ParseHide[T]) eval(args *State) (T, Error) {
	var comps []Comp

	args.swap_comps_with(&comps)

	res := newResult(self.inner.eval(args))

	args.swap_comps_with(&comps)
	if res.Err != nil && MessageF(Message(res.Err), MessageFs[bool]{Missing: func([]MissingItem) bool { return true }}) {
		return *new(T), Error(MessageMissing())
	}

	return res.unpack()
}

// Parser that hides inner parser from usage line
//
// No other changes to the inner parser
type ParseUsage[T any] struct {
	inner Parser[T]
	usage Doc
}

func (self ParseUsage[T]) eval(args *State) (T, Error) {
	return self.inner.eval(args)
}

func (self ParseUsage[P]) meta() Meta {
	return MetaCustomUsage(self.inner.meta, self.usage)
}

// Parser that tries to either of two parsers and uses one that succeeeds, created with [Parseror_else].
type ParseOrElse[T any] struct {
	this Parser[T]
	that Parser[T]
}

func (self ParseOrElse[T]) eval(args *State) (T, Error) {
	var comp_items []Comp

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

	ok, err := self.this.eval(args_a)
	var (
		res_a Option[T]
		err_a Option[Error]
	)
	if err != nil {
		res_a, err_a = Invalid[T](), Valid(err)
	} else {
		res_a, err_a = Valid(ok), Invalid[Error]()
	}

	ok, err = self.that.eval(args_b)
	var (
		res_b Option[T]
		err_b Option[Error]
	)
	if err != nil {
		res_b, err_b = Invalid[T](), Valid(err)
	} else {
		res_b, err_b = Valid(ok), Invalid[Error]()
	}

	zalupa, err := this_or_that_picks_first(
		err_a,
		err_b,
		args,
		args_a,
		args_b,
		&comp_items,
	).unpack()
	if err != nil {
		return *new(T), err
	}
	if zalupa {
		return res_a.unwrap(), nil
	} else {
		return res_b.unwrap(), nil
	}
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
	switch ccc := cmp.Compare(args_a.depth(), args_b.depth()); {
	case ccc < 0:
		stdmemswap(args, args_b)

		if comp, ok := args.comp_mut().Unpack(); ok {
			comp.extend_comps(*comp_stash)
		}
		err, ok := err_b.Unpack()
		if !ok {
			return Ok(false)
		}
		return Err[bool](err)
	case ccc == 0:
	case ccc > 0:
		stdmemswap(args, args_a)

		if comp, ok := args.comp_mut().Unpack(); ok {
			comp.extend_comps(*comp_stash)
		}
		err, ok := err_a.Unpack()
		if !ok {
			return Ok(true)
		}
		return Err[bool](err)
	}

	// otherwise pick based on the left most or successful one
	e1, a_ok := err_a.Unpack()
	e2, b_ok := err_b.Unpack()
	type resT fun.Pair[bool, Option[uint]]
	var res Result[resT]
	switch {
	case !a_ok && !b_ok:
		if args.len() == args_a.len() && args_a.len() == args_b.len() {
			res = Ok(resT{true, Invalid[uint]()})
		} else {
			a, b := args_a.pick_winner(*args_b)
			res = Ok(resT{a, b})
		}
	case a_ok && b_ok:
		res = Err[resT](e1.combine_with(e2))
	// otherwise either a or b are success, true means a is success
	default:
		res = Ok(resT{!a_ok, Invalid[uint]()})
	}

	{
		keep_a := true
		keep_b := true
		if args_a.len() != args_b.len() {
			// If neither parser consumed anything - both can produce valid completions, otherwise
			// look for the first "significant" consume and keep that parser
			//
			// This is needed to preserve completion from a choice between a positional and a flag
			// See https://github.com/pacak/bpaf/issues/303 for more details
			_, ok1 := args_a.comp_mut().Unpack()
			_, ok2 := args_b.comp_mut().Unpack()
			if ok1 && ok2 {
			CHECK:
				for ix, arg := range args_a.items {
					// During completion process named and unnamed arguments behave
					// different - `-` and `--` are positional arguments, but we want to produce
					// named items too. An empty string is also a special type of item that
					// gets passed when user starts completion without passing any actual data.
					//
					// All other strings are either valid named items or valid positional items
					// those are hopefully follow the right logic for being parsed/not parsed
					if ix+1 == len(args_a.items) {
						os := arg.str
						if os == "" || os == "-" || os == "--" {
							break CHECK
						}
					}
					a, ok1 := args_a.present(ix).Unpack()
					b, ok2 := args_b.present(ix).Unpack()
					if ok1 && ok2 {
						switch {
						case b:
							keep_b = false
							break CHECK
						case a:
							keep_a = false
							break CHECK
						}
					}
				}
			}
		}

		a, ok1 := args_a.comp_mut().Unpack()
		b, ok2 := args_b.comp_mut().Unpack()
		if ok1 && ok2 {
			if keep_a {
				*comp_stash = append(*comp_stash, a.drain_comps()...)
			}
			if keep_b {
				*comp_stash = append(*comp_stash, b.drain_comps()...)
			}
		}
	}

	if bbb, errrr := res.unpack(); errrr == nil { // or else no winner, keep the completions but don't touch args otherwise
		ix := bbb.V
		if bbb.K {
			if win, ok := ix.Unpack(); ok {
				args_a.save_conflicts(*args_b, int(win))
			}
			stdmemswap(args, args_a)
		} else {
			if win, ok := ix.Unpack(); ok {
				args_b.save_conflicts(*args_a, int(win))
			}
			stdmemswap(args, args_b)
		}
	}

	if comp, ok := args.comp_mut().Unpack(); ok {
		comp.extend_comps(*comp_stash)
	}

	resres, err := res.unpack()
	if err != nil {
		return Err[bool](err)
	}
	return Ok(resres.K)
}

// Parser that transforms parsed value with a failing function, created with [parse](Parserparse)
type ParseWith[T, R any] struct {
	inner    Parser[T]
	parse_fn func(T) Result[R]
}

func (self ParseWith[T, R]) Parser() Parser[R] {
	return Parser[R]{self.eval, self.inner.meta}
}

func (self ParseWith[T, R]) eval(args *State) (R, Error) {
	t, err := self.inner.eval(args)
	if err != nil {
		return *new(R), err
	}

	r, err := self.parse_fn(t).unpack()
	if err != nil {
		return *new(R), Error(MessageParseFailed(args.current, err.String()))
	}

	return r, nil
}

// Parser that substitutes missing value but not parse failure, created with [fallback](Parserfallback).
type ParseFallback[T any] struct {
	inner     Parser[T]
	value     T
	value_str string
}

func (self ParseFallback[T]) Parser() Parser[T] {
	return Parser[T]{self.eval, self.meta()}
}

func (self ParseFallback[T]) eval(args *State) (T, Error) {
	clone := args
	ok, e := self.inner.eval(clone)
	if e != nil {
		args.swap_comps(clone)
		if Message(e).can_catch() {
			return self.value, nil
		} else {
			return *new(T), Error(e)
		}
	}
	*args = *clone
	return ok, nil
}

func (self ParseFallback[T]) meta() Meta {
	m := MetaOptional(self.inner.meta)
	if self.value_str == "" {
		return m
	} else {
		buf := newDocFromString(self.value_str)
		return MetaSuffix(m, buf)
	}
}

// Show [fallback](Parserfallback) value in `--help` using [Display](stdfmtDisplay)
// representation
//
// #[cfg_attr(not(doctest), doc = include_str!("docs2/dis_fallback.md"))]
func (self *ParseFallback[T]) DisplayFallback() ParseFallback[T] {
	self.value_str = fmt.Sprintf("[default: %v]", self.value)
	return *self
}

// Show [fallback](Parserfallback) value in `--help` using [Debug](stdfmtDebug)
// representation
//
// #[cfg_attr(not(doctest), doc = include_str!("docs2/deb_fallback_with.md"))]
func (self *ParseFallback[T]) DebugFallback() ParseFallback[T] {
	self.value_str = fmt.Sprintf("[default: %#v]", self.value)
	return *self
}

// Show [fallback_with](Parserfallback_with) value in `--help` using [Display](stdfmtDisplay)
// representation
//
// # If fallback function fails - no value will show up
//
// #[cfg_attr(not(doctest), doc = include_str!("docs2/dis_fallback_with.md"))]
func (self *ParseFallbackWith[T]) DisplayFallback() ParseFallbackWith[T] {
	if val, err := self.fallback().unpack(); err == nil {
		self.value_str = fmt.Sprintf("[default: %v]", val)
	}
	return *self
}

// Show [fallback_with](Parserfallback_with) value in `--help` using [Debug](stdfmtDebug)
// representation
//
// # If fallback function fails - no value will show up
//
// #[cfg_attr(not(doctest), doc = include_str!("docs2/deb_fallback.md"))]
func (self *ParseFallbackWith[T]) debug_fallback() *ParseFallbackWith[T] {
	val, err := self.fallback().unpack()
	if err == nil {
		self.value_str = fmt.Sprintf("[default: %#v]", val)
	}
	return self
}

// Parser fails with a message if check returns false, created with [guard](Parserguard).
type ParseGuard[T any] struct {
	inner   Parser[T]
	check   func(T) bool
	message string
}

func (self ParseGuard[T]) Parser() Parser[T] {
	return Parser[T]{self.eval, self.inner.meta}
}

func (self ParseGuard[T]) eval(args *State) (T, Error) {
	t, err := self.inner.eval(args)
	if err != nil {
		return *new(T), err
	}
	if self.check(t) {
		return t, nil
	} else {
		return *new(T), Error(MessageGuardFailed(args.current, self.message))
	}
}

// try to parse
func parse_option[T any](
	parser Parser[T],
	len *uint,
	args *State,
	catch bool,
) (Option[T], Error) {
	orig_args := clone(args)
	val, err := parser.eval(args)
	if err != nil {
		// this is safe to return Ok(Invalid()) in following scenarios
		// when inner parser never consumed anything and
		// 1. produced ErrorMissing
		// 2. produced ErrorMessage(_, true)
		// 3. produced ErrorMessage and catch is enabled
		//
		// In all other scenarios we should return the original error
		//
		// When parser returns Ok(Invalid()) we should return the original arguments so if there's
		// anything left unconsumed - this won't be lost.

		missing := Message(err).IsMissing()

		if catch || (missing && orig_args.len() == args.len()) || (!missing && Message(err).can_catch()) {
			*orig_args = *args

			if orig_args.comp_mut().Valid {
				args.swap_comps(orig_args)
			}
			return Invalid[T](), nil
		} else {
			return Option[T]{}, err
		}
	}
	// we keep including values for as long as we consume values from the argument
	// list or at least one value
	if lll := uint(args.len()); lll < *len {
		*len = lll
		return Valid(val), nil
	} else {
		return Invalid[T](), nil
	}
}

// Parser that replaces metavar placeholders with actual info in shell completion

type ParseComp[T any] struct {
	inner Parser[T]
	op    func(T) iter.Seq2[string, Option[string]]
	group Option[string]
}

func (self ParseComp[T]) Parser() Parser[T] {
	return Parser[T]{self.eval, self.inner.meta}
}

// Attach group name to parsed values
func (self *ParseComp[T]) Group(group string) *ParseComp[T] {
	self.group = Valid(group)
	return self
}

func (self ParseComp[T]) eval(args *State) (T, Error) {
	// stash old
	comp_items := []Comp(nil)
	args.swap_comps_with(&comp_items)

	res := newResult(self.inner.eval(args))

	// restore old, now metavars added by inner parser, if any, are in comp_items
	args.swap_comps_with(&comp_items)

	if comp, ok := args.comp_mut().Unpack(); ok {
		if res.Err != nil {
			comp.extend_comps(comp_items)
			return res.unpack()
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
				// TODO: get back if
				// if len(suggestions) != 1 {
				// 	comp.push_comp(ci)
				// }
				for replacement, description := range suggestions {
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
	return res.unpack()
}

// type ParseCompStyle[T any] struct {
// 	inner Parser[T]
// 	style CompleteDecor
// }

// func (self ParseCompStyle[T]) eval(args *State) (T, Error) {
// 	comp_items := nil
// 	args.swap_comps_with(&comp_items)
// 	res := self.inner.eval(args)
// 	args.swap_comps_with(&comp_items)
// 	args.extend_with_style(self.style, &comp_items)
// 	return res
// }

// func (self ParseCompStyle[T]) meta() Meta {
// 	return self.inner.meta()
// }

type ParseAdjacent[T any] struct {
	inner Parser[T]
}

func (self ParseAdjacent[T]) Parser() Parser[T] {
	return Parser[T]{self.eval, MetaAdjacent(self.inner.meta)}
}

func (self ParseAdjacent[T]) eval(args *State) (T, Error) {
	original_scope := args.scope

	var first_item Item
	inner_meta := self.inner.meta

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

	iter := args.ranges(first_item)
	for {
		start_width, ok := iter.next()
		if !ok {
			break
		}
		start, width, this_arg := start_width.startOffset, start_width.widthOfTheFirstItem, start_width.state

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

		_, _ = self.inner.eval(&scratch)

		if before == scratch.len() {
			// failed to consume anything which means we don't start parsing at this point
			continue
		}

		this_arg.set_scope(Range{start, original_scope.end()})
		before = this_arg.len()

		// values consumed by adjacent must be actually adjacent - if a scope contains
		// already parsed values inside we need to trim it
		if original_scope.end()-start > before {
			this_arg.set_scope(this_arg.adjacently_available_from(start))
		}

		for {
			res, err := self.inner.eval(&this_arg)
			if err == nil {
				// there's a smaller adjacent scope, we must try it before returning.
				if adj_scope, ok := this_arg.adjacent_scope(*args).Unpack(); ok {
					this_arg = *args
					this_arg.set_scope(adj_scope)
				} else {
					stdmemswap(args, &this_arg)
					args.set_scope(original_scope)
					return res, nil
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
	return *new(T), Error(best_error)
}
