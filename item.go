package cli

import "fmt"

type itemI interface {
	isItem()
	clone() itemI
}

type itemAny struct {
	metavar Doc
	// used by any, moves it from positionals into arguments
	anywhere bool
}

func (itemAny) isItem()           {}
func (self itemAny) clone() itemI { return itemAny{self.metavar.clone(), self.anywhere} }

// Positional item, consumed from the the front of the arguments
// <FILE>
type itemPositional struct{ metavar Metavar }

func (itemPositional) isItem()           {}
func (self itemPositional) clone() itemI { return itemPositional{self.metavar} }

type itemCommand struct {
	name  string
	short Opt[byte]
	meta  meta
	info  info
}

func (itemCommand) isItem() {}
func (self itemCommand) clone() itemI {
	return itemCommand{
		self.name,
		self.short,
		self.meta.clone(),
		self.info.clone(),
	}
}

// short or long name, consumed anywhere
// -f
// --file
type itemFlag struct {
	name shortLong
	// used for disambiguation
	shorts []byte
	env    Opt[string]
}

func (itemFlag) isItem() {}
func (self itemFlag) clone() itemI {
	return itemFlag{
		self.name,
		append([]byte(nil), self.shorts...),
		self.env,
	}
}

// Short or long name followed by a value, consumed anywhere
// -f <VAL>
// --file <VAL>
type itemArgument struct {
	name shortLong
	// used for disambiguation
	shorts  []byte
	metavar Metavar
	env     Opt[string]
}

func (itemArgument) isItem() {}
func (self itemArgument) clone() itemI {
	return itemArgument{
		self.name,
		append([]byte(nil), self.shorts...),
		self.metavar,
		self.env,
	}
}

type item struct {
	help Opt[Doc]
	v    itemI
}

func (i item) clone() item {
	return item{
		Opt[Doc]{
			i.help.Value.clone(),
			i.help.Valid,
		},
		i.v.clone(),
	}
}

func (i item) String() string {
	switch v := i.v.(type) {
	case itemAny:
		return fmt.Sprintf(
			"Any{help: %v, metavar: %#v, anywhere: %v}",
			i.help, v.metavar, v.anywhere,
		)
	case itemPositional:
		return fmt.Sprintf(
			"Positional{help: %v, metavar: %s}",
			i.help, v.metavar,
		)
	case itemCommand:
		return fmt.Sprintf(
			"Command{help: %v, name: %s, short: %#v, meta: %#v, info: %#v}",
			i.help, v.name, v.short, v.meta, v.info,
		)
	case itemFlag:
		return fmt.Sprintf(
			"Flag{help: %v, name: %#v, shorts: %#v, env: %v}",
			i.help, v.name, v.shorts, v.env,
		)
	case itemArgument:
		return fmt.Sprintf(
			"Argument{help: %v, name: %#v, shorts: %#v, metavar: %s, env: %v}",
			i.help, v.name, v.shorts, v.metavar, v.env,
		)
	}
	panic("unreachable")
}

func (self item) is_pos() bool {
	switch v := self.v.(type) {
	case itemAny:
		return !v.anywhere
	case itemPositional:
		return true
	case itemCommand:
		return true
	}
	return false
}

// normalize name inside [shortLong] into either short or long
func (self *item) normalize(short bool) {
	switch v := self.v.(type) {
	case itemFlag:
		v.name.normalize(short)
		*self = item{self.help, itemFlag{v.name, v.shorts, v.env}}
	case itemArgument:
		v.name.normalize(short)
		*self = item{self.help, itemArgument{v.name, v.shorts, v.metavar, v.env}}
	}
}

// shortLong at least one field must be presented
type shortLong struct {
	Short byte
	Long  string
}

func (sl shortLong) is_short() bool { return sl.Short != 0 && sl.Long == "" }
func (sl shortLong) is_long() bool  { return sl.Short == 0 && sl.Long != "" }
func (sl shortLong) is_both() bool  { return sl.Short != 0 && sl.Long != "" }

func shortLongShort(b byte) shortLong          { return shortLong{Short: b} }
func shortLongLong(s string) shortLong         { return shortLong{Long: s} }
func shortLongBoth(b byte, s string) shortLong { return shortLong{Short: b, Long: s} }

func (self shortLong) as_long() (string, bool) {
	return self.Long, self.Long != ""
}

func (self shortLong) as_short() (byte, bool) {
	return self.Short, self.Short != 0
}

func (self shortLong) eq(other string) bool {
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

// Changes [shortLong](shortLongShortLong) variant into either short or long depending,
// leaves both Short and Long untouched
func (self *shortLong) normalize(short bool) {
	if !self.is_both() {
		return
	}
	if short {
		*self = shortLongShort(self.Short)
	} else {
		*self = shortLongLong(self.Long)
	}
}

func shortLongTryFrom(named NamedArg) (shortLong, bool) {
	switch {
	case len(named.short) == 0 && len(named.long) == 0:
		return shortLong{}, false
	case len(named.short) == 0:
		return shortLongLong(named.long[0]), true
	case len(named.long) == 0:
		return shortLongShort(named.short[0]), true
	default:
		return shortLongBoth(named.short[0], named.long[0]), true
	}
}

func (self item) required(required bool) meta {
	boxed := metaItem(self)
	if required {
		return boxed
	} else {
		return metaOptional{boxed}
	}
}
