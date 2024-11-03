package cli

import (
	"fmt"
	"strings"

	"github.com/rprtr258/fun"
)

type ErrorMessage func(
	// those can be caught ---------------------------------------------------------------
	// Tried to consume an env variable with no fallback, variable was not set
	NoEnv func(string),

	// User specified an error message on some
	ParseSome func(string),

	// User asked for parser to fail explicitly
	ParseFail func(string),

	// pure_with failed to parse a value
	PureFailed func(string),

	// Expected one of those values
	//
	// Used internally to generate better error messages
	Missing func([]MissingItem),

	// those cannot be caught-------------------------------------------------------------
	// Parsing failed and this is the final output
	ParseFailure func(ParseFailure),

	// Tried to consume a strict positional argument, value was present but was not strictly positional
	StrictPos func(uint, Metavar),

	// Tried to consume a non-strict positional argument, but the value was strict
	NonStrictPos func(uint, Metavar),

	// Parser provided by user failed to parse a value
	ParseFailed func(Option[uint], string),

	// Parser provided by user failed to validate a value
	GuardFailed func(Option[uint], string),

	// Argument requres a value but something else was passed,
	// required: --foo <BAR>
	// given: --foo --bar
	//        --foo -- bar
	//        --foo
	NoArgument func(int, Metavar),

	// Parser is expected to consume all the things from the command line
	// this item will contain an index of the unconsumed value
	Unconsumed func( /* TODO: unused? */ int),

	// argument is ambigoups - parser can accept it as both a set of flags and a short flag with no =
	Ambiguity func(int, string),

	// Suggested fixes for typos or missing input
	Suggestion func(uint, Suggestion),

	// Two arguments are mutually exclusive
	// --release --dev
	Conflict func( /* winner */ uint, uint),

	// Expected one or more items in the scope, got someting else if any
	Expected func([]Item, Option[uint]),

	// Parameter is accepted but only once
	OnlyOnce func( /* winner */ uint, uint),
)

func MessageNoEnv(s string) ErrorMessage {
	return func(NoEnv, _, _, _ func(string), _ func([]MissingItem), _ func(ParseFailure), _ func(uint, Metavar), _ func(uint, Metavar), _ func(Option[uint], string), _ func(Option[uint], string), _ func(int, Metavar), _ func(int), _ func(int, string), _ func(uint, Suggestion), _ func(uint, uint), _ func([]Item, Option[uint]), _ func(uint, uint)) {
		if NoEnv != nil {
			NoEnv(s)
		}
	}
}

func MessageParseFailed(a Option[int], s string) ErrorMessage {
	return func(
		_, _, _, _ func(string),
		_ func([]MissingItem),
		_ func(ParseFailure),
		_ func(uint, Metavar),
		_ func(uint, Metavar),
		f func(Option[uint], string),
		_ func(Option[uint], string),
		_ func(int, Metavar),
		_ func(int),
		_ func(int, string),
		_ func(uint, Suggestion),
		_ func(uint, uint),
		_ func([]Item, Option[uint]),
		_ func(uint, uint),
	) {
		if f != nil {
			f(Option[uint]{uint(a.Value), a.Valid}, s)
		}
	}
}

func MessageGuardFailed(a Option[int], s string) ErrorMessage {
	return func(
		_, _, _, _ func(string),
		_ func([]MissingItem),
		_ func(ParseFailure),
		_ func(uint, Metavar),
		_ func(uint, Metavar),
		_ func(Option[uint], string),
		f func(Option[uint], string),
		_ func(int, Metavar),
		_ func(int),
		_ func(int, string),
		_ func(uint, Suggestion),
		_ func(uint, uint),
		_ func([]Item, Option[uint]),
		_ func(uint, uint),
	) {
		if f != nil {
			f(Option[uint]{uint(a.Value), a.Valid}, s)
		}
	}
}

func MessagePureFailed(s string) ErrorMessage {
	return func(_, _, _, PureFailed func(string), _ func([]MissingItem), _ func(ParseFailure), _ func(uint, Metavar), _ func(uint, Metavar), _ func(Option[uint], string), _ func(Option[uint], string), _ func(int, Metavar), _ func(int), _ func(int, string), _ func(uint, Suggestion), _ func(uint, uint), _ func([]Item, Option[uint]), _ func(uint, uint)) {
		if PureFailed != nil {
			PureFailed(s)
		}
	}
}

func MessageParseFailure(e ParseFailure) ErrorMessage {
	return func(
		_, _, _, _ func(string),
		_ func([]MissingItem),
		ParseFailure func(ParseFailure),
		_ func(uint, Metavar),
		_ func(uint, Metavar),
		_ func(Option[uint], string),
		_ func(Option[uint], string),
		_ func(int, Metavar),
		_ func(int),
		_ func(int, string),
		_ func(uint, Suggestion),
		_ func(uint, uint),
		_ func([]Item, Option[uint]),
		_ func(uint, uint),
	) {
		if ParseFailure != nil {
			ParseFailure(e)
		}
	}
}

