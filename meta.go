package cli

import (
	"fmt"

	"github.com/rprtr258/fun"
)

type meta interface {
	isMeta()
	clone() meta
}

// metaAnd requires all arguments listed in the slice to be present.
type metaAnd []meta

func (metaAnd) isMeta() {}
func (self metaAnd) clone() meta {
	return metaAnd(fun.Map[meta](meta.clone, self...))
}

// metaOr requires one of the arguments listed in the slice to be present.
type metaOr []meta

func (metaOr) isMeta() {}
func (self metaOr) clone() meta {
	return metaOr(fun.Map[meta](meta.clone, self...))
}

// metaOptional renders the wrapped arguments as optional, encased in []
// when rendered.
type metaOptional struct{ m meta }

func (metaOptional) isMeta() {}
func (self metaOptional) clone() meta {
	return metaOptional{self.m.clone()}
}

// metaRequired renders the wrapped arguments as required, encased in ()
// when rendered.
type metaRequired struct{ m meta }

func (metaRequired) isMeta() {}
func (self metaRequired) clone() meta {
	return metaRequired{self.m.clone()}
}

// metaAdjacent requires the wrapped arguments to be adjacent to each other.
type metaAdjacent struct{ m meta }

func (metaAdjacent) isMeta() {}
func (self metaAdjacent) clone() meta {
	return metaAdjacent{self.m.clone()}
}

// metaItem is a primitive argument item.
type metaItem item

func (metaItem) isMeta() {}
func (self metaItem) clone() meta {
	return metaItem(item(self).clone())
}

// metaMany accepts multiple arguments.
type metaMany struct{ m meta }

func (metaMany) isMeta() {}
func (self metaMany) clone() meta {
	return metaMany{self.m.clone()}
}

// metaSubsection renders the arguments as a subsection with the doc as its
// header.
//
// The whole set of arguments goes into the same section as the first one.
type metaSubsection struct {
	m   meta
	doc Doc
}

func (metaSubsection) isMeta() {}
func (self metaSubsection) clone() meta {
	return metaSubsection{self.m.clone(), self.doc.clone()}
}

// metaSuffix renders the doc after the wrapped arguments.
type metaSuffix struct {
	m   meta
	doc Doc
}

func (metaSuffix) isMeta() {}
func (self metaSuffix) clone() meta {
	return metaSuffix{self.m.clone(), self.doc.clone()}
}

// metaSkip is not rendered in the help message.
type metaSkip struct{}

func (metaSkip) isMeta() {}
func (metaSkip) clone() meta {
	return metaSkip{}
}

// metaCustomUsage is a custom usage wrapper: a meta together with the Doc
// that replaces its derived usage line. The doc is always set where this
// variant is constructed (parse_usage).
type metaCustomUsage struct {
	m   meta
	doc Doc
}

func (metaCustomUsage) isMeta() {}
func (self metaCustomUsage) clone() meta {
	return metaCustomUsage{self.m.clone(), self.doc.clone()}
}

// metaStrict must be prefixed with -- in the usage group.
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

// strictNorm marks where [metaStrict] should bubble up to: top level, one of
// the "and" elements, or one of the "or" elements.
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

// is_command collapses duplicated commands, used by the normalization function.
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

// positional_invariant_check performs a nested invariant check.
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
		m = i.m
	}
	if _, ok := m.(metaOr); ok {
		m = metaRequired{m}
	}
	if norm == StrictNormPush {
		m = metaStrict{m}
	}
	return m
}

// metaFirstItem returns the front item of the meta; used by adjacent parsers
// since they inherit behavior of the front item.
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

// flatten_and flattens directly nested and-groups into a flat element list.
// Composition helpers like Map3 can nest them; strict `--` bubbling and
// rendering rely on the flat shape.
func flatten_and(m meta) []meta {
	if and, ok := m.(metaAnd); ok {
		res := make([]meta, 0, len(and))
		for _, x := range and {
			res = append(res, flatten_and(x)...)
		}
		return res
	}
	return []meta{m}
}

// normalize normalizes meta info for display as usage; required propagates
// outwards.
//
// Wrapping it in a pointer to the interface is intentional here.
//
//nolint:gocritic // `self *meta` is written through and must stay a pointer,
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
				// Strict bubbles up to one of the vec elements, the following
				// elements with their own Strict merge into the same `--`.
				(*xs)[i] = metaStrict{(*xs)[i]}
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
			return (*xs)[0], true
		default:
			return nil, false
		}
	}

	switch m := (*self).(type) {
	case metaAnd:
		flat := make([]meta, 0, len(m))
		for _, x := range m {
			flat = append(flat, flatten_and(x)...)
		}
		m = metaAnd(flat)
		if replacement, ok := normalize_vec((*[]meta)(&m), for_usage, norm, false); ok {
			*self = replacement
		} else {
			*self = m
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
			default:
				*self = metaRequired{m}
			}
		}
	case metaOptional:
		normalize(&m.m, for_usage, norm)
		*self = m // sync normalized inner into the stored value
		switch mm := m.m.(type) {
		case metaOptional:
			// Optional(Optional(m)) => Optional(m)
			*self = metaOptional{mm.m}
		case metaRequired:
			// Optional(Required(m)) => Optional(m)
			*self = metaOptional{mm.m}
		case metaMany:
			// Optional(Many(Required(m))) => Many(Optional(m))
			if x, ok := mm.m.(metaRequired); ok {
				*self = metaMany{metaOptional{x.m}}
			}
		case metaSkip:
			// Optional(Skip) => Skip
			*self = metaSkip{}
		}
	case metaRequired:
		normalize(&m.m, for_usage, norm)
		*self = m // sync normalized inner into the stored value
		switch {
		case isSkip(m.m):
			// Required(Skip) => Skip
			*self = metaSkip{}
		case isAnd(m.m) || isOr(m.m):
			// keep () around composite parsers
		default:
			// and strip them elsewhere
			*self = m.m
		}
	case metaMany:
		normalize(&m.m, for_usage, norm)
		*self = m // sync normalized inner into the stored value
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
		*self = m // sync normalized inner into the stored value
		// strip CustomUsage if we are not in usage so writer can simply render it
		if for_usage {
			// empty custom usage doc hides the parser from the usage line
			if m.doc.is_empty() {
				*self = metaSkip{}
			}
		} else {
			*self = m.m
		}
	case metaStrict:
		normalize(&m.m, for_usage, norm)
		*self = m // sync normalized inner into the stored value
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

// collect_shorts collects different kinds of short names for disambiguation.
func collect_shorts(self meta, flags, args *[]byte) {
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
