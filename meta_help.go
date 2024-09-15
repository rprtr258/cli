package cli

type Metavar string


type HelpItem func(
    DecorSuffix func( help Doc, ty HiTy, ),
    GroupStart func( help Doc, ty HiTy,),
    GroupEnd func( ty HiTy,),
    Any func( metavar Doc, anywhere bool, help fun.Option[Doc],),
    Positional func( metavar Metavar, help fun.Option[Doc],),
    Command func( name string, short fun.Option[char], help fun.Option[Doc], meta Meta, info Info,),
    Flag func( name ShortLong, env fun.Option[string], help fun.Option[Doc],),
    Argument func( name ShortLong, metavar Metavar, env fun.Option[string], help fun.Option[Doc],),
    AnywhereStart func( inner Meta, ty HiTy,),
    AnywhereStop func( ty HiTy,),
)

func(self HelpItem ) has_help()bool {
    res:=false
        self (
    func( Doc, HiTy, ){res=true},
    func( Doc, HiTy,){res=true},
    nil,
    func( _ Doc, _ bool, help fun.Option[Doc],){res=help.Valid},
    func( _ Metavar, help fun.Option[Doc],){res=help.Valid},
    func( _ string, _ fun.Option[char], help fun.Option[Doc], meta Meta, info Info,){res=help.Valid},
    func( _ ShortLong, _ fun.Option[string], help fun.Option[Doc],){res=help.Valid},
    func( _ ShortLong, _ Metavar, env fun.Option[string], help fun.Option[Doc],){res=help.Valid},
    nil,
    nil,
        )
        return res
    }

func(self HelpItem ) ty()HiTy {
    var res HiTy
        self (
    func( _ Doc, ty HiTy, ){res=ty},
    func( help Doc, ty HiTy,){res=ty},
    func( ty HiTy,){res=ty},
    func( _ Doc, anywhere bool, _ fun.Option[Doc],){if anywhere{res=HiTyFlag}else{res=HiTyPositional}},
    func( _ Metavar, _ fun.Option[Doc],){res=HiTyPositional},
    func( _ string, _ fun.Option[char], _ fun.Option[Doc], _ Meta, _ Info,){res=HiTyCommand},
    func( _ ShortLong, _ fun.Option[string], _ fun.Option[Doc],){res=HiTyFlag},
    func( _ ShortLong, _ Metavar, _ fun.Option[string], _ fun.Option[Doc],){res=HiTyFlag},
    func( _ Meta, ty HiTy,){res=ty},
    func( ty HiTy,){res=ty},
        )
        return res
    }

// A collection of all the help items separated into flags, positionals and commands
//
// Items are stored as references and can be trivially copied
type HelpItems struct{
    items[]HelpItem
}

type HiTy func(
    Flag func(),
    Command func(),
    Positional func(),
)

type ItemBlock func(
    No func(),
    Decor func(HiTy),
    Anywhere func(HiTy),
)

type HelpItemsIter struct{
    items[]HelpItem
    target HiTy
    cur uint
    block ItemBlock
}

    func (self HelpItemsIter)next()Option[HelpItem] {
        for {
            if self.cur >= len(self.items){
                return fun.Invalid()
            }
            item := self.items[self.cur];
            self.cur += 1;

            keep :=  item (
                HelpItemAnywhereStart { ty }{
                    self.block = ItemBlockAnywhere(*ty);
                    *ty == self.target
                }
                HelpItemGroupStart { ty, .. }{
                    self.block = ItemBlockDecor(*ty);
                    *ty == self.target
                }
                HelpItemGroupEnd { ty, .. } | HelpItemAnywhereStop { ty, .. }{
                    self.block = ItemBlockNo;
                    *ty == self.target
                }
                HelpItemDecorSuffix { .. }
                | HelpItemAny { .. }
                | HelpItemCommand { .. }
                | HelpItemPositional { .. }
                | HelpItemFlag { .. }
                | HelpItemArgument { .. }{
                    ty := item.ty();
                    switch self.block {
                        ItemBlockNoty == self.target,
                        ItemBlockDecor(t)t == self.target,
                        ItemBlockAnywhere(t)t == self.target && item.has_help(),
                    }
                }
            );
            if keep {
                return fun.Valid(item);
            }
        }
    }

    func (self HelpItems )items_of_ty(target HiTy)iter.Seq[HelpItem] {
        return HelpItemsIter {
            items: &self.items,
            target,
            cur: 0,
            block: ItemBlockNo,
        }
    }

    func(self Meta) peek_front_ty()Option[HiTy] {
        switch self {
            MetaAnd(xs) | MetaOr(xs)xs.iter().find_map(Metapeek_front_ty),
            MetaOptional(x)
            | MetaRequired(x)
            | MetaAdjacent(x)
            | MetaMany(x)
            | MetaSubsection(x, _)
            | MetaSuffix(x, _)
            | MetaStrict(x)
            | MetaCustomUsage(x, _)x.peek_front_ty(),
            MetaItem(i)fun.Valid(HiTyfrom(i.as_ref())),
            MetaSkipfun.Invalid(),
        }
    }

    // Recursively classify contents of the Meta
