// Package roff implements a faithful port of bpaf's ROFF document composer
//
// [ROFF] is a family of Unix text-formatting languages, implemented
// by the `nroff`, `troff`, and `groff` programs, among others. This
// structure is an abstract representation of a document in ROFF format.
// It is meant for writing code to generate ROFF documents, such as
// manual pages.
//
// [ROFF]: https://en.wikipedia.org/wiki/Roff_(software)
package roff

// Apostrophes handling configuration
//
// To generate manpages you most likely want to have this in Handle state.
// See https://bugs.debian.org/cgi-bin/bugreport.cgi?bug=507673#65 for more details
type Apostrophes uint8

const (
	// Replace apostrophes with special code that
	Handle Apostrophes = iota
	// Leave apostrophes as is
	DontHandle
)

// Replacement for apostrophes, if active
//
// https://bugs.debian.org/cgi-bin/bugreport.cgi?bug=507673#65
const apostrophe = `\*(Aq`

// A preamble added to the start of rendered output.
//
// This defines a string variable that contains an apostrophe. For
// historical reasons, there seems to be no other portable way to
// represent apostrophes across various implementations of the ROFF
// language. In implementations that produce output like PostScript
// or PDF, an apostrophe gets typeset as a right single quote, which
// looks different from an apostrophe. For terminal output ("ASCII"),
// such as when using nroff, an apostrophe looks indistinguishable
// from a right single quote. For manual pages, and similar content,
// an apostrophe is more generally desired than the right single
// quote, so we convert all apostrophe characters in input text into
// a use of the string variable defined in the preamble.
//
// Used when apostrophe handle is enabled.
//
// See: https://bugs.debian.org/cgi-bin/bugreport.cgi?bug=507673#65
const APOSTROPHE_PREABMLE = `.ie \n(.g .ds Aq \(aq
.el .ds Aq '
`

// Escaping rules
type Escape uint8

const (
	// Insert a newline character unless on a new line already
	EscapeUnescapedAtNewline Escape = iota

	// Escape space characters (`' '`) and newlines (`'\n'`)
	// This escape is used for control sequence arguments
	EscapeSpaces

	// Escape characters roff considers special:
	//   - `' '`, `'.'` and `'\''` at the beginning of the line,
	//   - `'-'` and `'\\'` when inside the body
	//   - replace `'\''` with APOSTROPHE if Apostrophes handling is enabled
	EscapeSpecial

	// Similar to [EscapeSpecial] but also replaces `'\n'` with `' '`
	EscapeSpecialNoNewline

	// Input is written as is
	EscapeUnescaped
)

// Escape a sequence of string slices according to escaping rules
//
// Writes results to out, result should be a valid utf8 string as long as out
// starts empty or contains a valid utf8 sequence.
func escapeSlice(payload string, labels []label, out *[]byte, ap Apostrophes) {
	at_line_start := true
	for _, l := range labels {
		meta := l.kind
		p := payload[l.start:l.end]
		if !at_line_start && meta == EscapeUnescapedAtNewline {
			*out = append(*out, '\n')
			at_line_start = true
		}
		for i := 0; i < len(p); i++ {
			c := p[i]
			switch meta {
			case EscapeSpaces:
				if c == ' ' || c == '\n' {
					*out = append(*out, '\\', ' ')
				} else {
					*out = append(*out, c)
				}
			case EscapeSpecial, EscapeSpecialNoNewline:
				if at_line_start && (c == '.' || c == '\'') {
					*out = append(*out, '\\', '&')
				}
				if c == '\\' || c == '-' {
					*out = append(*out, '\\')
				}
				switch {
				case ap == Handle && c == '\'':
					*out = append(*out, apostrophe...)
				case meta == EscapeSpecialNoNewline && c == '\n':
					*out = append(*out, ' ')
					at_line_start = false
					continue
				default:
					*out = append(*out, c)
				}
			case EscapeUnescaped, EscapeUnescapedAtNewline:
				*out = append(*out, c)
			}
			at_line_start = c == '\n'
		}
	}
}

// label marks a byte range of payload with an escape kind
type label struct {
	start, end int
	kind       Escape
}

// FreeMonoid on a set of annotated string slices: payload plus labels
type freeMonoid struct {
	payload string
	labels  []label
}

