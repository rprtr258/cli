package cli

import (
	"fmt"

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

func first[T any](xs []T) Option[T] {
	if len(xs) == 0 {
		return Invalid[T]()
	}
	return Valid(xs[0])
}

type IntoString = fmt.Stringer

func stdmemswap[T any](a, b *T) {
	*a, *b = *b, *a
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

func OptMap[I, O any](o Option[I], f func(I) O) Option[O] {
	if !o.Valid {
		return Invalid[O]()
	}
	return Valid(f(o.Value))
}

func OptFlatMap[I, O any](o Option[I], f func(I) Option[O]) Option[O] {
	if !o.Valid {
		return Invalid[O]()
	}
	return f(o.Value)
}
