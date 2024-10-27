package cli

import (
	"bytes"
	"math"
	"os"
	"path/filepath"
	"unicode/utf8"
)

// All currently present command line parameters with fun.Valid extra metainfo
// /
// Use it for unit tests and manual parsing. For production use you would want to replace the
// program name with [`set_name`](Argsset_name), but for tests passing a slice of strings to
// [`run_inner`](crateOptionParserrun_inner) is usually more convenient.
// /
// /
// The easiest way to create `Args` is by using its `From` instance.
// ```rust
// # use bpaf*;
// parser := short('f')
//
//	.switch()
//	.to_options();
//
// value := parser
//
//	.run_inner(Argsfrom(&["-f"]))
//	.unwrap();
//
// require.True(t,value);
// /
// // this also works
// value := parser.run_inner(&["-f"])
//
//	.unwrap();
//
// require.True(t,value);
// ```
type Args struct {
	items []string
	name  Option[string]
	c_rev Option[uint]
}

func newArgs(value ...string) Args {
	return Args{
		items: value,
		name:  Invalid[string](),
		c_rev: Invalid[uint](),
	}
}

// Enable completions with custom output revision style
// /
// Use revision 0 if you want to test completion mechanism
// /
// ```rust
// # use bpaf*;
// parser := short('f').switch().to_options();
// // ask bpaf to produce more input from "-", for
// // suggesting new items use "" at the end
// r := parser.run_inner(Argsfrom(&["-"])
//
//	.set_comp(0))
//	.unwrap_err()
//	.unwrap_stdout();
//
// require.Equal(t,r, "-f");
// ```
// /
// Note to self: shell passes "" as a parameter in situations like foo `--bar TAB`, bpaf
// completion stubs adopt this conventions add pass it along. This is needed so completer can
// tell the difference between `--bar` being completed or an argument to it in the example
// above.
func (self *Args) set_comp(rev uint) *Args {
	self.c_rev = Valid(rev)
	return self
}

// Add an application name for args created from custom input
// ```rust
// # use bpaf*;
// parser := short('f').switch().to_options();
// r := parser
//
//	.run_inner(Argsfrom(&["--help"]).set_name("my_app"))
//	.unwrap_err()
//	.unwrap_stdout();
//
// # drop(r);
// ```
func (self *Args) set_name(name string) *Args {
	self.name = Valid(name)
	return self
}

// Get a list of command line arguments from OS
func current_args() Args {
	name := os.Args[0]
	return Args{
		items: os.Args[1:],
		name:  Valid(filepath.Base(name)),
		c_rev: Invalid[uint](),
	}
}

type ItemStateKind uint8

const (
	ItemStateKindUnparsed ItemStateKind = iota // Value is yet to be parsed
	ItemStateKindConflict                      // Both branches succeeded, first parser was taken in favor of the second one
	ItemStateKindParsed                        // Value was parsed
)

// Shows which branch of [`ParseOrElse`] parsed the argument
type ItemState struct {
	kind     ItemStateKind
	conflict int
}

var ItemStateUnparsed = ItemState{kind: ItemStateKindUnparsed}

func ItemStateConflict(value int) ItemState {
	return ItemState{
		kind:     ItemStateKindConflict,
		conflict: value,
	}
}

var ItemStateParsed = ItemState{kind: ItemStateKindParsed}

func (self ItemState) parsed() bool {
	return self.kind == ItemStateKindParsed
}
func (self ItemState) present() bool {
	return self.kind != ItemStateKindParsed
}

func disambiguate_short(
	os string,
	short string,
	short_flags []byte,
	short_args []byte,
	items *[]Arg,
) Option[Message] {
	// block can start with 0 or more short flags
	// followed by zero or one short argument, possibly with a body

	// keep the old length around so we can trimp items to it and push a ArgWord
	// if we decide to give up
	original := len(*items)

	// first flag contains the original os string for error message and anywhere purposes
	for ix, c := range short {
		cc := byte(c)
		tail_ix := ix + utf8.RuneLen(c)
		rest := short[tail_ix:]

		// shortcircuit single character short fun.Options
		if ix == 0 && rest == "" {
			*items = append(*items, ArgShort(cc, false, os))
			return Invalid[Message]()
		}
		a := bytes.ContainsRune(short_flags, c)
		b := bytes.ContainsRune(short_args, c)
		switch {
		// short name that can be flag
		case a && !b:
			*items = append(*items, ArgShort(cc, false, os))
			// short name that can be argument
		case !a && b:
			adjacent_body := rest != ""
			*items = append(*items, ArgShort(cc, adjacent_body, os))
			if adjacent_body {
				*items = append(*items, ArgWord(rest))
			}
			return Invalid[Message]()
		// neither is valid and there's more than one character. fallback to using it as a Word
		case !a && !b:
			*items = (*items)[:original]
			*items = append(*items, ArgWord(os))
			return Invalid[Message]()
		// ambiguity, this is bad
		case a && b:
			msg := MessageAmbiguity(len(*items), short)
			*items = append(*items, ArgWord(os))
			return Valid(msg)
		}
	}
	return Invalid[Message]()
}

