package cli

// Parse
// Apply a failing transformation to a contained value
//
// Transformation preserves the present/absent state of the value: to Parse an fun.Optional value you
// can either first try to `Parse` it and then mark it as [Optional] or first
// deal with the fun.Optionality and then Parse a value wrapped in [Option]. In most cases
// the former approach is more concise.
//
// Similarly, it is possible to Parse multiple items with [Many] or
// [Some] by either parsing a single item first and then turning it into a slice
// or collecting them into a slice first and then parsing the whole vector. The former approach
// is more concise.
//
// This is a most general of transforming parsers and you can express
// [Parser.Map] and [Parser.Guard] in terms of it.
func (self Parser[T]) Parse[R any](f func(T) (R, error)) Parser[R] {
	return Parser[R]{
		eval: func(state *State) (R, Error) {
			t, err := self.eval(state)
			if err != nil {
				return *new(R), err
			}

			r, errr := f(t)
			if errr != nil {
				return *new(R), ErrorParseFailed{state.currentIdx, errr.Error()}
			}

			return r, nil
		},
		meta: self.meta,
	}
}

// Parse - Apply a failing transformation to a contained value
//
// Free-function form of [Parser.Parse] that also accepts builders:
//
//	Parse(Positional("N", parseUint32), strconv.ParseUint)
//
// is the same as
//
//	Positional("N", parseUint32).Parser().Parse(strconv.ParseUint)
func Parse[T, R any, P parser[T]](self P, f func(T) (R, error)) Parser[R] {
	return self.Parser().Parse(f)
}
