package cli

import "math"

// Apply inner parser as many times as it succeeds while consuming something and return this number
type ParseLast[T any] struct {
	inner Parser[T]
}

func (self ParseLast[T]) eval(args *State) (T, Error) {
	last := Invalid[T]()
	current := args.len()
	len := uint(math.MaxUint)
	for {
		val, err := parse_option(self.inner, &len, args, false)
		if err != nil || !val.Valid {
			break
		}
		last = val
		if current == args.len() {
			break
		}
		current = args.len()
	}
	if last, ok := last.Unpack(); ok {
		return last, nil
	} else {
		return self.inner.eval(args)
	}
}

func (self ParseLast[T]) meta() Meta {
	return MetaMany(MetaRequired(self.inner.meta))
}

// Apply the inner parser as many times as it succeeds, return the last value
//
// You can use this to allow users to pick contradicting fun.Options
// #[cfg_attr(not(doctest), doc = include_str!("docs2/last.md"))]
func (self Parser[T]) last() ParseLast[T] {
	return ParseLast[T]{inner: self}
}
