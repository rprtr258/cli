package cli

import (
	"fmt"
	"iter"
	"os"
	"unicode"

	"github.com/rprtr258/fun"
	"github.com/rprtr258/fun/set"
)

// Metavar is a placeholder for a value shown in help output, e.g. "FILE" or "N".
type Metavar string

type (
	helpItem            interface{ isHelpItem() }
	helpItemDecorSuffix struct {
		help Doc
		ty   hiTy
	}
)

func (helpItemDecorSuffix) isHelpItem() {}

type helpItemGroupStart struct {
	help Doc
	ty   hiTy
}

func (helpItemGroupStart) isHelpItem() {}

type helpItemGroupEnd struct{ ty hiTy }

func (helpItemGroupEnd) isHelpItem() {}

type helpItemAny struct {
	metavar  Doc
	anywhere bool
	help     Opt[Doc]
}

func (helpItemAny) isHelpItem() {}

type helpItemPositional struct {
	metavar Metavar
	help    Opt[Doc]
}

func (helpItemPositional) isHelpItem() {}

type helpItemCommand struct {
	name  string
	short Opt[byte]
	help  Opt[Doc]
	meta  meta
	info  info
}

func (helpItemCommand) isHelpItem() {}

type helpItemFlag struct {
	name shortLong
	env  Opt[string]
	help Opt[Doc]
}

func (helpItemFlag) isHelpItem() {}

type helpItemArgument struct {
	name    shortLong
	metavar Metavar
	env     Opt[string]
	help    Opt[Doc]
	// allowed values shown in help; empty means any
	choices []string
}

func (helpItemArgument) isHelpItem() {}

type helpItemAnywhereStart struct {
	inner meta
	ty    hiTy
}

func (helpItemAnywhereStart) isHelpItem() {}

type helpItemAnywhereStop struct{ ty hiTy }

func (helpItemAnywhereStop) isHelpItem() {}

func has_help(self helpItem) bool {
	switch self := self.(type) {
	case helpItemDecorSuffix, helpItemGroupStart:
		return true
	case helpItemAny:
		return self.help.Valid
	case helpItemPositional:
		return self.help.Valid
	case helpItemCommand:
		return self.help.Valid
	case helpItemFlag:
		return self.help.Valid
	case helpItemArgument:
		return self.help.Valid
	}
	panic("unreachable")
}

func ty(self helpItem) hiTy {
	switch self := self.(type) {
	case helpItemDecorSuffix:
		return self.ty
	case helpItemGroupStart:
		return self.ty
	case helpItemGroupEnd:
		return self.ty
	case helpItemAny:
		return fun.IF(self.anywhere, hiTyFlag, hiTyPositional)
	case helpItemPositional:
		return hiTyPositional
	case helpItemCommand:
		return hiTyCommand
	case helpItemFlag:
		return hiTyFlag
	case helpItemArgument:
		return hiTyFlag
	case helpItemAnywhereStart:
		return self.ty
	case helpItemAnywhereStop:
		return self.ty
	}
	panic("unreachable")
}

// helpItems is a collection of all the help items separated into flags,
// positionals and commands.
//
// Items are stored as references and can be trivially copied.
type helpItems []helpItem

func helpItemsDefault() helpItems {
	return []helpItem{}
}

type hiTy uint8

const (
	hiTyFlag hiTy = iota
	hiTyCommand
	hiTyPositional
)

// itemBlock is either not present, a decorated block with the given type, or
// an anywhere block with the given type.
type itemBlock struct {
	present bool
	isDecor bool // anywhere otherwise
	ty      hiTy
}

var itemBlockNo = itemBlock{}

func itemBlockDecor(ty hiTy) itemBlock {
	return itemBlock{true, true, ty}
}

func itemBlockAnywhere(ty hiTy) itemBlock {
	return itemBlock{true, false, ty}
}

type helpItemsIter struct {
	items  []helpItem
	target hiTy
	cur    int
	block  itemBlock
}

