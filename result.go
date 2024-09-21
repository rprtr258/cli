package cli

import (
	"encoding"
	"fmt"
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

func resultNew[T any](v T, err Error) Result[T] {
	return Result[T]{v, err}
}

func (r Result[T]) unpack() (T, Error) {
	return r.Value, r.Err
}

func (r Result[T]) unwrap() T {
	if r.Err != nil {
		panic(r.Err)
	}
	return r.Value
}

func (r Result[T]) check(f func(T) Error) Result[T] {
	if r.Err != nil {
		return r
	}
	if err := f(r.Value); err != nil {
		return Err[T](err)
	}
	return r
}

func (r Result[T]) or(t T) T {
	if r.Err != nil {
		return t
	}
	return r.Value
}

func get[T any](xs []T, ix int) Option[T] {
	if ix < 0 || ix >= len(xs) {
		return Invalid[T]()
	}
	return Valid(xs[ix])
}

func first[T any](xs []T) Option[T] {
	return get(xs, 0)
}

type IntoString = fmt.Stringer
type FromStr = encoding.TextUnmarshaler

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

func stris_empty(s string) bool {
	return s == ""
}

func starts_with(s string, prefix string) bool {
	return strings.HasPrefix(s, prefix)
}

func starts_withb(s string, prefix byte) bool {
	return starts_with(s, string(prefix))
}
