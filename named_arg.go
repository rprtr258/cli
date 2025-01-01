package cli

import "github.com/rprtr258/fun"

// A named thing used to create [flag](NamedArgflag), [switch](NamedArgswitch) or
// [argument](NamedArgargument)
//
// # Combinatoric usage
//
// Named items (`argument`, `flag` and `switch`) can have up to 2 visible names (one short and one long)
// and multiple hidden short and long aliases if needed. It's also possible to consume items from
// environment variables using [env](NamedArgenv()). You usually start with [short] or [long]
// function, then apply [short](NamedArgshort) / [long](NamedArglong) / [env](NamedArgenv()) /
// [help](NamedArghelp) repeatedly to build a desired set of names then transform it into
// a parser using `flag`, `switch` or `positional`.
//
// #[cfg_attr(not(doctest), doc = include_str!("docs2/named_arg_combine.md"))]
//
// # Derive usage
//
// When using derive API it is possible to omit some or all the details:
//  1. If no naming information is present at all - `bpaf` would use field name as a long name
//     (or a short name if field name consists of a single character)
//  2. If `short` or `long` annotation is present without an argument - `bpaf` would use first character
//     or a full name as long and short name respectively. It won't try to add implicit long or
//     short name from the previous item.
//  3. If `short` or `long` annotation is present with an argument - those are values `bpaf` would
//     use instead of the original field name
//  4. You can specify many `short` and `long` names, any past the first one of each type will
//     become hidden aliases
//  5. If `env(arg)` annotation is present - in addition to long/short names derived according to
//     rules 1..3 `bpaf` would also parse environment variable `arg` which can be a string literal
//     or an expression.
//
// #[cfg_attr(not(doctest), doc = include_str!("docs2/named_arg_derive.md"))]
type NamedArg struct {
	short []byte
	long  []string
	env   []string
	help  Option[Doc]
}

func (self NamedArg) clone() NamedArg {
	return NamedArg{
		append([]byte(nil), self.short...),
		append([]string(nil), self.long...),
		append([]string(nil), self.env...),
		self.help.clone(),
	}
}

func (self NamedArg) flag_item() (Item, bool) {
	x, ok := ShortLongtry_from(self)
	if !ok {
		return Item{}, false
	}
	return Item{self.help, ItemFlag{x, self.short, First(self.env)}}, true
}

// Add a short name to a flag/switch/argument
//
// #[cfg_attr(not(doctest), doc = include_str!("docs2/short_long_env.md"))]
func (self NamedArg) Short(short byte) NamedArg {
	self.short = append(self.short, short)
	return self
}

// Add a long name to a flag/switch/argument
//
// #[cfg_attr(not(doctest), doc = include_str!("docs2/short_long_env.md"))]
func (self NamedArg) Long(long string) NamedArg {
	self.long = append(self.long, long)
	return self
}

// Environment variable fallback
//
// If named value isn't present - try to fallback to this environment variable.
//
// You can specify it multiple times, `bpaf` would use items past the first one as hidden aliases.
//
// For [flag](NamedArgflag) and [switch](NamedArgswitch) environment variable being present
// gives the same result as the flag being present, allowing to implement things like `NO_COLOR`
// variables:
//
// ```console
// $ NO_COLOR=1 app --do-something
// ```
// #[cfg_attr(not(doctest), doc = include_str!("docs2/short_long_env.md"))]
func (self NamedArg) Env(variable string) NamedArg {
	self.env = append(self.env, variable)
	return self
}

// Add a help message to a `flag`/`switch`/`argument`
//
// `bpaf` converts doc comments and string into help by following those rules:
// 1. Everything up to the first blank line is included into a "short" help message
// 2. Everything is included into a "long" help message
// 3. `bpaf` preserves linebreaks followed by a line that starts with a space
// 4. Linebreaks are removed otherwise
//
// You can pass anything that can be converted into [Doc], if you are not using
// documentation generation functionality ([doc](cratedoc)) this can be `string`.
//
// #[cfg_attr(not(doctest), doc = include_str!("docs2/switch_help.md"))]
func (self NamedArg) Help(help string) NamedArg {
	self.help = Valid(newDocFromString(help))
	return self
}

// Simple boolean flag
//
// A special case of a [flag](NamedArgflag) that gets decoded into a `bool`, mostly serves as a convenient
// shortcut to `.flag(true, false)`.
//
// In Derive API bpaf would use `switch` for `bool` fields inside named structs that don't
// have other consumer annotations ([flag](NamedArgflag),
// [argument](NamedArgargument), etc).
//
// #[cfg_attr(not(doctest), doc = include_str!("docs2/switch.md"))]
func (self NamedArg) Switch() ParseFlag[bool] {
	return build_flag_parser(true, Valid(false), self)
}

// `adjacent` requires for the argument to be present in the same word as the flag:
// `-f bar` - no, `-fbar` or `-f=bar` - yes.
func (self NamedArg) matches_arg(arg Arg, adjacent bool) bool {
	res := false
	switch arg.kind {
	case ArgKindShortFlag:
		res = fun.Contains(arg.c, self.short...) && (!adjacent || arg.is_adj)
	case ArgKindLongFlag:
		res = fun.Contains(arg.s, self.long...) && (!adjacent || arg.is_adj)
	}
	return res
}

// Parse a [flag](NamedArgflag)/[switch](NamedArgswitch)/[argument](NamedArgargument) that has a Short name
//
// You can chain multiple [Short](NamedArgshort), [long](NamedArglong) and
// [env](NamedArgenv()) for multiple names. You can specify multiple names of the same type,
//
//	`bpaf` would use items past the first one as hidden aliases.
//
// #[cfg_attr(not(doctest), doc = include_str!("docs2/short_long_env.md"))]
func Short(short byte) NamedArg {
	return NamedArg{
		short: []byte{short},
		env:   nil,
		long:  nil,
		help:  Invalid[Doc](),
	}
}

// Parse a [flag](NamedArgflag)/[switch](NamedArgswitch)/[argument](NamedArgargument) that has a Long name
//
// You can chain multiple [short](NamedArgshort), [Long](NamedArglong) and
// [env](NamedArgenv()) for multiple names. You can specify multiple names of the same type,
//
//	`bpaf` would use items past the first one as hidden aliases.
//
// #[cfg_attr(not(doctest), doc = include_str!("docs2/short_long_env.md"))]
func Long(long string) NamedArg {
	return NamedArg{
		short: nil,
		long:  []string{long},
		env:   nil,
		help:  Invalid[Doc](),
	}
}

// Parse an environment variable
//
// You can chain multiple [short](NamedArgshort), [long](NamedArglong) and
// [Env](NamedArgenv()) for multiple names. You can specify multiple names of the same type,
//
//	`bpaf` would use items past the first one as hidden aliases.
//
// For [flag](NamedArgflag) and [switch](NamedArgswitch) environment variable being present
// gives the same result as the flag being present, allowing to implement things like `NO_COLOR`
// variables:
//
// ```console
// $ NO_COLOR=1 app --do-something
// ```
//
// If you don't specify a short or a long name - whole argument is going to be absent from the
// help message. Use it combined with a named or positional argument to have a hidden fallback
// that wouldn't leak sensitive info.
//
// #[cfg_attr(not(doctest), doc = include_str!("docs2/short_long_env.md"))]
func Env(variable string) NamedArg {
	return NamedArg{
		short: nil,
		long:  nil,
		help:  Invalid[Doc](),
		env:   []string{variable},
	}
}
