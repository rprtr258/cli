package cli

import (
	"bytes"
	"iter"
	"math"
	"os"
	"path/filepath"
	"slices"
	"unicode/utf8"

	"github.com/rprtr258/fun"
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
// assert!(value);
// /
// // this also works
// value := parser.run_inner(&["-f"])
//
//	.unwrap();
//
// assert!(value);
// ```
type Args struct {
	items iter.Seq[string]
	name  fun.Option[string]
	c_rev fun.Option[uint]
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
// assert_eq!(r, "-f");
// ```
// /
// Note to self: shell passes "" as a parameter in situations like foo `--bar TAB`, bpaf
// completion stubs adopt this conventions add pass it along. This is needed so completer can
// tell the difference between `--bar` being completed or an argument to it in the example
// above.
func (self *Args) set_comp(rev uint) *Args {
	self.c_rev = fun.Valid(rev)
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
	self.name = fun.Valid(name)
	return self
}

func newArgs(value []string) Args {
	return Args{
		items: slices.Values(value),
		name:  fun.Invalid[string](),
		c_rev: fun.Invalid[uint](),
	}
}

// Get a list of command line arguments from OS
func current_args() Args {
	name := os.Args[0]
	return Args{
		items: slices.Values(os.Args[1:]),
		name:  fun.Valid(filepath.Base(name)),
		c_rev: fun.Invalid[uint](),
	}
}

// Shows which branch of [`ParseOrElse`] parsed the argument
type ItemState func(
	// Value is yet to be parsed
	Unparsed func(),
	// Both branches succeeded, first parser was taken in favor of the second one
	Conflict func(uint),
	// Value was parsed
	Parsed func(),
)

var ItemStateUnparsed = ItemState(func(Unparsed func(), _ func(uint), _ func()) {
	if Unparsed != nil {
		Unparsed()
	}
})

func ItemStateConflict(value uint) ItemState {
	return ItemState(func(_ func(), Conflict func(uint), _ func()) {
		if Conflict != nil {
			Conflict(value)
		}
	})
}

var ItemStateParsed = ItemState(func(_ func(), _ func(uint), Parsed func()) {
	if Parsed != nil {
		Parsed()
	}
})

func (self ItemState) parsed() bool {
	res := false
	self(nil, nil, func() { res = true })
	return res
}
func (self ItemState) present() bool {
	res := true
	self(nil, nil, func() { res = false })
	return res
}

func disambiguate_short(
	os string,
	short string,
	short_flags []byte,
	short_args []byte,
	items *[]Arg,
) fun.Option[Message] {
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
			return fun.Invalid[Message]()
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
			return fun.Invalid[Message]()
		// neither is valid and there's more than one character. fallback to using it as a Word
		case !a && !b:
			*items = (*items)[:original]
			*items = append(*items, ArgWord(os))
			return fun.Invalid[Message]()
		// ambiguity, this is bad
		case a && b:
			msg := MessageAmbiguity(len(*items), short)
			*items = append(*items, ArgWord(os))
			return fun.Valid(msg)
		}
	}
	return fun.Invalid[Message]()
}

type State struct {
	// list of all available command line arguments, in `Rc` for cheap cloning
	items []Arg

	item_state []ItemState

	// performance optimization mostly - tracks removed item and gives cheap is_empty and len
	remaining uint

	// Used to render an error message for [`parse`][crateParserparse]
	// contains an index of a currently consumed item if we are parsing a single
	// item
	current fun.Option[uint]

	// path to current command, "deeper" parser should win in or_else branches
	path []string

	comp fun.Option[Complete]

	//        // A way to customize behavior for --help and error handling
	//        improve_error: superImprove,
	// Describes scope current parser will be consuming elements from. Usually it will be
	// considering the whole sequence of (unconsumed) arguments, but for "adjacent"
	// scope starts on the right of the first consumed item and might end before the end
	// of the list, similarly for "commands"
	scope Range
}

// Check if item at ixth position is still present (was not parsed)
func (self State) present(ix uint) fun.Option[bool] {
	if ix >= uint(len(self.item_state)) {
		return fun.Invalid[bool]()
	}
	return fun.Valid(self.item_state[ix].present())
}

func (self State) depth() uint {
	return uint(len(self.path))
}

type ArgsIter struct {
	args State
	cur  uint
}

