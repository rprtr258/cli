package cli

import (
	"fmt"
	"strings"

	"github.com/rprtr258/fun"
)

type ErrorMessage interface {
	isErrorMessage()
}

// those can be caught ---------------------------------------------------------------
// Tried to consume an env variable with no fallback, variable was not set
type ErrorMessageNoEnv string

func (ErrorMessageNoEnv) isErrorMessage() {}

// User specified an error message on some
type ErrorMessageParseSome string

func (ErrorMessageParseSome) isErrorMessage() {}

// User asked for parser to fail explicitly
type ErrorMessageParseFail string

func (ErrorMessageParseFail) isErrorMessage() {}

// pure_with failed to parse a value
type ErrorMessagePureFailed string

func (ErrorMessagePureFailed) isErrorMessage() {}

// Expected one of those values
//
// Used internally to generate better error messages
type ErrorMessageMissing []MissingItem

func (ErrorMessageMissing) isErrorMessage() {}

// those cannot be caught-------------------------------------------------------------
// Parsing failed and this is the final output
type ErrorMessageParseFailure struct{ e ParseFailure }

func (ErrorMessageParseFailure) isErrorMessage() {}

// Tried to consume a strict positional argument, value was present but was not strictly positional
type ErrorMessageStrictPos struct {
	a int
	b Metavar
}

func (ErrorMessageStrictPos) isErrorMessage() {}

// Tried to consume a non-strict positional argument, but the value was strict
type ErrorMessageNonStrictPos struct {
	a int
	b Metavar
}

func (ErrorMessageNonStrictPos) isErrorMessage() {}

// Parser provided by user failed to parse a value
type ErrorMessageParseFailed struct {
	a Option[int]
	b string
}

func (ErrorMessageParseFailed) isErrorMessage() {}

// Parser provided by user failed to validate a value
type ErrorMessageGuardFailed struct {
	a Option[int]
	b string
}

func (ErrorMessageGuardFailed) isErrorMessage() {}

// Argument requres a value but something else was passed,
// required: --foo <BAR>
// given: --foo --bar
//
//	--foo -- bar
//	--foo
type ErrorMessageNoArgument struct {
	a int
	b Metavar
}

func (ErrorMessageNoArgument) isErrorMessage() {}

// Parser is expected to consume all the things from the command line
// this item will contain an index of the unconsumed value
type ErrorMessageUnconsumed /* TODO: unused? */ int

func (ErrorMessageUnconsumed) isErrorMessage() {}

// argument is ambigoups - parser can accept it as both a set of flags and a short flag with no =
type ErrorMessageAmbiguity struct {
	a int
	b string
}

func (ErrorMessageAmbiguity) isErrorMessage() {}

// Suggested fixes for typos or missing input
type ErrorMessageSuggestion struct {
	a int
	b Suggestion
}

func (ErrorMessageSuggestion) isErrorMessage() {}

// Two arguments are mutually exclusive
// --release --dev
type ErrorMessageConflict struct{ winner, loser int }

func (ErrorMessageConflict) isErrorMessage() {}

// Expected one or more items in the scope, got someting else if any
type ErrorMessageExpected struct {
	a []Item
	b Option[int]
}

func (ErrorMessageExpected) isErrorMessage() {}

// Parameter is accepted but only once
type ErrorMessageOnlyOnce struct{ winner, loser int }

func (ErrorMessageOnlyOnce) isErrorMessage() {}

func Message(
	self ErrorMessage,
	errorMessageNoEnv func(string),
	errorMessageParseSome func(string),
	errorMessageParseFail func(string),
	errorMessagePureFailed func(string),
	errorMessageMissing func([]MissingItem),
	errorMessageParseFailure func(ParseFailure),
	errorMessageStrictPos func(int, Metavar),
	errorMessageNonStrictPos func(int, Metavar),
	errorMessageParseFailed func(Option[int], string),
	errorMessageGuardFailed func(Option[int], string),
	errorMessageNoArgument func(int, Metavar),
	errorMessageUnconsumed func(int),
	errorMessageAmbiguity func(int, string),
	errorMessageSuggestion func(int, Suggestion),
	errorMessageConflict func(int, int),
	errorMessageExpected func([]Item, Option[int]),
	errorMessageOnlyOnce func(int, int),
) {
	switch self := self.(type) {
	case ErrorMessageNoEnv:
		if errorMessageNoEnv != nil {
			errorMessageNoEnv(string(self))
		}
	case ErrorMessageParseSome:
		if errorMessageParseSome != nil {
			errorMessageParseSome(string(self))
		}
	case ErrorMessageParseFail:
		if errorMessageParseFail != nil {
			errorMessageParseFail(string(self))
		}
	case ErrorMessagePureFailed:
		if errorMessagePureFailed != nil {
			errorMessagePureFailed(string(self))
		}
	case ErrorMessageMissing:
		if errorMessageMissing != nil {
			errorMessageMissing(self)
		}
	case ErrorMessageParseFailure:
		if errorMessageParseFailure != nil {
			errorMessageParseFailure(self.e)
		}
	case ErrorMessageStrictPos:
		if errorMessageStrictPos != nil {
			errorMessageStrictPos(self.a, self.b)
		}
	case ErrorMessageNonStrictPos:
		if errorMessageNonStrictPos != nil {
			errorMessageNonStrictPos(self.a, self.b)
		}
	case ErrorMessageParseFailed:
		if errorMessageParseFailed != nil {
			errorMessageParseFailed(self.a, self.b)
		}
	case ErrorMessageGuardFailed:
		if errorMessageGuardFailed != nil {
			errorMessageGuardFailed(self.a, self.b)
		}
	case ErrorMessageNoArgument:
		if errorMessageNoArgument != nil {
			errorMessageNoArgument(self.a, self.b)
		}
	case ErrorMessageUnconsumed:
		if errorMessageUnconsumed != nil {
			errorMessageUnconsumed(int(self))
		}
	case ErrorMessageAmbiguity:
		if errorMessageAmbiguity != nil {
			errorMessageAmbiguity(self.a, self.b)
		}
	case ErrorMessageSuggestion:
		if errorMessageSuggestion != nil {
			errorMessageSuggestion(self.a, self.b)
		}
	case ErrorMessageConflict:
		if errorMessageConflict != nil {
			errorMessageConflict(self.winner, self.loser)
		}
	case ErrorMessageExpected:
		if errorMessageExpected != nil {
			errorMessageExpected(self.a, self.b)
		}
	case ErrorMessageOnlyOnce:
		if errorMessageOnlyOnce != nil {
			errorMessageOnlyOnce(self.winner, self.loser)
		}
	}
}

