package cli

import "math"

// Apply inner parser as many times as it succeeds while consuming something and return this number
type ParseLast[T any] struct {
	inner Parser[T]
}

func (self ParseLast[T]) Parser() Parser[T] {
	return Parser[T]{self.eval, self.meta()}
}

func (self ParseLast[T]) eval(state *State) (T, Error) {
	last := Invalid[T]()
	current := state.len()
	len := uint(math.MaxUint)
	for {
		val, err := parse_option(self.inner, &len, state, false)
		if err != nil || !val.Valid {
			break
		}
		last = val
		if current == state.len() {
			break
		}
		current = state.len()
	}
	if last, ok := last.Unpack(); ok {
		return last, nil
	} else {
		return self.inner.eval(state)
	}
}

func (self ParseLast[T]) meta() Meta {
	return MetaMany{MetaRequired{self.inner.meta}}
}

// Apply the inner parser as many times as it succeeds, return the Last value
//
// You can use this to allow users to pick contradicting fun.Options
// #[cfg_attr(not(doctest), doc = include_str!("docs2/Last.md"))]
func (self Parser[T]) Last() ParseLast[T] {
	return ParseLast[T]{inner: self}
}