func MessageStrictPos(a int, b Metavar) ErrorMessage {
	return func(
		_, _, _, _ func(string),
		_ func([]MissingItem),
		_ func(ParseFailure),
		f func(uint, Metavar),
		_ func(uint, Metavar),
		_ func(Option[uint], string),
		_ func(Option[uint], string),
		_ func(int, Metavar),
		_ func(int),
		_ func(int, string),
		_ func(uint, Suggestion),
		_ func(uint, uint),
		_ func([]Item, Option[uint]),
		_ func(uint, uint),
	) {
		if f != nil {
			f(uint(a), b)
		}
	}
}

func MessageNonStrictPos(a int, b Metavar) ErrorMessage {
	return func(
		_, _, _, _ func(string),
		_ func([]MissingItem),
		_ func(ParseFailure),
		_ func(uint, Metavar),
		f func(uint, Metavar),
		_ func(Option[uint], string),
		_ func(Option[uint], string),
		_ func(int, Metavar),
		_ func(int),
		_ func(int, string),
		_ func(uint, Suggestion),
		_ func(uint, uint),
		_ func([]Item, Option[uint]),
		_ func(uint, uint),
	) {
		if f != nil {
			f(uint(a), b)
		}
	}
}

func MessageMissing(xs ...MissingItem) ErrorMessage {
	return func(_, _, _, _ func(string), f func([]MissingItem), _ func(ParseFailure), _ func(uint, Metavar), _ func(uint, Metavar), _ func(Option[uint], string), _ func(Option[uint], string), _ func(int, Metavar), _ func(int), _ func(int, string), _ func(uint, Suggestion), _ func(uint, uint), _ func([]Item, Option[uint]), _ func(uint, uint)) {
		if f != nil {
			f(xs)
		}
	}
}

func MessageParseFail(s string) ErrorMessage {
	return func(_, _, _, f func(string), _ func([]MissingItem), _ func(ParseFailure), _ func(uint, Metavar), _ func(uint, Metavar), _ func(Option[uint], string), _ func(Option[uint], string), _ func(int, Metavar), _ func(int), _ func(int, string), _ func(uint, Suggestion), _ func(uint, uint), _ func([]Item, Option[uint]), _ func(uint, uint)) {
		if f != nil {
			f(s)
		}
	}
}

func MessageNoArgument(ix int, metavar Metavar) ErrorMessage {
	return func(_ func(string), _ func(string), _ func(string), _ func(string), _ func([]MissingItem), _ func(ParseFailure), _ func(uint, Metavar), _ func(uint, Metavar), _ func(Option[uint], string), _ func(Option[uint], string), f func(int, Metavar), _ func(int), _ func(int, string), _ func(uint, Suggestion), _ func(uint, uint), _ func([]Item, Option[uint]), _ func(uint, uint)) {
		if f != nil {
			f(ix, metavar)
		}
	}
}

func MessageUnconsumed(a int) ErrorMessage {
	return func(_, _, _, _ func(string), _ func([]MissingItem), _ func(ParseFailure), _ func(uint, Metavar), _ func(uint, Metavar), _ func(Option[uint], string), _ func(Option[uint], string), _ func(int, Metavar), f func(int), _ func(int, string), _ func(uint, Suggestion), _ func(uint, uint), _ func([]Item, Option[uint]), _ func(uint, uint)) {
		if f != nil {
			f(a)
		}
	}
}

func MessageAmbiguity(ix int, short string) ErrorMessage {
	return func(_, _, _, _ func(string), _ func([]MissingItem), _ func(ParseFailure), _ func(uint, Metavar), _ func(uint, Metavar), _ func(Option[uint], string), _ func(Option[uint], string), _ func(int, Metavar), _ func(int), f func(int, string), _ func(uint, Suggestion), _ func(uint, uint), _ func([]Item, Option[uint]), _ func(uint, uint)) {
		if f != nil {
			f(ix, short)
		}
	}
}

func MessageSuggestion(a int, b Suggestion) ErrorMessage {
	return func(_, _, _, _ func(string), _ func([]MissingItem), _ func(ParseFailure), _ func(uint, Metavar), _ func(uint, Metavar), _ func(Option[uint], string), _ func(Option[uint], string), _ func(int, Metavar), _ func(int), _ func(int, string), f func(uint, Suggestion), _ func(uint, uint), _ func([]Item, Option[uint]), _ func(uint, uint)) {
		if f != nil {
			f(uint(a), b)
		}
	}
}

func MessageConflict(a int, b int) ErrorMessage {
	return func(
		_, _, _, _ func(string),
		_ func([]MissingItem),
		_ func(ParseFailure),
		_ func(uint, Metavar),
		_ func(uint, Metavar),
		_ func(Option[uint], string),
		_ func(Option[uint], string),
		_ func(int, Metavar),
		_ func(int),
		_ func(int, string),
		_ func(uint, Suggestion),
		Conflict func(uint, uint),
		_ func([]Item, Option[uint]),
		_ func(uint, uint),
	) {
		if Conflict != nil {
			Conflict(uint(a), uint(b))
		}
	}
}

