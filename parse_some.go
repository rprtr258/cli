package cli

import "math"

// Repeatedly apply an inner parser and collect at least one result, created with
// [Some] or [SomeCatch]; fails with `message` if none are available.
func some_eval[T any](inner Parser[T], message string, catch bool, state *state) ([]T, parseError) {
	res := []T{}
	lenn := uint(math.MaxUint)
	for {
		val, err := parse_option(inner, &lenn, state, catch)
		if err != nil {
			break
		}
		if val.Valid {
			res = append(res, val.Value)
		} else {
			break
		}
	}
	if len(res) == 0 {
		return nil, errorParseFail(message)
	}

	return res, nil
}

// Some collects one or more items from a command line; if none are present it
// fails with `message`.
//
// Inner parser failures propagate. See [SomeCatch] for the variant that stops
// at the first failing item, keeping the values collected so far.
//
// Some will collect at most one result that does not consume anything from the
// argument list, allowing its use in combination with parsers with a fallback. After
// the first one, it keeps collecting the results as long as they consume something.
//
// With [Construct] a slice field parses as [Some] of the field parser.
//
// # See also
// [Many] also collects results to a slice but succeeds with
// no matching values; [Collect] collects results into a slice
func Some[T any](self parser[T], message string) Parser[[]T] {
	inner := self.Parser()
	return Parser[[]T]{
		eval: func(state *state) ([]T, parseError) { return some_eval(inner, message, false, state) },
		meta: metaMany{metaRequired{inner.meta}},
	}
}

// SomeCatch collects one or more items; if none are present it fails with
// `message`. Stops at the first inner parser failure, keeping the values
// collected so far.
func SomeCatch[T any](self parser[T], message string) Parser[[]T] {
	inner := self.Parser()
	return Parser[[]T]{
		eval: func(state *state) ([]T, parseError) { return some_eval(inner, message, true, state) },
		meta: metaMany{metaRequired{inner.meta}},
	}
}
