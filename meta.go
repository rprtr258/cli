package cli

import (
	"fmt"

	"github.com/rprtr258/fun"
)

type Meta = func(
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
	// TODO make it fun.Option[Doc]
	CustomUsage func(Meta, Doc),
	// this meta must be prefixed with -- in unsage group
	Strict func(Meta),
)

func MetaAnd(xs []Meta) Meta {
	return func(And, _ func([]Meta), _, _, _ func(Meta), _ func(Item), _ func(Meta), _, _, _, _ func(Meta, Doc), _ func(Meta)) {
		if And != nil {
			And(xs)
		}
	}
}

// to get stdmemtake to work
func DefaultMeta() Meta {
	return MetaSkip
}

// MetaStrict should bubble up to one of 3 places:
// - top level
// - one of "and" elements
// - one of "or" elements
type StrictNorm func(
	// starting at the top and looking for Strict inside
	Pull func(),
	Push func(),
	// Already accounted for one, can strip the rest
	Strip func(),
)

func (self *StrictNorm) push() {
	self(
		func() { *self = StrictNormPush },
		nil,
		nil,
	)
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
			i(nil, nil, func(string, fun.Option[byte], fun.Option[Doc], Meta, Info) { res = true }, nil, nil)
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
	GO := func(meta Meta, is_pos *bool, v bool) {
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
					out |= this_pos
				}
				*is_pos = out
			},
			func(Meta) { GO(m, is_pos, v) },
			func(Meta) { GO(m, is_pos, v) },
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
				i(nil, nil, func(_ string, _ fun.Option[byte], _ fun.Option[Doc], meta Meta, _ Info) {
					command_pos := false
					if v {
						fmt.Printf("Checking\n%#v\n", meta)
					}
					GO(meta, &command_pos, v)
				}, nil, nil)
			},
			func(Meta) { GO(m, is_pos, v) },
			func(Meta) { GO(m, is_pos, v) },
			func(Meta) { GO(m, is_pos, v) },
			func() {},
			func(Meta, Doc) { GO(m, is_pos, v) },
			func(Meta) { GO(m, is_pos, v) },
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
	m(nil, nil, nil, func(i Meta) { m = i }, nil, nil, nil, nil, nil, nil, nil)
	m(nil, func([]Meta) { m = MetaRequired(m) }, nil, nil, nil, nil, nil, nil, nil, nil, nil)
	norm(nil, func() { m = MetaStrict(m) }, nil)
	return m
}

// Used by adjacent parsers since it inherits behavior of the front item
func first_item(meta Meta) Option[Item] {
	res := fun.Invalid[Item]()
	meta(
		func(xs []Meta) { res = xs.first().and_then(Selffirst_item) },
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
		func(x Meta, _ Doc) { res = Metafirst_item(x) },
	)
	return res
}

