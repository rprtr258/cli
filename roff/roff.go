// Package roff renders documents in the ROFF text-formatting language, as
// used for manual pages.
//
// [ROFF] is a family of Unix text-formatting languages, implemented by the
// nroff, troff, and groff programs, among others. A [Roff] value is an
// abstract representation of a document in ROFF format. It is meant for
// writing code that generates ROFF documents, such as manual pages.
//
// [ROFF]: https://en.wikipedia.org/wiki/Roff_(software)
package roff

// Apostrophes controls how apostrophes are rendered.
//
// To generate manpages you most likely want this in the [Handle] state.
// See https://bugs.debian.org/cgi-bin/bugreport.cgi?bug=507673#65 for details.
type Apostrophes uint8

const (
	// Handle replaces apostrophes with a portable escape sequence.
	Handle Apostrophes = iota
	// DontHandle leaves apostrophes as is.
	DontHandle
)

// apostrophe is the replacement written for `'` when apostrophe handling is
// enabled.
//
// See https://bugs.debian.org/cgi-bin/bugreport.cgi?bug=507673#65.
const apostrophe = `\*(Aq`

// APOSTROPHE_PREAMBLE is the preamble added to the start of rendered output
// when apostrophe handling is enabled.
//
// It defines a string variable that contains an apostrophe. For historical
// reasons, there seems to be no other portable way to represent apostrophes
// across various implementations of the ROFF language. In implementations
// that produce output like PostScript or PDF, an apostrophe gets typeset as
// a right single quote, which looks different from an apostrophe. For
// terminal output ("ASCII"), such as when using nroff, an apostrophe looks
// indistinguishable from a right single quote. For manual pages and similar
// content, an apostrophe is more generally desired than the right single
// quote, so all apostrophe characters in input text are converted into uses
// of the string variable defined in this preamble.
//
// See: https://bugs.debian.org/cgi-bin/bugreport.cgi?bug=507673#65
const APOSTROPHE_PREAMBLE = `.ie \n(.g .ds Aq \(aq
.el .ds Aq '
`

// Escape selects the escaping rules applied to a piece of text.
type Escape uint8

const (
	// EscapeUnescapedAtNewline inserts a newline character unless the output
	// is already at the start of a line.
	EscapeUnescapedAtNewline Escape = iota

	// EscapeSpaces escapes space characters (' ') and newlines ('\n').
	// It is used for control sequence arguments.
	EscapeSpaces

	// EscapeSpecial escapes characters roff considers special:
	//   - ' ', '.' and '\'' at the beginning of the line,
	//   - '-' and '\\' when inside the body,
	//   - '\'' is replaced with [Handle]'s escape when apostrophe handling
	//     is enabled.
	EscapeSpecial

	// EscapeSpecialNoNewline is similar to [EscapeSpecial] but also replaces
	// '\n' with ' '.
	EscapeSpecialNoNewline

	// EscapeUnescaped writes the input as is.
	EscapeUnescaped
)

// escapeSlice escapes payload fragments into out according to the escape
// kind of each label. The result is valid UTF-8 as long as out starts empty
// or contains a valid UTF-8 sequence.
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

// freeMonoid is a payload string annotated with escape labels.
type freeMonoid struct {
	payload string
	labels  []label
}

func (self *freeMonoid) pushStr(kind Escape, payload string) {
	start := len(self.payload)
	self.payload += payload
	// No squashing: control lines and font runs are kept separate.
	self.labels = append(self.labels, label{start, len(self.payload), kind})
}

func (self *freeMonoid) payloadSize() int {
	return len(self.payload)
}

func (self *freeMonoid) iter() []label {
	return self.labels
}

// Font selects the font used to render a piece of text.
type Font uint8

const (
	// Roman is the regular font that should be used for most of the text.
	Roman Font = iota

	// Bold font.
	Bold

	// Italic font.
	Italic
)

// escape returns the escape code that switches to this font.
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

// RESTORE_FONT is the escape code that returns to the previous font.
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

// StripNewlines controls whether [Roff.Render] keeps newline characters
// inside added text.
//
// Newlines can carry a special meaning in ROFF. For example, newlines are
// automatically stripped from [Roff.Control] arguments when adding a
// section header.
func (self *Roff) StripNewlines(state bool) *Roff {
	self.stripNewlines = state
	return self
}

// Control inserts a raw control sequence.
//
// name should not contain the initial '.'. Arguments are escaped as space
// escaped payloads.
func (self *Roff) Control(name string, args ...string) *Roff {
	self.payload.pushStr(EscapeUnescapedAtNewline, ".")
	self.payload.pushStr(EscapeUnescaped, name)
	for _, arg := range args {
		// An empty macro argument is written as "", which is mostly useful
		// for the TH macro that takes several positional arguments.
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

// Control0 is a variant of [Roff.Control] that takes no arguments.
func (self *Roff) Control0(name string) *Roff {
	self.payload.pushStr(EscapeUnescapedAtNewline, ".")
	self.payload.pushStr(EscapeUnescaped, name)
	self.payload.pushStr(EscapeUnescapedAtNewline, "")
	return self
}

// RoffLinebreak inserts a line break in the ROFF document source.
//
// The break does not show up in the output of the roff program.
func (self *Roff) RoffLinebreak() *Roff {
	self.payload.pushStr(EscapeUnescapedAtNewline, "")
	return self
}

// Escape inserts a raw escape sequence.
//
// Any escape notation is copied to the output stream as is, without extra
// checks or escaping.
func (self *Roff) Escape(arg string) *Roff {
	self.payload.pushStr(EscapeUnescaped, arg)
	return self
}

// Plaintext inserts a plain text string, escaping special characters.
func (self *Roff) Plaintext(text string) *Roff {
	if self.stripNewlines {
		self.payload.pushStr(EscapeSpecialNoNewline, text)
	} else {
		self.payload.pushStr(EscapeSpecial, text)
	}
	return self
}

// Text inserts one or more text slices, each rendered with its own font.
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

// FontText is a text slice with the font to render it in.
type FontText struct {
	Font Font
	Text string
}

// Render renders the ROFF document to a string.
//
// The result is a valid ROFF document that can be fed to a ROFF
// implementation.
func (self Roff) Render(ap Apostrophes) string {
	res := make([]byte, 0, self.payload.payloadSize()*2)
	if ap == Handle {
		res = append(res, APOSTROPHE_PREAMBLE...)
	}
	escapeSlice(self.payload.payload, self.payload.iter(), &res, ap)
	return string(res)
}
