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
// [strict](Selfstrict), or [`non_strict`](Selfnon_strict) on this struct.
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
	return Parser[T]{self.eval, self.meta()}
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
	parse_pos_word := func(
		args *State,
		metavar Metavar,
		help Option[Doc],
		position Position,
	) (string, Error) {
		ix, is_strict, word, err := args.take_positional_word(metavar)
		if err != nil {
			if !args.check_no_pos_ahead() {
				args.push_metavar(string(metavar), help, false)
				args.set_no_pos_ahead()
			}
			return "", err
		}

		if position == PositionStrict && !is_strict {
			args.push_pos_sep()
			err = Error(MessageStrictPos(ix, metavar))
		} else if position == PositionNonStrict && is_strict {
			err = Error(MessageNonStrictPos(ix, metavar))
		}
		if err != nil {
			return "", err
		}

		if args.touching_last_remove() && !args.check_no_pos_ahead() {
			args.push_metavar(string(metavar), help, false)
			args.set_no_pos_ahead()
		}
		return word, nil
	}

	os, err := parse_pos_word(args, Metavar(self.metavar), self.help, self.position)
	if err != nil {
		return *new(T), err
	}
	t, errr := self.p(os)
	if errr != nil {
		return *new(T), Error(MessageParseFail(errr.Error()))
	}
	return t, nil
}

func (self ParsePositional[T]) meta() Meta {
	meta := MetaItem(ItemPositional(
		Metavar(self.metavar),
		self.help,
	))
	if self.position == PositionStrict {
		meta = MetaStrict(meta)
	}
	return meta
}
