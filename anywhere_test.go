package cli

import (
	"testing"

	"github.com/rprtr258/fun"
	"github.com/stretchr/testify/require"
)

func TestParse_anywhere_positional(t *testing.T) {
	a := Any("X", func(h string) (string, bool) { return h, h != "--help" }, true).
		Help("all the things")

	b := Short('b').Help("batch mode").Switch()
	parser := Pair(a, b).Parser().ToOptions()

	_, r := parser.run_inner(NewArgs("--help"))

	const expected = `Usage: X [-b]

Available options:
    X           all the things
    -b          batch mode
    -h, --help  Prints help information
`

	require.Equal(t, unwrap_stdout(r), expected)
	// this should be allowed because "anywhere" prevents anything inside from being positional
	parser.check_invariants(true)
}

func TestParse_anywhere_no_catch(t *testing.T) {
	a := ReqFlag(Short('a'), struct{}{})
	b := Positional("X", ParseUint)
	ab := Pair(a, b).Adjacent()
	c := Short('c').Switch()
	parser := Pair(ab, c).Parser().ToOptions()

	// Usage: -a <x> [-c],

	{
		_, r := parser.run_inner(NewArgs("3", "-a"))
		require.Equal(t, unwrap_stderr(r), "expected `X`, pass `--help` for usage information")
	}
	{
		_, r := parser.run_inner(NewArgs("-a"))
		require.Equal(t, unwrap_stderr(r), "expected `X`, pass `--help` for usage information")
	}
	{
		_, r := parser.run_inner(NewArgs("-a", "221b"))
		require.Equal(t, unwrap_stderr(r), "couldn't parse `221b`: invalid digit found in string")
	}
	{
		_, r := parser.run_inner(NewArgs("-c", "-a"))
		require.Equal(t, unwrap_stderr(r), "expected `X`, pass `--help` for usage information")
	}
	{
		_, r := parser.run_inner(NewArgs("-c", "-a", "221b"))
		require.Equal(t, unwrap_stderr(r), "couldn't parse `221b`: invalid digit found in string")
	}
	{
		_, r := parser.run_inner(NewArgs("-a", "-c"))
		require.Equal(t, unwrap_stderr(r), "expected `X`, pass `--help` for usage information")
	}
	{
		_, r := parser.run_inner(NewArgs("-a", "221b", "-c"))
		require.Equal(t, unwrap_stderr(r), "couldn't parse `221b`: invalid digit found in string")
	}
}

func TestAnywhere_catch_optional(t *testing.T) {
	a := ReqFlag(Short('a'), struct{}{})
	b := Positional("x", ParseUint)
	ab := Optional(Pair(a, b).Adjacent().Parser(), true)
	bc := Short('a').Switch()
	parser := Pair(ab, bc).Parser().ToOptions()

	type P = fun.Pair[struct{}, uint]
	type R = fun.Pair[Option[P], bool]
	{
		r, err := parser.run_inner(NewArgs("-a", "10"))
		require.Nil(t, err)
		require.Equal(t, r, R{Valid(P{struct{}{}, 10}), false})
	}
	{
		r, err := parser.run_inner(NewArgs("-a"))
		require.Nil(t, err)
		require.Equal(t, r, R{Invalid[P](), true})
	}
	{
		r, err := parser.run_inner(NewArgs())
		require.Nil(t, err)
		require.Equal(t, r, R{Invalid[P](), false})
	}
}

func TestAnywhere_catch_many(t *testing.T) {
	a := ReqFlag(Short('a'), struct{}{})
	b := Positional("x", ParseUint)
	ab := Many(Pair(a, b).Adjacent(), true)
	bc := Short('a').Switch()
	parser := Pair(ab, bc).Parser().ToOptions()

	type P = fun.Pair[struct{}, uint]
	type R = fun.Pair[[]fun.Pair[struct{}, uint], bool]
	{
		r, err := parser.run_inner(NewArgs("-a"))
		require.Nil(t, err)
		require.Equal(t, r, R{[]P{}, true})
	}
	{
		r, err := parser.run_inner(NewArgs("-a", "10"))
		require.Nil(t, err)
		require.Equal(t, r, R{[]P{{struct{}{}, 10}}, false})
	}
	{
		r, err := parser.run_inner(NewArgs())
		require.Nil(t, err)
		require.Equal(t, r, R{[]P{}, false})
	}
}

