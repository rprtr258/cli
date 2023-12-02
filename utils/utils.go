package utils

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

func Map[T, R any](f func(T) R, xs ...T) []R {
	ret := make([]R, len(xs))
	for i, v := range xs {
		ret[i] = f(v)
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
