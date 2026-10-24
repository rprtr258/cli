package cli

// Pure returns a parser that produces a given value without consuming anything,
// created with [Pure].
func Pure[T any](val T) Parser[T] {
	return Parser[T]{
		eval: func(state *state) (T, parseError) {
			state.currentIdx = Invalid[int]()
			return val, nil
		},
		meta: metaSkip{},
	}
}

// PureWith returns a parser that produces a fixed value.
//
// This parser produces `T` without consuming anything from the command line, which can be useful
// with [Construct]. As with any parsers, `T` should be `Clone` and `Debug`.
//
// Both `Pure` and [PureWith] are designed to put values into structures, to generate fallback
// you should be using [Parser.Fallback] and [Parser.FallbackWith].
//
// See also [Pure] for a computation that cannot fail.
func PureWith[T any](val func() (T, error)) Parser[T] {
	return Parser[T]{
		eval: func(*state) (T, parseError) {
			ok, e := val()
			if e == nil {
				return ok, nil
			}
			return *new(T), errorPureFailed(e.Error())
		},
		meta: metaSkip{},
	}
}
