package cli

import "iter"

type posKind uint8

const (
	PositionUnrestricted posKind = iota
	PositionStrict
	PositionNonStrict
)

// A positional item parser, created with [Positional]
//
// You can add extra information with [ParsePositional.Help],
// [ParsePositional.Strict] and [ParsePositional.NonStrict].
type ParsePositional[T any] struct {
	metavar  Metavar
	help     Opt[Doc]
	position posKind
	p        func(string) (T, error)
}

// Add a help message to the positional item
//
// The help message is displayed in the `--help` output.
func (self *ParsePositional[T]) Help(help string) *ParsePositional[T] {
	self.help = Valid(NewDoc(help))
	return self
}

func (self *ParsePositional[T]) Parser() Parser[T] {
	meta := meta(metaItem(item{self.help, itemPositional{Metavar(self.metavar)}}))
	if self.position == PositionStrict {
		meta = metaStrict{meta}
	}

	return Parser[T]{self.eval, meta}
}

// ToOptions prepares the parser for running, see [Parser.ToOptions].
func (self *ParsePositional[T]) ToOptions() *OptionParser[T] { return self.Parser().ToOptions() }

// Map applies a pure transformation to the parsed value, see [Parser.Map].
func (self *ParsePositional[T]) Map[R any](fn func(T) R) Parser[R] { return self.Parser().Map(fn) }

// Hide this parser from help generation, see [Parser.Hide].
func (self *ParsePositional[T]) Hide() Parser[T] { return self.Parser().Hide() }

// Hide this parser from the usage line, see [Parser.HideUsage].
func (self *ParsePositional[T]) HideUsage() Parser[T] { return self.Parser().HideUsage() }

// CustomUsage overrides the usage line, see [Parser.CustomUsage].
func (self *ParsePositional[T]) CustomUsage(usage Doc) Parser[T] {
	return self.Parser().CustomUsage(usage)
}

// GroupHelp attaches a help message to this parser, see [Parser.GroupHelp].
func (self *ParsePositional[T]) GroupHelp(message string) Parser[T] {
	return self.Parser().GroupHelp(message)
}

// Guard validates the parsed value, see [Parser.Guard].
func (self *ParsePositional[T]) Guard(check func(T) bool, message string) Parser[T] {
	return self.Parser().Guard(check, message)
}

// Fallback provides a default value, see [Parser.Fallback].
func (self *ParsePositional[T]) Fallback(value T) Parser[T] { return self.Parser().Fallback(value) }

// FallbackDisplay provides a default value shown in help, see [Parser.FallbackDisplay].
func (self *ParsePositional[T]) FallbackDisplay(value T) Parser[T] {
	return self.Parser().FallbackDisplay(value)
}

// FallbackWith provides a lazily computed default value, see [Parser.FallbackWith].
func (self *ParsePositional[T]) FallbackWith(fallback func() (T, error)) Parser[T] {
	return self.Parser().FallbackWith(fallback)
}

// FallbackWithDisplay provides a lazily computed default value shown in help, see [Parser.FallbackWithDisplay].
func (self *ParsePositional[T]) FallbackWithDisplay(fallback func() (T, error)) Parser[T] {
	return self.Parser().FallbackWithDisplay(fallback)
}

// Last keeps only the last matched value, see [Parser.Last].
func (self *ParsePositional[T]) Last() Parser[T] { return self.Parser().Last() }

// Count the number of matches, see [Parser.Count].
func (self *ParsePositional[T]) Count() Parser[int] { return self.Parser().Count() }

// Complete adds custom shell completion, see [Parser.Complete].
func (self *ParsePositional[T]) Complete(op func(T) iter.Seq2[string, Opt[string]]) ParseComp[T] {
	return self.Parser().Complete(op)
}

// CompleteShell adds file or directory shell completion, see [Parser.CompleteShell].
func (self *ParsePositional[T]) CompleteShell(op ShellComp) Parser[T] {
	return self.Parser().CompleteShell(op)
}

