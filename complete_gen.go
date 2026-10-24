package cli

import (
	"fmt"
	"iter"
	"slices"
	"strings"

	"github.com/rprtr258/fun"
)

type compExtra struct {
	depth int         // Used by complete_gen to separate commands from each other
	group Opt[string] // Render this option in a group along with all other items with the same name
	help  Opt[string] // help message attached to a completion item
}

type compKind uint8

const (
	// short or long flag
	compKindFlag compKind = iota
	// argument + metadata
	compKindArgument
	compKindCommand
	// comes from completed values, part of "dynamic" completion
	compKindValue
	compKindMetavariable
	compKindShell
)

// comp is a single completion hint.
type comp struct {
	kind compKind
	// shared metadata
	extra compExtra
	// flag or argument name
	name shortLong
	// argument or metavariable placeholder
	metavar Metavar
	// command name or value body
	body string
	// command short name
	short Opt[byte]
	// value/metavariable/shell: true when completing an argument value rather than a positional
	is_argument bool
	// shell completion operation
	script ShellComp
}

func compFlag(extra compExtra, name shortLong) comp {
	return comp{kind: compKindFlag, extra: extra, name: name}
}

func compArgument(extra compExtra, name shortLong, metavar Metavar) comp {
	return comp{kind: compKindArgument, extra: extra, name: name, metavar: metavar}
}

func compCommand(extra compExtra, body string, short Opt[byte]) comp {
	return comp{kind: compKindCommand, extra: extra, body: body, short: short}
}

func compValue(extra compExtra, body string, is_argument bool) comp {
	return comp{kind: compKindValue, extra: extra, body: body, is_argument: is_argument}
}

func compMetavariable(extra compExtra, metavar Metavar, is_argument bool) comp {
	return comp{kind: compKindMetavariable, extra: extra, metavar: metavar, is_argument: is_argument}
}

func compShell(extra compExtra, script ShellComp, is_argument bool) comp {
	return comp{kind: compKindShell, extra: extra, script: script, is_argument: is_argument}
}

// Complete implements the completion protocol:
//
//	static: flag names, command names
//	dynamic: argument values, positional item values
//
// For static completion, running collect only fails if any parser fails.
//
//	OR: combine completions
//	AND: usual logic without shortcircuits
//
// For static completion it's enough to collect items.
// For argument completion - only one argument (compMeta) should be active at once.
//
// For rendering prefer longer version of names.
// Complete short names to long names if possible.

type Complete struct {
	// completions accumulated so far
	comps      []comp
	output_rev uint

	// don't try to suggest any more positional items after there's a positional item failure
	// or parsing in progress
	no_pos_ahead bool
}

func newComplete(output_rev uint) Complete {
	return Complete{
		comps:        []comp{},
		output_rev:   output_rev,
		no_pos_ahead: false,
	}
}

// push_flag adds a new completion hint for the flag, if needed.
func (self *state) push_flag(named NamedArg) {
	depth := self.depth()
	if comp, ok := self.comp_mut(); ok {
		if name, ok := shortLongTryFrom(named); ok {
			comp.comps = append(comp.comps, compFlag(
				compExtra{
					depth: depth,
					group: Invalid[string](),
					help:  named.help.FlatMap(Doc.to_completion),
				},
				name,
			))
		}
	}
}

// push_argument adds a new completion hint for an argument, if needed.
func (self *state) push_argument(named NamedArg, metavar Metavar) {
	depth := self.depth()
	if comp, ok := self.comp_mut(); ok {
		if name, ok := shortLongTryFrom(named); ok {
			comp.comps = append(comp.comps, compArgument(
				compExtra{
					depth: depth,
					group: Invalid[string](),
					help:  named.help.FlatMap(Doc.to_completion),
				},
				name,
				metavar,
			))
		}
	}
}

// push_metavar adds a new completion hint for metadata, if needed.
//
// `is_argument` is set to true when we are trying to parse the value and false if
// when meta
func (self *state) push_metavar(
	meta Metavar,
	help Opt[Doc],
	is_argument bool,
) {
	depth := self.depth()
	if comp, ok := self.comp_mut(); ok {
		extra := compExtra{
			depth: depth,
			group: Invalid[string](),
			help:  help.FlatMap(Doc.to_completion),
		}

		comp.comps = append(comp.comps, compMetavariable(extra, meta, is_argument))
	}
}

// push_command adds a new completion hint for the command, if needed.
func (self *state) push_command(
	name string,
	short Opt[byte],
	help Opt[Doc],
) {
	depth := self.depth()
	if comp, ok := self.comp_mut(); ok {
		comp.comps = append(comp.comps, compCommand(
			compExtra{
				depth: depth,
				group: Invalid[string](),
				help:  help.FlatMap(Doc.to_completion),
			},
			name,
			short,
		))
	}
}

// clear_comps clears collected completions if completion is enabled.
func (self *state) clear_comps() {
	if comp, ok := self.comp_mut(); ok {
		comp.comps = comp.comps[:0]
	}
}

// push_pos_sep inserts a literal value with some description for completion.
//
// In practice it's "--"
func (self *state) push_pos_sep() {
	depth := self.depth()
	if comp, ok := self.comp_mut(); ok {
		comp.comps = append(comp.comps, compValue(
			compExtra{
				depth: depth,
				group: Invalid[string](),
				help:  Valid("Positional only items after this token"),
			},
			"--",
			false,
		))
	}
}

