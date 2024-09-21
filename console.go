package cli

import (
	"testing"
)

// help needs to support following features:
// - linebreak - insert linebreak whenever
// - newline - start text on a new line, don't start a new line if not already at one
// - margin - start text at some offset at a new line
// - tabstop - all tabstops are aligned within a section
// - section title - "usage", "available options", "available positionals", etc. starts a new
//         section - resets tabstops
// - literal text user needs to type - flags, command names, etc.
// - metavar - meta placehoder user needs to write something
// - subsection title - two spaces + text, used for adjacent groups

// help might want to render it:
// - monochrome - default mode
// - bright/dull/custom colors
// - export to markdown and groff
//
// monochrome and colors are rendered with different widths so tabstops are out of buffer rendering

// text formatting rules:
//
// want to be able to produce both brief and full versions of the documentation, it only makes
// sense to look for that in the plain text...
// - "\n " => hard line break, inserted always
// - "\n\n" => paragraphs are separated by this, only the first one in inserted unless in "full" mode
// // - "\n" => converted to spaces, text flows to the current margin value
//
// tabstops are aligned the same position within a section, tabstop sets a temporary margin for all
// the soft linebreaks, tabstop
//
// margin sets the minimal offset for any new text and retained until new margin is set:
// "hello" [margin 8] "world" is rendered as "hello   world"

const MAX_TAB uint = 24
const MAX_WIDTH uint = 100

// Default to dull color if colors are enabled,
type Color func(
	Monochrome func(),
	Dull func(),
	Bright func(),
)

var ColorMonochrome Color = func(Monochrome, Dull, Bright func()) {
	if Monochrome != nil {
		Monochrome()
	}
}
var ColorDull Color = func(Monochrome, Dull, Bright func()) {
	if Dull != nil {
		Dull()
	}
}
var ColorBright Color = func(Monochrome, Dull, Bright func()) {
	if Bright != nil {
		Bright()
	}
}

func Colordefault() Color {
	res := ColorDull
	if !(on(StreamStdout).is_some() && on(StreamStderr).is_some()) {
		res = ColorMonochrome
	}
	return res
}

func (self Color) push_str(style Style, res *string, item string) {
	self(
		func() {
			*res += item
		},
		func() {
			style(
				func() { *res += item },
				func() { *res += item.underline().bold() },
				func() { *res += item.bold() },
				func() { *res += item.underline() },
				func() { *res += item.bold().red() },
			)
		},
		func() {
			style(
				func() { *res += item },
				func() { *res += item.yellow().bold() },
				func() { *res += item.green().bold() },
				func() { *res += item.blue().bold() },
				func() { *res += item.red().bold() },
			)
		},
	)
}

const PADDING = "                                                  "

// Render a monochrome version of the document
//
// `full` indicates if full message should be rendered, this makes
// difference for rendered help message, otherwise you can pass `true`.
func (self Doc) monochrome(full bool) string {
	return self.render_console(full, ColorMonochrome, MAX_WIDTH)
}