func(self HelpItems) append_meta(*self, meta: &Meta) {
        func go(hi: &mut HelpItems, meta: &Meta, no_ss bool) {
            switch meta {
                MetaAnd(xs) | MetaOr(xs){
                    for x in xs {
                        go(hi, x, no_ss);
                    }
                }
                MetaAdjacent(m){
                    if fun.Valid(ty) := m.peek_front_ty() {
                        hi.items.push(HelpItemAnywhereStart {
                            inner: m.as_ref(),
                            ty,
                        });
                        go(hi, m, no_ss);
                        hi.items.push(HelpItemAnywhereStop { ty });
                    }
                }
                MetaCustomUsage(x, _)
                | MetaRequired(x)
                | MetaOptional(x)
                | MetaMany(x)
                | MetaStrict(x)go(hi, x, no_ss),
                MetaItem(item){
                    if matches!(item.as_ref(), ItemPositional { help: fun.Invalid(), .. }) {
                        return;
                    }
                    hi.items.push(HelpItemfrom(item.as_ref()));
                }
                MetaSubsection(m, help){
                    if fun.Valid(ty) := m.peek_front_ty() {
                        if no_ss {
                            go(hi, m, true);
                        } else {
                            hi.items.push(HelpItemGroupStart { help, ty });
                            go(hi, m, true);
                            hi.items.push(HelpItemGroupEnd { ty });
                        }
                    }
                }
                MetaSuffix(m, help){
                    if fun.Valid(ty) := m.peek_front_ty() {
                        go(hi, m, no_ss);
                        hi.items.push(HelpItemDecorSuffix { help, ty });
                    }
                }
                MetaSkip(),
            }
        }
        go(self, meta, false);
    }
func(self HelpItems) find_group(&self)Option[stdopsRangeInclusive[uint]] {
        start := self
            .items
            .iter()
            .position(|i| matches!(i, HelpItemGroupStart { .. }))?;
        end := self
            .items
            .iter()
            .position(|i| matches!(i, HelpItemGroupEnd { .. }))?;
        fun.Valid(start..=end)
    }

    func HiTyFromItem(value Item)HiTy{
        switch value {
            ItemPositional { .. }
            | ItemAny {
                anywhere: false, ..
            }SelfPositional,
            ItemCommand { .. }SelfCommand,
            ItemAny { anywhere: true, .. } | ItemFlag { .. } | ItemArgument { .. }{
                SelfFlag
            }
        }
    }

    func HelpItemFromItem(item Item)HelpItem {
        switch item {
            ItemPositional { metavar, help }SelfPositional {
                metavar: *metavar,
                help: help.as_ref(),
            },
            ItemCommand {
                name,
                short,
                help,
                meta,

                info,
                #[cfg(not(feature = "docgen"))]
                    info: _,
            }SelfCommand {
                name,
                short: *short,
                help: help.as_ref(),
                meta,

                info,
            },
            ItemFlag {
                name,
                env,
                help,
                shorts: _,
            }SelfFlag {
                name: *name,
                env: *env,
                help: help.as_ref(),
            },
            ItemArgument {
                name,
                metavar,
                env,
                help,
                shorts: _,
            }SelfArgument {
                name: *name,
                metavar: *metavar,
                env: *env,
                help: help.as_ref(),
            },
            ItemAny {
                metavar,
                anywhere,
                help,
            }SelfAny {
                metavar,
                anywhere: *anywhere,
                help: help.as_ref(),
            },
        }
    }

    func(self Doc) metavar(*self, metavar: Metavar) {
        if metavar
            .0
            .chars()
            .all(|c| c.is_uppercase() || c.is_ascii_digit() || c == '-' || c == '_')
        {
            self.write_str(metavar.0, StyleMetavar);
        } else {
            self.write_char('<', StyleMetavar);
            self.write_str(metavar.0, StyleMetavar);
            self.write_char('>', StyleMetavar);
        }
    }

