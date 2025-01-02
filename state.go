package cli

import (
	"math"

	"github.com/rprtr258/fun"
)

type State struct {
	// list of all available command line arguments
	items      []Arg
	item_state []ItemState

	// Used to render an error message for `parse`
	// contains an index of a currently consumed item if we are parsing a single item
	currentIdx Option[int]

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

func (self State) clone() State {
	s := self
	s.items = append([]Arg(nil), self.items...)
	s.item_state = append([]ItemState(nil), self.item_state...)
	s.comp.Value.comps = append([]Comp(nil), self.comp.Value.comps...)
	return s
}

// Check if item at ixth position is still present (was not parsed)
func (self State) present(ix int) (bool, bool) {
	inBounds := ix < len(self.item_state)
	return inBounds && !self.item_state[ix].parsed(), inBounds
}

func (self State) depth() int {
	return len(self.path)
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

func newState(
	args Args,
	short_flags []byte,
	short_args []byte,
) (State, Error) {
	items := []Arg{}
	pos_only := false
	double_dash_marker := Invalid[int]()

	comp_scanner := ArgScanner{
		revision: args.comp,
		name:     args.name,
	}

	var err Error
LOOP:
	for _, os := range args.items {
		if pos_only {
			items = append(items, ArgPositional(os))
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
			); ok {
				err = msg
				break LOOP
			}
		case ok && argType == ArgTypeShort && arg.Valid:
			items = append(items, ArgShort(os, s[0], true))
			items = append(items, arg.Value)
			// --key and --key=val
		case ok && argType == ArgTypeLong:
			items = append(items, ArgLong(os, s, arg.Valid))
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
					return ArgPositional(os)
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
	if ix, ok := double_dash_marker.Unpack(); ok {
		item_state[ix] = ItemStateParsed

		if comp_scanner.revision.Valid && ix == len(items)-1 {
			item_state[ix] = ItemStateUnparsed
		}
	}

	name, ok := args.name.Unpack()

	comp, ok := comp_scanner.done()
	return State{
		items,
		item_state,
		Invalid[int](),
		fun.IF(ok, []string{name}, []string{}),
		Option[Complete]{comp, ok},
		Range{0, len(items)},
	}, err
}

// creates iterator over remaining elements
func (self State) items_iter() func() (int, Arg, bool) {
	i := self.scope[0]
	return func() (int, Arg, bool) {
		for self.scope.contains(i) && i < len(self.item_state) {
			if !self.item_state[i].parsed() {
				a, b := i, self.items[i]
				i++
				return a, b, true
			}
			i++
		}
		return 0, Arg{}, false
	}
}

func (self *State) remove(index int) {
	if !self.scope.contains(index) || self.item_state[index].parsed() {
		return
	}

	self.currentIdx = Valid(index)
	self.item_state[index] = ItemStateParsed
}

func (self State) pick_winner(other State) (bool, Option[int]) {
	for i, me := range self.item_state {
		other := other.item_state[i]
		if me.parsed() != other.parsed() {
			return me.parsed(), Valid(i)
		}
	}
	return true, Invalid[int]()
}

// find first saved conflict
func (self State) conflict() (Range, bool) {
	next := self.items_iter()
	ix, _, ok := next()
	if !ok {
		return Range{}, false
	}
	if ix >= len(self.item_state) {
		return Range{}, false
	}
	if xxx := self.item_state[ix]; xxx.kind == ItemStateKindConflict {
		other := xxx.conflict
		return Range{ix, other}, true
	}
	return Range{}, false
}

func (self *State) save_conflicts(loser State, win int) {
	for i, winner := range self.item_state {
		loser := loser.item_state[i]
		if !winner.parsed() && loser.parsed() {
			self.item_state[i] = ItemStateConflict(win)
		}
	}
}

func (self State) len() int {
	res := 0
	for _, v := range self.item_state {
		if !v.parsed() {
			res++
		}
	}
	return res
}

// Get an argument from a scope that was not consumed yet
func (self State) get(ix int) (Arg, bool) {
	if !self.scope.contains(ix) || self.item_state[ix].parsed() {
		return Arg{}, false
	}

	return self.items[ix], true
}

// Narrow down scope of &self to adjacently consumed values compared to original.
func (self State) adjacent_scope(original State) (Range, bool) {
	if len(self.items) == 0 {
		return Range{}, false
	}

	// starting at the beginning of the scope look for the first mismatch
	start := self.scope[0]
	for offset := int(start); offset < len(self.item_state); offset++ {
		this := self.item_state[offset]
		orig := original.item_state[offset]
		// once there's a misswitch we have the scope we are looking for:
		// all the adjacent items consumed in this. It doesn't make sense to remove it if
		// it matches the original scope though...
		if !this.parsed() && !orig.parsed() {
			proposed_scope := Range{start, offset}
			if self.scope == proposed_scope {
				return Range{}, false
			} else {
				return proposed_scope, true
			}
		}
	}
	return Range{}, false
}

// Get a scope for an adjacently available block of item starting at start
func (self State) adjacently_available_from(start int) Range {
	span_size := 0
	for _, v := range self.item_state[start:] {
		if !v.parsed() {
			span_size++
		}
	}
	return Range{start, start + span_size}
}

type argRangesIterItem struct {
	startOffset         int
	widthOfTheFirstItem int
	state               State
}

func (args State) ranges(item Item) func() (argRangesIterItem, bool) {
	width := 0
	switch item.v.(type) {
	case ItemFlag:
		width = 1
	case ItemArgument:
		width = 2
	}

	itcur := 0
	return func() (argRangesIterItem, bool) {
		for {
			cur := itcur
			if cur > args.scope.end() {
				return argRangesIterItem{}, false
			}
			itcur += 1
			if v, ok := args.present(cur); ok && !v {
				continue
			}
			if cur+width > len(args.items) {
				return argRangesIterItem{}, false
			}
			// It should be possible to optimize this code a bit
			// by checking if first item can possibly
			args := args.clone()
			args.scope = Range{cur, len(args.items)}
			return argRangesIterItem{cur, width, args}, true
		}
	}
}

// check if bpaf tries to complete last consumed element
func (self State) touching_last_remove() bool {
	return self.comp.Valid && len(self.items)-1 == self.currentIdx.OrDefault(math.MaxInt)
}

func (self *State) comp_mut() (*Complete, bool) {
	return &self.comp.Value, self.comp.Valid
}

func (self *State) swap_comps(other *State) {
	*self, *other = *other, *self
}

func (self State) swap_comps_with(comps *[]Comp) {
	if self.comp.Valid {
		self.comp.Value.swap_comps(comps)
	}
}

// Get a short or long flag: `-f` / `--flag`
//
// Returns false if value isn't present
func (self *State) takeBoolFlag(named NamedArg) bool {
	var (
		ix int
		ok bool
	)
	iter := push2(self.items_iter())
	for i, arg := range iter {
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
// Returns Invalid() if flag isn't present
// Returns Err if flag is present but value is either missing or strange.
func (self *State) takeFlag(
	named NamedArg,
	adjacent bool,
	metavar Metavar,
) (Option[string], Error) {
	var (
		idx int
		ok  bool
	)
	for i, arg := range push2(self.items_iter()) {
		if named.matches_arg(arg, adjacent) {
			idx, ok = i, true
			break
		}
	}
	if !ok {
		return Invalid[string](), nil
	}

	ww, ok := self.get(idx + 1)
	if !ok || !fun.Contains(ww.kind, ArgKindFlagValue, ArgKindWord) {
		return Option[string]{}, ErrorNoArgument{idx, metavar}
	}

	self.currentIdx = Valid(idx + 1)
	self.remove(idx)
	self.remove(idx + 1)
	return Valid(ww.str), nil
}

// gets first positional argument present
//
// returns Ok(Invalid()) if input is empty
// returns Err if first positional argument is a flag
func (self *State) takePositional(metavar Metavar) (int, bool, string, Error) {
	var (
		ix     int
		strict bool
		w      Arg
		ok     bool
	)
	iter := push2(self.items_iter())
	for i, arg := range iter {
		switch arg.kind {
		case ArgKindWord, ArgKindPositional:
			ix = i
			strict, ok = arg.kind == ArgKindPositional, true
			w = arg
		}
		if ok {
			break
		}
	}
	if !ok {
		return 0, false, "", ErrorMissing([]MissingItem{{
			item:     Item{Invalid[Doc](), ItemPositional{metavar}},
			position: self.scope.start(),
			scope:    self.scope,
		}})
	}

	self.currentIdx = Valid(ix)
	self.remove(ix)
	return ix, strict, w.str, nil
}

// take a static string argument from the first present argument
func (self *State) take_cmd(w string) bool {
	ix, a, ok := self.items_iter()()
	if !ok || a.str != w {
		goto INVALID
	}

	if fun.Contains(a.kind, ArgKindShortFlag, ArgKindWord) || a.kind == ArgKindLongFlag && !a.is_adj {
		self.remove(ix)
		self.currentIdx = Valid(ix)
		return true
	}

INVALID:
	self.currentIdx = Invalid[int]()
	return false
}
