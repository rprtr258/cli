package cli

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPositional_with_help(t *testing.T) {
	user := Positional("USER", ParseString).Help("github user\nin two lines")
	api := Positional("API_KEY", ParseString).Help("api key to use")
	parser := Pair(user, api).ToOptions()

	const expected_help = `Usage: USER API_KEY

Available positional items:
    USER        github user in two lines
    API_KEY     api key to use

Available options:
    -h, --help  Prints help information
`
	assertHelp(t, parser, expected_help)
}

func TestHelp_for_positional(t *testing.T) {
	c := Positional("C", ParseString).Help("help for\nc")
	d := Positional("DDD", ParseString).Help("help for\nddd")
	parser := Pair(c, d).ToOptions()

	const expected_help = `Usage: C DDD

Available positional items:
    C           help for c
    DDD         help for ddd

Available options:
    -h, --help  Prints help information
`
	assertHelp(t, parser, expected_help)
}

func TestDash_is_positional(t *testing.T) {
	a := Positional("FILE", ParseString)
	parser := a.ToOptions()
	assertSuccess(t, "-", parser, "-")
}

func TestHelpful_error_message(t *testing.T) {
	parser := Positional("FOO", ParseString).Parser().Some("you need to specify at least one FOO").
		ToOptions()

	assertFailure(t, parser, "you need to specify at least one FOO")
}

func TestPositional_argument(t *testing.T) {
	p := Positional("FILE", ParseString).
		Help("file name").
		Parser().GroupHelp("File to process")
	parser := p.ToOptions()

	const expected = `Usage: FILE

File to process
    FILE        file name

Available options:
    -h, --help  Prints help information
`

	assertHelp(t, parser, expected)
}

func TestPositional_help_complain_1(t *testing.T) {
	v := catchPanic[string](t, func() {
		a := Positional("a", ParseString)
		b := Short('b').Switch()
		parser := Pair(a, b).ToOptions()

		_, err := parser.run(newArgList("--help"))
		unwrapStderr(err)
	})
	require.True(t, strings.HasPrefix(v, "cli usage BUG: all positional"))
}

func TestPositional_help_complain_2(t *testing.T) {
	v := catchPanic[string](t, func() {
		a := Positional("a", ParseString)
		b := Short('b').Switch()
		ba := Pair(b, a)
		c := Short('c').Switch()
		parser := Pair(ba, c).ToOptions()

		_, err := parser.run(newArgList("--help"))
		unwrapStderr(err)
	})
	require.True(t, strings.HasPrefix(v, "cli usage BUG: all positional"))
}

func TestPositional_help_complain_3(t *testing.T) {
	v := catchPanic[string](t, func() {
		a := Positional("a", ParseString)
		b := Short('b').Argument("B", ParseString)
		ba := Choice(b, a)
		c := Short('c').Switch()
		parser := Pair(ba, c).ToOptions()

		_, err := parser.run(newArgList("--help"))
		unwrapStderr(err)
	})
	require.True(t, strings.HasPrefix(v, "cli usage BUG: all positional"))
}

func TestPositional_help_complain_4(t *testing.T) {
	a := Positional("a", ParseString)
	b := Short('b').Argument("B", ParseString)
	parser := Choice(b, a).ToOptions()

	_, err := parser.run(newArgList("--help"))
	unwrapStdout(err)
}

func TestStrictly_positional(t *testing.T) {
	parser := Positional("A", ParseString).Strict().ToOptions()

	assertHelp(t, parser, "Usage: -- A\n\nAvailable options:\n    -h, --help  Prints help information\n")

	assertFailure(t, parser, "expected `A` to be on the right side of `--`", "a")

	assertSuccess(t, "a", parser, "-- a")

	assertFailure(t, parser, "expected `A` to be on the right side of `--`", "a", "--")

	assertFailure(t, parser, "expected `A`, pass `--help` for usage information", "--")
}

func TestNon_strictly_positional(t *testing.T) {
	parser := Positional("A", ParseString).NonStrict().ToOptions()

	assertSuccess(t, "a", parser, "a")

	assertFailure(t, parser, "expected `A` to be on the left side of `--`", "--", "a")

	assertFailure(t, parser, "expected `A`, pass `--help` for usage information", "--")
}
