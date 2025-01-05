package cli

import "os"

// Parser for a named switch, created with [NamedArgflag] or [NamedArgswitch]
type ParseFlag[T any] struct {
	present T
	absent  Option[T]
	named   NamedArg
}

func (self ParseFlag[T]) Parser() Parser[T] {
	return Parser[T]{self.eval, self.meta()}
}

func (self ParseFlag[T]) eval(state *State) (T, Error) {
	if state.takeBoolFlag(self.named) || func() bool {
		_, ok := find_map(self.named.env, os.LookupEnv)
		return ok
	}() {
		if state.touching_last_remove() {
			state.push_flag(self.named)
		}
		return self.present, nil
	} else {
		state.push_flag(self.named)
		if ok, okk := self.absent.Unpack(); okk {
			return ok, nil
		}

		if item, ok := self.named.flag_item(); ok {
			missing := MissingItem{
				item:     item,
				position: state.scope.start(),
				scope:    state.scope,
			}
			return *new(T), ErrorMissing([]MissingItem{missing})
		} else if len(self.named.env) > 0 {
			return *new(T), ErrorNoEnv(self.named.env[0])
		} else {
			panic(`todo!("no key!")`)
		}
	}
}

func (self ParseFlag[T]) meta() Meta {
	if item, ok := self.named.flag_item(); ok {
		return item.required(!self.absent.Valid)
	} else {
		return MetaSkip{}
	}
}

// Add a Help message to `flag`
//
// See [NamedArghelp]
func (self ParseFlag[T]) Help(help Doc) ParseFlag[T] {
	self.named.help = Valid(help)
	return self
}

// Flag with custom present/absent values
//
// More generic version of [switch](NamedArgswitch) that can use arbitrary type instead of
// [bool].
// #[cfg_attr(not(doctest), doc = include_str!("docs2/Flag.md"))]
func Flag[T any](self NamedArg, present, absent T) ParseFlag[T] {
	return ParseFlag[T]{present, Valid(absent), self}
}

// Required flag with custom value
//
// Similar to [flag](NamedArgflag) takes no Option arguments, but would only
// succeed if user specifies its name on a command line.
// Works best in combination with other parsers.
//
// In derive style API `bpaf` would transform field-less enum variants into a parser
// that accepts one of it's variant names as `ReqFlag`. Additionally `bpaf` handles `()`
// fields as `ReqFlag`.
// #[cfg_attr(not(doctest), doc = include_str!("docs2/ReqFlag.md"))]
func ReqFlag[T any](self NamedArg, present T) ParseFlag[T] {
	return ParseFlag[T]{present, Invalid[T](), self}
}