func (self *helpItemsIter) next() (helpItem, bool) {
	for {
		if self.cur >= len(self.items) {
			return nil, false
		}
		item := self.items[self.cur]
		self.cur += 1

		var keep bool
		switch it := item.(type) {
		case helpItemDecorSuffix, helpItemAny, helpItemPositional, helpItemCommand, helpItemFlag, helpItemArgument:
			ty := ty(item)
			switch {
			case !self.block.present:
				keep = ty == self.target
			case self.block.isDecor:
				keep = self.block.ty == self.target
			default:
				keep = self.block.ty == self.target && has_help(item)
			}
		case helpItemGroupStart:
			self.block = itemBlockDecor(it.ty)
			keep = it.ty == self.target
		case helpItemGroupEnd:
			self.block = itemBlockNo
			keep = it.ty == self.target
		case helpItemAnywhereStart:
			self.block = itemBlockAnywhere(it.ty)
			keep = it.ty == self.target
		case helpItemAnywhereStop:
			self.block = itemBlockNo
			keep = it.ty == self.target
		}
		if keep {
			return item, true
		}
	}
}

func (self helpItems) items_of_ty(target hiTy) iter.Seq[helpItem] {
	it := helpItemsIter{
		items:  self,
		target: target,
		cur:    0,
		block:  itemBlockNo,
	}
	return func(yield func(helpItem) bool) {
		for {
			x, ok := it.next()
			if !ok || !yield(x) {
				break
			}
		}
	}
}

func peek_front_ty(self meta) (hiTy, bool) {
	var (
		res hiTy
		ok  bool
	)
	switch m := self.(type) {
	case metaAnd:
		res, ok = find_map(m, peek_front_ty)
	case metaOr:
		res, ok = find_map(m, peek_front_ty)
	case metaOptional:
		res, ok = peek_front_ty(m.m)
	case metaRequired:
		res, ok = peek_front_ty(m.m)
	case metaAdjacent:
		res, ok = peek_front_ty(m.m)
	case metaItem:
		res, ok = hiTyFromItem(item(m)), true
	case metaMany:
		res, ok = peek_front_ty(m.m)
	case metaSubsection:
		res, ok = peek_front_ty(m.m)
	case metaSuffix:
		res, ok = peek_front_ty(m.m)
	case metaCustomUsage:
		res, ok = peek_front_ty(m.m)
	}
	return res, ok
}

// append_meta recursively classifies the contents of the meta.
func (self *helpItems) append_meta(mm meta) {
	var GO func(mm meta, no_ss bool)
	GO = func(mm meta, no_ss bool) {
		switch mm := mm.(type) {
		case metaAnd:
			for _, x := range mm {
				GO(x, no_ss)
			}
		case metaOr:
			for _, x := range mm {
				GO(x, no_ss)
			}
		case metaAdjacent:
			if ty, ok := peek_front_ty(mm.m); ok {
				*self = append(*self, helpItemAnywhereStart{mm.m, ty})
				GO(mm.m, no_ss)
				*self = append(*self, helpItemAnywhereStop{ty})
			}
		case metaStrict:
			GO(mm.m, no_ss)
		case metaCustomUsage:
			GO(mm.m, no_ss)
		case metaMany:
			GO(mm.m, no_ss)
		case metaRequired:
			GO(mm.m, no_ss)
		case metaOptional:
			GO(mm.m, no_ss)
		case metaItem:
			if _, ok := mm.v.(itemPositional); ok && !mm.help.Valid {
				return
			}
			*self = append(*self, helpItemFromItem(item(mm)))
		case metaSubsection:
			m := mm.m
			help := mm.doc
			if ty, ok := peek_front_ty(m); ok {
				if no_ss {
					GO(m, true)
				} else {
					*self = append(*self, helpItemGroupStart{help, ty})
					GO(m, true)
					*self = append(*self, helpItemGroupEnd{ty})
				}
			}
		case metaSuffix:
			m := mm.m
			help := mm.doc
			if ty, ok := peek_front_ty(m); ok {
				GO(m, no_ss)
				*self = append(*self, helpItemDecorSuffix{help, ty})
			}
		case metaSkip:
		}
	}
	GO(mm, false)
}

