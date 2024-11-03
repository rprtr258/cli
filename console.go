package cli

import (
	"strings"
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

const MAX_TAB = 24
const MAX_WIDTH = 100

// Default to dull color if colors are enabled,
type Color uint8

const (
	ColorMonochrome Color = iota
	ColorDull
	ColorBright
)

func (c Color) IsMonochrome() bool {
	return c == ColorMonochrome
}

var Colordefault = ColorMonochrome

func (self Color) push_str(style Style, item string) string {
	switch self {
	case ColorMonochrome:
		return item
	case ColorDull:
		switch style {
		case StyleText:
			return item
		case StyleEmphasis:
			return item /*.underline().bold()*/
		case StyleLiteral:
			return item /*.bold()*/
		case StyleMetavar:
			return item /*.underline()*/
		case StyleInvalid:
			return item /*.bold().red()*/
		}
	case ColorBright:
		switch style {
		case StyleText:
			return item
		case StyleEmphasis:
			return item /*.yellow().bold()*/
		case StyleLiteral:
			return item /*.green().bold()*/
		case StyleMetavar:
			return item /*.blue().bold()*/
		case StyleInvalid:
			return item /*.red().bold()*/
		}
	}
	return ""
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
	tabstop := 0
	byte_pos := 0
	{
		current := 0
		in_term := false
		// looking for widest term below MAX_TAB
		for _, token := range self.tokens {
			switch {
			case token.isText:
				if in_term {
					current += len(self.payload[byte_pos : byte_pos+token.bytes])
				}
				byte_pos += token.bytes
			case token.isStart:
				if token.block == BlockItemTerm {
					in_term = true
					current = 0
				}
			default:
				if token.block == BlockItemTerm {
					in_term = false
					if current > tabstop && current <= MAX_TAB {
						tabstop = current
					}
				}
			}
		}
		byte_pos = 0
	}
	tabstop += 4

	res := ""
	stack := []Block{}
	skip := Skip(0)
	char_pos := 0
	margins := []uint{}
	pending_newline := false    // a single new line, unless one exists
	pending_blank_line := false // a double newline, unless one exists
	pending_margin := false
	for _, token := range self.tokens {
		switch {
		case token.isText:
			bytes, style := token.bytes, token.style
			input := self.payload[byte_pos : byte_pos+bytes]
			byte_pos += bytes

			if skip.enabled() {
				continue
			}

			for chunk := range split(input) {
				chunk(
					func(s string, w uint) {
						margin := last(margins).OrDefault(0)
						if res != "" {
							if (pending_newline || pending_blank_line) && !strings.HasSuffix(res, "\n") {
								char_pos = 0
								res += "\n"
							}
							if pending_blank_line && !strings.HasSuffix(res, "\n\n") {
								res += "\n"
							}
							if char_pos+len(s) > int(max_width) {
								char_pos = 0
								res = strings.TrimRight(res, " \n")
								res += "\n"
								if s == " " {
									return
								}
							}
						}

						pushed := 0
						if int(margin) >= char_pos {
							missing := int(margin) - char_pos
							res += PADDING[:missing]
							char_pos = int(margin)
							pushed = missing
						}
						if pending_margin && char_pos >= MAX_TAB+4 && pushed < 2 {
							missing := 2 - pushed
							res += PADDING[:missing]
							char_pos += missing
						}

						pending_newline = false
						pending_blank_line = false
						pending_margin = false

						res += color.push_str(style, s)
						char_pos += int(w)
					},
					func() {
						res += "\n"
						char_pos = 0
						if !full {
							skip.enable()
							return
						}
					},
					func() {
						res += "\n"
						char_pos = 0
					},
				)
			}
		case token.isStart:
			block := token.block
			stack = append(stack, block)
			margin := last(margins).OrDefault(0)

			switch block {
			case BlockHeader, BlockSection2:
				pending_newline = true
				margins = append(margins, margin)
			case BlockSection3:
				pending_newline = true
				margins = append(margins, margin+2)
			case BlockItemTerm:
				pending_newline = true
				margins = append(margins, margin+4)
			case BlockItemBody:
				margins = append(margins, margin+uint(tabstop)+2)
				pending_margin = true
			case BlockInlineBlock:
				skip.push()
			case BlockBlock:
				margins = append(margins, margin)
			case BlockDefinitionList, BlockMeta, BlockMono:
			case BlockTermRef:
				if color.IsMonochrome() {
					res += "`"
					char_pos += 1
				}
			}
		default:
			block := token.block
			if testing.Testing() && (len(stack) == 0 || stack[len(stack)-1] != block) {
				panic("assert")
			}

			if len(margins) > 0 {
				margins = margins[:len(margins)-1]
			}
			switch block {
			case BlockItemBody:
				pending_margin = false
			case BlockHeader,
				BlockSection2,
				BlockSection3,
				BlockItemTerm,
				BlockDefinitionList,
				BlockMeta,
				BlockMono:
			case BlockInlineBlock:
				skip.pop()
			case BlockBlock:
				pending_blank_line = true
			case BlockTermRef:
				if color.IsMonochrome() {
					res += "`"
					char_pos += 1
				}
			}
		}
	}
	if pending_newline || pending_blank_line {
		res += "\n"
	}
	if testing.Testing() && len(stack) != 0 {
		panic("assert")
	}
	return res
}
