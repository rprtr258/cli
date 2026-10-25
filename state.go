package cli

import (
	"math"

	"github.com/rprtr258/fun"
)

type state struct {
	// list of all available command line arguments
	items      []arg
	item_state []itemState

	// Used to render an error message for `parse`
	// contains an index of a currently consumed item if we are parsing a single item
	currentIdx Opt[int]

	// path to current command, "deeper" parser should win in or_else branches
	path []string

	comp Opt[Complete]

	// A way to customize behavior for --help and error handling
	// improve_error: superImprove,

	// Describes scope current parser will be consuming elements from. Usually it will be
	// considering the whole sequence of (unconsumed) arguments, but for "adjacent"
	// scope starts on the right of the first consumed item and might end before the end
	// of the list, similarly for "commands"
	scope rng
}

func (self state) clone() state {
	s := self
	s.items = append([]arg(nil), self.items...)
	s.item_state = append([]itemState(nil), self.item_state...)
	s.comp.Value.comps = append([]comp(nil), self.comp.Value.comps...)
	return s
}

// present checks whether the item at the ixth position is still present (was
// not parsed).
func (self state) present(ix int) (present, exact bool) {
	inBounds := ix < len(self.item_state)
	return inBounds && !self.item_state[ix].parsed(), inBounds
}

func (self state) depth() int {
	return len(self.path)
}

func (self state) check_no_pos_ahead() bool {
	c, ok := self.comp.Unpack()
	return ok && c.no_pos_ahead
}

func (self *state) set_no_pos_ahead() {
	if _, ok := self.comp.Unpack(); ok {
		self.comp.Value.no_pos_ahead = true
	}
}

func newState(
	args argList,
	short_flags []byte,
	short_args []byte,
) (state, parseError) {
	items := make([]arg, 0, len(args.items))
	pos_only := false
	double_dash_marker := Invalid[int]()

	comp_scanner := argScanner{
		revision: args.comp,
		name:     args.name,
	}

	var err parseError
LOOP:
	for _, os := range args.items {
		if pos_only {
			items = append(items, argPositional(os))
			continue
		}

		if comp_scanner.check_next(os) {
			continue
		}

		argType, s, a, ok := splitOsArgument(os)
		switch {
		// -f and -fbar, but also -vvvvv
		case ok && argType == argTypeShort && !a.Valid:
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
		case ok && argType == argTypeShort && a.Valid:
			items = append(items, argShort(os, s[0], true), a.Value)
			// --key and --key=val
		case ok && argType == argTypeLong:
			items = append(items, argLong(os, s, a.Valid))
			if value, ok := a.Unpack(); ok {
				items = append(items, value)
			}
			// something that is not a short or long flag, keep them as positionals
			// handle "--" specifically as "end of flags" marker
		case !ok:
			if os == "--" {
				double_dash_marker = Valid(len(items))
				pos_only = true
			}
			items = append(items, func() arg {
				if pos_only {
					return argPositional(os)
				} else {
					return argWord(os)
				}
			}())
		}
	}

	item_state := make([]itemState, len(items))
	if ix, ok := double_dash_marker.Unpack(); ok {
		item_state[ix] = itemStateParsed

		if comp_scanner.revision.Valid && ix == len(items)-1 {
			item_state[ix] = itemStateUnparsed
		}
	}

	name, name_ok := args.name.Unpack()

	comp, comp_ok := comp_scanner.done()
	return state{
		items,
		item_state,
		Invalid[int](),
		fun.IF(name_ok, []string{name}, []string{}),
		Opt[Complete]{comp, comp_ok},
		rng{0, len(items)},
	}, err
}

// items_iter creates an iterator over the remaining elements.
func (self state) items_iter() func() (int, arg, bool) {
	i := self.scope[0]
	return func() (int, arg, bool) {
		for self.scope.contains(i) && i < len(self.item_state) {
			if !self.item_state[i].parsed() {
				a, b := i, self.items[i]
				i++
				return a, b, true
			}
			i++
		}
		return 0, arg{}, false
	}
}

func (self *state) remove(index int) {
	if !self.scope.contains(index) || self.item_state[index].parsed() {
		return
	}

	self.currentIdx = Valid(index)
	self.item_state[index] = itemStateParsed
}

func (self state) pick_winner(other state) (bool, Opt[int]) {
	for i, me := range self.item_state {
		other := other.item_state[i]
		if me.parsed() != other.parsed() {
			return me.parsed(), Valid(i)
		}
	}
	return true, Invalid[int]()
}

// conflict finds the first saved conflict.
func (self state) conflict() (rng, bool) {
	next := self.items_iter()
	ix, _, ok := next()
	if !ok {
		return rng{}, false
	}
	if ix >= len(self.item_state) {
		return rng{}, false
	}
	if xxx := self.item_state[ix]; xxx.kind == itemStateKindConflict {
		other := xxx.conflict
		return rng{ix, other}, true
	}
	return rng{}, false
}