type State struct {
	// list of all available command line arguments, in `Rc` for cheap cloning
	items []Arg

	item_state []ItemState

	// performance optimization mostly - tracks removed item and gives cheap is_empty and len
	remaining int

	// Used to render an error message for [`parse`][crateParserparse]
	// contains an index of a currently consumed item if we are parsing a single
	// item
	current Option[int]

	// path to current command, "deeper" parser should win in or_else branches
	path []string

	comp Option[Complete]

	// A way to customize behavior for --help and error handling
	// improve_error: superImprove,

	// Describes scope current parser will be consuming elements from. Usually it will be
	// considering the whole sequence of (unconsumed) arguments, but for "adjacent"
	// scope starts on the right of the first consumed item and might end before the end
	// of the list, similarly for "commands"
	scope Range
}

// Check if item at ixth position is still present (was not parsed)
func (self State) present(ix int) Option[bool] {
	if ix >= len(self.item_state) {
		return Invalid[bool]()
	}
	return Valid(self.item_state[ix].present())
}

func (self State) depth() uint {
	return uint(len(self.path))
}

func (self State) check_no_pos_ahead() bool {
	c, ok := self.comp.Unpack()
	return ok && c.no_pos_ahead
}

func (self *State) set_no_pos_ahead() {
	if _, ok := self.comp.Unpack(); ok {
		self.comp.Value.no_pos_ahead = true
	}
}

func stateconstruct(
	args Args,
	short_flags []byte,
	short_args []byte,
	err *Option[Message],
) State {
	items := []Arg{}
	pos_only := false
	double_dash_marker := Invalid[int]()

	comp_scanner := ArgScanner{
		revision: args.c_rev,
		name:     args.name,
	}

LOOP:
	for _, os := range args.items {
		if pos_only {
			items = append(items, ArgPosWord(os))
			continue
		}

		if comp_scanner.check_next(os) {
			continue
		}

		argType, s, arg, ok := split_os_argument(os)
		switch {
		// -f and -fbar, but also -vvvvv
		case ok && argType == ArgTypeShort && !arg.Valid:
			if msg, ok := disambiguate_short(
				os,
				s,
				short_flags,
				short_args,
				&items,
			).Unpack(); ok {
				*err = Valid(msg)
				break LOOP
			}
		case ok && argType == ArgTypeShort && arg.Valid:
			items = append(items, ArgShort(s[0], true, os))
			items = append(items, arg.Value)
			// --key and --key=val
		case ok && argType == ArgTypeLong:
			items = append(items, ArgLong(s, arg.Valid, os))
			if arg, ok := arg.Unpack(); ok {
				items = append(items, arg)
			}
			// something that is not a short or long flag, keep them as positionals
			// handle "--" specifically as "end of flags" marker
		case !ok:
			if os == "--" {
				double_dash_marker = Valid(len(items))
				pos_only = true
			}
			items = append(items, func() Arg {
				if pos_only {
					return ArgPosWord(os)
				} else {
					return ArgWord(os)
				}
			}())
		}
	}

	item_state := make([]ItemState, len(items))
	for i := range items {
		item_state[i] = ItemStateUnparsed
	}
	remaining := len(items)
	if ix, ok := double_dash_marker.Unpack(); ok {
		item_state[ix] = ItemStateParsed
		remaining -= 1

		if comp_scanner.revision.Valid && ix == len(items)-1 {
			remaining += 1
			item_state[ix] = ItemStateUnparsed
		}
	}

	path := []string{}
	if name, ok := args.name.Unpack(); ok {
		path = append(path, name)
	}

	comp := comp_scanner.done()
	return State{
		items,
		item_state,
		remaining,
		Invalid[int](),
		path,
		comp,
		Range{0, len(items)},
	}
}

