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

func errorToString(self Error) string {
	switch self := self.(type) {
	case ErrorNoEnv:
		return fmt.Sprintf("NoEnv(%#v)", self)
	case ErrorParseSome:
		return fmt.Sprintf("ParseSome(%#v)", self)
	case ErrorParseFail:
		return string(self)
	case ErrorPureFailed:
		return string(self)
	case ErrorMissing:
		return fmt.Sprintf("Missing(%s)", strings.Join(fun.Map[string]((MissingItem).String, self...), ", "))
	case ErrorParseFailure:
		return fmt.Sprintf("ParseFailure(%#v)", self.e)
	case ErrorStrictPos:
		return fmt.Sprintf("StrictPos(%#v, %#v)", self.a, self.metavar)
	case ErrorNonStrictPos:
		return fmt.Sprintf("NonStrictPos(%#v, %#v)", self.a, self.metavar)
	case ErrorParseFailed:
		return fmt.Sprintf("ParseFailed(%#v, %#v)", self.a, self.s)
	case ErrorGuardFailed:
		return fmt.Sprintf("GuardFailed(%#v, %#v)", self.a, self.s)
	case ErrorNoArgument:
		return fmt.Sprintf("NoArgument(%#v, %#v)", self.ix, self.metavar)
	case ErrorUnconsumed:
		return fmt.Sprintf("Unconsumed(%#v)", int(self))
	case ErrorAmbiguity:
		return fmt.Sprintf("Ambiguity(%#v, %#v)", self.ix, self.name)
	case ErrorSuggestion:
		return fmt.Sprintf("Suggestion(%#v, %#v)", self.ix, self.suggestion)
	case ErrorConflict:
		return fmt.Sprintf("Conflict(%#v, %#v)", self.winner, self.loser)
	case ErrorExpected:
		return fmt.Sprintf("Expected(%#v, %#v)", self.expected, self.actual)
	case ErrorOnlyOnce:
		return fmt.Sprintf("OnlyOnce(%#v, %#v)", self.winner, self.loser)
	}
	panic("unerasable")
}
