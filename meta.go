package cli

import (
	"fmt"

	"github.com/rprtr258/fun"
)

type Meta interface {
	isMeta()
}

// All arguments listed in a vector must be present
type MetaAnd []Meta

func (MetaAnd) isMeta() {}

// One of arguments listed in a vector must be present
type MetaOr []Meta

func (MetaOr) isMeta() {}

// Arguments are fun.Optional and encased in [] when rendered
type MetaOptional struct{ m Meta }

func (MetaOptional) isMeta() {}

// Arguments are requred and encased in () when rendered
type MetaRequired struct{ m Meta }

func (MetaRequired) isMeta() {}

// Argumens are required to be adjacent to each other
type MetaAdjacent struct{ m Meta }

func (MetaAdjacent) isMeta() {}

// Primitive argument as described
type MetaItem Item

func (MetaItem) isMeta() {}

// Accepts multiple arguments
type MetaMany struct{ m Meta }

func (MetaMany) isMeta() {}

// Arguments form a subsection with buffer being it's header
//
// whole set of arguments go into the same section as the first one
type MetaSubsection struct {
	m Meta
	u Doc
}

func (MetaSubsection) isMeta() {}

// Buffer is rendered after
type MetaSuffix struct {
	m Meta
	u Doc
}

func (MetaSuffix) isMeta() {}

// This item is not rendered in the help message
type MetaSkip struct{}

func (MetaSkip) isMeta() {}

// TODO make it Option[Doc]
type MetaCustomUsage struct {
	m Meta
	u Doc
}

func (MetaCustomUsage) isMeta() {}

// this meta must be prefixed with -- in unsage group
type MetaStrict struct{ m Meta }

func (MetaStrict) isMeta() {}

func DefaultMeta() Meta {
	return MetaSkip{}
}

func cloneMeta(m Meta) Meta {
	cloneDoc := func(d Doc) Doc {
		return Doc{d.payload, append([]Token(nil), d.tokens...)}
	}
	switch m := m.(type) {
	case MetaAnd:
		return MetaAnd(fun.Map[Meta](cloneMeta, m...))
	case MetaOr:
		return MetaOr(fun.Map[Meta](cloneMeta, m...))
	case MetaOptional:
		return MetaOptional{cloneMeta(m.m)}
	case MetaRequired:
		return MetaRequired{cloneMeta(m.m)}
	case MetaAdjacent:
		return MetaAdjacent{cloneMeta(m.m)}
	case MetaItem:
		return MetaItem(m) // TODO: clone item
	case MetaMany:
		return MetaMany{cloneMeta(m.m)}
	case MetaSubsection:
		return MetaSubsection{cloneMeta(m.m), cloneDoc(m.u)}
	case MetaSuffix:
		return MetaSuffix{cloneMeta(m.m), cloneDoc(m.u)}
	case MetaSkip:
		return MetaSkip{}
	case MetaCustomUsage:
		return MetaCustomUsage{cloneMeta(m.m), cloneDoc(m.u)}
	case MetaStrict:
		return MetaStrict{cloneMeta(m.m)}
	}
	panic("unreachable")
}

func IsAnd(m Meta) bool {
	_, ok := m.(MetaAnd)
	return ok
}

func IsOr(m Meta) bool {
	_, ok := m.(MetaOr)
	return ok
}

func IsSkip(m Meta) bool {
	_, ok := m.(MetaSkip)
	return ok
}

// MetaStrict should bubble up to one of 3 places:
// - top level
// - one of "and" elements
// - one of "or" elements
type StrictNorm uint8

const (
	// starting at the top and looking for Strict inside
	StrictNormPull StrictNorm = iota
	StrictNormPush
	// Already accounted for one, can strip the rest
	StrictNormStrip
)

func (self StrictNorm) IsPull() bool  { return self == StrictNormPull }
func (self StrictNorm) IsPush() bool  { return self == StrictNormPush }
func (self StrictNorm) IsStrip() bool { return self == StrictNormStrip }

func (self *StrictNorm) push() {
	if *self == StrictNormPull {
		*self = StrictNormPush
	}
}

// Used by normalization function to collapse duplicated commands.
// It seems to be fine to strip section information but not anything else
func is_command(self Meta) bool {
	switch self := self.(type) {
	case MetaItem:
		_, ok := self.v.(ItemCommand)
		return ok
	case MetaSubsection:
		return is_command(self.m)
	}
	return false
}

