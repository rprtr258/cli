package cli

import (
	"strings"

	"github.com/rprtr258/fun"
)

func NewDocFromPairs(val []fun.Pair[string, Style]) Doc {
	res := Doc{}
	for _, v := range val {
		text, style := v.K, v.V
		res.write_str(text, style)
	}
	return res
}

// Append a fragment of plain Text to [Doc]
//
// See [Doc] for usage examples
func (self *Doc) Text(text string) {
	self.write_str(text, StyleText)
}

// Append a fragment of Literal text to [Doc]
//
// See [Doc] for usage examples
func (self *Doc) Literal(text string) {
	self.write_str(text, StyleLiteral)
}

// Append a fragment of text with Emphasis to [Doc]
//
// See [Doc] for usage examples
func (self *Doc) Emphasis(text string) {
	self.write_str(text, StyleEmphasis)
}

// Append a fragment of unexpected user input to [Doc]
//
// See [Doc] for usage examples
func (self *Doc) Invalid(text string) {
	self.write_str(text, StyleInvalid)
}

// Append a fragment of parser metadata to [Doc]
//
// See [Doc] for usage examples
func (self *Doc) Meta(meta Meta, for_usage bool) {
	self.write_meta(meta, for_usage)
}

// Append a `Doc` to [Doc]
//
// See [Doc] for usage examples
func (self *Doc) Doc(buf Doc) {
	self.tokens = append(self.tokens, TokenBlockStart(BlockInlineBlock))
	self.tokens = append(self.tokens, buf.tokens...)
	self.payload += buf.payload
	self.tokens = append(self.tokens, TokenBlockEnd(BlockInlineBlock))
}

// Append a `Doc` to [Doc] for plaintext documents try to format
// first line as a help section header
func (self *Doc) EmDoc(buf Doc) {
	self.tokens = append(self.tokens, TokenBlockStart(BlockInlineBlock))
	if v, ok := First(buf.tokens).Unpack(); ok && v.isText {
		bytes := v.bytes
		prefix := buf.payload[0:bytes]
		if a, b, ok := strings.Cut(prefix, "\n"); ok {
			self.Emphasis(a)
			self.tokens = append(self.tokens, TokenBlockStart(BlockSection3))
			self.Text(b)

			if len(buf.tokens) > 1 {
				self.tokens = append(self.tokens, buf.tokens[1:]...)
				self.payload += buf.payload[bytes:]
			}
			self.tokens = append(self.tokens, TokenBlockEnd(BlockSection3))
		} else {
			self.Emphasis(prefix)
		}
	} else {
		self.tokens = append(self.tokens, buf.tokens...)
		self.payload += buf.payload
	}
	self.tokens = append(self.tokens, TokenBlockEnd(BlockInlineBlock))
}

func (self *Doc) write_shortlong(name ShortLong) {
	switch {
	case name.is_short() || name.is_both():
		self.write_char('-', StyleLiteral)
		self.write_char(name.Short, StyleLiteral)
	case name.is_long():
		l := name.Long
		self.write_str("--", StyleLiteral)
		self.write_str(l, StyleLiteral)
	}
}

func (self *Doc) write_item(item Item) {
	switch v := item.v.(type) {
	case ItemAny:
		self.Doc(v.metavar)
	case ItemPositional:
		self.metavar(v.metavar)
	case ItemCommand:
		self.write_str("COMMAND ...", StyleMetavar)
	case ItemFlag:
		self.write_shortlong(v.name)
	case ItemArgument:
		self.write_shortlong(v.name)
		self.write_char('=', StyleText)
		self.metavar(v.metavar)
	}
}