func MessageNoEnv(s string) ErrorMessage {
	return ErrorMessageNoEnv(s)
}

func MessageParseFailed(a Option[int], s string) ErrorMessage {
	return ErrorMessageParseFailed{a, s}
}

func MessageGuardFailed(a Option[int], s string) ErrorMessage {
	return ErrorMessageGuardFailed{a, s}
}

func MessagePureFailed(s string) ErrorMessage {
	return ErrorMessagePureFailed(s)
}

func MessageParseFailure(e ParseFailure) ErrorMessage {
	return ErrorMessageParseFailure{e}
}

func MessageStrictPos(a int, b Metavar) ErrorMessage {
	return ErrorMessageStrictPos{a, b}
}

func MessageNonStrictPos(a int, b Metavar) ErrorMessage {
	return ErrorMessageNonStrictPos{a, b}
}

func MessageMissing(xs ...MissingItem) ErrorMessage {
	return ErrorMessageMissing(xs)
}

func MessageParseFail(s string) ErrorMessage {
	return ErrorMessageParseFail(s)
}

func MessageNoArgument(ix int, metavar Metavar) ErrorMessage {
	return ErrorMessageNoArgument{ix, metavar}
}

func MessageUnconsumed(a int) ErrorMessage {
	return ErrorMessageUnconsumed(a)
}

func MessageAmbiguity(ix int, short string) ErrorMessage {
	return ErrorMessageAmbiguity{ix, short}
}

func MessageSuggestion(a int, b Suggestion) ErrorMessage {
	return ErrorMessageSuggestion{a, b}
}

func MessageConflict(a, b int) ErrorMessage {
	return ErrorMessageConflict{a, b}
}

func MessageExpected(a []Item, b Option[int]) ErrorMessage {
	return ErrorMessageExpected{a, b}
}

func MessageOnlyOnce(a, b int) ErrorMessage {
	return ErrorMessageOnlyOnce{a, b}
}

func IsMissing(m ErrorMessage) bool {
	return MessageF(m, MessageFs[bool]{
		Missing: func([]MissingItem) bool { return true },
	})
}

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

func MessageF[T any](m ErrorMessage, fs MessageFs[T]) T {
	var res T
	Message(m,
		func(a string) {
			if fs.NoEnv != nil {
				res = fs.NoEnv(a)
			}
		},
		func(a string) {
			if fs.ParseSome != nil {
				res = fs.ParseSome(a)
			}
		},
		func(a string) {
			if fs.ParseFail != nil {
				res = fs.ParseFail(a)
			}
		},
		func(a string) {
			if fs.PureFailed != nil {
				res = fs.PureFailed(a)
			}
		},
		func(a []MissingItem) {
			if fs.Missing != nil {
				res = fs.Missing(a)
			}
		},
		func(a ParseFailure) {
			if fs.ParseFailure != nil {
				res = fs.ParseFailure(a)
			}
		},
		func(a int, b Metavar) {
			if fs.StrictPos != nil {
				res = fs.StrictPos(a, b)
			}
		},
		func(a int, b Metavar) {
			if fs.NonStrictPos != nil {
				res = fs.NonStrictPos(a, b)
			}
		},
		func(a Option[int], b string) {
			if fs.ParseFailed != nil {
				res = fs.ParseFailed(a, b)
			}
		},
		func(a Option[int], b string) {
			if fs.GuardFailed != nil {
				res = fs.GuardFailed(a, b)
			}
		},
		func(a int, b Metavar) {
			if fs.NoArgument != nil {
				res = fs.NoArgument(a, b)
			}
		},
		func(a int) {
			if fs.Unconsumed != nil {
				res = fs.Unconsumed(a)
			}
		},
		func(a int, b string) {
			if fs.Ambiguity != nil {
				res = fs.Ambiguity(a, b)
			}
		},
		func(a int, b Suggestion) {
			if fs.Suggestion != nil {
				res = fs.Suggestion(a, b)
			}
		},
		func(a int, b int) {
			if fs.Conflict != nil {
				res = fs.Conflict(a, b)
			}
		},
		func(a []Item, b Option[int]) {
			if fs.Expected != nil {
				res = fs.Expected(a, b)
			}
		},
		func(a int, b int) {
			if fs.OnlyOnce != nil {
				res = fs.OnlyOnce(a, b)
			}
		},
	)
	return res
}

func String(self ErrorMessage) string {
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
