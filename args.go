package cli

import (
	"bytes"
	"unicode/utf8"
)

// argList holds all currently present command line parameters with some
// extra metainfo.
//
// Internal type used for parsing; for tests build one with [newArgList]
// and pass it to the internal run helpers.
type argList struct {
	items []string
	name  Opt[string]
	comp  Opt[uint]
}

func newArgList(args ...string) argList {
	return argList{
		items: args,
		name:  Invalid[string](),
		comp:  Invalid[uint](),
	}
}

// setComp enables completions with a custom output revision style.
//
// Revision 0 is useful to test the completion mechanism:
//
//	// ask the completer to produce more input from "-", for
//	// suggesting new items use "" at the end
//	r, err := parser.run(newArgList("-").setComp(0))
//
// Note: the shell passes "" as a parameter in situations like `foo --bar TAB`; the
// completion stubs adopt this convention and pass it along. This is needed so the
// completer can tell the difference between `--bar` being completed or an argument
// to it in the example above.
func (self argList) setComp(rev uint) argList {
	self.comp = Valid(rev)
	return self
}

type itemStateKind uint8

const (
	itemStateKindUnparsed itemStateKind = iota // Value is yet to be parsed
	itemStateKindConflict                      // Both branches succeeded, first parser was taken in favor of the second one
	itemStateKindParsed                        // Value was parsed
)

// itemState shows which branch of [Parser.OrElse] parsed the argument.
type itemState struct {
	kind     itemStateKind
	conflict int
}

var itemStateUnparsed = itemState{kind: itemStateKindUnparsed}

func itemStateConflict(value int) itemState {
	return itemState{
		kind:     itemStateKindConflict,
		conflict: value,
	}
}

var itemStateParsed = itemState{kind: itemStateKindParsed}

func (self itemState) parsed() bool {
	return self.kind == itemStateKindParsed
}

func disambiguate_short(
	os string,
	short string,
	short_flags []byte,
	short_args []byte,
	items *[]arg,
) (parseError, bool) {
	// block can start with 0 or more short flags
	// followed by zero or one short argument, possibly with a body

	// keep the old length around so we can trimp items to it and push a argWord
	// if we decide to give up
	original := len(*items)

	// first flag contains the original os string for error message and anywhere purposes
	first_flag := os
	for ix, c := range short {
		cc := byte(c)
		tail_ix := ix + utf8.RuneLen(c)
		rest := short[tail_ix:]

		// shortcircuit single character short flags
		if ix == 0 && rest == "" {
			*items = append(*items, argShort(first_flag, cc, false))
			return nil, false
		}
		a := bytes.ContainsRune(short_flags, c)
		b := bytes.ContainsRune(short_args, c)
		switch {
		// short name that can be flag
		case a && !b:
			*items = append(*items, argShort(first_flag, cc, false))
			first_flag = ""
			// short name that can be argument
		case !a && b:
			adjacent_body := rest != ""
			*items = append(*items, argShort(os, cc, adjacent_body))
			if adjacent_body {
				*items = append(*items, argWord(rest))
			}
			return nil, false
		// neither is valid and there's more than one character. fallback to using it as a Word
		case !a && !b:
			*items = (*items)[:original]
			*items = append(*items, argWord(os))
			return nil, false
		// ambiguity, this is bad
		case a && b:
			msg := errorAmbiguity{len(*items), short}
			*items = append(*items, argWord(os))
			return msg, true
		}
	}
	return nil, false
}
