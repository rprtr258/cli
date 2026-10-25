package cli

import (
	"fmt"
	"reflect"

	"github.com/rprtr258/fun"
)

// Map applies a pure transformation to the parsed value.
func (self Parser[T]) Map[R any](fn func(T) R) Parser[R] {
	return Parser[R]{
		eval: func(state *state) (R, parseError) {
			t, e := self.eval(state)
			if e != nil {
				return *new(R), e
			}
			return fn(t), nil
		},
		meta: self.meta,
	}
}

// Map2 combines two parsers into one: both must succeed and the combined parser
// produces the value f returns for their results. See also [Map3] through [Map6]
// and [Construct] for struct composition.
func Map2[A, B, R any](
	pa parser[A],
	pb parser[B],
	f func(A, B) R,
) Parser[R] {
	return Pair(pa, pb).Map(func(x fun.Pair[A, B]) R { return f(x.K, x.V) })
}

// Map3 combines three parsers into one: all must succeed and the combined parser
// produces the value f returns for their results.
func Map3[A, B, C, R any](
	pa parser[A],
	pb parser[B],
	pc parser[C],
	f func(A, B, C) R,
) Parser[R] {
	return Pair(Pair(pa, pb), pc).Map(func(x fun.Pair[fun.Pair[A, B], C]) R { return f(x.K.K, x.K.V, x.V) })
}

// Map4 combines four parsers into one: all must succeed and the combined parser
// produces the value f returns for their results.
func Map4[A, B, C, D, R any](
	pa parser[A],
	pb parser[B],
	pc parser[C],
	pd parser[D],
	f func(A, B, C, D) R,
) Parser[R] {
	return Pair(Pair(pa, pb), Pair(pc, pd)).Map(
		func(x fun.Pair[fun.Pair[A, B], fun.Pair[C, D]]) R {
			return f(x.K.K, x.K.V, x.V.K, x.V.V)
		},
	)
}

// Map5 combines five parsers into one: all must succeed and the combined parser
// produces the value f returns for their results.
func Map5[A, B, C, D, E, R any](
	pa parser[A],
	pb parser[B],
	pc parser[C],
	pd parser[D],
	pe parser[E],
	f func(A, B, C, D, E) R,
) Parser[R] {
	return Pair(Pair(pa, pb), Pair(pc, Pair(pd, pe))).Map(
		func(x fun.Pair[fun.Pair[A, B], fun.Pair[C, fun.Pair[D, E]]]) R {
			return f(x.K.K, x.K.V, x.V.K, x.V.V.K, x.V.V.V)
		},
	)
}

// Map6 combines six parsers into one: all must succeed and the combined parser
// produces the value f returns for their results.
func Map6[A, B, C, D, E, F, R any](
	pa parser[A],
	pb parser[B],
	pc parser[C],
	pd parser[D],
	pe parser[E],
	pf parser[F],
	f func(A, B, C, D, E, F) R,
) Parser[R] {
	return Pair(Pair(Pair(pa, pb), Pair(pc, pd)), Pair(pe, pf)).Map(
		func(x fun.Pair[fun.Pair[fun.Pair[A, B], fun.Pair[C, D]], fun.Pair[E, F]]) R {
			return f(x.K.K.K, x.K.K.V, x.K.V.K, x.K.V.V, x.V.K, x.V.V)
		},
	)
}

// mapStructType validates that T is a struct with exactly len(argTypes) exported
// fields whose types exactly match the parser result types; returns T's reflect
// type for [MapStruct2] through [MapStruct6] to fill fields in declaration order.
func mapStructType[T any](name string, argTypes ...reflect.Type) reflect.Type {
	t := reflect.TypeFor[T]()
	if t.Kind() != reflect.Struct {
		panic(fmt.Sprintf("cli.%s[%s]: type parameter must be a struct, got %s", name, t, t.Kind()))
	}
	if t.NumField() != len(argTypes) {
		panic(fmt.Sprintf("cli.%s[%s]: want struct with %d fields, got %d", name, t, len(argTypes), t.NumField()))
	}
	for i, argType := range argTypes {
		field := t.Field(i)
		if !field.IsExported() {
			panic(fmt.Sprintf("cli.%s[%s]: field %q is unexported, export it to fill it with a parser value", name, t, field.Name))
		}
		if field.Type != argType {
			panic(fmt.Sprintf("cli.%s[%s]: field %q: want %s, got parser producing %s", name, t, field.Name, field.Type, argType))
		}
	}
	return t
}

// MapStruct2 combines two parsers into a struct: both must succeed and the
// combined parser produces a value with fields filled from the parser results
// in declaration order.
//
// All T fields must be exported and have exactly the types of the parser
// results. The check happens at the MapStruct2 call and panics otherwise. See
// also [MapStruct3] through [MapStruct6] and [Construct] for tag-driven
// composition.
//
// # Example
//
//	type Rect struct{ Width, Height int }
//	rect := cli.MapStruct2[Rect](width, height)
func MapStruct2[T, A, B any](pa parser[A], pb parser[B]) Parser[T] {
	t := mapStructType[T]("MapStruct2", reflect.TypeFor[A](), reflect.TypeFor[B]())
	return Map2(pa, pb, func(a A, b B) T {
		out := reflect.New(t).Elem()
		out.Field(0).Set(reflect.ValueOf(&a).Elem())
		out.Field(1).Set(reflect.ValueOf(&b).Elem())
		return out.Interface().(T)
	})
}

