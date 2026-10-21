package cli

import "math"

// Apply inner parser several times and collect results into `Vec`, created with
// [Many], implements catch mode via the `catch` parameter.
func many_eval[T any](inner Parser[T], catch bool, state *State) ([]T, Error) {
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

// Consume zero or more items from a command line and collect them into a [Vec]
//
// `Many` preserves any parsing failures and propagates them outwards, with an extra
// [catch](ParseManycatch) statement you can instead stop at the first value
// that failed to parse and ignore it and all the subsequent ones.
//
// `Many` will collect at most one result that does not consume anything from the argument
// list allowing using it in combination with any parsers with a fallback. After the first
// one, it will keep collecting the results as long as they consume something.
//
// For derive usage `bpaf` would insert implicit `Many` when the resulting type is a
// vector.
//
// #[cfg_attr(not(doctest), doc = include_str!("docs2/Many.md"))]
//
// # See also
// [Parser.Some] also collects results to a vector but requires at least one
// element to succeed, [Parser.Collect] collects results into a collection
func Many[T any, P parser[T]](
	self P,
	// Handle parse failures
	//
	// Can be useful to decide to skip parsing of some items on a command line
	// When parser succeeds - `catch` version would return a value as usual
	// if it fails - `catch` would restore all the consumed values and return Invalid().
	//
	// There's several structures that implement this attribute: [Optional], [Many]
	// and [Parser.Some], behavior should be identical for all of them.
	catch bool,
) Parser[[]T] {
	inner := self.Parser()
	return Parser[[]T]{
		eval: func(state *State) ([]T, Error) { return many_eval(inner, catch, state) },
		meta: MetaMany{MetaOptional{inner.meta}},
	}
}
