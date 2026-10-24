package cli

import (
	"strings"

	"github.com/rprtr258/fun"

	"github.com/rprtr258/cli/roff"
)

type sectionKind uint8

const (
	SectionGeneral         sectionKind = iota // General commands
	SectionSystemCall                         // System calls
	SectionLibraryFunction                    // Library functions such as C standard library functions
	SectionSpecialFile                        // Special files (usually devices in /dev) and drivers
	SectionFileFormat                         // File formats and conventions
	SectionGame                               // Games and screensavers
	SectionMisc                               // Miscellaneous
	SectionSysadmin                           // System administration commands and daemons
	SectionCustom                             // Custom section
)

// Section is a fragment of the manpage: a section kind plus the text.
type Section struct {
	Kind sectionKind
	Text string
}

func (self Section) String() string {
	switch self.Kind {
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
		return self.Text
	}
	panic("unreachable")
}

// RenderManpage renders command line documentation for the app into a manpage.
func (self OptionParser[T]) RenderManpage(
	app string,
	section Section,
	last_update_date Opt[string],
	vendor Opt[string],
	application_title Opt[string],
) string {
	sections := []docSection{}
	root := self.inner.meta
	path := []string{app}

	extract_sections(root, &self.info, &path, &sections)

	var buf Doc

	if len(sections) > 1 {
		buf.token(tokenBlockStart(blockBlock))
		buf.token(tokenBlockStart(blockHeader))
		buf.Text("SYNOPSIS")
		buf.token(tokenBlockEnd(blockHeader))
		buf.token(tokenBlockEnd(blockBlock))

		buf.token(tokenBlockStart(blockMeta))
		for _, section := range sections {
			for _, p := range section.path {
				buf.Literal(p)
				buf.Text(" ")
			}

			buf.write_meta(section.meta, true)
			buf.Text("\n")
		}
		buf.token(tokenBlockEnd(blockMeta))
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
			buf.token(tokenBlockStart(blockHeader))
			buf.write_path(section.path)
			buf.token(tokenBlockEnd(blockHeader))
		}

		if descr, ok := section.info.descr.Unpack(); ok {
			buf.token(tokenBlockStart(blockHeader))
			buf.Text("NAME")
			buf.token(tokenBlockEnd(blockHeader))

			buf.Text(app)
			buf.Text(" - ")
			buf.Doc(descr)
		}

		buf.token(tokenBlockStart(blockHeader))
		buf.Text("SYNOPSIS")
		buf.token(tokenBlockEnd(blockHeader))
		buf.write_path(section.path)
		buf.write_meta(section.meta, true)

		if t, ok := section.info.header.Unpack(); ok {
			buf.token(tokenBlockStart(blockBlock))
			buf.Doc(t)
			buf.token(tokenBlockEnd(blockBlock))
		}

		var items helpItems
		items.append_meta(section.meta)
		help_meta := section.info.meta()
		items.append_meta(help_meta)
		buf.write_help_item_groups(items, false)

		if examples := section.info.examples; len(examples) > 0 {
			buf.token(tokenBlockStart(blockHeader))
			buf.Text("EXAMPLES")
			buf.token(tokenBlockEnd(blockHeader))

			buf.write_examples(section.path, examples)
		}

		if footer, ok := section.info.footer.Unpack(); ok {
			buf.token(tokenBlockStart(blockBlock))
			buf.Doc(footer)
			buf.token(tokenBlockEnd(blockBlock))
		}
	}

	manpage := roff.New()
	manpage.Control(
		"TH",
		app,
		section.String(),
		last_update_date.OrDefault("-"),
		vendor.OrDefault("-"),
		application_title.OrDefault(""),
	)

	return buf.render_roff(manpage)
}

type Font = roff.Font

func FontFrom(value Style) Font {
	switch value {
	case StyleText:
		return roff.Roman
	case StyleEmphasis, StyleLiteral:
		return roff.Bold
	case StyleMetavar, StyleInvalid:
		return roff.Italic
	}
	panic("unreachable")
}

func (self Doc) render_roff(roffDoc *roff.Roff) string {
	// Sections and subsections are implemented with .SH and .SS control
	// messages and it is easier to provide them right away. Styling is
	// stripped and section names are changed to all caps.
	capture := fun.Pair[string, bool]{K: "", V: false}

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
					roffDoc.Control0("SS")
				}
				roffDoc.Text(roff.FontText{Font: FontFrom(style), Text: input})
			}
		case token.isStart:
			block := token.block
			switch block {
			case blockHeader, blockSection2, blockSection3:
				capture.V = true
			case blockItemTerm:
				roffDoc.Control0("TP").StripNewlines(true)
			case blockMono,
				blockItemBody,
				blockDefinitionList,
				blockInlineBlock:
			case blockBlock:
				roffDoc.Control0("PP")
			case blockMeta:
				roffDoc.Control0("nf")
			case blockTermRef:
				// Unreachable: TermRef is only emitted by error rendering, which
				// never produces manpages.
				panic("TermRef is not supported in manpage rendering")
			}
		default:
			block := token.block
			switch block {
			case blockHeader:
				capture.V = false
				roffDoc.Control("SH", strings.ToUpper(capture.K))
				capture.K = ""
			case blockSection2, blockSection3:
				capture.V = false
				roffDoc.Control("SS", strings.ToUpper(capture.K))
				capture.K = ""
			case blockItemTerm:
				roffDoc.RoffLinebreak().StripNewlines(false)
			case blockItemBody:
				roffDoc.Control0("PP").StripNewlines(false)
			case blockMono, blockDefinitionList, blockBlock, blockInlineBlock:
			case blockMeta:
				roffDoc.Control0("fi")
			case blockTermRef:
				// Unreachable: TermRef is only emitted by error rendering, which
				// never produces manpages.
				panic("TermRef is not supported in manpage rendering")
			}
		}
	}

	return roffDoc.Render(roff.Handle)
}
