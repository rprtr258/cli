package cli

import (
	"fmt"

	"github.com/rprtr258/fun"
)

type meta interface {
	isMeta()
	clone() meta
}

// All arguments listed in a vector must be present
type metaAnd []meta

func (metaAnd) isMeta() {}
func (self metaAnd) clone() meta {
	return metaAnd(fun.Map[meta](meta.clone, self...))
}

// One of arguments listed in a vector must be present
type metaOr []meta

func (metaOr) isMeta() {}
func (self metaOr) clone() meta {
	return metaOr(fun.Map[meta](meta.clone, self...))
}

// Arguments are optional and encased in [] when rendered
type metaOptional struct{ m meta }

func (metaOptional) isMeta() {}
func (self metaOptional) clone() meta {
	return metaOptional{self.m.clone()}
}

// Arguments are requred and encased in () when rendered
type metaRequired struct{ m meta }

func (metaRequired) isMeta() {}
func (self metaRequired) clone() meta {
	return metaRequired{self.m.clone()}
}

// Argumens are required to be adjacent to each other
type metaAdjacent struct{ m meta }

func (metaAdjacent) isMeta() {}
func (self metaAdjacent) clone() meta {
	return metaAdjacent{self.m.clone()}
}

// Primitive argument as described
type metaItem item

func (metaItem) isMeta() {}
func (self metaItem) clone() meta {
	return metaItem(item(self).clone())
}

// Accepts multiple arguments
type metaMany struct{ m meta }

func (metaMany) isMeta() {}
func (self metaMany) clone() meta {
	return metaMany{self.m.clone()}
}

// Arguments form a subsection with buffer being it's header
//
// whole set of arguments go into the same section as the first one
type metaSubsection struct {
	m   meta
	doc Doc
}

func (metaSubsection) isMeta() {}
func (self metaSubsection) clone() meta {
	return metaSubsection{self.m.clone(), self.doc.clone()}
}

// Buffer is rendered after
type metaSuffix struct {
	m   meta
	doc Doc
}

func (metaSuffix) isMeta() {}
func (self metaSuffix) clone() meta {
	return metaSuffix{self.m.clone(), self.doc.clone()}
}

// This item is not rendered in the help message
type metaSkip struct{}

func (metaSkip) isMeta() {}
func (metaSkip) clone() meta {
	return metaSkip{}
}

// mirrors Rust's meta::CustomUsage(Box<meta>, Box<Doc>) - the doc is always set
// where this variant is constructed (parse_usage)
type metaCustomUsage struct {
	m   meta
	doc Doc
}

func (metaCustomUsage) isMeta() {}
func (self metaCustomUsage) clone() meta {
	return metaCustomUsage{self.m.clone(), self.doc.clone()}
}

// this meta must be prefixed with -- in unsage group
type metaStrict struct{ m meta }

func (metaStrict) isMeta() {}
func (self metaStrict) clone() meta {
	return metaStrict{self.m.clone()}
}

func defaultMeta() meta {
	return metaSkip{}
}

func cloneMeta(m meta) meta {
	cloneDoc := func(d Doc) Doc {
		return Doc{d.payload, append([]token(nil), d.tokens...)}
	}
	switch m := m.(type) {
	case metaAnd:
		return metaAnd(fun.Map[meta](cloneMeta, m...))
	case metaOr:
		return metaOr(fun.Map[meta](cloneMeta, m...))
	case metaOptional:
		return metaOptional{cloneMeta(m.m)}
	case metaRequired:
		return metaRequired{cloneMeta(m.m)}
	case metaAdjacent:
		return metaAdjacent{cloneMeta(m.m)}
	case metaItem:
		return metaItem(item(m).clone())
	case metaMany:
		return metaMany{cloneMeta(m.m)}
	case metaSubsection:
		return metaSubsection{cloneMeta(m.m), cloneDoc(m.doc)}
	case metaSuffix:
		return metaSuffix{cloneMeta(m.m), cloneDoc(m.doc)}
	case metaSkip:
		return metaSkip{}
	case metaCustomUsage:
		return metaCustomUsage{cloneMeta(m.m), cloneDoc(m.doc)}
	case metaStrict:
		return metaStrict{cloneMeta(m.m)}
	}
	panic("unreachable")
}

func isAnd(m meta) bool {
	_, ok := m.(metaAnd)
	return ok
}

func isOr(m meta) bool {
	_, ok := m.(metaOr)
	return ok
}

func isSkip(m meta) bool {
	_, ok := m.(metaSkip)
	return ok
}

// metaStrict should bubble up to one of 3 places:
// - top level
// - one of "and" elements
// - one of "or" elements
type strictNorm uint8

