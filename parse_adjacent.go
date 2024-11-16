package cli

type ParseAdjacent[T any] struct {
	inner Parser[T]
}

func (self ParseAdjacent[T]) Parser() Parser[T] {
	return Parser[T]{self.eval, MetaAdjacent{self.inner.meta}}
}

func (self ParseAdjacent[T]) eval(args *State) (T, Error) {
	original_scope := args.scope

	var first_item Item
	best_error := func() Error {
		item, ok := Metafirst_item(self.inner.meta).Unpack()
		if !ok {
			panic("bpaf usage BUG: adjacent should start with a required argument")
		}

		first_item = item
		missing_item := MissingItem{
			item:     item.clone(),
			position: original_scope.start(),
			scope:    original_scope,
		}
		return ErrorMissing([]MissingItem{missing_item})
	}()
	best_args := args.clone()
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
		scratch := this_arg.clone()
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
				if adj_scope, ok := this_arg.adjacent_scope(*args); ok {
					this_arg = args.clone()
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
					stdmemswap(&best_args, &this_arg)
					best_error = Error(err)
				}
				break
			}
		}
	}

	stdmemswap(args, &best_args)
	return *new(T), Error(best_error)
}