// push_with_group inserts a group of items.
func (self *state) push_with_group(group Opt[string], comps *[]comp) {
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
	self.comps = append(self.comps, compShell(
		compExtra{
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
	help Opt[string],
	group Opt[string],
	depth int,
	is_argument bool,
) {
	self.comps = append(self.comps, compValue(
		compExtra{depth, group, help},
		body,
		is_argument,
	))
}

func (self *Complete) push_comp(c *comp) {
	self.comps = append(self.comps, *c)
}

func (self *Complete) extend_comps(comps []comp) {
	self.comps = append(self.comps, comps...)
}

// depth returns the completion depth, used to avoid leaking items with
// higher depth into items with lower depth.
func (self *comp) depth() int {
	return self.extra.depth
}

// is_metavar reports whether the comp is a metavariable placeholder and
// whether it stands for an argument or a positional meta.
func (self *comp) is_metavar() (is_argument, ok bool) {
	if self.kind == compKindMetavariable {
		return self.is_argument, true
	}
	return false, false
}

func (self *comp) set_group(group string) {
	if !self.extra.group.Valid {
		self.extra.group = Valid(group)
	}
}

type ShowComp struct {
	// value to be actually inserted by the autocomplete system
	subst string

	// pretty rendering which might include metavars, etc
	pretty string

	extra compExtra
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

// prefix inspects the preceding item, if any - mostly to tell whether an
// argument rather than a flag is being completed.
type prefix = func(
	NA func(),
	Short func(byte),
	Long func(string),
)

func prefixNA(f func(), _ func(byte), _ func(string)) { f() }
func prefixShort(s byte) prefix {
	return func(_ func(), f func(byte), _ func(string)) {
		if f != nil {
			f(s)
		}
	}
}

func prefixLong(s string) prefix {
	return func(_ func(), _ func(byte), f func(string)) {
		if f != nil {
			f(s)
		}
	}
}

// check_complete generates completion from collected heads.
//
// Before calling this method the parser runs in "complete" mode and collects
// live heads inside self.comp; this part goes over the collected heads and
// generates possible completions from that.
func (self *state) check_complete() (string, bool) {
	comp, ok := self.comp.Unpack()
	if !ok {
		return "", false
	}

	next, cancel := iter.Pull2(func(yield func(arg, string) bool) {
		for _, s := range slices.Backward(self.items) {
			// empty string is only expected for short flags in a block like -abc
			if s.kind == argKindShortFlag && s.str == "" {
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
	// can't do much completing with non-utf8 values since they need to be printed to stdout
	cur, lit, ok := next()
	if !ok {
		return "", false
	}

	// For cases like "-k=val", "-kval", "--key=val", "--key val"
	// last value is going  to be either argWord or argArgWord
	// so to perform full completion we look at the preceeding item
	// and use it's value if it was a composite short/long argument
	preceeding, preceeding_os, ok := next()
	var (
		pos_only bool
		full_lit string
	)
	switch {
	case ok && fun.Contains(preceeding.kind, argKindShortFlag, argKindLongFlag) && preceeding.isAdjacent:
		pos_only, full_lit = false, preceeding_os
	case ok && preceeding.kind == argKindPositional:
		pos_only, full_lit = true, lit
	default:
		pos_only, full_lit = false, lit
	}

	is_named := fun.Contains(cur.kind, argKindShortFlag, argKindLongFlag)

	prefix := prefixNA
	switch preceeding.kind {
	case argKindShortFlag:
		if preceeding.isAdjacent {
			prefix = prefixShort(preceeding.c)
		}
	case argKindLongFlag:
		if preceeding.isAdjacent {
			prefix = prefixLong(preceeding.s)
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

// preferred_name expands short string names into long names if possible.
func preferred_name(name shortLong) string {
	if name.is_short() {
		return fmt.Sprintf("-%c", name.Short)
	} else {
		return fmt.Sprintf("--%s", name.Long)
	}
}

func stris_empty(s string) bool {
	return s == ""
}

// arg_matches checks whether the argument can possibly switch the argument
// passed in and returns a preferable replacement.
func arg_matches(arg string, name shortLong) (string, bool) {
	// "" and "-" switch any flag
	if arg == "" || arg == "-" ||
		// separately check for short and long names, fancy strip prefix things is here to avoid allocations and cloning
		strip_prefixb(arg, '-').
			FlatMap(func(a string) Opt[string] { return strip_prefixb(a, name.Short) }).
			MapOr(false, stris_empty) ||
		// and long string too, name must actually have a long variant
		strip_prefix(arg, "--").MapOr(false, func(s string) bool {
			return name.Long != "" && strings.HasPrefix(name.Long, s)
		}) {
		return preferred_name(name), true
	}

	return "", false
}

func cmd_matches(arg, name string, short Opt[byte]) (string, bool) {
	// partial long name and exact short name switch anything
	if strings.HasPrefix(name, arg) || short.MapOr(false, func(s byte) bool {
		// avoid allocations
		return strip_prefixb(arg, s).MapOr(false, stris_empty)
	}) {
		return name, true
	}

	return "", false
}

// only_value reports whether this completion should suppress anything else
// that is not a value.
func (self *comp) only_value() bool {
	switch self.kind {
	case compKindValue, compKindMetavariable, compKindShell:
		return self.is_argument
	default:
		return false
	}
}

func (self comp) is_pos() bool {
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
	prefix prefix,
) ([]ShowComp, []ShellComp) {
	max_depth := 0
	for i := range self.comps {
		max_depth = max(max_depth, self.comps[i].depth())
	}

	only_values := false
	items := []ShowComp{}
	shell := []ShellComp{}
F:
	for i := range self.comps {
		c := &self.comps[i]
		if c.depth() != max_depth || (pos_only && !c.is_pos()) {
			continue F
		}
		item := c
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
