package cli

import "math"

// Apply inner parser several times and collect results, created with [Parser.Collect].
func collect_eval[T any](inner Parser[T], catch bool, state *state) ([]T, parseError) {
	len := uint(math.MaxUint)
	res := []T{}
	for {
		xx, err := parse_option(inner, &len, state, catch)
		if err != nil {
			return nil, err
		}
		x, ok := xx.Unpack()
		if !ok {
			break
		}
		res = append(res, x)
	}
	return res, nil
}

// Collect collects zero or more items from a command line into a slice
//
// A generic variant of [Many] collecting values into a slice.
// Inner parser failures propagate. See [CollectCatch] for the variant that
// stops at the first failing item, keeping the values collected so far.
//
// Collect will collect at most one result that does not consume anything from the
// argument list, allowing its use in combination with parsers with a fallback. After
// the first one, it keeps collecting the results as long as they consume something.
func Collect[T any, P parser[T]](self P) Parser[[]T] {
	inner := self.Parser()
	return Parser[[]T]{
		eval: func(state *state) ([]T, parseError) { return collect_eval(inner, false, state) },
		meta: metaMany{metaRequired{inner.meta}},
	}
}

// CollectCatch collects zero or more items, stopping at the first inner parser
// failure and keeping the values collected so far.
func CollectCatch[T any, P parser[T]](self P) Parser[[]T] {
	inner := self.Parser()
	return Parser[[]T]{
		eval: func(state *state) ([]T, parseError) { return collect_eval(inner, true, state) },
		meta: metaMany{metaRequired{inner.meta}},
	}
}
