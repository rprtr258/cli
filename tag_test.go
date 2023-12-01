package cli

import "testing"

func TestTagMissingColon(t *testing.T) {
	var opts struct {
		TestValue bool `short`
	}

	assertParseFailed(t, ErrTag, "expected `:' after key name, but got end of tag (in `short`)", &opts, "")
}

func TestTagMissingValue(t *testing.T) {
	var opts struct {
		TestValue bool `short:`
	}

	assertParseFailed(t, ErrTag, "expected `\"' to start tag value at end of tag (in `short:`)", &opts, "")
}

func TestTagMissingQuote(t *testing.T) {
	var opts struct {
		TestValue bool `short:"v`
	}

	assertParseFailed(t, ErrTag, "expected end of tag value `\"' at end of tag (in `short:\"v`)", &opts, "")
}

func TestTagNewline(t *testing.T) {
	var opts struct {
		TestValue bool `long:"verbose" description:"verbose
something"`
	}

	assertParseFailed(t, ErrTag, "unexpected newline in tag value `description' (in `long:\"verbose\" description:\"verbose\nsomething\"`)", &opts, "")
}
