package internal

import "cmp"

func OneOf[T comparable](x T, ys ...T) bool {
	for _, y := range ys {
		if x == y {
			return true
		}
	}
	return false
}

func Or[T comparable](x T, ys ...T) T {
	var zero T
	res := x
	for _, y := range ys {
		if res == zero {
			res = y
		}
	}
	return res
}

func MaxBy[T any, R cmp.Ordered](f func(T) R, xs ...T) R {
	if len(xs) == 0 {
		var zero R
		return zero
	}

	ret := f(xs[0])
	for _, v := range xs[1:] {
		ret = max(ret, f(v))
	}
	return ret
}

func Filter[T any](f func(T) bool, xs ...T) []T {
	ret := make([]T, 0, len(xs))
	for _, v := range xs {
		if f(v) {
			ret = append(ret, v)
		}
	}
	return ret
}

func FilterMap[T, R any](f func(T) (R, bool), xs ...T) []R {
	ret := make([]R, 0, len(xs))
	for _, v := range xs {
		r, ok := f(v)
		if ok {
			ret = append(ret, r)
		}
	}
	return ret
}

// Iverson returns value if b is true and zero if b is false
func Iverson[T any](b bool, value T) T {
	if b {
		return value
	}

	var zero T
	return zero
}

func Join[T any](f func(T) string, sep string, xs ...T) string {
	ret := ""
	for i, v := range xs {
		if i != 0 {
			ret += sep
		}
		ret += f(v)
	}
	return ret
}
