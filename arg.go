package cli

import "fmt"

type argKind uint8

const (
	// -f
	argKindShortFlag argKind = iota
	// --flag
	argKindLongFlag
	// separate word that can be command, positional or a separate argument to a flag
	//
	// Can start with `-` or `--`, doesn't have to be valid utf8
	//
	// e.g. `hello`
	argKindWord
	// separate word that goes after `--`, strictly positional
	// Can start with `-` or `--`, doesn't have to be valid utf8
	argKindPositional
)

// Preprocessed command line argument
// [string] in Short/Long correspond to orignal command line item used for errors
type arg struct {
	str  string
	kind argKind
	// kind is short or long flag
	// if it looks like --key=val or not
	// if following item is also part of this Short
	isAdjacent bool
	// kind is short flag
	c byte
	// kind is long flag
	s string
}

func argWord(str string) arg {
	return arg{str: str, kind: argKindWord}
}

func argPositional(str string) arg {
	return arg{str: str, kind: argKindPositional}
}

func argShort(str string, c byte, b bool) arg {
	return arg{str: str, kind: argKindShortFlag, c: c, isAdjacent: b}
}

func argLong(str, s string, b bool) arg {
	return arg{str: str, kind: argKindLongFlag, s: s, isAdjacent: b}
}

func (self arg) IsShort() bool {
	return self.kind == argKindShortFlag
}

func (self arg) IsLong() bool {
	return self.kind == argKindLongFlag
}

func (self arg) IsPositional() bool {
	return self.kind == argKindPositional
}

func (self arg) match_short(val byte) bool {
	return self.IsShort() && self.c == val
}

func (self arg) match_long(val string) bool {
	return self.IsLong() && self.s == val
}

func (self arg) String() string {
	switch self.kind {
	case argKindShortFlag:
		return fmt.Sprintf("Short(%s, %c, %v)", self.str, self.c, self.isAdjacent)
	case argKindLongFlag:
		return fmt.Sprintf("Long(%s, %s, %v)", self.str, self.s, self.isAdjacent)
	case argKindWord:
		return fmt.Sprintf("Word(%s)", self.str)
	case argKindPositional:
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
func (self arg) GoString() string {
	switch self.kind {
	case argKindShortFlag:
		return fmt.Sprintf("-%c", self.c)
	case argKindLongFlag:
		return fmt.Sprintf("--%s", self.s)
	default:
		return self.str
	}
}

type argType bool

const (
	argTypeShort argType = false
	argTypeLong  argType = true
)

// split [string] into argument specific bits
//
// takes a possibly non-utf8 string looking like "--name=value" and splits it into bits:
// "--" - type, "name" - name, must be representable as utf8, "=" - optional, "value" - flag
//
// dashes and equals sign are low codepoint values and - can look for them literally in a string.
// This probably means not supporting dashes with diacritics, but that's okay
//
// name must be valid utf8 after conversion and must not include `=`
//
// argument is optional and can be non-valid utf8.
//
// The idea is to split the [string] into opaque parts by looking only at the parts simple parts
// and let stdlib to handle the decoding of those parts.
//
// performance wise this (at least on unix) works some small number percentage slower than the
// previous version
//
// Notation -fbar is ambigous and could mean either `-f -b -a -r` or `-f=bar`, resolve it into
// [argAmbiguity] and let subparser disambiguate it later depending on available short flag and
// arguments
func split_os_argument(input string) (argType, string, Opt[arg], bool) {
	// but in either case dashes and equals are just literal values just with different width
	const DASH = '-'
	const EQUALS = '='

	// preallocate something to store the name. oversized but avoids extra allocations/copying
	name := make([]byte, 0, len(input))

	items := []byte(input)

	// first item must be dash, otherwise it's positional or a flag value
	if len(items) == 0 || items[0] != DASH {
		return false, "", Invalid[arg](), false
	}
	items = items[1:]

	// second item may or may not be, but should be present
	if len(items) == 0 {
		return false, "", Invalid[arg](), false
	}
	var ty argType
	switch items[0] {
	case DASH:
		ty = argTypeLong
	default:
		ty = argTypeShort
		name = append(name, items[0])
	}
	items = items[1:]

	// keep collecting until = or the end of the input
	for i, c := range items {
		switch c {
		case EQUALS:
			if ty == argTypeShort && len(name) > 1 {
				body := name[1:]
				body = append(body, EQUALS)
				body = append(body, items[i+1:]...)
				name = name[:1]
				os := argWord(string(body))
				return ty, string(name), Valid(os), true
			}

			word := argWord(string(items[i+1:]))
			return ty, string(name), Valid(word), true
		default:
			name = append(name, c)
		}
	}

	if len(name) == 0 {
		return false, "", Invalid[arg](), false
	}
	return ty, string(name), Invalid[arg](), true
}
