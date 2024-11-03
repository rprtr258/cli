package cli

import "github.com/rprtr258/fun"

// Parser that transforms parsed value with a function, created with [map](Parsermap).
type ParseMap[T, R any] struct {
	inner Parser[T]
	fn    func(T) R
}

func (self ParseMap[T, R]) eval(args *State) (R, Error) {
	t, e := self.inner.eval(args)
	if e != nil {
		return *new(R), e
	}
	return self.fn(t), nil
}

// Map - Apply a pure transformation to a contained value
//
// A common case of the [parse](Parserparse) method, exists mostly for convenience.
//
// # Derive usage:
// The `map` takes a single parameter: function name to call. This function should transform
// the value produced by the parser into a new value of the same or different type.
//
// #[cfg_attr(not(doctest), doc = include_str!("docs2/map.md"))]
func Map[T, R any](self Parser[T], fn func(T) R) Parser[R] {
	p := ParseMap[T, R]{
		inner: self,
		fn:    fn,
	}
	return Parser[R]{p.eval, self.meta}
}

func Map2[A, B, R any](
	pa Parser[A],
	pb Parser[B],
	f func(A, B) R,
) Parser[R] {
	return Map(Pair(pa, pb).Parser(), func(x fun.Pair[A, B]) R { return f(x.K, x.V) })
}

func Map3[A, B, C, R any](
	pa Parser[A],
	pb Parser[B],
	pc Parser[C],
	f func(A, B, C) R,
) Parser[R] {
	return Map(Pair(Pair(pa, pb).Parser(), pc).Parser(), func(x fun.Pair[fun.Pair[A, B], C]) R { return f(x.K.K, x.K.V, x.V) })
}

func Map4[A, B, C, D, R any](
	pa Parser[A],
	pb Parser[B],
	pc Parser[C],
	pd Parser[D],
	f func(A, B, C, D) R,
) Parser[R] {
	return Map(Pair(
		Pair(
			pa,
			pb,
		).Parser(),
		Pair(
			pc,
			pd,
		).Parser(),
	).Parser(), func(x fun.Pair[fun.Pair[A, B], fun.Pair[C, D]]) R {
		return f(x.K.K, x.K.V, x.V.K, x.V.V)
	})
}

// TODO: Map5

func Map6[A, B, C, D, E, F, R any](
	pa Parser[A],
	pb Parser[B],
	pc Parser[C],
	pd Parser[D],
	pe Parser[E],
	pf Parser[F],
	f func(A, B, C, D, E, F) R,
) Parser[R] {
	return Map(Pair(
		Pair(
			Pair(
				pa,
				pb,
			).Parser(),
			Pair(
				pc,
				pd,
			).Parser(),
		).Parser(),
		Pair(
			pe,
			pf,
		).Parser(),
	).Parser(), func(x fun.Pair[fun.Pair[fun.Pair[A, B], fun.Pair[C, D]], fun.Pair[E, F]]) R {
		return f(x.K.K.K, x.K.K.V, x.K.V.K, x.K.V.V, x.V.K, x.V.V)
	})
}
