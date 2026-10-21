package cli

// Parser that returns a given value without consuming anything, created with [Pure].
func Pure[T any](val T) Parser[T] {
	return Parser[T]{
		eval: func(state *State) (T, Error) {
			state.currentIdx = Invalid[int]()
			return val, nil
		},
		meta: MetaSkip{},
	}
}

// Parser that produces a fixed value
//
// This parser produces `T` without consuming anything from the command line, which can be useful
// with [Construct]. As with any parsers, `T` should be `Clone` and `Debug`.
//
// Both `Pure` and [PureWith] are designed to put values into structures, to generate fallback
// you should be using [Parser.Fallback] and [Parser.FallbackWith].
//
// See also [PureWith] for a Pure computation that can fail.
func PureWith[T any](val func() (T, Error)) Parser[T] {
	return Parser[T]{
		eval: func(*State) (T, Error) {
			ok, e := val()
			if e == nil {
				return ok, nil
			}
			return *new(T), ErrorPureFailed(errorToString(e))
		},
		meta: MetaSkip{},
	}
}
