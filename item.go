package cli

import "fmt"

type item interface {
	isItem()
	clone() item
}

type ItemAny struct {
	metavar Doc
	// used by any, moves it from positionals into arguments
	anywhere bool
}

func (ItemAny) isItem()          {}
func (self ItemAny) clone() item { return ItemAny{self.metavar.clone(), self.anywhere} }

// Positional item, consumed from the the front of the arguments
// <FILE>
type ItemPositional struct{ metavar Metavar }

func (ItemPositional) isItem()          {}
func (self ItemPositional) clone() item { return ItemPositional{self.metavar} }

type ItemCommand struct {
	name  string
	short Option[byte]
	meta  Meta
	info  Info
}

func (ItemCommand) isItem() {}
func (self ItemCommand) clone() item {
	return ItemCommand{
		self.name,
		self.short,
		self.meta.clone(),
		self.info.clone(),
	}
}

// short or long name, consumed anywhere
// -f
// --file
type ItemFlag struct {
	name ShortLong
	// used for disambiguation
	shorts []byte
	env    Option[string]
}

func (ItemFlag) isItem() {}
func (self ItemFlag) clone() item {
	return ItemFlag{
		self.name,
		append([]byte(nil), self.shorts...),
		self.env,
	}
}

// Short or long name followed by a value, consumed anywhere
// -f <VAL>
// --file <VAL>
type ItemArgument struct {
	name ShortLong
	// used for disambiguation
	shorts  []byte
	metavar Metavar
	env     Option[string]
}

func (ItemArgument) isItem() {}
func (self ItemArgument) clone() item {
	return ItemArgument{
		self.name,
		append([]byte(nil), self.shorts...),
		self.metavar,
		self.env,
	}
}

type Item struct {
	help Option[Doc]
	v    item
}

func (i Item) clone() Item {
	return Item{
		Option[Doc]{
			i.help.Value.clone(),
			i.help.Valid,
		},
		i.v.clone(),
	}
}

func (i Item) String() string {
	switch v := i.v.(type) {
	case ItemAny:
		return fmt.Sprintf(
			"Any{help: %v, metavar: %#v, anywhere: %v}",
			i.help, v.metavar, v.anywhere,
		)
	case ItemPositional:
		return fmt.Sprintf(
			"Positional{help: %v, metavar: %s}",
			i.help, v.metavar,
		)
	case ItemCommand:
		return fmt.Sprintf(
			"Command{help: %v, name: %s, short: %#v, meta: %#v, info: %#v}",
			i.help, v.name, v.short, v.meta, v.info,
		)
	case ItemFlag:
		return fmt.Sprintf(
			"Flag{help: %v, name: %#v, shorts: %#v, env: %v}",
			i.help, v.name, v.shorts, v.env,
		)
	case ItemArgument:
		return fmt.Sprintf(
			"Argument{help: %v, name: %#v, shorts: %#v, metavar: %s, env: %v}",
			i.help, v.name, v.shorts, v.metavar, v.env,
		)
	}
	panic("unreachable")
}

func (self Item) is_pos() bool {
	switch v := self.v.(type) {
	case ItemAny:
		return !v.anywhere
	case ItemPositional:
		return true
	case ItemCommand:
		return true
	}
	return false
}

// normalize name inside [ShortLong] into either short or long
func (self *Item) normalize(short bool) {
	switch v := self.v.(type) {
	case ItemFlag:
		v.name.normalize(short)
		*self = Item{self.help, ItemFlag{v.name, v.shorts, v.env}}
	case ItemArgument:
		v.name.normalize(short)
		*self = Item{self.help, ItemArgument{v.name, v.shorts, v.metavar, v.env}}
	}
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

func (self ShortLong) as_long() (string, bool) {
	return self.Long, self.Long != ""
}

func (self ShortLong) as_short() (byte, bool) {
	return self.Short, self.Short != 0
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
	case len(named.short) == 0 && len(named.long) == 0:
		return ShortLong{}, false
	case len(named.short) == 0:
		return ShortLongLong(named.long[0]), true
	case len(named.long) == 0:
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
		return MetaOptional{boxed}
	}
}
