package cli

import (
	"cmp"
	"fmt"
	"slices"
	"strings"

	"github.com/rprtr258/fun"
)

// Unsuccessful command line parsing outcome, internal representation
type parseError interface {
	isError()
}

// those can be caught ---------------------------------------------------------------
// Tried to consume an env variable with no fallback, variable was not set
type errorNoEnv string

func (errorNoEnv) isError() {}

// User specified an error message on some
type errorParseSome string

func (errorParseSome) isError() {}

// User asked for parser to fail explicitly
type errorParseFail string

func (errorParseFail) isError() {}

// pure_with failed to parse a value
type errorPureFailed string

func (errorPureFailed) isError() {}

// Expected one of those values
//
// Used internally to generate better error messages
type errorMissing []missingItem

func (errorMissing) isError() {}

// those cannot be caught-------------------------------------------------------------
// Parsing failed and this is the final output
type errorParseFailure struct{ e parseFailure }

func (errorParseFailure) isError() {}

// Tried to consume a strict positional argument, value was present but was not strictly positional
type errorStrictPos struct {
	a       int
	metavar Metavar
}

func (errorStrictPos) isError() {}

// Tried to consume a non-strict positional argument, but the value was strict
type errorNonStrictPos struct {
	a       int
	metavar Metavar
}

func (errorNonStrictPos) isError() {}

// Parser provided by user failed to parse a value
type errorParseFailed struct {
	pos Opt[int]
	msg string
}

func (errorParseFailed) isError() {}

// Parser provided by user failed to validate a value
type errorGuardFailed struct {
	pos Opt[int]
	msg string
}

func (errorGuardFailed) isError() {}

// Argument requres a value but something else was passed,
// required: --foo <BAR>
// given: --foo --bar
//
//	--foo -- bar
//	--foo
type errorNoArgument struct {
	ix      int
	metavar Metavar
}

func (errorNoArgument) isError() {}

// Parser is expected to consume all the things from the command line
// this item will contain an index of the unconsumed value
type errorUnconsumed int

func (errorUnconsumed) isError() {}

// argument is ambigoups - parser can accept it as both a set of flags and a short flag with no =
type errorAmbiguity struct {
	ix   int
	name string
}

func (errorAmbiguity) isError() {}

// Suggested fixes for typos or missing input
type errorSuggestion struct {
	ix         int
	suggestion suggestion
}

func (errorSuggestion) isError() {}

// Two arguments are mutually exclusive
// --release --dev
type errorConflict struct{ winner, loser int }

func (errorConflict) isError() {}

// Expected one or more items in the scope, got someting else if any
type errorExpected struct {
	expected []item
	actual   Opt[int]
}

func (errorExpected) isError() {}

// Parameter is accepted but only once
type errorOnlyOnce struct{ winner, loser int }

func (errorOnlyOnce) isError() {}

func errorToString(self parseError) string {
	switch self := self.(type) {
	case errorNoEnv:
		return fmt.Sprintf("NoEnv(%#v)", self)
	case errorParseSome:
		return fmt.Sprintf("ParseSome(%#v)", self)
	case errorParseFail:
		return string(self)
	case errorPureFailed:
		return string(self)
	case errorMissing:
		return fmt.Sprintf("Missing(%s)", strings.Join(fun.Map[string]((missingItem).String, self...), ", "))
	case errorParseFailure:
		return fmt.Sprintf("parseFailure(%#v)", self.e)
	case errorStrictPos:
		return fmt.Sprintf("StrictPos(%#v, %#v)", self.a, self.metavar)
	case errorNonStrictPos:
		return fmt.Sprintf("NonStrictPos(%#v, %#v)", self.a, self.metavar)
	case errorParseFailed:
		return fmt.Sprintf("ParseFailed(%#v, %#v)", self.pos, self.msg)
	case errorGuardFailed:
		return fmt.Sprintf("GuardFailed(%#v, %#v)", self.pos, self.msg)
	case errorNoArgument:
		return fmt.Sprintf("NoArgument(%#v, %#v)", self.ix, self.metavar)
	case errorUnconsumed:
		return fmt.Sprintf("Unconsumed(%#v)", int(self))
	case errorAmbiguity:
		return fmt.Sprintf("Ambiguity(%#v, %#v)", self.ix, self.name)
	case errorSuggestion:
		return fmt.Sprintf("suggestion(%#v, %#v)", self.ix, self.suggestion)
	case errorConflict:
		return fmt.Sprintf("Conflict(%#v, %#v)", self.winner, self.loser)
	case errorExpected:
		return fmt.Sprintf("Expected(%#v, %#v)", self.expected, self.actual)
	case errorOnlyOnce:
		return fmt.Sprintf("OnlyOnce(%#v, %#v)", self.winner, self.loser)
	}
	panic("unerasable")
}