func TestAnywhere_catch_fallback(t *testing.T) {
	type P = fun.Pair[struct{}, uint]
	a := ReqFlag(Short('a'), struct{}{})
	b := Positional("x", ParseUint)
	ab := Pair(a, b).Adjacent().Parser().Fallback(P{struct{}{}, 10})
	bc := Short('a').Switch()
	parser := Pair(ab, bc).Parser().ToOptions()

	type R = fun.Pair[P, bool]
	{
		r, err := parser.run_inner(NewArgs("-a", "12"))
		require.Nil(t, err)
		require.Equal(t, R{P{struct{}{}, 12}, false}, r)
	}
	{
		t.Skip() // TODO: fix
		r, err := parser.run_inner(NewArgs("-a"))
		require.Nil(t, err)
		require.Equal(t, R{P{struct{}{}, 10}, true}, r)
	}
	{
		r, err := parser.run_inner(NewArgs())
		require.Nil(t, err)
		require.Equal(t, R{P{struct{}{}, 10}, false}, r)
	}
}

func TestParse_anywhere_catch_optional(t *testing.T) {
	a := ReqFlag(Short('a'), struct{}{})
	b := Positional("x", ParseUint)

	// optional + catch makes it so parser succeeds without consuming anything
	// usually leaving `-a` untouched to be consumed by something else
	ab := Optional(Pair(a, b).Adjacent().Parser(), true)
	c := Short('c').Switch()
	parser := Pair(ab, c).Parser().ToOptions()

	{
		_, r := parser.run_inner(NewArgs("-a", "221b"))
		require.Equal(t, unwrap_stderr(r), "`-a` is not expected in this context")
	}
	{
		_, r := parser.run_inner(NewArgs("3", "-a"))
		require.Equal(t, unwrap_stderr(r), "`3` is not expected in this context")
	}
	{
		_, r := parser.run_inner(NewArgs("-a"))
		require.Equal(t, unwrap_stderr(r), "`-a` is not expected in this context")
	}
	{
		_, r := parser.run_inner(NewArgs("-c", "-a"))
		require.Equal(t, unwrap_stderr(r), "`-a` is not expected in this context")
	}
	{
		_, r := parser.run_inner(NewArgs("-c", "-a", "221b"))
		require.Equal(t, unwrap_stderr(r), "`-a` is not expected in this context")
	}
	{
		_, r := parser.run_inner(NewArgs("-a", "-c"))
		require.Equal(t, unwrap_stderr(r), "`-a` is not expected in this context")
	}
	{
		_, r := parser.run_inner(NewArgs("-a", "221b", "-c"))
		require.Equal(t, unwrap_stderr(r), "`-a` is not expected in this context")
	}
}

func TestAnywhere_literal(t *testing.T) {
	tag := Any(
		"-mode",
		func(x string) (struct{}, bool) { return struct{}{}, x == "-mode" },
		false,
	)
	mode := Positional("value", ParseUint)
	a := Many(Pair(tag, mode).Adjacent(), true)
	b := Short('b').Switch()
	type P = fun.Pair[struct{}, uint]
	type R = fun.Pair[[]P, bool]
	parser := Pair(a, b).Parser().ToOptions()

	{
		t.Skip() // TODO: fix
		r, err := parser.run_inner(NewArgs("-b", "-mode", "12"))
		require.Nil(t, err)
		require.Equal(t, r, R{[]P{{struct{}{}, 12}}, true})
	}
	{
		r, err := parser.run_inner(NewArgs("-mode", "12", "-b"))
		require.Nil(t, err)
		require.Equal(t, r, R{[]P{{struct{}{}, 12}}, true})
	}
	{
		r, err := parser.run_inner(NewArgs("-mode", "12"))
		require.Nil(t, err)
		require.Equal(t, r, R{[]P{{struct{}{}, 12}}, false})
	}
}
