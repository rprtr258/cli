package cli

// Parser with attached message to several fields, created with [group_help](Parsergroup_help).
type ParseGroupHelp[T any] struct {
	inner   Parser[T]
	message Doc
}

func (self ParseGroupHelp[T]) Parser() Parser[T] {
	return Parser[T]{self.eval, MetaSubsection{self.inner.meta, self.message}}
}

func (self ParseGroupHelp[T]) eval(state *State) (T, Error) {
	comp_items := []Comp(nil)
	state.swap_comps_with(&comp_items)
	res, err := self.inner.eval(state)
	state.swap_comps_with(&comp_items)
	state.push_with_group(self.message.to_completion(), &comp_items)
	return res, err
}

// Parser with attached message to several fields, created with [group_help](Parsergroup_help).
type ParseWithGroupHelp[T any] struct {
	inner Parser[T]
	f     func(Meta) Doc
}

func (self ParseWithGroupHelp[T]) Parser() Parser[T] {
	return Parser[T]{self.eval, self.meta()}
}

func (self ParseWithGroupHelp[T]) eval(state *State) (T, Error) {
	return self.inner.eval(state)
}

func (self ParseWithGroupHelp[T]) meta() Meta {
	meta := self.inner.meta
	buf := self.f(meta)
	return MetaSubsection{meta, buf}
}

// Parser that returns results as usual but not shown in `--help` output, created with [Parserhide]
type ParseHide[T any] struct {
	inner Parser[T]
}

func (self ParseHide[T]) eval(state *State) (T, Error) {
	var comps []Comp

	state.swap_comps_with(&comps)

	res, err := self.inner.eval(state)

	state.swap_comps_with(&comps)
	if _, ok := err.(ErrorMissing); err != nil && ok {
		return *new(T), ErrorMissing(nil)
	}

	return res, err
}

// try to parse
func parse_option[T any](
	parser Parser[T],
	len *uint,
	state *State,
	catch bool,
) (Option[T], Error) {
	orig_args := state.clone()
	val, err := parser.eval(state)
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

		if _, missing := err.(ErrorMissing); catch ||
			missing && orig_args.len() == state.len() ||
			!missing && can_catch(Error(err)) {
			orig_args, *state = *state, orig_args
			if orig_args.comp.Valid {
				state.swap_comps(&orig_args)
			}
			return Invalid[T](), nil
		} else {
			return Option[T]{}, err
		}
	}
	// we keep including values for as long as we consume values from the argument
	// list or at least one value
	if lll := uint(state.len()); lll < *len {
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

// func (self ParseCompStyle[T]) eval(state *State) (T, Error) {
// 	comp_items := nil
// 	state.swap_comps_with(&comp_items)
// 	res := self.inner.eval(state)
// 	state.swap_comps_with(&comp_items)
// 	state.extend_with_style(self.style, &comp_items)
// 	return res
// }

// func (self ParseCompStyle[T]) meta() Meta {
// 	return self.inner.meta()
// }
