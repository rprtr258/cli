package cli

// try to parse
func parse_option[T any](
	parser Parser[T],
	len *uint,
	state *state,
	catch bool,
) (Opt[T], parseError) {
	stateOrig := state.clone()
	val, err := parser.eval(state)
	if err != nil {
		// this is safe to return Ok(Invalid()) in following scenarios
		// when inner parser never consumed anything and
		// 1. produced errorMissing
		// 2. produced errorMessage(_, true)
		// 3. produced errorMessage and catch is enabled
		//
		// In all other scenarios we should return the original error
		//
		// When parser returns Ok(Invalid()) we should return the original arguments so if there's
		// anything left unconsumed - this won't be lost.

		if _, missing := err.(errorMissing); catch ||
			missing && stateOrig.len() == state.len() ||
			!missing && can_catch(parseError(err)) {
			if !stateOrig.comp.Valid {
				*state = stateOrig
			}
			return Invalid[T](), nil
		} else {
			return Opt[T]{}, err
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

// func (self ParseCompStyle[T]) eval(state *state) (T, parseError) {
// 	comp_items := nil
// 	state.swap_comps_with(&comp_items)
// 	res := self.inner.eval(state)
// 	state.swap_comps_with(&comp_items)
// 	state.extend_with_style(self.style, &comp_items)
// 	return res
// }

// func (self ParseCompStyle[T]) meta() meta {
// 	return self.inner.meta()
// }
