package cli

import "os"

// Parser for a named argument, created with [NamedArgargument].
type ParseArgument[T any] struct {
	named    NamedArg
	metavar  string
	adjacent bool
	p        func(string) (T, error)
}

func (self ParseArgument[T]) Parser() Parser[T] {
	return Parser[T]{self.eval, self.meta()}
}

// Add a help message to an `argument`
//
// See [NamedArghelp]
func (self *ParseArgument[T]) help(help Doc) *ParseArgument[T] {
	self.named.help = Valid(help)
	return self
}

// Restrict parsed arguments to have both flag and a value in the same word:
//
// In other words adjacent restricted `ParseArgument` would accept `--flag=value` or
// `-fbar` but not `--flag value`. Note, this is different from [adjacent](crateParseConadjacent),
// just plays a similar role.
//
// # Should allow to parse some of the more unusual things
//
// #[cfg_attr(not(doctest), doc = include_str!("docs2/adjacent_argument.md"))]
func (self *ParseArgument[T]) Adjacent() *ParseArgument[T] {
	self.adjacent = true
	return self
}

func (self ParseArgument[T]) item() Option[Item] {
	name, ok := ShortLongtry_from(self.named).Unpack()
	if !ok {
		return Invalid[Item]()
	}
	return Valid(ItemArgument(
		name,
		self.named.short,
		Metavar(self.metavar),
		First(self.named.env),
		self.named.help,
	))
}

func (self ParseArgument[T]) take_argument(args *State) (string, Error) {
	wopt, err := args.take_arg(self.named, self.adjacent, Metavar(self.metavar)).unpack()
	if err != nil {
		args.push_argument(self.named, self.metavar)
		return "", err
	} else if w, ok := wopt.Unpack(); ok {
		if args.touching_last_remove() {
			args.push_metavar(self.metavar, self.named.help, true)
		}
		return w, nil
	} else {
		args.push_argument(self.named, self.metavar)
		if val, ok := find_map(self.named.env, func(e string) Option[string] {
			r, ok := os.LookupEnv(e)
			return Option[string]{r, ok}
		}).Unpack(); ok {
			args.current = Invalid[int]()
			return val, nil
		}

		if item, ok := self.item().Unpack(); ok {
			missing := MissingItem{
				item:     item,
				position: args.scope.start(),
				scope:    args.scope,
			}
			return "", Error(MessageMissing(missing))
		} else if name, ok := First(self.named.env).Unpack(); ok {
			return "", Error(MessageNoEnv(name))
		} else {
			panic("unreachable!()")
		}
	}
}

func (self ParseArgument[T]) eval(args *State) (T, Error) {
	os, err := self.take_argument(args)
	if err != nil {
		return *new(T), err
	}

	t, errr := self.p(os)
	if errr != nil {
		return *new(T), Error(MessageParseFailed(args.current, errr.Error()))
	}
	return t, nil
}

func (self ParseArgument[T]) meta() Meta {
	if item, ok := self.item().Unpack(); ok {
		return MetaItem(item)
	} else {
		return MetaSkip
	}
}

// Argument
//
// A short (`-a`) or long (`--name`) name followed by  either a space or `=` and
// then by a string literal.  `-f foo`, `--flag bar` or `-o=-` are all valid Argument examples. Note, string
// literal can't start with `-` unless separated from the flag with `=`. For short flags value
// can follow immediately: `-fbar`.
//
// When using combinatoring API you can specify the type with turbofish, for parsing types
// that don't implement [FromStr] you can use consume a `String`/`string` first and parse
// it by hands.
//
// For `metavar` value you should pick something short and descriptive about the parameter,
// usually in capital letters. For example for an abstract file parameter it could be
// `"FILE"`, for a username - `"USER"`, etc.
//
// #[cfg_attr(not(doctest), doc = include_str!("docs2/Argument.md"))]
//
// You can further restrict it using [adjacent](ParseArgumentadjacent)
func Argument[T any](self NamedArg, metavar string, p func(string) (T, error)) ParseArgument[T] {
	return ParseArgument[T]{
		named:    self,
		metavar:  metavar,
		adjacent: false,
		p:        p,
	}
}
