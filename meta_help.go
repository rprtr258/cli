package cli

import (
	"fmt"
	"iter"
	"os"
	"unicode"

	"github.com/rprtr258/fun"
	"github.com/rprtr258/fun/set"
)

type Metavar string

type HelpItem interface{ isHelpItem() }
type HelpItemDecorSuffix struct {
	help Doc
	ty   HiTy
}

func (HelpItemDecorSuffix) isHelpItem() {}

type HelpItemGroupStart struct {
	help Doc
	ty   HiTy
}

func (HelpItemGroupStart) isHelpItem() {}

type HelpItemGroupEnd struct{ ty HiTy }

func (HelpItemGroupEnd) isHelpItem() {}

type HelpItemAny struct {
	metavar  Doc
	anywhere bool
	help     Option[Doc]
}

func (HelpItemAny) isHelpItem() {}

type HelpItemPositional struct {
	metavar Metavar
	help    Option[Doc]
}

func (HelpItemPositional) isHelpItem() {}

type HelpItemCommand struct {
	name  string
	short Option[byte]
	help  Option[Doc]
	meta  Meta
	info  Info
}

func (HelpItemCommand) isHelpItem() {}

type HelpItemFlag struct {
	name ShortLong
	env  Option[string]
	help Option[Doc]
}

func (HelpItemFlag) isHelpItem() {}

type HelpItemArgument struct {
	name    ShortLong
	metavar Metavar
	env     Option[string]
	help    Option[Doc]
}

func (HelpItemArgument) isHelpItem() {}

type HelpItemAnywhereStart struct {
	inner Meta
	ty    HiTy
}

func (HelpItemAnywhereStart) isHelpItem() {}

type HelpItemAnywhereStop struct{ ty HiTy }

func (HelpItemAnywhereStop) isHelpItem() {}

func has_help(self HelpItem) bool {
	switch self := self.(type) {
	case HelpItemDecorSuffix, HelpItemGroupStart:
		return true
	case HelpItemAny:
		return self.help.Valid
	case HelpItemPositional:
		return self.help.Valid
	case HelpItemCommand:
		return self.help.Valid
	case HelpItemFlag:
		return self.help.Valid
	case HelpItemArgument:
		return self.help.Valid
	}
	panic("unreachable")
}

func ty(self HelpItem) HiTy {
	switch self := self.(type) {
	case HelpItemDecorSuffix:
		return self.ty
	case HelpItemGroupStart:
		return self.ty
	case HelpItemGroupEnd:
		return self.ty
	case HelpItemAny:
		return fun.IF(self.anywhere, HiTyFlag, HiTyPositional)
	case HelpItemPositional:
		return HiTyPositional
	case HelpItemCommand:
		return HiTyCommand
	case HelpItemFlag:
		return HiTyFlag
	case HelpItemArgument:
		return HiTyFlag
	case HelpItemAnywhereStart:
		return self.ty
	case HelpItemAnywhereStop:
		return self.ty
	}
	panic("unreachable")
}

// A collection of all the help items separated into flags, positionals and commands
//
// Items are stored as references and can be trivially copied
type HelpItems []HelpItem

func HelpItemsdefault() HelpItems {
	return []HelpItem{}
}

type HiTy uint8

const (
	HiTyFlag HiTy = iota
	HiTyCommand
	HiTyPositional
)

// No | Decor(HiTy) | Anywhere(HiTy)
type ItemBlock struct {
	present bool
	isDecor bool // anywhere otherwise
	ty      HiTy
}

var ItemBlockNo = ItemBlock{}

func ItemBlockDecor(ty HiTy) ItemBlock {
	return ItemBlock{true, true, ty}
}

func ItemBlockAnywhere(ty HiTy) ItemBlock {
	return ItemBlock{true, false, ty}
}

type HelpItemsIter struct {
	items  []HelpItem
	target HiTy
	cur    int
	block  ItemBlock
}

