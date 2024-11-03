package cli

import "math"

// Apply inner parser as many times as it succeeds while consuming something and return this number
type ParseCount[T any] struct {
	inner Parser[T]
}

func (self ParseCount[T]) Parser() Parser[int] {
	return Parser[int]{self.eval, MetaMany{MetaOptional{self.inner.meta}}}
}

func (self ParseCount[T]) eval(args *State) (int, Error) {
	res := 0
	current := args.len()
	len := uint(math.MaxUint)
	for {
		vv, err := parse_option(self.inner, &len, args, false)
		if err != nil {
			return 0, err
		}
		if !vv.Valid {
			break
		}
		res += 1
		if current == args.len() {
			break
		}
		current = args.len()
	}
	return res, nil
}

// Count how many times the inner parser succeeds, and return that number.
//
// When you are dealing with a parser that can succeed without consuming
// anything from a command line - `bpaf` will Count first such success as well.
//
// #[cfg_attr(not(doctest), doc = include_str!("docs2/Count.md"))]
func (self Parser[T]) Count() ParseCount[T] {
	return ParseCount[T]{
		inner: self,
	}
}
