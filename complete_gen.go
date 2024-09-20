package cli

import (
	"fmt"

	"github.com/rprtr258/fun"
)

type CompExtra struct {
	depth uint               // Used by complete_gen to separate commands from each other
	group fun.Option[string] // Render this fun.Option in a group along with all other items with the same name
	help  fun.Option[string] // help message attached to a completion item
}

// TODO: move out extra
type Comp = func(
	// short or long flag
	Flag func(extra CompExtra, name ShortLong),

	// argument + metadata
	Argument func(
		extra CompExtra,
		name ShortLong,
		metavar string,
	),

	Command func(
		extra CompExtra,
		name string,
		short fun.Option[byte],
	),

	// comes from completed values, part of "dynamic" completion
	Value func(
		extra CompExtra,
		body string,
		// values from arguments (say -p=SPEC and user already typed "-p b")
		// should suppress all other fun.Options except for metavaraiables?
		is_argument bool,
	),

	Metavariable func(
		extra CompExtra,
		meta string,
		// AKA not positional
		is_argument bool,
	),

	Shell func(
		extra CompExtra,
		script ShellComp,
		// AKA not positional
		is_argument bool,
	),
)

type Complete struct {
	// completions accumulated so far
	comps      []Comp
	output_rev uint

	// don't try to suggest any more positional items after there's a positional item failure
	// or parsing in progress
	no_pos_ahead bool
}

func Completenew(output_rev uint) Complete {
	return Complete{
		comps:        []Comp{},
		output_rev:   output_rev,
		no_pos_ahead: false,
	}
}

// completion:
// static: flag names, command names
// dynamic: argument values, positional item values
//
// for static when running collect any parser that fails
//
// OR: combine completions
// AND: usual logic without shortcircuits
//
// for static completion it's enough to collect items
// for argument completion - only one argument(CompMeta) should be active at once
//
// for rendering prefer longer version of names
//
// complete short names to long names if possible

type Complete struct {
	// completions accumulated so far
	comps      []Comp
	output_rev uint

	// don't try to suggest any more positional items after there's a positional item failure
	// or parsing in progress
	no_pos_ahead bool
}

func Completenew(output_rev uint) Complete {
	return Complete{
		comps: nil,
		output_rev,
		no_pos_ahead: false,
	}
}

// Add a new completion hint for flag, if needed
func (self *State) push_flag(named NamedArg) {
	depth := self.depth()
	if comp, ok := self.comp_mut().Unpack(); ok {
		if name, err := ShortLongtry_from(named); err == nil {
			comp.comps.push(CompFlag{
				extra: CompExtra{
					depth,
					group: fun.Invalid(),
					help:  named.help.as_ref().and_then(Docto_completion),
				},
				name,
			})
		}
	}
}

// Add a new completion hint for an argument, if needed
func (self *State) push_argument(named NamedArg, metavar string) {
	depth := self.depth()
	if comp, ok := self.comp_mut().Unpack(); ok {
		if name, err := ShortLongtry_from(named); err == nil {
			comp.comps.push(CompArgument{
				extra: CompExtra{
					depth,
					group: fun.Invalid(),
					help:  named.help.as_ref().and_then(Docto_completion),
				},
				metavar,
				name,
			})
		}
	}
}

// Add a new completion hint for metadata, if needed
//
// `is_argument` is set to true when we are trying to parse the value and false if
// when meta
func (self *State) push_metavar(
	meta string,
	help Option[Doc],
	is_argument bool,
) {
	depth := self.depth()
	if comp, ok := self.comp_mut().Unpack(); ok {
		extra := CompExtra{
			depth,
			group: fun.Invalid(),
			help:  help.as_ref().and_then(Docto_completion),
		}

		comp.comps.push(CompMetavariable{
			extra,
			meta,
			is_argument,
		})
	}
}

// Add a new completion hint for command, if needed
func (self *State) push_command(
	name string,
	short fun.Option[char],
	help Option[Doc],
) {
	depth := self.depth()
	if comp, ok := self.comp_mut().Unpack(); ok {
		comp.comps.push(CompCommand{
			extra: CompExtra{
				depth,
				group: fun.Invalid(),
				help:  help.as_ref().and_then(Docto_completion),
			},
			name,
			short,
		})
	}
}

