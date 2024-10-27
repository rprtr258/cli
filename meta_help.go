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

type HelpItem func(
	DecorSuffix func(help Doc, ty HiTy),
	GroupStart func(help Doc, ty HiTy),
	GroupEnd func(ty HiTy),
	Any func(metavar Doc, anywhere bool, help Option[Doc]),
	Positional func(metavar Metavar, help Option[Doc]),
	Command func(name string, short Option[byte], help Option[Doc], meta Meta, info Info),
	Flag func(name ShortLong, env Option[string], help Option[Doc]),
	Argument func(name ShortLong, metavar Metavar, env Option[string], help Option[Doc]),
	AnywhereStart func(inner Meta, ty HiTy),
	AnywhereStop func(ty HiTy),
)

func HelpItemDecorSuffix(help Doc, ty HiTy) HelpItem {
	return func(
		f func(help Doc, ty HiTy),
		_ func(help Doc, ty HiTy),
		_ func(ty HiTy),
		_ func(metavar Doc, anywhere bool, help Option[Doc]),
		_ func(metavar Metavar, help Option[Doc]),
		_ func(name string, short Option[byte], help Option[Doc], meta Meta, info Info),
		_ func(name ShortLong, env Option[string], help Option[Doc]),
		_ func(name ShortLong, metavar Metavar, env Option[string], help Option[Doc]),
		_ func(inner Meta, ty HiTy),
		_ func(ty HiTy),
	) {
		if f != nil {
			f(help, ty)
		}
	}
}

func HelpItemGroupStart(help Doc, ty HiTy) HelpItem {
	return func(
		_ func(help Doc, ty HiTy),
		f func(help Doc, ty HiTy),
		_ func(ty HiTy),
		_ func(metavar Doc, anywhere bool, help Option[Doc]),
		_ func(metavar Metavar, help Option[Doc]),
		_ func(name string, short Option[byte], help Option[Doc], meta Meta, info Info),
		_ func(name ShortLong, env Option[string], help Option[Doc]),
		_ func(name ShortLong, metavar Metavar, env Option[string], help Option[Doc]),
		_ func(inner Meta, ty HiTy),
		_ func(ty HiTy),
	) {
		if f != nil {
			f(help, ty)
		}
	}
}

func HelpItemGroupEnd(ty HiTy) HelpItem {
	return func(
		_ func(help Doc, ty HiTy),
		_ func(help Doc, ty HiTy),
		f func(ty HiTy),
		_ func(metavar Doc, anywhere bool, help Option[Doc]),
		_ func(metavar Metavar, help Option[Doc]),
		_ func(name string, short Option[byte], help Option[Doc], meta Meta, info Info),
		_ func(name ShortLong, env Option[string], help Option[Doc]),
		_ func(name ShortLong, metavar Metavar, env Option[string], help Option[Doc]),
		_ func(inner Meta, ty HiTy),
		_ func(ty HiTy),
	) {
		if f != nil {
			f(ty)
		}
	}
}

func HelpItemAny(metavar Doc, anywhere bool, help Option[Doc]) HelpItem {
	return func(
		_ func(help Doc, ty HiTy),
		_ func(help Doc, ty HiTy),
		_ func(ty HiTy),
		f func(metavar Doc, anywhere bool, help Option[Doc]),
		_ func(metavar Metavar, help Option[Doc]),
		_ func(name string, short Option[byte], help Option[Doc], meta Meta, info Info),
		_ func(name ShortLong, env Option[string], help Option[Doc]),
		_ func(name ShortLong, metavar Metavar, env Option[string], help Option[Doc]),
		_ func(inner Meta, ty HiTy),
		_ func(ty HiTy),
	) {
		if f != nil {
			f(metavar, anywhere, help)
		}
	}
}

func HelpItemPositional(metavar Metavar, help Option[Doc]) HelpItem {
	return func(
		_ func(help Doc, ty HiTy),
		_ func(help Doc, ty HiTy),
		_ func(ty HiTy),
		_ func(metavar Doc, anywhere bool, help Option[Doc]),
		f func(metavar Metavar, help Option[Doc]),
		_ func(name string, short Option[byte], help Option[Doc], meta Meta, info Info),
		_ func(name ShortLong, env Option[string], help Option[Doc]),
		_ func(name ShortLong, metavar Metavar, env Option[string], help Option[Doc]),
		_ func(inner Meta, ty HiTy),
		_ func(ty HiTy),
	) {
		if f != nil {
			f(metavar, help)
		}
	}
}

