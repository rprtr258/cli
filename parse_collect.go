package cli

import "math"

// Apply inner parser several times and collect results into `FromIterator`, created with
// [collect](Parsercollect),
// Implements [catch](ParseCollectcatch)
type ParseCollect[T any] struct {
	inner Parser[T]
	catch bool
}

// Handle parse failures
//
// Can be useful to decide to skip parsing of some items on a command line
// When parser succeeds - `catch` version would return a value as usual
// if it fails - `catch` would restore all the consumed values and return Invalid().
//
// There's several structures that implement this attribute: [ParseOptional], [ParseMany]
// and [ParseSome], behavior should be identical for all of them.
// #[cfg_attr(not(doctest), doc = include_str!("docs2/some_catch.md"))]
func (self *ParseCollect[T]) Catch() *ParseCollect[T] {
	self.catch = true
	return self
}

// impl<T, C, P> Parser<C> for ParseCollect<P, C, T>
func (self ParseCollect[T]) eval(args *State) Result[[]T] {
	len := uint(math.MaxUint)
	res := []T{}
	for {
		xx, err := parse_option(self.inner, &len, args, self.catch)
		if err != nil {
			return Err[[]T](err)
		}
		x, ok := xx.Unpack()
		if !ok {
			break
		}
		res = append(res, x)
	}
	return Ok(res)
}

func (self ParseCollect[T]) meta() Meta {
	return MetaMany(MetaRequired(self.inner.meta))
}
