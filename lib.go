// Package cli provides a lightweight command line argument parser with a
// combinator-style API.
//
// A parser is assembled from small building blocks, composed into larger
// parsers, and finally run against the command line. The value produced by a
// successful parse is a regular Go value - the parser either produces it or
// prints a message describing the problem (missing argument, validation
// failure, --help request, etc).
//
// # Overview
//
// It is best to think of a [Parser[T]] as a container with a value inside.
// Containers are composed with [Pair], [Triple], [Choice] and [Construct];
// the only way to extract the value is to finalize the parser with
// [Parser.ToOptions] and run it with [OptionParser.Run] or
// [OptionParser.RunE].
//
// Values inside can be of any type, as long as they are comparable with ==
// where needed. Parse functions convert raw command line strings to the value
// types, e.g. strconv.ParseUint for numbers. It is better to perform as much
// parsing and validation inside the parser as possible, so the program itself
// gets strictly typed and correct values while the user gets immediate
// feedback on what is wrong with the arguments they pass.
//
// # Quick start
//
// Suppose the program needs the user to specify the dimensions of a rectangle,
// with sides 1..20 units long and the total area not exceeding 200 square
// units:
//
//	type Rectangle struct {
//	    Width, Height int
//	}
//
//	func rectangle() cli.Parser[Rectangle] {
//	    invalidSize := "Sides of a rectangle must be 1..20 units long"
//	    invalidArea := "Area of a rectangle must not exceed 200 units square"
//	    width := cli.Argument[int](cli.Long("width").Help("Width of the rectangle"), "PX", strconv.Atoi).
//	        Parser().
//	        Guard(func(x int) bool { return 1 <= x && x <= 20 }, invalidSize)
//	    height := cli.Argument[int](cli.Long("height").Help("Height of the rectangle"), "PX", strconv.Atoi).
//	        Parser().
//	        Guard(func(x int) bool { return 1 <= x && x <= 20 }, invalidSize)
//	    return cli.Pair(width, height).
//	        Map(func(wh fun.Pair[int, int]) Rectangle {
//	            return Rectangle{Width: wh.K, Height: wh.V}
//	        }).
//	        Guard(func(r Rectangle) bool { return r.Width*r.Height <= 200 }, invalidArea)
//	}
//
//	func main() {
//	    rect := rectangle().ToOptions().Run()
//	    fmt.Println(rect)
//	}
//
// # Consuming items
//
// [Short], [Long] and [Env] build a [NamedArg]. [NamedArg.Switch] makes a boolean
// flag; [Flag] and [ReqFlag] make flags with custom present/absent values;
// [Argument] and [Positional] consume a value, converting the raw string with a
// parse function whose error is reported to the user.
//
// # Transforming parsers
//
// Each combinator returns a new parser without changing the original:
// [Parser.Map] and [Parse] convert values, [Parser.Guard] validates them,
// [Parser.Fallback] and [Parser.FallbackWith] provide defaults,
// [Optional] makes a parser optional, [Many] and [Some] collect repeated
// items, [Collect] fills any slice, [Count] counts occurrences.
//
// # Combining parsers
//
// [Pair] and [Triple] combine a fixed number of parsers, [Construct]
// combines them into a struct, [Choice] and [Parser.OrElse] pick the first
// parser that succeeds, [Parser.Adjacent] restricts a group to adjacent
// command line items, [Any] consumes items at arbitrary positions.
//
// # Improving user experience
//
// [NamedArg.Help] documents flags and arguments, [Parser.Hide] hides them
// from --help, [Parser.GroupHelp] groups related options,
// [OptionParser.Description], [OptionParser.Header], [OptionParser.Footer]
// and [OptionParser.Usage] customize help output, [OptionParser.Version]
// adds --version, [OptionParser.HelpParser] and
// [OptionParser.VersionParser] customize the flags used to trigger help and
// version, [OptionParser.MaxWidth] limits the rendered width.
//
// # Shell completion
//
// Flag and value names are completed out of the box through the complete
// protocol of the shell. Custom dynamic completion is available with
// [Parser.CompleteShell] and [Parser.Complete].
//
// # Running the parser
//
// [OptionParser.Run] parses os.Args and returns the value, exiting the
// process on --help, --version or parse failures. [OptionParser.RunE]
// returns an error instead of exiting, which is useful for tests and
// libraries.

package cli

