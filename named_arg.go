package cli

import "github.com/rprtr258/fun"

// NamedArg is a named thing used to create a flag with [Flag] or [ReqFlag],
// a switch with [NamedArg.Switch] or an argument with [NamedArg.Argument].
//
// # Combinatoric usage
//
// Named items can have up to 2 visible names (one short and one long)
// and multiple hidden short and long aliases if needed. It's also possible to consume items from
// environment variables using [NamedArg.Env]. You usually start with [Short], [Long] or [Env],
// then apply [NamedArg.Short] / [NamedArg.Long] / [NamedArg.Env] / [NamedArg.Help] repeatedly
// to build the desired set of names, then transform it into a parser using [NamedArg.Switch],
// [Flag], [ReqFlag] or [NamedArg.Argument].
//
// With [Construct] the `cli` struct tag names each field: `short=v`, `long=name`,
// `env=VAR` choose the names, additional `short`/`long` entries become hidden aliases.
type NamedArg struct {
	short []byte
	long  []string
	env   []string
	help  Opt[Doc]
}

func (self NamedArg) clone() NamedArg {
	return NamedArg{
		append([]byte(nil), self.short...),
		append([]string(nil), self.long...),
		append([]string(nil), self.env...),
		self.help.clone(),
	}
}

func (self NamedArg) flag_item() (item, bool) {
	x, ok := shortLongTryFrom(self)
	if !ok {
		return item{}, false
	}
	return item{self.help, itemFlag{x, self.short, first(self.env)}}, true
}

// Short adds a short name to a flag/switch/argument.
//
// You can chain multiple [NamedArg.Short], [NamedArg.Long] and [NamedArg.Env] calls
// for multiple names. Multiple names of the same type are allowed -
// items past the first one become hidden aliases.
func (self NamedArg) Short(short byte) NamedArg {
	self.short = append(self.short, short)
	return self
}

// Long adds a long name to a flag/switch/argument.
//
// You can chain multiple [NamedArg.Short], [NamedArg.Long] and [NamedArg.Env] calls
// for multiple names. Multiple names of the same type are allowed -
// items past the first one become hidden aliases.
func (self NamedArg) Long(long string) NamedArg {
	self.long = append(self.long, long)
	return self
}

// Env adds an environment variable fallback.
//
// If the named value isn't present - try to fallback to this environment variable.
//
// You can specify it multiple times, items past the first one become hidden aliases.
//
// For [Flag] and [NamedArg.Switch] the environment variable being present
// gives the same result as the flag being present, allowing to implement things like `NO_COLOR`
// variables:
//
//	$ NO_COLOR=1 app --do-something
func (self NamedArg) Env(variable string) NamedArg {
	self.env = append(self.env, variable)
	return self
}

// Help adds a help message to a flag/switch/argument.
//
// The help message is displayed in the --help output next to the option name.
func (self NamedArg) Help(help string) NamedArg {
	self.help = Valid(NewDoc(help))
	return self
}

// Switch is a simple boolean flag.
//
// A special case of [Flag] that produces a bool - a convenient shortcut
// for `Flag(self, true, false)`.
//
// With [Construct] a `bool` field without a consumer tag parses as a switch.
func (self NamedArg) Switch() Parser[bool] {
	return flagParser(true, Valid(false), self)
}

// matches_arg reports whether the argument matches the named item. With
// adjacent set, the argument must be present in the same word as the flag:
// `-f bar` - no, `-fbar` or `-f=bar` - yes.
func (self NamedArg) matches_arg(arg arg, adjacent bool) bool {
	return (!adjacent || arg.isAdjacent) &&
		(arg.kind == argKindShortFlag && fun.Contains(arg.c, self.short...) ||
			arg.kind == argKindLongFlag && fun.Contains(arg.s, self.long...))
}

// Short starts building a flag/switch/argument from a short name.
//
// You can chain multiple [NamedArg.Short], [NamedArg.Long] and [NamedArg.Env] calls
// for multiple names. Multiple names of the same type are allowed -
// items past the first one become hidden aliases.
func Short(short byte) NamedArg {
	return NamedArg{
		short: []byte{short},
		env:   nil,
		long:  nil,
		help:  Invalid[Doc](),
	}
}

// Long starts building a flag/switch/argument from a long name.
//
// You can chain multiple [NamedArg.Short], [NamedArg.Long] and [NamedArg.Env] calls
// for multiple names. Multiple names of the same type are allowed -
// items past the first one become hidden aliases.
func Long(long string) NamedArg {
	return NamedArg{
		short: nil,
		long:  []string{long},
		env:   nil,
		help:  Invalid[Doc](),
	}
}

// Env starts building a flag/switch/argument from an environment variable.
//
// You can chain multiple [NamedArg.Short], [NamedArg.Long] and [NamedArg.Env] calls
// for multiple names. Multiple names of the same type are allowed -
// items past the first one become hidden aliases.
//
// For [Flag] and [NamedArg.Switch] the environment variable being present
// gives the same result as the flag being present, allowing to implement things like `NO_COLOR`
// variables:
//
//	$ NO_COLOR=1 app --do-something
//
// If you don't specify a short or a long name - the whole item is going to be absent from the
// help message. Use it combined with a named or positional argument to have a hidden fallback
// that wouldn't leak sensitive info.
func Env(variable string) NamedArg {
	return NamedArg{
		short: nil,
		long:  nil,
		help:  Invalid[Doc](),
		env:   []string{variable},
	}
}
