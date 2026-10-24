package cli

func adjacent_eval[T any](inner Parser[T], state *state) (T, parseError) {
	original_scope := state.scope

	var first_item item
	best_error := func() parseError {
		item, ok := metaFirstItem(inner.meta).Unpack()
		if !ok {
			panic("cli usage BUG: adjacent should start with a required argument")
		}

		first_item = item
		missing_item := missingItem{
			item:     item.clone(),
			position: original_scope.start(),
			scope:    original_scope,
		}
		return errorMissing([]missingItem{missing_item})
	}()
	best_args := state.clone()
	best_consumed := 0

	next := state.ranges(first_item)
	for {
		start_width, ok := next()
		if !ok {
			break
		}

		start := start_width.startOffset
		width := start_width.widthOfTheFirstItem
		this_arg := start_width.state

		// since we only want to parse things to the right of the first item we perform
		// parsing in two passes:
		// - try to run the parser showing only single argument available at all the indices
		// - try to run the parser showing starting at that argument and to the right of it
		// this means constructing argument parsers from req flag and positional works as
		// expected:
		// consider examples "42 -n" and "-n 42"
		// without multi step approach first command line also parses into 42
		scratch := this_arg.clone()
		scratch.scope = rng{start, start + width}
		before := scratch.len()

		// nothing to consume, might as well skip this segment right now
		// it will most likely fail, but it doesn't matter, we are only looking for the
		// left most match
		if before == 0 {
			continue
		}

		_, _ = inner.eval(&scratch)

		if before == scratch.len() {
			// failed to consume anything which means we don't start parsing at this point
			continue
		}

		this_arg.scope = rng{start, original_scope.end()}
		before = this_arg.len()

		// values consumed by adjacent must be actually adjacent - if a scope contains
		// already parsed values inside we need to trim it
		if original_scope.end()-start > before {
			this_arg.scope = this_arg.adjacently_available_from(start)
		}

		for {
			res, err := inner.eval(&this_arg)
			if err == nil {
				// there's a smaller adjacent scope, we must try it before returning.
				if adj_scope, ok := this_arg.adjacent_scope(*state); ok {
					this_arg = state.clone()
					this_arg.scope = adj_scope
				} else {
					*state, this_arg = this_arg, *state
					state.scope = original_scope
					return res, nil
				}
			} else {
				consumed := before - this_arg.len()
				if consumed > best_consumed {
					best_consumed = consumed
					best_args, this_arg = this_arg, best_args
					best_error = parseError(err)
				}
				break
			}
		}
	}

	*state, best_args = best_args, *state
	return *new(T), parseError(best_error)
}
