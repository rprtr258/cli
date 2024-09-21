package cli

import (
	"fmt"
	"strings"

	"github.com/rprtr258/fun"
)

type CompExtra struct {
	depth uint           // Used by complete_gen to separate commands from each other
	group Option[string] // Render this fun.Option in a group along with all other items with the same name
	help  Option[string] // help message attached to a completion item
}

// TODO: move out extra
type Comp func(
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
		short Option[byte],
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

func CompFlag(extra CompExtra, name ShortLong) Comp {
	return func(
		Flag func(extra CompExtra, name ShortLong),
		_ func(CompExtra, ShortLong, string),
		_ func(CompExtra, string, Option[byte]),
		_ func(CompExtra, string, bool),
		_ func(CompExtra, string, bool),
		_ func(CompExtra, ShellComp, bool),
	) {
		if Flag != nil {
			Flag(extra, name)
		}
	}
}

func CompArgument(extra CompExtra, name ShortLong, metavar string) Comp {
	return func(
		_ func(CompExtra, ShortLong),
		Argument func(extra CompExtra, name ShortLong, metavar string),
		_ func(CompExtra, string, Option[byte]),
		_ func(CompExtra, string, bool),
		_ func(CompExtra, string, bool),
		_ func(CompExtra, ShellComp, bool),
	) {
		if Argument != nil {
			Argument(extra, name, metavar)
		}
	}
}

func CompValue(
	extra CompExtra,
	body string,
	is_argument bool,
) Comp {
	return func(
		_ func(CompExtra, ShortLong),
		_ func(CompExtra, ShortLong, string),
		_ func(CompExtra, string, Option[byte]),
		Value func(extra CompExtra, body string, is_argument bool),
		_ func(CompExtra, string, bool),
		_ func(CompExtra, ShellComp, bool),
	) {
		if Value != nil {
			Value(extra, body, is_argument)
		}
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
		comps:        []Comp{},
		output_rev:   output_rev,
		no_pos_ahead: false,
	}
}

// Add a new completion hint for flag, if needed
func (self *State) push_flag(named NamedArg) {
	depth := self.depth()
	if comp, ok := self.comp_mut().Unpack(); ok {
		if name, err := ShortLongtry_from(named); err == nil {
			comp.comps = append(comp.comps, CompFlag(
				CompExtra{
					depth: depth,
					group: Invalid[string](),
					help:  named.help.and_then(Docto_completion),
				},
				name,
			))
		}
	}
}

// Add a new completion hint for an argument, if needed
func (self *State) push_argument(named NamedArg, metavar string) {
	depth := self.depth()
	if comp, ok := self.comp_mut().Unpack(); ok {
		if name, err := ShortLongtry_from(named); err == nil {
			comp.comps = append(comp.comps, CompArgument(
				CompExtra{
					depth: depth,
					group: Invalid[string](),
					help:  named.help.and_then(Docto_completion),
				},
				name,
				metavar,
			))
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
			depth: depth,
			group: Invalid[string](),
			help:  help.and_then(Docto_completion),
		}

		comp.comps = append(comp.comps, CompMetavariable{
			extra,
			meta,
			is_argument,
		})
	}
}

// Add a new completion hint for command, if needed
func (self *State) push_command(
	name string,
	short Option[byte],
	help Option[Doc],
) {
	depth := self.depth()
	if comp, ok := self.comp_mut().Unpack(); ok {
		comp.comps = append(comp.comps, CompCommand{
			extra: CompExtra{
				depth: depth,
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
		comp.comps = comp.comps[:0]
	}
}

// Insert a literal value with some description for completion
//
// In practice it's "--"
func (self *State) push_pos_sep() {
	depth := self.depth()
	if comp, ok := self.comp_mut().Unpack(); ok {
		comp.comps = append(comp.comps, CompValue{
			extra: CompExtra{
				depth: depth,
				group: Invalid[string](),
				help:  Valid("Positional only items after this token"),
			},
			body:        "--",
			is_argument: false,
		})
	}
}

// Insert a bunch of items
func (self *State) push_with_group(group Option[string], comps *[]Comp) {
	if comp, ok := self.comp_mut().Unpack(); ok {
		for i := range *comps {
			if group, ok := group.Unpack(); ok {
				(*comps)[i].set_group(group)
			}
			comp.comps = append(comp.comps, (*comps)[i])
		}
	}
}

func (self *Complete) push_shell(op ShellComp, is_argument bool, depth uint) {
	self.comps = append(self.comps, CompShell{
		extra: CompExtra{
			depth: depth,
			group: Invalid[string](),
			help:  Invalid[string](),
		},
		script: op,
		is_argument,
	})
}

func (self *Complete) push_value(
	body string,
	help Option[string],
	group Option[string],
	depth uint,
	is_argument bool,
) {
	self.comps = append(self.comps, CompValue(
		CompExtra{depth, group, help},
		body,
		is_argument,
	))
}

func (self *Complete) push_comp(comp Comp) {
	self.comps = append(self.comps, comp)
}

func (self *Complete) extend_comps(comps []Comp) {
	self.comps = append(self.comps, comps...)
}

func (self *Complete) drain_comps() []Comp {
	return self.comps
}

func (self *Complete) swap_comps(other *[]Comp) {
	stdmemswap(other, &self.comps)
}

// to avoid leaking items with higher depth into items with lower depth
func (self Comp) depth() uint {
	var res uint
	self(
		func(extra CompExtra, _ ShortLong) { res = extra.depth },
		func(extra CompExtra, _ ShortLong, _ string) { res = extra.depth },
		func(extra CompExtra, _ string, _ Option[byte]) { res = extra.depth },
		func(extra CompExtra, _ string, _ bool) { res = extra.depth },
		func(extra CompExtra, _ string, _ bool) { res = extra.depth },
		func(extra CompExtra, _ ShellComp, _ bool) { res = extra.depth },
	)
	return res
}

// completer needs to replace meta placeholder with actual values - uses this
//
// value indicates if it's an argument or a positional meta
func (self Comp) is_metavar() Option[bool] {
	res := Invalid[bool]()
	self(nil, nil, nil, nil, func(_ CompExtra, _ string, is_argument bool) { res = Valid(is_argument) }, nil)
	return res
}

func (self *Comp) set_group(group string) {
	var extraa CompExtra
	(*self)(
		func(extra CompExtra, _ ShortLong) { extraa = extra },
		func(extra CompExtra, _ ShortLong, _ string) { extraa = extra },
		func(extra CompExtra, _ string, _ Option[byte]) { extraa = extra },
		func(extra CompExtra, _ string, _ bool) { extraa = extra },
		func(extra CompExtra, _ string, _ bool) { extraa = extra },
		func(extra CompExtra, _ ShellComp, _ bool) { extraa = extra },
	)
	if !extraa.group.Valid {
		extraa.group = Valid(group)
	}
}

type ShowComp struct {
	// value to be actually inserted by the autocomplete system
	subst string

	// pretty rendering which might include metavars, etc
	pretty string

	extra CompExtra
}

func (self ShowComp) fmt(f *strings.Builder) {
	if help, ok := self.extra.help.Unpack(); ok && self.subst == "" {
		fmt.Fprintf(f, "{}: {}", self.pretty, help)
	} else if ok {
		fmt.Fprintf(f, "{:24} -- {}", self.pretty, help)
	} else {
		fmt.Fprintf(f, "{}", self.pretty)
	}
}

func (self Arg) and_os_string() Option[string] {
	res := Invalid[string]()
	self.value(
		func(byte, bool) {
			if self.os_str != "" {
				res = Valid(self.os_str)
			}
		},
		func(string, bool) { res = Valid(self.os_str) },
		func() { res = Valid(self.os_str) },
		func() { res = Valid(self.os_str) },
		func() { res = Valid(self.os_str) },
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
		return Invalid[string]()
	}

	items := self.
		items.
		rev().
		filter_map(Argand_os_string)

	// try get a current item to complete - must be non-virtual right most one
	// value must be present here, and can fail only for non-utf8 values
	// can't do much completing with non-utf8 values since bpaf needs to print them to stdout
	cur, lit, ok := items.next().Unpack()
	if !ok {
		return Invalid[string]()
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
	// case ok && Valid((ArgShort(_, true, _os) | ArgLong(_, true, _os), full_lit)):
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
	//     case Valid((ArgShort(s, true, _os), _lit)): prefix = PrefixShort(*s)
	//     case Valid((ArgLong(l, true, _os), _lit)):prefix=PrefixLong(l)
	//     default:PrefixNA
	// };

	items, shell := comp.complete(lit, pos_only, is_named, prefix)

	return Valid(func() string {
		switch comp.output_rev {
		case 0:
			return render_test(&items, shell, full_lit)
		case 1:
			return render_simple(&items) // <- AKA elvish
		case 7:
			return render_zsh(&items, shell, full_lit)
		case 8:
			return render_bash(&items, shell, full_lit)
		case 9:
			return render_fish(&items, shell, full_lit, self.path[0])
		default:
			panic(fmt.Sprintf("Unsupported output revision {}, you need to genenerate your shell completion files for the app", unk))
		}
	}())
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
	if arg == "" || arg == "-" {
		return Valid(preferred_name(name))
	}

	can_switch := false

	// separately check for short and long names, fancy strip prefix things is here to avoid
	// allocations and cloning
	name(
		func(s byte) {
			can_switch |= arg.
				strip_prefix('-').
				and_then(func(a string) Option[string] { return a.strip_prefix(s) }).
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

	if !can_switch {
		return Invalid[string]()
	}

	return Valid(preferred_name(name))
}

func cmd_matches(arg string, name string, short Option[byte]) Option[string] {
	// partial long name and exact short name switch anything
	if name.starts_with(arg) || short.map_or(false, func(s string) Option[string] {
		// avoid allocations
		return arg.strip_prefix(s).map_or(false, stris_empty)
	}) {
		return Valid(name)
	}

	return Invalid[string]()
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
	max_depth := uint(0)
	for _, c := range self.comps {
		max_depth = max(max_depth, c.depth())
	}

	only_values := false
	items := []ShowComp{}
	shell := []ShellComp{}
F:
	for _, item := range self.
		comps.
		iter().
		filter(func(c Comp) bool { return c.depth() == max_depth && (!pos_only || c.is_pos()) }) {
		switch {
		case only_values == item.only_value():
		case only_values:
			continue F
		case item.only_value():
			only_values = true
			items = items[:0]
		}

		item(
			func(extra CompExtra, name ShortLong) {
				if long, ok := arg_matches(arg, *name).Unpack(); ok {
					items = append(items, ShowComp{
						pretty: long,
						subst:  long,
						extra:  extra,
					})
				}
			},
			func(
				extra CompExtra,
				name ShortLong,
				metavar string,
			) {
				if long, ok := arg_matches(arg, *name).Unpack(); ok {
					items = append(items, ShowComp{
						pretty: fmt.Sprintf("{}={}", long, metavar),
						subst:  long,
						extra:  extra,
					})
				}
			},
			func(
				extra CompExtra,
				name string,
				short Option[byte],
			) {
				if long, ok := cmd_matches(arg, name, *short).Unpack(); ok {
					items = append(items, ShowComp{
						subst:  long.to_string(),
						pretty: long.to_string(),
						extra:  extra,
					})
				}
			},
			func(
				extra CompExtra,
				body string,
				_ bool,
			) {
				items = append(items, ShowComp{
					pretty: body,
					extra:  extra,
					subst: func() string {
						var res string
						prefix(
							func() { res = body },
							func(s byte) { res = fmt.Sprintf("-{}={}", s, body) },
							func(l string) { res = fmt.Sprintf("--{}={}", l, body) },
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
				items = append(items, ShowComp{
					subst:  "",
					pretty: meta,
					extra:  extra,
				})
			},
			func(
				_ CompExtra,
				script ShellComp,
				_ bool,
			) {
				if !is_named {
					shell = append(shell, script)
				}
			},
		)
	}

	return items, shell
}
