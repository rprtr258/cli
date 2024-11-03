package cli

import (
	"cmp"
	"fmt"
	"os"
	"slices"
)

type Range [2]int

func (r Range) start() int           { return r[0] }
func (r Range) end() int             { return r[1] }
func (r Range) contains(ix int) bool { return r[0] <= ix && ix < r[1] }
func (r Range) String() string       { return fmt.Sprintf("[%d, %d]", r[0], r[1]) }

// Missing item in a context
type MissingItem struct {
	// Item that is missing
	item Item
	// Position it is missing from - exact for positionals, earliest possible for flags
	position int
	// Range where search was performed, important for combinators that narrow the search scope such as adjacent
	scope Range
}

func (mi MissingItem) String() string {
	return fmt.Sprintf(
		"MissingItem{item: %s, position: %d, scope: %s}",
		mi.item.String(), mi.position, mi.scope.String(),
	)
}

// Unsuccessful command line parsing outcome, use it for unit tests
//
// When [OptionParserrun_inner](crateOptionParserrun_inner) produces `Err(ParseFailure)`
// it means that the parser couldn't produce the value it supposed to produce and the program
// should terminate.
//
// If you are handling variants manually - `Stdout` contains formatted output and you can use any
// logging framework to produce the output, `Completion` should be printed to stdout unchanged -
// shell completion mechanism relies on that. In both cases application should exit with error
// code of 0. `Stderr` variant indicates a genuinly parsing error which should be printed to
// stderr or a logging framework of your choice as an error and the app should exit with error
// code of 1. [ParseFailureexit_code] is a helper method that performs printing and produces
// the exit code to use.
//
// For purposes of for unit testing for user parsers, you can consume it with
// [ParseFailureunwrap_stdout] and [ParseFailureunwrap_stdout] - both of which produce a
// an unformatted `String` that parser might produce if failure type is correct or panics
// otherwise.
type ParseFailure interface {
	isParseFailure()
}

// Print this to stdout and exit with success code
type ParseFailureStdout struct {
	msg  Doc
	full bool
}

func (ParseFailureStdout) isParseFailure() {}

// This also goes to stdout with exit code of 0,
// this cannot be Doc because completion needs more control about rendering
type ParseFailureCompletion string

func (ParseFailureCompletion) isParseFailure() {}

// Print this to stderr and exit with failure code
type ParseFailureStderr Doc

func (ParseFailureStderr) isParseFailure() {}

// TODO: OPTIMIZE: less memory could be used
func dist(s, t string) int {
	if s == "" {
		return len(t)
	}

	if t == "" {
		return len(s)
	}

	w := len(t) + 1
	buf := make([]int, (len(s)+1)*w)

	dists := make([][]int, len(s)+1)
	for i := range dists {
		dists[i] = buf[w*i : w*(i+1)]
		dists[i][0] = i
	}

	for j := range t {
		dists[0][j] = j
	}

	for i, sc := range s {
		for j, tc := range t {
			if sc == tc {
				dists[i+1][j+1] = dists[i][j]
			} else {
				dists[i+1][j+1] = min(
					dists[i][j],
					dists[i+1][j],
					dists[i][j+1],
				) + 1
			}
		}
	}

	return dists[len(s)][len(t)]
}

func getClosest(s string, ss []string) (closest_ string, distMin_ int) {
	if len(ss) == 0 {
		return "", 0
	}

	closest, distMin := "", -1
	for _, t := range ss {
		if distCur := dist(s, t); distMin == -1 || distCur < distMin {
			closest, distMin = t, distCur
		}
	}
	return closest, distMin
}

func can_catch(self Error) bool {
	switch self.(type) {
	case
		ErrorNoEnv,
		ErrorParseSome,
		ErrorParseFail,
		ErrorPureFailed,
		ErrorMissing,
		ErrorNonStrictPos:
		return true
	}
	return false
}

func combine_with(self, other Error) Error {
	// help output takes priority
	if _, ok := self.(ErrorParseFailure); ok {
		return self
	}
	if _, ok := other.(ErrorParseFailure); ok {
		return other
	}

	// combine missing elements
	a, oka := self.(ErrorMissing)
	b, okb := other.(ErrorMissing)
	if oka && okb {
		return ErrorMissing(append(a, b...))
	}

	// otherwise earliest wins
	if can_catch(self) {
		return other
	} else {
		return self
	}
}

