package cli

// Parser that returns a given value without consuming anything, created with [pure](cratepure).
type parsePure[T any] struct {
	inner T
}

func (self parsePure[T]) Parser() Parser[T] {
	return Parser[T]{self.eval, MetaSkip{}}
}

func (self parsePure[T]) eval(args *State) (T, Error) {
	args.current = Invalid[int]()
	return self.inner, nil
}

// Parser that produces a fixed value
//
// This parser produces `T` without consuming anything from the command line, which can be useful
// with [construct]. As with any parsers, `T` should be `Clone` and `Debug`.
//
// Both `Pure` and [pure_with] are designed to put values into structures, to generate fallback
// you should be using [fallback](Parserfallback) and [fallback_with](Parserfallback_with).
//
// See also [pure_with] for a Pure computation that can fail.
//
// #[cfg_attr(not(doctest), doc = include_str!("docs2/Pure.md"))]
func Pure[T any](val T) parsePure[T] {
	return parsePure[T]{val}
}

type parsePureWith[T any] struct{ inner func() (T, Error) }

func (self parsePureWith[T]) Parser() Parser[T] {
	return Parser[T]{self.eval, MetaSkip{}}
}

func (self parsePureWith[T]) eval(*State) (T, Error) {
	ok, e := self.inner()
	if e == nil {
		return ok, nil
	}
	return *new(T), ErrorPureFailed(errorToString(e))
}

// Wrap a calculated value into a `Parser`
//
// This parser represents a possibly failing equivalent to [pure].
// It produces `T` by invoking the provided callback without consuming anything from the command
// line, which can be useful with [construct]. As with any parsers, `T` should be `Clone`
// and `Debug`.
//
// Both [pure] and `PureWith` are designed to put values into structures, to generate fallback
// you should be using [fallback](Parserfallback) and [fallback_with](Parserfallback_with).
//
// See also [pure] for a pure computation that can't fail.
//
// #[cfg_attr(not(doctest), doc = include_str!("docs2/PureWith.md"))]
func PureWith[T any](val func() (T, Error)) parsePureWith[T] {
	return parsePureWith[T]{val}
}
