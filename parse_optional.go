package cli

import "math"

// optional_eval applies the inner parser, returns the value if all items it
// requests are present, restores the state and returns an empty [Opt] if any
// are missing. Created with [Parser.Optional] or [Parser.OptionalCatch].
func optional_eval[T any](inner Parser[T], catch bool, state *state) (Opt[T], parseError) {
	maxlen := uint(math.MaxUint)
	return parse_option(inner, &maxlen, state, catch)
}

// Optional makes a parser optional
//
// Optional produces an [Opt] with the parsed value or
// nothing if the parser fails softly, without consuming input.
// See [Parser.OptionalCatch] for the variant that also catches consumed-input failures.
func (self Parser[T]) Optional[R Opt[T]]() Parser[R] {
	inner := self.Parser()
	return Parser[R]{
		eval: func(state *state) (R, parseError) {
			x, err := optional_eval(inner, false, state)
			return R(x), err
		},
		meta: metaOptional{inner.meta},
	}
}

// OptionalCatch makes a parser optional and, additionally, restores all
// consumed values and yields nothing if the inner parser fails.
//
// This is the behavior [Construct] applies to fields with an Opt[T] type.
func (self Parser[T]) OptionalCatch[R Opt[T]]() Parser[R] {
	inner := self.Parser()
	return Parser[R]{
		eval: func(state *state) (R, parseError) {
			x, err := optional_eval(inner, true, state)
			return R(x), err
		},
		meta: metaOptional{inner.meta},
	}
}