// do a nested invariant check
func positional_invariant_check(self Meta, verbose bool) {
	var GO func(meta Meta, is_pos *bool, v bool)
	GO = func(meta Meta, is_pos *bool, v bool) {
		switch meta := meta.(type) {
		case MetaAnd:
			for _, x := range meta {
				GO(x, is_pos, v)
			}
		case MetaOr:
			out := *is_pos
			for _, x := range meta {
				this_pos := *is_pos
				GO(x, &this_pos, v)
				out = out || this_pos
			}
			*is_pos = out
		case MetaOptional:
			GO(meta.m, is_pos, v)
		case MetaRequired:
			GO(meta.m, is_pos, v)
		case MetaAdjacent:
			if i, ok := Metafirst_item(meta.m).Unpack(); ok {
				if i.is_pos() {
					GO(meta.m, is_pos, v)
				} else {
					inner := false
					GO(meta.m, &inner, v)
				}
			}
		case MetaItem:
			i := Item(meta)
			switch {
			case *is_pos == i.is_pos():
			case *is_pos:
				panic(fmt.Sprintf(`bpaf usage BUG: all positional and command items must be placed in the right most position of the structure or tuple they are in but %#v breaks this rule. See bpaf documentation for "positional" for details.`, i))
			case i.is_pos():
				*is_pos = true
			}
			m, ok := i.v.(ItemCommand)
			if ok {
				command_pos := false
				if v {
					fmt.Printf("Checking\n%#v\n", m.meta)
				}
				GO(m.meta, &command_pos, v)
			}
		case MetaMany:
			GO(meta.m, is_pos, v)
		case MetaSubsection:
			GO(meta.m, is_pos, v)
		case MetaSuffix:
			GO(meta.m, is_pos, v)
		case MetaSkip:
		case MetaCustomUsage:
			GO(meta.m, is_pos, v)
		case MetaStrict:
			GO(meta.m, is_pos, v)
		}
	}
	is_pos := false
	if verbose {
		fmt.Printf("Checking\n%#v\n", self)
	}
	GO(self, &is_pos, verbose)
}

func normalized(m Meta, for_usage bool) Meta {
	m = cloneMeta(m)
	norm := StrictNormPull
	normalize(&m, for_usage, &norm)
	// strip outer () around meta unless inner
	if i, ok := m.(MetaRequired); ok {
		m = i
	}
	if _, ok := m.(MetaOr); ok {
		m = MetaRequired{m}
	}
	if norm == StrictNormPush {
		m = MetaStrict{m}
	}
	return m
}

// Used by adjacent parsers since it inherits behavior of the front item
func Metafirst_item(meta Meta) Option[Item] {
	switch meta := meta.(type) {
	case MetaAnd:
		return OptFlatMap(First(meta), Metafirst_item)
	case MetaItem:
		return Valid(Item(meta))
	case MetaOptional:
		return Metafirst_item(meta.m)
	case MetaRequired:
		return Metafirst_item(meta.m)
	case MetaAdjacent:
		return Metafirst_item(meta.m)
	case MetaMany:
		return Metafirst_item(meta.m)
	case MetaSubsection:
		return Metafirst_item(meta.m)
	case MetaSuffix:
		return Metafirst_item(meta.m)
	case MetaCustomUsage:
		return Metafirst_item(meta.m)
	case MetaStrict:
		return Metafirst_item(meta.m)
	}
	return Invalid[Item]()
}

