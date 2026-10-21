package cli

import (
	"fmt"
	"math"
	"strings"

	"github.com/rprtr258/fun"
)

type variant func(
	CommandLong func(string),
	Flag func(shortLong),
)

func variantCommandLong(name string) variant {
	return func(CommandLong func(string), _ func(shortLong)) {
		if CommandLong != nil {
			CommandLong(name)
		}
	}
}

func variantFlag(name shortLong) variant {
	return func(_ func(string), Flag func(shortLong)) {
		if Flag != nil {
			Flag(name)
		}
	}
}

type suggestion func(
	variant func(variant),
	// expected --foo, actual -foo
	MissingDash func(string),
	// expected -f, actual --f
	ExtraDash func(byte),
	Nested func(string, variant),
)

func suggestionVariant(v variant) suggestion {
	return func(vv func(variant), _ func(string), _ func(byte), _ func(string, variant)) {
		if vv != nil {
			vv(v)
		}
	}
}

func suggestionMissingDash(long string) suggestion {
	return func(_ func(variant), MissingDash func(string), _ func(byte), _ func(string, variant)) {
		if MissingDash != nil {
			MissingDash(long)
		}
	}
}

func suggestionExtraDash(short byte) suggestion {
	return func(_ func(variant), _ func(string), ExtraDash func(byte), _ func(string, variant)) {
		if ExtraDash != nil {
			ExtraDash(short)
		}
	}
}

func suggestionNested(name string, v variant) suggestion {
	return func(_ func(variant), _ func(string), _ func(byte), Nested func(string, variant)) {
		if Nested != nil {
			Nested(name, v)
		}
	}
}

// Looks for potential typos
func suggest_zalupa(state state, meta meta, return_ func(int, suggestion, bool)) {
	next := state.items_iter()
	ix, arg, ok := next()
	if !ok {
		return_(0, nil, false)
	}

	// suggesting typos for parts of group of short names (-vvv, typo in third v) would be strange
	if arg.str == "" {
		return_(0, nil, false)
	}
	// don't try to suggest fixes for typos in strictly positional items
	if arg.IsPositional() {
		return_(0, nil, false)
	}
	// it also should be a printable name
	actual := arg.GoString()

	// all the help items one level deep
	hi := helpItemsDefault()
	hi.append_meta(meta)

	// this will be used to avoid reallocations on scannign
	nested := helpItemsDefault()

	// while scanning keep the closest match
	best_switch := Invalid[variant]()
	best_dist := math.MaxInt
	improve := func(dist int, val variant) {
		if best_dist > dist && dist > 0 && dist < 4 {
			best_dist = dist
			best_switch = Valid(val)
		}
	}

	nest := Invalid[fun.Pair[string, variant]]()
	for _, item := range hi {
		switch item := item.(type) {
		case helpItemCommand:
			name, meta := item.name, item.meta
			// command can result in 2 types of suggestions:
			// - typo in a short or a long name
			// - there is a nested command that matches perfectly - try using that
			distance := damerau_levenshtein(actual, name)
			improve(distance, variantCommandLong(name))

			// scan nested items and look for exact matches only
			nested = nil
			nested.append_meta(meta)
			for _, item := range nested {
				switch item := item.(type) {
				case helpItemCommand:
					nname := item.name
					if nname == actual {
						nest = Valid(fun.Pair[string, variant]{name, variantCommandLong(nname)})
					}
				case helpItemFlag:
					nname := item.name
					if nname.eq(actual) {
						nest = Valid(fun.Pair[string, variant]{name, variantFlag(nname)})
					}
				case helpItemArgument:
					nname := item.name
					if nname.eq(actual) {
						nest = Valid(fun.Pair[string, variant]{name, variantFlag(nname)})
					}
				}
			}
		case helpItemFlag:
			name := item.name
			if long, ok := name.as_long(); ok {
				distance := damerau_levenshtein(actual, fmt.Sprintf("--%v", long))
				improve(distance, variantFlag(name))
			}
			if short, ok := name.as_short(); ok {
				if act, ok := strings.CutPrefix(actual, "--"); ok {
					if act == string(short) {
						return_(ix, suggestionExtraDash(short), true)
					}
				}
			}
		case helpItemArgument:
			name := item.name
			if long, ok := name.as_long(); ok {
				distance := damerau_levenshtein(actual, fmt.Sprintf("--%v", long))
				improve(distance, variantFlag(name))
			}
			if short, ok := name.as_short(); ok {
				if act, ok := strings.CutPrefix(actual, "--"); ok {
					if act == string(short) {
						return_(ix, suggestionExtraDash(short), true)
					}
				}
			}
		}
	}

	if namevariant, ok := nest.Unpack(); ok {
		return_(ix, suggestionNested(namevariant.K, namevariant.V), true)
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
			func(n shortLong) {
				if long, ok := n.as_long(); ok {
					if s, ok := strings.CutPrefix(actual, "-"); ok && s == long {
						return_(ix, suggestionMissingDash(long), true)
					}
				}

			},
		)
		return_(ix, suggestionVariant(best_match), true)
	}

}

func suggest(state state, meta meta) (a int, b suggestion, c bool) {
	type zalupa struct{}
	defer func() {
		if r := recover(); r != nil {
			if _, ok := r.(zalupa); ok {
				return
			}
			panic(r)
		}
	}()

	suggest_zalupa(state, meta, func(aa int, bb suggestion, cc bool) {
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
