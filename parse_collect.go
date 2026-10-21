package cli

import "math"

// Apply inner parser several times and collect results, created with [Parser.Collect].
func collect_eval[T any](inner Parser[T], catch bool, state *State) ([]T, Error) {
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

// Transform parser into a collection parser
//
// A generic variant of [many](Parsermany), instead of collecting into a vector
// it collects into any collection that implements [FromIterator] trait
//
// `collect` preserves any parsing failures and propagates them outwards, with extra
// [catch](ParseCollectcatch) statement you can instead stop at the first value
// that failed to parse and ignore it and all the subsequent ones.
//
// #[cfg_attr(not(doctest), doc = include_str!("docs2/collect.md"))]
//
// `collect` will collect at most one result that does not consume anything from the argument
// list allowing using it in combination of any parsers with a fallback. After the first one
// it will keep collecting the results as long as they consume something.
func Collect[T any, P parser[T]](
	self P,
	// Handle parse failures
	//
	// Can be useful to decide to skip parsing of some items on a command line
	// When parser succeeds - `catch` version would return a value as usual
	// if it fails - `catch` would restore all the consumed values and return Invalid().
	//
	// There's several structures that implement this attribute: [Optional], [Many]
	// and [Some], behavior should be identical for all of them.
	catch bool,
) Parser[[]T] {
	// free function, not a Parser method: Go forbids methods on Parser[T]
	// that instantiate Parser[[]T] (generic instantiation cycle)
	inner := self.Parser()
	return Parser[[]T]{
		eval: func(state *State) ([]T, Error) { return collect_eval(inner, catch, state) },
		meta: MetaMany{MetaRequired{inner.meta}},
	}
}
