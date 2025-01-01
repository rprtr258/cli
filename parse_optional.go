package cli

import "math"

// Apply inner parser, return a value in `Some` if items requested by it are all present, restore
// and return `None` if any are missing. Created with [optional](Parser::optional). Implements [catch](ParseOptional::catch)
type ParseOptional[T any] struct {
	inner Parser[T]
	catch bool
}

func (self ParseOptional[T]) eval(args *State) (Option[T], Error) {
	len := uint(math.MaxUint)
	return parse_option(self.inner, &len, args, self.catch)
}

func (self *ParseOptional[T]) Parser() Parser[Option[T]] {
	return Parser[Option[T]]{self.eval, MetaOptional{self.inner.meta}}
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
func Optional[T any](
	self Parser[T],
	// Handle parse failures for fun.Optional parsers
	//
	// Can be useful to decide to skip parsing of some items on a command line.
	// When parser succeeds - `catch` version would return a value as usual
	// if it fails - `catch` would restore all the consumed values and return Invalid().
	//
	// There's several structures that implement this attribute: [ParseOptional], [ParseMany]
	// and [ParseSome], behavior should be identical for all of them.
	//
	// Those examples are very artificial and designed to show what difference `catch` makes, to
	// actually parse arguments like in examples you should [parse](Parserparse) or construct
	// enum with alternative branches
	// #[cfg_attr(not(doctest), doc = include_str!("docs2/optional_catch.md"))]
	catch bool,
) *ParseOptional[T] {
	return &ParseOptional[T]{
		inner: self,
		catch: catch,
	}
}