func HelpItemCommand(name string, short Option[byte], help Option[Doc], meta Meta, info Info) HelpItem {
	return func(
		_ func(help Doc, ty HiTy),
		_ func(help Doc, ty HiTy),
		_ func(ty HiTy),
		_ func(metavar Doc, anywhere bool, help Option[Doc]),
		_ func(metavar Metavar, help Option[Doc]),
		f func(name string, short Option[byte], help Option[Doc], meta Meta, info Info),
		_ func(name ShortLong, env Option[string], help Option[Doc]),
		_ func(name ShortLong, metavar Metavar, env Option[string], help Option[Doc]),
		_ func(inner Meta, ty HiTy),
		_ func(ty HiTy),
	) {
		if f != nil {
			f(name, short, help, meta, info)
		}
	}
}

func HelpItemFlag(name ShortLong, env Option[string], help Option[Doc]) HelpItem {
	return func(
		_ func(help Doc, ty HiTy),
		_ func(help Doc, ty HiTy),
		_ func(ty HiTy),
		_ func(metavar Doc, anywhere bool, help Option[Doc]),
		_ func(metavar Metavar, help Option[Doc]),
		_ func(name string, short Option[byte], help Option[Doc], meta Meta, info Info),
		f func(name ShortLong, env Option[string], help Option[Doc]),
		_ func(name ShortLong, metavar Metavar, env Option[string], help Option[Doc]),
		_ func(inner Meta, ty HiTy),
		_ func(ty HiTy),
	) {
		if f != nil {
			f(name, env, help)
		}
	}
}

func HelpItemArgument(name ShortLong, metavar Metavar, env Option[string], help Option[Doc]) HelpItem {
	return func(
		_ func(help Doc, ty HiTy),
		_ func(help Doc, ty HiTy),
		_ func(ty HiTy),
		_ func(metavar Doc, anywhere bool, help Option[Doc]),
		_ func(metavar Metavar, help Option[Doc]),
		_ func(name string, short Option[byte], help Option[Doc], meta Meta, info Info),
		_ func(name ShortLong, env Option[string], help Option[Doc]),
		f func(name ShortLong, metavar Metavar, env Option[string], help Option[Doc]),
		_ func(inner Meta, ty HiTy),
		_ func(ty HiTy),
	) {
		if f != nil {
			f(name, metavar, env, help)
		}
	}
}

func HelpItemAnywhereStart(inner Meta, ty HiTy) HelpItem {
	return func(
		_ func(help Doc, ty HiTy),
		_ func(help Doc, ty HiTy),
		_ func(ty HiTy),
		_ func(metavar Doc, anywhere bool, help Option[Doc]),
		_ func(metavar Metavar, help Option[Doc]),
		_ func(name string, short Option[byte], help Option[Doc], meta Meta, info Info),
		_ func(name ShortLong, env Option[string], help Option[Doc]),
		_ func(name ShortLong, metavar Metavar, env Option[string], help Option[Doc]),
		f func(inner Meta, ty HiTy),
		_ func(ty HiTy),
	) {
		if f != nil {
			f(inner, ty)
		}
	}
}

func HelpItemAnywhereStop(ty HiTy) HelpItem {
	return func(
		_ func(help Doc, ty HiTy),
		_ func(help Doc, ty HiTy),
		_ func(ty HiTy),
		_ func(metavar Doc, anywhere bool, help Option[Doc]),
		_ func(metavar Metavar, help Option[Doc]),
		_ func(name string, short Option[byte], help Option[Doc], meta Meta, info Info),
		_ func(name ShortLong, env Option[string], help Option[Doc]),
		_ func(name ShortLong, metavar Metavar, env Option[string], help Option[Doc]),
		_ func(inner Meta, ty HiTy),
		f func(ty HiTy),
	) {
		if f != nil {
			f(ty)
		}
	}
}

