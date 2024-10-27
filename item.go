package cli

import "fmt"

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

func (i Item) String() string {
	var res string
	i.v(
		func(metavar Doc, anywhere bool) {
			res = fmt.Sprintf(
				"Any{help: %v, metavar: %#v, anywhere: %v}",
				i.help, metavar, anywhere,
			)
		},
		func(metavar Metavar) {
			res = fmt.Sprintf(
				"Positional{help: %v, metavar: %s}",
				i.help, metavar,
			)
		},
		func(name string, short Option[byte], meta Meta, info Info) {
			res = fmt.Sprintf(
				"Command{help: %v, name: %s, short: %#v, meta: %#v, info: %#v}",
				i.help, name, short, meta, info,
			)
		},
		func(name ShortLong, shorts []byte, env Option[string]) {
			res = fmt.Sprintf(
				"Flag{help: %v, name: %#v, shorts: %#v, env: %v}",
				i.help, name, shorts, env,
			)
		},
		func(name ShortLong, shorts []byte, metavar Metavar, env Option[string]) {
			res = fmt.Sprintf(
				"Argument{help: %v, name: %#v, shorts: %#v, metavar: %s, env: %v}",
				i.help, name, shorts, metavar, env,
			)
		},
	)
	return res
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

// Normalize name inside [ShortLong] into either short or long
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

// ShortLong at least one field must be presented
type ShortLong struct {
	Short byte
	Long  string
}

func (sl ShortLong) is_short() bool { return sl.Short != 0 && sl.Long == "" }
func (sl ShortLong) is_long() bool  { return sl.Short == 0 && sl.Long != "" }
func (sl ShortLong) is_both() bool  { return sl.Short != 0 && sl.Long != "" }

func ShortLongShort(b byte) ShortLong          { return ShortLong{Short: b} }
func ShortLongLong(s string) ShortLong         { return ShortLong{Long: s} }
func ShortLongBoth(b byte, s string) ShortLong { return ShortLong{Short: b, Long: s} }

func (self ShortLong) as_long() Option[string] {
	return Option[string]{self.Long, self.Long != ""}
}

func (self ShortLong) as_short() Option[byte] {
	return Option[byte]{self.Short, self.Short != 0}
}

func (self ShortLong) eq(other string) bool {
	short_eq := func(c byte, s string) bool {
		return strip_prefixb(s, '-') == Valid(string([]byte{c}))
	}
	long_eq := func(l string, s string) bool {
		return strip_prefix(s, "--") == Valid(l)
	}
	switch {
	case self.is_short():
		return short_eq(self.Short, other)
	case self.is_long():
		return long_eq(self.Long, other)
	default:
		return short_eq(self.Short, other) || long_eq(self.Long, other)
	}
}

// Changes [ShortLong](ShortLongShortLong) variant into either short or long depending,
// leaves both Short and Long untouched
func (self *ShortLong) normalize(short bool) {
	if !self.is_both() {
		return
	}
	if short {
		*self = ShortLongShort(self.Short)
	} else {
		*self = ShortLongLong(self.Long)
	}
}

func ShortLongtry_from(named NamedArg) (ShortLong, bool) {
	switch {
	case is_empty(named.short) && is_empty(named.long):
		return ShortLong{}, false
	case is_empty(named.short):
		return ShortLongLong(named.long[0]), true
	case is_empty(named.long):
		return ShortLongShort(named.short[0]), true
	default:
		return ShortLongBoth(named.short[0], named.long[0]), true
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
