package cli

import (
	"testing"

	"github.com/rprtr258/fun"
	"github.com/stretchr/testify/require"
)

func TestParse_anywhere_positional(t *testing.T) {
	a := Any("X", func(h string) (string, bool) { return h, h != "--help" }).
		Help("all the things").
		Anywhere()

	b := Short('b').Help("batch mode").Switch()
	var parser *OptionParser[fun.Pair[string, bool]] = Pair(a, b).Parser().ToOptions()

	_, r := parser.run_inner(newArgs("--help"))

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
	b := Positional("X", parseUint)
	ab := Pair(a, b).Adjacent()
	c := Short('c').Switch()
	parser := Pair(ab, c).Parser().ToOptions()

	// Usage: -a <x> [-c],

	{
		_, r := parser.run_inner(newArgs("3", "-a"))
		require.Equal(t, unwrap_stderr(r), "expected `X`, pass `--help` for usage information")
	}
	{
		_, r := parser.run_inner(newArgs("-a"))
		require.Equal(t, unwrap_stderr(r), "expected `X`, pass `--help` for usage information")
	}
	{
		_, r := parser.run_inner(newArgs("-a", "221b"))
		require.Equal(t, unwrap_stderr(r), "couldn't parse `221b`: invalid digit found in string")
	}
	{
		_, r := parser.run_inner(newArgs("-c", "-a"))
		require.Equal(t, unwrap_stderr(r), "expected `X`, pass `--help` for usage information")
	}
	{
		_, r := parser.run_inner(newArgs("-c", "-a", "221b"))
		require.Equal(t, unwrap_stderr(r), "couldn't parse `221b`: invalid digit found in string")
	}
	{
		_, r := parser.run_inner(newArgs("-a", "-c"))
		require.Equal(t, unwrap_stderr(r), "expected `X`, pass `--help` for usage information")
	}
	{
		_, r := parser.run_inner(newArgs("-a", "221b", "-c"))
		require.Equal(t, unwrap_stderr(r), "couldn't parse `221b`: invalid digit found in string")
	}
}

func TestAnywhere_catch_optional(t *testing.T) {
	var a = ReqFlag(Short('a'), struct{}{})
	var b = Positional("x", parseUint)
	var ab = Optional(Pair(a, b).Adjacent().Parser()).Catch()
	var bc = Short('a').Switch()
	var parser = Pair(ab, bc).Parser().ToOptions()

	type P = fun.Pair[struct{}, uint]
	type R = fun.Pair[Option[P], bool]
	{
		r, err := parser.run_inner(newArgs("-a", "10"))
		require.Nil(t, err)
		require.Equal(t, r, R{Valid(P{struct{}{}, 10}), false})
	}
	{
		r, err := parser.run_inner(newArgs("-a"))
		require.Nil(t, err)
		require.Equal(t, r, R{Invalid[P](), true})
	}
	{
		r, err := parser.run_inner(newArgs())
		require.Nil(t, err)
		require.Equal(t, r, R{Invalid[P](), false})
	}
}

func TestAnywhere_catch_many(t *testing.T) {
	var a = ReqFlag(Short('a'), struct{}{})
	var b = Positional("x", parseUint)
	var ab = Many(Pair(a, b).Adjacent().Parser()).Catch()
	var bc = Short('a').Switch()
	var parser = Pair(ab, bc).Parser().ToOptions()

	type P = fun.Pair[struct{}, uint]
	type R = fun.Pair[[]fun.Pair[struct{}, uint], bool]
	{
		r, err := parser.run_inner(newArgs("-a"))
		require.Nil(t, err)
		require.Equal(t, r, R{[]P{}, true})
	}
	{
		r, err := parser.run_inner(newArgs("-a", "10"))
		require.Nil(t, err)
		require.Equal(t, r, R{[]P{{struct{}{}, 10}}, false})
	}
	{
		r, err := parser.run_inner(newArgs())
		require.Nil(t, err)
		require.Equal(t, r, R{[]P{}, false})
	}
}

