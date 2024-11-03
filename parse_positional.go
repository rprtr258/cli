package cli

type Position uint8

const (
	PositionUnrestricted Position = iota
	PositionStrict
	PositionNonStrict
)

// Parse a positional item, created with [positional](cratepositional)
//
// You can add extra information to positional parsers with [help](Selfhelp),
// [strict](Selfstrict), or [non_strict](Selfnon_strict) on this struct.
type ParsePositional[T any] struct {
	metavar  string
	help     Option[Doc]
	position Position
	p        func(string) (T, error)
}

// Add a help message to a [positional] parser
//
// `bpaf` converts doc comments and string into help by following those rules:
// 1. Everything up to the first blank line is included into a "short" help message
// 2. Everything is included into a "long" help message
// 3. `bpaf` preserves linebreaks followed by a line that starts with a space
// 4. Linebreaks are removed otherwise
//
// You can pass anything that can be converted into [Doc], if you are not using
// documentation generation functionality ([doc](cratedoc)) this can be `string`.
//
// #[cfg_attr(not(doctest), doc = include_str!("docs2/positional.md"))]
func (self *ParsePositional[T]) Help(help string) *ParsePositional[T] {
	self.help = Valid(newDocFromString(help))
	return self
}

func (self *ParsePositional[T]) Parser() Parser[T] {
	meta := Meta(MetaItem(Item{self.help, ItemPositional{Metavar(self.metavar)}}))
	if self.position == PositionStrict {
		meta = MetaStrict{meta}
	}

	return Parser[T]{self.eval, meta}
}

// Changes positional parser to be a "strict" positional
//
// Usually positional items can appear anywhere on a command line:
// ```console
// $ ls -d bpaf
// $ ls bpaf -d
// ```
// here `ls` takes a positional item `bpaf` and a flag `-d`
//
// But in some cases it might be useful to have a stricter separation between
// positonal items and flags, such as passing arguments to a subprocess:
// ```console
// $ cargo run --example basic -- --help
// ```
//
// here `cargo` takes a `--help` as a positional item and passes it to the example
//
// `bpaf` allows to require user to pass `--` for positional items with `strict` annotation.
// `bpaf` would display such positional elements differently in usage line as well.
// #[cfg_attr(not(doctest), doc = include_str!("docs2/positional_strict.md"))]
func (self *ParsePositional[T]) strict() *ParsePositional[T] {
	self.position = PositionStrict
	return self
}

// Changes positional parser to be a "not strict" positional
//
// Ensures the parser always rejects "strict" positions to the right of the separator, `--`.
// Essentially the inverse operation to [ParsePositionalstrict], which can be used to ensure
// adjacent strict and nonstrict args never conflict with eachother.
func (self *ParsePositional[T]) non_strict() *ParsePositional[T] {
	self.position = PositionNonStrict
	return self
}

func (self ParsePositional[T]) eval(args *State) (T, Error) {
	metavar := Metavar(self.metavar)
	help := self.help
	position := self.position
	ix, is_strict, word, err := args.take_positional_word(metavar)
	if err != nil {
		if !args.check_no_pos_ahead() {
			args.push_metavar(string(metavar), help, false)
			args.set_no_pos_ahead()
		}
		return *new(T), err
	}

	if position == PositionStrict && !is_strict {
		args.push_pos_sep()
		return *new(T), ErrorStrictPos{ix, metavar}
	} else if position == PositionNonStrict && is_strict {
		return *new(T), ErrorNonStrictPos{ix, metavar}
	}

	if args.touching_last_remove() && !args.check_no_pos_ahead() {
		args.push_metavar(string(metavar), help, false)
		args.set_no_pos_ahead()
	}

	t, errr := self.p(word)
	if errr != nil {
		return *new(T), ErrorParseFail(errr.Error())
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
// When using combinatoric API you can specify the type with turbofish, for parsing types
// that don't implement [FromStr] you can use consume a `String`/`string` first and parse
// it by hand.
//
//	func parse_pos() Parser[uint] {
//	    return Positional[uint]("POS")
//	}
//
// # Important restriction
// To parse Positional arguments from a command line you should place parsers for all your
// named values before parsers for Positional items and commands. In derive API fields parsed as
// Positional items or commands should be at the end of your `struct`/`enum`. The same rule applies
// to parsers with Positional fields or commands inside: such parsers should go to the end as well.
//
// Use [check_invariants](OptionParsercheck_invariants) in your test to ensure correctness.
//
// For example for non-Positional `non_pos` and Positional `pos` parsers
//
//	# non_pos := short('n').switch();
//	# pos := Positional[string]("POS");
//	valid := Pair(non_pos, pos);
//	invalid := Pair(pos, non_pos);
//
// **`bpaf` panics during help generation unless this restriction holds**
//
// Without using `--` `bpaf` would only accept items that don't start with `-` as Positional, you
// can use [any] to work around this restriction.
//
// By default `bpaf` accepts Positional items with or without `--` where values permit, you can
// further restrict the parser to accept Positional items only on the right side of `--` using
// [strict](ParsePositionalstrict).
// #[cfg_attr(not(doctest), doc = include_str!("docs2/Positional.md"))]
func Positional[T any](metavar string, parse func(string) (T, error)) *ParsePositional[T] {
	return &ParsePositional[T]{
		metavar:  metavar,
		help:     Invalid[Doc](),
		position: PositionUnrestricted,
		p:        parse,
	}
}