func (self HelpItem) has_help() bool {
	res := false
	self(
		func(Doc, HiTy) { res = true },
		func(Doc, HiTy) { res = true },
		nil,
		func(_ Doc, _ bool, help Option[Doc]) { res = help.Valid },
		func(_ Metavar, help Option[Doc]) { res = help.Valid },
		func(_ string, _ Option[byte], help Option[Doc], meta Meta, info Info) { res = help.Valid },
		func(_ ShortLong, _ Option[string], help Option[Doc]) { res = help.Valid },
		func(_ ShortLong, _ Metavar, env Option[string], help Option[Doc]) { res = help.Valid },
		nil,
		nil,
	)
	return res
}

func (self HelpItem) ty() HiTy {
	var res HiTy
	self(
		func(_ Doc, ty HiTy) { res = ty },
		func(help Doc, ty HiTy) { res = ty },
		func(ty HiTy) { res = ty },
		func(_ Doc, anywhere bool, _ Option[Doc]) {
			if anywhere {
				res = HiTyFlag
			} else {
				res = HiTyPositional
			}
		},
		func(_ Metavar, _ Option[Doc]) { res = HiTyPositional },
		func(_ string, _ Option[byte], _ Option[Doc], _ Meta, _ Info) { res = HiTyCommand },
		func(_ ShortLong, _ Option[string], _ Option[Doc]) { res = HiTyFlag },
		func(_ ShortLong, _ Metavar, _ Option[string], _ Option[Doc]) { res = HiTyFlag },
		func(_ Meta, ty HiTy) { res = ty },
		func(ty HiTy) { res = ty },
	)
	return res
}

// A collection of all the help items separated into flags, positionals and commands
//
// Items are stored as references and can be trivially copied
type HelpItems struct {
	items []HelpItem
}

func HelpItemsdefault() HelpItems {
	return HelpItems{[]HelpItem{}}
}

type HiTy func(
	Flag func(),
	Command func(),
	Positional func(),
)

var HiTyFlag HiTy = func(Flag, _, _ func()) {
	if Flag != nil {
		Flag()
	}
}
var HiTyCommand HiTy = func(_, Command, _ func()) {
	if Command != nil {
		Command()
	}
}
var HiTyPositional HiTy = func(_, _, Positional func()) {
	if Positional != nil {
		Positional()
	}
}

func HiTyCompare(a, b HiTy) bool {
	var res bool
	a(
		func() { b(func() { res = true }, nil, nil) },
		func() { b(nil, func() { res = true }, nil) },
		func() { b(nil, nil, func() { res = true }) },
	)
	return res
}

type ItemBlock func(
	No func(),
	Decor func(HiTy),
	Anywhere func(HiTy),
)

func ItemBlockNo() ItemBlock {
	return func(
		f func(),
		_ func(HiTy),
		_ func(HiTy),
	) {
		if f != nil {
			f()
		}
	}
}

func ItemBlockDecor(ty HiTy) ItemBlock {
	return func(
		_ func(),
		f func(HiTy),
		_ func(HiTy),
	) {
		if f != nil {
			f(ty)
		}
	}
}

func ItemBlockAnywhere(ty HiTy) ItemBlock {
	return func(
		_ func(),
		_ func(HiTy),
		f func(HiTy),
	) {
		if f != nil {
			f(ty)
		}
	}
}

type HelpItemsIter struct {
	items  []HelpItem
	target HiTy
	cur    int
	block  ItemBlock
}

func (self HelpItemsIter) next() Option[HelpItem] {
	for {
		if self.cur >= len(self.items) {
			return Invalid[HelpItem]()
		}
		item := self.items[self.cur]
		self.cur += 1

		var keep bool
		dhh := func() {
			ty := item.ty()
			self.block(
				func() { keep = HiTyCompare(ty, self.target) },
				func(t HiTy) { keep = HiTyCompare(t, self.target) },
				func(t HiTy) { keep = HiTyCompare(t, self.target) && item.has_help() },
			)
		}
		item(
			func(Doc, HiTy) { dhh() },
			func(_ Doc, ty HiTy) {
				self.block = ItemBlockDecor(ty)
				keep = HiTyCompare(ty, self.target)
			},
			func(ty HiTy) {
				self.block = ItemBlockNo()
				keep = HiTyCompare(ty, self.target)
			},
			func(Doc, bool, Option[Doc]) { dhh() },
			func(Metavar, Option[Doc]) { dhh() },
			func(string, Option[byte], Option[Doc], Meta, Info) { dhh() },
			func(ShortLong, Option[string], Option[Doc]) { dhh() },
			func(ShortLong, Metavar, Option[string], Option[Doc]) { dhh() },
			func(inner Meta, ty HiTy) {
				self.block = ItemBlockAnywhere(ty)
				keep = HiTyCompare(ty, self.target)
			},
			func(ty HiTy) {
				self.block = ItemBlockNo()
				keep = HiTyCompare(ty, self.target)
			},
		)
		if keep {
			return Valid(item)
		}
	}
}