func (self *ArgsIter) next() (uint, Arg, bool) {
	for {
		ix := self.cur
		if !self.args.scope.contains(ix) {
			return 0, Arg{}, false
		}
		self.cur += 1
		if ix >= uint(len(self.args.item_state)) {
			return 0, Arg{}, false
		}
		if self.args.item_state[ix].present() {
			return ix, self.args.items[ix], true
		}
	}
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
	err *fun.Option[Message],
) State {
	items := []Arg{}
	pos_only := false
	double_dash_marker := fun.Invalid[int]()

	comp_scanner := ArgScanner{
		revision: args.c_rev,
		name:     args.name,
	}

LOOP:
	for os := range args.items {
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
				*err = fun.Valid(msg)
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
				double_dash_marker = fun.Valid(len(items))
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
	remaining := uint(len(items))
	if ix, ok := double_dash_marker.Unpack(); ok {
		item_state[ix] = ItemStateParsed
		remaining -= 1

		if comp_scanner.revision.Valid && ix == len(items)-1 {
			remaining += 1
			item_state[ix] = ItemStateUnparsed
		}
	}

	path := []string{}

	comp := comp_scanner.done()

	if name, ok := args.name.Unpack(); ok {
		path = append(path, name)
	}
	return State{
		items,
		item_state,
		remaining,
		fun.Invalid[uint](),
		path,
		comp,
		Range{0, uint(len(items))},
	}
}

// creates iterator over remaining elements
func (self State) items_iter() ArgsIter {
	return ArgsIter{
		self,
		self.scope[0],
	}
}

func (self *State) remove(index uint) {
	if self.scope.contains(index) && self.item_state[index].present() {
		self.current = fun.Valid(index)
		self.remaining -= 1
		self.item_state[index] = ItemStateParsed
	}
}

func (self State) pick_winner(other State) (bool, fun.Option[uint]) {
	for ix := 0; ix < len(self.item_state); ix++ {
		me := self.item_state[ix]
		other := other.item_state[ix]
		if me.parsed() != other.parsed() {
			return me.parsed(), fun.Valid(uint(ix))
		}
	}
	return true, fun.Invalid[uint]()
}

// find first saved conflict
func (self State) conflict() fun.Option[Range] {
	iter := self.items_iter()
	ix, _, ok := iter.next()
	if !ok {
		return fun.Invalid[Range]()
	}
	if ix >= uint(len(self.item_state)) {
		return fun.Invalid[Range]()
	}
	res := fun.Invalid[Range]()
	self.item_state[ix](nil, func(other uint) {
		res = fun.Valid(Range{ix, other})
	}, nil)
	return res
}

func (self *State) save_conflicts(loser State, win uint) {
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

func (self State) len() uint {
	return self.remaining
}

// Get an argument from a scope that was not consumed yet
func (self State) get(ix int) fun.Option[Arg] {
	if self.scope.contains(ix) && self.item_state[ix].present() {
		return fun.Valid(self.items[ix])
	} else {
		return fun.Invalid[Arg]()
	}
}

// Check if parser performs autocompletion
// /
// used by construct macro
func (self State) is_comp() bool {
	return self.comp.Valid
}

// Narrow down scope of &self to adjacently consumed values compared to original.
func (self State) adjacent_scope(original State) fun.Option[Range] {
	if len(self.items) == 0 {
		return fun.Invalid[Range]()
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
			proposed_scope := Range{start, uint(offset)}
			if self.scope == proposed_scope {
				return fun.Invalid[Range]()
			} else {
				return fun.Valid(proposed_scope)
			}
		}
	}
	return fun.Invalid[Range]()
}

// Get a scope for an adjacently available block of item starting at start
func (self State) adjacently_available_from(start uint) Range {
	span_size := uint(0)
	for _, v := range self.item_state[start:] {
		if v.present() {
			span_size++
		}
	}
	return Range{start, start + span_size}
}

type ArgRangesIter struct {
	args  State
	width uint
	cur   uint
}

type ArgRangesIterItem = struct {
	startOffset         uint
	widthOfTheFirstItem uint
	state               State
}

func (self *ArgRangesIter) next() fun.Option[ArgRangesIterItem] {
	for {
		cur := self.cur
		if cur > self.args.scope.end() {
			return fun.Invalid[ArgRangesIterItem]()
		}
		self.cur += 1
		if v := self.args.present(); v.Valid && !v.Value {
			continue
		}
		if cur+self.width > uint(len(self.args.items)) {
			return fun.Invalid[ArgRangesIterItem]()
		}
		// It should be possible to optimize this code a bit
		// by checking if first item can possibly
		args := self.args
		args.set_scope(Range{cur, uint(len(self.args.items))})
		return fun.Valid(ArgRangesIterItem{cur, self.width, args})
	}
}

type ArgsIterItem = fun.Pair[uint, Arg]

func (self *ArgRangesIter) next() fun.Option[ArgsIterItem] {
	for {
		ix := self.cur
		if !self.args.scope.contains(ix) {
			return fun.Invalid[ArgsIterItem]()
		}
		self.cur += 1
		if ix >= uint(len(self.args.item_state)) {
			return fun.Invalid[ArgsIterItem]()
		}
		if self.args.item_state[ix].present() {
			return fun.Valid(ArgsIterItem{ix, self.args.items[ix]})
		}
	}
}

func (self State) ranges(item Item) ArgRangesIter {
	width := uint(0)
	item(
		nil, nil, nil,
		func(ShortLong, []byte, fun.Option[string], fun.Option[Doc]) { width = 1 },
		func(ShortLong, []byte, Metavar, fun.Option[string], fun.Option[Doc]) { width = 2 },
	)
	return ArgRangesIter{self, width, 0}
}

// Mark everything outside of `range` as removed
func (self *State) set_scope(scope Range) {
	self.scope = scope
	remaining := uint(0)
	for _, v := range self.item_state[self.scope[0]:self.scope[1]] {
		if v.present() {
			remaining++
		}
	}
	self.remaining = remaining
}

// check if bpaf tries to complete last consumed element
func (self State) touching_last_remove() bool {
	return self.comp.Valid && uint(len(self.items)-1) == self.current.OrDefault(math.MaxUint)
}

func (self *State) comp_mut() fun.Option[*Complete] {
	return fun.Option[*Complete]{&self.comp.Value, self.comp.Valid}
}

func (self State) comp_ref() fun.Option[Complete] {
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
	if ix, _, ok := self.items_iter().find(func(arg) { return named.matches_arg(arg[1], false) }).Unpack(); ok {
		self.remove(ix)
		return true
	} else {
		return false
	}
}

// get a short or long arguments
//
// Returns Ok(fun.Invalid()) if flag isn't present
// Returns Err if flag is present but value is either missing or strange.
func (self State) take_arg(
	named NamedArg,
	adjacent bool,
	metavar Metavar,
) Result[fun.Option[string]] {
	var key_ix int
	var _arg Arg
	if v, ok := self.items_iter().find(func(arg) { return named.matches_arg(arg[1], adjacent) }).Unpack(); ok {
		key_ix, _arg = v[0], v[1]
	} else {
		return Ok(fun.Invalid())
	}

	val_ix := key_ix + 1
	var val Arg
	if ww, ok := self.get(val_ix).Unpack(); ok && func() bool {
		res := false
		ww.value(nil, nil, func() { res = true }, func() { res = true }, nil)
		return res
	}() {
		val = w
	} else {
		return Err(Error(MessageNoArgument(key_ix, metavar)))
	}
	val := val
	self.current = fun.Valid(val_ix)
	self.remove(key_ix)
	self.remove(val_ix)
	return Ok(fun.Valid(val))
}

// gets first positional argument present
//
// returns Ok(fun.Invalid()) if input is empty
// returns Err if first positional argument is a flag
func (self State) take_positional_word(metavar Metavar) (uint, bool, string, Error) {
	var (
		ix     uint
		strict bool
		w      Arg
		ok     bool
	)
	iter := self.items_iter()
	for {
		ix, arg, ok := iter.next().Unpack()
		arg.value(nil, nil, func() {
			ix, strict, w, ok = ix, false, w, true
		}, nil, func() {
			ix, strict, w, ok = ix, true, w, true
		})
	}
	if ok {
		self.current = fun.Valid(ix)
		self.remove(ix)
		return ix, strict, w, nil
	} else {
		scope := self.scope()
		missing := MissingItem{
			item: ItemPositional{
				help: fun.Invalid(),
				metavar,
			},
			position: scope.start,
			scope,
		}
		return 0, false, "", Error(MessageMissing([]MissingItem{missing}))
	}
}

// take a static string argument from the first present argument
func (self State) take_cmd(w string) bool {
	iter := self.items_iter()
	ix, a, ok := iter.next()
	if !ok || a.os_str != w {
		goto INVALID
	}

	if func() bool {
		kek := false
		a.value(func(byte, bool) { kek = true }, func(_ string, b bool) { kek = !b }, func() { kek = true }, nil, nil)
		return kek
	}() {
		self.remove(ix)
		self.current = fun.Valid(ix)
		return true
	}

INVALID:
	self.current = fun.Invalid[uint]()
	return false
}

// #[cfg(test)]
func (self State) peek() fun.Option[Arg] {
	iter := self.items_iter()
	_, a, ok := iter.next()
	return fun.Optional(a, ok)
}
