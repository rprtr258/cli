package cli

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPositional_with_help(t *testing.T) {
	user := Positional("USER", parseString).Help("github user\nin two lines")
	api := Positional("API_KEY", parseString).Help("api key to use")
	parser := Pair(user, api).Parser().ToOptions()

	_, err := parser.run_inner(newArgs("--help"))
	const expected_help = `Usage: USER API_KEY

Available positional items:
    USER        github user in two lines
    API_KEY     api key to use

Available options:
    -h, --help  Prints help information
`
	require.Equal(t, expected_help, unwrap_stdout(err))
}

func TestHelp_for_positional(t *testing.T) {
	c := Positional("C", parseString).Help("help for\nc")
	d := Positional("DDD", parseString).Help("help for\nddd")
	parser := Pair(c, d).Parser().ToOptions()
	_, err := parser.run_inner(newArgs("--help"))

	const expected_help = `Usage: C DDD

Available positional items:
    C           help for c
    DDD         help for ddd

Available options:
    -h, --help  Prints help information
`
	require.Equal(t, expected_help, unwrap_stdout(err))
}

func TestDash_is_positional(t *testing.T) {
	a := Positional("FILE", parseString)
	parser := a.Parser().ToOptions()
	res, err := parser.run_inner(newArgs("-"))
	require.Nil(t, err)
	require.Equal(t, "-", res)
}

func TestHelpful_error_message(t *testing.T) {
	parser := ParseSomeParser(Positional("FOO", parseString).
		Parser().Some("you need to specify at least one FOO")).
		ToOptions()

	_, err := parser.run_inner(newArgs())
	require.Equal(t, "you need to specify at least one FOO", unwrap_stderr(err))
}

func TestPositional_argument(t *testing.T) {
	p := Positional("FILE", parseString).
		Help("file name").
		Parser().GroupHelp("File to process")
	parser := p.Parser().ToOptions()

	_, err := parser.run_inner(newArgs("--help"))

	const expected = `Usage: FILE

File to process
    FILE        file name

Available options:
    -h, --help  Prints help information
`

	require.Equal(t, expected, unwrap_stdout(err))
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

		_, err := parser.run_inner(newArgs("--help"))
		unwrap_stderr(err)
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

		_, err := parser.run_inner(newArgs("--help"))
		unwrap_stderr(err)
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

		_, err := parser.run_inner(newArgs("--help"))
		unwrap_stderr(err)
	})
	require.True(t, strings.HasPrefix(v, "bpaf usage BUG: all positional"))
}

func TestPositional_help_complain_4(t *testing.T) {
	a := Positional("a", parseString)
	b := Argument(Short('b'), "B", parseString)
	parser := Choice(b, a).Parser().ToOptions()

	_, err := parser.run_inner(newArgs("--help"))
	unwrap_stdout(err)
}

func TestStrictly_positional(t *testing.T) {
	parser := Positional("A", parseString).strict().Parser().ToOptions()

	{
		_, r := parser.run_inner(newArgs("--help"))
		require.Equal(t,
			"Usage: -- A\n\nAvailable options:\n    -h, --help  Prints help information\n",
			unwrap_stdout(r),
		)
	}

	{
		_, r := parser.run_inner(newArgs("a"))
		require.Equal(t, "expected `A` to be on the right side of `--`", unwrap_stderr(r))
	}

	{
		r, err := parser.run_inner(newArgs("--", "a"))
		require.Nil(t, err)
		require.Equal(t, "a", r)
	}

	{
		_, r := parser.run_inner(newArgs("a", "--"))
		require.Equal(t, "expected `A` to be on the right side of `--`", unwrap_stderr(r))
	}

	{
		_, r := parser.run_inner(newArgs("--"))
		require.Equal(t, "expected `A`, pass `--help` for usage information", unwrap_stderr(r))
	}
}

func TestNon_strictly_positional(t *testing.T) {
	parser := Positional("A", parseString).non_strict().Parser().ToOptions()

	{
		r, err := parser.run_inner(newArgs("a"))
		require.Nil(t, err)
		require.Equal(t, r, "a")
	}

	{
		_, r := parser.run_inner(newArgs("--", "a"))
		require.Equal(t, "expected `A` to be on the left side of `--`", unwrap_stderr(r))
	}

	{
		_, r := parser.run_inner(newArgs("--"))
		require.Equal(t, "expected `A`, pass `--help` for usage information", unwrap_stderr(r))
	}
}
