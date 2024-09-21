package cli

// TODO: move out Option[Doc]
type Item func(
	Any func(
		metavar Doc,
		// used by any, moves it from positionals into arguments
		anywhere bool,
		help Option[Doc],
	),
	// Positional item, consumed from the the front of the arguments
	// <FILE>
	Positional func(metavar Metavar, help Option[Doc]),
	Command func(
		name string,
		short Option[byte],
		help Option[Doc],
		meta Meta,
		info Info,
	),
	// short or long name, consumed anywhere
	// -f
	// --file
	Flag func(
		name ShortLong,
		// used for disambiguation
		shorts []byte,
		env Option[string],
		help Option[Doc],
	),
	// Short or long name followed by a value, consumed anywhere
	// -f <VAL>
	// --file <VAL>
	Argument func(
		name ShortLong,
		// used for disambiguation
		shorts []byte,
		metavar Metavar,
		env Option[string],
		help Option[Doc],
	),
)

func ItemAny(metavar Doc, anywhere bool, help Option[Doc]) Item {
	return func(
		Any func(Doc, bool, Option[Doc]),
		_ func(Metavar, Option[Doc]),
		_ func(string, Option[byte], Option[Doc], Meta, Info),
		_ func(ShortLong, []byte, Option[string], Option[Doc]),
		_ func(ShortLong, []byte, Metavar, Option[string], Option[Doc]),
	) {
		if Any != nil {
			Any(metavar, anywhere, help)
		}
	}
}

func ItemPositional(metavar Metavar, help Option[Doc]) Item {
	return func(
		_ func(Doc, bool, Option[Doc]),
		Positional func(Metavar, Option[Doc]),
		_ func(string, Option[byte], Option[Doc], Meta, Info),
		_ func(ShortLong, []byte, Option[string], Option[Doc]),
		_ func(ShortLong, []byte, Metavar, Option[string], Option[Doc]),
	) {
		if Positional != nil {
			Positional(metavar, help)
		}
	}
}

func (self Item) is_pos() bool {
	var res bool
	self(
		func(_ Doc, anywhere bool, _ Option[Doc]) { res = !anywhere },
		func(Metavar, Option[Doc]) { res = true },
		func(string, Option[byte], Option[Doc], Meta, Info) { res = true },
		nil,
		nil,
	)
	return res
}

// Normalize name inside [`ShortLong`] into either short or long
func (self *Item) normalize(short bool) {
	self(
		nil,
		nil,
		nil,
		func(
			name ShortLong,
			shorts []byte,
			env Option[string],
			help Option[Doc],
		) {
			*self = SelfFlag(name.normalize(short), shorts, env, help)
		},
		func(
			name ShortLong,
			shorts []byte,
			metavar Metavar,
			env Option[string],
			help Option[Doc],
		) {
			*self = SelfArgument(name.normalize(short), shorts, metavar, env, help)
		},
	)
}

type ShortLong func(
	Short func(byte),
	Long func(string),
	Both func(byte, string),
)

func (self ShortLong) as_long() Option[string] {
	var res Option[string]
	self(
		nil,
		func(l string) { res = Valid(l) },
		func(_ byte, l string) { res = Valid(l) },
	)
	return res
}

func (self ShortLong) as_short() Option[char] {
	var res Option[char]
	self(
		func(s byte) { res = Valid(s) },
		nil,
		func(s byte, _ string) { res = Valid(s) },
	)
	return res
}

func (self ShortLong) eq(other string) bool {
	short_eq := func(c char, s string) bool {
		tmp := [4]u8{}
		s.strip_prefix('-') == Valid(c.encode_utf8(&tmp))
	}
	long_eq := func(l string, s string) bool {
		Valid(l) == s.strip_prefix("--")
	}
	var res bool
	self(
		func(s byte) { res = short_eq(*s, other) },
		func(l string) { res = long_eq(l, other) },
		func(s byte, l string) { res = short_eq(*s, other) || long_eq(l, other) },
	)
	return res
}

// Changes [`ShortLong`](ShortLongShortLong) variant into either short or long depending,
// leaves both Short and Long untouched
func (self *ShortLong) normalize(short bool) {
	self(
		nil,
		nil,
		func(s byte, l string) {
			if short {
				*self = SelfShort(*s)
			} else {
				*self = SelfLong(l)
			}
		},
	)
}

func ShortLongFromNamedArg(named NamedArg) Option[ShortLong] {
	switch {
	case named.short.is_empty() && named.long.is_empty():
		return Invalid[ShortLong]()
	case named.short.is_empty():
		return Valid(SelfLong(named.long[0]))
	case named.long.is_empty():
		return Valid(SelfShort(named.short[0]))
	default:
		return Valid(SelfBoth(named.short[0], named.long[0]))
	}
}

func (self Item) required(required bool) Meta {
	boxed := Metafrom(self)
	if required {
		return boxed
	} else {
		return MetaOptional(Boxnew(boxed))
	}
}