func (self HelpItems) items_of_ty(target HiTy) iter.Seq[HelpItem] {
	iter := HelpItemsIter{
		items:  self.items,
		target: target,
		cur:    0,
		block:  ItemBlockNo(),
	}
	return func(yield func(HelpItem) bool) {
		for {
			x, ok := iter.next().Unpack()
			if !ok || !yield(x) {
				break
			}
		}
	}
}

func (self Meta) peek_front_ty() (HiTy, bool) {
	var (
		res HiTy
		ok  bool
	)
	self(
		func(xs []Meta) { res, ok = find_map(xs, (Meta).peek_front_ty) },
		func(xs []Meta) { res, ok = find_map(xs, (Meta).peek_front_ty) },
		func(x Meta) { res, ok = x.peek_front_ty() },
		func(x Meta) { res, ok = x.peek_front_ty() },
		func(x Meta) { res, ok = x.peek_front_ty() },
		func(i Item) { res, ok = HiTyFromItem(i), true },
		func(x Meta) { res, ok = x.peek_front_ty() },
		func(x Meta, _ Doc) { res, ok = x.peek_front_ty() },
		func(x Meta, _ Doc) { res, ok = x.peek_front_ty() },
		nil,
		func(x Meta, _ Doc) { res, ok = x.peek_front_ty() },
		nil,
	)
	return res, ok
}

// Recursively classify contents of the Meta
func (self *HelpItems) append_meta(meta Meta) {
	GO := func(hi *HelpItems, meta Meta, no_ss bool) {
		switch meta {
		// MetaAnd(xs) | MetaOr(xs){
		//     for x in xs {
		//         GO(hi, x, no_ss);
		//     }
		// }
		// MetaAdjacent(m){
		//     if Valid(ty) := m.peek_front_ty() {
		//         hi.items=append(hi.items,HelpItemAnywhereStart {
		//             inner: m.as_ref(),
		//             ty,
		//         });
		//         GO(hi, m, no_ss);
		//         hi.items=append(hi.items,HelpItemAnywhereStop { ty });
		//     }
		// }
		// MetaCustomUsage(x, _)
		// | MetaRequired(x)
		// | MetaOptional(x)
		// | MetaMany(x)
		// | MetaStrict(x)GO(hi, x, no_ss),
		// MetaItem(item){
		//     if matches(item.as_ref(), ItemPositional { help: Invalid(), .. }) {
		//         return;
		//     }
		//     hi.items=append(hi.items,HelpItemfrom(item.as_ref()));
		// }
		// MetaSubsection(m, help){
		//     if Valid(ty) := m.peek_front_ty() {
		//         if no_ss {
		//             GO(hi, m, true);
		//         } else {
		//             hi.items=append(hi.items,HelpItemGroupStart { help, ty });
		//             GO(hi, m, true);
		//             hi.items=append(hi.items,HelpItemGroupEnd { ty });
		//         }
		//     }
		// }
		// MetaSuffix(m, help){
		//     if Valid(ty) := m.peek_front_ty() {
		//         GO(hi, m, no_ss);
		//         hi.items=append(hi.items,HelpItemDecorSuffix { help, ty });
		//     }
		// }
		// MetaSkip(),
		}
	}
	GO(self, meta, false)
}
func (self HelpItems) find_group() Option[Range] {
	start, end := -1, -1
	for i, x := range self.items {
		var isStart, isEnd bool
		x(
			nil,
			func(help Doc, ty HiTy) { isStart = true },
			func(ty HiTy) { isEnd = true },
			nil, nil, nil, nil, nil, nil, nil,
		)
		if start == -1 && isStart {
			start = i
		}
		if end == -1 && isEnd {
			end = i
		}
	}
	if start == -1 || end == -1 {
		return Invalid[Range]()
	}
	return Valid(Range{start, end + 1})
}