const (
	// starting at the top and looking for Strict inside
	StrictNormPull strictNorm = iota
	StrictNormPush
	// Already accounted for one, can strip the rest
	StrictNormStrip
)

func (self strictNorm) IsPull() bool  { return self == StrictNormPull }
func (self strictNorm) IsPush() bool  { return self == StrictNormPush }
func (self strictNorm) IsStrip() bool { return self == StrictNormStrip }

func (self *strictNorm) push() {
	if *self == StrictNormPull {
		*self = StrictNormPush
	}
}

// Used by normalization function to collapse duplicated commands.
// It seems to be fine to strip section information but not anything else
func is_command(self meta) bool {
	switch self := self.(type) {
	case metaItem:
		_, ok := self.v.(itemCommand)
		return ok
	case metaSubsection:
		return is_command(self.m)
	}
	return false
}

// do a nested invariant check
func positional_invariant_check(self meta, verbose bool) {
	var GO func(meta meta, is_pos *bool, v bool)
	GO = func(meta meta, is_pos *bool, v bool) {
		switch meta := meta.(type) {
		case metaAnd:
			for _, x := range meta {
				GO(x, is_pos, v)
			}
		case metaOr:
			out := *is_pos
			for _, x := range meta {
				this_pos := *is_pos
				GO(x, &this_pos, v)
				out = out || this_pos
			}
			*is_pos = out
		case metaOptional:
			GO(meta.m, is_pos, v)
		case metaRequired:
			GO(meta.m, is_pos, v)
		case metaAdjacent:
			if i, ok := metaFirstItem(meta.m).Unpack(); ok {
				if i.is_pos() {
					GO(meta.m, is_pos, v)
				} else {
					inner := false
					GO(meta.m, &inner, v)
				}
			}
		case metaItem:
			i := item(meta)
			switch {
			case *is_pos == i.is_pos():
			case *is_pos:
				panic(fmt.Sprintf(`cli usage BUG: all positional and command items must be placed in the right most position of the structure or tuple they are in but %#v breaks this rule. See the "Positional" documentation for details.`, i))
			case i.is_pos():
				*is_pos = true
			}
			m, ok := i.v.(itemCommand)
			if ok {
				command_pos := false
				if v {
					fmt.Printf("Checking\n%#v\n", m.meta)
				}
				GO(m.meta, &command_pos, v)
			}
		case metaMany:
			GO(meta.m, is_pos, v)
		case metaSubsection:
			GO(meta.m, is_pos, v)
		case metaSuffix:
			GO(meta.m, is_pos, v)
		case metaSkip:
		case metaCustomUsage:
			GO(meta.m, is_pos, v)
		case metaStrict:
			GO(meta.m, is_pos, v)
		}
	}
	if verbose {
		fmt.Printf("Checking\n%#v\n", self)
	}
	is_pos := false
	GO(self, &is_pos, verbose)
}

func normalized(m meta, for_usage bool) meta {
	m = cloneMeta(m)
	norm := StrictNormPull
	normalize(&m, for_usage, &norm)
	// strip outer () around meta unless inner
	if i, ok := m.(metaRequired); ok {
		m = i
	}
	if _, ok := m.(metaOr); ok {
		m = metaRequired{m}
	}
	if norm == StrictNormPush {
		m = metaStrict{m}
	}
	return m
}

// Used by adjacent parsers since it inherits behavior of the front item
func metaFirstItem(meta meta) Opt[item] {
	switch meta := meta.(type) {
	case metaAnd:
		return optFlatMap(first(meta), metaFirstItem)
	case metaItem:
		return Valid(item(meta))
	case metaOptional:
		return metaFirstItem(meta.m)
	case metaRequired:
		return metaFirstItem(meta.m)
	case metaAdjacent:
		return metaFirstItem(meta.m)
	case metaMany:
		return metaFirstItem(meta.m)
	case metaSubsection:
		return metaFirstItem(meta.m)
	case metaSuffix:
		return metaFirstItem(meta.m)
	case metaCustomUsage:
		return metaFirstItem(meta.m)
	case metaStrict:
		return metaFirstItem(meta.m)
	}
	return Invalid[item]()
}

