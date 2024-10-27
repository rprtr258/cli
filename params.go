package cli

import (
	"os"
)

//! Tools to define primitive parsers
//!
//! # Ways to consume data
//!
//! ## Flag
//!
//! - [flag](NamedArgflag) - a string that consists of two dashes (`--flag`) and a name and a single
//! dash and a single character (`-f`) created with [long](NamedArglong) and [short](NamedArgshort)
//! respectively. Depending if this name is present or absent on the command line
//! primitive flag parser produces one of two values. User can combine several short flags in a single
//! invocation: `-a -b -c` is the same as `-abc`.
//!
// #![cfg_attr(not(doctest), doc = include_str!("docs2/flag.md"))]
//!
//! ## Required flag
//!
//! Similar to `flag`, but instead of falling back to the second value required flag parser would
//! fail. Mostly useful in combination with other parsers, created with [NamedArgreq_flag].
//!
// #![cfg_attr(not(doctest), doc = include_str!("docs2/req_flag.md"))]
//!
//! ## Switch
//!
//! A special case of a flag that gets decoded into a `bool`, mostly serves as a convenient
//! shortcut to `.flag(true, false)`. Created with [NamedArgswitch].
//!
// #![cfg_attr(not(doctest), doc = include_str!("docs2/switch.md"))]
//!
//! ## Argument
//!
//! A short or long `flag` followed by either a space or `=` and
//! then by a string literal.  `-f foo`, `--flag bar` or `-o=-` are all valid argument examples. Note, string
//! literal can't start with `-` unless separated from the flag with `=`. For short flags value
//! can follow immediately: `-fbar`.
//!
// #![cfg_attr(not(doctest), doc = include_str!("docs2/argument.md"))]
//!
//! ## Positional
//!
//! A positional argument with no additonal name, for example in `vim main.rs` `main.rs`
//! is a positional argument. Can't start with `-`, created with [positional].
//!
// #![cfg_attr(not(doctest), doc = include_str!("docs2/positional.md"))]
//!
//! ## Any
//!
//! Also a positional argument with no additional name, but unlike [positional] itself, [any]
//! isn't restricted to positional looking structure and would consume any items as they appear on
//! a command line. Can be useful to collect anything unused to pass to other applications.
//!
// #![cfg_attr(not(doctest), doc = include_str!("docs2/any_simple.md"))]
// #![cfg_attr(not(doctest), doc = include_str!("docs2/any_literal.md"))]
//!
//! ## Command
//!
//! A command defines a starting point for an independent subparser. Name must be a valid utf8
//! string. For example `cargo build` invokes command `"build"` and after `"build"` `cargo`
//! starts accepting values it won't accept otherwise
//!
// #![cfg_attr(not(doctest), doc = include_str!("docs2/command.md"))]
//!

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

func build_flag_parser[T any](present T, absent Option[T], named NamedArg) ParseFlag[T] {
	return ParseFlag[T]{
		present,
		absent,
		named,
	}
}

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