func (self *HelpItemsIter) next() (HelpItem, bool) {
	for {
		if self.cur >= len(self.items) {
			return nil, false
		}
		item := self.items[self.cur]
		self.cur += 1

		var keep bool
		switch it := item.(type) {
		case HelpItemDecorSuffix, HelpItemAny, HelpItemPositional, HelpItemCommand, HelpItemFlag, HelpItemArgument:
			ty := ty(item)
			switch {
			case !self.block.present:
				keep = ty == self.target
			case self.block.isDecor:
				keep = self.block.ty == self.target
			default:
				keep = self.block.ty == self.target && has_help(item)
			}
		case HelpItemGroupStart:
			self.block = ItemBlockDecor(it.ty)
			keep = it.ty == self.target
		case HelpItemGroupEnd:
			self.block = ItemBlockNo
			keep = it.ty == self.target
		case HelpItemAnywhereStart:
			self.block = ItemBlockAnywhere(it.ty)
			keep = it.ty == self.target
		case HelpItemAnywhereStop:
			self.block = ItemBlockNo
			keep = it.ty == self.target
		}
		if keep {
			return item, true
		}
	}
}

func (self HelpItems) items_of_ty(target HiTy) iter.Seq[HelpItem] {
	iter := HelpItemsIter{
		items:  self,
		target: target,
		cur:    0,
		block:  ItemBlockNo,
	}
	return func(yield func(HelpItem) bool) {
		for {
			x, ok := iter.next()
			if !ok || !yield(x) {
				break
			}
		}
	}
}

func peek_front_ty(self Meta) (HiTy, bool) {
	var (
		res HiTy
		ok  bool
	)
	switch m := self.(type) {
	case MetaAnd:
		res, ok = find_map(m, peek_front_ty)
	case MetaOr:
		res, ok = find_map(m, peek_front_ty)
	case MetaOptional:
		res, ok = peek_front_ty(m.m)
	case MetaRequired:
		res, ok = peek_front_ty(m.m)
	case MetaAdjacent:
		res, ok = peek_front_ty(m.m)
	case MetaItem:
		res, ok = HiTyFromItem(Item(m)), true
	case MetaMany:
		res, ok = peek_front_ty(m.m)
	case MetaSubsection:
		res, ok = peek_front_ty(m.m)
	case MetaSuffix:
		res, ok = peek_front_ty(m.m)
	case MetaCustomUsage:
		res, ok = peek_front_ty(m.m)
	}
	return res, ok
}

// Recursively classify contents of the Meta
func (self *HelpItems) append_meta(meta Meta) {
	var GO func(meta Meta, no_ss bool)
	GO = func(meta Meta, no_ss bool) {
		switch meta := meta.(type) {
		case MetaAnd:
			for _, x := range meta {
				GO(x, no_ss)
			}
		case MetaOr:
			for _, x := range meta {
				GO(x, no_ss)
			}
		case MetaAdjacent:
			if ty, ok := peek_front_ty(meta.m); ok {
				*self = append(*self, HelpItemAnywhereStart{meta.m, ty})
				GO(meta.m, no_ss)
				*self = append(*self, HelpItemAnywhereStop{ty})
			}
		case MetaStrict:
			GO(meta.m, no_ss)
		case MetaCustomUsage:
			GO(meta.m, no_ss)
		case MetaMany:
			GO(meta.m, no_ss)
		case MetaRequired:
			GO(meta.m, no_ss)
		case MetaOptional:
			GO(meta.m, no_ss)
		case MetaItem:
			if _, ok := meta.v.(ItemPositional); ok && !meta.help.Valid {
				return
			}
			*self = append(*self, HelpItemFromItem(Item(meta)))
		case MetaSubsection:
			m := meta.m
			help := meta.u
			if ty, ok := peek_front_ty(m); ok {
				if no_ss {
					GO(m, true)
				} else {
					*self = append(*self, HelpItemGroupStart{help, ty})
					GO(m, true)
					*self = append(*self, HelpItemGroupEnd{ty})
				}
			}
		case MetaSuffix:
			m := meta.m
			help := meta.u
			if ty, ok := peek_front_ty(m); ok {
				GO(m, no_ss)
				*self = append(*self, HelpItemDecorSuffix{help, ty})
			}
		case MetaSkip:
		}
	}
	GO(meta, false)
}
func (self HelpItems) find_group() (Range, bool) {
	start, end := -1, -1
	for i, x := range self {
		var isStart, isEnd bool
		switch x.(type) {
		case HelpItemGroupStart:
			isStart = true
		case HelpItemGroupEnd:
			isEnd = true
		}
		if start == -1 && isStart {
			start = i
		}
		if end == -1 && isEnd {
			end = i
		}
	}
	if start == -1 || end == -1 {
		return Range{}, false
	}
	return Range{start, end + 1}, true
}