// Normalize meta info for display as usage. Required propagates outwards
func normalize(self *meta, for_usage bool, norm *strictNorm) {
	normalize_vec := func(
		xs *[]meta,
		for_usage bool,
		norm *strictNorm,
		or bool,
	) (meta, bool) {
		final_norm := *norm
		for i := range *xs {
			this_norm := *norm
			normalize(&(*xs)[i], for_usage, &this_norm)
			var target *strictNorm = fun.IF(or, &final_norm, norm)
			switch {
			// switch (*target, this_norm) {
			case this_norm.IsPull() || target.IsStrip():
			case target.IsPull() && this_norm.IsPush():
				*self = metaStrict{*self}
				*target = StrictNormStrip
			default:
				*target = this_norm
			}
		}
		*xs = fun.Filter(func(m meta) bool { return !isSkip(m) }, *xs...)

		*norm = final_norm

		switch len(*xs) {
		case 0:
			return metaSkip{}, true
		case 1:
			res := (*xs)[0]
			*xs = (*xs)[1:]
			return res, true
		default:
			return nil, false
		}
	}

	switch m := (*self).(type) {
	case metaAnd:
		if replacement, ok := normalize_vec((*[]meta)(&m), for_usage, norm, false); ok {
			*self = replacement
		}
	// or should have either () or [] around it
	case metaOr:
		if replacement, ok := normalize_vec((*[]meta)(&m), for_usage, norm, true); ok {
			*self = replacement
		} else {
			saw_cmd := false
			// drop all the commands apart from the first one
			m = fun.Filter(func(m meta) bool {
				is_cmd := is_command(m)
				keep := !(is_cmd && saw_cmd)
				saw_cmd = saw_cmd || is_cmd
				return keep
			}, m...)
			switch len(m) {
			case 0:
				*self = metaSkip{}
			case 1:
				*self = m[0]
				m = m[1:]
			default:
				*self = metaRequired{m}
			}
		}
	case metaOptional:
		normalize(&m.m, for_usage, norm)
		// *self = m
		switch mm := m.m.(type) {
		case metaOptional:
			*self = mm.m
		case metaRequired:
			*self = mm.m
		case metaMany:
			if x, ok := mm.m.(metaRequired); ok {
				*self = metaMany(metaOptional(x))
			}
		case metaSkip:
			*self = metaSkip{}
		}
	case metaRequired:
		normalize(&m.m, for_usage, norm)
		switch {
		case isSkip(m.m):
			// Required(skip)skip
			*self = metaSkip{}
		case isAnd(m) || isOr(m):
			// keep () around composite parsers
		default:
			// and strip them elsewhere
			*self = m
		}
	case metaMany:
		normalize(&m.m, for_usage, norm)
		*self = m
		if isSkip(m.m) {
			*self = metaSkip{}
		}
	case metaItem:
		i := item(m)
		i.normalize(for_usage)
	case metaSkip:
		// nothing to do with items and skip just bubbles upwards
	case metaCustomUsage:
		normalize(&m.m, for_usage, norm)
		// strip CustomUsage if we are not in usage so writer can simply render it
		if for_usage {
			*self = m
		} else if m.doc.is_empty() {
			*self = metaSkip{}
		}
	case metaStrict:
		normalize(&m.m, for_usage, norm)
		norm.push()
		*self = m.m
	case metaAdjacent:
		normalize(&m.m, for_usage, norm)
		*self = m.m
	case metaSubsection:
		normalize(&m.m, for_usage, norm)
		*self = m.m
	case metaSuffix:
		normalize(&m.m, for_usage, norm)
		*self = m.m
	}
}

func metaFromItem(value item) meta {
	return metaItem(value)
}

func alts(self meta, to *[]meta) {
	switch self := self.(type) {
	case metaOr:
		*to = append(*to, self...)
	case metaSkip:
	default:
		*to = append(*to, self)
	}
}

func or(self, other meta) meta {
	res := []meta{}
	alts(self, &res)
	alts(other, &res)
	switch len(res) {
	case 0:
		return metaSkip{}
	case 1:
		return res[0]
	default:
		return metaOr(res)
	}
}

// collect different kinds of short names for disambiguation
func collect_shorts(self meta, flags *[]byte, args *[]byte) {
	switch self := self.(type) {
	case metaAnd:
		for _, x := range self {
			collect_shorts(x, flags, args)
		}
	case metaOr:
		for _, x := range self {
			collect_shorts(x, flags, args)
		}
	case metaItem:
		switch v := self.v.(type) {
		case itemCommand:
			collect_shorts(v.meta, flags, args)
		case itemFlag:
			*flags = append(*flags, v.shorts...)
		case itemArgument:
			*args = append(*args, v.shorts...)
		}
	case metaOptional:
		collect_shorts(self.m, flags, args)
	case metaRequired:
		collect_shorts(self.m, flags, args)
	case metaAdjacent:
		collect_shorts(self.m, flags, args)
	case metaMany:
		collect_shorts(self.m, flags, args)
	case metaSubsection:
		collect_shorts(self.m, flags, args)
	case metaSuffix:
		collect_shorts(self.m, flags, args)
	case metaCustomUsage:
		collect_shorts(self.m, flags, args)
	}
}
