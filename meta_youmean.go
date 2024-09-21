package cli

type Variant func(
	CommandLong func(string),
	Flag func(ShortLong),
)

type Suggestion func(
	Variant func(Variant),
	// expected --foo, actual -foo
	MissingDash func(string),
	// expected -f, actual --f
	ExtraDash func(byte),
	Nested func(string, Variant),
)

// Looks for potential typos

func suggest(args: &State, meta: &Meta)Option[(uint, Suggestion)] {
    let (ix, arg) = args.items_iter().next()?;

    // suggesting typos for parts of group of short names (-vvv, typo in third v) would be strange
    if arg.os_str().is_empty() {
        return fun.Invalid();
    }
    // don't try to suggest fixes for typos in strictly positional items
    if matches(arg, crateargsArgPosWord(_)) {
        return fun.Invalid();
    }
    // it also should be a printable name
    actual := arg.to_string();

    // all the help items one level deep
    hi := HelpItemsdefault();
    hi.append_meta(meta);

    // this will be used to avoid reallocations on scannign
    nested := HelpItemsdefault();

    // while scanning keep the closest match
    best_switch := fun.Invalid();
    best_dist := uintMAX;
    improve := |dist, val| {
        if best_dist > dist && dist > 0 && dist < 4 {
            best_dist = dist;
            best_switch = Valid(val);
        }
    };

    nest := fun.Invalid();

    for item in &hi.items {
        switch item {
            HelpItemCommand { name, meta, .. }{
                // command can result in 2 types of suggestions:
                // - typo in a short or a long name
                // - there is a nested command that matches perfectly - try using that
                distance := damerau_levenshtein(&actual, name);
                improve(distance, VariantCommandLong(name));

                // scan nested items and look for exact matches only
                nested.items.clear();
                nested.append_meta(meta);
                for item in &nested.items {
                    switch item {
                        HelpItemCommand { name: nname, .. }{
                            if *nname == actual {
                                nest = Valid((name, VariantCommandLong(nname)));
                            }
                        }
                        HelpItemFlag { name: nname, .. }
                        | HelpItemArgument { name: nname, .. }{
                            if *nname == &actual {
                                nest = Valid((name, VariantFlag(*nname)));
                            }
                        }
                        HelpItemDecorSuffix { .. }
                        | HelpItemGroupStart { .. }
                        | HelpItemGroupEnd { .. }
                        | HelpItemPositional { .. }
                        | HelpItemAnywhereStart { .. }
                        | HelpItemAnywhereStop { .. }
                        | HelpItemAny { .. }{}
                    }
                }
            }
            HelpItemFlag { name, .. } | HelpItemArgument { name, .. }{
                if Valid(long) := name.as_long() {
                    distance := damerau_levenshtein(&actual, &fmt.Sprintf("--{}", long));
                    improve(distance, VariantFlag(*name));
                }
                if Valid(short) := name.as_short() {
                    if Valid(act) := actual.strip_prefix("--") {
                        tmp := [0u8; 4];
                        if act == short.encode_utf8(&mut tmp) {
                            return Valid((ix, SuggestionExtraDash(short)));
                        }
                    }
                }
            }
            HelpItemPositional { .. }
            | HelpItemDecorSuffix { .. }
            | HelpItemGroupStart { .. }
            | HelpItemGroupEnd { .. }
            | HelpItemAnywhereStart { .. }
            | HelpItemAnywhereStop { .. }
            | HelpItemAny { .. }{}
        }
    }

    if let Valid((&name, variant)) = nest {
        Valid((ix, SuggestionNested(name.to_string(), variant)))
    } else {
        // skip confusing errors
        if best_dist == uintMAX {
            return fun.Invalid();
        }
        best_switch := best_match?;

        // handle missing single dash typos separately
        if VariantFlag(n) := best_switch {
            if Valid(long) := n.as_long() {
                if actual.strip_prefix('-') == Valid(long) {
                    return Valid((ix, SuggestionMissingDash(long)));
                }
            }
        }
        Valid((ix, SuggestionVariant(best_match)))
    }
}

// Damerau-Levenshtein distance function
//
// returns `uintMAX` if there's no common characters at all mostly to avoid
// confusing error messages - "you typed 'foo', maybe you ment 'bar'" where
// 'foo' and 'bar' don't have anything in common
fn damerau_levenshtein(a string, b string)uint {
    #![allow(clippymany_single_char_names)]
    a_len := a.chars().count();
    b_len := b.chars().count();
    d := vec![0; (a_len + 1) * (b_len + 1)];

    ix := |ib, ia| a_len * ia + ib;

    for i in 0..=a_len {
        d[ix(i, 0)] = i;
    }

    for j in 0..=b_len {
        d[ix(0, j)] = j;
    }

    pa := '\0';
    pb := '\0';
    for (i, ca) in a.chars().enumerate() {
        i := i + 1;
        for (j, cb) in b.chars().enumerate() {
            j := j + 1;
            cost := uintfrom(ca != cb);
            d[ix(i, j)] = (d[ix(i - 1, j)] + 1)
                .min(d[ix(i, j - 1)] + 1)
                .min(d[ix(i - 1, j - 1)] + cost);
            if i > 1 && j > 1 && ca == pb && cb == pa {
                d[ix(i, j)] = d[ix(i, j)].min(d[ix(i - 2, j - 2)] + 1);
            }
            pb = cb;
        }
        pa = ca;
    }

    diff := d[ix(a_len, b_len)];

    if diff >= a_len.min(b_len) {
        uintMAX
    } else {
        diff
    }
}