func MessageExpected(a []Item, b Option[int]) ErrorMessage {
	return func(
		_, _, _, _ func(string),
		_ func([]MissingItem),
		_ func(ParseFailure),
		_ func(uint, Metavar),
		_ func(uint, Metavar),
		_ func(Option[uint], string),
		_ func(Option[uint], string),
		_ func(int, Metavar),
		_ func(int),
		_ func(int, string),
		_ func(uint, Suggestion),
		_ func(uint, uint),
		f func([]Item, Option[uint]),
		_ func(uint, uint),
	) {
		if f != nil {
			f(a, Option[uint]{uint(b.Value), b.Valid})
		}
	}
}

func MessageOnlyOnce(a, b int) ErrorMessage {
	return func(
		_, _, _, _ func(string),
		_ func([]MissingItem),
		_ func(ParseFailure),
		_ func(uint, Metavar),
		_ func(uint, Metavar),
		_ func(Option[uint], string),
		_ func(Option[uint], string),
		_ func(int, Metavar),
		_ func(int),
		_ func(int, string),
		_ func(uint, Suggestion),
		_ func(uint, uint),
		_ func([]Item, Option[uint]),
		f func(uint, uint),
	) {
		if f != nil {
			f(uint(a), uint(b))
		}
	}
}

func (m ErrorMessage) IsMissing() bool {
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
	StrictPos    func(uint, Metavar) T
	NonStrictPos func(uint, Metavar) T
	ParseFailed  func(Option[uint], string) T
	GuardFailed  func(Option[uint], string) T
	NoArgument   func(int, Metavar) T
	Unconsumed   func(int) T
	Ambiguity    func(int, string) T
	Suggestion   func(uint, Suggestion) T
	Conflict     func(uint, uint) T
	Expected     func([]Item, Option[uint]) T
	OnlyOnce     func(uint, uint) T
}

func MessageF[T any](m ErrorMessage, fs MessageFs[T]) T {
	var res T
	m(
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
		func(a uint, b Metavar) {
			if fs.StrictPos != nil {
				res = fs.StrictPos(a, b)
			}
		},
		func(a uint, b Metavar) {
			if fs.NonStrictPos != nil {
				res = fs.NonStrictPos(a, b)
			}
		},
		func(a Option[uint], b string) {
			if fs.ParseFailed != nil {
				res = fs.ParseFailed(a, b)
			}
		},
		func(a Option[uint], b string) {
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
		func(a uint, b Suggestion) {
			if fs.Suggestion != nil {
				res = fs.Suggestion(a, b)
			}
		},
		func(a uint, b uint) {
			if fs.Conflict != nil {
				res = fs.Conflict(a, b)
			}
		},
		func(a []Item, b Option[uint]) {
			if fs.Expected != nil {
				res = fs.Expected(a, b)
			}
		},
		func(a uint, b uint) {
			if fs.OnlyOnce != nil {
				res = fs.OnlyOnce(a, b)
			}
		},
	)
	return res
}

func (self ErrorMessage) String() string {
	return MessageF(self, MessageFs[string]{
		func(msg string) string { return fmt.Sprintf("NoEnv(%#v)", msg) },
		func(msg string) string { return fmt.Sprintf("ParseSome(%#v)", msg) },
		func(msg string) string { return fmt.Sprintf("ParseFail(%#v)", msg) },
		func(msg string) string { return msg },
		func(xs []MissingItem) string {
			return fmt.Sprintf("Missing(%s)", strings.Join(fun.Map[string]((MissingItem).String, xs...), ", "))
		},
		func(a ParseFailure) string { return fmt.Sprintf("ParseFailure(%#v)", a) },
		func(a uint, b Metavar) string { return fmt.Sprintf("StrictPos(%#v, %#v)", a, b) },
		func(a uint, b Metavar) string { return fmt.Sprintf("NonStrictPos(%#v, %#v)", a, b) },
		func(a Option[uint], b string) string { return fmt.Sprintf("ParseFailed(%#v, %#v)", a, b) },
		func(a Option[uint], b string) string { return fmt.Sprintf("GuardFailed(%#v, %#v)", a, b) },
		func(a int, b Metavar) string { return fmt.Sprintf("NoArgument(%#v, %#v)", a, b) },
		func(a int) string { return fmt.Sprintf("Unconsumed(%#v)", a) },
		func(a int, b string) string { return fmt.Sprintf("Ambiguity(%#v, %#v)", a, b) },
		func(a uint, b Suggestion) string { return fmt.Sprintf("Suggestion(%#v, %#v)", a, b) },
		func(a uint, b uint) string { return fmt.Sprintf("Conflict(%#v, %#v)", a, b) },
		func(items []Item, b Option[uint]) string { return fmt.Sprintf("Expected(%#v, %#v)", items, b) },
		func(a uint, b uint) string { return fmt.Sprintf("OnlyOnce(%#v, %#v)", a, b) },
	})
}
