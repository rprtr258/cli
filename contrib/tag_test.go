package flags

import (
	"testing"
)

func TestTagMissingColon(t *testing.T) {
	opts := struct {
		TestValue bool `short`
	}{}

	assertParseFail(t, &Error{
		Kind:    ErrTag,
		Message: "expected `:' after key name, but got end of tag (in `short`)",
	}, &opts, "")
}

func TestTagMissingValue(t *testing.T) {
	opts := struct {
		TestValue bool `short:`
	}{}

	assertParseFail(t, &Error{
		Kind:    ErrTag,
		Message: "expected `\"' to start tag value at end of tag (in `short:`)",
	}, &opts, "")
}

func TestTagMissingQuote(t *testing.T) {
	opts := struct {
		TestValue bool `short:"v`
	}{}

	assertParseFail(t, &Error{
		Kind:    ErrTag,
		Message: "expected end of tag value `\"' at end of tag (in `short:\"v`)",
	}, &opts, "")
}

func TestTagNewline(t *testing.T) {
	opts := struct {
		TestValue bool `long:"verbose" description:"verbose
something"`
	}{}

	assertParseFail(t, &Error{
		Kind:    ErrTag,
		Message: "unexpected newline in tag value `description' (in `long:\"verbose\" description:\"verbose\nsomething\"`)",
	}, &opts, "")
}