func HiTyFromItem(value Item) HiTy {
	switch v := value.v.(type) {
	case ItemAny:
		return fun.IF(v.anywhere, HiTyFlag, HiTyPositional)
	case ItemPositional:
		return HiTyPositional
	case ItemCommand:
		return HiTyCommand
	case ItemFlag:
		return HiTyFlag
	case ItemArgument:
		return HiTyFlag
	}
	panic("unreachable")
}

func HelpItemFromItem(item Item) HelpItem {
	switch v := item.v.(type) {
	case ItemAny:
		return HelpItemAny{v.metavar, v.anywhere, item.help}
	case ItemPositional:
		return HelpItemPositional{v.metavar, item.help}
	case ItemCommand:
		return HelpItemCommand{v.name, v.short, item.help, v.meta, v.info}
	case ItemFlag:
		return HelpItemFlag{v.name, v.env, item.help}
	case ItemArgument:
		return HelpItemArgument{v.name, v.metavar, v.env, item.help}
	}
	panic("unreachable")
}

func (self *Doc) metavar(metavar Metavar) {
	if fun.All(func(c rune) bool { return unicode.IsUpper(c) || unicode.IsDigit(c) || c == '-' || c == '_' }, []rune(metavar)...) {
		self.write_str(string(metavar), StyleMetavar)
	} else {
		self.write_char('<', StyleMetavar)
		self.write_str(string(metavar), StyleMetavar)
		self.write_char('>', StyleMetavar)
	}
}

