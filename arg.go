package cli

import "fmt"

type ArgKind uint8

const (
	// -f
	ArgKindShortFlag ArgKind = iota
	// --flag
	ArgKindLongFlag
	// "val" part of --key=val -k=val -kval
	ArgKindArgWord
	// separate word that can be command, positional or a separate argument to a flag
	//
	// Can start with `-` or `--`, doesn't have to be valid utf8
	//
	// e.g. `hello`
	ArgKindWord
	// separate word that goes after `--`, strictly positional
	// Can start with `-` or `--`, doesn't have to be valid utf8
	ArgKindPositional
)

// Preprocessed command line argument
// [string] in Short/Long correspond to orignal command line item used for errors
type Arg struct {
	str  string
	kind ArgKind
	// kind is short or long flag
	// if it looks like --key=val or not
	// if following item is also part of this Short
	is_adj bool
	// kind is short flag
	c byte
	// kind is long flag
	s string
}

func ArgWord(str string) Arg {
	return Arg{str: str, kind: ArgKindWord}
}

func ArgArgWord(str string) Arg {
	return Arg{str: str, kind: ArgKindArgWord}
}

func ArgPositional(str string) Arg {
	return Arg{str: str, kind: ArgKindPositional}
}

func ArgShort(str string, c byte, b bool) Arg {
	return Arg{str: str, kind: ArgKindShortFlag, c: c, is_adj: b}
}

func ArgLong(str string, s string, b bool) Arg {
	return Arg{str: str, kind: ArgKindLongFlag, s: s, is_adj: b}
}

func (self Arg) IsShort() bool {
	return self.kind == ArgKindShortFlag
}

func (self Arg) IsLong() bool {
	return self.kind == ArgKindLongFlag
}

func (self Arg) IsPositional() bool {
	return self.kind == ArgKindPositional
}

func (self Arg) match_short(val byte) bool {
	return self.IsShort() && self.c == val
}

func (self Arg) match_long(val string) bool {
	return self.IsLong() && self.s == val
}

func (self Arg) String() string {
	switch self.kind {
	case ArgKindShortFlag:
		return fmt.Sprintf("Short(%s, %c, %v)", self.str, self.c, self.is_adj)
	case ArgKindLongFlag:
		return fmt.Sprintf("Long(%s, %s, %v)", self.str, self.s, self.is_adj)
	case ArgKindArgWord:
		return fmt.Sprintf("ArgWord(%s)", self.str)
	case ArgKindWord:
		return fmt.Sprintf("Word(%s)", self.str)
	case ArgKindPositional:
		return fmt.Sprintf("Positional(%s)", self.str)
	}
	panic("unreachable")
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
// -abc - assuming both short a b c AND there's argument -a - need to disambiguate on a context level
//
// 1. parse argument into ambigous representation that can store both short flags and argument
// 2. collect short flag/arg when entering the subparsre
// 3. when reaching ambi
func (self Arg) GoString() string {
	switch self.kind {
	case ArgKindShortFlag:
		return fmt.Sprintf("-%c", self.c)
	case ArgKindLongFlag:
		return fmt.Sprintf("--%s", self.s)
	default:
		return self.str
	}
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
	if len(items) == 0 {
		return false, "", Invalid[Arg](), false
	} else if items[0] == DASH {
		ty = ArgTypeLong
	} else {
		ty = ArgTypeShort
		name = append(name, items[0])
	}
	items = items[1:]

	// keep collecting until = or the end of the input
	for i, c := range items {
		switch c {
		case EQUALS:
			if ty == ArgTypeShort && len(name) > 1 {
				body := name[1:]
				body = append(body, EQUALS)
				body = append(body, items[i+1:]...)
				name = name[:1]
				os := ArgWord(string(body))
				return ty, string(name), Valid(os), true
			}

			word := ArgWord(string(items[i+1:]))
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
