package cli

// Parser that hides inner parser from usage line
//
// No other changes to the inner parser.

// Ignore this parser when generating a usage line
//
// Parsers hidden from usage will still show up in the available arguments list. Best used on
// optional things that augment the main application functionality but don't define it.
// Alternatively, you can use [Parser.CustomUsage] to replace a single
// parser or a group of them with some other text.
func (self Parser[T]) HideUsage() Parser[T] {
	return Parser[T]{self.eval, metaCustomUsage{self.meta, Doc{}}}
}

// Customize how this parser looks like in the usage line
func (self Parser[T]) CustomUsage(usage Doc) Parser[T] {
	return Parser[T]{self.eval, metaCustomUsage{self.meta, usage}}
}
