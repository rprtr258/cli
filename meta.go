package cli

import (
	"fmt"

	"github.com/rprtr258/fun"
)

type Meta func(
	// All arguments listed in a vector must be present
	And func([]Meta),
	// One of arguments listed in a vector must be present
	Or func([]Meta),
	// Arguments are fun.Optional and encased in [] when rendered
	Optional func(Meta),
	// Arguments are requred and encased in () when rendered
	Required func(Meta),
	// Argumens are required to be adjacent to each other
	Adjacent func(Meta),
	// Primitive argument as described
	Item func(Item),
	// Accepts multiple arguments
	Many func(Meta),
	// Arguments form a subsection with buffer being it's header
	//
	// whole set of arguments go into the same section as the first one
	Subsection func(Meta, Doc),
	// Buffer is rendered after
	Suffix func(Meta, Doc),
	// This item is not rendered in the help message
	Skip func(),
	// TODO make it Option[Doc]
	CustomUsage func(Meta, Doc),
	// this meta must be prefixed with -- in unsage group
	Strict func(Meta),
)

func MetaAnd(xs ...Meta) Meta {
	return func(And, _ func([]Meta), _, _, _ func(Meta), _ func(Item), _ func(Meta), _, _ func(Meta, Doc), _ func(), _ func(Meta, Doc), _ func(Meta)) {
		if And != nil {
			And(xs)
		}
	}
}

func MetaOr(xs []Meta) Meta {
	return func(_, Or func([]Meta), _, _, _ func(Meta), _ func(Item), _ func(Meta), _, _ func(Meta, Doc), _ func(), _ func(Meta, Doc), _ func(Meta)) {
		if Or != nil {
			Or(xs)
		}
	}
}

func MetaOptional(m Meta) Meta {
	return func(_, _ func([]Meta), Optional, _, _ func(Meta), _ func(Item), _ func(Meta), _, _ func(Meta, Doc), _ func(), _ func(Meta, Doc), _ func(Meta)) {
		if Optional != nil {
			Optional(m)
		}
	}
}

func MetaRequired(m Meta) Meta {
	return func(_, _ func([]Meta), _, Required, _ func(Meta), _ func(Item), _ func(Meta), _, _ func(Meta, Doc), _ func(), _ func(Meta, Doc), _ func(Meta)) {
		if Required != nil {
			Required(m)
		}
	}
}

func MetaAdjacent(m Meta) Meta {
	return func(_, _ func([]Meta), _, _, Adjacent func(Meta), _ func(Item), _ func(Meta), _, _ func(Meta, Doc), _ func(), _ func(Meta, Doc), _ func(Meta)) {
		if Adjacent != nil {
			Adjacent(m)
		}
	}
}

func MetaItem(i Item) Meta {
	return func(_, _ func([]Meta), _, _, _ func(Meta), Item func(Item), _ func(Meta), _, _ func(Meta, Doc), _ func(), _ func(Meta, Doc), _ func(Meta)) {
		if Item != nil {
			Item(i)
		}
	}
}

func MetaMany(m Meta) Meta {
	return func(_, _ func([]Meta), _, _, _ func(Meta), _ func(Item), Many func(Meta), _, _ func(Meta, Doc), _ func(), _ func(Meta, Doc), _ func(Meta)) {
		if Many != nil {
			Many(m)
		}
	}
}

func MetaSubsection(m Meta, d Doc) Meta {
	return func(_, _ func([]Meta), _, _, _ func(Meta), _ func(Item), _ func(Meta), f, _ func(Meta, Doc), _ func(), _ func(Meta, Doc), _ func(Meta)) {
		if f != nil {
			f(m, d)
		}
	}
}

func MetaSuffix(m Meta, d Doc) Meta {
	return func(_, _ func([]Meta), _, _, _ func(Meta), _ func(Item), _ func(Meta), _, f func(Meta, Doc), _ func(), _ func(Meta, Doc), _ func(Meta)) {
		if f != nil {
			f(m, d)
		}
	}
}

var MetaSkip Meta = func(_, _ func([]Meta), _, _, _ func(Meta), _ func(Item), _ func(Meta), _, _ func(Meta, Doc), Skip func(), _ func(Meta, Doc), _ func(Meta)) {
	if Skip != nil {
		Skip()
	}
}

func MetaCustomUsage(m Meta, d Doc) Meta {
	return func(_, _ func([]Meta), _, _, _ func(Meta), _ func(Item), _ func(Meta), _, _ func(Meta, Doc), _ func(), f func(Meta, Doc), _ func(Meta)) {
		if f != nil {
			f(m, d)
		}
	}
}

func MetaStrict(m Meta) Meta {
	return func(_, _ func([]Meta), _, _, _ func(Meta), _ func(Item), _ func(Meta), _, _ func(Meta, Doc), _ func(), _ func(Meta, Doc), Strict func(Meta)) {
		if Strict != nil {
			Strict(m)
		}
	}
}

// to get stdmemtake to work
func DefaultMeta() Meta {
	return MetaSkip
}

func (m Meta) IsAnd() bool {
	var res bool
	m(func([]Meta) { res = true }, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil)
	return res
}

