package cli

import (
	"iter"

	"github.com/rprtr258/fun"
)

// ParseCommand is a builder structure for a subcommand parser.
//
// Created with [Command]; implements a parser for the inner structure, allows attaching
// a help message with [ParseCommand.Help].
type ParseCommand[T any] struct {
	longs  []string
	shorts []byte
	// short help!
	help      Opt[Doc]
	subparser OptionParser[T]
	adjacent  bool
}

// Help adds a brief description to a command.
//
// The library uses this description along with the command name in the help output
// so it shouldn't exceed one or two lines. If `help` isn't specified the library falls
// back to the description of the inner parser set with [OptionParser.Description].
//
// # Usage
//
//	inner := cli.Short('i').
//	    Help("Mysterious inner switch").
//	    Switch().
//	    Parser().
//	    ToOptions().
//	    Description("performs an operation")
//
//	mysteryParser := inner.
//	    Command("mystery").
//	    Help("This command performs a mystery operation")
//
// # Example
//
//	$ app --help
//	Available commands:
//	  mystery  This command performs a mystery operation// # Example
//
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

// Short adds a custom short alias for a command.
//
// Behavior is similar to [NamedArg.Short]: only the first short name is visible.
func (self *ParseCommand[P]) Short(short byte) *ParseCommand[P] {
	self.shorts = append(self.shorts, short)
	return self
}

// Long adds a custom hidden long alias for a command.
//
// Behavior is similar to [NamedArg.Long], but since you had to specify the first long
// name when making the command - this one becomes a hidden alias.
func (self *ParseCommand[P]) Long(long string) *ParseCommand[P] {
	self.longs = append(self.longs, long)
	return self
}

// Adjacent allows for the command to succeed even if there are unconsumed items present.
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
// The library panics during help generation if this restriction is violated.
func (self *ParseCommand[P]) Adjacent() *ParseCommand[P] {
	self.adjacent = true
	return self
}

func (self ParseCommand[T]) Parser() Parser[T] {
	return Parser[T]{self.eval, metaItem(self.item())}
}

// ToOptions prepares the parser for running, see [Parser.ToOptions].
func (self ParseCommand[T]) ToOptions() *OptionParser[T] { return self.Parser().ToOptions() }

// Map applies a pure transformation to the parsed command value, see [Parser.Map].
func (self ParseCommand[T]) Map[R any](fn func(T) R) Parser[R] { return self.Parser().Map(fn) }

// Hide hides this parser from help generation, see [Parser.Hide].
func (self ParseCommand[T]) Hide() Parser[T] { return self.Parser().Hide() }

// HideUsage hides this parser from the usage line, see [Parser.HideUsage].
func (self ParseCommand[T]) HideUsage() Parser[T] { return self.Parser().HideUsage() }

// CustomUsage overrides the usage line, see [Parser.CustomUsage].
func (self ParseCommand[T]) CustomUsage(usage Doc) Parser[T] { return self.Parser().CustomUsage(usage) }

// GroupHelp attaches a help message to this parser, see [Parser.GroupHelp].
func (self ParseCommand[T]) GroupHelp(message string) Parser[T] {
	return self.Parser().GroupHelp(message)
}

// Guard validates the parsed value, see [Parser.Guard].
func (self ParseCommand[T]) Guard(check func(T) bool, message string) Parser[T] {
	return self.Parser().Guard(check, message)
}

// Fallback provides a default value, see [Parser.Fallback].
func (self ParseCommand[T]) Fallback(value T) Parser[T] { return self.Parser().Fallback(value) }

// FallbackDisplay provides a default value shown in help, see [Parser.FallbackDisplay].
func (self ParseCommand[T]) FallbackDisplay(value T) Parser[T] {
	return self.Parser().FallbackDisplay(value)
}

// FallbackWith provides a lazily computed default value, see [Parser.FallbackWith].
func (self ParseCommand[T]) FallbackWith(fallback func() (T, error)) Parser[T] {
	return self.Parser().FallbackWith(fallback)
}

