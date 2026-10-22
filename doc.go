package cli

import (
	"strings"

	"github.com/rprtr258/fun"
)

// String with styled segments.
//
// A [Doc] accumulates styled text via [Doc.Text], [Doc.Emphasis], [Doc.Literal] and
// [Doc.Invalid] helpers and renders it into help messages and
// usage lines. For the simplest possible results you can pass a plain string
// wherever a [Doc] is required.
type Doc struct {
	payload string
	tokens  []token
}

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
func (self *Doc) meta(meta meta, for_usage bool) {
	self.write_meta(meta, for_usage)
}

// Append a `Doc` to [Doc]
//
// See [Doc] for usage examples
func (self *Doc) Doc(buf Doc) {
	self.tokens = append(self.tokens, tokenBlockStart(blockInlineBlock))
	self.tokens = append(self.tokens, buf.tokens...)
	self.payload += buf.payload
	self.tokens = append(self.tokens, tokenBlockEnd(blockInlineBlock))
}

// Append a `Doc` to [Doc] for plaintext documents try to format
// first line as a help section header
func (self *Doc) emDoc(buf Doc) {
	self.tokens = append(self.tokens, tokenBlockStart(blockInlineBlock))
	if v, ok := first(buf.tokens).Unpack(); ok && v.isText {
		bytes := v.bytes
		prefix := buf.payload[0:bytes]
		if a, b, ok := strings.Cut(prefix, "\n"); ok {
			self.Emphasis(a)
			self.tokens = append(self.tokens, tokenBlockStart(blockSection3))
			self.Text(b)

			if len(buf.tokens) > 1 {
				self.tokens = append(self.tokens, buf.tokens[1:]...)
				self.payload += buf.payload[bytes:]
			}
			self.tokens = append(self.tokens, tokenBlockEnd(blockSection3))
		} else {
			self.Emphasis(prefix)
		}
	} else {
		self.tokens = append(self.tokens, buf.tokens...)
		self.payload += buf.payload
	}
	self.tokens = append(self.tokens, tokenBlockEnd(blockInlineBlock))
}