func (self *freeMonoid) pushStr(kind Escape, payload string) {
	start := len(self.payload)
	self.payload += payload
	// no squashing: control lines and font runs are kept separate like upstream
	self.labels = append(self.labels, label{start, len(self.payload), kind})
}

func (self *freeMonoid) payloadSize() int {
	return len(self.payload)
}

func (self *freeMonoid) iter() []label {
	return self.labels
}

// Font selector
type Font uint8

const (
	// Roman font - regular font that should be used for most of the text
	Roman Font = iota

	// Bold font
	Bold

	// Italic font
	Italic
)

// Escape code needed to set this font
func (self Font) escape() string {
	switch self {
	case Bold:
		return `\fB`
	case Italic:
		return `\fI`
	case Roman:
		return `\fR`
	}
	panic("unreachable")
}

// Escape code used to return to the previous font
const RESTORE_FONT = `\fP`

// Roff is an abstract representation of a document in ROFF format.
type Roff struct {
	payload freeMonoid
	// keep or strip newlines from inserted text
	stripNewlines bool
}

// New creates a new raw Roff document
func New() *Roff {
	return &Roff{}
}

// Chainable setter for `stripNewlines` field
//
// `stripNewlines` specifies if [Roff.Render] should keep all the newline characters
// inside added text newlines can come with a special meaning, for example adding a
// section header relies, newlines are automatically stripped from [Roff.Control]
// arguments.
func (self *Roff) StripNewlines(state bool) *Roff {
	self.stripNewlines = state
	return self
}

// Insert a raw control sequence
//
// `name` should not contain initial `'.'`.
// Arguments are escaped accordingly
func (self *Roff) Control(name string, args ...string) *Roff {
	self.payload.pushStr(EscapeUnescapedAtNewline, ".")
	self.payload.pushStr(EscapeUnescaped, name)
	for _, arg := range args {
		// empty macro argument can be specified as "", mostly useful for TH macro
		// which takes several arguments positionally
		s := arg
		if s == "" {
			s = `""`
		}
		self.payload.pushStr(EscapeUnescaped, " ")
		self.payload.pushStr(EscapeSpaces, s)
	}
	self.payload.pushStr(EscapeUnescapedAtNewline, "")
	return self
}

// Control0 is a variant of Control that takes no parameters
func (self *Roff) Control0(name string) *Roff {
	self.payload.pushStr(EscapeUnescapedAtNewline, ".")
	self.payload.pushStr(EscapeUnescaped, name)
	self.payload.pushStr(EscapeUnescapedAtNewline, "")
	return self
}

// RoffLinebreak inserts a line break in the Roff document source
//
// This will not show up in the output of the roff program.
func (self *Roff) RoffLinebreak() *Roff {
	self.payload.pushStr(EscapeUnescapedAtNewline, "")
	return self
}

// Escape inserts a raw escape sequence
//
// You can use all the notations for the escapes, they will be copied into the
// output stream as is without extra checks or escapes.
func (self *Roff) Escape(arg string) *Roff {
	self.payload.pushStr(EscapeUnescaped, arg)
	return self
}

// Plaintext inserts a plain text string, special characters are escaped
func (self *Roff) Plaintext(text string) *Roff {
	if self.stripNewlines {
		self.payload.pushStr(EscapeSpecialNoNewline, text)
	} else {
		self.payload.pushStr(EscapeSpecial, text)
	}
	return self
}

// Text inserts one or more string slices using custom font for each one
func (self *Roff) Text(text ...FontText) *Roff {
	var prev_font Font
	var prev_set bool
	for _, item := range text {
		if prev_set && prev_font == item.Font {
			self.Plaintext(item.Text)
		} else {
			self.Escape(item.Font.escape())
			self.Plaintext(item.Text)
			prev_font, prev_set = item.Font, true
		}
	}
	if prev_set {
		self.Escape(RESTORE_FONT)
	}
	return self
}

// FontText is a text slice with the font to render it in
type FontText struct {
	Font Font
	Text string
}

// Render renders Roff document to a string
//
// This method creates a valid ROFF document which can be fed to a ROFF implementation
func (self Roff) Render(ap Apostrophes) string {
	res := make([]byte, 0, self.payload.payloadSize()*2)
	if ap == Handle {
		res = append(res, APOSTROPHE_PREABMLE...)
	}
	escapeSlice(self.payload.payload, self.payload.iter(), &res, ap)
	return string(res)
}