// Simple or composed argument parser
//
// # Overview
//
// It's best to think of a [Parser[T]] as a container with a value inside that is composable
// with other containers using [Pair], [Triple] or [Construct]; the only way to extract the
// value is by finalizing the parser with [Parser.ToOptions] and running it with
// [OptionParser.Run] or [OptionParser.RunE]. At that point you either get your value out or
// the library generates a message describing a problem (missing argument, validation
// failure, user requested help, etc).
//
// Values inside can be of any type. Parse functions convert raw command line strings into
// the value type - for numbers that's usually a strconv helper. It's better to perform as
// much parsing and validation inside the parser as possible, so the program itself gets
// strictly typed and correct values while the user gets immediate feedback on what's wrong
// with the arguments they pass.
//
// Order of operations matters: each subsequent transformation gets the output of the earlier
// one. Both parsers below collect multiple numeric values below 10, but `a` validates each
// value as it is consumed while `b` first collects and then validates the whole slice. The
// former approach is usually more readable:
//
//	a := Argument[uint32](Short('a'), "N", parseUint32).
//	    Guard(func(a uint32) bool { return a < 10 }, "`a` must be below 10")
//	collectA := Many(a)
//
//	b := Many(Argument[uint](Short('b'), "N", parseUint32))
//	collectB := Parse(b, func(bs []uint32) ([]uint32, error) {
//	    for _, b := range bs {
//	        if b >= 10 {
//	            return nil, errors.New("`b` must be below 10")
//	        }
//	    }
//	    return bs, nil
//	})
//
// For example suppose the program needs the user to specify dimensions of a rectangle, with
// sides being 1..20 units long and the total area must not exceed 200 square units. A parser
// that consumes it might look like this:
//
//	type Rectangle struct {
//	    Width, Height uint32
//	}
//
//	func rectangle() cli.Parser[Rectangle] {
//	    invalidSize := "Sides of a rectangle must be 1..20 units long"
//	    invalidArea := "Area of a rectangle must not exceed 200 units square"
//	    width := cli.Argument[uint32](cli.Long("width"), "PX", parseUint32).
//	        Help("Width of the rectangle").
//	        Guard(func(x uint32) bool { return 1 <= x && x <= 20 }, invalidSize)
//	    height := cli.Argument[uint32](cli.Long("height"), "PX", parseUint32).
//	        Help("Height of the rectangle").
//	        Guard(func(x uint32) bool { return 1 <= x && x <= 20 }, invalidSize)
//	    return cli.Pair(width, height).
//	        Map(func(wh fun.Pair[uint32, uint32]) Rectangle {
//	            return Rectangle{Width: wh.K, Height: wh.V}
//	        }).
//	        Guard(func(r Rectangle) bool { return r.Width*r.Height <= 200 }, invalidArea)
//	}
//
// # Composition with structs
//
// [Construct] builds a parser for a struct type from its fields: each exported field with a
// `cli` tag is parsed according to the tag, producing the value for that field.
type Parser[T any] struct {
	eval func(*state) (T, parseError) // Evaluate inner function
	meta meta                         // Included information about the parser
}

func (self Parser[T]) Parser() Parser[T] {
	return self
}

type parser[T any] interface {
	Parser() Parser[T]
}

// If first parser fails - try the second one
//
// For parser to succeed either of the components needs to succeed. If both succeed the
// result comes from the one that consumed the left most value. The second flag on the
// command line remains unconsumed by [Parser.OrElse].
//
// # Combinatoric usage
// There are two ways to write this combinator with identical results:
//
//	func a() cli.Parser[uint32] {
//	    return cli.Argument[uint32](cli.Short('a'), "NUM", parseUint32).Parser()
//	}
//
//	func b() cli.Parser[uint32] {
//	    return cli.Argument[uint32](cli.Short('b'), "NUM", parseUint32).Parser()
//	}
//
//	func aOrB() cli.Parser[uint32] {
//	    return cli.Choice(a(), b())
//	}
//
// # Example
//
//	$ app -a 12 -b 3
//	12
//	$ app -b 3 -a 12
//	3
//	$ app -b 13
//	13
//	$ app
//	fails asking for either -a NUM or -b NUM
//
// Deprecated: instead of a.OrElse(b) you should use Choice(a, b)
func (self Parser[T]) OrElse(alt Parser[T]) Parser[T] {
	return Parser[T]{
		eval: func(state *state) (T, parseError) {
			return or_else_eval(self, alt, state)
		},
		meta: or(self.meta, alt.meta),
	}
}

// Ignore this parser during any sort of help generation
//
// Best used for optional parsers or parsers with a defined fallback, usually for implementing
// backward compatibility or hidden aliases
func (self Parser[T]) Hide() Parser[T] {
	inner := self
	return Parser[T]{
		eval: func(state *state) (T, parseError) {
			var comps []comp
			state.swap_comps_with(&comps)
			res, err := inner.eval(state)
			state.swap_comps_with(&comps)
			if _, ok := err.(errorMissing); err != nil && ok {
				return *new(T), errorMissing(nil)
			}
			return res, err
		},
		meta: metaSkip{},
	}
}

// Attach a help message to a complex parser
//
// The library inserts the group help message before the block with all the options
// from the inner parser and an empty line after the block.
func (self Parser[T]) GroupHelp(message string) Parser[T] {
	inner := self
	buf := NewDoc(message)
	return Parser[T]{
		eval: func(state *state) (T, parseError) {
			comp_items := []comp(nil)
			state.swap_comps_with(&comp_items)
			res, err := inner.eval(state)
			state.swap_comps_with(&comp_items)
			state.push_with_group(buf.to_completion(), &comp_items)
			return res, err
		},
		meta: metaSubsection{inner.meta, buf},
	}
}

// Make a help message for a complex parser from its inner structure
func (self Parser[T]) withGroupHelp(f func(meta) Doc) Parser[T] {
	inner := self
	return Parser[T]{
		eval: inner.eval,
		meta: metaSubsection{inner.meta, f(inner.meta)},
	}
}

// Transform `Parser` into [OptionParser] to get ready to run it
//
// The returned [OptionParser] carries the presentation and runtime options:
// [OptionParser.Run] or [OptionParser.RunE] parse the command line, and
// [OptionParser.Description], [OptionParser.Usage], [OptionParser.Version]
// and friends customize the help output.
func (self Parser[T]) ToOptions() *OptionParser[T] {
	return &OptionParser[T]{
		info:  newInfo(),
		inner: self,
	}
}
