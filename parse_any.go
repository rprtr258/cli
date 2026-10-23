package cli

import (
	"iter"

	"github.com/rprtr258/fun"
)

// ParseAny consumes an arbitrary value that satisfies a condition, created
// with [Any], optionally in "anywhere" mode.
type ParseAny[T any] struct {
	metavar  Doc
	help     Opt[Doc]
	parse    func(string) (T, bool)
	anywhere bool
}

func (self ParseAny[T]) Parser() Parser[T] {
	return Parser[T]{self.eval, metaItem(self.item())}
}

// ToOptions prepares the parser for running, see [Parser.ToOptions].
func (self ParseAny[T]) ToOptions() *OptionParser[T] { return self.Parser().ToOptions() }

// Map applies a pure transformation to the parsed value, see [Parser.Map].
func (self ParseAny[T]) Map[R any](fn func(T) R) Parser[R] { return self.Parser().Map(fn) }

// Hide hides this parser from help generation, see [Parser.Hide].
func (self ParseAny[T]) Hide() Parser[T] { return self.Parser().Hide() }

// HideUsage hides this parser from the usage line, see [Parser.HideUsage].
func (self ParseAny[T]) HideUsage() Parser[T] { return self.Parser().HideUsage() }

// CustomUsage overrides the usage line, see [Parser.CustomUsage].
func (self ParseAny[T]) CustomUsage(usage Doc) Parser[T] { return self.Parser().CustomUsage(usage) }

// GroupHelp attaches a help message to this parser, see [Parser.GroupHelp].
func (self ParseAny[T]) GroupHelp(message string) Parser[T] { return self.Parser().GroupHelp(message) }

// Guard validates the parsed value, see [Parser.Guard].
func (self ParseAny[T]) Guard(check func(T) bool, message string) Parser[T] {
	return self.Guard(check, message)
}

// Fallback provides a default value, see [Parser.Fallback].
func (self ParseAny[T]) Fallback(value T) Parser[T] { return self.Parser().Fallback(value) }

// FallbackDisplay provides a default value shown in help, see [Parser.FallbackDisplay].
func (self ParseAny[T]) FallbackDisplay(value T) Parser[T] {
	return self.Parser().FallbackDisplay(value)
}

// FallbackWith provides a lazily computed default value, see [Parser.FallbackWith].
func (self ParseAny[T]) FallbackWith(fallback func() (T, error)) Parser[T] {
	return self.Parser().FallbackWith(fallback)
}

// FallbackWithDisplay provides a lazily computed default value shown in help, see [Parser.FallbackWithDisplay].
func (self ParseAny[T]) FallbackWithDisplay(fallback func() (T, error)) Parser[T] {
	return self.Parser().FallbackWithDisplay(fallback)
}

// Last keeps only the last matched value, see [Parser.Last].
func (self ParseAny[T]) Last() Parser[T] { return self.Parser().Last() }

// Count the number of matches, see [Parser.Count].
func (self ParseAny[T]) Count() Parser[int] { return self.Parser().Count() }

// Complete adds custom shell completion, see [Parser.Complete].
func (self ParseAny[T]) Complete(op func(T) iter.Seq2[string, Opt[string]]) ParseComp[T] {
	return self.Parser().Complete(op)
}

// CompleteShell adds file or directory shell completion, see [Parser.CompleteShell].
func (self ParseAny[T]) CompleteShell(op ShellComp) Parser[T] { return self.Parser().CompleteShell(op) }

func (self ParseAny[T]) item() item {
	return item{self.help, itemAny{self.metavar, self.anywhere}}
}

// Help adds a help message to the [Any] parser.
func (self *ParseAny[T]) Help(help string) *ParseAny[T] {
	self.help = Valid(NewDoc(help))
	return self
}

// Metavar replaces the metavar with a custom value.
//
// The metavar is the placeholder shown in the usage line and help output.
func (self *ParseAny[T]) Metavar(metavar Doc) *ParseAny[T] {
	self.metavar = metavar
	return self
}

func (self ParseAny[T]) eval(state *state) (T, parseError) {
	for ix, x := range push2(state.items_iter()) {
		if i, ok := self.parse(x.str); ok {
			state.remove(ix)
			if x.isAdjacent {
				state.remove(ix + 1)
			}

			return i, nil
		}
		if !self.anywhere {
			break
		}
	}
	return *new(T), errorMissing([]missingItem{{
		item:     self.item(),
		position: state.scope.start(),
		scope:    state.scope,
	}})
}

// Any parses a single arbitrary item from a command line.
//
// **[Any] is designed to consume items that don't fit into the usual flag, switch,
// argument, positional or command classification - in most cases you don't need to use it**
//
// By default, Any behaves similarly to a positional item: place it near the rightmost end
// of the composed parser and it will only try to parse the first unconsumed item
// on the command line. It is possible to lift this restriction by enabling the `anywhere`
// mode.
//
// The `parse` function decides whether to consume an item: when it returns true the parser
// succeeds and produces the value, otherwise it behaves as if it hasn't seen the item. In
// `anywhere` mode it will try to parse all other unconsumed items, otherwise it fails after
// the first unconsumed item.
//
// # Use Any to capture the remaining arguments
// Normally you would use [Positional] with [ParsePositional.Strict] for
// that, but Any allows you to blur the boundary between arguments for the child process and
// this process a bit more.
//
// # Use Any to parse a non standard flag
//
// # Use Any to parse a non standard argument
// Normally Any displays itself as a metavariable in the usage line and
// generated help; customize that with [ParseAny.Metavar].
//
// # See also
// [Literal] - a specialized version of Any that tries to parse a fixed literal
func Any[T any](
	metavar string,
	parse func(string) (T, bool),
	// Try to apply the parser to each unconsumed element instead of just the front one
	//
	// By default `any` tries to parse just the front unconsumed item behaving similar to
	// [positional] parser, `anywhere` changes it so it applies to every unconsumed item,
	// similar to argument parser.
	anywhere bool,
) *ParseAny[T] {
	return &ParseAny[T]{
		metavar: NewDocFromPairs([]fun.Pair[string, Style]{
			{metavar, StyleMetavar},
		}),
		help:     Invalid[Doc](),
		parse:    parse,
		anywhere: anywhere,
	}
}

// Literal is a specialized version of [Any] that consumes an arbitrary string.
//
// By default Literal behaves similarly to a positional item: place it near the
// rightmost end of the composed parser and it will only try to parse the first unconsumed
// item on the command line. It is possible to lift this restriction by enabling the
// `anywhere` mode.
//
// # See also
// [Any] - a generic version of Literal that uses a function to decide if the value is to be parsed
// or not.
func Literal(val string) Parser[struct{}] {
	p := Any("", func(s string) (struct{}, bool) {
		return struct{}{}, s == val
	}, false).Metavar(NewDocFromPairs([]fun.Pair[string, Style]{
		{val, StyleLiteral},
	}))
	return Parser[struct{}]{p.eval, metaItem(p.item())}
}