type rng [2]int

func (r rng) start() int           { return r[0] }
func (r rng) end() int             { return r[1] }
func (r rng) contains(ix int) bool { return r[0] <= ix && ix < r[1] }
func (r rng) String() string       { return fmt.Sprintf("[%d, %d]", r[0], r[1]) }

// Missing item in a context
type missingItem struct {
	// item that is missing
	item item
	// posKind it is missing from - exact for positionals, earliest possible for flags
	position int
	// rng where search was performed, important for combinators that narrow the search scope such as adjacent
	scope rng
}

func (mi missingItem) String() string {
	return fmt.Sprintf(
		"missingItem{item: %s, position: %d, scope: %s}",
		mi.item.String(), mi.position, mi.scope.String(),
	)
}

// Unsuccessful command line parsing outcome, use it for unit tests
//
// When the internal run helper produces a failure it means that the parser couldn't
// produce the value it was supposed to produce and the program should terminate.
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
type parseFailure interface {
	isParseFailure()
}

// Print this to stdout and exit with success code
type parseFailureStdout struct {
	msg  Doc
	full bool
}

func (parseFailureStdout) isParseFailure() {}

// This also goes to stdout with exit code of 0,
// this cannot be Doc because completion needs more control about rendering
type parseFailureCompletion string

func (parseFailureCompletion) isParseFailure() {}

// Print this to stderr and exit with failure code
type parseFailureStderr Doc

func (parseFailureStderr) isParseFailure() {}

func can_catch(self parseError) bool {
	switch self.(type) {
	case
		errorNoEnv,
		errorParseSome,
		errorParseFail,
		errorPureFailed,
		errorMissing,
		errorNonStrictPos:
		return true
	}
	return false
}