// Clear collected completions if enabled
func (self *State) clear_comps() {
	if comp, ok := self.comp_mut().Unpack(); ok {
		comp.comps.clear()
	}
}

// Insert a literal value with some description for completion
//
// In practice it's "--"
func (self *State) push_pos_sep() {
	depth := self.depth()
	if comp, ok := self.comp_mut().Unpack(); ok {
		comp.comps.push(CompValue{
			extra: CompExtra{
				depth,
				group: fun.Invalid(),
				help:  fun.Valid("Positional only items after this token".to_owned()),
			},
			body:        "--".to_owned(),
			is_argument: false,
		})
	}
}

// Insert a bunch of items
func (self *State) push_with_group(group Option[string], comps *[]Comp) {
	if comp, ok := self.comp_mut().Unpack(); ok {
		for i := range comps {
			if group, ok := group.as_ref().Unpack(); ok {
				comps[i].set_group(group)
			}
			comp.comps.push(comps[i])
		}
	}
}

func (self *Complete) push_shell(op ShellComp, is_argument bool, depth uint) {
	self.comps.push(CompShell{
		extra: CompExtra{
			depth,
			group: fun.Invalid(),
			help:  fun.Invalid(),
		},
		script: op,
		is_argument,
	})
}

func (self *Complete) push_value(
	body string,
	help fun.Option[string],
	group fun.Option[string],
	depth uint,
	is_argument bool,
) {
	self.comps.push(CompValue{
		body,
		is_argument,
		extra: CompExtra{depth, group, help},
	})
}

func (self *Complete) push_comp(comp Comp) {
	self.comps.push(comp)
}

func (self *Complete) extend_comps(comps []Comp) {
	self.comps.extend(comps)
}

func (self *Complete) drain_comps() []Comp {
	return self.comps
}

func (self *Complete) swap_comps(other *[]Comp) {
	stdmemswap(other, *self.comps)
}

type CompExtra struct {
	// Used by complete_gen to separate commands from each other
	depth uint
	// Render this fun.Option in a group along with all other items with the same name
	group fun.Option[string]
	// help message attached to a completion item
	help fun.Option[string]
}

// to avoid leaking items with higher depth into items with lower depth
func (self Comp) depth() uint {
	var res uint
	self(
		func(extra CompExtra, _ ShortLong) { res = extra.depth },
		func(extra CompExtra, _ ShortLong, _ string) { res = extra.depth },
		func(extra CompExtra, _ string, _ fun.Option[byte]) { res = extra.depth },
		func(extra CompExtra, _ string, _ bool) { res = extra.depth },
		func(extra CompExtra, _ string, _ bool) { res = extra.depth },
		func(extra CompExtra, _ ShellComp, _ bool) { res = extra.depth },
	)
	return uint
}

// completer needs to replace meta placeholder with actual values - uses this
//
// value indicates if it's an argument or a positional meta
func (self Comp) is_metavar() Option[bool] {
	res := fun.Invalid[bool]()
	self(nil, nil, nil, nil, func(_ CompExtra, _ string, is_argument bool) { res = fun.Valid(*is_argument) }, nil)
	return res
}

func (self *Comp) set_group(group string) {
	var extraa CompExtra
	self(
		func(extra CompExtra, _ ShortLong) { extraa = extra.depth },
		func(extra CompExtra, _ ShortLong, _ string) { extraa = extra.depth },
		func(extra CompExtra, _ string, _ fun.Option[byte]) { extraa = extra.depth },
		func(extra CompExtra, _ string, _ bool) { extraa = extra.depth },
		func(extra CompExtra, _ string, _ bool) { extraa = extra.depth },
		func(extra CompExtra, _ ShellComp, _ bool) { extraa = extra.depth },
	)
	if extraa.group.is_fun.Invalid()() {
		extraa.group = fun.Valid(group)
	}
}

type ShowComp struct {
	// value to be actually inserted by the autocomplete system
	subst string

	// pretty rendering which might include metavars, etc
	pretty string

	extra CompExtra
}

