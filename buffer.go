package cli

import (
	"strings"

	"github.com/rprtr258/fun"
)

func newDocFromPairs(val []fun.Pair[string, Style]) Doc {
	res := NewDoc()
	for _, v := range val {
		text, style := v.First, v.Second
		res.write_str(text, *style)
	}
	res
}

// Parser metainformation
//
// This is a newtype around internal parser metainfo representation, generated
// with [`Parserwith_group_help`](crateParserwith_group_help) and consumed by
// [`Docmeta`](Docmeta)
type MetaInfo = Meta

// Append a fragment of plain text to [`Doc`]
//
// See [`Doc`] for usage examples
func (self Doc) text(text string) {
	self.write_str(text, StyleText)
}

// Append a fragment of literal text to [`Doc`]
//
// See [`Doc`] for usage examples
func (self Doc) literal(text string) {
	self.write_str(text, StyleLiteral)
}

// Append a fragment of text with emphasis to [`Doc`]
//
// See [`Doc`] for usage examples
func (self Doc) emphasis(text string) {
	self.write_str(text, StyleEmphasis)
}

// Append a fragment of unexpected user input to [`Doc`]
//
// See [`Doc`] for usage examples
func (self Doc) invalid(text string) {
	self.write_str(text, StyleInvalid)
}

// Append a fragment of parser metadata to [`Doc`]
//
// See [`Doc`] for usage examples
func (self Doc) meta(meta MetaInfo, for_usage bool) {
	self.write_meta(meta[0], for_usage)
}

// Append a `Doc` to [`Doc`]
//
// See [`Doc`] for usage examples
func (self Doc) doc(buf Doc) {
	self.tokens.push(TokenBlockStart(BlockInlineBlock))
	self.tokens.extend(&buf.tokens)
	self.payload.push_str(&buf.payload)
	self.tokens.push(TokenBlockEnd(BlockInlineBlock))
}

// Append a `Doc` to [`Doc`] for plaintext documents try to format
// first line as a help section header
func (self Doc) em_doc(buf Doc) {
	self.tokens.push(TokenBlockStart(BlockInlineBlock))
	if v, ok := buf.tokens.first().copied().Unpack(); ok {
		bytes, style := v.bytes, v.stle
		prefix := buf.payload[0:bytes]
		if a, b, ok := strings.Cut(prefix, '\n'); ok {
			self.emphasis(a)
			self.tokens.push(TokenBlockStart(BlockSection3))
			self.text(b)

			if buf.tokens.len() > 1 {
				self.tokens.extend(buf.tokens[1:])
				self.payload.push_str(&buf.payload[bytes:])
			}
			self.tokens.push(TokenBlockEnd(BlockSection3))
		} else {
			self.emphasis(prefix)
		}
	} else {
		self.tokens.extend(&buf.tokens)
		self.payload.push_str(&buf.payload)
	}

	self.tokens.push(TokenBlockEnd(BlockInlineBlock))
}

func (self Doc) write_shortlong(name ShortLong) {
	name(
		func(s) {
			self.write_char('-', StyleLiteral)
			self.write_char(*s, StyleLiteral)
		},
		func(l) {
			self.write_str("--", StyleLiteral)
			self.write_str(l, StyleLiteral)
		},
		func(_, l) {
			self.write_str("--", StyleLiteral)
			self.write_str(l, StyleLiteral)
		},
	)
}

func (self Doc) write_item(item Item) {
	item(
		func(metavar Doc, _ bool, _ fun.Option[Doc]) {
			self.doc(metavar)
		},
		func(metavar Metavar, _ fun.Option[Doc]) {
			self.metavar(*metavar)
		},
		func(string, fun.Option[byte], fun.Option[Doc], Meta, Info) {
			self.write_str("COMMAND ...", StyleMetavar)
		},
		func(name ShortLong, _ []byte, _ fun.Option[string], _ fun.Option[Doc]) {
			self.write_shortlong(name)
		},
		func(name ShortLong, _ []byte, _ Metavar, _ fun.Option[string], _ fun.Option[Doc]) {
			self.write_shortlong(name)
			self.write_char('=', StyleText)
			self.metavar(*metavar)
		},
	)
}

func (self Doc) write_meta(meta Meta, for_usage bool) {
	gO := func(meta Meta, f *Doc) {
		meta(
			func(xs []Meta) {
				for ix, x := range xs {
					if ix != 0 {
						f.write_str(" ", StyleText)
					}
					gO(x, f)
				}
			},
			func(xs []Meta) {
				for ix, x := range xs {
					if ix != 0 {
						f.write_str(" | ", StyleText)
					}
					gO(x, f)
				}
			},
			func(m Meta) {
				f.write_str("[", StyleText)
				gO(m, f)
				f.write_str("]", StyleText)
			},
			func(m Meta) {
				f.write_str("(", StyleText)
				gO(m, f)
				f.write_str(")", StyleText)
			},
			func(i Item) { f.write_item(i) },
			func(m Meta) {
				gO(m, f)
				f.write_str("...", StyleText)
			},
			func(m Meta) { gO(m, f) },
			func(m Meta, _ Doc) { gO(m, f) },
			func(m Meta, _ Doc) { gO(m, f) },
			func() { /* f.write_str("no parameters expected", StyleText), */ },
			func(_, u Doc) { f.doc(u) },
			func(m Meta) {
				f.write_str("--", StyleLiteral)
				f.write_str(" ", StyleText)
				gO(m, f)
			},
		)
	}

	meta := meta.normalized(for_usage)
	self.token(TokenBlockStart(BlockMono))
	gO(&meta, self)
	self.token(TokenBlockEnd(BlockMono))
}

