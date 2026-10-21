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

type CompKind uint8

const (
	// short or long flag
	compKindFlag CompKind = iota
	// argument + metadata
	compKindArgument
	compKindCommand
	// comes from completed values, part of "dynamic" completion
	compKindValue
	compKindMetavariable
	compKindShell
)

// A single completion hint, mirrors the Rust `Comp` enum
type Comp struct {
	kind CompKind
	// shared metadata
	extra CompExtra
	// flag or argument name
	name ShortLong
	// argument or metavariable placeholder
	metavar Metavar
	// command name or value body
	body string
	// command short name
	short Option[byte]
	// value/metavariable/shell: true when completing an argument value rather than a positional
	is_argument bool
	// shell completion operation
	script ShellComp
}

func CompFlag(extra CompExtra, name ShortLong) Comp {
	return Comp{kind: compKindFlag, extra: extra, name: name}
}

func CompArgument(extra CompExtra, name ShortLong, metavar Metavar) Comp {
	return Comp{kind: compKindArgument, extra: extra, name: name, metavar: metavar}
}

func CompCommand(extra CompExtra, body string, short Option[byte]) Comp {
	return Comp{kind: compKindCommand, extra: extra, body: body, short: short}
}

func CompValue(extra CompExtra, body string, is_argument bool) Comp {
	return Comp{kind: compKindValue, extra: extra, body: body, is_argument: is_argument}
}

func CompMetavariable(extra CompExtra, metavar Metavar, is_argument bool) Comp {
	return Comp{kind: compKindMetavariable, extra: extra, metavar: metavar, is_argument: is_argument}
}

func CompShell(extra CompExtra, script ShellComp, is_argument bool) Comp {
	return Comp{kind: compKindShell, extra: extra, script: script, is_argument: is_argument}
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

// to avoid leaking items with higher depth into items with lower depth
func (self Comp) depth() int {
	return self.extra.depth
}

// completer needs to replace meta placeholder with actual values - uses this
//
// value indicates if it's an argument or a positional meta
func (self Comp) is_metavar() (bool, bool) {
	if self.kind == compKindMetavariable {
		return self.is_argument, true
	}
	return false, false
}

func (self *Comp) set_group(group string) {
	if !self.extra.group.Valid {
		self.extra.group = Valid(group)
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
			// empty string is only expected for short flags in a block like -abc
			if s.kind == ArgKindShortFlag && s.str == "" {
				continue
			}
			if !yield(s, s.str) {
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
	preceeding, preceeding_os, ok := next()
	var (
		pos_only bool
		full_lit string
	)
	switch {
	case ok && fun.Contains(preceeding.kind, ArgKindShortFlag, ArgKindLongFlag) && preceeding.isAdjacent:
		pos_only, full_lit = false, preceeding_os
	case ok && preceeding.kind == ArgKindPositional:
		pos_only, full_lit = true, lit
	default:
		pos_only, full_lit = false, lit
	}

	is_named := fun.Contains(cur.kind, ArgKindShortFlag, ArgKindLongFlag)

	prefix := PrefixNA
	switch preceeding.kind {
	case ArgKindShortFlag:
		if preceeding.isAdjacent {
			prefix = PrefixShort(preceeding.c)
		}
	case ArgKindLongFlag:
		if preceeding.isAdjacent {
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
		// and long string too, name must actually have a long variant
		map_or(strip_prefix(arg, "--"), false, func(s string) bool {
			return name.Long != "" && strings.HasPrefix(name.Long, s)
		}) {
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
	switch self.kind {
	case compKindValue, compKindMetavariable, compKindShell:
		return self.is_argument
	default:
		return false
	}
}

func (self Comp) is_pos() bool {
	switch self.kind {
	case compKindValue:
		return !self.is_argument
	case compKindMetavariable, compKindShell:
		return true
	default:
		return false
	}
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

		switch item.kind {
		case compKindFlag:
			if long, ok := arg_matches(arg, item.name); ok {
				items = append(items, ShowComp{
					pretty: long,
					subst:  long,
					extra:  item.extra,
				})
			}
		case compKindArgument:
			if long, ok := arg_matches(arg, item.name); ok {
				items = append(items, ShowComp{
					pretty: fmt.Sprintf("%v=%v", long, item.metavar),
					subst:  long,
					extra:  item.extra,
				})
			}
		case compKindCommand:
			if long, ok := cmd_matches(arg, item.body, item.short); ok {
				items = append(items, ShowComp{
					subst:  long,
					pretty: long,
					extra:  item.extra,
				})
			}
		case compKindValue:
			items = append(items, ShowComp{
				pretty: item.body,
				extra:  item.extra,
				subst: func() string {
					var res string
					prefix(
						func() { res = item.body },
						func(s byte) { res = fmt.Sprintf("-%v=%v", s, item.body) },
						func(l string) { res = fmt.Sprintf("--%v=%v", l, item.body) },
					)
					return res
				}(),
			})
		case compKindMetavariable:
			if item.is_argument || pos_only || !starts_withb(arg, '-') {
				items = append(items, ShowComp{
					subst:  "",
					pretty: string(item.metavar),
					extra:  item.extra,
				})
			}
		case compKindShell:
			if !is_named {
				shell = append(shell, item.script)
			}
		}
	}

	return items, shell
}
