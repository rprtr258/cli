package cli

import "github.com/rprtr258/fun"

// Map - Apply a pure transformation to a contained value
func (self Parser[T]) Map[R any](fn func(T) R) Parser[R] {
	return Parser[R]{
		eval: func(state *State) (R, Error) {
			t, e := self.eval(state)
			if e != nil {
				return *new(R), e
			}
			return fn(t), nil
		},
		meta: self.meta,
	}
}

// Map - Apply a pure transformation to a contained value
//
// A common case of the [Parse] function, exists mostly for convenience.
//
// Free-function form of [Parser.Map] that also accepts builders:
//
//	Map(Positional("N", parseUint32), strconv.Itoa)
//
// is the same as
//
//	Positional("N", parseUint32).Parser().Map(strconv.Itoa)
func Map[T, R any, P parser[T]](self P, fn func(T) R) Parser[R] {
	return self.Parser().Map(fn)
}

func Map2[A, B, R any](
	pa parser[A],
	pb parser[B],
	f func(A, B) R,
) Parser[R] {
	return Map(Pair(pa, pb), func(x fun.Pair[A, B]) R { return f(x.K, x.V) })
}

func Map3[A, B, C, R any](
	pa parser[A],
	pb parser[B],
	pc parser[C],
	f func(A, B, C) R,
) Parser[R] {
	return Map(Pair(Pair(pa, pb), pc), func(x fun.Pair[fun.Pair[A, B], C]) R { return f(x.K.K, x.K.V, x.V) })
}

func Map4[A, B, C, D, R any](
	pa parser[A],
	pb parser[B],
	pc parser[C],
	pd parser[D],
	f func(A, B, C, D) R,
) Parser[R] {
	return Map(
		Pair(Pair(pa, pb), Pair(pc, pd)),
		func(x fun.Pair[fun.Pair[A, B], fun.Pair[C, D]]) R {
			return f(x.K.K, x.K.V, x.V.K, x.V.V)
		},
	)
}

func Map5[A, B, C, D, E, R any](
	pa parser[A],
	pb parser[B],
	pc parser[C],
	pd parser[D],
	pe parser[E],
	f func(A, B, C, D, E) R,
) Parser[R] {
	return Map(
		Pair(Pair(pa, pb), Pair(pc, Pair(pd, pe))),
		func(x fun.Pair[fun.Pair[A, B], fun.Pair[C, fun.Pair[D, E]]]) R {
			return f(x.K.K, x.K.V, x.V.K, x.V.V.K, x.V.V.V)
		},
	)
}

func Map6[A, B, C, D, E, F, R any](
	pa parser[A],
	pb parser[B],
	pc parser[C],
	pd parser[D],
	pe parser[E],
	pf parser[F],
	f func(A, B, C, D, E, F) R,
) Parser[R] {
	return Map(
		Pair(Pair(Pair(pa, pb), Pair(pc, pd)), Pair(pe, pf)),
		func(x fun.Pair[fun.Pair[fun.Pair[A, B], fun.Pair[C, D]], fun.Pair[E, F]]) R {
			return f(x.K.K.K, x.K.K.V, x.K.V.K, x.K.V.V, x.V.K, x.V.V)
		},
	)
}
