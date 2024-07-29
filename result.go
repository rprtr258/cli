package cli

type result[T any] struct {
	Value T
	Err   error
}

func resultOk[T any](v T) result[T] {
	return result[T]{v, nil}
}

func resultErr[T any](err error) result[T] {
	return result[T]{*new(T), err}
}

func resultNew[T any](v T, err error) result[T] {
	return result[T]{v, err}
}

func (r result[T]) check(f func(T) error) result[T] {
	if r.Err != nil {
		return r
	}
	if err := f(r.Value); err != nil {
		return resultErr[T](err)
	}
	return r
}