func (self ShowComp) fmt(f *stdfmtFormatter) Result {
	if help, ok := self.extra.help.Unpack(); ok && self.subst.is_empty() {
		fmt.Fprintf(f, "{}: {}", self.pretty, help)
	} else if ok {
		fmt.Fprintf(f, "{:24} -- {}", self.pretty, help)
	} else {
		fmt.Fprintf(f, "{}", self.pretty)
	}
}

func (self Arg) and_os_string() Option[string] {
	res := fun.Invalid[string]()
	self.value(
		func(byte, bool) {
			if self.os_str.is_empty() {
				res = fun.Invalid()
			} else {
				res = fun.Valid(self.os_str)
			}
		},
		func(string, bool) { res = fun.Valid(self.os_str) },
		func() { res = fun.Valid(self.os_str) },
		func() { res = fun.Valid(self.os_str) },
		func() { res = fun.Valid(self.os_str) },
	)
	return res
}

// What is the preceeding item, if any
//
// Mostly is there to tell if we are trying to complete and argument or not...
type Prefix = func(
	NA func(),
	Short func(byte),
	Long func(string),
)

// Generate completion from collected heads
//
// before calling this method we run parser in "complete" mode and collect live heads inside
// `self.comp`, this part goes over collected heads and generates possible completions from
// that
func (self State) check_complete() Option[string] {
	comp, ok := self.comp_ref().Unpack()
	if !ok {
		return fun.Invalid[string]()
	}

	items := self.
		items.
		iter().
		rev().
		filter_map(Argand_os_string)

	// try get a current item to complete - must be non-virtual right most one
	// value must be present here, and can fail only for non-utf8 values
	// can't do much completing with non-utf8 values since bpaf needs to print them to stdout
	let(cur, lit, ok) = items.next().Unpack()
	if !ok {
		return fun.Invalid[string]()
	}

	// For cases like "-k=val", "-kval", "--key=val", "--key val"
	// last value is going  to be either ArgWord or ArgArgWord
	// so to perform full completion we look at the preceeding item
	// and use it's value if it was a composite short/long argument
	preceeding, ok := items.next()
	var (
		pos_only bool
		full_lit string
	)
	switch {
	// case ok && fun.Valid((ArgShort(_, true, _os) | ArgLong(_, true, _os), full_lit)):
	//     pos_only,full_lit=false, full_lit
	// case ok && (ArgPosWord(_), _): pos_only,full_lit=true, lit
	// default: pos_only,full_lit=false, lit
	}

	is_named := false
	cur(
		func(byte, bool) { is_named = true },
		func(string, bool) { is_named = true },
		nil,
		nil,
		nil,
	)

	var prefix Prefix
	// switch preceeding {
	//     case fun.Valid((ArgShort(s, true, _os), _lit)): prefix = PrefixShort(*s)
	//     case fun.Valid((ArgLong(l, true, _os), _lit)):prefix=PrefixLong(l.as_str())
	//     default:PrefixNA
	// };

	let(items, shell) = comp.complete(lit, pos_only, is_named, prefix)

	return fun.Valid(func() string {
		switch comp.output_rev {
		case 0:
			return render_test(&items, &shell, full_lit)
		case 1:
			return render_simple(&items) // <- AKA elvish
		case 7:
			return render_zsh(&items, &shell, full_lit)
		case 8:
			return render_bash(&items, &shell, full_lit)
		case 9:
			return render_fish(&items, &shell, full_lit, self.path[0].as_str())
		default:
			panic(fmt.Sprintf("Unsupported output revision {}, you need to genenerate your shell completion files for the app", unk))
		}
	})
}

// Try to expand short string names into long names if possible
func preferred_name(name ShortLong) string {
	var res string
	name(
		func(s byte) { res = fmt.Sprintf("-{}", s) },
		func(l string) { res = fmt.Sprintf("--{}", l) },
		func(_ byte, l string) { res = fmt.Sprintf("--{}", l) },
	)
	return res
}

