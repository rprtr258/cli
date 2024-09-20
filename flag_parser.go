package cli

import "fmt"

type FlagParser[T any] struct {
	run      func(*Context) Result[T]
	complete func(string) []Completion
}

func Bool(name string) FlagParser[bool] {
	return FlagParser[bool]{func(ctx *Context) Result[bool] {
		_, ok := ctx.Bools[name]
		return Ok(ok)
	}, nil}
}

func String(name string) FlagParser[string] {
	return FlagParser[string]{func(ctx *Context) Result[string] {
		val, ok := ctx.Flags[name]
		if !ok {
			return Err[string](fmt.Errorf("%q was not found", name))
		}
		return Ok(val)
	}, nil}
}

func Pure[T any](t T) FlagParser[T] {
	return FlagParser[T]{func(*Context) Result[T] {
		return Ok(t)
	}, nil}
}

func Fail[T any](err Error) FlagParser[T] {
	return FlagParser[T]{func(*Context) Result[T] {
		return Err[T](err)
	}, nil}
}
