package cli

import "math"

// Apply inner parser several times and collect results into `Vec`, created with
// [Parser.Some], requires for at least one item to be available to succeed.
func some_eval[T any](inner Parser[T], message string, catch bool, state *State) ([]T, Error) {
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
		return nil, ErrorParseFail(message)
	}

	return res, nil
}

// Consume one or more items from a command line and collect them into a [Vec]
//
// # Takes a string used as an error message if there are no specified parameters
//
// `some` preserves any parsing failures and propagates them outwards, with an extra
// [catch](ParseSomecatch) statement you can instead stop at the first value
// that failed to parse and ignore it and all the subsequent ones.
//
// `some` will collect at most one result that does not consume anything from the argument
// list allowing using it in combination with any parsers with a fallback. After the first
// one, it will keep collecting the results as long as they consume something.
//
// #[cfg_attr(not(doctest), doc = include_str!("docs2/some.md"))]
//
// # See also
// [many](Parsermany) also collects results to a vector but succeeds with
// no matching values. [collect](Parsercollect) collects results into a [FromIterator]
// structure
func Some[T any, P parser[T]](
	self P,
	message string,
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
		eval: func(state *State) ([]T, Error) { return some_eval(inner, message, catch, state) },
		meta: MetaMany{MetaRequired{inner.meta}},
	}
}