func (self Doc) render_console(full bool, color Color, max_width uint) string {
	res := ""
	tabstop := 0
	byte_pos := 0
	{
		current := 0
		in_term := false
		// looking for widest term below MAX_TAB
		for _, token := range self.tokens {
			token(
				func(Text TokenText) {
					bytes, style := Text.bytes, Text.style
					if in_term {
						current += self.payload[byte_pos : byte_pos+bytes].chars().count()
					}
					byte_pos += bytes
				},
				func(BlockStart Block) {
					if BlockStart == BlockItemTerm {
						in_term = true
						current = 0
					}
				},
				func(BlockEnd Block) {
					if BlockStart == BlockItemTerm {
						in_term = false
						if current > tabstop && current <= MAX_TAB {
							tabstop = current
						}
					}
				},
			)
		}
		byte_pos = 0
	}
	tabstop += 4

	stack := Vec__new()
	skip := Skip__default()
	char_pos := 0

	margins := []uint{}

	// a single new line, unless one exists
	pending_newline := false
	// a double newline, unless one exists
	pending_blank_line := false

	pending_margin := false

	for _, token := range self.tokens.iter().copied() {
		token(
			func(x TokenText) {
				bytes, style := x.bytes, x.style
				input := &self.payload[byte_pos : byte_pos+bytes]
				byte_pos += bytes

				if skip.enabled() {
					continue
				}

				for _, chunk := range split(input) {
					switch chunk {
					case Chunk__Raw(s, w):
						{
							margin := margins.last().copied().unwrap_or(0)
							if !res.is_empty() {
								if (pending_newline || pending_blank_line) && !res.ends_with('\n') {
									char_pos = 0
									res.push('\n')
								}
								if pending_blank_line && !res.ends_with("\n\n") {
									res.push('\n')
								}
								if char_pos+s.len() > max_width {
									char_pos = 0
									res.truncate(res.trim_end().len())
									res.push('\n')
									if s == " " {
										continue
									}
								}
							}

							pushed := 0
							if missing, ok := margin.checked_sub(char_pos).Unpack(); ok {
								res.push_str(&PADDING[:missing])
								char_pos = margin
								pushed = missing
							}
							if pending_margin && char_pos >= MAX_TAB+4 && pushed < 2 {
								missing := 2 - pushed
								res.push_str(&PADDING[:missing])
								char_pos += missing
							}

							pending_newline = false
							pending_blank_line = false
							pending_margin = false

							color.push_str(style, &res, s)
							char_pos += w
						}
					case Chunk__Paragraph:
						{
							res.push('\n')
							char_pos = 0
							if !full {
								skip.enable()
								break
							}
						}
					case Chunk__LineBreak:
						{
							res.push('\n')
							char_pos = 0
						}
					}
				}
			},
			func(block Block) {
				stack.push(block)
				margin := margins.last().copied().unwrap_or(0)

				switch block {
				case Block__Header, Block__Section2:
					pending_newline = true
					margins.push(margin)
				case Block__Section3:
					pending_newline = true
					margins.push(margin + 2)
				case Block__ItemTerm:
					pending_newline = true
					margins.push(margin + 4)
				case Block__ItemBody:
					margins.push(margin + tabstop + 2)
					pending_margin = true
				case Block__InlineBlock:
					skip.push()
				case Block__Block:
					margins.push(margin)
				case Block__DefinitionList, Block__Meta, Block__Mono:
				case Block__TermRef:
					if color == Color__Monochrome {
						res.push('`')
						char_pos += 1
					}
				}
			},
			func(block Block) {
				if testing.Testing() && stack.pop() != Some(block) {
					panic("assert")
				}

				margins.pop()
				switch block {
				case Block__ItemBody:
					pending_margin = false
				case Block__Header,
					Block__Section2,
					Block__Section3,
					Block__ItemTerm,
					Block__DefinitionList,
					Block__Meta,
					Block__Mono:
				case Block__InlineBlock:
					skip.pop()
				case Block__Block:
					pending_blank_line = true
				case Block__TermRef:
					if color == Color__Monochrome {
						res.push('`')
						char_pos += 1
					}
				}
			},
		)
	}
	if pending_newline || pending_blank_line {
		res.push('\n')
	}
	if testing.Testing() && stack.len() != 0 {
		panic("assert")
	}
	return res
}

