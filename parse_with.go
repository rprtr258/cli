package cli

// Parse
// Apply a failing transformation to a contained value
//
// Transformation preserves the present/absent state of the value: to transform an optional value you
// can either first try to `Parse` it and then mark it as [Optional] or first
// deal with the optionality first and then transform the value wrapped in [Opt]. In most cases
// the former approach is more concise.
//
// Similarly, it is possible to Parse multiple items with [Many] or
// [Some] by either parsing a single item first and then turning it into a slice
// or collecting them into a slice first and then parsing the whole vector. The former approach
// is more concise.
//
// This is a most general of transforming parsers and you can express
// [Map] and [Guard] in terms of it.
func Parse[T, R any, P parser[T]](self P, f func(T) (R, error)) Parser[R] {
	return Parser[R]{
		eval: func(state *state) (R, parseError) {
			t, err := self.Parser().eval(state)
			if err != nil {
				return *new(R), err
			}

			r, errr := f(t)
			if errr != nil {
				return *new(R), errorParseFailed{state.currentIdx, errr.Error()}
			}

			return r, nil
		},
		meta: self.Parser().meta,
	}
}