// Normalize meta info for display as usage. Required propagates outwards
func (meta *Meta) normalize(for_usage bool, norm *StrictNorm) {
	normalize_vec := func(
		xs *[]Meta,
		for_usage bool,
		norm *StrictNorm,
		or bool,
	) Option[Meta] {
		final_norm := *norm
		for i := range *xs {
			this_norm := *norm
			xs[i].normalize(for_usage, &this_norm)
			var target *StrictNorm = fun.IF(or, &final_norm, norm)
			// switch (*target, this_norm) {
			//     (_, StrictNormPull) | (StrictNormStrip, _){}
			//     (StrictNormPull, StrictNormPush){
			//         *m = MetaStrict(stdmemtake(m));
			//         *target = StrictNormStrip;
			//     }
			//     _{
			//         *target = this_norm;
			//     }
			// }
		}
		xs = fun.Filter(xs, func(m) { return !matches(m, MetaSkip) })

		*norm = final_norm

		switch xs.len() {
		case 0:
			return fun.Valid(MetaSkip)
		case 1:
			return fun.Valid(xs.remove(0))
		default:
			return fun.Invalid()
		}
	}

	self(
		func(xs []Meta) {
			if replacement, ok := normalize_vec(xs, for_usage, norm, false); ok {
				*self = replacement
			}
		},
		// or should have either () or [] around it
		func(xs []Meta) {
			if replacement, ok := normalize_vec(xs, for_usage, norm, true); ok {
				*self = replacement
			} else {
				saw_cmd := false
				// drop all the commands apart from the first one
				xs = fun.Filter(func(m Meta) {
					is_cmd := m.is_command()
					keep := !(is_cmd && saw_cmd)
					saw_cmd |= is_cmd
					keep
				}, xs...)
				switch xs.len() {
				case 0:
					*self = MetaSkip
				case 1:
					*self = xs.remove(0)
				default:
					*self = MetaRequired(Boxnew(stdmemtake(self)))
				}
			}
		},
		func(m Meta) {
			m.normalize(for_usage, norm)
			m(
				nil,
				nil,
				func(mm Meta) { *m = stdmemtake(mm) },
				func(mm Meta) { *m = stdmemtake(mm) },
				nil,
				nil,
				func(many Meta) {
					many(nil, nil, nil, func(x Meta) { *self = MetaMany(Boxnew(MetaOptional(stdmemtake(x)))) }, nil, nil, nil, nil, nil)
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
			if matches(**m, MetaSkip) {
				// Required(Skip)Skip
				*self = MetaSkip
			} else if matches(**m, MetaAnd(_)|MetaOr(_)) {
				// keep () around composite parsers
			} else {
				// and strip them elsewhere
				*self = stdmemtake(m)
			}
		},
		func(m Meta) {
			m.normalize(for_usage, norm)
			*self = stdmemtake(m)
		},
		func(i Item) { i.normalize(for_usage) },
		func(m Meta) {
			m.normalize(for_usage, norm)
			if matches(**m, MetaSkip) {
				*self = MetaSkip
			}
		},
		func(m Meta, _ Doc) {
			m.normalize(for_usage, norm)
			*self = stdmemtake(m)
		},
		func(m Meta, _ Doc) {
			m.normalize(for_usage, norm)
			*self = stdmemtake(m)
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
				*self = stdmemtake(m)
			}
		},
		func(m Meta) {
			m.normalize(for_usage, norm)
			norm.push()
			*self = stdmemtake(m)
		},
	)
}

func MetaFromItem(value Item) Meta {
	return MetaItem(Boxnew(value))
}

func (self Meta) alts(to *[]Meta) {
	self(
		func() { to.push(meta) },
		func(xs) { to.append(&xs) },
		func() { to.push(meta) },
		func() { to.push(meta) },
		func() { to.push(meta) },
		func() { to.push(meta) },
		func() { to.push(meta) },
		func() { to.push(meta) },
		func() { to.push(meta) },
		nil,
		func() { to.push(meta) },
		func() { to.push(meta) },
	)
}

func (self Meta) or(other Meta) Self {
	res := Vecnew()
	self.alts(&res)
	other.alts(&res)
	switch res.len() {
	case 0:
		MetaSkip
	case 1:
		res.remove(0)
	default:
		MetaOr(res)
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
			m(
				nil, nil,
				func(_ string, _ fun.Option[byte], _ fun.Option[Doc], meta Meta, _ Info) {
					meta.collect_shorts(flags, args)
				},
				func(_ ShortLong, shorts []byte, _ fun.Option[string], _ fun.Option[Doc]) { flags.extend(shorts) },
				func(_ ShortLong, shorts []byte, _ Metavar, _ fun.Option[string], _ fun.Option[Doc]) {
					args.extend(shorts)
				},
			)
		},
		func(m Meta) { m.collect_shorts(flags, args) },
		func(m Meta, _ Doc) { m.collect_shorts(flags, args) },
		func(m Meta, _ Doc) { m.collect_shorts(flags, args) },
		nil,
		func(m Meta) { m.collect_shorts(flags, args) },
		nil,
	)
}
