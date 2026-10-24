package cli

import (
	"fmt"
	"iter"
	"slices"
	"strings"

	"github.com/rprtr258/fun"
	funiter "github.com/rprtr258/fun/iter"
)

func get[T any](xs []T, ix int) Opt[T] {
	if ix < 0 || ix >= len(xs) {
		return Invalid[T]()
	}
	return Valid(xs[ix])
}

func first[T any](xs []T) Opt[T] {
	return get(xs, 0)
}

func last[T any](xs []T) Opt[T] {
	return get(xs, len(xs)-1)
}

func strip_prefix(s, prefix string) Opt[string] {
	x, ok := strings.CutPrefix(s, prefix)
	return Opt[string]{x, ok}
}

func strip_prefixb(s string, prefix byte) Opt[string] {
	return strip_prefix(s, string(prefix))
}

// Opt is either value or nothing.
type Opt[T any] struct {
	Value T
	Valid bool
}

func (o Opt[T]) String() string {
	if o.Valid {
		return fmt.Sprintf("Some(%v)", o.Value)
	} else {
		return "None"
	}
}

func (o Opt[T]) clone() Opt[T] {
	switch t := any(o.Value).(type) {
	case interface{ clone() T }:
		return Opt[T]{t.clone(), o.Valid}
	default:
		panic(1)
	}
}

func Invalid[T any]() Opt[T] {
	return Opt[T]{}
}

func Valid[T any](t T) Opt[T] {
	return Opt[T]{
		Value: t,
		Valid: true,
	}
}

func (o Opt[T]) Unpack() (T, bool) {
	return o.Value, o.Valid
}

func (o Opt[T]) unwrap() T {
	if !o.Valid {
		panic("invalid optional value unwrap")
	}
	return o.Value
}

func (o Opt[T]) Or(other Opt[T]) Opt[T] {
	return fun.IF(o.Valid, o, other)
}

func (o Opt[T]) OrDefault(value T) T {
	return fun.IF(o.Valid, o.Value, value)
}

func (o Opt[I]) Map[O any](f func(I) O) Opt[O] {
	if !o.Valid {
		return Invalid[O]()
	}
	return Valid(f(o.Value))
}

func (o Opt[I]) MapOr[O any](defVal O, f func(I) O) O {
	return o.Map(f).OrDefault(defVal)
}

func (o Opt[I]) FlatMap[O any](f func(I) Opt[O]) Opt[O] {
	if !o.Valid {
		return Invalid[O]()
	}
	return f(o.Value)
}

func starts_withb(s string, prefix byte) bool {
	return strings.HasPrefix(s, string(prefix))
}

func find_map[T, R any](xs []T, f func(T) (R, bool)) (R, bool) {
	for _, x := range xs {
		if res, ok := f(x); ok {
			return res, true
		}
	}
	return *new(R), false
}

func rev[T any](xs []T) iter.Seq[T] {
	return iter.Seq[T](funiter.Seq2[int, T](slices.Backward(xs)).Values())
}

func filter_map[T, R any](it iter.Seq[T], f func(T) (R, bool)) iter.Seq[R] {
	return func(yield func(R) bool) {
		for x := range it {
			if y, ok := f(x); ok && !yield(y) {
				break
			}
		}
	}
}

func fromPull[T any](next func() (T, bool)) iter.Seq[T] {
	return func(yield func(T) bool) {
		for {
			x, ok := next()
			if !ok || !yield(x) {
				break
			}
		}
	}
}

func position[T any](it iter.Seq[T], f func(T) bool) (int, bool) {
	i := 0
	for x := range it {
		if f(x) {
			return i, true
		}
		i++
	}
	return 0, false
}

func max_by_key[T any](xs iter.Seq[T], cmp func(a, b T) int) (T, bool) {
	ok := false
	var res T
	for x := range xs {
		if !ok || cmp(res, x) > 0 {
			res = x
			ok = true
		}
	}
	return res, ok
}

func push2[A, B any](next func() (A, B, bool)) iter.Seq2[A, B] {
	return func(yield func(A, B) bool) {
		for {
			if a, b, ok := next(); !ok || !yield(a, b) {
				return
			}
		}
	}
}

func lenOf[T any](xs []T) int {
	return len(xs)
}