/*
#[cfg(test)]
mod test {
    use super__*;

    #[test]
    fn tabstop_works() {
        // tabstop followed by newline
        let mut m = Buffer__default();
        m.token(Token__TermStart);
        m.text("aa");
        m.token(Token__TermStop);
        m.token(Token__LineBreak);

        m.token(Token__TermStart);
        m.text("b");
        m.token(Token__TermStop);
        m.text("c");
        m.token(Token__LineBreak);
        require.Equal(t,m.monochrome(true), "    aa\n    b   c\n");
        m.clear();

        // plain, narrow first
        m.token(Token__TermStart);
        m.text("1");
        m.token(Token__TermStop);
        m.text("22");
        m.token(Token__LineBreak);

        m.token(Token__TermStart);
        m.text("33");
        m.token(Token__TermStop);
        m.text("4");
        m.token(Token__LineBreak);
        require.Equal(t,m.monochrome(true), "    1   22\n    33  4\n");
        m.clear();

        // plain, wide first
        m.token(Token__TermStart);
        m.text("aa");
        m.token(Token__TermStop);

        m.text("b");
        m.token(Token__LineBreak);

        m.token(Token__TermStart);
        m.text("c");
        m.token(Token__TermStop);

        m.text("dd");
        m.token(Token__LineBreak);
        require.Equal(t,m.monochrome(true), "    aa  b\n    c   dd\n");
        m.clear();

        // two different styles first
        m.token(Token__TermStart);
        m.text("a");
        m.literal("b");
        m.token(Token__TermStop);

        m.literal("c");
        m.token(Token__LineBreak);
        m.token(Token__TermStart);
        m.text("d");
        m.token(Token__TermStop);

        m.literal("e");
        m.token(Token__LineBreak);
        require.Equal(t,m.monochrome(true), "    ab  c\n    d   e\n");
    }

    #[test]
    fn linewrap_works() {
        let mut m = Buffer__default();
        m.token(Token__TermStart);
        m.write_str("--hello", Style__Literal);
        m.token(Token__TermStop);
        for i in 0..25 {
            m.write_str(&format!("and word{i} "), Style__Text)
        }
        m.write_str("and last word", Style__Text);
        m.token(Token__LineBreak);

        let expected =
"    --hello  and word0 and word1 and word2 and word3 and word4 and word5 and word6 and word7 and word8
             and word9 and word10 and word11 and word12 and word13 and word14 and word15 and word16 and
             word17 and word18 and word19 and word20 and word21 and word22 and word23 and word24 and last
             word
";

        require.Equal(t,m.monochrome(true), expected);
    }

    #[test]
    fn very_long_tabstop() {
        let mut m = Buffer__default();
        m.token(Token__TermStart);
        m.write_str(
            "--this-is-a-very-long-option <DON'T DO THIS AT HOME>",
            Style__Literal,
        );
        m.token(Token__TermStop);
        for i in 0..15 {
            m.write_str(&format!("and word{i} "), Style__Text)
        }
        m.write_str("and last word", Style__Text);
        m.token(Token__LineBreak);

        let expected =
"    --this-is-a-very-long-option <DON'T DO THIS AT HOME>  and word0 and word1 and word2 and word3 and word4
      and word5 and word6 and word7 and word8 and word9 and word10 and word11 and word12 and word13 and
      word14 and last word
";

        require.Equal(t,m.monochrome(true), expected);
    }

    #[test]
    fn line_breaking_rules() {
        let mut buffer = Buffer__default();
        buffer.write_str("hello ", Style__Text);
        require.Equal(t,buffer.monochrome(true), "hello ");
        buffer.clear();

        buffer.write_str("hello\n world\n", Style__Text);
        require.Equal(t,buffer.monochrome(true), "hello\nworld ");
        buffer.clear();

        buffer.write_str("hello\nworld", Style__Text);
        require.Equal(t,buffer.monochrome(true), "hello world");
        buffer.clear();

        buffer.write_str("hello\nworld\n", Style__Text);
        require.Equal(t,buffer.monochrome(true), "hello world ");
        buffer.clear();

        buffer.write_str("hello\n\nworld", Style__Text);
        require.Equal(t,buffer.monochrome(false), "hello\n");
        buffer.clear();

        buffer.write_str("hello\n\nworld", Style__Text);
        require.Equal(t,buffer.monochrome(true), "hello\nworld");
        buffer.clear();
    }

    #[test]
    fn splitter_works() {
        require.Equal(t,
            split("hello ").collect__<Vec<_>>(),
            [Chunk__Raw("hello", 5), Chunk__Raw(" ", 1)]
        );

        require.Equal(t,
            split("hello\nworld").collect__<Vec<_>>(),
            [
                Chunk__Raw("hello", 5),
                Chunk__Raw(" ", 1),
                Chunk__Raw("world", 5)
            ]
        );

        require.Equal(t,
            split("hello\n world").collect__<Vec<_>>(),
            [
                Chunk__Raw("hello", 5),
                Chunk__HardLineBreak,
                Chunk__Raw("world", 5)
            ]
        );

        require.Equal(t,
            split("hello\n\nworld").collect__<Vec<_>>(),
            [
                Chunk__Raw("hello", 5),
                Chunk__SoftLineBreak,
                Chunk__Raw("world", 5)
            ]
        );
    }
}*/
