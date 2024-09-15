package cli

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
