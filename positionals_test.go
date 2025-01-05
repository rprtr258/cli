package cli_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	. "github.com/rprtr258/cli"
)

func TestPositional_with_help(t *testing.T) {
	user := Positional("USER", parseString).Help("github user\nin two lines")
	api := Positional("API_KEY", parseString).Help("api key to use")
	parser := Pair(user, api).Parser().ToOptions()

	_, err := parser.RunInner(NewArgs("--help"))
	const expected_help = `Usage: USER API_KEY

Available positional items:
    USER        github user in two lines
    API_KEY     api key to use

Available options:
    -h, --help  Prints help information
`
	require.Equal(t, expected_help, UnwrapStdout(err))
}

func TestHelp_for_positional(t *testing.T) {
	c := Positional("C", parseString).Help("help for\nc")
	d := Positional("DDD", parseString).Help("help for\nddd")
	parser := Pair(c, d).Parser().ToOptions()
	_, err := parser.RunInner(NewArgs("--help"))

	const expected_help = `Usage: C DDD

Available positional items:
    C           help for c
    DDD         help for ddd

Available options:
    -h, --help  Prints help information
`
	require.Equal(t, expected_help, UnwrapStdout(err))
}

func TestDash_is_positional(t *testing.T) {
	a := Positional("FILE", parseString)
	parser := a.Parser().ToOptions()
	res, err := parser.RunInner(NewArgs("-"))
	require.Nil(t, err)
	require.Equal(t, "-", res)
}

func TestHelpful_error_message(t *testing.T) {
	parser := ParseSomeParser(Positional("FOO", parseString).
		Parser().Some("you need to specify at least one FOO", false)).
		ToOptions()

	_, err := parser.RunInner(NewArgs())
	require.Equal(t, "you need to specify at least one FOO", UnwrapStderr(err))
}

func TestPositional_argument(t *testing.T) {
	p := Positional("FILE", parseString).
		Help("file name").
		Parser().GroupHelp("File to process")
	parser := p.Parser().ToOptions()

	_, err := parser.RunInner(NewArgs("--help"))

	const expected = `Usage: FILE

File to process
    FILE        file name

Available options:
    -h, --help  Prints help information
`

	require.Equal(t, expected, UnwrapStdout(err))
}

func usePanic[T any](t *testing.T, f func()) (res T) {
	t.Helper()

	defer func() {
		if r := recover(); r != nil {
			res = r.(T)
		} else {
			t.FailNow()
		}
	}()
	f()
	return
}

func TestPositional_help_complain_1(t *testing.T) {
	v := usePanic[string](t, func() {
		a := Positional("a", parseString)
		b := Short('b').Switch()
		parser := Pair(a, b).Parser().ToOptions()

		_, err := parser.RunInner(NewArgs("--help"))
		UnwrapStderr(err)
	})
	require.True(t, strings.HasPrefix(v, "bpaf usage BUG: all positional"))
}

func TestPositional_help_complain_2(t *testing.T) {
	v := usePanic[string](t, func() {
		a := Positional("a", parseString)
		b := Short('b').Switch()
		ba := Pair(b, a)
		c := Short('c').Switch()
		parser := Pair(ba, c).Parser().ToOptions()

		_, err := parser.RunInner(NewArgs("--help"))
		UnwrapStderr(err)
	})
	require.True(t, strings.HasPrefix(v, "bpaf usage BUG: all positional"))
}

func TestPositional_help_complain_3(t *testing.T) {
	v := usePanic[string](t, func() {
		a := Positional("a", parseString)
		b := Argument(Short('b'), "B", parseString)
		ba := Choice(b, a)
		c := Short('c').Switch()
		parser := Pair(ba, c).Parser().ToOptions()

		_, err := parser.RunInner(NewArgs("--help"))
		UnwrapStderr(err)
	})
	require.True(t, strings.HasPrefix(v, "bpaf usage BUG: all positional"))
}

func TestPositional_help_complain_4(t *testing.T) {
	a := Positional("a", parseString)
	b := Argument(Short('b'), "B", parseString)
	parser := Choice(b, a).Parser().ToOptions()

	_, err := parser.RunInner(NewArgs("--help"))
	UnwrapStdout(err)
}

func TestStrictly_positional(t *testing.T) {
	parser := Positional("A", parseString).Strict().Parser().ToOptions()

	{
		_, r := parser.RunInner(NewArgs("--help"))
		require.Equal(t,
			"Usage: -- A\n\nAvailable options:\n    -h, --help  Prints help information\n",
			UnwrapStdout(r),
		)
	}

	{
		_, r := parser.RunInner(NewArgs("a"))
		require.Equal(t, "expected `A` to be on the right side of `--`", UnwrapStderr(r))
	}

	{
		r, err := parser.RunInner(NewArgs("--", "a"))
		require.Nil(t, err)
		require.Equal(t, "a", r)
	}

	{
		_, r := parser.RunInner(NewArgs("a", "--"))
		require.Equal(t, "expected `A` to be on the right side of `--`", UnwrapStderr(r))
	}

	{
		_, r := parser.RunInner(NewArgs("--"))
		require.Equal(t, "expected `A`, pass `--help` for usage information", UnwrapStderr(r))
	}
}

func TestNon_strictly_positional(t *testing.T) {
	parser := Positional("A", parseString).NonStrict().Parser().ToOptions()

	{
		r, err := parser.RunInner(NewArgs("a"))
		require.Nil(t, err)
		require.Equal(t, r, "a")
	}

	{
		_, r := parser.RunInner(NewArgs("--", "a"))
		require.Equal(t, "expected `A` to be on the left side of `--`", UnwrapStderr(r))
	}

	{
		_, r := parser.RunInner(NewArgs("--"))
		require.Equal(t, "expected `A`, pass `--help` for usage information", UnwrapStderr(r))
	}
}