// creates iterator over remaining elements
func (self *State) items_iter() func() (int, Arg, bool) {
	args, cur := self, self.scope[0]
	return func() (int, Arg, bool) {
		for {
			if !args.scope.contains(cur) || cur >= len(args.item_state) {
				return 0, Arg{}, false
			}
			if args.item_state[cur].present() {
				defer func() { cur++ }()
				return cur, args.items[cur], true
			}
		}
	}
}

func (self *State) remove(index int) {
	if self.scope.contains(index) && self.item_state[index].present() {
		self.current = Valid(index)
		self.remaining -= 1
		self.item_state[index] = ItemStateParsed
	}
}

func (self State) pick_winner(other State) (bool, Option[uint]) {
	for ix := 0; ix < len(self.item_state); ix++ {
		me := self.item_state[ix]
		other := other.item_state[ix]
		if me.parsed() != other.parsed() {
			return me.parsed(), Valid(uint(ix))
		}
	}
	return true, Invalid[uint]()
}

// find first saved conflict
func (self State) conflict() Option[Range] {
	next := self.items_iter()
	ix, _, ok := next()
	if !ok {
		return Invalid[Range]()
	}
	if ix >= len(self.item_state) {
		return Invalid[Range]()
	}
	if xxx := self.item_state[ix]; xxx.kind == ItemStateKindConflict {
		other := xxx.conflict
		return Valid(Range{ix, other})
	}
	return Invalid[Range]()
}

func (self *State) save_conflicts(loser State, win int) {
	for i := 0; i < len(self.item_state); i++ {
		winner := self.item_state[i]
		loser := loser.item_state[i]
		if winner.present() && loser.parsed() {
			self.item_state[i] = ItemStateConflict(win)
		}
	}
}

// it is in use when autocomplete is enabled
func (self State) is_empty() bool {
	return self.remaining == 0
}

func (self State) len() int {
	return self.remaining
}

// Get an argument from a scope that was not consumed yet
func (self State) get(ix int) Option[Arg] {
	if self.scope.contains(ix) && self.item_state[ix].present() {
		return Valid(self.items[ix])
	} else {
		return Invalid[Arg]()
	}
}

// Check if parser performs autocompletion
// /
// used by construct macro
func (self State) is_comp() bool {
	return self.comp.Valid
}

// Narrow down scope of &self to adjacently consumed values compared to original.
func (self State) adjacent_scope(original State) Option[Range] {
	if len(self.items) == 0 {
		return Invalid[Range]()
	}

	// starting at the beginning of the scope look for the first mismatch
	start := self.scope[0]
	for offset := int(start); offset < len(self.item_state); offset++ {
		this := self.item_state[offset]
		orig := original.item_state[offset]
		// once there's a misswitch we have the scope we are looking for:
		// all the adjacent items consumed in this. It doesn't make sense to remove it if
		// it matches the original scope though...
		if this.present() && orig.present() {
			proposed_scope := Range{start, offset}
			if self.scope == proposed_scope {
				return Invalid[Range]()
			} else {
				return Valid(proposed_scope)
			}
		}
	}
	return Invalid[Range]()
}

// Get a scope for an adjacently available block of item starting at start
func (self State) adjacently_available_from(start int) Range {
	span_size := 0
	for _, v := range self.item_state[start:] {
		if v.present() {
			span_size++
		}
	}
	return Range{start, start + span_size}
}

type ArgRangesIter struct {
	args  State
	width int
	cur   int
}

type ArgRangesIterItem struct {
	startOffset         int
	widthOfTheFirstItem int
	state               State
}

func (self *ArgRangesIter) next() Option[ArgRangesIterItem] {
	for {
		cur := self.cur
		if cur > self.args.scope.end() {
			return Invalid[ArgRangesIterItem]()
		}
		self.cur += 1
		if v := self.args.present(cur); v.Valid && !v.Value {
			continue
		}
		if cur+self.width > len(self.args.items) {
			return Invalid[ArgRangesIterItem]()
		}
		// It should be possible to optimize this code a bit
		// by checking if first item can possibly
		args := self.args
		args.set_scope(Range{cur, len(self.args.items)})
		return Valid(ArgRangesIterItem{cur, self.width, args})
	}
}

