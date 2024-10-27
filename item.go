package cli

type Item struct {
	help Option[Doc]
	v    func(
		Any func(
			metavar Doc,
			// used by any, moves it from positionals into arguments
			anywhere bool,
		),
		// Positional item, consumed from the the front of the arguments
		// <FILE>
		Positional func(metavar Metavar),
		Command func(
			name string,
			short Option[byte],
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
		),
	)
}

func ItemAny(metavar Doc, anywhere bool, help Option[Doc]) Item {
	return Item{help, func(
		f func(Doc, bool),
		_ func(Metavar),
		_ func(string, Option[byte], Meta, Info),
		_ func(ShortLong, []byte, Option[string]),
		_ func(ShortLong, []byte, Metavar, Option[string]),
	) {
		if f != nil {
			f(metavar, anywhere)
		}
	}}
}

func ItemPositional(metavar Metavar, help Option[Doc]) Item {
	return Item{help, func(
		_ func(Doc, bool),
		f func(Metavar),
		_ func(string, Option[byte], Meta, Info),
		_ func(ShortLong, []byte, Option[string]),
		_ func(ShortLong, []byte, Metavar, Option[string]),
	) {
		if f != nil {
			f(metavar)
		}
	}}
}

func ItemCommand(name string, short Option[byte], help Option[Doc], meta Meta, info Info) Item {
	return Item{help, func(
		_ func(Doc, bool),
		_ func(Metavar),
		f func(string, Option[byte], Meta, Info),
		_ func(ShortLong, []byte, Option[string]),
		_ func(ShortLong, []byte, Metavar, Option[string]),
	) {
		if f != nil {
			f(name, short, meta, info)
		}
	}}
}

func ItemFlag(name ShortLong, shorts []byte, env Option[string], help Option[Doc]) Item {
	return Item{help, func(
		_ func(Doc, bool),
		_ func(Metavar),
		_ func(string, Option[byte], Meta, Info),
		f func(ShortLong, []byte, Option[string]),
		_ func(ShortLong, []byte, Metavar, Option[string]),
	) {
		if f != nil {
			f(name, shorts, env)
		}
	}}
}

func ItemArgument(name ShortLong, shorts []byte, metavar Metavar, env Option[string], help Option[Doc]) Item {
	return Item{help, func(
		_ func(Doc, bool),
		_ func(Metavar),
		_ func(string, Option[byte], Meta, Info),
		_ func(ShortLong, []byte, Option[string]),
		f func(ShortLong, []byte, Metavar, Option[string]),
	) {
		if f != nil {
			f(name, shorts, metavar, env)
		}
	}}
}

func (self Item) is_pos() bool {
	var res bool
	self.v(
		func(_ Doc, anywhere bool) { res = !anywhere },
		func(Metavar) { res = true },
		func(string, Option[byte], Meta, Info) { res = true },
		nil,
		nil,
	)
	return res
}

// Normalize name inside [`ShortLong`] into either short or long
func (self *Item) normalize(short bool) {
	self.v(
		nil,
		nil,
		nil,
		func(
			name ShortLong,
			shorts []byte,
			env Option[string],
		) {
			name.normalize(short)
			*self = ItemFlag(name, shorts, env, self.help)
		},
		func(
			name ShortLong,
			shorts []byte,
			metavar Metavar,
			env Option[string],
		) {
			name.normalize(short)
			*self = ItemArgument(name, shorts, metavar, env, self.help)
		},
	)
}

type ShortLong func(
	Short func(byte),
	Long func(string),
	Both func(byte, string),
)

func ShortLongShort(s byte) ShortLong {
	return func(
		f func(byte),
		_ func(string),
		_ func(byte, string),
	) {
		if f != nil {
			f(s)
		}
	}
}

func ShortLongLong(s string) ShortLong {
	return func(
		_ func(byte),
		f func(string),
		_ func(byte, string),
	) {
		if f != nil {
			f(s)
		}
	}
}

func ShortLongBoth(b byte, s string) ShortLong {
	return func(
		_ func(byte),
		_ func(string),
		f func(byte, string),
	) {
		if f != nil {
			f(b, s)
		}
	}
}

func (self ShortLong) as_long() Option[string] {
	var res Option[string]
	self(
		nil,
		func(l string) { res = Valid(l) },
		func(_ byte, l string) { res = Valid(l) },
	)
	return res
}

func (self ShortLong) as_short() Option[byte] {
	var res Option[byte]
	self(
		func(s byte) { res = Valid(s) },
		nil,
		func(s byte, _ string) { res = Valid(s) },
	)
	return res
}

func (self ShortLong) eq(other string) bool {
	short_eq := func(c byte, s string) bool {
		return strip_prefixb(s, '-') == Valid(string([]byte{c}))
	}
	long_eq := func(l string, s string) bool {
		return strip_prefix(s, "--") == Valid(l)
	}
	var res bool
	self(
		func(s byte) { res = short_eq(s, other) },
		func(l string) { res = long_eq(l, other) },
		func(s byte, l string) { res = short_eq(s, other) || long_eq(l, other) },
	)
	return res
}

// Changes [`ShortLong`](ShortLongShortLong) variant into either short or long depending,
// leaves both Short and Long untouched
func (self *ShortLong) normalize(short bool) {
	(*self)(
		nil,
		nil,
		func(s byte, l string) {
			if short {
				*self = ShortLongShort(s)
			} else {
				*self = ShortLongLong(l)
			}
		},
	)
}

func ShortLongtry_from(named NamedArg) Option[ShortLong] {
	switch {
	case is_empty(named.short) && is_empty(named.long):
		return Invalid[ShortLong]()
	case is_empty(named.short):
		return Valid(ShortLongLong(named.long[0]))
	case is_empty(named.long):
		return Valid(ShortLongShort(named.short[0]))
	default:
		return Valid(ShortLongBoth(named.short[0], named.long[0]))
	}
}

func (self Item) required(required bool) Meta {
	boxed := MetaItem(self)
	if required {
		return boxed
	} else {
		return MetaOptional(boxed)
	}
}
