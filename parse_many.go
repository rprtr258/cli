package cli

import "math"

// Repeatedly apply an inner parser and collect the results, created with
// [Parser.Many] or [Parser.ManyCatch].
func many_eval[T any](inner Parser[T], catch bool, state *state) ([]T, parseError) {
	len := uint(math.MaxUint)
	res := []T{}
	for {
		vr, err := parse_option(inner, &len, state, catch)
		if err == nil && !vr.Valid {
			break
		}
		if err != nil {
			return nil, err
		}

		res = append(res, vr.Value)
	}
	return res, nil
}

// Many collects zero or more items from a command line.
//
// Inner parser failures propagate outward and abort the collection.
// See [Parser.ManyCatch] for the variant that stops at the first failing item, keeping
// the values collected so far.
//
// Many will collect at most one result that does not consume anything from the
// argument list, allowing its use in combination with parsers with a fallback. After
// the first one, it keeps collecting the results as long as they consume something.
//
// With [Construct] a slice field parses as [Parser.Many] of the field parser.
//
// # See also
// [Parser.Some] also collects results to a slice but requires at least one
// element to succeed; [Parser.Collect] collects results into a slice
func (self Parser[T]) Many[R []T]() Parser[R] {
	inner := self.Parser()
	return Parser[R]{
		eval: func(state *state) (R, parseError) { return many_eval(inner, false, state) },
		meta: metaMany{metaOptional{inner.meta}},
	}
}

// ManyCatch collects zero or more items, stopping at the first inner parser
// failure and keeping the values collected so far.
func (self Parser[T]) ManyCatch[R []T]() Parser[R] {
	inner := self.Parser()
	return Parser[R]{
		eval: func(state *state) (R, parseError) { return many_eval(inner, true, state) },
		meta: metaMany{metaOptional{inner.meta}},
	}
}
