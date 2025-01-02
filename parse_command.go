package cli

import "github.com/rprtr258/fun"

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
//
//	func inner()OptionParser<bool> {
//	    short('i')
//	        .help("Mysterious inner switch")
//	        .Switch()
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
	self.help = Valid(NewDoc(help))
	return self
}

// Add a custom short alias for a command
//
// Behavior is similar to [short](NamedArgshort), only first short name is visible.
func (self *ParseCommand[P]) Short(short byte) *ParseCommand[P] {
	self.shorts = append(self.shorts, short)
	return self
}

// Add a custom hidden long alias for a command
//
// Behavior is similar to [long](NamedArglong), but since you had to specify the first long
// name when making the command - this one becomes a hidden alias.
func (self *ParseCommand[P]) Long(long string) *ParseCommand[P] {
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

func (self ParseCommand[T]) eval(state *State) (T, Error) {
	// used to avoid allocations for short names
	tmp := ""
	if fun.Any(func(long string) bool { return state.take_cmd(long) }, self.longs...) || fun.Any(func(s byte) bool {
		tmp = string([]byte{s})
		return state.take_cmd(tmp)
	}, self.shorts...) {
		if state.touching_last_remove() {
			// in completion mode prefer to autocomplete the command name vs going inside the parser
			state.clear_comps()
			state.push_command(self.longs[0], First(self.shorts), self.help)
			return *new(T), ErrorMissing(nil)
		}

		if cur, ok := state.currentIdx.Unpack(); ok {
			state.scope = Range{cur, state.scope.end()}
		}

		state.path = append(state.path, self.longs[0])
		if self.adjacent {
			orig_args := state.clone()

			// narrow down the scope to adjacently available elements
			state.scope = state.adjacently_available_from(state.scope.start() + 1)

			ok, xxx := self.subparser.run_subparser(state)
			if xxx != nil {
				orig_scope := state.scope
				if narrow_scope, ok := state.adjacent_scope(orig_args); ok {
					orig_args.scope = narrow_scope
					if res, err := self.subparser.run_subparser(&orig_args); err == nil {
						orig_args.scope = orig_scope
						orig_args, *state = *state, orig_args
						return res, nil
					}
				}
				return *new(T), Error(ErrorParseFailure{xxx})
			}

			state.scope = orig_args.scope
			return ok, nil
		} else {
			v, e := self.subparser.run_subparser(state)
			if e != nil {
				return *new(T), ErrorParseFailure{e}
			}
			return v, nil
		}
	} else {
		state.push_command(self.longs[0], First(self.shorts), self.help)
		return *new(T), ErrorMissing([]MissingItem{{
			item:     self.item(),
			position: state.scope.start(),
			scope:    state.scope,
		}})
	}
}

func (self ParseCommand[T]) item() Item {
	return Item{self.help, ItemCommand{
		self.longs[0],
		First(self.shorts),
		self.subparser.inner.meta,
		self.subparser.info,
	}}
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
// # non_pos := || short('n').Switch();
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

// Deprecated: You should switch from command(name, sub) to sub.command(name)
func command[T any](name string, subparser OptionParser[T]) ParseCommand[T] {
	return ParseCommand[T]{
		longs:     []string{name},
		shorts:    nil,
		help:      subparser.short_descr(),
		subparser: subparser,
		adjacent:  false,
	}
}
