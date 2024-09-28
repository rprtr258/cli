package cli

import "fmt"

type Message func(
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
	Unconsumed func( /* TODO: unused? */ uint),

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

func MessageNoEnv(s string) Message {
	return func(NoEnv, _, _, _ func(string), _ func([]MissingItem), _ func(ParseFailure), _ func(uint, Metavar), _ func(uint, Metavar), _ func(Option[uint], string), _ func(Option[uint], string), _ func(int, Metavar), _ func(uint), _ func(int, string), _ func(uint, Suggestion), _ func(uint, uint), _ func([]Item, Option[uint]), _ func(uint, uint)) {
		if NoEnv != nil {
			NoEnv(s)
		}
	}
}

func MessageParseFailed(a Option[int], s string) Message {
	return func(
		_, _, _, _ func(string),
		_ func([]MissingItem),
		_ func(ParseFailure),
		_ func(uint, Metavar),
		_ func(uint, Metavar),
		f func(Option[uint], string),
		_ func(Option[uint], string),
		_ func(int, Metavar),
		_ func(uint),
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

func MessageGuardFailed(a Option[int], s string) Message {
	return func(
		_, _, _, _ func(string),
		_ func([]MissingItem),
		_ func(ParseFailure),
		_ func(uint, Metavar),
		_ func(uint, Metavar),
		_ func(Option[uint], string),
		f func(Option[uint], string),
		_ func(int, Metavar),
		_ func(uint),
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

func MessagePureFailed(s string) Message {
	return func(_, _, _, PureFailed func(string), _ func([]MissingItem), _ func(ParseFailure), _ func(uint, Metavar), _ func(uint, Metavar), _ func(Option[uint], string), _ func(Option[uint], string), _ func(int, Metavar), _ func(uint), _ func(int, string), _ func(uint, Suggestion), _ func(uint, uint), _ func([]Item, Option[uint]), _ func(uint, uint)) {
		if PureFailed != nil {
			PureFailed(s)
		}
	}
}

func MessageParseFailure(e ParseFailure) Message {
	return func(
		_, _, _, _ func(string),
		_ func([]MissingItem),
		ParseFailure func(ParseFailure),
		_ func(uint, Metavar),
		_ func(uint, Metavar),
		_ func(Option[uint], string),
		_ func(Option[uint], string),
		_ func(int, Metavar),
		_ func(uint),
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

func MessageStrictPos(a int, b Metavar) Message {
	return func(
		_, _, _, _ func(string),
		_ func([]MissingItem),
		_ func(ParseFailure),
		f func(uint, Metavar),
		_ func(uint, Metavar),
		_ func(Option[uint], string),
		_ func(Option[uint], string),
		_ func(int, Metavar),
		_ func(uint),
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

func MessageNonStrictPos(a int, b Metavar) Message {
	return func(
		_, _, _, _ func(string),
		_ func([]MissingItem),
		_ func(ParseFailure),
		_ func(uint, Metavar),
		f func(uint, Metavar),
		_ func(Option[uint], string),
		_ func(Option[uint], string),
		_ func(int, Metavar),
		_ func(uint),
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

func MessageMissing(xs ...MissingItem) Message {
	return func(_, _, _, _ func(string), Missing func([]MissingItem), _ func(ParseFailure), _ func(uint, Metavar), _ func(uint, Metavar), _ func(Option[uint], string), _ func(Option[uint], string), _ func(int, Metavar), _ func(uint), _ func(int, string), _ func(uint, Suggestion), _ func(uint, uint), _ func([]Item, Option[uint]), _ func(uint, uint)) {
		if Missing != nil {
			Missing(xs)
		}
	}
}

func MessageParseFail(s string) Message {
	return func(_, _, _, ParseFail func(string), _ func([]MissingItem), _ func(ParseFailure), _ func(uint, Metavar), _ func(uint, Metavar), _ func(Option[uint], string), _ func(Option[uint], string), _ func(int, Metavar), _ func(uint), _ func(int, string), _ func(uint, Suggestion), _ func(uint, uint), _ func([]Item, Option[uint]), _ func(uint, uint)) {
		if ParseFail != nil {
			ParseFail(s)
		}
	}
}

func MessageNoArgument(ix int, metavar Metavar) Message {
	return func(_ func(string), _ func(string), _ func(string), _ func(string), _ func([]MissingItem), _ func(ParseFailure), _ func(uint, Metavar), _ func(uint, Metavar), _ func(Option[uint], string), _ func(Option[uint], string), NoArgument func(int, Metavar), _ func(uint), _ func(int, string), _ func(uint, Suggestion), _ func(uint, uint), _ func([]Item, Option[uint]), _ func(uint, uint)) {
		if NoArgument != nil {
			NoArgument(ix, metavar)
		}
	}
}

func MessageUnconsumed(a int) Message {
	return func(_, _, _, _ func(string), _ func([]MissingItem), _ func(ParseFailure), _ func(uint, Metavar), _ func(uint, Metavar), _ func(Option[uint], string), _ func(Option[uint], string), _ func(int, Metavar), Unconsumed func(uint), _ func(int, string), _ func(uint, Suggestion), _ func(uint, uint), _ func([]Item, Option[uint]), _ func(uint, uint)) {
		if Unconsumed != nil {
			Unconsumed(uint(a))
		}
	}
}

func MessageAmbiguity(ix int, short string) Message {
	return func(_, _, _, _ func(string), _ func([]MissingItem), _ func(ParseFailure), _ func(uint, Metavar), _ func(uint, Metavar), _ func(Option[uint], string), _ func(Option[uint], string), _ func(int, Metavar), _ func(uint), Ambiguity func(int, string), _ func(uint, Suggestion), _ func(uint, uint), _ func([]Item, Option[uint]), _ func(uint, uint)) {
		if Ambiguity != nil {
			Ambiguity(ix, short)
		}
	}
}

func MessageSuggestion(a int, b Suggestion) Message {
	return func(_, _, _, _ func(string), _ func([]MissingItem), _ func(ParseFailure), _ func(uint, Metavar), _ func(uint, Metavar), _ func(Option[uint], string), _ func(Option[uint], string), _ func(int, Metavar), _ func(uint), _ func(int, string), Suggestion func(uint, Suggestion), _ func(uint, uint), _ func([]Item, Option[uint]), _ func(uint, uint)) {
		if Suggestion != nil {
			Suggestion(uint(a), b)
		}
	}
}

func MessageConflict(a int, b int) Message {
	return func(
		_, _, _, _ func(string),
		_ func([]MissingItem),
		_ func(ParseFailure),
		_ func(uint, Metavar),
		_ func(uint, Metavar),
		_ func(Option[uint], string),
		_ func(Option[uint], string),
		_ func(int, Metavar),
		_ func(uint),
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

func MessageExpected(a []Item, b Option[int]) Message {
	return func(
		_, _, _, _ func(string),
		_ func([]MissingItem),
		_ func(ParseFailure),
		_ func(uint, Metavar),
		_ func(uint, Metavar),
		_ func(Option[uint], string),
		_ func(Option[uint], string),
		_ func(int, Metavar),
		_ func(uint),
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

func MessageOnlyOnce(a, b int) Message {
	return func(
		_, _, _, _ func(string),
		_ func([]MissingItem),
		_ func(ParseFailure),
		_ func(uint, Metavar),
		_ func(uint, Metavar),
		_ func(Option[uint], string),
		_ func(Option[uint], string),
		_ func(int, Metavar),
		_ func(uint),
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

func (m Message) IsMissing() bool {
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
	Unconsumed   func(uint) T
	Ambiguity    func(int, string) T
	Suggestion   func(uint, Suggestion) T
	Conflict     func(uint, uint) T
	Expected     func([]Item, Option[uint]) T
	OnlyOnce     func(uint, uint) T
}

func MessageF[T any](m Message, fs MessageFs[T]) T {
	var res T
	m(
		func(a string) { res = fs.NoEnv(a) },
		func(a string) { res = fs.ParseSome(a) },
		func(a string) { res = fs.ParseFail(a) },
		func(a string) { res = fs.PureFailed(a) },
		func(a []MissingItem) { res = fs.Missing(a) },
		func(a ParseFailure) { res = fs.ParseFailure(a) },
		func(a uint, b Metavar) { res = fs.StrictPos(a, b) },
		func(a uint, b Metavar) { res = fs.NonStrictPos(a, b) },
		func(a Option[uint], b string) { res = fs.ParseFailed(a, b) },
		func(a Option[uint], b string) { res = fs.GuardFailed(a, b) },
		func(a int, b Metavar) { res = fs.NoArgument(a, b) },
		func(a uint) { res = fs.Unconsumed(a) },
		func(a int, b string) { res = fs.Ambiguity(a, b) },
		func(a uint, b Suggestion) { res = fs.Suggestion(a, b) },
		func(a uint, b uint) { res = fs.Conflict(a, b) },
		func(a []Item, b Option[uint]) { res = fs.Expected(a, b) },
		func(a uint, b uint) { res = fs.OnlyOnce(a, b) },
	)
	return res
}

func (self Message) to_string() string {
	return MessageF(self, MessageFs[string]{
		func(a string) string { return fmt.Sprintf("NoEnv(%#v)", a) },
		func(a string) string { return fmt.Sprintf("ParseSome(%#v)", a) },
		func(a string) string { return fmt.Sprintf("ParseFail(%#v)", a) },
		func(a string) string { return fmt.Sprintf("PureFailed(%#v)", a) },
		func(a []MissingItem) string { return fmt.Sprintf("Missing(%#v)", a) },
		func(a ParseFailure) string { return fmt.Sprintf("ParseFailure(%#v)", a) },
		func(a uint, b Metavar) string { return fmt.Sprintf("StrictPos(%#v, %#v)", a, b) },
		func(a uint, b Metavar) string { return fmt.Sprintf("NonStrictPos(%#v, %#v)", a, b) },
		func(a Option[uint], b string) string { return fmt.Sprintf("ParseFailed(%#v, %#v)", a, b) },
		func(a Option[uint], b string) string { return fmt.Sprintf("GuardFailed(%#v, %#v)", a, b) },
		func(a int, b Metavar) string { return fmt.Sprintf("NoArgument(%#v, %#v)", a, b) },
		func(a uint) string { return fmt.Sprintf("Unconsumed(%#v)", a) },
		func(a int, b string) string { return fmt.Sprintf("Ambiguity(%#v, %#v)", a, b) },
		func(a uint, b Suggestion) string { return fmt.Sprintf("Suggestion(%#v, %#v)", a, b) },
		func(a uint, b uint) string { return fmt.Sprintf("Conflict(%#v, %#v)", a, b) },
		func(a []Item, b Option[uint]) string { return fmt.Sprintf("Expected(%#v, %#v)", a, b) },
		func(a uint, b uint) string { return fmt.Sprintf("OnlyOnce(%#v, %#v)", a, b) },
	})
}
