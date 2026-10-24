package cli

import (
	"fmt"
	"strings"
)

// example is a single documented usage example.
type example struct {
	// descr explains what the example demonstrates.
	descr string
	// command is the command line to run, relative to the parser the example
	// is declared on: without the program name and, for commands, without the
	// command name. The program name and the command path are prepended
	// automatically when the example is rendered or validated.
	command string
}

// split_example_args splits a shell-like command line into words. Single and
// double quotes group whitespace-separated words into a single word and are
// removed; no variable expansion or escaping is performed.
func split_example_args(command string) ([]string, error) {
	words := []string{}
	var word strings.Builder
	in_word := false
	quote := byte(0)
	for i := 0; i < len(command); i++ {
		c := command[i]
		switch {
		case quote != 0:
			if c == quote {
				quote = 0
			} else {
				word.WriteByte(c)
				in_word = true
			}
		case c == '\'' || c == '"':
			quote = c
			in_word = true
		case c == ' ' || c == '\t' || c == '\n':
			if in_word {
				words = append(words, word.String())
				word.Reset()
				in_word = false
			}
		default:
			word.WriteByte(c)
			in_word = true
		}
	}

	if quote != 0 {
		return nil, fmt.Errorf("unbalanced %c quote in example command %q", quote, command)
	}
	if in_word {
		words = append(words, word.String())
	}
	return words, nil
}

// write_examples renders the "Examples:" section with a command line for each
// example. The path is the command path to the parser the examples are
// declared on, starting with the program name.
func (self *Doc) write_examples(path []string, examples []example) {
	if len(examples) == 0 {
		return
	}

	self.token(tokenBlockStart(blockBlock))
	self.token(tokenBlockStart(blockSection2))
	self.write_str("Examples:", StyleEmphasis)
	self.token(tokenBlockEnd(blockSection2))
	self.token(tokenBlockStart(blockDefinitionList))

	for _, ex := range examples {
		self.token(tokenBlockStart(blockItemTerm))
		self.Doc(NewDoc(ex.descr))
		self.token(tokenBlockEnd(blockItemTerm))

		self.token(tokenBlockStart(blockItemBody))
		self.token(tokenBlockStart(blockMono))
		self.Literal("$ ")
		for _, item := range path {
			self.Literal(item)
			self.write_char(' ', StyleText)
		}
		self.Literal(ex.command)
		self.token(tokenBlockEnd(blockMono))
		self.token(tokenBlockEnd(blockItemBody))
	}

	self.token(tokenBlockEnd(blockDefinitionList))
	self.token(tokenBlockEnd(blockBlock))
}

// validate_examples parses every declared example with the parser it is
// declared on, panicking if any example fails to parse. For commands, the
// command path is prepended to the example command line before parsing.
//
// Examples are checked by [OptionParser.Run], [OptionParser.RunE] and
// [OptionParser.CheckInvariants].
func (self OptionParser[T]) validate_examples() {
	sections := []docSection{}
	path := []string{""}
	extract_sections(self.inner.meta, &self.info, &path, &sections)

	for _, section := range sections {
		for _, ex := range section.info.examples {
			words, err := split_example_args(ex.command)
			if err != nil {
				// The command was already checked for well-formedness at
				// declaration; this should never happen.
				panic("cli usage BUG: " + err.Error())
			}

			args := append(append([]string{"prog"}, section.path[1:]...), words...)
			if _, failure := self.run(argList{
				items: args[1:],
				name:  Valid("prog"),
				comp:  Invalid[uint](),
			}); failure != nil {
				panic(fmt.Sprintf(
					"cli usage BUG: example %q doesn't parse: %s",
					strings.Join(args[1:], " "),
					FailureError{failure: failure, maxWidth: self.info.max_width}.Error(),
				))
			}
		}
	}
}
