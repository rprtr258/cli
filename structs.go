package cli

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