// Changes positional parser to be a "Strict" positional
//
// Usually positional items can appear anywhere on a command line:
//
//	$ ls -d file
//	$ ls file -d
//
// here `ls` takes a positional item `file` and a flag `-d`
//
// But in some cases it might be useful to have a stricter separation between
// positional items and flags, such as passing arguments to a subprocess:
//
//	$ go run --example basic -- --help
//
// here `go` takes `--help` as a positional item and passes it to the example.
//
// [ParsePositional.Strict] requires the user to pass `--` before such positional items.
// Such positional elements are displayed differently in the usage line as well.
func (self *ParsePositional[T]) Strict() *ParsePositional[T] {
	self.position = PositionStrict
	return self
}

// Changes positional parser to be a "not strict" positional
//
// Ensures the parser always rejects "strict" positions to the right of the separator, `--`.
// Essentially the inverse operation to [ParsePositional.Strict], which can be used to ensure
// adjacent strict and non-strict positionals never conflict with each other.
func (self *ParsePositional[T]) NonStrict() *ParsePositional[T] {
	self.position = PositionNonStrict
	return self
}

func (self ParsePositional[T]) eval(state *state) (T, parseError) {
	pos, isStrict, word, err := state.takePositional(self.metavar)
	if err != nil {
		if !state.check_no_pos_ahead() {
			state.push_metavar(self.metavar, self.help, false)
			state.set_no_pos_ahead()
		}
		return *new(T), err
	}

	if self.position == PositionStrict && !isStrict {
		state.push_pos_sep()
		return *new(T), errorStrictPos{pos, self.metavar}
	} else if self.position == PositionNonStrict && isStrict {
		return *new(T), errorNonStrictPos{pos, self.metavar}
	}

	if state.touching_last_remove() && !state.check_no_pos_ahead() {
		state.push_metavar(self.metavar, self.help, false)
		state.set_no_pos_ahead()
	}

	t, errr := self.p(word)
	if errr != nil {
		return *new(T), errorParseFailed{state.currentIdx, errr.Error()}
	}
	return t, nil
}

// Parse a Positional argument
//
// For named flags and arguments ordering generally doesn't matter: most programs would
// understand `-O2 -v` the same way as `-v -O2`, but for Positional items order matters: in *nix
// `cat hello world` and `cat world hello` would display contents of the same two files but in
// a different order.
//
// # Parse a positional item
//
// For named flags and arguments ordering generally doesn't matter: most programs would
// understand `-O2 -v` the same way as `-v -O2`, but for positional items order matters: in *nix
// `cat hello world` and `cat world hello` would display contents of the same two files but in
// a different order.
//
// The type parameter is the parsed value type, the `parse` function converts the raw string
// into it; ready-made converters [ParseString], [ParseInt], [ParseUint], [ParseFloat],
// [ParseBool] and [ParseDuration] cover the common types:
//
//	func parsePos() cli.Parser[uint] {
//	    return cli.Positional("POS", cli.ParseUint)
//	}
//
// # Important restriction
// To parse positional items from a command line you should place parsers for all your
// named values before parsers for positional items and commands. With [Construct] the fields
// parsed as positional items or commands should be at the end of the struct. The same rule applies
// to parsers with positional fields or commands inside: such parsers should go to the end as well.
//
// Use [OptionParser.CheckInvariants] in your tests to ensure correctness.
//
// For example for non-positional `nonPos` and positional `pos` parsers:
//
//	valid := cli.Pair(nonPos, pos)
//	invalid := cli.Pair(pos, nonPos)
//
// **The library panics during help generation unless this restriction holds**
//
// Without using `--` the library would only accept items that don't start with `-` as
// positional, you can use [Any] to work around this restriction.
//
// By default the parser accepts positional items with or without `--` where values permit, you can
// further restrict the parser to accept positional items only on the right side of `--` using
// [ParsePositional.Strict].
func Positional[T any](metavar Metavar, parse func(string) (T, error)) *ParsePositional[T] {
	return &ParsePositional[T]{
		metavar:  metavar,
		help:     Invalid[Doc](),
		position: PositionUnrestricted,
		p:        parse,
	}
}
