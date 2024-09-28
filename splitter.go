package cli

import (
	"iter"
	"strings"
)

type Code func(
	No func(),
	First func(),
	Rest func(),
)

func CodeNo() Code {
	return func(f, _, _ func()) {
		if f != nil {
			f()
		}
	}
}

func CodeFirst() Code {
	return func(_, f, _ func()) {
		if f != nil {
			f()
		}
	}
}

func CodeRest() Code {
	return func(_, _, f func()) {
		if f != nil {
			f()
		}
	}
}

func (c Code) IsNo() bool {
	var res bool
	c(func() { res = true }, nil, nil)
	return res
}

func (c Code) IsFirst() bool {
	var res bool
	c(nil, func() { res = true }, nil)
	return res
}

func (c Code) IsRest() bool {
	var res bool
	c(nil, nil, func() { res = true })
	return res
}

type Chunk func(
	Raw func(string, uint),
	Paragraph func(),
	LineBreak func(),
)

func ChunkRaw(code string, code_type uint) Chunk {
	return func(f func(string, uint), _, _ func()) {
		if f != nil {
			f(code, code_type)
		}
	}
}
func ChunkParagraph(_ func(string, uint), f, _ func()) {
	if f != nil {
		f()
	}
}
func ChunkLineBreak(_ func(string, uint), _, f func()) {
	if f != nil {
		f()
	}
}

const ChunkCODE = 1_000_000
const ChunkTICKED_CODE = 1_000_001

// Split payload into chunks annotated with character width and containing no newlines according
// to text formatting rules
func split(input string) iter.Seq[Chunk] {
	type Splitter struct {
		input string
		code  Code
	}

	self := Splitter{
		input: input,
		code:  CodeNo(),
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
	next := func() (Chunk, bool) {
		if self.input == "" {
			return nil, false
		}

		if !self.code.IsNo() {
			if self.code.IsRest() && strings.HasPrefix(self.input, "```") {
				self.code = CodeNo()
			}
			if self.code.IsFirst() {
				self.code = CodeRest()
			}

			tail := self.input

			var code string
			if code, rest, ok := strings.Cut(self.input, "\n"); ok {
				tail := tail[len(code):]
				if strings.HasPrefix(tail, "\n\n") && self.code.IsNo() {
					self.input = tail
				} else {
					self.input = rest
				}
			} else {
				self.input = ""
				code = tail
			}
			return ChunkRaw(code, ChunkTICKED_CODE), true
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
				return ChunkRaw(code, ChunkCODE), true
			} else if strings.HasPrefix(tail, "\n```") {
				self.code = CodeFirst()
				self.input = tail[1:]
				return ChunkParagraph, true
			} else if strings.HasPrefix(tail, "\n    ") {
				self.input = tail
				return ChunkParagraph, true
			} else if tail, ok := strip_prefixb(tail, '\n').Unpack(); ok {
				self.input = tail
				return ChunkParagraph, true
			} else if tail, ok := strip_prefixb(tail, ' ').Unpack(); ok {
				self.input = tail
				return ChunkLineBreak, true
			} else {
				self.input = tail
				return ChunkRaw(" ", 1), true
			}
		} else if tail, ok := strip_prefixb(self.input, ' ').Unpack(); ok {
			self.input = tail
			return ChunkRaw(" ", 1), true
		} else {
			// there's iterator position but it won't give me character length of the rest of the input
			char_ix := uint(0)
			for byte_ix, chr := range []byte(self.input) {
				if chr == '\n' || chr == ' ' {
					head := self.input[:byte_ix]
					tail := self.input[byte_ix:]
					self.input = tail
					return ChunkRaw(head, char_ix), true
				}
				char_ix += 1
			}
			head := self.input
			self.input = ""
			return ChunkRaw(head, char_ix), true
		}
	}

	return fromPull(next)
}