// FallbackWithDisplay provides a lazily computed default value shown in help, see [Parser.FallbackWithDisplay].
func (self ParseCommand[T]) FallbackWithDisplay(fallback func() (T, error)) Parser[T] {
	return self.Parser().FallbackWithDisplay(fallback)
}

// Last keeps only the last matched value, see [Parser.Last].
func (self ParseCommand[T]) Last() Parser[T] { return self.Parser().Last() }

// Count the number of matches, see [Parser.Count].
func (self ParseCommand[T]) Count() Parser[int] { return self.Parser().Count() }

// Complete adds custom shell completion, see [Parser.Complete].
func (self ParseCommand[T]) Complete(op func(T) iter.Seq2[string, Opt[string]]) ParseComp[T] {
	return self.Parser().Complete(op)
}

// CompleteShell adds file or directory shell completion, see [Parser.CompleteShell].
func (self ParseCommand[T]) CompleteShell(op ShellComp) Parser[T] {
	return self.Parser().CompleteShell(op)
}

func (self ParseCommand[T]) eval(state *state) (T, parseError) {
	// used to avoid allocations for short names
	tmp := ""
	if fun.Any(func(long string) bool { return state.takeCmd(long) }, self.longs...) || fun.Any(func(s byte) bool {
		tmp = string([]byte{s})
		return state.takeCmd(tmp)
	}, self.shorts...) {
		if state.touching_last_remove() {
			// in completion mode prefer to autocomplete the command name vs going inside the parser
			state.clear_comps()
			state.push_command(self.longs[0], first(self.shorts), self.help)
			return *new(T), errorMissing(nil)
		}

		if cur, ok := state.currentIdx.Unpack(); ok {
			state.scope = rng{cur, state.scope.end()}
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
				return *new(T), parseError(errorParseFailure{xxx})
			}

			state.scope = orig_args.scope
			return ok, nil
		} else {
			v, e := self.subparser.run_subparser(state)
			if e != nil {
				return *new(T), errorParseFailure{e}
			}
			return v, nil
		}
	} else {
		state.push_command(self.longs[0], first(self.shorts), self.help)
		return *new(T), errorMissing([]missingItem{{
			item:     self.item(),
			position: state.scope.start(),
			scope:    state.scope,
		}})
	}
}

func (self ParseCommand[T]) item() item {
	return item{self.help, itemCommand{
		self.longs[0],
		first(self.shorts),
		self.subparser.inner.meta,
		self.subparser.info,
	}}
}

// Command parses a subcommand.
//
// Subcommands allow to use a totally independent parser inside a current one. Inner parser
// can have its own help message, description, version and so on. You can nest them arbitrarily
// too.
//
// # Important restriction
// When parsing command arguments from command lines you should place parsers for all your
// named values before parsers for commands and positional items. With [Construct] the fields
// parsed as commands or positional items should be at the end of the struct. The same rule applies
// to parsers with positional fields or commands inside: such parsers should go to the end as well.
//
// Use [OptionParser.CheckInvariants] in your tests to ensure correctness.
//
// For example for a non-positional `nonPos` and a command `command` parsers:
//
//	valid := cli.Pair(nonPos, command)
//	invalid := cli.Pair(command, nonPos)
//
// **The library panics during help generation if this restriction is violated.**
//
// You can attach a single visible short alias and multiple hidden short and long aliases
// using [ParseCommand.Short] and [ParseCommand.Long] methods.
//
// To represent multiple possible commands with [Construct] define a field per command in the
// struct and make each one [Parser.Optional].
func (self OptionParser[T]) Command(name string) *ParseCommand[T] {
	return &ParseCommand[T]{
		longs:     []string{name},
		shorts:    nil,
		help:      self.short_descr(),
		subparser: self,
		adjacent:  false,
	}
}

// command builds a subcommand parser with the given long name.
//
// Deprecated: switch from command(name, sub) to sub.command(name).
func command[T any](name string, subparser OptionParser[T]) ParseCommand[T] {
	return ParseCommand[T]{
		longs:     []string{name},
		shorts:    nil,
		help:      subparser.short_descr(),
		subparser: subparser,
		adjacent:  false,
	}
}