func (self helpItems) find_group() (rng, bool) {
	start, end := -1, -1
	for i, x := range self {
		var isStart, isEnd bool
		switch x.(type) {
		case helpItemGroupStart:
			isStart = true
		case helpItemGroupEnd:
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
		return rng{}, false
	}
	return rng{start, end + 1}, true
}

func hiTyFromItem(value item) hiTy {
	switch v := value.v.(type) {
	case itemAny:
		return fun.IF(v.anywhere, hiTyFlag, hiTyPositional)
	case itemPositional:
		return hiTyPositional
	case itemCommand:
		return hiTyCommand
	case itemFlag:
		return hiTyFlag
	case itemArgument:
		return hiTyFlag
	}
	panic("unreachable")
}

func helpItemFromItem(item item) helpItem {
	switch v := item.v.(type) {
	case itemAny:
		return helpItemAny{v.metavar, v.anywhere, item.help}
	case itemPositional:
		return helpItemPositional{v.metavar, item.help}
	case itemCommand:
		return helpItemCommand{v.name, v.short, item.help, v.meta, v.info}
	case itemFlag:
		return helpItemFlag{v.name, v.env, item.help}
	case itemArgument:
		return helpItemArgument{v.name, v.metavar, v.env, item.help, v.choices}
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

func write_help_item(buf *Doc, item helpItem, include_env bool) {
	switch item := item.(type) {
	case helpItemDecorSuffix:
		buf.token(tokenBlockStart(blockItemTerm))
		buf.token(tokenBlockEnd(blockItemTerm))
		buf.token(tokenBlockStart(blockItemBody))
		buf.Doc(item.help)
		buf.token(tokenBlockEnd(blockItemBody))
	case helpItemGroupStart:
		buf.token(tokenBlockStart(blockBlock))
		buf.token(tokenBlockStart(blockSection2))
		buf.emDoc(item.help)
		buf.token(tokenBlockEnd(blockSection2))
		buf.token(tokenBlockStart(blockDefinitionList))
	case helpItemGroupEnd:
		buf.token(tokenBlockEnd(blockDefinitionList))
		buf.token(tokenBlockEnd(blockBlock))
	case helpItemAny:
		buf.token(tokenBlockStart(blockItemTerm))
		buf.Doc(item.metavar)
		buf.token(tokenBlockEnd(blockItemTerm))
		if help, ok := item.help.Unpack(); ok {
			buf.token(tokenBlockStart(blockItemBody))
			buf.Doc(help)
			buf.token(tokenBlockEnd(blockItemBody))
		}
	case helpItemPositional:
		buf.token(tokenBlockStart(blockItemTerm))
		buf.metavar(item.metavar)
		buf.token(tokenBlockEnd(blockItemTerm))
		if help, ok := item.help.Unpack(); ok {
			buf.token(tokenBlockStart(blockItemBody))
			buf.Doc(help)
			buf.token(tokenBlockEnd(blockItemBody))
		}
	case helpItemCommand:
		buf.token(tokenBlockStart(blockItemTerm))
		buf.write_str(item.name, StyleLiteral)
		if short, ok := item.short.Unpack(); ok {
			buf.write_str(", ", StyleText)
			buf.write_char(short, StyleLiteral)
		}
		buf.token(tokenBlockEnd(blockItemTerm))
		if help, ok := item.help.Unpack(); ok {
			buf.token(tokenBlockStart(blockItemBody))
			buf.Doc(help)
			buf.token(tokenBlockEnd(blockItemBody))
		}
	case helpItemFlag:
		buf.token(tokenBlockStart(blockItemTerm))
		write_shortlong(buf, item.name)
		buf.token(tokenBlockEnd(blockItemTerm))
		if help, ok := item.help.Unpack(); ok {
			buf.token(tokenBlockStart(blockItemBody))
			buf.Doc(help)
			buf.token(tokenBlockEnd(blockItemBody))
		}
		if env, ok := item.env.Unpack(); ok {
			_, found := os.LookupEnv(env)
			val := fun.IF(found, ": set", ": not set")
			if item.help.Valid {
				buf.token(tokenBlockStart(blockItemTerm))
				buf.token(tokenBlockEnd(blockItemTerm))
			}
			buf.token(tokenBlockStart(blockItemBody))
			if include_env {
				buf.write_str(fmt.Sprintf("[env:%v%v]", env, val), StyleText)
			} else {
				buf.Text("Uses environment variable ")
				buf.Literal(env)
			}
			buf.token(tokenBlockEnd(blockItemBody))
		}
	case helpItemArgument:
		buf.token(tokenBlockStart(blockItemTerm))
		write_shortlong(buf, item.name)
		buf.write_str("=", StyleText)
		buf.metavar(item.metavar)
		buf.token(tokenBlockEnd(blockItemTerm))

		if help, ok := item.help.Unpack(); ok {
			buf.token(tokenBlockStart(blockItemBody))
			buf.Doc(help)
			buf.token(tokenBlockEnd(blockItemBody))
		}

		if len(item.choices) > 0 {
			buf.token(tokenBlockStart(blockItemBody))
			buf.Text("(one of: ")
			for i, choice := range item.choices {
				if i > 0 {
					buf.Text(", ")
				}
				buf.Literal(choice)
			}
			buf.Text(")")
			buf.token(tokenBlockEnd(blockItemBody))
		}

		if env, ok := item.env.Unpack(); ok {
			val := ": N/A"
			if s, ok := os.LookupEnv(env); ok {
				val = fmt.Sprintf(" = %#v", s)
			}

			if item.help.Valid {
				buf.token(tokenBlockStart(blockItemTerm))
				buf.token(tokenBlockEnd(blockItemTerm))
			}
			buf.token(tokenBlockStart(blockItemBody))

			if include_env {
				buf.write_str(fmt.Sprintf("[env:%v%v]", env, val), StyleText)
			} else {
				buf.Text("Uses environment variable ")
				buf.Literal(env)
			}

			buf.token(tokenBlockEnd(blockItemBody))
		}
	case helpItemAnywhereStart:
		buf.token(tokenBlockStart(blockSection3))
		buf.write_meta(item.inner, true)
		buf.token(tokenBlockEnd(blockSection3))
	case helpItemAnywhereStop:
		buf.token(tokenBlockStart(blockBlock))
		buf.token(tokenBlockEnd(blockBlock))
	}
}

func write_shortlong(buf *Doc, name shortLong) {
	switch {
	case name.is_short():
		buf.write_char('-', StyleLiteral)
		buf.write_char(name.Short, StyleLiteral)
	case name.is_long():
		// long only arguments are indented to match the short ones
		buf.write_str("    --", StyleLiteral)
		buf.write_str(name.Long, StyleLiteral)
	case name.is_both():
		buf.write_char('-', StyleLiteral)
		buf.write_char(name.Short, StyleLiteral)
		buf.write_str(", ", StyleText)
		buf.write_str("--", StyleLiteral)
		buf.write_str(name.Long, StyleLiteral)
	}
}

func render_help(
	path []string,
	info *info,
	parser_meta meta,
	help_meta meta,
) Doc {
	positional_invariant_check(parser_meta)
	buf := Doc{}

	if t, ok := info.descr.Unpack(); ok {
		buf.token(tokenBlockStart(blockBlock))
		buf.Doc(t)
		buf.token(tokenBlockEnd(blockBlock))
	}

	buf.token(tokenBlockStart(blockBlock))
	if usage, ok := info.usage.Unpack(); ok {
		buf.Doc(usage)
	} else {
		buf.write_str("Usage", StyleEmphasis)
		buf.write_str(": ", StyleText)
		buf.token(tokenBlockStart(blockMono))
		buf.write_path(path)
		buf.write_meta(parser_meta, true)
		buf.token(tokenBlockEnd(blockMono))
	}
	buf.token(tokenBlockEnd(blockBlock))

	if t, ok := info.header.Unpack(); ok {
		buf.token(tokenBlockStart(blockBlock))
		buf.Doc(t)
		buf.token(tokenBlockEnd(blockBlock))
	}

	items := helpItemsDefault()
	items.append_meta(parser_meta)
	items.append_meta(help_meta)

	buf.write_help_item_groups(items, true)

	buf.write_examples(path, info.examples)

	if footer, ok := info.footer.Unpack(); ok {
		buf.token(tokenBlockStart(blockBlock))
		buf.Doc(footer)
		buf.token(tokenBlockEnd(blockBlock))
	}
	return buf
}

type dedup struct {
	items set.Set[string]
	keep  bool
}

func dedupDefault() dedup {
	return dedup{set.New[string](0), false}
}

func insert[K comparable](s *set.Set[K], item K) bool {
	if s.Contains(item) {
		return false
	}
	s.Add(item)
	return true
}

// check deduplicates the item, storing it and reporting whether it should
// be kept in the help output.
func (self *dedup) check(item helpItem) bool {
	switch item := item.(type) {
	case helpItemDecorSuffix:
		return self.keep
	case helpItemAny:
		self.keep = insert(&self.items, fmt.Sprintf("%#v %#v", item.metavar, item.help))
		return self.keep
	case helpItemPositional:
		self.keep = insert(&self.items, fmt.Sprintf("%#v %#v", item.metavar, item.help))
		return self.keep
	case helpItemCommand:
		self.keep = insert(&self.items, fmt.Sprintf("%#v %#v", item.name, item.help))
		return self.keep
	case helpItemFlag:
		self.keep = insert(&self.items, fmt.Sprintf("%#v %#v", item.name, item.help))
		return self.keep
	case helpItemArgument:
		self.keep = insert(&self.items, fmt.Sprintf("%#v %v %#v %#v", item.name, item.metavar, item.help, item.choices))
		return self.keep
	case helpItemGroupStart, helpItemGroupEnd, helpItemAnywhereStart, helpItemAnywhereStop:
		self.keep = true
		return true
	}
	panic("unreachable")
}

func (self *Doc) write_help_item_groups(items helpItems, include_env bool) {
	for {
		rangee, ok := items.find_group()
		if !ok {
			break
		}

		dd := dedupDefault()
		drain := items[rangee.start():rangee.end()]
		items = append(items[:rangee.start():rangee.start()], items[rangee.end():]...)
		for _, item := range drain {
			if dd.check(item) {
				write_help_item(self, item, include_env)
			}
		}
	}

	for _, tyname := range []struct {
		hiTy hiTy
		name string
	}{
		{hiTyPositional, "Available positional items:"},
		{hiTyFlag, "Available options:"},
		{hiTyCommand, "Available commands:"},
	} {
		ty, name := tyname.hiTy, tyname.name
		self.write_help_items(items, ty, name, include_env)
	}
}

func (self *Doc) write_help_items(items helpItems, ty hiTy, name string, include_env bool) {
	zalupa := func() bool {
		for range items.items_of_ty(ty) {
			return true
		}
		return false
	}()
	xs := items.items_of_ty(ty)
	if zalupa {
		self.token(tokenBlockStart(blockBlock))
		self.token(tokenBlockStart(blockSection2))
		self.write_str(name, StyleEmphasis)
		self.token(tokenBlockEnd(blockSection2))
		self.token(tokenBlockStart(blockDefinitionList))
		dd := dedupDefault()
		for item := range xs {
			if dd.check(item) {
				write_help_item(self, item, include_env)
			}
		}
		self.token(tokenBlockEnd(blockDefinitionList))
		self.token(tokenBlockEnd(blockBlock))
	}
}

func (self *Doc) write_path(path []string) {
	for _, item := range path {
		self.write_str(item, StyleLiteral)
		self.write_char(' ', StyleText)
	}
}