// Returns the contained `stderr` values - for unit tests
//
// # Panics
//
// Panics if failure contains `stdout`
func unwrap_stderr(self ParseFailure) string {
	switch self := self.(type) {
	case ParseFailureStdout:
		panic(fmt.Sprintf("not an stderr: %#v", self))
	case ParseFailureCompletion:
		panic(fmt.Sprintf("not an stderr: %#v", self))
	case ParseFailureStderr:
		return Doc(self).monochrome(true)
	}
	panic("unreachable")
}

// Returns the contained `stdout` values - for unit tests
//
// # Panics
//
// Panics if failure contains `stderr`
func unwrap_stdout(self ParseFailure) string {
	switch self := self.(type) {
	case ParseFailureStdout:
		return self.msg.monochrome(self.full)
	case ParseFailureCompletion:
		return string(self)
	case ParseFailureStderr:
		panic(fmt.Sprintf("not an stdout: %#v", self))
	}
	panic("unreachable")
}

// Returns the exit code for the failure
func exit_code(self ParseFailure) int {
	if _, ok := self.(ParseFailureStderr); ok {
		return 1
	}
	return 0
}

// Prints a message to `stdout` or `stderr` appropriate to the failure.
func print_message(self ParseFailure, max_width uint) {
	color := Colordefault
	switch self := self.(type) {
	case ParseFailureStdout:
		fmt.Println(self.msg.render_console(self.full, color, max_width))
	case ParseFailureCompletion:
		fmt.Print(string(self))
	case ParseFailureStderr:
		fmt.Fprintln(os.Stderr, color.push_str(StyleInvalid, "Error: ")+Doc(self).render_console(true, color, max_width))
	}
}

func check_conflicts(args State) (Error, bool) {
	loserwinner, ok := args.conflict()
	if !ok {
		return nil, false
	}
	loser, winner := loserwinner.start(), loserwinner.end()
	return ErrorConflict{winner, loser}, true
}

func textual_part(args State, ixxx Option[int]) (string, bool) {
	ixx, ok := ixxx.Unpack()
	if !ok {
		return "", false
	}

	ix, ok := get(args.items, ixx).Unpack()
	if !ok {
		return "", false
	}

	res, ok := "", false
	ix.value(
		nil,
		nil,
		func() { res, ok = ix.str, true },
		func() { res, ok = ix.str, true },
		func() { res, ok = ix.str, true },
	)
	return res, ok
}

func only_once(args State, cur int) (int, bool) {
	if cur == 0 {
		return 0, false
	}

	iter := rev(args.items[:cur])
	v, ok := get(args.items, cur).Unpack()
	if !ok {
		return 0, false
	}

	var (
		offset int
		okk    bool
	)
	v.value(
		func(s byte, _ bool) { offset, okk = position(iter, func(a Arg) bool { return a.match_short(s) }) },
		func(l string, _ bool) { offset, okk = position(iter, func(a Arg) bool { return a.match_long(l) }) },
		nil, nil, nil,
	)
	if !okk {
		return 0, false
	}
	return cur - offset - 1, true
}

