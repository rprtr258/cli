package cli

import (
	"os"

	"github.com/rprtr258/fun"
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

// Parse a subcommand
//
// Subcommands allow to use a totally independent parser inside a current one. Inner parser
// can have its own help message, description, version and so on. You can nest them arbitrarily
// too.
//
// # Important restriction
// When parsing Command arguments from Command lines you should have parsers for all your
// named values before parsers for commands and positional items. In derive API fields parsed as
// positional should be at the end of your `struct`/`enum`. Same rule applies
// to parsers with positional fields or commands inside: such parsers should go to the end as well.
//
// Use [check_invariants](OptionParsercheck_invariants) in your test to ensure correctness.
//
// For example for non positional `non_pos` and a Command `Command` parsers
// ```rust
// # use bpaf*;
// # non_pos := || short('n').switch();
// # Command := || pure(struct{}{}).to_options().Command("POS");
// valid := construct(non_pos(), Command());
// invalid := construct(Command(), non_pos());
// ```
//
// **`bpaf` panics during help generation unless if this restriction holds**
//
// You can attach a single visible short alias and multiple hiddden short and long aliases
// using [short](ParseCommandshort) and [long](ParseCommandlong) methods.
//
// #[cfg_attr(not(doctest), doc = include_str!("docs2/Command.md"))]
//
// To represent multiple possible commands it is convenient to use enums
// #[cfg_attr(not(doctest), doc = include_str!("docs2/command_enum.md"))]
func (self OptionParser[T]) Command(name string) *ParseCommand[T] {
	return &ParseCommand[T]{
		longs:     []string{name},
		shorts:    nil,
		help:      self.short_descr(),
		subparser: self,
		adjacent:  false,
	}
}

// Builder structure for the [command]
//
// Created with [command], implements parser for the inner structure, gives access to [help](ParseCommandhelp).
type ParseCommand[T any] struct {
	longs  []string
	shorts []byte
	// short help!
	help      Option[Doc]
	subparser OptionParser[T]
	adjacent  bool
}

// Add a brief description to a command
//
// `bpaf` uses this description along with the command name
// in help output so it shouldn't exceed one or two lines. If `help` isn't specified
// `bpaf` falls back to [descr](OptionParserdescr) from the inner parser.
//
// # Combinatoric usage
//
// ```rust
// # use bpaf*;
//
//	func inner()OptionParser<bool> {
//	    short('i')
//	        .help("Mysterious inner switch")
//	        .switch()
//	        .to_options()
//	        .descr("performs an operation")
//	}
//
//	func mysterious_parser()impl Parser<bool> {
//	    inner().command("mystery")
//	        .help("This command performs a mystery operation")
//	}
//
// ```
//
// # Derive usage
// `bpaf_derive` uses doc comments for inner parser, no specific fun.Options are available.
// See [descr](OptionParserdescr) for more details
// ```rust
// # use bpaf*;
// // This command performs a mystery operation
//
// #[bpaf(command)]
//
//	struct Mystery {
//	    #[bpaf(short)]
//	    // Mysterious inner switch
//	    inner bool,
//	}
//
// ```
//
// # Example
// ```console
// $ app --help
//
//	<skip>
//
// Available commands:
//
//	mystery  This command performs a mystery operation
//
// ```
func (self *ParseCommand[P]) Help(help string) *ParseCommand[P] {
	self.help = Valid(newDocFromString(help))
	return self
}

// Add a custom short alias for a command
//
// Behavior is similar to [short](NamedArgshort), only first short name is visible.
func (self *ParseCommand[P]) short(short byte) *ParseCommand[P] {
	self.shorts = append(self.shorts, short)
	return self
}

// Add a custom hidden long alias for a command
//
// Behavior is similar to [long](NamedArglong), but since you had to specify the first long
// name when making the command - this one becomes a hidden alias.
func (self *ParseCommand[P]) long(long string) *ParseCommand[P] {
	self.longs = append(self.longs, long)
	return self
}

// Allow for the command to succeed even if there are non consumed items present
//
// Normally a subcommand parser should handle the rest of the unconsumed elements thus
// allowing only "vertical" chaining of commands. `adjacent` modifier lets command parser to
// succeed if there are leftovers for as long as all comsumed items form a single adjacent
// block. This opens possibilities to chain commands sequentially.
//
// Let's consider two examples with consumed items marked in bold :
//
// - <code>**cmd** **-a** -b **-c** -d</code>
// - <code>**cmd** **-a** **-c** -b -d</code>
//
// In the first example `-b` breaks the adjacency for all the consumed items so parsing will fail,
// while here in the second one the name and all the consumed items are adjacent to each other so
// parsing will succeed.
//
// #[cfg_attr(not(doctest), doc = include_str!("docs2/adjacent_command.md"))]
func (self *ParseCommand[P]) Adjacent() *ParseCommand[P] {
	self.adjacent = true
	return self
}

func (self ParseCommand[T]) Parser() Parser[T] {
	return Parser[T]{self.eval, MetaItem(self.item())}
}

func (self ParseCommand[T]) eval(args *State) (T, Error) {
	// used to avoid allocations for short names
	tmp := ""
	if fun.Any(func(long string) bool { return args.take_cmd(long) }, self.longs...) || fun.Any(func(s byte) bool {
		tmp = string([]byte{s})
		return args.take_cmd(tmp)
	}, self.shorts...) {
		if args.touching_last_remove() {
			// in completion mode prefer to autocomplete the command name vs going inside the
			// parser
			args.clear_comps()
			args.push_command(self.longs[0], First(self.shorts), self.help)
			return *new(T), Error(MessageMissing())
		}

		if cur, ok := args.current.Unpack(); ok {
			args.set_scope(Range{cur, args.scope.end()})
		}

		args.path = append(args.path, self.longs[0])
		if self.adjacent {
			orig_args := args

			// narrow down the scope to adjacently available elements
			args.set_scope(args.adjacently_available_from(args.scope.start() + 1))

			ok, xxx := self.
				subparser.
				run_subparser(args)
			var err Message
			if xxx != nil {
				err = MessageParseFailure(xxx)
			}
			if err == nil {
				args.set_scope(orig_args.scope)
				return ok, nil
			} else {
				orig_scope := args.scope
				if narrow_scope, ok := args.adjacent_scope(*orig_args).Unpack(); ok {
					orig_args.set_scope(narrow_scope)
					if res, err := self.subparser.run_subparser(orig_args); err == nil {
						orig_args.set_scope(orig_scope)
						stdmemswap(orig_args, args)
						return res, nil
					}
				}
				return *new(T), Error(err)
			}
		} else {
			v, e := self.subparser.
				run_subparser(args)
			if e != nil {
				return *new(T), Error(MessageParseFailure(e))
			}
			return v, nil
		}
	} else {
		args.push_command(self.longs[0], First(self.shorts), self.help)

		missing := MissingItem{
			item:     self.item(),
			position: args.scope.start(),
			scope:    args.scope,
		}
		return *new(T), Error(MessageMissing(missing))
	}
}

func (self ParseCommand[T]) item() Item {
	return ItemCommand(
		self.longs[0],
		First(self.shorts),
		self.help,
		self.subparser.inner.meta,
		self.subparser.info,
	)
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

// Add a help message to an `argument`
//
// See [NamedArghelp]
func (self *ParseArgument[T]) help(help Doc) *ParseArgument[T] {
	self.named.help = Valid(help)
	return self
}

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

func (self ParseArgument[T]) take_argument(args *State) Result[string] {
	wopt, err := args.take_arg(self.named, self.adjacent, Metavar(self.metavar)).unpack()
	if err != nil {
		args.push_argument(self.named, self.metavar)
		return Err[string](err)
	} else if w, ok := wopt.Unpack(); ok {
		if args.touching_last_remove() {
			args.push_metavar(self.metavar, self.named.help, true)
		}
		return Ok(w)
	} else {
		args.push_argument(self.named, self.metavar)
		if val, ok := find_map(self.named.env, func(e string) Option[string] {
			r, ok := os.LookupEnv(e)
			return Option[string]{r, ok}
		}).Unpack(); ok {
			args.current = Invalid[int]()
			return Ok(val)
		}

		if item, ok := self.item().Unpack(); ok {
			missing := MissingItem{
				item:     item,
				position: args.scope.start(),
				scope:    args.scope,
			}
			return Err[string](Error(MessageMissing(missing)))
		} else if name, ok := First(self.named.env).Unpack(); ok {
			return Err[string](Error(MessageNoEnv(name)))
		} else {
			panic("unreachable!()")
		}
	}
}

func (self ParseArgument[T]) eval(args *State) (T, Error) {
	os, err := self.take_argument(args).unpack()
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

// Consume an arbitrary value that satisfies a condition, created with [any], implements [anywhere](ParseAnyanywhere).
type ParseAny[T any] struct {
	metavar  Doc
	help     Option[Doc]
	check    func(string) Option[T]
	anywhere bool
}

func (self ParseAny[T]) Parser() Parser[T] {
	return Parser[T]{self.eval, MetaItem(self.item())}
}

func (self ParseAny[T]) item() Item {
	return ItemAny(
		self.metavar,
		self.anywhere,
		self.help,
	)
}

// Add a help message to [any] parser.
// See examples in [any]
func (self *ParseAny[T]) Help(help string) *ParseAny[T] {
	self.help = Valid(newDocFromString(help))
	return self
}

// Replace metavar with a custom value
// See examples in [any]
func (self *ParseAny[T]) Metavar(metavar Doc) *ParseAny[T] {
	self.metavar = metavar
	return self
}

// Try to apply the parser to each unconsumed element instead of just the front one
//
// By default `any` tries to parse just the front unconsumed item behaving similar to
// [positional] parser, `anywhere` changes it so it applies to every unconsumed item,
// similar to argument parser.
//
// See examples in [any]
func (self *ParseAny[T]) Anywhere() *ParseAny[T] {
	self.anywhere = true
	return self
}

func (self ParseAny[T]) eval(args *State) (T, Error) {
	for ix, x := range push2(args.items_iter()) {
		var next bool
		x.value(
			func(_ byte, c bool) { next = c },
			func(_ string, c bool) { next = c },
			nil, nil, nil,
		)
		os := x.str
		if i, ok := (self.check)(os).Unpack(); ok {
			args.remove(ix)
			if next {
				args.remove(ix + 1)
			}

			return i, nil
		}
		if !self.anywhere {
			break
		}
	}
	missing_item := MissingItem{
		item:     self.item(),
		position: args.scope.start(),
		scope:    args.scope,
	}
	return *new(T), Error(MessageMissing(missing_item))
}
