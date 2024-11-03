package cli

import (
	"fmt"
	"strings"

	"github.com/rprtr258/fun"
)

// Unsuccessful command line parsing outcome, internal representation
type Error interface {
	isError()
}

// those can be caught ---------------------------------------------------------------
// Tried to consume an env variable with no fallback, variable was not set
type ErrorNoEnv string

func (ErrorNoEnv) isError() {}

// User specified an error message on some
type ErrorParseSome string

func (ErrorParseSome) isError() {}

// User asked for parser to fail explicitly
type ErrorParseFail string

func (ErrorParseFail) isError() {}

// pure_with failed to parse a value
type ErrorPureFailed string

func (ErrorPureFailed) isError() {}

// Expected one of those values
//
// Used internally to generate better error messages
type ErrorMissing []MissingItem

func (ErrorMissing) isError() {}

// those cannot be caught-------------------------------------------------------------
// Parsing failed and this is the final output
type ErrorParseFailure struct{ e ParseFailure }

func (ErrorParseFailure) isError() {}

// Tried to consume a strict positional argument, value was present but was not strictly positional
type ErrorStrictPos struct {
	a       int
	metavar Metavar
}

func (ErrorStrictPos) isError() {}

// Tried to consume a non-strict positional argument, but the value was strict
type ErrorNonStrictPos struct {
	a       int
	metavar Metavar
}

func (ErrorNonStrictPos) isError() {}

// Parser provided by user failed to parse a value
type ErrorParseFailed struct {
	a Option[int]
	s string
}

func (ErrorParseFailed) isError() {}

// Parser provided by user failed to validate a value
type ErrorGuardFailed struct {
	a Option[int]
	s string
}

func (ErrorGuardFailed) isError() {}

// Argument requres a value but something else was passed,
// required: --foo <BAR>
// given: --foo --bar
//
//	--foo -- bar
//	--foo
type ErrorNoArgument struct {
	ix      int
	metavar Metavar
}

func (ErrorNoArgument) isError() {}

// Parser is expected to consume all the things from the command line
// this item will contain an index of the unconsumed value
type ErrorUnconsumed /* TODO: unused? */ int

func (ErrorUnconsumed) isError() {}

// argument is ambigoups - parser can accept it as both a set of flags and a short flag with no =
type ErrorAmbiguity struct {
	ix   int
	name string
}

func (ErrorAmbiguity) isError() {}

// Suggested fixes for typos or missing input
type ErrorSuggestion struct {
	ix         int
	suggestion Suggestion
}

func (ErrorSuggestion) isError() {}

// Two arguments are mutually exclusive
// --release --dev
type ErrorConflict struct{ winner, loser int }

func (ErrorConflict) isError() {}

// Expected one or more items in the scope, got someting else if any
type ErrorExpected struct {
	expected []Item
	actual   Option[int]
}

func (ErrorExpected) isError() {}

// Parameter is accepted but only once
type ErrorOnlyOnce struct{ winner, loser int }

func (ErrorOnlyOnce) isError() {}

type MessageFs[T any] struct {
	NoEnv        func(string) T
	ParseSome    func(string) T
	ParseFail    func(string) T
	PureFailed   func(string) T
	Missing      func([]MissingItem) T
	ParseFailure func(ParseFailure) T
	StrictPos    func(int, Metavar) T
	NonStrictPos func(int, Metavar) T
	ParseFailed  func(Option[int], string) T
	GuardFailed  func(Option[int], string) T
	NoArgument   func(int, Metavar) T
	Unconsumed   func(int) T
	Ambiguity    func(int, string) T
	Suggestion   func(int, Suggestion) T
	Conflict     func(int, int) T
	Expected     func([]Item, Option[int]) T
	OnlyOnce     func(int, int) T
}

