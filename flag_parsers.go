package cli

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/rprtr258/fun"
)

type NamedArg struct {
	short []byte
	long  []string
	env   []string
	help  fun.Option[Doc]
}

func (p FlagParser[T]) Help(text string) FlagParser[T] {
	// TODO: implement
	return p
}

func Many[T any](p FlagParser[T]) FlagParser[[]T] {
	// TODO: implement ?
	return FlagParser[[]T]{}
}

func (p FlagParser[T]) Default(v T) FlagParser[T] {
	return FlagParser[T]{func(ctx *Context) Result[T] {
		return Ok(p.run(ctx).or(v))
	}, p.complete}
}

func (p FlagParser[T]) Guard(f func(T) bool, err string) FlagParser[T] {
	return FlagParser[T]{func(ctx *Context) Result[T] {
		return p.run(ctx).check(func(t T) error {
			if !f(t) {
				return errors.New(err)
			}
			return nil
		})
	}, p.complete}
}

func (p FlagParser[T]) completion(f func(string) []Completion) FlagParser[T] {
	return FlagParser[T]{p.run, f}
}

func ExampleGuard() {
	_ = Parse(String("speed"), strconv.Atoi).
		Help("Speed in km/h").
		Guard(func(x int) bool { return x < 100 }, "speed must be less than 100")
}

func Parse[T, R any](p FlagParser[T], f func(T) (R, error)) FlagParser[R] {
	return FlagParser[R]{func(ctx *Context) Result[R] {
		t := p.run(ctx)
		if t.Err != nil {
			return Err[R](t.Err)
		}
		return resultNew(f(t.Value))
	}, nil}
}

func Map[T, R any](
	p FlagParser[T],
	f func(T) R,
) FlagParser[R] {
	return Parse(p, func(t T) (R, error) { return f(t), nil })
}

func Map2[A, B, T any](
	pa FlagParser[A],
	pb FlagParser[B],
	f func(A, B) T,
) FlagParser[T] {
	return FlagParser[T]{func(ctx *Context) Result[T] {
		a := pa.run(ctx)
		if a.Err != nil {
			return Err[T](a.Err)
		}

		b := pb.run(ctx)
		if b.Err != nil {
			return Err[T](b.Err)
		}

		return Ok(f(a.Value, b.Value))
	}, nil}
}

func Map3[A, B, C, T any](
	pa FlagParser[A],
	pb FlagParser[B],
	pc FlagParser[C],
	f func(A, B, C) T,
) FlagParser[T] {
	return FlagParser[T]{func(ctx *Context) Result[T] {
		a := pa.run(ctx)
		if a.Err != nil {
			return Err[T](a.Err)
		}

		b := pb.run(ctx)
		if b.Err != nil {
			return Err[T](b.Err)
		}

		c := pc.run(ctx)
		if c.Err != nil {
			return Err[T](c.Err)
		}

		return Ok(f(a.Value, b.Value, c.Value))
	}, nil}
}

func Alt2[A, B, T any](
	pa FlagParser[A], fa func(A) T,
	pb FlagParser[B], fb func(B) T,
) FlagParser[T] {
	return FlagParser[T]{func(ctx *Context) Result[T] {
		a := pa.run(ctx)
		if a.Err == nil {
			return Ok(fa(a.Value))
		}

		b := pb.run(ctx)
		if b.Err == nil {
			return Ok(fb(b.Value))
		}

		return Err[T](fmt.Errorf("alt: %w, %w", a.Err, b.Err))
	}, func(s string) []Completion {
		ca := pa.complete(s)
		cb := pb.complete(s)
		return append(ca, cb...)
	}}
}

func Optional[T, R any](
	p FlagParser[T],
	fyes func(T) R,
	fno func() R,
) FlagParser[R] {
	return FlagParser[R]{func(ctx *Context) Result[R] {
		t := p.run(ctx)
		if t.Err != nil {
			return Ok(fno())
		}
		return Ok(fyes(t.Value))
	}, p.complete}
}