// MapStruct3 combines three parsers into a struct: all must succeed and the
// combined parser produces a value with fields filled from the parser results
// in declaration order.
//
// All T fields must be exported and have exactly the types of the parser
// results. The check happens at the MapStruct3 call and panics otherwise. See
// also [MapStruct4] through [MapStruct6] and [Construct] for tag-driven
// composition.
//
// # Example
//
//	type Point struct{ X, Y, Z int }
//	point := cli.MapStruct3[Point](x, y, z)
func MapStruct3[T, A, B, C any](pa parser[A], pb parser[B], pc parser[C]) Parser[T] {
	t := mapStructType[T]("MapStruct3", reflect.TypeFor[A](), reflect.TypeFor[B](), reflect.TypeFor[C]())
	return Map3(pa, pb, pc, func(a A, b B, c C) T {
		out := reflect.New(t).Elem()
		out.Field(0).Set(reflect.ValueOf(&a).Elem())
		out.Field(1).Set(reflect.ValueOf(&b).Elem())
		out.Field(2).Set(reflect.ValueOf(&c).Elem())
		return out.Interface().(T)
	})
}

// MapStruct4 combines four parsers into a struct: all must succeed and the
// combined parser produces a value with fields filled from the parser results
// in declaration order.
//
// All T fields must be exported and have exactly the types of the parser
// results. The check happens at the MapStruct4 call and panics otherwise. See
// also [MapStruct5], [MapStruct6] and [Construct] for tag-driven composition.
//
// # Example
//
//	type Rect struct{ X, Y, Width, Height int }
//	rect := cli.MapStruct4[Rect](x, y, width, height)
func MapStruct4[T, A, B, C, D any](pa parser[A], pb parser[B], pc parser[C], pd parser[D]) Parser[T] {
	t := mapStructType[T]("MapStruct4", reflect.TypeFor[A](), reflect.TypeFor[B](), reflect.TypeFor[C](), reflect.TypeFor[D]())
	return Map4(pa, pb, pc, pd, func(a A, b B, c C, d D) T {
		out := reflect.New(t).Elem()
		out.Field(0).Set(reflect.ValueOf(&a).Elem())
		out.Field(1).Set(reflect.ValueOf(&b).Elem())
		out.Field(2).Set(reflect.ValueOf(&c).Elem())
		out.Field(3).Set(reflect.ValueOf(&d).Elem())
		return out.Interface().(T)
	})
}

// MapStruct5 combines five parsers into a struct: all must succeed and the
// combined parser produces a value with fields filled from the parser results
// in declaration order.
//
// All T fields must be exported and have exactly the types of the parser
// results. The check happens at the MapStruct5 call and panics otherwise. See
// also [MapStruct6] and [Construct] for tag-driven composition.
//
// # Example
//
//	type Rect struct{ X, Y, Width, Height, Depth int }
//	rect := cli.MapStruct5[Rect](x, y, width, height, depth)
func MapStruct5[T, A, B, C, D, E any](pa parser[A], pb parser[B], pc parser[C], pd parser[D], pe parser[E]) Parser[T] {
	t := mapStructType[T]("MapStruct5", reflect.TypeFor[A](), reflect.TypeFor[B](), reflect.TypeFor[C](), reflect.TypeFor[D](), reflect.TypeFor[E]())
	return Map5(pa, pb, pc, pd, pe, func(a A, b B, c C, d D, e E) T {
		out := reflect.New(t).Elem()
		out.Field(0).Set(reflect.ValueOf(&a).Elem())
		out.Field(1).Set(reflect.ValueOf(&b).Elem())
		out.Field(2).Set(reflect.ValueOf(&c).Elem())
		out.Field(3).Set(reflect.ValueOf(&d).Elem())
		out.Field(4).Set(reflect.ValueOf(&e).Elem())
		return out.Interface().(T)
	})
}

// MapStruct6 combines six parsers into a struct: all must succeed and the
// combined parser produces a value with fields filled from the parser results
// in declaration order.
//
// All T fields must be exported and have exactly the types of the parser
// results. The check happens at the MapStruct6 call and panics otherwise. See
// also [Construct] for tag-driven composition.
//
// # Example
//
//	type Rect struct{ X, Y, Z, Width, Height, Depth int }
//	rect := cli.MapStruct6[Rect](x, y, z, width, height, depth)
func MapStruct6[T, A, B, C, D, E, F any](
	pa parser[A],
	pb parser[B],
	pc parser[C],
	pd parser[D],
	pe parser[E],
	pf parser[F],
) Parser[T] {
	t := mapStructType[T]("MapStruct6", reflect.TypeFor[A](), reflect.TypeFor[B](), reflect.TypeFor[C](), reflect.TypeFor[D](), reflect.TypeFor[E](), reflect.TypeFor[F]())
	return Map6(pa, pb, pc, pd, pe, pf, func(a A, b B, c C, d D, e E, f F) T {
		out := reflect.New(t).Elem()
		out.Field(0).Set(reflect.ValueOf(&a).Elem())
		out.Field(1).Set(reflect.ValueOf(&b).Elem())
		out.Field(2).Set(reflect.ValueOf(&c).Elem())
		out.Field(3).Set(reflect.ValueOf(&d).Elem())
		out.Field(4).Set(reflect.ValueOf(&e).Elem())
		out.Field(5).Set(reflect.ValueOf(&f).Elem())
		return out.Interface().(T)
	})
}
