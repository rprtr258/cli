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
	meta := meta(metaSkip{})
	if item, ok := self.item(); ok {
		meta = metaItem(item)
	}

	return Parser[T]{self.eval, meta}
}

// ToOptions prepares the parser for running, see [Parser.ToOptions].
func (self parseArgument[T]) ToOptions() *OptionParser[T] { return self.Parser().ToOptions() }

// Map applies a pure transformation to the parsed value, see [Parser.Map].
func (self parseArgument[T]) Map[R any](fn func(T) R) Parser[R] { return self.Parser().Map(fn) }

// Add a help message to an `argument`
//
// See [NamedArghelp]
func (self *parseArgument[T]) help(help Doc) *parseArgument[T] {
	self.named.help = Valid(help)
	return self
}

// Restrict parsed arguments to have both flag and a value in the same word:
//
// In other words adjacent restricted [Argument] would accept `--flag=value` or
// `-fbar` but not `--flag value`. Note, this is different from [ParseCon.Adjacent],
// it just plays a similar role.
func (self *parseArgument[T]) Adjacent() *parseArgument[T] {
	self.adjacent = true
	return self
}

func (self parseArgument[T]) item() (item, bool) {
	name, ok := shortLongTryFrom(self.named)
	if !ok {
		return item{}, false
	}
	return item{self.named.help, itemArgument{
		name,
		self.named.short,
		Metavar(self.metavar),
		first(self.named.env),
	}}, true
}

func (self parseArgument[T]) eval(state *state) (T, parseError) {
	// take argument
	valueStr, err := func() (string, parseError) {
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
				missing := missingItem{
					item:     item,
					position: state.scope.start(),
					scope:    state.scope,
				}
				return "", errorMissing([]missingItem{missing})
			} else if name, ok := first(self.named.env).Unpack(); ok {
				return "", errorNoEnv(name)
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
		return *new(T), errorParseFailed{state.currentIdx, errr.Error()}
	}
	return t, nil
}

// Parse a value following a named flag: a short (`-a`) or long (`--name`) name
// followed by either a space or `=` and then by a value. `-f foo`, `--flag bar` or `-o=-`
// are all valid Argument examples. Note, the value can't start with `-` unless separated
// from the flag with `=`. For short flags the value can follow immediately: `-fbar`.
//
// The type parameter is the parsed value type, the `parse` function converts the raw
// string into it. For types with custom parsing logic pass a custom function.
//
// For `metavar` you should pick something short and descriptive about the parameter,
// usually in capital letters. For example for an abstract file parameter it could be
// "FILE", for a username - "USER", etc.
//
// You can further restrict it with [parseArgument.Adjacent]
func Argument[T any](self NamedArg, metavar Metavar, p func(string) (T, error)) *parseArgument[T] {
	return &parseArgument[T]{
		named:    self,
		metavar:  metavar,
		adjacent: false,
		parse:    p,
	}
}
