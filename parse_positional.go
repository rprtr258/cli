package cli

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
// into it:
//
//	func parsePos() cli.Parser[uint] {
//	    return cli.Positional[uint]("POS", parseUint)
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