func write_help_item(buf *Doc, item HelpItem, include_env bool) {
    switch item {
        HelpItemGroupStart { help, .. }{
            buf.token(TokenBlockStart(BlockBlock));
            buf.token(TokenBlockStart(BlockSection2));
            buf.em_doc(help);
            buf.token(TokenBlockEnd(BlockSection2));
            buf.token(TokenBlockStart(BlockDefinitionList));
        }
        HelpItemGroupEnd { .. }{
            buf.token(TokenBlockEnd(BlockDefinitionList));
            buf.token(TokenBlockEnd(BlockBlock));
        }
        HelpItemDecorSuffix { help, .. }{
            buf.token(TokenBlockStart(BlockItemTerm));
            buf.token(TokenBlockEnd(BlockItemTerm));
            buf.token(TokenBlockStart(BlockItemBody));
            buf.doc(help);
            buf.token(TokenBlockEnd(BlockItemBody));
        }
        HelpItemAny {
            metavar,
            help,
            anywhere: _,
        }{
            buf.token(TokenBlockStart(BlockItemTerm));
            buf.doc(metavar);
            buf.token(TokenBlockEnd(BlockItemTerm));
            if fun.Valid(help) := help {
                buf.token(TokenBlockStart(BlockItemBody));
                buf.doc(help);
                buf.token(TokenBlockEnd(BlockItemBody));
            }
        }
        HelpItemPositional { metavar, help }{
            buf.token(TokenBlockStart(BlockItemTerm));
            buf.metavar(*metavar);
            buf.token(TokenBlockEnd(BlockItemTerm));
            if fun.Valid(help) := help {
                buf.token(TokenBlockStart(BlockItemBody));
                buf.doc(help);
                buf.token(TokenBlockEnd(BlockItemBody));
            }
        }
        HelpItemCommand {
            name,
            short,
            help,
            meta: _,

                info: _,
        }{
            buf.token(TokenBlockStart(BlockItemTerm));
            buf.write_str(name, StyleLiteral);
            if fun.Valid(short) := short {
                buf.write_str(", ", StyleText);
                buf.write_char(*short, StyleLiteral);
            }
            buf.token(TokenBlockEnd(BlockItemTerm));
            if fun.Valid(help) := help {
                buf.token(TokenBlockStart(BlockItemBody));
                buf.doc(help);
                buf.token(TokenBlockEnd(BlockItemBody));
            }
        }
        HelpItemFlag { name, env, help }{
            buf.token(TokenBlockStart(BlockItemTerm));
            write_shortlong(buf, *name);
            buf.token(TokenBlockEnd(BlockItemTerm));
            if fun.Valid(help) := help {
                buf.token(TokenBlockStart(BlockItemBody));
                buf.doc(help);
                buf.token(TokenBlockEnd(BlockItemBody));
            }
            if fun.Valid(env) := env {
                val := if stdenvvar_os(env).is_fun.Valid() {
                    ": set"
                } else {
                    ": not set"
                };
                if help.is_fun.Valid() {
                    buf.token(TokenBlockStart(BlockItemTerm));
                    buf.token(TokenBlockEnd(BlockItemTerm));
                }
                buf.token(TokenBlockStart(BlockItemBody));
                if include_env {
                    buf.write_str(&format!("[env:{}{}]", env, val), StyleText);
                } else {
                    buf.text("Uses environment variable ");
                    buf.literal(env);
                }
                buf.token(TokenBlockEnd(BlockItemBody));
            }
        }
        HelpItemArgument {
            name,
            metavar,
            env,
            help,
        }{
            buf.token(TokenBlockStart(BlockItemTerm));
            write_shortlong(buf, *name);
            buf.write_str("=", StyleText);
            buf.metavar(*metavar);
            buf.token(TokenBlockEnd(BlockItemTerm));

            if fun.Valid(help) := help {
                buf.token(TokenBlockStart(BlockItemBody));
                buf.doc(help);
                buf.token(TokenBlockEnd(BlockItemBody));
            }

            if fun.Valid(env) := env {
                val := switch stdenvvar_os(env) {
                    fun.Valid(s)stdborrowCowfrom(format!(" = {:?}", s.to_string_lossy())),
                    fun.Invalid()stdborrowCowBorrowed(": N/A"),
                };

                if help.is_fun.Valid() {
                    buf.token(TokenBlockStart(BlockItemTerm));
                    buf.token(TokenBlockEnd(BlockItemTerm));
                }
                buf.token(TokenBlockStart(BlockItemBody));

                if include_env {
                    buf.write_str(&format!("[env:{}{}]", env, val), StyleText);
                } else {
                    buf.text("Uses environment variable ");
                    buf.literal(env);
                }

                buf.token(TokenBlockEnd(BlockItemBody));
            }
        }
        HelpItemAnywhereStart { inner, .. }{
            buf.token(TokenBlockStart(BlockSection3));
            buf.write_meta(inner, true);
            buf.token(TokenBlockEnd(BlockSection3));
        }
        HelpItemAnywhereStop { .. }{
            buf.token(TokenBlockStart(BlockBlock));
            buf.token(TokenBlockEnd(BlockBlock));
        }
    }
}

