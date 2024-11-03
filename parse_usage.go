package cli

// Parser that hides inner parser from usage line
//
// No other changes to the inner parser
type ParseUsage[T any] struct {
	inner Parser[T]
	usage Doc
}

func (self ParseUsage[T]) eval(args *State) (T, Error) {
	return self.inner.eval(args)
}

func (self ParseUsage[P]) meta() Meta {
	return MetaCustomUsage{self.inner.meta, self.usage}
}

// Ignore this parser when generating a usage line
//
// Parsers hidden from usage will still show up in the available arguments list. Best used on
// fun.Optional things that augment the main application functionality but not define it.
// Alternatively, you can use [custom_usage](Parsercustom_usage) to replace a single
// fun.Option or a group of them with some other text.
// #[cfg_attr(not(doctest), doc = include_str!("docs2/HideUsage.md"))]
func (self Parser[T]) HideUsage() ParseUsage[T] {
	return ParseUsage[T]{
		inner: self,
		usage: Doc{},
	}
}

// Customize how this parser looks like in the usage line
//
// #[cfg_attr(not(doctest), doc = include_str!("docs2/CustomUsage.md"))]
func (self Parser[T]) CustomUsage(usage Doc) ParseUsage[T] {
	return ParseUsage[T]{
		inner: self,
		usage: usage,
	}
}
