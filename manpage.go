package cli

import (
	"fmt"
	"strings"

	"github.com/rprtr258/fun"

	"github.com/rprtr258/cli/roff"
)

type SectionKind uint8

const (
	SectionGeneral         SectionKind = iota // General commands
	SectionSystemCall                         // System calls
	SectionLibraryFunction                    // Library functions such as C standard library functions
	SectionSpecialFile                        // Special files (usually devices in /dev) and drivers
	SectionFileFormat                         // File formats and conventions
	SectionGame                               // Games and screensavers
	SectionMisc                               // Miscellaneous
	SectionSysadmin                           // System administration commands and daemons
	SectionCustom                             // Custom section
)

// Manual page section
type Section struct {
	kind SectionKind
	s    string
}

func (self Section) String() string {
	switch self.kind {
	case SectionGeneral:
		return "1"
	case SectionSystemCall:
		return "2"
	case SectionLibraryFunction:
		return "3"
	case SectionSpecialFile:
		return "4"
	case SectionFileFormat:
		return "5"
	case SectionGame:
		return "6"
	case SectionMisc:
		return "7"
	case SectionSysadmin:
		return "8"
	case SectionCustom:
		return self.s
	}
	panic("unreachable")
}

// Render command line documentation for the app into a manpage
func (self OptionParser[T]) render_manpage(
	app string,
	section Section,
	last_update_date Option[string],
	vendor Option[string],
	application_title Option[string],
) string {
	sections := []DocSection{}
	root := self.inner.meta
	path := []string{app}

	extract_sections(root, self.info, &path, &sections)

	var buf Doc

	if len(sections) > 1 {
		buf.token(TokenBlockStart(BlockBlock))
		buf.token(TokenBlockStart(BlockHeader))
		buf.Text("SYNOPSIS")
		buf.token(TokenBlockEnd(BlockHeader))
		buf.token(TokenBlockEnd(BlockBlock))

		buf.token(TokenBlockStart(BlockMeta))
		for _, section := range sections {
			for _, p := range section.path {
				buf.Literal(p)
				buf.Text(" ")
			}

			buf.write_meta(section.meta, true)
			buf.Text("\n")
		}
		buf.token(TokenBlockEnd(BlockMeta))
	}

	// NAME
	// SYNOPSIS
	// DESCRIPTION
	// EXIT STATUS
	// REPORTING BUGS
	// EXAMPLES
	// SEE ALSO

	for _, section := range sections {
		if len(sections) > 1 {
			buf.token(TokenBlockStart(BlockHeader))
			buf.write_path(section.path)
			buf.token(TokenBlockEnd(BlockHeader))
		}

		if descr, ok := section.info.descr.Unpack(); ok {
			buf.token(TokenBlockStart(BlockHeader))
			buf.Text("NAME")
			buf.token(TokenBlockEnd(BlockHeader))

			buf.Text(app)
			buf.Text(" - ")
			buf.Doc(descr)
		}

		buf.token(TokenBlockStart(BlockHeader))
		buf.Text("SYNOPSIS")
		buf.token(TokenBlockEnd(BlockHeader))
		buf.write_path(section.path)
		buf.write_meta(section.meta, true)

		if t, ok := section.info.header.Unpack(); ok {
			buf.token(TokenBlockStart(BlockBlock))
			buf.Doc(t)
			buf.token(TokenBlockEnd(BlockBlock))
		}

		var items HelpItems
		items.append_meta(section.meta)
		help_meta := section.info.meta()
		items.append_meta(help_meta)
		buf.write_help_item_groups(items, false)

		if footer, ok := section.info.footer.Unpack(); ok {
			buf.token(TokenBlockStart(BlockBlock))
			buf.Doc(footer)
			buf.token(TokenBlockEnd(BlockBlock))
		}
	}

	manpage := roff.NewDocument()
	fmt.Fprintf(&manpage.B,
		"TH %s %s %s %s %s\n",
		app,
		section.String(),
		vendor.OrDefault("-"),
		application_title.OrDefault(""),
		last_update_date.OrDefault("-"),
	)

	return buf.render_roff(manpage)
}

type Font = string

func Fontfrom(value Style) Font {
	switch value {
	case StyleText:
		return ""
	case StyleEmphasis, StyleLiteral:
		return roff.Bold
	case StyleMetavar, StyleInvalid:
		return roff.Italic
	}
	panic("unreachable")
}

func (self Doc) render_roff(roff *roff.Document) string {
	// sections and subsections are implemented with .SH and .SS
	// control messages and it is easier to provide them right away
	// We also strip styling from them and change sections to all caps
	capture := fun.Pair[string, bool]{"", false}

	byte_pos := 0
	for _, token := range self.tokens {
		switch {
		case token.isText:
			bytes, style := token.bytes, token.style
			input := self.payload[byte_pos : byte_pos+bytes]
			byte_pos += bytes

			if capture.V {
				capture.K += input
				continue
			} else {
				if style == StyleEmphasis {
					roff.B.WriteString("SS")
				}
				roff.Text(Fontfrom(style) + input)
			}
		case token.isStart:
			block := token.block
			switch block {
			case BlockHeader, BlockSection2, BlockSection3:
				capture.V = true
			case BlockItemTerm:
				roff.B.WriteString("TP")
			case BlockMono,
				BlockItemBody,
				BlockDefinitionList,
				BlockInlineBlock:
			case BlockBlock:
				roff.Paragraph()
			case BlockMeta:
				roff.B.WriteString("nf")
			case BlockTermRef:
				panic("TODO: ")
			}
		default:
			block := token.block
			switch block {
			case BlockHeader:
				capture.V = false
				roff.Section(capture.K)
				capture.K = ""
			case BlockSection2, BlockSection3:
				capture.V = false
				fmt.Fprint(&roff.B, "SS", []string{strings.ToUpper(capture.K)})
				capture.K = ""
			case BlockItemTerm:
				fmt.Fprint(&roff.B, "\n")
			case BlockItemBody:
				roff.Paragraph()
			case BlockMono | BlockDefinitionList | BlockBlock | BlockInlineBlock:
			case BlockMeta:
				fmt.Fprint(&roff.B, "fi")
			case BlockTermRef:
				panic("TODO: ")
			}
		}
	}

	return roff.String( /*ApostrophesHandle*/ )
}