// check if argument can possibly switch the argument passed in and returns a preferrable replacement
func arg_matches(arg string, name ShortLong) Option[string] {
	// "" and "-" switch any flag
	if arg.is_empty() || arg == "-" {
		return fun.Valid(preferred_name(name))
	}

	can_switch := false

	// separately check for short and long names, fancy strip prefix things is here to avoid
	// allocations and cloning
	name(
		func(s byte) {
			can_switch |= arg.
				strip_prefix('-').
				and_then((func(a) a.strip_prefix)(s)).
				map_or(false, stris_empty)
		},
		nil,
		func(s byte, _ string) {
			can_switch |= arg.
				strip_prefix('-').
				and_then((func(a) a.strip_prefix)(s)).
				map_or(false, stris_empty)
		},
	)

	// and long string too
	name(
		nil,
		func(l string) { can_switch |= arg.strip_prefix("--").map_or(false, (func(s) l.starts_with)(s)) },
		func(_ byte, l string) { can_switch |= arg.strip_prefix("--").map_or(false, (func(s) l.starts_with)(s)) },
	)

	if can_switch {
		return fun.Valid(preferred_name(name))
	} else {
		return fun.Invalid()
	}
}

func cmd_matches(arg string, name string, short fun.Option[char]) Option[string] {
	// partial long name and exact short name switch anything
	if name.starts_with(arg) || short.map_or(false, func(s) {
		// avoid allocations
		arg.strip_prefix(s).map_or(false, stris_empty)
	}) {
		return fun.Valid(name)
	} else {
		return fun.Invalid()
	}
}

// this completion should suppress anything else that is not a value
func (self Comp) only_value() bool {
	// switch self {
	//     CompFlag { .. } | CompArgument { .. } | CompCommand { .. }false,
	//     CompMetavariable { is_argument, .. }
	//     | CompValue { is_argument, .. }
	//     | CompShell { is_argument, .. }*is_argument,
	// }
}

func (self Comp) is_pos() bool {
	// switch self {
	//     CompFlag { .. } | CompArgument { .. } | CompCommand { .. }false,
	//     CompValue { is_argument, .. }!is_argument,
	//     CompMetavariable { .. } | CompShell { .. }true,
	// }
}

func (self Complete) complete(
	arg string,
	pos_only bool,
	is_named bool,
	prefix Prefix,
) ([]ShowComp, []ShellComp) {
	items := []ShowComp{}
	shell := Vecnew()
	max_depth := self.comps.iter().Map(Compdepth).max().unwrap_or(0)
	only_values := false

F:
	for _, item := range self.
		comps.
		iter().
		filter((func(c) c.depth)() == max_depth && (!pos_only || c.is_pos())) {
		switch {
		case only_values == item.only_value():
		case only_values:
			continue F
		case item.only_value():
			only_values = true
			items.clear()
		}

		item(
			func(extra CompExtra, name ShortLong) {
				if long, ok := arg_matches(arg, *name).Unpack(); ok {
					items.push(ShowComp{
						pretty: long,
						subst:  long,
						extra,
					})
				}
			},
			func(
				extra CompExtra,
				name ShortLong,
				metavar string,
			) {
				if long, ok := arg_matches(arg, *name).Unpack(); ok {
					items.push(ShowComp{
						pretty: fmt.Sprintf("{}={}", long, metavar),
						subst:  long,
						extra,
					})
				}
			},
			func(
				extra CompExtra,
				name string,
				short fun.Option[byte],
			) {
				if long, ok := cmd_matches(arg, name, *short).Unpack(); ok {
					items.push(ShowComp{
						subst:  long.to_string(),
						pretty: long.to_string(),
						extra,
					})
				}
			},
			func(
				extra CompExtra,
				body string,
				_ bool,
			) {
				items.push(ShowComp{
					pretty: body,
					extra,
					subst: func() string {
						var res string
						prefix(
							func() { res = body },
							func(s) { res = fmt.Sprintf("-{}={}", s, body) },
							func(l) { res = fmt.Sprintf("--{}={}", l, body) },
						)
						return res
					}(),
				})
			},
			func(
				extra CompExtra,
				meta string,
				is_argument bool,
			) {
				if !is_argument && !pos_only && arg.starts_with('-') {
					continue F
				}
				items.push(ShowComp{
					subst:  "",
					pretty: (*meta).to_string(),
					extra,
				})
			},
			func(
				_ CompExtra,
				script ShellComp,
				_ bool,
			) {
				if !is_named {
					shell.push(*script)
				}
			},
		)
	}

	return items, shell
}