func MessageF[T any](m Error, fs MessageFs[T]) T {
	switch m := m.(type) {
	case ErrorNoEnv:
		if fs.NoEnv != nil {
			return fs.NoEnv(string(m))
		}
	case ErrorParseSome:
		if fs.ParseSome != nil {
			return fs.ParseSome(string(m))
		}
	case ErrorParseFail:
		if fs.ParseFail != nil {
			return fs.ParseFail(string(m))
		}
	case ErrorPureFailed:
		if fs.PureFailed != nil {
			return fs.PureFailed(string(m))
		}
	case ErrorMissing:
		if fs.Missing != nil {
			return fs.Missing(m)
		}
	case ErrorParseFailure:
		if fs.ParseFailure != nil {
			return fs.ParseFailure(m.e)
		}
	case ErrorStrictPos:
		if fs.StrictPos != nil {
			return fs.StrictPos(m.a, m.metavar)
		}
	case ErrorNonStrictPos:
		if fs.NonStrictPos != nil {
			return fs.NonStrictPos(m.a, m.metavar)
		}
	case ErrorParseFailed:
		if fs.ParseFailed != nil {
			return fs.ParseFailed(m.a, m.s)
		}
	case ErrorGuardFailed:
		if fs.GuardFailed != nil {
			return fs.GuardFailed(m.a, m.s)
		}
	case ErrorNoArgument:
		if fs.NoArgument != nil {
			return fs.NoArgument(m.ix, m.metavar)
		}
	case ErrorUnconsumed:
		if fs.Unconsumed != nil {
			return fs.Unconsumed(int(m))
		}
	case ErrorAmbiguity:
		if fs.Ambiguity != nil {
			return fs.Ambiguity(m.ix, m.name)
		}
	case ErrorSuggestion:
		if fs.Suggestion != nil {
			return fs.Suggestion(m.ix, m.suggestion)
		}
	case ErrorConflict:
		if fs.Conflict != nil {
			return fs.Conflict(m.winner, m.loser)
		}
	case ErrorExpected:
		if fs.Expected != nil {
			return fs.Expected(m.expected, m.actual)
		}
	case ErrorOnlyOnce:
		if fs.OnlyOnce != nil {
			return fs.OnlyOnce(m.winner, m.loser)
		}
	}
	var res T
	return res
}

func errorToString(self Error) string {
	return MessageF(self, MessageFs[string]{
		func(msg string) string { return fmt.Sprintf("NoEnv(%#v)", msg) },
		func(msg string) string { return fmt.Sprintf("ParseSome(%#v)", msg) },
		func(msg string) string { return msg },
		func(msg string) string { return msg },
		func(xs []MissingItem) string {
			return fmt.Sprintf("Missing(%s)", strings.Join(fun.Map[string]((MissingItem).String, xs...), ", "))
		},
		func(a ParseFailure) string { return fmt.Sprintf("ParseFailure(%#v)", a) },
		func(a int, b Metavar) string { return fmt.Sprintf("StrictPos(%#v, %#v)", a, b) },
		func(a int, b Metavar) string { return fmt.Sprintf("NonStrictPos(%#v, %#v)", a, b) },
		func(a Option[int], b string) string { return fmt.Sprintf("ParseFailed(%#v, %#v)", a, b) },
		func(a Option[int], b string) string { return fmt.Sprintf("GuardFailed(%#v, %#v)", a, b) },
		func(a int, b Metavar) string { return fmt.Sprintf("NoArgument(%#v, %#v)", a, b) },
		func(a int) string { return fmt.Sprintf("Unconsumed(%#v)", a) },
		func(a int, b string) string { return fmt.Sprintf("Ambiguity(%#v, %#v)", a, b) },
		func(a int, b Suggestion) string { return fmt.Sprintf("Suggestion(%#v, %#v)", a, b) },
		func(a int, b int) string { return fmt.Sprintf("Conflict(%#v, %#v)", a, b) },
		func(items []Item, b Option[int]) string { return fmt.Sprintf("Expected(%#v, %#v)", items, b) },
		func(a int, b int) string { return fmt.Sprintf("OnlyOnce(%#v, %#v)", a, b) },
	})
}
