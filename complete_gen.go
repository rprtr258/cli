package cli

import (
	"fmt"
	"iter"
	"slices"
	"strings"

	"github.com/rprtr258/fun"
	funiter "github.com/rprtr258/fun/iter"
)

type CompExtra struct {
	depth int            // Used by complete_gen to separate commands from each other
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
		metavar Metavar,
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
		metavar Metavar,
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
		_ func(CompExtra, ShortLong, Metavar),
		_ func(CompExtra, string, Option[byte]),
		_ func(CompExtra, string, bool),
		_ func(CompExtra, Metavar, bool),
		_ func(CompExtra, ShellComp, bool),
	) {
		if Flag != nil {
			Flag(extra, name)
		}
	}
}

func CompArgument(extra CompExtra, name ShortLong, metavar Metavar) Comp {
	return func(
		_ func(CompExtra, ShortLong),
		Argument func(extra CompExtra, name ShortLong, metavar Metavar),
		_ func(CompExtra, string, Option[byte]),
		_ func(CompExtra, string, bool),
		_ func(CompExtra, Metavar, bool),
		_ func(CompExtra, ShellComp, bool),
	) {
		if Argument != nil {
			Argument(extra, name, metavar)
		}
	}
}

func CompCommand(a CompExtra, b string, c Option[byte]) Comp {
	return func(
		_ func(CompExtra, ShortLong),
		_ func(extra CompExtra, name ShortLong, metavar Metavar),
		f func(CompExtra, string, Option[byte]),
		_ func(CompExtra, string, bool),
		_ func(CompExtra, Metavar, bool),
		_ func(CompExtra, ShellComp, bool),
	) {
		if f != nil {
			f(a, b, c)
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
		_ func(CompExtra, ShortLong, Metavar),
		_ func(CompExtra, string, Option[byte]),
		Value func(extra CompExtra, body string, is_argument bool),
		_ func(CompExtra, Metavar, bool),
		_ func(CompExtra, ShellComp, bool),
	) {
		if Value != nil {
			Value(extra, body, is_argument)
		}
	}
}

func CompMetavariable(
	a CompExtra,
	metavar Metavar,
	c bool,
) Comp {
	return func(
		_ func(CompExtra, ShortLong),
		_ func(CompExtra, ShortLong, Metavar),
		_ func(CompExtra, string, Option[byte]),
		_ func(extra CompExtra, body string, is_argument bool),
		f func(CompExtra, Metavar, bool),
		_ func(CompExtra, ShellComp, bool),
	) {
		if f != nil {
			f(a, metavar, c)
		}
	}
}

func CompShell(
	a CompExtra,
	b ShellComp,
	c bool,
) Comp {
	return func(
		_ func(CompExtra, ShortLong),
		_ func(CompExtra, ShortLong, Metavar),
		_ func(CompExtra, string, Option[byte]),
		_ func(extra CompExtra, body string, is_argument bool),
		_ func(CompExtra, Metavar, bool),
		f func(CompExtra, ShellComp, bool),
	) {
		if f != nil {
			f(a, b, c)
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
	if comp, ok := self.comp_mut(); ok {
		if name, ok := ShortLongtry_from(named); ok {
			comp.comps = append(comp.comps, CompFlag(
				CompExtra{
					depth: depth,
					group: Invalid[string](),
					help:  OptFlatMap(named.help, (Doc).to_completion),
				},
				name,
			))
		}
	}
}

// Add a new completion hint for an argument, if needed
func (self *State) push_argument(named NamedArg, metavar Metavar) {
	depth := self.depth()
	if comp, ok := self.comp_mut(); ok {
		if name, ok := ShortLongtry_from(named); ok {
			comp.comps = append(comp.comps, CompArgument(
				CompExtra{
					depth: depth,
					group: Invalid[string](),
					help:  OptFlatMap(named.help, (Doc).to_completion),
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
	meta Metavar,
	help Option[Doc],
	is_argument bool,
) {
	depth := self.depth()
	if comp, ok := self.comp_mut(); ok {
		extra := CompExtra{
			depth: depth,
			group: Invalid[string](),
			help:  OptFlatMap(help, (Doc).to_completion),
		}

		comp.comps = append(comp.comps, CompMetavariable(extra, meta, is_argument))
	}
}

// Add a new completion hint for command, if needed
func (self *State) push_command(
	name string,
	short Option[byte],
	help Option[Doc],
) {
	depth := self.depth()
	if comp, ok := self.comp_mut(); ok {
		comp.comps = append(comp.comps, CompCommand(
			CompExtra{
				depth: depth,
				group: Invalid[string](),
				help:  OptFlatMap(help, (Doc).to_completion),
			},
			name,
			short,
		))
	}
}

// Clear collected completions if enabled
func (self *State) clear_comps() {
	if comp, ok := self.comp_mut(); ok {
		comp.comps = comp.comps[:0]
	}
}

// Insert a literal value with some description for completion
//
// In practice it's "--"
func (self *State) push_pos_sep() {
	depth := self.depth()
	if comp, ok := self.comp_mut(); ok {
		comp.comps = append(comp.comps, CompValue(
			CompExtra{
				depth: depth,
				group: Invalid[string](),
				help:  Valid("Positional only items after this token"),
			},
			"--",
			false,
		))
	}
}

// Insert a bunch of items
func (self *State) push_with_group(group Option[string], comps *[]Comp) {
	if comp, ok := self.comp_mut(); ok {
		for i := range *comps {
			if group, ok := group.Unpack(); ok {
				(*comps)[i].set_group(group)
			}
			comp.comps = append(comp.comps, (*comps)[i])
		}
	}
}

func (self *Complete) push_shell(op ShellComp, is_argument bool, depth int) {
	self.comps = append(self.comps, CompShell(
		CompExtra{
			depth: depth,
			group: Invalid[string](),
			help:  Invalid[string](),
		},
		op,
		is_argument,
	))
}

func (self *Complete) push_value(
	body string,
	help Option[string],
	group Option[string],
	depth int,
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

func (self *Complete) swap_comps(other *[]Comp) {
	*other, self.comps = self.comps, *other
}

// to avoid leaking items with higher depth into items with lower depth
func (self Comp) depth() int {
	var res int
	self(
		func(extra CompExtra, _ ShortLong) { res = extra.depth },
		func(extra CompExtra, _ ShortLong, _ Metavar) { res = extra.depth },
		func(extra CompExtra, _ string, _ Option[byte]) { res = extra.depth },
		func(extra CompExtra, _ string, _ bool) { res = extra.depth },
		func(extra CompExtra, _ Metavar, _ bool) { res = extra.depth },
		func(extra CompExtra, _ ShellComp, _ bool) { res = extra.depth },
	)
	return res
}

// completer needs to replace meta placeholder with actual values - uses this
//
// value indicates if it's an argument or a positional meta
func (self Comp) is_metavar() (bool, bool) {
	var res, ok bool
	self(nil, nil, nil, nil, func(_ CompExtra, _ Metavar, is_argument bool) { res, ok = is_argument, true }, nil)
	return res, ok
}

func (self *Comp) set_group(group string) {
	var extraa CompExtra
	(*self)(
		func(extra CompExtra, _ ShortLong) { extraa = extra },
		func(extra CompExtra, _ ShortLong, _ Metavar) { extraa = extra },
		func(extra CompExtra, _ string, _ Option[byte]) { extraa = extra },
		func(extra CompExtra, _ string, _ bool) { extraa = extra },
		func(extra CompExtra, _ Metavar, _ bool) { extraa = extra },
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
		fmt.Fprintf(f, "%v: %v", self.pretty, help)
	} else if ok {
		fmt.Fprintf(f, "%24s -- %v", self.pretty, help)
	} else {
		fmt.Fprintf(f, "%v", self.pretty)
	}
}

func (self ShowComp) to_string() string {
	var res strings.Builder
	self.fmt(&res)
	return res.String()
}

func (self Arg) and_os_string() Option[string] {
	switch self.kind {
	case ArgKindShortFlag:
		if self.str != "" {
			return Valid(self.str)
		}
	default:
		return Valid(self.str)
	}
	return Invalid[string]()
}

// What is the preceeding item, if any
//
// Mostly is there to tell if we are trying to complete and argument or not...
type Prefix = func(
	NA func(),
	Short func(byte),
	Long func(string),
)

func PrefixNA(f func(), _ func(byte), _ func(string)) { f() }
func PrefixShort(s byte) Prefix {
	return func(_ func(), f func(byte), _ func(string)) {
		if f != nil {
			f(s)
		}
	}
}
func PrefixLong(s string) Prefix {
	return func(_ func(), _ func(byte), f func(string)) {
		if f != nil {
			f(s)
		}
	}
}

// Generate completion from collected heads
// before calling this method we run parser in "complete" mode and collect live heads inside
// `self.comp`, this part goes over collected heads and generates possible completions from that
func (self State) check_complete() (string, bool) {
	comp, ok := self.comp.Unpack()
	if !ok {
		return "", false
	}

	next, cancel := iter.Pull2(func(yield func(Arg, string) bool) {
		for _, s := range slices.Backward(self.items) {
			if s.str != "" && !yield(s, s.str) {
				break
			}
		}
	})
	defer cancel()

	// try get a current item to complete - must be non-virtual right most one
	// value must be present here, and can fail only for non-utf8 values
	// can't do much completing with non-utf8 values since bpaf needs to print them to stdout
	cur, lit, ok := next()
	if !ok {
		return "", false
	}

	// For cases like "-k=val", "-kval", "--key=val", "--key val"
	// last value is going  to be either ArgWord or ArgArgWord
	// so to perform full completion we look at the preceeding item
	// and use it's value if it was a composite short/long argument
	preceeding, _, ok := next()
	var (
		pos_only bool
		full_lit string
	)
	switch {
	// case ok && Valid((ArgShort(_, true, _os) | ArgLong(_, true, _os), full_lit)):
	//     pos_only,full_lit=false, full_lit
	// case ok && (ArgPositional(_), _): pos_only,full_lit=true, lit
	// default: pos_only,full_lit=false, lit
	}

	is_named := fun.Contains(cur.kind, ArgKindShortFlag, ArgKindLongFlag)

	prefix := PrefixNA
	switch preceeding.kind {
	case ArgKindShortFlag:
		if preceeding.is_adj {
			prefix = PrefixShort(preceeding.c)
		}
	case ArgKindLongFlag:
		if preceeding.is_adj {
			prefix = PrefixLong(preceeding.s)
		}
	}

	items, shell := comp.complete(lit, pos_only, is_named, prefix)

	return func() string {
		switch comp.output_rev {
		case 0:
			return render_test(items, shell, full_lit)
		case 1:
			return render_simple(items) // <- AKA elvish
		case 7:
			return render_zsh(items, shell, full_lit)
		case 8:
			return render_bash(items, shell, full_lit)
		case 9:
			return render_fish(items, shell, full_lit, self.path[0])
		default:
			panic(fmt.Sprintf("Unsupported output revision %v, you need to genenerate your shell completion files for the app", comp.output_rev))
		}
	}(), true
}

// Try to expand short string names into long names if possible
func preferred_name(name ShortLong) string {
	if name.is_short() {
		return fmt.Sprintf("-%c", name.Short)
	} else {
		return fmt.Sprintf("--%s", name.Long)
	}
}

func stris_empty(s string) bool {
	return s == ""
}

// check if argument can possibly switch the argument passed in and returns a preferrable replacement
func arg_matches(arg string, name ShortLong) (string, bool) {
	// "" and "-" switch any flag
	if arg == "" || arg == "-" ||
		// separately check for short and long names, fancy strip prefix things is here to avoid allocations and cloning
		map_or(strip_prefixb(arg, '-').
			and_then(func(a string) Option[string] { return strip_prefixb(a, name.Short) }),
			false, stris_empty) ||
		// and long string too
		map_or(strip_prefix(arg, "--"), false, func(s string) bool { return strings.HasPrefix(name.Long, s) }) {
		return preferred_name(name), true
	}

	return "", false
}

func cmd_matches(arg string, name string, short Option[byte]) (string, bool) {
	// partial long name and exact short name switch anything
	if strings.HasPrefix(name, arg) || map_or(short, false, func(s byte) bool {
		// avoid allocations
		return map_or(strip_prefixb(arg, s), false, stris_empty)
	}) {
		return name, true
	}

	return "", false
}

// this completion should suppress anything else that is not a value
func (self Comp) only_value() bool {
	var res bool
	self(
		nil,
		nil,
		nil,
		func(_ CompExtra, _ string, is_argument bool) { res = is_argument },
		func(_ CompExtra, _ Metavar, is_argument bool) { res = is_argument },
		func(_ CompExtra, _ ShellComp, is_argument bool) { res = is_argument },
	)
	return res
}

func (self Comp) is_pos() bool {
	var res bool
	self(
		nil,
		nil,
		nil,
		func(_ CompExtra, _ string, is_argument bool) { res = !is_argument },
		func(_ CompExtra, _ Metavar, is_argument bool) { res = true },
		func(_ CompExtra, _ ShellComp, is_argument bool) { res = true },
	)
	return res
}

func (self Complete) complete(
	arg string,
	pos_only bool,
	is_named bool,
	prefix Prefix,
) ([]ShowComp, []ShellComp) {
	max_depth := 0
	for _, c := range self.comps {
		max_depth = max(max_depth, c.depth())
	}

	only_values := false
	items := []ShowComp{}
	shell := []ShellComp{}
F:
	for item := range funiter.
		FromMany(self.comps...).
		Filter(func(c Comp) bool { return c.depth() == max_depth && (!pos_only || c.is_pos()) }) {
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
				if long, ok := arg_matches(arg, name); ok {
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
				metavar Metavar,
			) {
				if long, ok := arg_matches(arg, name); ok {
					items = append(items, ShowComp{
						pretty: fmt.Sprintf("%v=%v", long, metavar),
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
				if long, ok := cmd_matches(arg, name, short); ok {
					items = append(items, ShowComp{
						subst:  long,
						pretty: long,
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
							func(s byte) { res = fmt.Sprintf("-%v=%v", s, body) },
							func(l string) { res = fmt.Sprintf("--%v=%v", l, body) },
						)
						return res
					}(),
				})
			},
			func(
				extra CompExtra,
				meta Metavar,
				is_argument bool,
			) {
				if is_argument || pos_only || !starts_withb(arg, '-') {
					items = append(items, ShowComp{
						subst:  "",
						pretty: string(meta),
						extra:  extra,
					})
				}
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
