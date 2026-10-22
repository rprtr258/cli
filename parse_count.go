package cli

import "math"

// Apply inner parser as many times as it succeeds while consuming something and return this number
func count_eval[T any](inner Parser[T], state *state) (int, parseError) {
	res := 0
	current := state.len()
	maxlen := uint(math.MaxUint)
	for {
		vv, err := parse_option(inner, &maxlen, state, false)
		if err != nil {
			return 0, err
		}
		if !vv.Valid {
			break
		}
		res += 1
		if current == state.len() {
			break
		}
		current = state.len()
	}
	return res, nil
}

// Count how many times the inner parser succeeds, and return that number.
//
// When you are dealing with a parser that can succeed without consuming
// anything from a command line - the first such success counts as well.
//
// # Example
//
//	count := cli.Short('v').Switch().Count() // -vvv -> 3
func (self Parser[T]) Count() Parser[int] {
	inner := self
	return Parser[int]{
		eval: func(state *state) (int, parseError) { return count_eval(inner, state) },
		meta: metaMany{metaOptional{inner.meta}},
	}
}
