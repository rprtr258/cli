package cli

import "math"

// Apply inner parser as many times as it succeeds, return the Last value
func last_eval[T any](inner Parser[T], state *state) (T, parseError) {
	last := Invalid[T]()
	current := state.len()
	len := uint(math.MaxUint)
	for {
		val, err := parse_option(inner, &len, state, false)
		if err != nil || !val.Valid {
			break
		}
		last = val
		if current == state.len() {
			break
		}
		current = state.len()
	}
	if last, ok := last.Unpack(); ok {
		return last, nil
	}
	return inner.eval(state)
}

// Apply the inner parser as many times as it succeeds, return the Last value
//
// # You can use this to allow users to pick contradicting options
//
// # Example
//
//	pick := cli.Choice(flagA.Last(), flagB.Last())
func (self Parser[T]) Last() Parser[T] {
	inner := self
	return Parser[T]{
		eval: func(state *state) (T, parseError) { return last_eval(inner, state) },
		meta: metaMany{metaRequired{inner.meta}},
	}
}
