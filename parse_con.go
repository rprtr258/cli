package cli

import (
	"github.com/rprtr258/fun"
)

// Create parser from a function, [construct](crateconstruct) uses it internally
type ParseCon[T any] struct {
	// inner parser closure
	inner func(bool, *State) (T, Error)
	// metas for inner parsers
	meta Meta
	// To produce a better error messages while parsing constructed values
	// we want to look at all the items so values that can be consumed are consumed
	// autocomplete relies on the same logic
	//
	// However when dealing with adjacent restriction detecting the first item relies on failing fast
	failfast bool
}

func (self ParseCon[T]) Parser() Parser[T] {
	return Parser[T]{self.eval, self.meta}
}

func (self ParseCon[T]) eval(state *State) (T, Error) {
	res, err := self.inner(self.failfast, state)
	state.currentIdx = Invalid[int]()
	return res, err
}

// Automagically restrict the inner parser scope to accept Adjacent values only
//
// `Adjacent` can solve surprisingly wide variety of problems: sequential command chaining,
// multi-value arguments, fun.Option-structs to name a few. If you want to run a parser on a
// sequential subset of arguments - `Adjacent` might be able to help you. Check the examples
// for better intuition.
//
// Let's consider two examples with consumed items marked in bold and constructor containing
// parsers for `-c` and `-d`.
//
// - <code>**-a** -b **-c** -d</code>
// - <code>**-a** **-c** -b -d</code>
//
// In the first example `-b` breaks the adjacency for all the consumed items so parsing will fail,
// while here in the second one all the consumed items are Adjacent to each other so
// parsing will succeed.
//
// # Multi-value arguments
//
// Parsing things like `--point X Y Z`
// #[cfg_attr(not(doctest), doc = include_str!("docs2/adjacent_struct_0.md"))]
//
// # Structure groups
//
// Parsing things like `--rect --width W --height H --rect --height H --width W`
// #[cfg_attr(not(doctest), doc = include_str!("docs2/adjacent_struct_1.md"))]
//
// # Chaining commands
// This example explains [Adjacent](crateparamsParseCommandadjacent), but the same idea holds.
// Parsing things like `cmd1 --arg1 cmd2 --arg2 --arg3 cmd3 --flag`
//
// #[cfg_attr(not(doctest), doc = include_str!("docs2/adjacent_command.md"))]
//
// # Capturing everything between markers
//
// Parsing things like `find . --exec foo {} -bar ; --more`
//
// #[cfg_attr(not(doctest), doc = include_str!("docs2/adjacent_struct_3.md"))]
//
// # Multi-value arguments with fun.Optional flags
//
// Parsing things like `--foo ARG1 --flag --inner ARG2`
//
// So you can parse things while parsing things. Not sure why you might need this, but you can
// :)
//
// #[cfg_attr(not(doctest), doc = include_str!("docs2/adjacent_struct_4.md"))]
//
// # Performance and other considerations
//
// `bpaf` can run adjacently restricted parsers multiple times to refine the guesses. It's
// best not to have complex inter-fields verification since they might trip up the detection
// logic: instead of restricting, for example "sum of two fields to be 5 or greater" *inside* the
// `Adjacent` parser, you can restrict it *outside*, once `Adjacent` done the parsing.
//
// There's also similar method [Adjacent](crateparsersParseArgument) that allows to restrict argument
// parser to work only for arguments where both key and a value are in the same shell word:
// `-f=bar` or `-fbar`, but not `-f bar`.
func (self ParseCon[T]) Adjacent() ParseAdjacent[T] {
	self.failfast = true
	return ParseAdjacent[T]{inner: Parser[T]{self.eval, self.meta}}
}

func Pair[A, B any](ppa parser[A], ppb parser[B]) *ParseCon[fun.Pair[A, B]] {
	pa := ppa.Parser()
	pb := ppb.Parser()
	meta := MetaAnd{pa.meta, pb.meta}
	type R = fun.Pair[A, B]
	inner := func(failfast bool, state *State) (R, Error) {
		a, erra := pa.eval(state)
		if failfast && erra != nil {
			return R{}, erra
		}
		b, errb := pb.eval(state)
		if erra != nil {
			return R{}, erra
		}
		if errb != nil {
			return R{}, errb
		}
		state.currentIdx = Invalid[int]()
		return R{a, b}, nil
	}
	return &ParseCon[R]{inner, meta, false}
}

type triple[A, B, C any] struct {
	A A
	B B
	C C
}

func Triple[A, B, C any](ppa parser[A], ppb parser[B], ppc parser[C]) *ParseCon[triple[A, B, C]] {
	pa, pb, pc := ppa.Parser(), ppb.Parser(), ppc.Parser()
	meta := MetaAnd{pa.meta, pb.meta, pc.meta}
	type R = triple[A, B, C]
	inner := func(failfast bool, state *State) (R, Error) {
		a, erra := pa.eval(state)
		if failfast && erra != nil {
			return R{}, erra
		}
		b, errb := pb.eval(state)
		c, errc := pc.eval(state)
		if erra != nil {
			return R{}, erra
		}
		if errb != nil {
			return R{}, errb
		}
		if errc != nil {
			return R{}, errc
		}
		state.currentIdx = Invalid[int]()
		return R{a, b, c}, nil
	}
	return &ParseCon[R]{inner, meta, false}
}
