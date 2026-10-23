package cli

import (
	"strings"
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

const (
	maxTab         = 24
	maxRenderWidth = 100
)

// color selects the color scheme; it defaults to dull color if colors are
// enabled.
type color uint8

const (
	colorMonochrome color = iota
	colorDull
	colorBright
)

func (c color) IsMonochrome() bool {
	return c == colorMonochrome
}

var colorDefault = colorMonochrome

// ANSI escape sequences used to style the rendered text.
const (
	ansiReset     = "\x1b[0m"
	ansiBold      = "\x1b[1m"
	ansiUnderline = "\x1b[4m"
	ansiRed       = "\x1b[31m"
	ansiYellow    = "\x1b[33m"
	ansiGreen     = "\x1b[32m"
	ansiBlue      = "\x1b[34m"
)

// styled wraps item with the given opening sequences and a trailing reset
func styled(item string, seqs ...string) string {
	if len(seqs) == 0 {
		return item
	}
	var res strings.Builder
	for _, s := range seqs {
		res.WriteString(s)
	}
	res.WriteString(item)
	res.WriteString(ansiReset)
	return res.String()
}

func (self color) push_str(style Style, item string) string {
	switch self {
	case colorMonochrome:
		return item
	case colorDull:
		switch style {
		case StyleText:
			return item
		case StyleEmphasis:
			return styled(item, ansiBold, ansiUnderline)
		case StyleLiteral:
			return styled(item, ansiBold)
		case StyleMetavar:
			return styled(item, ansiUnderline)
		case StyleInvalid:
			return styled(item, ansiBold, ansiRed)
		}
	case colorBright:
		switch style {
		case StyleText:
			return item
		case StyleEmphasis:
			return styled(item, ansiBold, ansiYellow)
		case StyleLiteral:
			return styled(item, ansiBold, ansiGreen)
		case StyleMetavar:
			return styled(item, ansiBold, ansiBlue)
		case StyleInvalid:
			return styled(item, ansiBold, ansiRed)
		}
	}
	return ""
}

const padding = "                                                  "

// monochrome renders a monochrome version of the document.
//
// full indicates if the full message should be rendered, this makes
// difference for rendered help message, otherwise you can pass `true`.
func (self Doc) monochrome(full bool) string {
	return self.render_console(full, colorMonochrome, maxRenderWidth)
}

func (self Doc) render_console(full bool, color color, max_width uint) string {
	tabstop := 0
	byte_pos := 0
	{
		current := 0
		in_term := false
		// looking for widest term below maxTab
		for _, token := range self.tokens {
			switch {
			case token.isText:
				if in_term {
					current += len(self.payload[byte_pos : byte_pos+token.bytes])
				}
				byte_pos += token.bytes
			case token.isStart:
				if token.block == blockItemTerm {
					in_term = true
					current = 0
				}
			default:
				if token.block == blockItemTerm {
					in_term = false
					if current > tabstop && current <= maxTab {
						tabstop = current
					}
				}
			}
		}
		byte_pos = 0
	}
	tabstop += 4

	res := ""
	stack := []block{}
	skip := skip(0)
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
				switch {
				case chunk.i >= 0:
					s, w := chunk.s, uint(chunk.i)
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
								break
							}
						}
					}

					pushed := 0
					if int(margin) >= char_pos {
						missing := int(margin) - char_pos
						res += padding[:missing]
						char_pos = int(margin)
						pushed = missing
					}
					if pending_margin && char_pos >= maxTab+4 && pushed < 2 {
						missing := 2 - pushed
						res += padding[:missing]
						char_pos += missing
					}

					pending_newline = false
					pending_blank_line = false
					pending_margin = false

					res += color.push_str(style, s)
					char_pos += int(w)
				case chunk.i == -1:
					res += "\n"
					char_pos = 0
					if !full {
						skip.enable()
					}
				case chunk.i == -2:
					res += "\n"
					char_pos = 0
				}
			}
		case token.isStart:
			block := token.block
			stack = append(stack, block)
			margin := last(margins).OrDefault(0)

			switch block {
			case blockHeader, blockSection2:
				pending_newline = true
				margins = append(margins, margin)
			case blockSection3:
				pending_newline = true
				margins = append(margins, margin+2)
			case blockItemTerm:
				pending_newline = true
				margins = append(margins, margin+4)
			case blockItemBody:
				margins = append(margins, margin+uint(tabstop)+2)
				pending_margin = true
			case blockInlineBlock:
				skip.push()
			case blockBlock:
				margins = append(margins, margin)
			case blockDefinitionList, blockMeta, blockMono:
			case blockTermRef:
				if color.IsMonochrome() {
					res += "`"
					char_pos += 1
				}
			}
		default:
			block := token.block

			if len(margins) > 0 {
				margins = margins[:len(margins)-1]
			}
			switch block {
			case blockItemBody:
				pending_margin = false
			case blockHeader,
				blockSection2,
				blockSection3,
				blockItemTerm,
				blockDefinitionList,
				blockMeta,
				blockMono:
			case blockInlineBlock:
				skip.pop()
			case blockBlock:
				pending_blank_line = true
			case blockTermRef:
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

	for i := strings.Index(res, "\n\n\n"); i != -1; i = strings.Index(res, "\n\n") {
		res = strings.ReplaceAll(res, "\n\n\n", "\n\n")
	}
	return res
}