func write_help_item(buf *Doc, item HelpItem, include_env bool) {
	switch item := item.(type) {
	case HelpItemDecorSuffix:
		buf.token(TokenBlockStart(BlockItemTerm))
		buf.token(TokenBlockEnd(BlockItemTerm))
		buf.token(TokenBlockStart(BlockItemBody))
		buf.Doc(item.help)
		buf.token(TokenBlockEnd(BlockItemBody))
	case HelpItemGroupStart:
		buf.token(TokenBlockStart(BlockBlock))
		buf.token(TokenBlockStart(BlockSection2))
		buf.EmDoc(item.help)
		buf.token(TokenBlockEnd(BlockSection2))
		buf.token(TokenBlockStart(BlockDefinitionList))
	case HelpItemGroupEnd:
		buf.token(TokenBlockEnd(BlockDefinitionList))
		buf.token(TokenBlockEnd(BlockBlock))
	case HelpItemAny:
		buf.token(TokenBlockStart(BlockItemTerm))
		buf.Doc(item.metavar)
		buf.token(TokenBlockEnd(BlockItemTerm))
		if help, ok := item.help.Unpack(); ok {
			buf.token(TokenBlockStart(BlockItemBody))
			buf.Doc(help)
			buf.token(TokenBlockEnd(BlockItemBody))
		}
	case HelpItemPositional:
		buf.token(TokenBlockStart(BlockItemTerm))
		buf.metavar(item.metavar)
		buf.token(TokenBlockEnd(BlockItemTerm))
		if help, ok := item.help.Unpack(); ok {
			buf.token(TokenBlockStart(BlockItemBody))
			buf.Doc(help)
			buf.token(TokenBlockEnd(BlockItemBody))
		}
	case HelpItemCommand:
		buf.token(TokenBlockStart(BlockItemTerm))
		buf.write_str(item.name, StyleLiteral)
		if short, ok := item.short.Unpack(); ok {
			buf.write_str(", ", StyleText)
			buf.write_char(short, StyleLiteral)
		}
		buf.token(TokenBlockEnd(BlockItemTerm))
		if help, ok := item.help.Unpack(); ok {
			buf.token(TokenBlockStart(BlockItemBody))
			buf.Doc(help)
			buf.token(TokenBlockEnd(BlockItemBody))
		}
	case HelpItemFlag:
		buf.token(TokenBlockStart(BlockItemTerm))
		write_shortlong(buf, item.name)
		buf.token(TokenBlockEnd(BlockItemTerm))
		if help, ok := item.help.Unpack(); ok {
			buf.token(TokenBlockStart(BlockItemBody))
			buf.Doc(help)
			buf.token(TokenBlockEnd(BlockItemBody))
		}
		if env, ok := item.env.Unpack(); ok {
			_, found := os.LookupEnv(env)
			val := fun.IF(found, ": set", ": not set")
			if item.help.Valid {
				buf.token(TokenBlockStart(BlockItemTerm))
				buf.token(TokenBlockEnd(BlockItemTerm))
			}
			buf.token(TokenBlockStart(BlockItemBody))
			if include_env {
				buf.write_str(fmt.Sprintf("[env:%v%v]", env, val), StyleText)
			} else {
				buf.Text("Uses environment variable ")
				buf.Literal(env)
			}
			buf.token(TokenBlockEnd(BlockItemBody))
		}
	case HelpItemArgument:
		buf.token(TokenBlockStart(BlockItemTerm))
		write_shortlong(buf, item.name)
		buf.write_str("=", StyleText)
		buf.metavar(item.metavar)
		buf.token(TokenBlockEnd(BlockItemTerm))

		if help, ok := item.help.Unpack(); ok {
			buf.token(TokenBlockStart(BlockItemBody))
			buf.Doc(help)
			buf.token(TokenBlockEnd(BlockItemBody))
		}

		if env, ok := item.env.Unpack(); ok {
			val := ": N/A"
			if s, ok := os.LookupEnv(env); ok {
				val = fmt.Sprintf(" = %#v", s)
			}

			if item.help.Valid {
				buf.token(TokenBlockStart(BlockItemTerm))
				buf.token(TokenBlockEnd(BlockItemTerm))
			}
			buf.token(TokenBlockStart(BlockItemBody))

			if include_env {
				buf.write_str(fmt.Sprintf("[env:%v%v]", env, val), StyleText)
			} else {
				buf.Text("Uses environment variable ")
				buf.Literal(env)
			}

			buf.token(TokenBlockEnd(BlockItemBody))
		}
	case HelpItemAnywhereStart:
		buf.token(TokenBlockStart(BlockSection3))
		buf.write_meta(item.inner, true)
		buf.token(TokenBlockEnd(BlockSection3))
	case HelpItemAnywhereStop:
		buf.token(TokenBlockStart(BlockBlock))
		buf.token(TokenBlockEnd(BlockBlock))
	}
}

func write_shortlong(buf *Doc, name ShortLong) {
	if s := name.Short; s != 0 {
		buf.write_char('-', StyleLiteral)
		buf.write_char(s, StyleLiteral)
	}
	if name.Short != 0 && name.Long != "" {
		buf.write_str(", ", StyleText)
	}
	if l := name.Long; l != "" {
		buf.write_str("--", StyleLiteral)
		buf.write_str(l, StyleLiteral)
	}
}