func (self State) ranges(item Item) ArgRangesIter {
	width := 0
	item.v(
		nil, nil, nil,
		func(ShortLong, []byte, Option[string]) { width = 1 },
		func(ShortLong, []byte, Metavar, Option[string]) { width = 2 },
	)
	return ArgRangesIter{self, width, 0}
}

// Mark everything outside of `range` as removed
func (self *State) set_scope(scope Range) {
	self.scope = scope
	remaining := 0
	for _, v := range self.item_state[self.scope[0]:self.scope[1]] {
		if v.present() {
			remaining++
		}
	}
	self.remaining = remaining
}

// check if bpaf tries to complete last consumed element
func (self State) touching_last_remove() bool {
	return self.comp.Valid && len(self.items)-1 == self.current.OrDefault(math.MaxInt)
}

func (self *State) comp_mut() Option[*Complete] {
	return Option[*Complete]{&self.comp.Value, self.comp.Valid}
}

func (self State) comp_ref() Option[Complete] {
	return self.comp
}

func (self *State) swap_comps(other *State) {
	*self, *other = *other, *self
}

func (self State) swap_comps_with(comps *[]Comp) {
	if comp, ok := self.comp_mut().Unpack(); ok {
		comp.swap_comps(comps)
	}
}

// Get a short or long flag: `-f` / `--flag`
//
// Returns false if value isn't present
func (self State) take_flag(named NamedArg) bool {
	var (
		ix int
		ok bool
	)
	iter := self.items_iter()
	for i, arg := range push2(iter) {
		if named.matches_arg(arg, false) {
			ix = i
			ok = true
			break
		}
	}
	if ok {
		self.remove(ix)
	}
	return ok
}

// get a short or long arguments
//
// Returns Ok(Invalid()) if flag isn't present
// Returns Err if flag is present but value is either missing or strange.
func (self State) take_arg(
	named NamedArg,
	adjacent bool,
	metavar Metavar,
) Result[Option[string]] {
	var (
		key_ix int
		ok     bool
	)
	iter := self.items_iter()
	for i, arg := range push2(iter) {
		if named.matches_arg(arg, adjacent) {
			key_ix = i
			ok = true
			break
		}
	}

	if !ok {
		return Ok(Invalid[string]())
	}

	val_ix := key_ix + 1
	var val string
	if ww, ok := self.get(val_ix).Unpack(); ok && func() bool {
		res := false
		ww.value(nil, nil, func() { res = true }, func() { res = true }, nil)
		return res
	}() {
		val = ww.os_str
	} else {
		return Err[Option[string]](Error(MessageNoArgument(key_ix, metavar)))
	}
	self.current = Valid(val_ix)
	self.remove(key_ix)
	self.remove(val_ix)
	return Ok(Valid(val))
}

// gets first positional argument present
//
// returns Ok(Invalid()) if input is empty
// returns Err if first positional argument is a flag
func (self State) take_positional_word(metavar Metavar) (int, bool, string, Error) {
	var (
		ix     int
		strict bool
		w      Arg
		ok     bool
	)
	for _, arg := range push2(self.items_iter()) {
		arg.value(nil, nil, nil, func() {
			strict, ok = false, true
		}, func() {
			strict, ok = true, true
		})
	}
	if ok {
		self.current = Valid(ix)
		self.remove(ix)
		return ix, strict, w.os_str, nil
	} else {
		scope := self.scope
		missing := MissingItem{
			item: ItemPositional(
				metavar,
				Invalid[Doc](),
			),
			position: scope.start(),
			scope:    scope,
		}
		return 0, false, "", Error(MessageMissing(missing))
	}
}

// take a static string argument from the first present argument
func (self State) take_cmd(w string) bool {
	next := self.items_iter()
	ix, a, ok := next()
	if !ok || a.os_str != w {
		goto INVALID
	}

	if func() bool {
		res := false
		a.value(func(byte, bool) { res = true }, func(_ string, b bool) { res = !b }, func() { res = true }, nil, nil)
		return res
	}() {
		self.remove(ix)
		self.current = Valid(ix)
		return true
	}

INVALID:
	self.current = Invalid[int]()
	return false
}

// #[cfg(test)]
func (self State) peek() Option[Arg] {
	next := self.items_iter()
	_, a, ok := next()
	return Option[Arg]{a, ok}
}
