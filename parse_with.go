package cli

// Parser that transforms parsed value with a failing function, created with [parse](Parserparse)
type ParseWith[T, R any] struct {
	inner    Parser[T]
	parse_fn func(T) (R, Error)
}

func (self ParseWith[T, R]) Parser() Parser[R] {
	return Parser[R]{self.eval, self.inner.meta}
}

func (self ParseWith[T, R]) eval(args *State) (R, Error) {
	t, err := self.inner.eval(args)
	if err != nil {
		return *new(R), err
	}

	r, err := self.parse_fn(t)
	if err != nil {
		return *new(R), Error(MessageParseFailed(args.current, err.String()))
	}

	return r, nil
}

// Parse
// Apply a failing transformation to a contained value
//
// Transformation preserves the present/absent state of the value: to Parse an fun.Optional value you
// can either first try to `Parse` it and then mark it as [optional](Parseroptional) or first
// deal with the fun.Optionality and then Parse a value wrapped in [Option]. In most cases
// the former approach is more concise.
//
// Similarly, it is possible to Parse multiple items with [many](Parsermany) or
// [some](Parsersome) by either parsing a single item first and then turning it into a [Vec]
// or collecting them into a [Vec] first and then parsing the whole vector. The former approach
// is more concise.
//
// This is a most general of transforming parsers and you can express
// [map](Parsermap) and [guard](Parserguard) in terms of it.
//
// Examples are a bit artificial, to Parse a value from a string you can specify
// the type directly in the `argument`'s turbofish and then apply `map`.
//
// # Derive usage:
// `Parse` takes a single parameter: function name to call. Function type should match
// parameter `F` used by `Parse` in combinatoric API.
//
// #[cfg_attr(not(doctest), doc = include_str!("docs2/Parse.md"))]
func Parse[T, R any](self Parser[T], f func(T) (R, Error)) ParseWith[T, R] {
	return ParseWith[T, R]{
		inner:    self,
		parse_fn: f,
	}
}
