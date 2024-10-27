package cli

import "github.com/rprtr258/fun"

// Consume an arbitrary value that satisfies a condition, created with [any], implements [anywhere](ParseAnyanywhere).
type ParseAny[T any] struct {
	metavar  Doc
	help     Option[Doc]
	check    func(string) (T, bool)
	anywhere bool
}

func (self ParseAny[T]) Parser() Parser[T] {
	return Parser[T]{self.eval, MetaItem(self.item())}
}

func (self ParseAny[T]) item() Item {
	return ItemAny(
		self.metavar,
		self.anywhere,
		self.help,
	)
}

// Add a help message to [any] parser.
// See examples in [any]
func (self *ParseAny[T]) Help(help string) *ParseAny[T] {
	self.help = Valid(newDocFromString(help))
	return self
}

// Replace metavar with a custom value
// See examples in [any]
func (self *ParseAny[T]) Metavar(metavar Doc) *ParseAny[T] {
	self.metavar = metavar
	return self
}

// Try to apply the parser to each unconsumed element instead of just the front one
//
// By default `any` tries to parse just the front unconsumed item behaving similar to
// [positional] parser, `anywhere` changes it so it applies to every unconsumed item,
// similar to argument parser.
//
// See examples in [any]
func (self *ParseAny[T]) Anywhere() *ParseAny[T] {
	self.anywhere = true
	return self
}

func (self ParseAny[T]) eval(args *State) (T, Error) {
	for ix, x := range push2(args.items_iter()) {
		var next bool
		x.value(
			func(_ byte, c bool) { next = c },
			func(_ string, c bool) { next = c },
			nil, nil, nil,
		)
		os := x.str
		if i, ok := self.check(os); ok {
			args.remove(ix)
			if next {
				args.remove(ix + 1)
			}

			return i, nil
		}
		if !self.anywhere {
			break
		}
	}
	missing_item := MissingItem{
		item:     self.item(),
		position: args.scope.start(),
		scope:    args.scope,
	}
	return *new(T), Error(MessageMissing(missing_item))
}

// Parse a single arbitrary item from a command line
//
// **`any` is designed to consume items that don't fit into the usual [flag](NamedArgflag)
// /[switch](NamedArgswitch)/[argument](NamedArgargument)/[positional]/
// [command](OptionParsercommand) classification, in most cases you don't need to use it**
//
// By default, `any` behaves similarly to [positional] so you should be using it near the
// rightmost end of the consumer struct and it will only try to parse the first unconsumed item
// on the command line. It is possible to lift this restriction by calling
// [anywhere](ParseAnyanywhere) on the parser.
//
// `check` argument is a function from any type `I` that implements `FromStr` to `T`.
// Usually this should be `String` or `string`, but feel free to experiment. When
// running `any` tries to parse an item on a command line into that `I` and applies the `check`
// function. If the `check` succeeds - parser `any` succeeds and produces `T`, otherwise it behaves
// as if it hasn't seen it. If `any` works in `anywhere` mode - it will try to parse all other
// unconsumed items, otherwise, `any` fails.
//
// # Use `any` to capture the remaining arguments
// Normally you would use [positional] with [strict](ParsePositionalstrict) annotation for
// that, but using any allows you to blur the boundary between arguments for child process and self
// process a bit more.
// #[cfg_attr(not(doctest), doc = include_str!("docs2/any_simple.md"))]
//
// # Use `any` to parse a non standard flag
//
// #[cfg_attr(not(doctest), doc = include_str!("docs2/any_switch.md"))]
//
// # Use `any` to parse a non standard argument
// Normally `any` would try to display itself as a usual metavariable in the usage line and
// generated help, you can customize that with [metavar](ParseAnymetavar) method:
//
// #[cfg_attr(not(doctest), doc = include_str!("docs2/any_literal.md"))]
//
// # See also
// [literal] - a specialized version of `any` that tries to parse a fixed literal
func Any[T any](metavar string, check func(string) (T, bool)) *ParseAny[T] {
	return &ParseAny[T]{
		metavar: NewDocFromPairs([]fun.Pair[string, Style]{
			{metavar, StyleMetavar},
		}),
		help:     Invalid[Doc](),
		check:    check,
		anywhere: false,
	}
}

// A specialized version of [any] that consumes an arbitrary string
//
// By default `Literal` behaves similarly to [positional] so you should be using it near the
// rightmost end of the consumer struct and it will only try to parse the first unconsumed
// item on the command line. It is possible to lift this restriction by calling
// [anywhere](ParseAnyanywhere) on the parser.
//
// #[cfg_attr(not(doctest), doc = include_str!("docs2/any_literal.md"))]
//
// # See also
// [any] - a generic version of `Literal` that uses function to decide if value is to be parsed
// or not.
func Literal(val string) Parser[struct{}] {
	p := Any("", func(s string) (struct{}, bool) {
		return struct{}{}, s == val
	}).Metavar(NewDocFromPairs([]fun.Pair[string, Style]{
		{val, StyleLiteral},
	}))
	return Parser[struct{}]{p.eval, MetaItem(p.item())}
}
