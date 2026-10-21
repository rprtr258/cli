package cli

import (
	"iter"
	"strings"
)

type code uint8

const (
	codeNo code = iota
	codeFirst
	codeRest
)

// Raw(string, uint) | Paragraph | LineBreak
type chunk struct {
	s string
	i int // -1 for paragraph, -2 for LineBreak
}

func chunkRaw(code string, code_type uint) chunk {
	return chunk{code, int(code_type)}
}

var chunkParagraph = chunk{"", -1}
var chunkLineBreak = chunk{"", -2}

const chunkCode = 1_000_000
const chunkTickedCode = 1_000_001

// Split payload into chunks annotated with character width and containing no newlines according
// to text formatting rules
func split(input string) iter.Seq[chunk] {
	type Splitter struct {
		input string
		code  code
	}

	self := Splitter{
		input: input,
		code:  codeNo,
	}

	// 1. paragraphs are separated by a blank line.
	// 2. code blocks are aligned by 4 spaces and are kept intact
	// 3. linebreaks followed by space are preserved
	// 4. leftovers are fed word by word
	// 5. everything between "^```" is passed as is

	// 1. "\n\n" = Paragraph
	// 2. "\n " = LineBreak
	// 3. "\n" = " "
	// 4. "\n    " = code block
	// 5. take next word
	next := func() (chunk, bool) {
		if self.input == "" {
			return chunk{}, false
		}

		if self.code != codeNo {
			if self.code == codeRest && strings.HasPrefix(self.input, "```") {
				self.code = codeNo
			}
			if self.code == codeFirst {
				self.code = codeRest
			}

			tail := self.input

			var code string
			if code, rest, ok := strings.Cut(self.input, "\n"); ok {
				tail := tail[len(code):]
				if strings.HasPrefix(tail, "\n\n") && self.code == codeNo {
					self.input = tail
				} else {
					self.input = rest
				}
			} else {
				self.input = ""
				code = tail
			}
			return chunkRaw(code, chunkTickedCode), true
		}

		if tail, ok := strip_prefixb(self.input, '\n').Unpack(); ok {
			if tail, ok := strip_prefix(tail, "    ").Unpack(); ok {
				var code string
				if code, _, ok = strings.Cut(tail, "\n"); ok {
					self.input = tail[len(code):]
				} else {
					self.input = ""
					code = tail
				}
				return chunkRaw(code, chunkCode), true
			} else if strings.HasPrefix(tail, "\n```") {
				self.code = codeFirst
				self.input = tail[1:]
				return chunkParagraph, true
			} else if strings.HasPrefix(tail, "\n    ") {
				self.input = tail
				return chunkParagraph, true
			} else if tail, ok := strip_prefixb(tail, '\n').Unpack(); ok {
				self.input = tail
				return chunkParagraph, true
			} else if tail, ok := strip_prefixb(tail, ' ').Unpack(); ok {
				self.input = tail
				return chunkLineBreak, true
			} else {
				self.input = tail
				return chunkRaw(" ", 1), true
			}
		} else if tail, ok := strip_prefixb(self.input, ' ').Unpack(); ok {
			self.input = tail
			return chunkRaw(" ", 1), true
		} else {
			// there's iterator position but it won't give me character length of the rest of the input
			char_ix := uint(0)
			for byte_ix, chr := range []byte(self.input) {
				if chr == '\n' || chr == ' ' {
					head := self.input[:byte_ix]
					tail := self.input[byte_ix:]
					self.input = tail
					return chunkRaw(head, char_ix), true
				}
				char_ix += 1
			}
			head := self.input
			self.input = ""
			return chunkRaw(head, char_ix), true
		}
	}

	return fromPull(next)
}
