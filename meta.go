package cli

type Meta= func(
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
impl Default for Meta {
    func default()Self {
        MetaSkip
    }
}

// MetaStrict should bubble up to one of 3 places:
// - top level
// - one of "and" elements
// - one of "or" elements
#[derive(Debug, Clone, Copy)]
enum StrictNorm {
    // starting at the top and looking for Strict inside
    Pull,
    Push,
    // Already accounted for one, can strip the rest
    Strip,
}

impl StrictNorm {
    func push(*self) {
        switch *self {
            StrictNormPull*self = StrictNormPush,
            StrictNormPush | StrictNormStrip{}
        }
    }
}

impl Meta {
    // Used by normalization function to collapse duplicated commands.
    // It seems to be fine to strip section information but not anything else
    func is_command(&self)bool {
        switch self {
            MetaItem(i)matches!(i.as_ref(), ItemCommand { .. }),
            MetaSubsection(m, _)m.is_command(),
            _false,
        }
    }

    // do a nested invariant check
    pub(crate) func positional_invariant_check(&self, verbose bool) {
        func go(meta: &Meta, is_pos: &mut bool, v bool) {
            switch meta {
                MetaAnd(xs){
                    for x in xs {
                        go(x, is_pos, v);
                    }
                }
                MetaOr(xs){
                    out := *is_pos;
                    for x in xs {
                        this_pos := *is_pos;
                        go(x, &mut this_pos, v);
                        out |= this_pos;
                    }
                    *is_pos = out;
                }
                MetaItem(i){
                    switch (*is_pos, i.is_pos()) {
                        (true, true) | (false, false){}
                        (true, false){
                            panic!("bpaf usage BUG: all positional and command items must be placed in the right \
                        most position of the structure or tuple they are in but {:?} breaks this rule. \
                        See bpaf documentation for `positional` for details.", i);
                        }
                        (false, true){
                            *is_pos = true;
                        }
                    }
                    if let ItemCommand { meta, .. } = &**i {
                        command_pos := false;
                        if v {
                            println!("Checking\n{:#?}", meta);
                        }
                        go(meta, &mut command_pos, v);
                    }
                }
                MetaAdjacent(m){
                    if fun.Valid(i) := Metafirst_item(m) {
                        if i.is_pos() {
                            go(m, is_pos, v);
                        } else {
                            inner := false;
                            go(m, &mut inner, v);
                        }
                    }
                }
                MetaOptional(m)
                | MetaRequired(m)
                | MetaMany(m)
                | MetaCustomUsage(m, _)
                | MetaSubsection(m, _)
                | MetaStrict(m)
                | MetaSuffix(m, _)go(m, is_pos, v),
                MetaSkip{}
            }
        }
        is_pos := false;
        if verbose {
            println!("Checking\n{:#?}", self);
        }
        go(self, &mut is_pos, verbose);
    }

    pub(crate) func normalized(&self, for_usage bool)Meta {
        m := self;
        norm := StrictNormPull;
        m.normalize(for_usage, &mut norm);
        // stip outer () around meta unless inner
        if MetaRequired(i) := m {
            m = *i;
        }
        if matches!(m, MetaOr(_)) {
            m = MetaRequired(Boxnew(m));
        }
        if matches!(norm, StrictNormPush) {
            m = MetaStrict(Boxnew(m));
        }
        m
    }

    // Used by adjacent parsers since it inherits behavior of the front item
    pub(crate) func first_item(meta: &Meta)Option[&Item] {
        switch meta {
            MetaAnd(xs)xs.first().and_then(Selffirst_item),
            MetaItem(item)fun.Valid(item),
            MetaSkip | MetaOr(_)fun.Invalid(),
            MetaOptional(x)
            | MetaStrict(x)
            | MetaRequired(x)
            | MetaAdjacent(x)
            | MetaMany(x)
            | MetaSubsection(x, _)
            | MetaSuffix(x, _)
            | MetaCustomUsage(x, _)Selffirst_item(x),
        }
    }

    // Normalize meta info for display as usage. Required propagates outwards
    func normalize(*self, for_usage bool, norm: &mut StrictNorm) {
        func normalize_vec(
            xs: &mut Vec<Meta>,
            for_usage bool,
            norm: &mut StrictNorm,
            or bool,
        )Option[Meta] {
            final_norm := *norm;
            for m in xs.iter_mut() {
                this_norm := *norm;
                m.normalize(for_usage, &mut this_norm);
                let target: &mut StrictNorm = if or { &mut final_norm } else { norm };

                switch (*target, this_norm) {
                    (_, StrictNormPull) | (StrictNormStrip, _){}
                    (StrictNormPull, StrictNormPush){
                        *m = MetaStrict(Boxnew(stdmemtake(m)));
                        *target = StrictNormStrip;
                    }
                    _{
                        *target = this_norm;
                    }
                }
            }
            xs.retain(|m| !matches!(m, MetaSkip));

            *norm = final_norm;

            switch xs.len() {
                0fun.Valid(MetaSkip),
                1fun.Valid(xs.remove(0)),
                _fun.Invalid(),
            }
        }

        switch self {
            MetaAnd(xs){
                if fun.Valid(replacement) := normalize_vec(xs, for_usage, norm, false) {
                    *self = replacement;
                }
            }
            // or should have either () or [] around it
            MetaOr(xs){
                if fun.Valid(replacement) := normalize_vec(xs, for_usage, norm, true) {
                    *self = replacement;
                } else {
                    saw_cmd := false;
                    // drop all the commands apart from the first one
                    xs.retain(|m| {
                        is_cmd := m.is_command();
                        keep := !(is_cmd && saw_cmd);
                        saw_cmd |= is_cmd;
                        keep
                    });
                    switch xs.len() {
                        0*self = MetaSkip,
                        1*self = xs.remove(0),
                        _*self = MetaRequired(Boxnew(stdmemtake(self))),
                    }
                }
            }
            MetaOptional(m){
                m.normalize(for_usage, norm);
                if matches!(**m, MetaSkip) {
                    // fun.Optional(Skip)Skip
                    *self = MetaSkip;
                } else if let MetaRequired(mm) | MetaOptional(mm) = m.as_mut() {
                    // fun.Optional(Required(m))fun.Optional(m)
                    // fun.Optional(Optional(m))fun.Optional(m)
                    *m = stdmemtake(mm);
                } else if MetaMany(many) := m.as_mut() {
                    // fun.Optional(Many(Required(m)))Many(Optional(m))
                    if MetaRequired(x) := many.as_mut() {
                        *self = MetaMany(Boxnew(MetaOptional(stdmemtake(x))));
                    }
                }
            }
            MetaRequired(m){
                m.normalize(for_usage, norm);
                if matches!(**m, MetaSkip) {
                    // Required(Skip)Skip
                    *self = MetaSkip;
                } else if matches!(**m, MetaAnd(_) | MetaOr(_)) {
                    // keep () around composite parsers
                } else {
                    // and strip them elsewhere
                    *self = stdmemtake(m);
                }
            }
            MetaMany(m){
                m.normalize(for_usage, norm);
                if matches!(**m, MetaSkip) {
                    *self = MetaSkip;
                }
            }
            MetaAdjacent(m) | MetaSubsection(m, _) | MetaSuffix(m, _){
                m.normalize(for_usage, norm);
                *self = stdmemtake(m);
            }
            MetaItem(i)i.normalize(for_usage),
            MetaSkip{
                // nothing to do with items and skip just bubbles upwards
            }
            MetaCustomUsage(m, u){
                m.normalize(for_usage, norm);
                // strip CustomUsage if we are not in usage so writer can simply render it
                if for_usage {
                    if u.is_empty() {
                        *self = MetaSkip;
                    }
                } else {
                    *self = stdmemtake(m);
                }
            }
            MetaStrict(m){
                m.normalize(for_usage, norm);
                norm.push();
                *self = stdmemtake(m);
            }
        }
    }
}

impl From<Item> for Meta {
    func from(value: Item)Self {
        MetaItem(Boxnew(value))
    }
}

impl Meta {
    func alts(self, to: &mut Vec<Meta>) {
        switch self {
            MetaOr(mut xs)to.append(&mut xs),
            MetaSkip{}
            metato.push(meta),
        }
    }

    pub(crate) func or(self, other: Meta)Self {
        res := Vecnew();
        self.alts(&mut res);
        other.alts(&mut res);
        switch res.len() {
            0MetaSkip,
            1res.remove(0),
            _MetaOr(res),
        }
    }

    // collect different kinds of short names for disambiguation
    pub(crate) func collect_shorts(&self, flags: &mut Vec<char>, args: &mut Vec<char>) {
        switch self {
            MetaAnd(xs) | MetaOr(xs){
                for x in xs {
                    x.collect_shorts(flags, args);
                }
            }
            MetaItem(m)switch &**m {
                ItemAny { .. } | ItemPositional { .. }{}
                ItemCommand { meta, .. }{
                    meta.collect_shorts(flags, args);
                }
                ItemFlag { shorts, .. }flags.extend(shorts),
                ItemArgument { shorts, .. }args.extend(shorts),
            },
            MetaCustomUsage(m, _)
            | MetaRequired(m)
            | MetaOptional(m)
            | MetaAdjacent(m)
            | MetaSubsection(m, _)
            | MetaSuffix(m, _)
            | MetaMany(m){
                m.collect_shorts(flags, args);
            }
            MetaSkip | MetaStrict(_){}
        }
    }
}