func (self *Doc) write_meta_go(meta Meta) {
	switch m := meta.(type) {
	case MetaAnd:
		for ix, x := range m {
			if ix != 0 {
				self.write_str(" ", StyleText)
			}
			self.write_meta_go(x)
		}
	case MetaOr:
		for ix, x := range m {
			if ix != 0 {
				self.write_str(" | ", StyleText)
			}
			self.write_meta_go(x)
		}
	case MetaOptional:
		self.write_str("[", StyleText)
		self.write_meta_go(m.m)
		self.write_str("]", StyleText)
	case MetaRequired:
		self.write_str("(", StyleText)
		self.write_meta_go(m.m)
		self.write_str(")", StyleText)
	case MetaItem:
		self.write_item(Item(m))
	case MetaMany:
		self.write_meta_go(m.m)
		self.write_str("...", StyleText)
	case MetaSkip:
		// self.write_str("no parameters expected", StyleText)
	case MetaCustomUsage:
		self.Doc(m.doc)
	case MetaStrict:
		self.write_str("--", StyleLiteral)
		self.write_str(" ", StyleText)
		self.write_meta_go(m.m)
	case MetaAdjacent:
		self.write_meta_go(m.m)
	case MetaSubsection:
		self.write_meta_go(m.m)
	case MetaSuffix:
		self.write_meta_go(m.m)
	}
}

func (self *Doc) write_meta(meta Meta, for_usage bool) {
	meta = normalized(meta, for_usage)
	self.token(TokenBlockStart(BlockMono))
	self.write_meta_go(meta)
	self.token(TokenBlockEnd(BlockMono))
}

// Style of a text fragment inside of [Doc]
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
	BlockMono           // Monospaced font that goes around [Meta]
)

// Token = Text | BlockStart | BlockEnd
type Token struct {
	isText bool
	bytes  int
	style  Style

	isStart bool
	block   Block
}

func TokenText(bytes int, style Style) Token { return Token{isText: true, bytes: bytes, style: style} }
func TokenBlockStart(block Block) Token      { return Token{isStart: true, block: block} }
func TokenBlockEnd(block Block) Token        { return Token{isStart: false, block: block} }

// String with styled segments.
//
// You can add style information to generated documentation and help messages
// For simpliest possible results you can also pass a string slice in all the places
// that require `impl Into<Doc>`
type Doc struct {
	payload string
	tokens  []Token
}

func (d Doc) clone() Doc {
	return Doc{
		d.payload,
		append([]Token(nil), d.tokens...),
	}
}

func (d Doc) String() string {
	return d.render_console(true, Colordefault, MAX_WIDTH)
}

func (d Doc) to_completion() Option[string] {
	l, ok := d.first_line()
	if !ok {
		return Invalid[string]()
	}

	s := l.monochrome(false)
	return Valid(strings.TrimRight(s, " "))
}

func NewDoc(s string) Doc {
	var res Doc
	res.write_str(s, StyleText)
	return res
}
func (d Doc) is_empty() bool { return len(d.tokens) == 0 }
func (d Doc) first_line() (Doc, bool) {
	if d.is_empty() {
		return Doc{}, false
	}

	res := Doc{}
	cur := 0
	for _, t := range d.tokens {
		if !t.isText {
			break
		}

		bytes, style := t.bytes, t.style
		payload := d.payload[cur : cur+bytes]
		if first, _, ok := strings.Cut(payload, "\n"); ok {
			res.tokens = append(res.tokens, TokenText(len(first), style))
			res.payload += first
		} else {
			res.tokens = append(res.tokens, t)
			res.payload += d.payload[cur : cur+bytes]
			cur += bytes
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

func (self *Doc) token(token Token) {
	self.tokens = append(self.tokens, token)
}

func (self *Doc) write(input string, style Style) {
	old_len := len(self.payload)
	self.payload += input
	self.set_style(len(self.payload)-old_len, style)
}

func (self *Doc) set_style(lenn int, style Style) {
	// buffer chunks are unified with previous chunks of the same type
	// [metavar]"foo" + [metavar]"bar" => [metavar]"foobar"
	if n := len(self.tokens); n != 0 && self.tokens[n-1].isText && self.tokens[n-1].style == style {
		self.tokens[n-1].bytes += lenn
	} else {
		self.tokens = append(self.tokens, TokenText(lenn, style))
	}
}

func (self *Doc) write_str(input string, style Style) {
	self.payload += input
	self.set_style(len(input), style)
}

func (self *Doc) write_char(c byte, style Style) {
	self.write_str(string(c), style)
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
	hi := HelpItems{}
	hi.append_meta(meta)
	for _, item := range hi {
		switch item := item.(type) {
		case HelpItemCommand:
			*path = append(*path, item.name)
			extract_sections(item.meta, item.info, path, sections)
			*path = (*path)[0 : len(*path)-1]
		}
	}
}