func render(self *Error, args State, meta Meta) ParseFailure {
	// try to come up with a better error message for a few cases
	switch e := (*self).(type) {
	case ErrorMissing:
		*self = summarize_missing(e, meta, args)
	case ErrorUnconsumed:
		ix := int(e)
		if conflict, ok := check_conflicts(args); ok {
			*self = conflict
		} else if prev_ix, ok := only_once(args, ix); ok {
			*self = ErrorOnlyOnce{prev_ix, ix}
		} else if ix, suggestion, ok := suggest(args, meta); ok {
			*self = ErrorSuggestion{ix, suggestion}
		}
	}

	doc := Doc{}
	switch e := (*self).(type) {
	case ErrorNoEnv:
		name := string(e)
		// Error: environment variable FOO is not set
		doc.Text("environment variable ")
		doc.token(TokenBlockStart(BlockTermRef))
		doc.Invalid(name)
		doc.token(TokenBlockEnd(BlockTermRef))
		doc.Text(" is not set")
	case ErrorParseSome:
		s := string(e)
		// Error: <message from some or fail>
		doc.Text(s)
	case ErrorParseFail:
		s := string(e)
		// Error: <message from some or fail>
		doc.Text(s)
	case ErrorPureFailed:
		s := string(e)
		// Error: <message from pure_with>
		doc.Text(s)
	case ErrorMissing:
		// this case is handled above(?)
		// this one is unreachable
		// already rendered
	case ErrorParseFailure:
		return e.e
	case ErrorStrictPos:
		// Error: FOO expected to be  in the right side of --
		doc.Text("expected ")
		doc.token(TokenBlockStart(BlockTermRef))
		doc.metavar(e.metavar)
		doc.token(TokenBlockEnd(BlockTermRef))
		doc.Text(" to be on the right side of ")
		doc.token(TokenBlockStart(BlockTermRef))
		doc.Literal("--")
		doc.token(TokenBlockEnd(BlockTermRef))
	case ErrorNonStrictPos:
		// Error: FOO expected to be on the left side of --
		doc.Text("expected ")
		doc.token(TokenBlockStart(BlockTermRef))
		doc.metavar(e.metavar)
		doc.token(TokenBlockEnd(BlockTermRef))
		doc.Text(" to be on the left side of ")
		doc.token(TokenBlockStart(BlockTermRef))
		doc.Literal("--")
		doc.token(TokenBlockEnd(BlockTermRef))
	case ErrorParseFailed:
		mix, s := e.a, e.s
		// Error: couldn't parse FIELD: <FromStr message>
		doc.Text("couldn't parse")
		if field, ok := textual_part(args, Option[int]{int(mix.Value), mix.Valid}); ok {
			doc.Text(" ")
			doc.token(TokenBlockStart(BlockTermRef))
			doc.Invalid(field)
			doc.token(TokenBlockEnd(BlockTermRef))
		}
		doc.Text(": ")
		doc.Text(s)
	case ErrorGuardFailed:
		mix, s := e.a, e.s
		// Error: ( FIELD:  | check failed: ) <message from guard>
		if field, ok := textual_part(args, Option[int]{int(mix.Value), mix.Valid}); ok {
			doc.token(TokenBlockStart(BlockTermRef))
			doc.Invalid(field)
			doc.token(TokenBlockEnd(BlockTermRef))
			doc.Text(": ")
		} else {
			doc.Text("check failed: ")
		}
		doc.Text(s)
	case ErrorNoArgument:
		x, mv := e.ix, e.metavar
		// Error: --foo requires an argument FOO, got a flag --bar, try --foo=-bar to use it as an argument
		// Error: --foo requires an argument FOO
		a, ok := args.get(x + 1)
		if ok && (a.IsShort() || a.IsLong()) {
			arg := &args.items[x]

			doc.token(TokenBlockStart(BlockTermRef))
			doc.write(arg.GoString(), StyleLiteral)
			doc.token(TokenBlockEnd(BlockTermRef))
			doc.Text(" requires an argument ")
			doc.token(TokenBlockStart(BlockTermRef))
			doc.metavar(mv)
			doc.token(TokenBlockEnd(BlockTermRef))
			doc.Text(", got a flag ")
			doc.token(TokenBlockStart(BlockTermRef))
			doc.write(a.str, StyleInvalid)
			doc.token(TokenBlockEnd(BlockTermRef))
			doc.Text(", try ")
			doc.token(TokenBlockStart(BlockTermRef))
			doc.write(arg.GoString(), StyleLiteral)
			doc.Literal("=")
			doc.write(a.str, StyleLiteral)
			doc.token(TokenBlockEnd(BlockTermRef))
			doc.Text(" to use it as an argument")
		} else {
			// "Some" part of this branch is actually unreachable
			doc.token(TokenBlockStart(BlockTermRef))
			doc.write(args.items[x].GoString(), StyleLiteral)
			doc.token(TokenBlockEnd(BlockTermRef))
			doc.Text(" requires an argument ")
			doc.token(TokenBlockStart(BlockTermRef))
			doc.metavar(mv)
			doc.token(TokenBlockEnd(BlockTermRef))
		}
	case ErrorUnconsumed:
		ix := int(e)
		// Error: --foo is not expected in this context
		doc.token(TokenBlockStart(BlockTermRef))
		doc.write(args.items[ix].GoString(), StyleInvalid)
		doc.token(TokenBlockEnd(BlockTermRef))
		doc.Text(" is not expected in this context")
	case ErrorAmbiguity:
		name, ix := e.name, e.ix
		// Error: app supports -f as both an fun.Option and an fun.Option-argument, try to split -foo
		// into invididual fun.Options (-f -o ..) or use -f=oo syntax to disambiguate
		first := name[0]
		second := name[1]
		rest := name[2:]
		s := args.items[ix].str

		if name, ok := First(args.path).Unpack(); ok {
			doc.Literal(name)
			doc.Text(" supports ")
		} else {
			doc.Text("app supports ")
		}

		doc.token(TokenBlockStart(BlockTermRef))
		doc.Literal("-")
		doc.write_char(first, StyleLiteral)
		doc.token(TokenBlockEnd(BlockTermRef))
		doc.Text(" as both an fun.Option and an fun.Option-argument, try to split ")
		doc.token(TokenBlockStart(BlockTermRef))
		doc.write(s, StyleLiteral)
		doc.token(TokenBlockEnd(BlockTermRef))
		doc.Text(" into individual fun.Options (")
		doc.Literal("-")
		doc.write_char(first, StyleLiteral)
		doc.Literal(" -")
		doc.write_char(second, StyleLiteral)
		doc.Literal(" ..")
		doc.Text(") or use ")
		doc.token(TokenBlockStart(BlockTermRef))
		doc.Literal("-")
		doc.write_char(first, StyleLiteral)
		doc.Literal("=")
		doc.Literal(rest)
		doc.token(TokenBlockEnd(BlockTermRef))
		doc.Text(" syntax to disambiguate")
	case ErrorSuggestion:
		ix, suggestion := e.ix, e.suggestion
		// Error: No such (flag|argument|command), did you mean  ...
		actual := args.items[ix].GoString()
		suggestion(
			func(v Variant) {
				var ty string
				if starts_withb(actual, '-') {
					ty = "flag"
				} else {
					args.items[ix].value(
						func(byte, bool) { ty = "flag" },
						func(string, bool) { ty = "flag" },
						func() { ty = "argument value" },
						func() { ty = "command or positional" },
						func() { ty = "command or positional" },
					)
				}

				doc.Text("no such ")
				doc.Text(ty)
				doc.Text(": ")
				doc.token(TokenBlockStart(BlockTermRef))
				doc.Invalid(actual)
				doc.token(TokenBlockEnd(BlockTermRef))
				doc.Text(", did you mean ")
				doc.token(TokenBlockStart(BlockTermRef))

				v(
					func(name string) {
						doc.Literal(name)
					},
					func(f ShortLong) {
						if f.is_short() {
							doc.Literal("-")
							doc.write_char(f.Short, StyleLiteral)
						} else {
							doc.Literal("--")
							doc.Literal(f.Long)
						}
					},
				)

				doc.token(TokenBlockEnd(BlockTermRef))
				doc.Text("?")
			},
			func(name string) {
				doc.Text("no such flag: ")
				doc.token(TokenBlockStart(BlockTermRef))
				doc.Literal("-")
				doc.Literal(name)
				doc.token(TokenBlockEnd(BlockTermRef))
				doc.Text(" (with one dash), did you mean ")
				doc.token(TokenBlockStart(BlockTermRef))
				doc.Literal("--")
				doc.Literal(name)
				doc.token(TokenBlockEnd(BlockTermRef))
				doc.Text("?")
			},
			func(name byte) {
				doc.Text("no such flag: ")
				doc.token(TokenBlockStart(BlockTermRef))
				doc.Literal("--")
				doc.write_char(name, StyleLiteral)
				doc.token(TokenBlockEnd(BlockTermRef))
				doc.Text(" (with two dashes), did you mean ")
				doc.token(TokenBlockStart(BlockTermRef))
				doc.Literal("-")
				doc.write_char(name, StyleLiteral)
				doc.token(TokenBlockEnd(BlockTermRef))
				doc.Text("?")
			},
			func(x string, v Variant) {
				var ty string
				v(
					func(string) { ty = "subcommand" },
					func(ShortLong) { ty = "flag" },
				)
				doc.Text(ty)
				doc.Text(" ")
				doc.token(TokenBlockStart(BlockTermRef))
				doc.Literal(actual)
				doc.token(TokenBlockEnd(BlockTermRef))
				doc.Text(
					" is not valid in this context, did you mean to pass it to command ",
				)
				doc.token(TokenBlockStart(BlockTermRef))
				doc.Literal(x)
				doc.token(TokenBlockEnd(BlockTermRef))
				doc.Text("?")
			},
		)
	case ErrorConflict:
		winner, loser := e.winner, e.loser
		// Error: --intel cannot be used at the same time as --att
		doc.token(TokenBlockStart(BlockTermRef))
		doc.write(args.items[loser].GoString(), StyleLiteral)
		doc.token(TokenBlockEnd(BlockTermRef))
		doc.Text(" cannot be used at the same time as ")
		doc.token(TokenBlockStart(BlockTermRef))
		doc.write(args.items[winner].GoString(), StyleLiteral)
		doc.token(TokenBlockEnd(BlockTermRef))
	case ErrorExpected:
		exp, actual := e.expected, e.actual
		// Error: Expected (no arguments|--foo), got ..., pass --help
		doc.Text("expected ")
		switch len(exp) {
		case 0:
			doc.Text("no arguments")
		case 1:
			doc.token(TokenBlockStart(BlockTermRef))
			doc.write_item(exp[0])
			doc.token(TokenBlockEnd(BlockTermRef))
		case 2:
			doc.token(TokenBlockStart(BlockTermRef))
			doc.write_item(exp[0])
			doc.token(TokenBlockEnd(BlockTermRef))
			doc.Text(" or ")
			doc.token(TokenBlockStart(BlockTermRef))
			doc.write_item(exp[1])
			doc.token(TokenBlockEnd(BlockTermRef))
		default:
			doc.token(TokenBlockStart(BlockTermRef))
			doc.write_item(exp[0])
			doc.token(TokenBlockEnd(BlockTermRef))
			doc.Text(", ")
			doc.token(TokenBlockStart(BlockTermRef))
			doc.write_item(exp[1])
			doc.token(TokenBlockEnd(BlockTermRef))
			doc.Text(", or more")
		}

		if actual, ok := actual.Unpack(); ok {
			doc.Text(", got ")
			doc.token(TokenBlockStart(BlockTermRef))
			doc.write(args.items[actual].GoString(), StyleInvalid)
			doc.token(TokenBlockEnd(BlockTermRef))
			doc.Text(". Pass ")
		} else {
			doc.Text(", pass ")
		}
		doc.token(TokenBlockStart(BlockTermRef))
		doc.Literal("--help")
		doc.token(TokenBlockEnd(BlockTermRef))
		doc.Text(" for usage information")
	case ErrorOnlyOnce:
		loser := e.loser
		// Error: argument FOO cannot be used multiple times in this context
		doc.Text("argument ")
		doc.token(TokenBlockStart(BlockTermRef))
		doc.write(args.items[loser].GoString(), StyleLiteral)
		doc.token(TokenBlockEnd(BlockTermRef))
		doc.Text(" cannot be used multiple times in this context")
	}
	return ParseFailureStderr(doc)
}

