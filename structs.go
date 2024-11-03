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
	res, err := self.inner.eval(args)
	args.swap_comps_with(&comp_items)
	args.push_with_group(self.message.to_completion(), &comp_items)
	return res, err
}

func (self ParseGroupHelp[T]) meta() Meta {
	meta := self.inner.meta
	return MetaSubsection{meta, self.message}
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

	return MetaSubsection{meta, buf}
}

// Parser that returns results as usual but not shown in `--help` output, created with [Parserhide]
type ParseHide[T any] struct {
	inner Parser[T]
}

func (self ParseHide[T]) eval(args *State) (T, Error) {
	var comps []Comp

	args.swap_comps_with(&comps)

	res, err := self.inner.eval(args)

	args.swap_comps_with(&comps)
	if err != nil && MessageF(err, MessageFs[bool]{Missing: func([]MissingItem) bool { return true }}) {
		return *new(T), ErrorMissing(nil)
	}

	return res, err
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

		_, missing := err.(ErrorMissing)
		if catch || missing && orig_args.len() == args.len() || !missing && can_catch(Error(err)) {
			*orig_args = *args

			if _, ok := orig_args.comp_mut(); ok {
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
