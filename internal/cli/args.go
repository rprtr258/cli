package cli

import (
	"bytes"
	"unicode/utf8"
)

// All currently present command line parameters with some extra metainfo
//
// Use it for unit tests and manual parsing. For production use you would want to replace the
// program name with [set_name](Argsset_name), but for tests passing a slice of strings to
// run_inner is usually more convenient.
//
// The easiest way to create `Args` is by using `newArgs`.
//
//	use bpaf*;
//	parser := short('f')
//		.Switch()
//		.to_options();
//
//	value, _ := parser.run_inner(Argsfrom(&["-f"]))
//	require.True(t, value);
//
//	// this also works
//	value, _ := parser.run_inner(newArgs("-f"])
//	require.True(t, value);
type Args struct {
	items []string
	name  Option[string]
	comp  Option[uint]
}

func NewArgs(args ...string) Args {
	return Args{
		items: args,
		name:  Invalid[string](),
		comp:  Invalid[uint](),
	}
}

// Enable completions with custom output revision style
//
// # Use revision 0 if you want to test completion mechanism
//
//	parser := short('f').Switch().to_options();
//	// ask bpaf to produce more input from "-", for
//	// suggesting new items use "" at the end
//	r := parser.run_inner(Argsfrom(&["-"])
//		.SetComp(0))
//		.unwrap_err()
//		.unwrap_stdout();
//	require.Equal(t,r, "-f");
//
// Note to self: shell passes "" as a parameter in situations like foo `--bar TAB`, bpaf
// completion stubs adopt this conventions add pass it along. This is needed so completer can
// tell the difference between `--bar` being completed or an argument to it in the example above.
func (self Args) SetComp(rev uint) Args {
	self.comp = Valid(rev)
	return self
}

type ItemStateKind uint8

const (
	ItemStateKindUnparsed ItemStateKind = iota // Value is yet to be parsed
	ItemStateKindConflict                      // Both branches succeeded, first parser was taken in favor of the second one
	ItemStateKindParsed                        // Value was parsed
)

// Shows which branch of [ParseOrElse] parsed the argument
type ItemState struct {
	kind     ItemStateKind
	conflict int
}

var ItemStateUnparsed = ItemState{kind: ItemStateKindUnparsed}

func ItemStateConflict(value int) ItemState {
	return ItemState{
		kind:     ItemStateKindConflict,
		conflict: value,
	}
}

var ItemStateParsed = ItemState{kind: ItemStateKindParsed}

func (self ItemState) parsed() bool {
	return self.kind == ItemStateKindParsed
}
func disambiguate_short(
	os string,
	short string,
	short_flags []byte,
	short_args []byte,
	items *[]Arg,
) (Error, bool) {
	// block can start with 0 or more short flags
	// followed by zero or one short argument, possibly with a body

	// keep the old length around so we can trimp items to it and push a ArgWord
	// if we decide to give up
	original := len(*items)

	// first flag contains the original os string for error message and anywhere purposes
	for ix, c := range short {
		cc := byte(c)
		tail_ix := ix + utf8.RuneLen(c)
		rest := short[tail_ix:]

		// shortcircuit single character short fun.Options
		if ix == 0 && rest == "" {
			*items = append(*items, ArgShort(os, cc, false))
			return nil, false
		}
		a := bytes.ContainsRune(short_flags, c)
		b := bytes.ContainsRune(short_args, c)
		switch {
		// short name that can be flag
		case a && !b:
			*items = append(*items, ArgShort(os, cc, false))
			// short name that can be argument
		case !a && b:
			adjacent_body := rest != ""
			*items = append(*items, ArgShort(os, cc, adjacent_body))
			if adjacent_body {
				*items = append(*items, ArgWord(rest))
			}
			return nil, false
		// neither is valid and there's more than one character. fallback to using it as a Word
		case !a && !b:
			*items = (*items)[:original]
			*items = append(*items, ArgWord(os))
			return nil, false
		// ambiguity, this is bad
		case a && b:
			msg := ErrorAmbiguity{len(*items), short}
			*items = append(*items, ArgWord(os))
			return msg, true
		}
	}
	return nil, false
}
