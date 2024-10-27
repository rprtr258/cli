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

func (self ParseFlag[T]) eval(args *State) (T, Error) {
	if args.take_flag(self.named) || find_map(self.named.env, func(e string) Option[string] {
		r, ok := os.LookupEnv(e)
		return Option[string]{r, ok}
	}).Valid {
		if args.touching_last_remove() {
			args.push_flag(self.named)
		}
		return self.present, nil
	} else {
		args.push_flag(self.named)
		if ok, okk := self.absent.Unpack(); okk {
			return ok, nil
		} else {
			if item, ok := self.named.flag_item().Unpack(); ok {
				missing := MissingItem{
					item:     item,
					position: args.scope.start(),
					scope:    args.scope,
				}
				return *new(T), Error(MessageMissing(missing))
			} else if name, ok := First(self.named.env).Unpack(); ok {
				return *new(T), Error(MessageNoEnv(name))
			} else {
				panic(`todo!("no key!")`)
			}
		}
	}
}

func (self ParseFlag[T]) meta() Meta {
	if item, ok := self.named.flag_item().Unpack(); ok {
		return item.required(!self.absent.Valid)
	} else {
		return MetaSkip
	}
}

// Add a help message to `flag`
//
// See [NamedArghelp]
func (self *ParseFlag[T]) help(help Doc) *ParseFlag[T] {
	self.named.help = Valid(help)
	return self
}

func build_flag_parser[T any](present T, absent Option[T], named NamedArg) ParseFlag[T] {
	return ParseFlag[T]{
		present,
		absent,
		named,
	}
}

// Flag with custom present/absent values
//
// More generic version of [switch](NamedArgswitch) that can use arbitrary type instead of
// [bool].
// #[cfg_attr(not(doctest), doc = include_str!("docs2/Flag.md"))]
func Flag[T any](self NamedArg, present T, absent T) ParseFlag[T] {
	return build_flag_parser(present, Valid(absent), self)
}

// Required flag with custom value
//
// Similar to [flag](NamedArgflag) takes no fun.Option arguments, but would only
// succeed if user specifies its name on a command line.
// Works best in combination with other parsers.
//
// In derive style API `bpaf` would transform field-less enum variants into a parser
// that accepts one of it's variant names as `ReqFlag`. Additionally `bpaf` handles `()`
// fields as `ReqFlag`.
// #[cfg_attr(not(doctest), doc = include_str!("docs2/ReqFlag.md"))]
func ReqFlag[T any](self NamedArg, present T) ParseFlag[T] {
	return build_flag_parser(present, Invalid[T](), self)
}