func HiTyFromItem(value Item) HiTy {
	var res HiTy
	value.v(
		func(_ Doc, anywhere bool) {
			res = fun.IF(anywhere, HiTyFlag, HiTyPositional)
		},
		func(Metavar) { res = HiTyPositional },
		func(string, Option[byte], Meta, Info) { res = HiTyCommand },
		func(ShortLong, []byte, Option[string]) { res = HiTyFlag },
		func(ShortLong, []byte, Metavar, Option[string]) { res = HiTyFlag },
	)
	return res
}

func HelpItemFromItem(item Item) HelpItem {
	var res HelpItem
	item.v(
		func(metavar Doc, anywhere bool) {
			res = HelpItemAny(metavar, anywhere, item.help)
		},
		func(metavar Metavar) {
			res = HelpItemPositional(metavar, item.help)
		},
		func(name string, short Option[byte], meta Meta, info Info) {
			res = HelpItemCommand(name, short, item.help, meta, info)
		},
		func(name ShortLong, shorts []byte, env Option[string]) {
			res = HelpItemFlag(name, env, item.help)
		},
		func(name ShortLong, shorts []byte, metavar Metavar, env Option[string]) {
			res = HelpItemArgument(name, metavar, env, item.help)
		},
	)
	return res
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
	item(
		func(help Doc, _ HiTy) {
			buf.token(TokenBlockStart(BlockItemTerm))
			buf.token(TokenBlockEnd(BlockItemTerm))
			buf.token(TokenBlockStart(BlockItemBody))
			buf.Doc(help)
			buf.token(TokenBlockEnd(BlockItemBody))
		},
		func(help Doc, _ HiTy) {
			buf.token(TokenBlockStart(BlockBlock))
			buf.token(TokenBlockStart(BlockSection2))
			buf.EmDoc(help)
			buf.token(TokenBlockEnd(BlockSection2))
			buf.token(TokenBlockStart(BlockDefinitionList))
		},
		func(HiTy) {
			buf.token(TokenBlockEnd(BlockDefinitionList))
			buf.token(TokenBlockEnd(BlockBlock))
		},
		func(metavar Doc, _ bool, help Option[Doc]) {
			buf.token(TokenBlockStart(BlockItemTerm))
			buf.Doc(metavar)
			buf.token(TokenBlockEnd(BlockItemTerm))
			if help, ok := help.Unpack(); ok {
				buf.token(TokenBlockStart(BlockItemBody))
				buf.Doc(help)
				buf.token(TokenBlockEnd(BlockItemBody))
			}
		},
		func(metavar Metavar, help Option[Doc]) {
			buf.token(TokenBlockStart(BlockItemTerm))
			buf.metavar(metavar)
			buf.token(TokenBlockEnd(BlockItemTerm))
			if help, ok := help.Unpack(); ok {
				buf.token(TokenBlockStart(BlockItemBody))
				buf.Doc(help)
				buf.token(TokenBlockEnd(BlockItemBody))
			}
		},
		func(name string, short Option[byte], help Option[Doc], _ Meta, _ Info) {
			buf.token(TokenBlockStart(BlockItemTerm))
			buf.write_str(name, StyleLiteral)
			if short, ok := short.Unpack(); ok {
				buf.write_str(", ", StyleText)
				buf.write_char(short, StyleLiteral)
			}
			buf.token(TokenBlockEnd(BlockItemTerm))
			if help, ok := help.Unpack(); ok {
				buf.token(TokenBlockStart(BlockItemBody))
				buf.Doc(help)
				buf.token(TokenBlockEnd(BlockItemBody))
			}
		},
		func(name ShortLong, env Option[string], help Option[Doc]) {
			buf.token(TokenBlockStart(BlockItemTerm))
			write_shortlong(buf, name)
			buf.token(TokenBlockEnd(BlockItemTerm))
			if help, ok := help.Unpack(); ok {
				buf.token(TokenBlockStart(BlockItemBody))
				buf.Doc(help)
				buf.token(TokenBlockEnd(BlockItemBody))
			}
			if env, ok := env.Unpack(); ok {
				_, found := os.LookupEnv(env)
				val := fun.IF(found, ": set", ": not set")
				if help.Valid {
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
		},
		func(name ShortLong, metavar Metavar, env Option[string], help Option[Doc]) {
			buf.token(TokenBlockStart(BlockItemTerm))
			write_shortlong(buf, name)
			buf.write_str("=", StyleText)
			buf.metavar(metavar)
			buf.token(TokenBlockEnd(BlockItemTerm))

			if help, ok := help.Unpack(); ok {
				buf.token(TokenBlockStart(BlockItemBody))
				buf.Doc(help)
				buf.token(TokenBlockEnd(BlockItemBody))
			}

			if env, ok := env.Unpack(); ok {
				val := ": N/A"
				if s, ok := os.LookupEnv(env); ok {
					val = fmt.Sprintf(" = %#v", s)
				}

				if help.Valid {
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
		},
		func(inner Meta, _ HiTy) {
			buf.token(TokenBlockStart(BlockSection3))
			buf.write_meta(inner, true)
			buf.token(TokenBlockEnd(BlockSection3))
		},
		func(HiTy) {
			buf.token(TokenBlockStart(BlockBlock))
			buf.token(TokenBlockEnd(BlockBlock))
		},
	)
}

func write_shortlong(buf *Doc, name ShortLong) {
	if s := name.Short; s != 0 {
		buf.write_char('-', StyleLiteral)
		buf.write_char(s, StyleLiteral)
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
	parser_meta.positional_invariant_check(false)
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

// TODO: mutate somehow vahui
func (self Dedup) check(item HelpItem) bool {
	var res bool
	dhh := func() { self.keep = true; res = true }
	item(
		func(Doc, HiTy) { res = self.keep },
		func(Doc, HiTy) { dhh() },
		func(HiTy) { dhh() },
		func(metavar Doc, _ bool, help Option[Doc]) {
			//     self.keep = self.items.insert(fmt.Sprintf("%#v %#v", metavar, help));
			res = self.keep
		},
		func(metavar Metavar, help Option[Doc]) {
			//     self.keep = self.items.insert(fmt.Sprintf("%#v %#v", metavar, help));
			res = self.keep
		},
		func(name string, _ Option[byte], help Option[Doc], _ Meta, _ Info) {
			//     self.keep = self.items.insert(fmt.Sprintf("%#v %#v", name, help));
			res = self.keep
		},
		func(name ShortLong, _ Option[string], help Option[Doc]) {
			//     self.keep = self.items.insert(fmt.Sprintf("%#v %#v", name, help));
			res = self.keep
		},
		func(name ShortLong, metavar Metavar, _ Option[string], help Option[Doc]) {
			//     self.keep = self
			//         .items
			//         .insert(fmt.Sprintf("%#v %v %#v", name, metavar, help));
			res = self.keep
		},
		func(Meta, HiTy) { dhh() },
		func(HiTy) { dhh() },
	)
	return res
}

func (self Doc) write_help_item_groups(items HelpItems, include_env bool) {
	for {
		rangee, ok := items.find_group().Unpack()
		if !ok {
			break
		}

		dd := Dedupdefault()
		drain := items.items[rangee.start():rangee.end()]
		items.items = append(items.items[rangee.start():], items.items[rangee.end():]...)
		for _, item := range drain {
			if dd.check(item) {
				write_help_item(&self, item, include_env)
			}
		}
	}

	for _, tyname := range []struct {
		HiTy HiTy
		name string
	}{
		{HiTyPositional, "Available positional items:"},
		{HiTyFlag, "Available fun.Options:"},
		{HiTyCommand, "Available commands:"},
	} {
		ty, name := tyname.HiTy, tyname.name
		self.write_help_items(items, ty, name, include_env)
	}
}

func (self Doc) write_help_items(items HelpItems, ty HiTy, name string, include_env bool) {
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
				write_help_item(&self, item, include_env)
			}
		}
		self.token(TokenBlockEnd(BlockDefinitionList))
		self.token(TokenBlockEnd(BlockBlock))
	}
}

func (self Doc) write_path(path []string) {
	for _, item := range path {
		self.write_str(item, StyleLiteral)
		self.write_char(' ', StyleText)
	}
}
