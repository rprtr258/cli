package cli

import (
	"testing"
)

func TestMapParse(t *testing.T) {
	parser := Long("point").Map("K:V", ParseString, ParseInt).ToOptions()

	assertSuccess(t, map[string]int{}, parser)
	assertSuccess(t, map[string]int{"x": 1}, parser, "--point", "x:1")
	assertSuccess(t, map[string]int{"x": 1, "y": 2}, parser,
		"--point", "x:1", "--point", "y:2")
	// repeated keys: last wins
	assertSuccess(t, map[string]int{"x": 2}, parser, "--point", "x:1", "--point", "x:2")
}

func TestMapValueWithColons(t *testing.T) {
	parser := Long("point").Map("", ParseString, ParseString).ToOptions()

	// split at the first colon
	assertSuccess(t, map[string]string{"k": "a:b"}, parser, "--point", "k:a:b")
}

func TestMapAdjacent(t *testing.T) {
	parser := Long("point").Map("", ParseString, ParseInt).Adjacent().ToOptions()

	assertSuccess(t, map[string]int{"x": 1}, parser, "--point=x:1")
	assertSuccess(t, map[string]int{"x": 1, "y": 2}, parser, "--point=x:1", "--point=y:2")
	assertFailure(t, parser, "`--point` is not expected in this context", "--point", "x:1")
}

func TestMapFailureNoColon(t *testing.T) {
	parser := Long("point").Map("", ParseString, ParseInt).ToOptions()

	assertFailure(t, parser, "couldn't parse `x`: expected `key:value`", "--point", "x")
}

func TestMapFailureBadValue(t *testing.T) {
	parser := Long("point").Map("", ParseString, ParseInt).ToOptions()

	assertFailure(t, parser, "couldn't parse `x:a`: invalid syntax", "--point", "x:a")
}

func TestMapEnv(t *testing.T) {
	parser := Long("point").Env("POINT").Map("", ParseString, ParseInt).ToOptions()

	assertSuccess(t, map[string]int{"x": 1}, parser, "--point", "x:1")
	assertSuccess(t, map[string]int{"x": 1, "y": 2}, parser, "--point", "x:1", "--point", "y:2")

	t.Setenv("POINT", "x:1,y:2")
	// env is a fallback: only used when the flag is absent entirely
	assertSuccess(t, map[string]int{"x": 1, "y": 2}, parser)
	assertSuccess(t, map[string]int{"a": 3}, parser, "--point", "a:3")
}

func TestMapHelp(t *testing.T) {
	parser := Long("point").Map("", ParseString, ParseInt).ToOptions()

	assertHelp(t, parser, `Usage: --point=<KEY:VALUE>

Available options:
        --point=<KEY:VALUE>
    -h, --help               Prints help information
`)
}

func TestMapCompletion(t *testing.T) {
	parser := Long("point").Map("", ParseString, ParseInt).ToOptions()

	assertCompletion(t, parser, "\tKEY:VALUE\t\t\n\n", "--point", "")
}