func (self *Doc) write_shortlong(name shortLong) {
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

func (self *Doc) write_item(item item) {
	switch v := item.v.(type) {
	case itemAny:
		self.Doc(v.metavar)
	case itemPositional:
		self.metavar(v.metavar)
	case itemCommand:
		self.write_str("COMMAND ...", StyleMetavar)
	case itemFlag:
		self.write_shortlong(v.name)
	case itemArgument:
		self.write_shortlong(v.name)
		self.write_char('=', StyleText)
		self.metavar(v.metavar)
	}
}

func (self *Doc) write_meta_go(meta meta) {
	switch m := meta.(type) {
	case metaAnd:
		for ix, x := range m {
			if ix != 0 {
				self.write_str(" ", StyleText)
			}
			self.write_meta_go(x)
		}
	case metaOr:
		for ix, x := range m {
			if ix != 0 {
				self.write_str(" | ", StyleText)
			}
			self.write_meta_go(x)
		}
	case metaOptional:
		self.write_str("[", StyleText)
		self.write_meta_go(m.m)
		self.write_str("]", StyleText)
	case metaRequired:
		self.write_str("(", StyleText)
		self.write_meta_go(m.m)
		self.write_str(")", StyleText)
	case metaItem:
		self.write_item(item(m))
	case metaMany:
		self.write_meta_go(m.m)
		self.write_str("...", StyleText)
	case metaSkip:
	case metaCustomUsage:
		self.Doc(m.doc)
	case metaStrict:
		self.write_str("--", StyleLiteral)
		self.write_str(" ", StyleText)
		self.write_meta_go(m.m)
	case metaAdjacent:
		self.write_meta_go(m.m)
	case metaSubsection:
		self.write_meta_go(m.m)
	case metaSuffix:
		self.write_meta_go(m.m)
	}
}

func (self *Doc) write_meta(meta meta, for_usage bool) {
	meta = normalized(meta, for_usage)
	self.token(tokenBlockStart(blockMono))
	self.write_meta_go(meta)
	self.token(tokenBlockEnd(blockMono))
}

// style of a text fragment inside of [Doc]
type Style uint8

const (
	StyleText     Style = iota // Plain text, no decorations
	StyleEmphasis              // Word with emphasis - things like "Usage", "Available options", etc
	StyleLiteral               // Something user needs to type literally - command names, etc
	StyleMetavar               // Metavavar placeholder - something user needs to replace with own input
	StyleInvalid               // Invalid input given by user - used to display invalid parts of the input
)

type block uint8

const (
	blockHeader block = iota // level 1 section header, block for separate command inside manpage, not used in --help
	blockSection2
	blockSection3       // 2 margin level 3 section header, "group_help" header, etc.
	blockItemTerm       // inline, 4 margin, no nesting -h, --help
	blockItemBody       // widest term up below 20 margin margin plus two, but at least 4. print usage information, but also items inside Numbered/Unnumbered lists
	blockDefinitionList // 0 margin Definition list,
	blockBlock          // block of text, blocks are separated by a blank line in man or help can contain headers or other items inside
	blockInlineBlock    // inserted when block is written into a block. single blank line in the input fast forward until the end of current inline block
	blockTermRef        // inline displayed with `` in monochrome or not when rendered with colors. In markdown this becomes a link to a term if one is defined
	blockMeta           // Surrounds metavars block in manpage used only inside render_manpage at the moment
	blockMono           // Monospaced font that goes around [meta]
)

// token = Text | blockStart | blockEnd
type token struct {
	isText bool
	bytes  int
	style  Style

	isStart bool
	block   block
}

func tokenText(bytes int, style Style) token { return token{isText: true, bytes: bytes, style: style} }
func tokenBlockStart(block block) token      { return token{isStart: true, block: block} }
func tokenBlockEnd(block block) token        { return token{isStart: false, block: block} }

func (d Doc) clone() Doc {
	return Doc{
		d.payload,
		append([]token(nil), d.tokens...),
	}
}

func (d Doc) String() string {
	return d.render_console(true, colorDefault, maxRenderWidth)
}

func (d Doc) to_completion() Opt[string] {
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
			res.tokens = append(res.tokens, tokenText(len(first), style))
			res.payload += first
		} else {
			res.tokens = append(res.tokens, t)
			res.payload += d.payload[cur : cur+bytes]
			cur += bytes
		}
	}
	return res, true
}

type skip uint

func (s skip) enabled() bool { return s > 0 }
func (s *skip) enable()      { *s = 1 }
func (s *skip) push() {
	if *s > 0 {
		*s++
	}
}

func (s *skip) pop() {
	if *s > 0 {
		*s--
	}
}

func (self *Doc) token(token token) {
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
		self.tokens = append(self.tokens, tokenText(lenn, style))
	}
}

func (self *Doc) write_str(input string, style Style) {
	self.payload += input
	self.set_style(len(input), style)
}

func (self *Doc) write_char(c byte, style Style) {
	self.write_str(string(c), style)
}

type docSection struct {
	// Path name to the command name starting from the application
	path []string
	info *info
	meta meta
}

func extract_sections(
	meta meta,
	info *info,
	path *[]string,
	sections *[]docSection,
) {
	// clone the path: docSection must own its slice, otherwise the next
	// append(path, ...) overwrites the backing array of previously pushed sections
	p := make([]string, len(*path))
	copy(p, *path)
	*sections = append(*sections, docSection{p, info, meta})
	hi := helpItems{}
	hi.append_meta(meta)
	for _, item := range hi {
		if cmd, ok := item.(helpItemCommand); ok {
			*path = append(*path, cmd.name)
			extract_sections(cmd.meta, &cmd.info, path, sections)
			*path = (*path)[0 : len(*path)-1]
		}
	}
}
