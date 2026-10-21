package cli

import "math"

// Apply inner parser, return a value in `Some` if items requested by it are all present, restore
// and return `None` if any are missing. Created with [Optional].
func optional_eval[T any](inner Parser[T], catch bool, state *State) (Option[T], Error) {
	len := uint(math.MaxUint)
	return parse_option(inner, &len, state, catch)
}

// Turn a required argument into an fun.Optional one
//
// `Optional` converts any missing items into `Invalid()` and passes the remaining parsing
// failures untouched. With an extra [catch](ParseOptionalcatch) statement, you can handle
// those failures too.
//
// # Derive usage
//
// By default, `bpaf` would automatically use fun.Optional for fields of type `Option[T]`,
// for as long as it's not prevented from doing so by present postprocessing fun.Options.
// But it's also possible to specify it explicitly.
//
// #[cfg_attr(not(doctest), doc = include_str!("docs2/Optional.md"))]
// TODO: compiler ne mogyet v return Parser[Option[T]] in method
func Optional[T any, P parser[T]](
	self P,
	// Handle parse failures for fun.Optional parsers
	//
	// Can be useful to decide to skip parsing of some items on a command line.
	// When parser succeeds - `catch` version would return a value as usual
	// if it fails - `catch` would restore all the consumed values and return Invalid().
	//
	// There's several structures that implement this attribute: [Optional], [Many]
	// and [Parser.Some], behavior should be identical for all of them.
	catch bool,
) Parser[Option[T]] {
	inner := self.Parser()
	return Parser[Option[T]]{
		eval: func(state *State) (Option[T], Error) { return optional_eval(inner, catch, state) },
		meta: MetaOptional{inner.meta},
	}
}