func (self *state) save_conflicts(loser state, win int) {
	for i, winner := range self.item_state {
		loser := loser.item_state[i]
		if !winner.parsed() && loser.parsed() {
			self.item_state[i] = itemStateConflict(win)
		}
	}
}

func (self state) len() int {
	res := 0
	for _, v := range self.item_state {
		if !v.parsed() {
			res++
		}
	}
	return res
}

// get returns an argument from a scope that was not consumed yet.
func (self state) get(ix int) (arg, bool) {
	if !self.scope.contains(ix) || self.item_state[ix].parsed() {
		return arg{}, false
	}

	return self.items[ix], true
}

// adjacent_scope narrows the scope to adjacently consumed values compared to
// the original.
func (self state) adjacent_scope(original state) (rng, bool) {
	if len(self.items) == 0 {
		return rng{}, false
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
			proposed_scope := rng{start, offset}
			if self.scope == proposed_scope {
				return rng{}, false
			} else {
				return proposed_scope, true
			}
		}
	}
	return rng{}, false
}

// adjacently_available_from returns a scope for an adjacently available block
// of items starting at start.
func (self state) adjacently_available_from(start int) rng {
	span_size := 0
	for _, v := range self.item_state[start:] {
		if !v.parsed() {
			span_size++
		}
	}
	return rng{start, start + span_size}
}

type argRangesIterItem struct {
	startOffset         int
	widthOfTheFirstItem int
	state               state
}

func (self state) ranges(item item) func() (argRangesIterItem, bool) {
	width := 0
	switch item.v.(type) {
	case itemAny, itemPositional, itemCommand, itemFlag:
		width = 1
	case itemArgument:
		width = 2
	}

	itcur := 0
	return func() (argRangesIterItem, bool) {
		for itcur <= self.scope.end() {
			cur := itcur
			itcur++
			if v, ok := self.present(cur); ok && !v {
				continue
			}
			if cur+width > len(self.items) {
				return argRangesIterItem{}, false
			}
			// It should be possible to optimize this code a bit
			// by checking if first item can possibly
			args := self.clone()
			args.scope = rng{cur, len(args.items)}
			return argRangesIterItem{cur, width, args}, true
		}
		return argRangesIterItem{}, false
	}
}

// touching_last_remove checks whether the library tries to complete the last
// consumed element.
func (self state) touching_last_remove() bool {
	return self.comp.Valid && len(self.items)-1 == self.currentIdx.OrDefault(math.MaxInt)
}

func (self *state) comp_mut() (*Complete, bool) {
	return &self.comp.Value, self.comp.Valid
}

func (self *state) swap_comps_with(comps *[]comp) {
	if self.comp.Valid {
		*comps, self.comp.Value.comps = self.comp.Value.comps, *comps
	}
}

// takeBoolFlag takes a short or long flag: `-f` / `--flag`.
//
// Returns false if value isn't present
func (self *state) takeBoolFlag(named NamedArg) bool {
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

// takeFlag takes a short or long argument.
//
// Returns Invalid() if flag isn't present
// Returns Err if flag is present but value is either missing or strange.
func (self *state) takeFlag(
	named NamedArg,
	adjacent bool,
	metavar Metavar,
) (Opt[string], parseError) {
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
	if !ok || ww.kind != argKindWord {
		return Opt[string]{}, errorNoArgument{idx, metavar}
	}

	self.currentIdx = Valid(idx + 1)
	self.remove(idx)
	self.remove(idx + 1)
	return Valid(ww.str), nil
}

// takePositional takes the first positional argument present.
//
// returns Ok(Invalid()) if input is empty
// returns Err if first positional argument is a flag
func (self *state) takePositional(metavar Metavar) (ix int, strict bool, value string, err parseError) {
	iter := push2(self.items_iter())
	for i, arg := range iter {
		if fun.Contains(arg.kind, argKindWord, argKindPositional) {
			strict := arg.kind == argKindPositional
			self.currentIdx = Valid(i)
			self.remove(i)
			return i, strict, arg.str, nil
		}
	}
	return 0, false, "", errorMissing([]missingItem{{
		item:     item{Invalid[Doc](), itemPositional{metavar}},
		position: self.scope.start(),
		scope:    self.scope,
	}})
}

// takeCmd takes a static string argument from the first present argument.
func (self *state) takeCmd(cmd string) bool {
	ix, a, ok := self.items_iter()()
	if ok && a.str == cmd &&
		(fun.Contains(a.kind, argKindShortFlag, argKindWord) || a.kind == argKindLongFlag && !a.isAdjacent) {
		self.currentIdx = Valid(ix)
		self.remove(ix)
		return true
	}

	self.currentIdx = Invalid[int]()
	return false
}
