package cli

import "math"

// Apply inner parser, return a value in `Some` if items requested by it are all present, restore
// and return `None` if any are missing. Created with [Optional].
func optional_eval[T any](inner Parser[T], catch bool, state *state) (Opt[T], parseError) {
	len := uint(math.MaxUint)
	return parse_option(inner, &len, state, catch)
}

// Make a parser optional
//
// Optional produces an [Opt] with the parsed value or
// nothing if the parser fails softly, without consuming input.
// See [OptionalCatch] for the variant that also catches consumed-input failures.
func Optional[T any, P parser[T]](self P) Parser[Opt[T]] {
	inner := self.Parser()
	return Parser[Opt[T]]{
		eval: func(state *state) (Opt[T], parseError) { return optional_eval(inner, false, state) },
		meta: metaOptional{inner.meta},
	}
}

// OptionalCatch makes a parser optional and, additionally, restores all
// consumed values and yields nothing if the inner parser fails.
//
// This is the behavior [Construct] applies to fields with an Opt[T] type.
func OptionalCatch[T any, P parser[T]](self P) Parser[Opt[T]] {
	inner := self.Parser()
	return Parser[Opt[T]]{
		eval: func(state *state) (Opt[T], parseError) { return optional_eval(inner, true, state) },
		meta: metaOptional{inner.meta},
	}
}