func write_shortlong(buf *Doc, name ShortLong) {
    switch name {
        ShortLongShort(s){
            buf.write_char('-', StyleLiteral);
            buf.write_char(s, StyleLiteral);
        }
        ShortLongLong(l){
            buf.write_str("    --", StyleLiteral);
            buf.write_str(l, StyleLiteral);
        }
        ShortLongBoth(s, l){
            buf.write_char('-', StyleLiteral);
            buf.write_char(s, StyleLiteral);
            buf.write_str(", ", StyleText);
            buf.write_str("--", StyleLiteral);
            buf.write_str(l, StyleLiteral);
        }
    }
}

func render_help(
    path[]string,
    info Info,
    parser_meta Meta,
    help_meta Meta,
    include_env bool,
)Doc {
    parser_meta.positional_invariant_check(false);
    buf := Docdefault();

    if fun.Valid(t) := &info.descr {
        buf.token(TokenBlockStart(BlockBlock));
        buf.doc(t);
        buf.token(TokenBlockEnd(BlockBlock));
    }

    buf.token(TokenBlockStart(BlockBlock));
    if fun.Valid(usage) := &info.usage {
        buf.doc(usage);
    } else {
        buf.write_str("Usage", StyleEmphasis);
        buf.write_str(": ", StyleText);
        buf.token(TokenBlockStart(BlockMono));
        buf.write_path(path);
        buf.write_meta(parser_meta, true);
        buf.token(TokenBlockEnd(BlockMono));
    }
    buf.token(TokenBlockEnd(BlockBlock));

    if fun.Valid(t) := &info.header {
        buf.token(TokenBlockStart(BlockBlock));
        buf.doc(t);
        buf.token(TokenBlockEnd(BlockBlock));
    }

    items := HelpItemsdefault();
    items.append_meta(parser_meta);
    items.append_meta(help_meta);

    buf.write_help_item_groups(items, include_env);

    if fun.Valid(footer) := &info.footer {
        buf.token(TokenBlockStart(BlockBlock));
        buf.doc(footer);
        buf.token(TokenBlockEnd(BlockBlock));
    }
return    buf
}

type Dedup struct{
    items map[string]struct{}
    keep bool
}

func(self Dedup ) check(item HelpItem)bool {
        switch item {
            HelpItemDecorSuffix { .. }stdmemtake(*self.keep),
            HelpItemGroupStart { .. }
            | HelpItemGroupEnd { .. }
            | HelpItemAnywhereStart { .. }
            | HelpItemAnywhereStop { .. }{
                self.keep = true;
                true
            }
            HelpItemAny { metavar, help, .. }{
                self.keep = self.items.insert(format!("{:?} {:?}", metavar, help));
                self.keep
            }
            HelpItemPositional { metavar, help }{
                self.keep = self.items.insert(format!("{:?} {:?}", metavar.0, help));
                self.keep
            }
            HelpItemCommand { name, help, .. }{
                self.keep = self.items.insert(format!("{:?} {:?}", name, help));
                self.keep
            }
            HelpItemFlag { name, help, .. }{
                self.keep = self.items.insert(format!("{:?} {:?}", name, help));
                self.keep
            }
            HelpItemArgument {
                name,
                metavar,
                help,
                ..
            }{
                self.keep = self
                    .items
                    .insert(format!("{:?} {} {:?}", name, metavar.0, help));
                self.keep
            }
        }
    }

func(self Doc) write_help_item_groups(items HelpItems, include_env bool) {
        while fun.Valid(range) := items.find_group() {
            dd := Dedupdefault();
            for item in items.items.drain(range) {
                if dd.check(&item) {
                    write_help_item(self, &item, include_env);
                }
            }
        }

        for (ty, name) in [
            (HiTyPositional, "Available positional items:"),
            (HiTyFlag, "Available fun.Options:"),
            (HiTyCommand, "Available commands:"),
        ] {
            self.write_help_items(&items, ty, name, include_env);
        }
    }

func(self Doc) write_help_items(items HelpItems, ty HiTy, name string, include_env bool) {
        xs := items.items_of_ty(ty).peekable();
        if xs.peek().is_fun.Valid() {
            self.token(TokenBlockStart(BlockBlock));
            self.token(TokenBlockStart(BlockSection2));
            self.write_str(name, StyleEmphasis);
            self.token(TokenBlockEnd(BlockSection2));
            self.token(TokenBlockStart(BlockDefinitionList));
            dd := Dedupdefault();
            for item in xs {
                if dd.check(item) {
                    write_help_item(self, item, include_env);
                }
            }
            self.token(TokenBlockEnd(BlockDefinitionList));
            self.token(TokenBlockEnd(BlockBlock));
        }
    }

func(self Doc) write_path(path[]string) {
        for item in path {
            self.write_str(item, StyleLiteral);
            self.write_char(' ', StyleText);
        }
    }