// go over all the missing items, pick the left most scope
func summarize_missing(items []MissingItem, inner Meta, args State) Error {
	// missing items can belong to different scopes, pick the best scope to work with
	best_item, ok := max_by_key(
		slices.Values(items),
		func(a, b MissingItem) int {
			aa, ab := a.position, a.scope.start()
			ba, bb := b.position, b.scope.start()
			return cmp.Or(
				cmp.Compare(aa, ba),
				cmp.Compare(ab, bb),
			)
		},
	)
	if !ok {
		return ErrorParseFail("parser requires an extra flag, argument or parameter, but its name is hidden by the author")
	}

	best_scope := best_item.scope

	saw_command := false
	expected := slices.Collect(filter_map(slices.Values(items), func(i MissingItem) (Item, bool) {
		var cmd bool
		i.item.v(
			nil,
			nil,
			func(string, Option[byte], Meta, Info) { cmd = true },
			nil,
			nil,
		)
		if i.scope == best_scope && !(saw_command && cmd) {
			saw_command = saw_command || cmd
			return i.item, true
		} else {
			return Item{}, false
		}
	}))

	best_scope[0] = max(best_scope.start(), best_item.position)
	args.set_scope(best_scope)
	next := args.items_iter()
	if _, _, ok := next(); ok {
		if ix, sugg, ok := suggest(args, inner); ok {
			return ErrorSuggestion{ix, sugg}
		} else {
			return ErrorExpected{expected, Valid(ix)}
		}
	} else {
		return ErrorExpected{expected, Invalid[int]()}
	}
}
