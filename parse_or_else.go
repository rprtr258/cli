package cli

import (
	"cmp"

	"github.com/rprtr258/fun"
)

// Parser that tries to either of two parsers and uses one that succeeeds, created with [Parseror_else].
type ParseOrElse[T any] struct {
	this Parser[T]
	that Parser[T]
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
) (bool, Error) {
	// if higher depth parser succeeds - it takes a priority
	// completion from different depths should never mix either
	switch ccc := cmp.Compare(args_a.depth(), args_b.depth()); {
	case ccc < 0:
		stdmemswap(args, args_b)

		if comp, ok := args.comp_mut(); ok {
			comp.extend_comps(*comp_stash)
		}
		err, ok := err_b.Unpack()
		if !ok {
			return false, nil
		}
		return false, err
	case ccc == 0:
	case ccc > 0:
		stdmemswap(args, args_a)

		if comp, ok := args.comp_mut(); ok {
			comp.extend_comps(*comp_stash)
		}
		err, ok := err_a.Unpack()
		if !ok {
			return true, nil
		}
		return false, err
	}

	// otherwise pick based on the left most or successful one
	e1, a_ok := err_a.Unpack()
	e2, b_ok := err_b.Unpack()
	type resT fun.Pair[bool, Option[uint]]
	var (
		res    resT
		resErr Error
	)
	switch {
	case !a_ok && !b_ok:
		if args.len() == args_a.len() && args_a.len() == args_b.len() {
			res = resT{true, Invalid[uint]()}
		} else {
			a, b := args_a.pick_winner(*args_b)
			res = resT{a, b}
		}
	case a_ok && b_ok:
		resErr = combine_with(e1, e2)
	// otherwise either a or b are success, true means a is success
	default:
		res = resT{!a_ok, Invalid[uint]()}
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
			_, ok1 := args_a.comp_mut()
			_, ok2 := args_b.comp_mut()
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
					a, ok1 := args_a.present(ix)
					b, ok2 := args_b.present(ix)
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

		a, ok1 := args_a.comp_mut()
		b, ok2 := args_b.comp_mut()
		if ok1 && ok2 {
			if keep_a {
				*comp_stash = append(*comp_stash, a.drain_comps()...)
			}
			if keep_b {
				*comp_stash = append(*comp_stash, b.drain_comps()...)
			}
		}
	}

	if resErr == nil { // or else no winner, keep the completions but don't touch args otherwise
		ix := res.V
		if res.K {
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

	if comp, ok := args.comp_mut(); ok {
		comp.extend_comps(*comp_stash)
	}

	if resErr != nil {
		return false, resErr
	}
	return res.K, nil
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

	// run both parsers, expand Result[T, Error] into Option[T] + fun.Option[Error]
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
	)
	if err != nil {
		return *new(T), err
	}
	if zalupa {
		return res_a.unwrap(), nil
	} else {
		return res_b.unwrap(), nil
	}
}

// Choose between several parsers specified at runtime
//
// You can use this function to create multiple parsers that produce the same type of value at a runtime
// and let bpaf to pick one that best fits best. This function is designed to work in Combinatoric
// API, but you can use it in Derive API with `extern`.
//
// #[cfg_attr(not(doctest), doc = include_str!("docs2/Choice.md"))]
func Choice[T any](parsers ...Parser[T]) Parser[T] {
	if len(parsers) == 0 {
		p := Fail[T]("Invalid choice usage")
		return Parser[T]{p.eval, MetaSkip{}}
	}

	this := parsers[0]
	for _, that := range parsers[1:] {
		p := ParseOrElse[T]{this, that}
		this = Parser[T]{p.eval, or(p.this.meta, p.that.meta)}
	}
	return this
}