func TestAnywhere_catch_fallback(t *testing.T) {
	type P = fun.Pair[struct{}, uint]
	var a = ReqFlag(Short('a'), struct{}{})
	var b = Positional("x", parseUint)
	var ab = Pair(a, b).Adjacent().Parser().Fallback(P{struct{}{}, 10})
	var bc = Short('a').Switch()
	var parser = Pair(ab, bc).Parser().ToOptions()

	type R = fun.Pair[P, bool]
	{
		r, err := parser.run_inner(newArgs("-a", "12"))
		require.Nil(t, err)
		require.Equal(t, R{P{struct{}{}, 12}, false}, r)
	}
	{
		t.Skip() // TODO: fix
		r, err := parser.run_inner(newArgs("-a"))
		require.Nil(t, err)
		require.Equal(t, R{P{struct{}{}, 10}, true}, r)
	}
	{
		r, err := parser.run_inner(newArgs())
		require.Nil(t, err)
		require.Equal(t, R{P{struct{}{}, 10}, false}, r)
	}
}

func TestParse_anywhere_catch_optional(t *testing.T) {
	var a = ReqFlag(Short('a'), struct{}{})
	var b = Positional("x", parseUint)

	// optional + catch makes it so parser succeeds without consuming anything
	// usually leaving `-a` untouched to be consumed by something else
	var ab = Optional(Pair(a, b).Adjacent().Parser()).Catch()
	var c = Short('c').Switch()
	var parser = Pair(ab, c).Parser().ToOptions()

	{
		_, r := parser.run_inner(newArgs("-a", "221b"))
		require.Equal(t, unwrap_stderr(r), "`-a` is not expected in this context")
	}
	{
		_, r := parser.run_inner(newArgs("3", "-a"))
		require.Equal(t, unwrap_stderr(r), "`3` is not expected in this context")
	}
	{
		_, r := parser.run_inner(newArgs("-a"))
		require.Equal(t, unwrap_stderr(r), "`-a` is not expected in this context")
	}
	{
		_, r := parser.run_inner(newArgs("-c", "-a"))
		require.Equal(t, unwrap_stderr(r), "`-a` is not expected in this context")
	}
	{
		_, r := parser.run_inner(newArgs("-c", "-a", "221b"))
		require.Equal(t, unwrap_stderr(r), "`-a` is not expected in this context")
	}
	{
		_, r := parser.run_inner(newArgs("-a", "-c"))
		require.Equal(t, unwrap_stderr(r), "`-a` is not expected in this context")
	}
	{
		_, r := parser.run_inner(newArgs("-a", "221b", "-c"))
		require.Equal(t, unwrap_stderr(r), "`-a` is not expected in this context")
	}
}

func TestAnywhere_literal(t *testing.T) {
	var tag = Any(
		"-mode",
		func(x string) (struct{}, bool) { return struct{}{}, x == "-mode" },
	)
	var mode = Positional("value", parseUint)
	var a = Many(Pair(tag, mode).Adjacent().Parser()).Catch()
	var b = Short('b').Switch()
	type P = fun.Pair[struct{}, uint]
	type R = fun.Pair[[]P, bool]
	var parser *OptionParser[R] = Pair(a, b).Parser().ToOptions()

	{
		t.Skip() // TODO: fix
		r, err := parser.run_inner(newArgs("-b", "-mode", "12"))
		require.Nil(t, err)
		require.Equal(t, r, R{[]P{{struct{}{}, 12}}, true})
	}
	{
		r, err := parser.run_inner(newArgs("-mode", "12", "-b"))
		require.Nil(t, err)
		require.Equal(t, r, R{[]P{{struct{}{}, 12}}, true})
	}
	{
		r, err := parser.run_inner(newArgs("-mode", "12"))
		require.Nil(t, err)
		require.Equal(t, r, R{[]P{{struct{}{}, 12}}, false})
	}
}
