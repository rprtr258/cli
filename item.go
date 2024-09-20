package cli

import "github.com/rprtr258/fun"

// TODO: move out fun.Option[Doc]
type Item func(
	Any func(
		metavar Doc,
		// used by any, moves it from positionals into arguments
		anywhere bool,
		help fun.Option[Doc],
	),
	// Positional item, consumed from the the front of the arguments
	// <FILE>
	Positional func(metavar Metavar, help fun.Option[Doc]),
	Command func(
		name string,
		short fun.Option[byte],
		help fun.Option[Doc],
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
		env fun.Option[string],
		help fun.Option[Doc],
	),
	// Short or long name followed by a value, consumed anywhere
	// -f <VAL>
	// --file <VAL>
	Argument func(
		name ShortLong,
		// used for disambiguation
		shorts []byte,
		metavar Metavar,
		env fun.Option[string],
		help fun.Option[Doc],
	),
)

func (self Item) is_pos() bool {
	var res bool
	self(
		func(_ Doc, anywhere bool, _ fun.Option[Doc]) { res = !anywhere },
		func(Metavar, fun.Option[Doc]) { res = true },
		func(string, fun.Option[byte], fun.Option[Doc], Meta, Info) { res = true },
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
			env fun.Option[string],
			help fun.Option[Doc],
		) {
			*self = SelfFlag(name.normalize(short), shorts, env, help)
		},
		func(
			name ShortLong,
			shorts []byte,
			metavar Metavar,
			env fun.Option[string],
			help fun.Option[Doc],
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
	var res fun.Option[string]
	self(
		nil,
		func(l string) { res = fun.Valid(l) },
		func(_ byte, l string) { res = fun.Valid(l) },
	)
	return res
}

func (self ShortLong) as_short() Option[char] {
	var res fun.Option[char]
	self(
		func(s byte) { res = fun.Valid(s) },
		nil,
		func(s byte, _ string) { res = fun.Valid(s) },
	)
	return res
}

func (self ShortLong) eq(other string) bool {
	short_eq := func(c char, s string) bool {
		tmp := [4]u8{}
		s.strip_prefix('-') == fun.Valid(c.encode_utf8(&tmp))
	}
	long_eq := func(l string, s string) bool {
		fun.Valid(l) == s.strip_prefix("--")
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

func ShortLongFromNamedArg(named NamedArg) fun.Option[ShortLong] {
	switch {
	case named.short.is_empty() && named.long.is_empty():
		return fun.Invalid[ShortLong]()
	case named.short.is_empty():
		return fun.Valid(SelfLong(named.long[0]))
	case named.long.is_empty():
		return fun.Valid(SelfShort(named.short[0]))
	default:
		return fun.Valid(SelfBoth(named.short[0], named.long[0]))
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
