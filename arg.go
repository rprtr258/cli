package cli

import "fmt"

// Preprocessed command line argument
// [string] in Short/Long correspond to orignal command line item used for errors
type Arg struct {
	str   string
	value func(
		// short flag: `-f`
		//
		// bool indicates if following item is also part of this Short
		Short func(byte, bool),

		// long flag: `--flag`
		// bool tells if it looks like --key=val or not
		Long func(string, bool),

		// "val" part of --key=val -k=val -kval
		ArgWord func(),

		// separate word that can be command, positional or a separate argument to a flag
		//
		// Can start with `-` or `--`, doesn't have to be valid utf8
		//
		// `hello`
		Word func(),

		// separate word that goes after `--`, strictly positional
		//
		// Can start with `-` or `--`, doesn't have to be valid utf8
		PosWord func(),
	)
}

func ArgShort(c byte, b bool, s string) Arg {
	return Arg{str: s, value: func(Short func(byte, bool), _ func(string, bool), _, _, _ func()) {
		if Short != nil {
			Short(c, b)
		}
	}}
}

func ArgLong(s string, b bool, os string) Arg {
	return Arg{str: os, value: func(_ func(byte, bool), Long func(string, bool), _, _, _ func()) {
		if Long != nil {
			Long(s, b)
		}
	}}
}

func ArgArgWord(s string) Arg {
	return Arg{str: s, value: func(_ func(byte, bool), _ func(string, bool), ArgWord, _, _ func()) {
		if ArgWord != nil {
			ArgWord()
		}
	}}
}

func ArgWord(s string) Arg {
	return Arg{str: s, value: func(_ func(byte, bool), _ func(string, bool), _, Word, _ func()) {
		if Word != nil {
			Word()
		}
	}}
}

func ArgPosWord(s string) Arg {
	return Arg{str: s, value: func(_ func(byte, bool), _ func(string, bool), _, _, PosWord func()) {
		if PosWord != nil {
			PosWord()
		}
	}}
}

func (self Arg) IsShort() bool {
	var res bool
	self.value(
		func(byte, bool) { res = true },
		nil, nil, nil, nil,
	)
	return res
}

func (self Arg) IsLong() bool {
	var res bool
	self.value(
		nil,
		func(string, bool) { res = true },
		nil, nil, nil,
	)
	return res
}

func (self Arg) IsPosWord() bool {
	var res bool
	self.value(
		nil, nil, nil, nil,
		func() { res = true },
	)
	return res
}

func (self Arg) match_short(val byte) bool {
	res := false
	self.value(func(c byte, b bool) { res = c == val }, nil, nil, nil, nil)
	return res
}

func (self Arg) match_long(val string) bool {
	res := false
	self.value(nil, func(c string, b bool) { res = c == val }, nil, nil, nil)
	return res
}

func (self Arg) String() string {
	var res string
	self.value(
		func(c byte, b bool) { res = fmt.Sprintf("Short(%s, %c, %v)", self.str, c, b) },
		func(l string, b bool) { res = fmt.Sprintf("Long(%s, %s, %v)", self.str, l, b) },
		func() { res = fmt.Sprintf("ArgWord(%s)", self.str) },
		func() { res = fmt.Sprintf("Word(%s)", self.str) },
		func() { res = fmt.Sprintf("PosWord(%s)", self.str) },
	)
	return res
}

// short flag disambiguations:
//
// Short flags | short arg
// No          | No        | no problem
// Yes         | No        | use flag
// No          | Yes       | use arg
// Yes         | Yes       | ask user?
//
// -a  - just a regular short flag: "-a"
// -abc - assuming there are short flags a, b and c: "-a -b -c", assuming utf8 values AND there's no argument -a
// -abc - assuming there's no -a -b -c: "-a bc"
// -abc - assuming both short a b c AND there's argument -a - need to disambiguate  on a context level
//
// 1. parse argument into ambigous representation that can store both short flags and argument
// 2. collect short flag/arg when entering the subparsre
// 3. when reaching ambi
func (self Arg) GoString() string {
	var res string
	self.value(
		func(c byte, _ bool) { res = fmt.Sprintf("-%c", c) },
		func(l string, _ bool) { res = fmt.Sprintf("--%s", l) },
		func() { res = self.str },
		func() { res = self.str },
		func() { res = self.str },
	)
	return res
}

type ArgType bool

const (
	ArgTypeShort ArgType = false
	ArgTypeLong  ArgType = true
)

// split [string] into argument specific bits
//
// takes a possibly non-utf8 string looking like "--name=value" and splits it into bits:
// "--" - type, "name" - name, must be representable as utf8, "=" - fun.Optional, "value" - flag
//
// dashes and equals sign are low codepoint values and - can look for them literally in a string.
// This probably means not supporting dashes with diacritics, but that's okay
//
// name must be valid utf8 after conversion and must not include `=`
//
// argument is fun.Optional and can be non valid utf8.
//
// The idea is to split the [string] into opaque parts by looking only at the parts simple parts
// and let stdlib to handle the decoding of those parts.
//
// performance wise this (at least on unix) works some small number percentage slower than the
// previous version
//
// Notation -fbar is ambigous and could mean either `-f -b -a -r` or `-f=bar`, resolve it into
// [ArgAmbiguity] and let subparser disambiguate it later depending on available short flag and
// arguments
func split_os_argument(input string) (ArgType, string, Option[Arg], bool) {
	// but in either case dashes and equals are just literal values just with different width
	const DASH = '-'
	const EQUALS = '='

	// preallocate something to store the name. oversized but avoids extra allocations/copying
	name := make([]byte, 0, len(input))

	items := []byte(input)

	// first item must be dash, otherwise it's positional or a flag value
	if len(items) == 0 || items[0] != DASH {
		return false, "", Invalid[Arg](), false
	}
	items = items[1:]

	// second item may or may not be, but should be present
	var ty ArgType
	if len(items) > 0 && items[0] == DASH {
		ty = ArgTypeLong
	} else {
		ty = ArgTypeShort
		name = append(name, items[0])
	}
	items = items[1:]

	// keep collecting until = or the end of the input
	for _, c := range items {
		switch c {
		case EQUALS:
			if ty == ArgTypeShort && len(name) > 1 {
				body := name[1:]
				body = append(body, EQUALS)
				body = append(body, items...)
				name = name[1:]
				os := ArgArgWord(string(body))
				return ty, string(name), Valid(os), true
			}

			word := ArgArgWord(string(items))
			return ty, string(name), Valid(word), true
		default:
			name = append(name, c)
		}
	}

	if len(name) == 0 {
		return false, "", Invalid[Arg](), false
	}
	return ty, string(name), Invalid[Arg](), true
}
