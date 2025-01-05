package cli

import "os"

// Parser for a named argument, created with [NamedArgargument].
type parseArgument[T any] struct {
	named    NamedArg
	metavar  Metavar
	adjacent bool
	parse    func(string) (T, error)
}

func (self parseArgument[T]) Parser() Parser[T] {
	meta := Meta(MetaSkip{})
	if item, ok := self.item(); ok {
		meta = MetaItem(item)
	}

	return Parser[T]{self.eval, meta}
}

// Add a help message to an `argument`
//
// See [NamedArghelp]
func (self *parseArgument[T]) help(help Doc) *parseArgument[T] {
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
func (self *parseArgument[T]) Adjacent() *parseArgument[T] {
	self.adjacent = true
	return self
}

func (self parseArgument[T]) item() (Item, bool) {
	name, ok := ShortLongtry_from(self.named)
	if !ok {
		return Item{}, false
	}
	return Item{self.named.help, ItemArgument{
		name,
		self.named.short,
		Metavar(self.metavar),
		First(self.named.env),
	}}, true
}

func (self parseArgument[T]) eval(state *State) (T, Error) {
	// take argument
	valueStr, err := func() (string, Error) {
		valueStrOpt, err := state.takeFlag(self.named, self.adjacent, self.metavar)
		if err != nil {
			state.push_argument(self.named, self.metavar)
			return "", err
		}

		valueStr, ok := valueStrOpt.Unpack()
		if !ok {
			state.push_argument(self.named, self.metavar)
			if val, ok := find_map(self.named.env, os.LookupEnv); ok {
				state.currentIdx = Invalid[int]()
				return val, nil
			}

			if item, ok := self.item(); ok {
				missing := MissingItem{
					item:     item,
					position: state.scope.start(),
					scope:    state.scope,
				}
				return "", ErrorMissing([]MissingItem{missing})
			} else if name, ok := First(self.named.env).Unpack(); ok {
				return "", ErrorNoEnv(name)
			}
			panic("unreachable")
		}

		if state.touching_last_remove() {
			state.push_metavar(self.metavar, self.named.help, true)
		}
		return valueStr, nil
	}()
	if err != nil {
		return *new(T), err
	}

	t, errr := self.parse(valueStr)
	if errr != nil {
		return *new(T), ErrorParseFailed{state.currentIdx, errr.Error()}
	}
	return t, nil
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
func Argument[T any](self NamedArg, metavar Metavar, p func(string) (T, error)) *parseArgument[T] {
	return &parseArgument[T]{
		named:    self,
		metavar:  metavar,
		adjacent: false,
		parse:    p,
	}
}