func combine_with(self, other parseError) parseError {
	// help output takes priority
	if _, ok := self.(errorParseFailure); ok {
		return self
	}
	if _, ok := other.(errorParseFailure); ok {
		return other
	}

	// combine missing elements
	a, oka := self.(errorMissing)
	b, okb := other.(errorMissing)
	if oka && okb {
		return errorMissing(append(a, b...))
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
func unwrap_stderr(self parseFailure) string {
	switch self := self.(type) {
	case parseFailureStdout:
		panic(fmt.Sprintf("not an stderr: %#v", self))
	case parseFailureCompletion:
		panic(fmt.Sprintf("not an stderr: %#v", self))
	case parseFailureStderr:
		return Doc(self).monochrome(true)
	}
	panic(fmt.Sprintf("unreachable %T", self))
}

// Returns the contained `stdout` values - for unit tests
//
// # Panics
//
// Panics if failure contains `stderr`
func unwrap_stdout(self parseFailure) string {
	switch self := self.(type) {
	case parseFailureStdout:
		return self.msg.monochrome(self.full)
	case parseFailureCompletion:
		return string(self)
	case parseFailureStderr:
		panic(fmt.Sprintf("not an stdout: %#v", self))
	}
	panic(fmt.Sprintf("unreachable %T", self))
}

func textual_part(state state, ixxx Opt[int]) (string, bool) {
	ixx, ok := ixxx.Unpack()
	if !ok {
		return "", false
	}

	ix, ok := get(state.items, ixx).Unpack()
	if !ok {
		return "", false
	}

	switch ix.kind {
	case argKindWord, argKindPositional:
		return ix.str, true
	}
	return "", false
}

func only_once(state state, cur int) (int, bool) {
	if cur == 0 {
		return 0, false
	}

	iter := rev(state.items[:cur])
	v, ok := get(state.items, cur).Unpack()
	if !ok {
		return 0, false
	}

	var (
		offset int
		okk    bool
	)
	switch v.kind {
	case argKindShortFlag:
		offset, okk = position(iter, func(a arg) bool { return a.match_short(v.c) })
	case argKindLongFlag:
		offset, okk = position(iter, func(a arg) bool { return a.match_long(v.s) })
	}
	if !okk {
		return 0, false
	}
	return cur - offset - 1, true
}

func render(self *parseError, state state, meta meta) parseFailure {
	// try to come up with a better error message for a few cases
	switch e := (*self).(type) {
	case errorMissing:
		*self = summarize_missing(e, meta, state)
	case errorUnconsumed:
		ix := int(e)

		if loserwinner, ok := state.conflict(); ok {
			loser, winner := loserwinner.start(), loserwinner.end()
			*self = errorConflict{winner, loser}
		} else if prev_ix, ok := only_once(state, ix); ok {
			*self = errorOnlyOnce{prev_ix, ix}
		} else if ix, suggestion, ok := suggest(state, meta); ok {
			*self = errorSuggestion{ix, suggestion}
		}
	}

	doc := Doc{}
	switch e := (*self).(type) {
	case errorNoEnv:
		name := string(e)
		// parseError: environment variable FOO is not set
		doc.Text("environment variable ")
		doc.token(tokenBlockStart(blockTermRef))
		doc.Invalid(name)
		doc.token(tokenBlockEnd(blockTermRef))
		doc.Text(" is not set")
	case errorParseSome:
		s := string(e)
		// parseError: <message from some or fail>
		doc.Text(s)
	case errorParseFail:
		s := string(e)
		// parseError: <message from some or fail>
		doc.Text(s)
	case errorPureFailed:
		s := string(e)
		// parseError: <message from pure_with>
		doc.Text(s)
	case errorMissing:
		// this case is handled above(?)
		// this one is unreachable
		// already rendered
	case errorParseFailure:
		return e.e
	case errorStrictPos:
		// parseError: FOO expected to be  in the right side of --
		doc.Text("expected ")
		doc.token(tokenBlockStart(blockTermRef))
		doc.metavar(e.metavar)
		doc.token(tokenBlockEnd(blockTermRef))
		doc.Text(" to be on the right side of ")
		doc.token(tokenBlockStart(blockTermRef))
		doc.Literal("--")
		doc.token(tokenBlockEnd(blockTermRef))
	case errorNonStrictPos:
		// parseError: FOO expected to be on the left side of --
		doc.Text("expected ")
		doc.token(tokenBlockStart(blockTermRef))
		doc.metavar(e.metavar)
		doc.token(tokenBlockEnd(blockTermRef))
		doc.Text(" to be on the left side of ")
		doc.token(tokenBlockStart(blockTermRef))
		doc.Literal("--")
		doc.token(tokenBlockEnd(blockTermRef))
	case errorParseFailed:
		// parseError: couldn't parse FIELD: <FromStr message>
		doc.Text("couldn't parse")
		if field, ok := textual_part(state, e.pos); ok {
			doc.Text(" ")
			doc.token(tokenBlockStart(blockTermRef))
			doc.Invalid(field)
			doc.token(tokenBlockEnd(blockTermRef))
		}
		doc.Text(": ")
		doc.Text(e.msg)
	case errorGuardFailed:
		// parseError: ( FIELD:  | check failed: ) <message from guard>
		if field, ok := textual_part(state, e.pos); ok {
			doc.token(tokenBlockStart(blockTermRef))
			doc.Invalid(field)
			doc.token(tokenBlockEnd(blockTermRef))
			doc.Text(": ")
		} else {
			doc.Text("check failed: ")
		}
		doc.Text(e.msg)
	case errorNoArgument:
		// parseError: --foo requires an argument FOO, got a flag --bar, try --foo=-bar to use it as an argument
		// parseError: --foo requires an argument FOO
		a, ok := state.get(e.ix + 1)
		if ok && (a.IsShort() || a.IsLong()) {
			arg := &state.items[e.ix]

			doc.token(tokenBlockStart(blockTermRef))
			doc.write(arg.GoString(), StyleLiteral)
			doc.token(tokenBlockEnd(blockTermRef))
			doc.Text(" requires an argument ")
			doc.token(tokenBlockStart(blockTermRef))
			doc.metavar(e.metavar)
			doc.token(tokenBlockEnd(blockTermRef))
			doc.Text(", got a flag ")
			doc.token(tokenBlockStart(blockTermRef))
			doc.write(a.str, StyleInvalid)
			doc.token(tokenBlockEnd(blockTermRef))
			doc.Text(", try ")
			doc.token(tokenBlockStart(blockTermRef))
			doc.write(arg.GoString(), StyleLiteral)
			doc.Literal("=")
			doc.write(a.str, StyleLiteral)
			doc.token(tokenBlockEnd(blockTermRef))
			doc.Text(" to use it as an argument")
		} else {
			// "Some" part of this branch is actually unreachable
			doc.token(tokenBlockStart(blockTermRef))
			doc.write(state.items[e.ix].GoString(), StyleLiteral)
			doc.token(tokenBlockEnd(blockTermRef))
			doc.Text(" requires an argument ")
			doc.token(tokenBlockStart(blockTermRef))
			doc.metavar(e.metavar)
			doc.token(tokenBlockEnd(blockTermRef))
		}
	case errorUnconsumed:
		ix := int(e)
		// parseError: --foo is not expected in this context
		doc.token(tokenBlockStart(blockTermRef))
		doc.write(state.items[ix].GoString(), StyleInvalid)
		doc.token(tokenBlockEnd(blockTermRef))
		doc.Text(" is not expected in this context")
	case errorAmbiguity:
		name, ix := e.name, e.ix
		// parseError: app supports -f as both an option and an argument, try to split -foo
		// into individual options (-f -o ..) or use -f=oo syntax to disambiguate
		c0 := name[0]
		c1 := name[1]
		rest := name[1:]
		s := state.items[ix].str

		if name, ok := first(state.path).Unpack(); ok {
			doc.Literal(name)
			doc.Text(" supports ")
		} else {
			doc.Text("app supports ")
		}

		doc.token(tokenBlockStart(blockTermRef))
		doc.Literal("-")
		doc.write_char(c0, StyleLiteral)
		doc.token(tokenBlockEnd(blockTermRef))
		doc.Text(" as both an option and an option-argument, try to split ")
		doc.token(tokenBlockStart(blockTermRef))
		doc.write(s, StyleLiteral)
		doc.token(tokenBlockEnd(blockTermRef))
		doc.Text(" into individual options (")
		doc.Literal("-")
		doc.write_char(c0, StyleLiteral)
		doc.Literal(" -")
		doc.write_char(c1, StyleLiteral)
		doc.Literal(" ..")
		doc.Text(") or use ")
		doc.token(tokenBlockStart(blockTermRef))
		doc.Literal("-")
		doc.write_char(c0, StyleLiteral)
		doc.Literal("=")
		doc.Literal(rest)
		doc.token(tokenBlockEnd(blockTermRef))
		doc.Text(" syntax to disambiguate")
	case errorSuggestion:
		ix, suggestion := e.ix, e.suggestion
		// parseError: No such (flag|argument|command), did you mean  ...
		actual := state.items[ix].GoString()
		suggestion(
			func(v variant) {
				var ty string
				if starts_withb(actual, '-') {
					ty = "flag"
				} else {
					switch state.items[ix].kind {
					case argKindShortFlag:
						ty = "flag"
					case argKindLongFlag:
						ty = "flag"
					case argKindWord:
						ty = "command or positional"
					case argKindPositional:
						ty = "command or positional"
					}
				}

				doc.Text("no such ")
				doc.Text(ty)
				doc.Text(": ")
				doc.token(tokenBlockStart(blockTermRef))
				doc.Invalid(actual)
				doc.token(tokenBlockEnd(blockTermRef))
				doc.Text(", did you mean ")
				doc.token(tokenBlockStart(blockTermRef))

				v(
					func(name string) {
						doc.Literal(name)
					},
					func(f shortLong) {
						if f.is_short() {
							doc.Literal("-")
							doc.write_char(f.Short, StyleLiteral)
						} else {
							doc.Literal("--")
							doc.Literal(f.Long)
						}
					},
				)

				doc.token(tokenBlockEnd(blockTermRef))
				doc.Text("?")
			},
			func(name string) {
				doc.Text("no such flag: ")
				doc.token(tokenBlockStart(blockTermRef))
				doc.Literal("-")
				doc.Literal(name)
				doc.token(tokenBlockEnd(blockTermRef))
				doc.Text(" (with one dash), did you mean ")
				doc.token(tokenBlockStart(blockTermRef))
				doc.Literal("--")
				doc.Literal(name)
				doc.token(tokenBlockEnd(blockTermRef))
				doc.Text("?")
			},
			func(name byte) {
				doc.Text("no such flag: ")
				doc.token(tokenBlockStart(blockTermRef))
				doc.Literal("--")
				doc.write_char(name, StyleLiteral)
				doc.token(tokenBlockEnd(blockTermRef))
				doc.Text(" (with two dashes), did you mean ")
				doc.token(tokenBlockStart(blockTermRef))
				doc.Literal("-")
				doc.write_char(name, StyleLiteral)
				doc.token(tokenBlockEnd(blockTermRef))
				doc.Text("?")
			},
			func(x string, v variant) {
				var ty string
				v(
					func(string) { ty = "subcommand" },
					func(shortLong) { ty = "flag" },
				)
				doc.Text(ty)
				doc.Text(" ")
				doc.token(tokenBlockStart(blockTermRef))
				doc.Literal(actual)
				doc.token(tokenBlockEnd(blockTermRef))
				doc.Text(
					" is not valid in this context, did you mean to pass it to command ",
				)
				doc.token(tokenBlockStart(blockTermRef))
				doc.Literal(x)
				doc.token(tokenBlockEnd(blockTermRef))
				doc.Text("?")
			},
		)
	case errorConflict:
		winner, loser := e.winner, e.loser
		// parseError: --intel cannot be used at the same time as --att
		doc.token(tokenBlockStart(blockTermRef))
		doc.write(state.items[loser].GoString(), StyleLiteral)
		doc.token(tokenBlockEnd(blockTermRef))
		doc.Text(" cannot be used at the same time as ")
		doc.token(tokenBlockStart(blockTermRef))
		doc.write(state.items[winner].GoString(), StyleLiteral)
		doc.token(tokenBlockEnd(blockTermRef))
	case errorExpected:
		exp, actual := e.expected, e.actual
		// parseError: Expected (no arguments|--foo), got ..., pass --help
		doc.Text("expected ")
		switch len(exp) {
		case 0:
			doc.Text("no arguments")
		case 1:
			doc.token(tokenBlockStart(blockTermRef))
			doc.write_item(exp[0])
			doc.token(tokenBlockEnd(blockTermRef))
		case 2:
			doc.token(tokenBlockStart(blockTermRef))
			doc.write_item(exp[0])
			doc.token(tokenBlockEnd(blockTermRef))
			doc.Text(" or ")
			doc.token(tokenBlockStart(blockTermRef))
			doc.write_item(exp[1])
			doc.token(tokenBlockEnd(blockTermRef))
		default:
			doc.token(tokenBlockStart(blockTermRef))
			doc.write_item(exp[0])
			doc.token(tokenBlockEnd(blockTermRef))
			doc.Text(", ")
			doc.token(tokenBlockStart(blockTermRef))
			doc.write_item(exp[1])
			doc.token(tokenBlockEnd(blockTermRef))
			doc.Text(", or more")
		}

		if actual, ok := actual.Unpack(); ok {
			doc.Text(", got ")
			doc.token(tokenBlockStart(blockTermRef))
			doc.write(state.items[actual].GoString(), StyleInvalid)
			doc.token(tokenBlockEnd(blockTermRef))
			doc.Text(". Pass ")
		} else {
			doc.Text(", pass ")
		}
		doc.token(tokenBlockStart(blockTermRef))
		doc.Literal("--help")
		doc.token(tokenBlockEnd(blockTermRef))
		doc.Text(" for usage information")
	case errorOnlyOnce:
		loser := e.loser
		// parseError: argument FOO cannot be used multiple times in this context
		doc.Text("argument ")
		doc.token(tokenBlockStart(blockTermRef))
		doc.write(state.items[loser].GoString(), StyleLiteral)
		doc.token(tokenBlockEnd(blockTermRef))
		doc.Text(" cannot be used multiple times in this context")
	}
	return parseFailureStderr(doc)
}

// go over all the missing items, pick the left most scope
func summarize_missing(items []missingItem, inner meta, state state) parseError {
	// missing items can belong to different scopes, pick the best scope to work with
	best_item, ok := max_by_key(
		slices.Values(items),
		func(a, b missingItem) int {
			aa, ab := a.position, a.scope.start()
			ba, bb := b.position, b.scope.start()
			return cmp.Or(
				cmp.Compare(aa, ba),
				cmp.Compare(ab, bb),
			)
		},
	)
	if !ok {
		return errorParseFail("parser requires an extra flag, argument or parameter, but its name is hidden by the author")
	}

	best_scope := best_item.scope

	saw_command := false
	expected := slices.Collect(filter_map(slices.Values(items), func(i missingItem) (item, bool) {
		if _, cmd := i.item.v.(itemCommand); i.scope == best_scope && !(saw_command && cmd) {
			saw_command = saw_command || cmd
			return i.item, true
		} else {
			return item{}, false
		}
	}))

	best_scope[0] = max(best_scope.start(), best_item.position)
	state = state.clone()
	state.scope = best_scope
	if ix, _, ok := state.items_iter()(); ok {
		if iy, sugg, ok := suggest(state, inner); ok {
			return errorSuggestion{iy, sugg}
		} else {
			return errorExpected{expected, Valid(ix)}
		}
	}

	return errorExpected{expected, Invalid[int]()}
}
