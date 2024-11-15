package cli

import "fmt"

type arg interface{ isArg() }

// short flag: `-f`
//
// bool indicates if following item is also part of this Short
type argShort struct {
	c      byte
	is_adj bool
}

func (argShort) isArg() {}

// long flag: `--flag`
// bool tells if it looks like --key=val or not
type argLong struct {
	s      string
	is_adj bool
}

func (argLong) isArg() {}

// "val" part of --key=val -k=val -kval
type argArgWord struct{}

func (argArgWord) isArg() {}

// separate word that can be command, positional or a separate argument to a flag
//
// Can start with `-` or `--`, doesn't have to be valid utf8
//
// `hello`
type argWord struct{}

func (argWord) isArg() {}

// separate word that goes after `--`, strictly positional
//
// Can start with `-` or `--`, doesn't have to be valid utf8
type argPosWord struct{}

func (argPosWord) isArg() {}

// Preprocessed command line argument
// [string] in Short/Long correspond to orignal command line item used for errors
type Arg struct {
	str   string
	value arg
}

func ArgWord(str string) Arg {
	return Arg{str: str, value: argWord{}}
}

func ArgArgWord(str string) Arg {
	return Arg{str: str, value: argArgWord{}}
}

func ArgPosWord(str string) Arg {
	return Arg{str: str, value: argPosWord{}}
}

func ArgShort(str string, c byte, b bool) Arg {
	return Arg{str: str, value: argShort{c, b}}
}

func ArgLong(str string, s string, b bool) Arg {
	return Arg{str: str, value: argLong{s, b}}
}

func (self Arg) IsShort() bool {
	_, ok := self.value.(argShort)
	return ok
}

func (self Arg) IsLong() bool {
	_, ok := self.value.(argLong)
	return ok
}

func (self Arg) IsPosWord() bool {
	_, ok := self.value.(argPosWord)
	return ok
}

func (self Arg) match_short(val byte) bool {
	sh, ok := self.value.(argShort)
	return ok && sh.c == val
}

func (self Arg) match_long(val string) bool {
	sh, ok := self.value.(argLong)
	return ok && sh.s == val
}

func (self Arg) String() string {
	switch x := self.value.(type) {
	case argShort:
		return fmt.Sprintf("Short(%s, %c, %v)", self.str, x.c, x.is_adj)
	case argLong:
		return fmt.Sprintf("Long(%s, %s, %v)", self.str, x.s, x.is_adj)
	case argArgWord:
		return fmt.Sprintf("ArgWord(%s)", self.str)
	case argWord:
		return fmt.Sprintf("Word(%s)", self.str)
	case argPosWord:
		return fmt.Sprintf("PosWord(%s)", self.str)
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
// -abc - assuming both short a b c AND there's argument -a - need to disambiguate  on a context level
//
// 1. parse argument into ambigous representation that can store both short flags and argument
// 2. collect short flag/arg when entering the subparsre
// 3. when reaching ambi
func (self Arg) GoString() string {
	switch x := self.value.(type) {
	case argShort:
		return fmt.Sprintf("-%c", x.c)
	case argLong:
		return fmt.Sprintf("--%s", x.s)
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
	if len(items) > 0 && items[0] == DASH {
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
				name = name[1:]
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
