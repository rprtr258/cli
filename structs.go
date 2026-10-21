package cli

// try to parse
func parse_option[T any](
	parser Parser[T],
	len *uint,
	state *State,
	catch bool,
) (Option[T], Error) {
	stateOrig := state.clone()
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
			missing && stateOrig.len() == state.len() ||
			!missing && can_catch(Error(err)) {
			if !stateOrig.comp.Valid {
				*state = stateOrig
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