// Style of a text fragment inside of [`Doc`]
type Style uint8

const (
	StyleText     Style = iota // Plain text, no decorations
	StyleEmphasis              // Word with emphasis - things like "Usage", "Available fun.Options", etc
	StyleLiteral               // Something user needs to type literally - command names, etc
	StyleMetavar               // Metavavar placeholder - something user needs to replace with own input
	StyleInvalid               // Invalid input given by user - used to display invalid parts of the input
)

type Block uint8

const (
	BlockHeader Block = iota // level 1 section header, block for separate command inside manpage, not used in --help
	BlockSection2
	BlockSection3       // 2 margin level 3 section header, "group_help" header, etc.
	BlockItemTerm       // inline, 4 margin, no nesting -h, --help
	BlockItemBody       // widest term up below 20 margin margin plus two, but at least 4. print usage information, but also items inside Numbered/Unnumbered lists
	BlockDefinitionList // 0 margin Definition list,
	BlockBlock          // block of text, blocks are separated by a blank line in man or help can contain headers or other items inside
	BlockInlineBlock    // inserted when block is written into a block. single blank line in the input fast forward until the end of current inline block
	BlockTermRef        // inline displayed with `` in monochrome or not when rendered with colors. In markdown this becomes a link to a term if one is defined
	BlockMeta           // Surrounds metavars block in manpage used only inside render_manpage at the moment
	BlockMono           // Monospaced font that goes around [`Meta`]
)

type TokenText struct {
	bytes int
	style Style
}

type Token = func(
	func(Text TokenText),
	func(BlockStart Block),
	func(BlockEnd Block),
)

// String with styled segments.
//
// You can add style information to generated documentation and help messages
// For simpliest possible results you can also pass a string slice in all the places
// that require `impl Into<Doc>`
type Doc struct {
	payload string
	tokens  []Token
}

func newDocFromString(s string) Doc { return Doc{payload: s} }
func (d Doc) isEmpty() bool         { return len(d.tokens) == 0 }
func (d Doc) firstLine() (Doc, bool) {
	if d.isEmpty() {
		return Doc{}, false
	}

	res := Doc{}
	cur := 0
	doBreak := false
	for _, t := range d.tokens {
		t(func(Text TokenText) {
			payload := d.payload[cur : cur+Text.bytes]
			first, _, ok := strings.Cut(payload, "\n")
			if ok {
				res.payload += first
				res.tokens = append(res.tokens, func(ftext func(Text TokenText), _, _ func(BlockEnd Block)) {
					if ftext != nil {
						ftext(TokenText{len(first), Text.style})
					}
				})
				return
			} else {
				res.payload += d.payload[cur : cur+Text.bytes]
				res.tokens = append(res.tokens, t)
				cur += Text.bytes
			}
		}, func(Block) { doBreak = true }, func(Block) { doBreak = true })
		if doBreak {
			break
		}
	}
	return res, true
}

type Skip uint

func (s Skip) enabled() bool { return s > 0 }
func (s *Skip) enable()      { *s = 1 }
func (s *Skip) push() {
	if *s > 0 {
		*s++
	}
}
func (s *Skip) pop() {
	if *s > 0 {
		*s--
	}
}

func (self Doc) token(token Token) {
	self.tokens.push(token)
}

func (self Doc) write(input fmt.Display, style Style) {
	old_len := self.payload.len()
	_ := write(self.payload, "{}", input)
	self.set_style(self.payload.len()-old_len, style)
}

func (self Doc) set_style(len uint, style Style) {
	// buffer chunks are unified with previous chunks of the same type
	// [metavar]"foo" + [metavar]"bar"[metavar]"foobar"
	if n := len(self.tokens); n == 0 {
		self.tokens.push(TokenText{bytes: len, style})
	} else {
		if self.tokens[n-1].style == style {
			self.tokens[n-1].bytes += len
		}
	}
}

func (self Doc) write_str(input string, style Style) {
	self.payload.push_str(input)
	self.set_style(input.len(), style)
}

func (self Doc) write_char(c char, style Style) {
	self.write_str(c.encode_utf8([4]byte{}), style)
}

type DocSection struct {
	// Path name to the command name starting from the application
	path []string
	info Info
	meta Meta
}

func extract_sections(
	meta Meta,
	info Info,
	path *[]string,
	sections *[]DocSection,
) {
	*sections = append(*sections, DocSection{*path, info, meta})
	hi := HelpItemsDefault()
	hi.append_meta(meta)
	for item := range hi.items {
		item(nil, nil, nil, nil, nil, func(name string, _ fun.Option[byte], _ fun.Option[Doc], meta Meta, info Info) {
			*path = append(*path, name)
			extract_sections(meta, info, path, sections)
			*path = (*path)[0 : len(*path)-1]
		}, nil, nil, nil, nil)
	}
}
