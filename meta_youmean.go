package cli

import (
	"fmt"
	"math"
	"strings"

	"github.com/rprtr258/fun"
)

type Variant func(
	CommandLong func(string),
	Flag func(ShortLong),
)

func VariantCommandLong(name string) Variant {
	return func(CommandLong func(string), _ func(ShortLong)) {
		if CommandLong != nil {
			CommandLong(name)
		}
	}
}

func VariantFlag(name ShortLong) Variant {
	return func(_ func(string), Flag func(ShortLong)) {
		if Flag != nil {
			Flag(name)
		}
	}
}

type Suggestion func(
	Variant func(Variant),
	// expected --foo, actual -foo
	MissingDash func(string),
	// expected -f, actual --f
	ExtraDash func(byte),
	Nested func(string, Variant),
)

func SuggestionVariant(variant Variant) Suggestion {
	return func(Variant func(Variant), _ func(string), _ func(byte), _ func(string, Variant)) {
		if Variant != nil {
			Variant(variant)
		}
	}
}

func SuggestionMissingDash(long string) Suggestion {
	return func(_ func(Variant), MissingDash func(string), _ func(byte), _ func(string, Variant)) {
		if MissingDash != nil {
			MissingDash(long)
		}
	}
}

func SuggestionExtraDash(short byte) Suggestion {
	return func(_ func(Variant), _ func(string), ExtraDash func(byte), _ func(string, Variant)) {
		if ExtraDash != nil {
			ExtraDash(short)
		}
	}
}

func SuggestionNested(name string, variant Variant) Suggestion {
	return func(_ func(Variant), _ func(string), _ func(byte), Nested func(string, Variant)) {
		if Nested != nil {
			Nested(name, variant)
		}
	}
}

// Looks for potential typos
func suggest_zalupa(args State, meta Meta, return_ func(int, Suggestion, bool)) {
	next := args.items_iter()
	ix, arg, ok := next()
	if !ok {
		return_(0, nil, false)
	}

	// suggesting typos for parts of group of short names (-vvv, typo in third v) would be strange
	if arg.str == "" {
		return_(0, nil, false)
	}
	// don't try to suggest fixes for typos in strictly positional items
	if arg.IsPosWord() {
		return_(0, nil, false)
	}
	// it also should be a printable name
	actual := arg.str

	// all the help items one level deep
	hi := HelpItemsdefault()
	hi.append_meta(meta)

	// this will be used to avoid reallocations on scannign
	nested := HelpItemsdefault()

	// while scanning keep the closest match
	best_switch := Invalid[Variant]()
	best_dist := math.MaxInt
	improve := func(dist int, val Variant) {
		if best_dist > dist && dist > 0 && dist < 4 {
			best_dist = dist
			best_switch = Valid(val)
		}
	}

	nest := Invalid[fun.Pair[string, Variant]]()
	for _, item := range hi {
		switch item := item.(type) {
		case HelpItemCommand:
			name, meta := item.name, item.meta
			// command can result in 2 types of suggestions:
			// - typo in a short or a long name
			// - there is a nested command that matches perfectly - try using that
			distance := damerau_levenshtein(actual, name)
			improve(distance, VariantCommandLong(name))

			// scan nested items and look for exact matches only
			nested = nil
			nested.append_meta(meta)
			for _, item := range nested {
				switch item := item.(type) {
				case HelpItemCommand:
					nname := item.name
					if nname == actual {
						nest = Valid(fun.Pair[string, Variant]{name, VariantCommandLong(nname)})
					}
				case HelpItemFlag:
					nname := item.name
					if nname.eq(actual) {
						nest = Valid(fun.Pair[string, Variant]{name, VariantFlag(nname)})
					}
				case HelpItemArgument:
					nname := item.name
					if nname.eq(actual) {
						nest = Valid(fun.Pair[string, Variant]{name, VariantFlag(nname)})
					}
				}
			}
		case HelpItemFlag:
			name := item.name
			if long, ok := name.as_long(); ok {
				distance := damerau_levenshtein(actual, fmt.Sprintf("--%v", long))
				improve(distance, VariantFlag(name))
			}
			if short, ok := name.as_short(); ok {
				if act, ok := strings.CutPrefix(actual, "--"); ok {
					if act == string(short) {
						return_(ix, SuggestionExtraDash(short), true)
					}
				}
			}
		case HelpItemArgument:
			name := item.name
			if long, ok := name.as_long(); ok {
				distance := damerau_levenshtein(actual, fmt.Sprintf("--%v", long))
				improve(distance, VariantFlag(name))
			}
			if short, ok := name.as_short(); ok {
				if act, ok := strings.CutPrefix(actual, "--"); ok {
					if act == string(short) {
						return_(ix, SuggestionExtraDash(short), true)
					}
				}
			}
		}
	}

	if namevariant, ok := nest.Unpack(); ok {
		return_(ix, SuggestionNested(namevariant.K, namevariant.V), true)
	} else {
		// skip confusing errors
		if best_dist == math.MaxInt {
			return_(0, nil, false)
		}
		best_match, ok := best_switch.Unpack()
		if !ok {
			return_(0, nil, false)
		}

		// handle missing single dash typos separately
		best_match(
			nil,
			func(n ShortLong) {
				if long, ok := n.as_long(); ok {
					if s, ok := strings.CutPrefix(actual, "-"); ok && s == long {
						return_(ix, SuggestionMissingDash(long), true)
					}
				}

			},
		)
		return_(ix, SuggestionVariant(best_match), true)
	}

}

func suggest(args State, meta Meta) (a int, b Suggestion, c bool) {
	type zalupa struct{}
	defer func() {
		if r := recover(); r != nil {
			if _, ok := r.(zalupa); ok {
				return
			}
			panic(r)
		}
	}()

	suggest_zalupa(args, meta, func(aa int, bb Suggestion, cc bool) {
		a, b, c = aa, bb, cc
		panic(zalupa{})
	})
	return
}

// Damerau-Levenshtein distance function
//
// returns `math.MaxInt` if there's no common characters at all mostly to avoid
// confusing error messages - "you typed 'foo', maybe you ment 'bar'" where
// 'foo' and 'bar' don't have anything in common
func damerau_levenshtein(a, b string) int {
	a_len := len(a)
	b_len := len(b)
	d := make([]int, (a_len+1)*(b_len+1))

	ix := func(ib, ia int) int { return a_len*ia + ib }

	for i := range a_len + 1 {
		d[ix(i, 0)] = i
	}

	for j := range b_len + 1 {
		d[ix(0, j)] = j
	}

	pa := rune(0)
	pb := rune(0)
	for i, ca := range a {
		i := i + 1
		for j, cb := range b {
			j := j + 1
			cost := fun.IF(ca != cb, 1, 0)
			d[ix(i, j)] = min(
				d[ix(i-1, j)]+1,
				d[ix(i, j-1)]+1,
				d[ix(i-1, j-1)]+cost,
			)
			if i > 1 && j > 1 && ca == pb && cb == pa {
				d[ix(i, j)] = min(d[ix(i, j)], d[ix(i-2, j-2)]+1)
			}
			pb = cb
		}
		pa = ca
	}

	diff := d[ix(a_len, b_len)]

	if diff >= min(a_len, b_len) {
		return math.MaxInt
	} else {
		return diff
	}
}