func render_help(
	path []string,
	info Info,
	parser_meta Meta,
	help_meta Meta,
	include_env bool,
) Doc {
	positional_invariant_check(parser_meta, false)
	buf := Doc{}

	if t, ok := info.descr.Unpack(); ok {
		buf.token(TokenBlockStart(BlockBlock))
		buf.Doc(t)
		buf.token(TokenBlockEnd(BlockBlock))
	}

	buf.token(TokenBlockStart(BlockBlock))
	if usage, ok := info.usage.Unpack(); ok {
		buf.Doc(usage)
	} else {
		buf.write_str("Usage", StyleEmphasis)
		buf.write_str(": ", StyleText)
		buf.token(TokenBlockStart(BlockMono))
		buf.write_path(path)
		buf.write_meta(parser_meta, true)
		buf.token(TokenBlockEnd(BlockMono))
	}
	buf.token(TokenBlockEnd(BlockBlock))

	if t, ok := info.header.Unpack(); ok {
		buf.token(TokenBlockStart(BlockBlock))
		buf.Doc(t)
		buf.token(TokenBlockEnd(BlockBlock))
	}

	items := HelpItemsdefault()
	items.append_meta(parser_meta)
	items.append_meta(help_meta)

	buf.write_help_item_groups(items, include_env)

	if footer, ok := info.footer.Unpack(); ok {
		buf.token(TokenBlockStart(BlockBlock))
		buf.Doc(footer)
		buf.token(TokenBlockEnd(BlockBlock))
	}
	return buf
}

type Dedup struct {
	items set.Set[string]
	keep  bool
}

func Dedupdefault() Dedup {
	return Dedup{set.New[string](0), false}
}

func insert[K comparable](s *set.Set[K], item K) bool {
	if s.Contains(item) {
		return false
	}
	s.Add(item)
	return true
}

// TODO: mutate somehow vahui
func (self *Dedup) check(item HelpItem) bool {
	switch item := item.(type) {
	case HelpItemDecorSuffix:
		return self.keep
	case HelpItemAny:
		self.keep = insert(&self.items, fmt.Sprintf("%#v %#v", item.metavar, item.help))
		return self.keep
	case HelpItemPositional:
		self.keep = insert(&self.items, fmt.Sprintf("%#v %#v", item.metavar, item.help))
		return self.keep
	case HelpItemCommand:
		self.keep = insert(&self.items, fmt.Sprintf("%#v %#v", item.name, item.help))
		return self.keep
	case HelpItemFlag:
		self.keep = insert(&self.items, fmt.Sprintf("%#v %#v", item.name, item.help))
		return self.keep
	case HelpItemArgument:
		self.keep = insert(&self.items, fmt.Sprintf("%#v %v %#v", item.name, item.metavar, item.help))
		return self.keep
	case HelpItemGroupStart, HelpItemGroupEnd, HelpItemAnywhereStart, HelpItemAnywhereStop:
		self.keep = true
		return true
	}
	panic("unreachable")
}

func (self *Doc) write_help_item_groups(items HelpItems, include_env bool) {
	for {
		rangee, ok := items.find_group()
		if !ok {
			break
		}

		dd := Dedupdefault()
		drain := items[rangee.start():rangee.end()]
		items = append(items[:rangee.start():rangee.start()], items[rangee.end():]...)
		for _, item := range drain {
			if dd.check(item) {
				write_help_item(self, item, include_env)
			}
		}
	}

	for _, tyname := range []struct {
		HiTy HiTy
		name string
	}{
		{HiTyPositional, "Available positional items:"},
		{HiTyFlag, "Available options:"},
		{HiTyCommand, "Available commands:"},
	} {
		ty, name := tyname.HiTy, tyname.name
		self.write_help_items(items, ty, name, include_env)
	}
}

func (self *Doc) write_help_items(items HelpItems, ty HiTy, name string, include_env bool) {
	zalupa := func() bool {
		for range items.items_of_ty(ty) {
			return true
		}
		return false
	}()
	xs := items.items_of_ty(ty)
	if zalupa {
		self.token(TokenBlockStart(BlockBlock))
		self.token(TokenBlockStart(BlockSection2))
		self.write_str(name, StyleEmphasis)
		self.token(TokenBlockEnd(BlockSection2))
		self.token(TokenBlockStart(BlockDefinitionList))
		dd := Dedupdefault()
		for item := range xs {
			if dd.check(item) {
				write_help_item(self, item, include_env)
			}
		}
		self.token(TokenBlockEnd(BlockDefinitionList))
		self.token(TokenBlockEnd(BlockBlock))
	}
}

func (self *Doc) write_path(path []string) {
	for _, item := range path {
		self.write_str(item, StyleLiteral)
		self.write_char(' ', StyleText)
	}
}
