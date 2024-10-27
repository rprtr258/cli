package cli

import (
	"fmt"
	"iter"
	"strings"

	"github.com/rprtr258/fun"
)

type Result[T any] struct {
	Value T
	Err   Error
}

func Ok[T any](v T) Result[T] {
	return Result[T]{v, nil}
}

func Err[T any](err Error) Result[T] {
	return Result[T]{*new(T), err}
}

func NewResult[T any](v T, err Error) Result[T] {
	return Result[T]{v, err}
}

func (r Result[T]) unpack() (T, Error) {
	return r.Value, r.Err
}

func (r Result[T]) is_ok() bool {
	return r.Err == nil
}

func (r Result[T]) unwrap() T {
	if r.Err != nil {
		panic(r.Err)
	}
	return r.Value
}

func get[T any](xs []T, ix int) Option[T] {
	if ix < 0 || ix >= len(xs) {
		return Invalid[T]()
	}
	return Valid(xs[ix])
}

func First[T any](xs []T) Option[T] {
	return get(xs, 0)
}

func last[T any](xs []T) Option[T] {
	return get(xs, len(xs)-1)
}

func stdmemswap[T any](a, b *T) {
	*a, *b = *b, *a
}

func strip_prefix(s string, prefix string) Option[string] {
	x, ok := strings.CutPrefix(s, prefix)
	return Option[string]{x, ok}
}

func strip_prefixb(s string, prefix byte) Option[string] {
	return strip_prefix(s, string(prefix))
}

// Option is either value or nothing.
type Option[T any] struct {
	Value T
	Valid bool
}

func (o Option[T]) String() string {
	if o.Valid {
		return fmt.Sprintf("Some(%v)", o.Value)
	} else {
		return "None"
	}
}

func Invalid[T any]() Option[T] {
	return Option[T]{}
}

func Valid[T any](t T) Option[T] {
	return Option[T]{
		Value: t,
		Valid: true,
	}
}

func (o Option[T]) Unpack() (T, bool) {
	return o.Value, o.Valid
}

func (o Option[T]) unwrap() T {
	if !o.Valid {
		panic("invalid optional value unwrap")
	}
	return o.Value
}

func (o Option[T]) Or(other Option[T]) Option[T] {
	return fun.IF(o.Valid, o, other)
}

func (o Option[T]) OrDefault(value T) T {
	return fun.IF(o.Valid, o.Value, value)
}

func (o Option[T]) and_then(f func(T) Option[T]) Option[T] {
	if !o.Valid {
		return Invalid[T]()
	}
	return f(o.Value)
}

func OptMap[I, O any](o Option[I], f func(I) O) Option[O] {
	if !o.Valid {
		return Invalid[O]()
	}
	return Valid(f(o.Value))
}

func map_or[I, O any](o Option[I], defVal O, f func(I) O) O {
	return OptMap(o, f).OrDefault(defVal)
}

func OptFlatMap[I, O any](o Option[I], f func(I) Option[O]) Option[O] {
	if !o.Valid {
		return Invalid[O]()
	}
	return f(o.Value)
}

func starts_withb(s string, prefix byte) bool {
	return strings.HasPrefix(s, string(prefix))
}

func find_map[T, R any](xs []T, f func(T) Option[R]) Option[R] {
	for _, x := range xs {
		if res := f(x); res.Valid {
			return res
		}
	}
	return Invalid[R]()
}

func is_empty[T any](xs []T) bool {
	return len(xs) == 0
}

func rev[T any](xs []T) iter.Seq[T] {
	return func(yield func(T) bool) {
		for i := len(xs) - 1; i >= 0; i-- {
			if !yield(xs[i]) {
				break
			}
		}
	}
}

func filter[T any](it iter.Seq[T], f func(T) bool) iter.Seq[T] {
	return func(yield func(T) bool) {
		for x := range it {
			if f(x) && !yield(x) {
				break
			}
		}
	}
}

func filter_map[T, R any](it iter.Seq[T], f func(T) Option[R]) iter.Seq[R] {
	return func(yield func(R) bool) {
		for x := range it {
			if y, ok := f(x).Unpack(); ok && !yield(y) {
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

func position[T any](it iter.Seq[T], f func(T) bool) Option[int] {
	i := 0
	for x := range it {
		if f(x) {
			return Valid(i)
		}
		i++
	}
	return Invalid[int]()
}

func max_by_key[T any](xs iter.Seq[T], cmp func(a, b T) int) Option[T] {
	ok := false
	var res T
	for x := range xs {
		if !ok || cmp(res, x) > 0 {
			res = x
			ok = true
		}
	}
	return Option[T]{res, ok}
}