// Normalize meta info for display as usage. Required propagates outwards
func normalize(self *Meta, for_usage bool, norm *StrictNorm) {
	normalize_vec := func(
		xs *[]Meta,
		for_usage bool,
		norm *StrictNorm,
		or bool,
	) (Meta, bool) {
		final_norm := *norm
		for i := range *xs {
			this_norm := *norm
			normalize(&(*xs)[i], for_usage, &this_norm)
			var target *StrictNorm = fun.IF(or, &final_norm, norm)
			switch {
			// switch (*target, this_norm) {
			case this_norm.IsPull() || target.IsStrip():
			case target.IsPull() && this_norm.IsPush():
				*self = MetaStrict{*self}
				*target = StrictNormStrip
			default:
				*target = this_norm
			}
		}
		*xs = fun.Filter(func(m Meta) bool { return !IsSkip(m) }, *xs...)

		*norm = final_norm

		switch len(*xs) {
		case 0:
			return MetaSkip{}, true
		case 1:
			res := (*xs)[0]
			*xs = (*xs)[1:]
			return res, true
		default:
			return nil, false
		}
	}

	switch m := (*self).(type) {
	case MetaAnd:
		if replacement, ok := normalize_vec((*[]Meta)(&m), for_usage, norm, false); ok {
			*self = replacement
		}
	// or should have either () or [] around it
	case MetaOr:
		if replacement, ok := normalize_vec((*[]Meta)(&m), for_usage, norm, true); ok {
			*self = replacement
		} else {
			saw_cmd := false
			// drop all the commands apart from the first one
			m = fun.Filter(func(m Meta) bool {
				is_cmd := is_command(m)
				keep := !(is_cmd && saw_cmd)
				saw_cmd = saw_cmd || is_cmd
				return keep
			}, m...)
			switch len(m) {
			case 0:
				*self = MetaSkip{}
			case 1:
				*self = m[0]
				m = m[1:]
			default:
				*self = MetaRequired{m}
			}
		}
	case MetaOptional:
		normalize(&m.m, for_usage, norm)
		switch mm := m.m.(type) {
		// TODO: mutate issue
		case MetaOptional:
			*self = mm.m
		case MetaRequired:
			*self = mm.m
		case MetaMany:
			if x, ok := mm.m.(MetaRequired); ok {
				*self = MetaMany(MetaOptional(x))
			}
		case MetaSkip:
			*self = MetaSkip{}
		}
	case MetaRequired:
		normalize(&m.m, for_usage, norm)
		switch {
		case IsSkip(m.m):
			// Required(Skip)Skip
			*self = MetaSkip{}
		case IsAnd(m) || IsOr(m):
			// keep () around composite parsers
		default:
			// and strip them elsewhere
			*self = m
		}
	case MetaMany:
		normalize(&m.m, for_usage, norm)
		if IsSkip(m.m) {
			*self = MetaSkip{}
		}
	case MetaItem:
		i := Item(m)
		i.normalize(for_usage)
	case MetaSkip:
		// nothing to do with items and skip just bubbles upwards
	case MetaCustomUsage:
		normalize(&m.m, for_usage, norm)
		// strip CustomUsage if we are not in usage so writer can simply render it
		if for_usage {
			*self = m
		} else if m.u.is_empty() {
			*self = MetaSkip{}
		}
	case MetaStrict:
		normalize(&m.m, for_usage, norm)
		norm.push()
		*self = m
	case MetaAdjacent:
		normalize(&m.m, for_usage, norm)
		*self = m
	case MetaSubsection:
		normalize(&m.m, for_usage, norm)
		*self = m
	case MetaSuffix:
		normalize(&m.m, for_usage, norm)
		*self = m
	}
}

func MetaFromItem(value Item) Meta {
	return MetaItem(value)
}

func alts(self Meta, to *[]Meta) {
	switch self := self.(type) {
	case MetaOr:
		*to = append(*to, self...)
	case MetaSkip:
	default:
		*to = append(*to, self)
	}
}

func or(self, other Meta) Meta {
	res := []Meta{}
	alts(self, &res)
	alts(other, &res)
	switch len(res) {
	case 0:
		return MetaSkip{}
	case 1:
		return res[0]
	default:
		return MetaOr(res)
	}
}

// collect different kinds of short names for disambiguation
func collect_shorts(self Meta, flags *[]byte, args *[]byte) {
	switch self := self.(type) {
	case MetaAnd:
		for _, x := range self {
			collect_shorts(x, flags, args)
		}
	case MetaOr:
		for _, x := range self {
			collect_shorts(x, flags, args)
		}
	case MetaItem:
		switch v := self.v.(type) {
		case ItemCommand:
			collect_shorts(v.meta, flags, args)
		case ItemFlag:
			*flags = append(*flags, v.shorts...)
		case ItemArgument:
			*args = append(*args, v.shorts...)
		}
	case MetaOptional:
		collect_shorts(self.m, flags, args)
	case MetaRequired:
		collect_shorts(self.m, flags, args)
	case MetaAdjacent:
		collect_shorts(self.m, flags, args)
	case MetaMany:
		collect_shorts(self.m, flags, args)
	case MetaSubsection:
		collect_shorts(self.m, flags, args)
	case MetaSuffix:
		collect_shorts(self.m, flags, args)
	case MetaCustomUsage:
		collect_shorts(self.m, flags, args)
	}
}
