package cli

import (
	"iter"

	"github.com/rprtr258/fun"
)

// A composed parser over several inner parsers, created with [Pair], [Triple] or [Construct]
type ParseCon[T any] struct {
	// inner parser closure
	inner func(bool, *state) (T, parseError)
	// metas for inner parsers
	meta meta
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

// ToOptions prepares the parser for running, see [Parser.ToOptions].
func (self ParseCon[T]) ToOptions() *OptionParser[T] { return self.Parser().ToOptions() }

// Map applies a pure transformation to the composed value, see [Parser.Map].
func (self ParseCon[T]) Map[R any](fn func(T) R) Parser[R] { return self.Parser().Map(fn) }

// Hide this parser from help generation, see [Parser.Hide].
func (self ParseCon[T]) Hide() Parser[T] { return self.Parser().Hide() }

// Hide this parser from the usage line, see [Parser.HideUsage].
func (self ParseCon[T]) HideUsage() Parser[T] { return self.Parser().HideUsage() }

// CustomUsage overrides the usage line, see [Parser.CustomUsage].
func (self ParseCon[T]) CustomUsage(usage Doc) Parser[T] { return self.Parser().CustomUsage(usage) }

// GroupHelp attaches a help message to this parser, see [Parser.GroupHelp].
func (self ParseCon[T]) GroupHelp(message string) Parser[T] { return self.Parser().GroupHelp(message) }

// Guard validates the parsed value, see [Parser.Guard].
func (self ParseCon[T]) Guard(check func(T) bool, message string) Parser[T] {
	return self.Parser().Guard(check, message)
}

// Fallback provides a default value, see [Parser.Fallback].
func (self ParseCon[T]) Fallback(value T) Parser[T] { return self.Parser().Fallback(value) }

// FallbackDisplay provides a default value shown in help, see [Parser.FallbackDisplay].
func (self ParseCon[T]) FallbackDisplay(value T) Parser[T] {
	return self.Parser().FallbackDisplay(value)
}

// FallbackWith provides a lazily computed default value, see [Parser.FallbackWith].
func (self ParseCon[T]) FallbackWith(fallback func() (T, error)) Parser[T] {
	return self.Parser().FallbackWith(fallback)
}

// FallbackWithDisplay provides a lazily computed default value shown in help, see [Parser.FallbackWithDisplay].
func (self ParseCon[T]) FallbackWithDisplay(fallback func() (T, error)) Parser[T] {
	return self.Parser().FallbackWithDisplay(fallback)
}

// Last keeps only the last matched value, see [Parser.Last].
func (self ParseCon[T]) Last() Parser[T] { return self.Parser().Last() }

// Count the number of matches, see [Parser.Count].
func (self ParseCon[T]) Count() Parser[int] { return self.Parser().Count() }

// Complete adds custom shell completion, see [Parser.Complete].
func (self ParseCon[T]) Complete(op func(T) iter.Seq2[string, Opt[string]]) ParseComp[T] {
	return self.Parser().Complete(op)
}

// CompleteShell adds file or directory shell completion, see [Parser.CompleteShell].
func (self ParseCon[T]) CompleteShell(op ShellComp) Parser[T] { return self.Parser().CompleteShell(op) }

func (self ParseCon[T]) eval(state *state) (T, parseError) {
	res, err := self.inner(self.failfast, state)
	state.currentIdx = Invalid[int]()
	return res, err
}

// Restrict the inner parser scope to accept adjacent values only
//
// Adjacent can solve a surprisingly wide variety of problems: sequential command chaining,
// multi-value arguments, groups of options to name a few. If you want to run a parser on a
// sequential subset of arguments - Adjacent might be able to help you. Check the examples
// for a better intuition.
//
// Consider two examples with consumed items marked in bold and a group containing
// parsers for `-c` and `-d`:
//
//	-a -b -c -d   // adjacency holds: -a, -c, -d form one block
//	-a -c -b -d   // adjacency holds: -a, -c, -d form one block
//
// # Multi-value arguments
//
// Parsing things like `--point X Y Z` - a group of a flag followed by several arguments
// consumes them only when they are adjacent on the command line.
//
// # Option groups
//
// Parsing things like `--rect --width W --height H --rect --height H --width W` - each group
// starts with a flag and consumes the following options.
//
// # Chaining commands
//
// Parsing things like `cmd1 --arg1 cmd2 --arg2 --arg3 cmd3 --flag` - see [ParseCommand.Adjacent].
//
// # Capturing everything between markers
//
// Parsing things like `find . --exec foo {} -bar ; --more`.
//
// # Multi-value arguments with optional flags
//
// Parsing things like `--foo ARG1 --flag --inner ARG2`.
//
// # Performance and other considerations
//
// The library can run adjacently restricted parsers multiple times to refine the guesses. It's
// best not to have complex inter-field verification since it might trip up the detection
// logic: instead of restricting, for example, "sum of two fields to be 5 or greater" *inside* the
// Adjacent parser, restrict it *outside*, once Adjacent finished the parsing.
//
// There's also a similar restriction on [parseArgument.Adjacent] that allows to restrict an
// argument parser to work only when both key and value are in the same shell word:
// `-f=bar` or `-fbar`, but not `-f bar`.
func (self ParseCon[T]) Adjacent() Parser[T] {
	self.failfast = true
	inner := Parser[T]{self.eval, self.meta}
	return Parser[T]{
		eval: func(state *state) (T, parseError) { return adjacent_eval(inner, state) },
		meta: metaAdjacent{inner.meta},
	}
}

// Combine two parsers into one producing a [fun.Pair] of their results
//
// Both parsers must succeed for the combined parser to succeed. Usually [Map2]
// is more convenient since it produces the combined value directly. See also
// [Triple] for three parsers and [Construct] for struct composition.
func Pair[A, B any](ppa parser[A], ppb parser[B]) *ParseCon[fun.Pair[A, B]] {
	pa := ppa.Parser()
	pb := ppb.Parser()
	meta := metaAnd{pa.meta, pb.meta}
	type R = fun.Pair[A, B]
	inner := func(failfast bool, state *state) (R, parseError) {
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

// Tuple is the result of [Triple]: three combined parser values.
type Tuple[A, B, C any] struct {
	A A
	B B
	C C
}

// Combine three parsers into one producing a [Tuple] of their results
//
// All three parsers must succeed for the combined parser to succeed. Usually
// [Map3] is more convenient since it produces the combined value directly.
func Triple[A, B, C any](ppa parser[A], ppb parser[B], ppc parser[C]) *ParseCon[Tuple[A, B, C]] {
	pa, pb, pc := ppa.Parser(), ppb.Parser(), ppc.Parser()
	meta := metaAnd{pa.meta, pb.meta, pc.meta}
	type R = Tuple[A, B, C]
	inner := func(failfast bool, state *state) (R, parseError) {
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