func (m Meta) IsOr() bool {
	var res bool
	m(nil, func([]Meta) { res = true }, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil)
	return res
}

func (m Meta) IsSkip() bool {
	var res bool
	m(nil, nil, nil, nil, nil, nil, nil, nil, nil, func() { res = true }, nil, nil)
	return res
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
func (self Meta) is_command() bool {
	res := false
	self(
		nil,
		nil,
		nil,
		nil,
		nil,
		func(i Item) {
			i.v(nil, nil, func(string, Option[byte], Meta, Info) { res = true }, nil, nil)
		},
		nil,
		func(m Meta, _ Doc) { res = m.is_command() },
		nil,
		nil,
		nil,
		nil,
	)
	return res
}

// do a nested invariant check
func (self Meta) positional_invariant_check(verbose bool) {
	var GO func(meta Meta, is_pos *bool, v bool)
	GO = func(meta Meta, is_pos *bool, v bool) {
		meta(
			func(xs []Meta) {
				for _, x := range xs {
					GO(x, is_pos, v)
				}
			},
			func(xs []Meta) {
				out := *is_pos
				for _, x := range xs {
					this_pos := *is_pos
					GO(x, &this_pos, v)
					out = out || this_pos
				}
				*is_pos = out
			},
			func(m Meta) { GO(m, is_pos, v) },
			func(m Meta) { GO(m, is_pos, v) },
			func(m Meta) {
				if i, ok := Metafirst_item(m).Unpack(); ok {
					if i.is_pos() {
						GO(m, is_pos, v)
					} else {
						inner := false
						GO(m, &inner, v)
					}
				}
			},
			func(i Item) {
				switch {
				case *is_pos == i.is_pos():
				case *is_pos:
					panic(fmt.Sprintf(`bpaf usage BUG: all positional and command items must be placed in the right most position of the structure or tuple they are in but %#v breaks this rule. See bpaf documentation for "positional" for details.`, i))
				case i.is_pos():
					*is_pos = true
				}
				i.v(nil, nil, func(_ string, _ Option[byte], meta Meta, _ Info) {
					command_pos := false
					if v {
						fmt.Printf("Checking\n%#v\n", meta)
					}
					GO(meta, &command_pos, v)
				}, nil, nil)
			},
			func(m Meta) { GO(m, is_pos, v) },
			func(m Meta, _ Doc) { GO(m, is_pos, v) },
			func(m Meta, _ Doc) { GO(m, is_pos, v) },
			func() {},
			func(m Meta, _ Doc) { GO(m, is_pos, v) },
			func(m Meta) { GO(m, is_pos, v) },
		)
	}
	is_pos := false
	if verbose {
		fmt.Printf("Checking\n%#v\n", self)
	}
	GO(self, &is_pos, verbose)
}

func (self Meta) normalized(for_usage bool) Meta {
	m := self
	norm := StrictNormPull
	m.normalize(for_usage, &norm)
	// stip outer () around meta unless inner
	m(nil, nil, nil, func(i Meta) { m = i }, nil, nil, nil, nil, nil, nil, nil, nil)
	m(nil, func([]Meta) { m = MetaRequired(m) }, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil)
	if norm == StrictNormPush {
		m = MetaStrict(m)
	}
	return m
}

// Used by adjacent parsers since it inherits behavior of the front item
func Metafirst_item(meta Meta) Option[Item] {
	res := Invalid[Item]()
	meta(
		func(xs []Meta) { res = OptFlatMap(First(xs), Metafirst_item) },
		nil,
		func(x Meta) { res = Metafirst_item(x) },
		func(x Meta) { res = Metafirst_item(x) },
		func(x Meta) { res = Metafirst_item(x) },
		nil,
		func(x Meta) { res = Metafirst_item(x) },
		func(x Meta, _ Doc) { res = Metafirst_item(x) },
		func(x Meta, _ Doc) { res = Metafirst_item(x) },
		nil,
		func(x Meta, _ Doc) { res = Metafirst_item(x) },
		func(x Meta) { res = Metafirst_item(x) },
	)
	return res
}

// Normalize meta info for display as usage. Required propagates outwards
func (self *Meta) normalize(for_usage bool, norm *StrictNorm) {
	normalize_vec := func(
		xs *[]Meta,
		for_usage bool,
		norm *StrictNorm,
		or bool,
	) (Meta, bool) {
		final_norm := *norm
		for i := range *xs {
			this_norm := *norm
			(*xs)[i].normalize(for_usage, &this_norm)
			var target *StrictNorm = fun.IF(or, &final_norm, norm)
			switch {
			// switch (*target, this_norm) {
			case this_norm.IsPull() || target.IsStrip():
			case target.IsPull() && this_norm.IsPush():
				*self = MetaStrict(*self)
				*target = StrictNormStrip
			default:
				*target = this_norm
			}
		}
		*xs = fun.Filter(func(m Meta) bool { return !m.IsSkip() }, *xs...)

		*norm = final_norm

		switch len(*xs) {
		case 0:
			return MetaSkip, true
		case 1:
			res := (*xs)[0]
			*xs = (*xs)[1:]
			return res, true
		default:
			return nil, false
		}
	}

	(*self)(
		func(xs []Meta) {
			if replacement, ok := normalize_vec(&xs, for_usage, norm, false); ok {
				*self = replacement
			}
		},
		// or should have either () or [] around it
		func(xs []Meta) {
			if replacement, ok := normalize_vec(&xs, for_usage, norm, true); ok {
				*self = replacement
			} else {
				saw_cmd := false
				// drop all the commands apart from the first one
				xs = fun.Filter(func(m Meta) bool {
					is_cmd := m.is_command()
					keep := !(is_cmd && saw_cmd)
					saw_cmd = saw_cmd || is_cmd
					return keep
				}, xs...)
				switch len(xs) {
				case 0:
					*self = MetaSkip
				case 1:
					*self = xs[0]
					xs = xs[1:]
				default:
					*self = MetaRequired(*self)
				}
			}
		},
		func(m Meta) {
			m.normalize(for_usage, norm)
			m(
				nil,
				nil,
				// TODO: mutate issue
				func(mm Meta) { /* *m = mm*/ },
				func(mm Meta) { /* *m = mm*/ },
				nil,
				nil,
				func(many Meta) {
					many(nil, nil, nil, func(x Meta) { *self = MetaMany(MetaOptional(x)) }, nil, nil, nil, nil, nil, nil, nil, nil)
				},
				nil,
				nil,
				func() { *self = MetaSkip },
				nil,
				nil,
			)
		},
		func(m Meta) {
			m.normalize(for_usage, norm)
			switch {
			case m.IsSkip():
				// Required(Skip)Skip
				*self = MetaSkip
			case m.IsAnd() || m.IsOr():
				// keep () around composite parsers
			default:
				// and strip them elsewhere
				*self = m
			}
		},
		func(m Meta) {
			m.normalize(for_usage, norm)
			*self = m
		},
		func(i Item) { i.normalize(for_usage) },
		func(m Meta) {
			m.normalize(for_usage, norm)
			if m.IsSkip() {
				*self = MetaSkip
			}
		},
		func(m Meta, _ Doc) {
			m.normalize(for_usage, norm)
			*self = m
		},
		func(m Meta, _ Doc) {
			m.normalize(for_usage, norm)
			*self = m
		},
		nil, // nothing to do with items and skip just bubbles upwards
		func(m Meta, u Doc) {
			m.normalize(for_usage, norm)
			// strip CustomUsage if we are not in usage so writer can simply render it
			if for_usage {
				if u.is_empty() {
					*self = MetaSkip
				}
			} else {
				*self = m
			}
		},
		func(m Meta) {
			m.normalize(for_usage, norm)
			norm.push()
			*self = m
		},
	)
}

func MetaFromItem(value Item) Meta {
	return MetaItem(value)
}

func (self Meta) alts(to *[]Meta) {
	self(
		func([]Meta) { *to = append(*to, self) },
		func(xs []Meta) { *to = append(*to, xs...) },
		func(Meta) { *to = append(*to, self) },
		func(Meta) { *to = append(*to, self) },
		func(Meta) { *to = append(*to, self) },
		func(Item) { *to = append(*to, self) },
		func(Meta) { *to = append(*to, self) },
		func(Meta, Doc) { *to = append(*to, self) },
		func(Meta, Doc) { *to = append(*to, self) },
		nil,
		func(Meta, Doc) { *to = append(*to, self) },
		func(Meta) { *to = append(*to, self) },
	)
}

func (self Meta) or(other Meta) Meta {
	res := []Meta{}
	self.alts(&res)
	other.alts(&res)
	switch len(res) {
	case 0:
		return MetaSkip
	case 1:
		return res[0]
	default:
		return MetaOr(res)
	}
}

// collect different kinds of short names for disambiguation
func (self Meta) collect_shorts(flags *[]byte, args *[]byte) {
	self(
		func(xs []Meta) {
			for _, x := range xs {
				x.collect_shorts(flags, args)
			}
		},
		func(xs []Meta) {
			for _, x := range xs {
				x.collect_shorts(flags, args)
			}
		},
		func(m Meta) { m.collect_shorts(flags, args) },
		func(m Meta) { m.collect_shorts(flags, args) },
		func(m Meta) { m.collect_shorts(flags, args) },
		func(m Item) {
			m.v(
				nil, nil,
				func(_ string, _ Option[byte], meta Meta, _ Info) {
					meta.collect_shorts(flags, args)
				},
				func(_ ShortLong, shorts []byte, _ Option[string]) {
					*flags = append(*flags, shorts...)
				},
				func(_ ShortLong, shorts []byte, _ Metavar, _ Option[string]) {
					*args = append(*args, shorts...)
				},
			)
		},
		func(m Meta) { m.collect_shorts(flags, args) },
		func(m Meta, _ Doc) { m.collect_shorts(flags, args) },
		func(m Meta, _ Doc) { m.collect_shorts(flags, args) },
		nil,
		func(m Meta, _ Doc) { m.collect_shorts(flags, args) },
		nil,
	)
}
