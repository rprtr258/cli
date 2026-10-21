package cli_test

import (
	"testing"

	"github.com/rprtr258/fun"
	"github.com/stretchr/testify/require"

	. "github.com/rprtr258/cli"
)

// Combinators return Parser directly: no intermediate .Parser() calls needed
// anywhere in the chain, and builders compose into combinators directly.
func TestFluentChaining(t *testing.T) {
	verbose := Many(Short('v').Long("verbose").Switch(), false)
	speed := Argument[int](Short('s').Long("speed"), "SPEED", parseInt).
		Parser().
		FallbackDisplay(42).
		Guard(func(x int) bool { return x < 1000 }, "too fast").
		GroupHelp("Speed settings").
		Hide()
	rest := Optional(Positional[string]("REST", parseString), false)

	parser := Map(Pair(verbose, Pair(speed, rest)), func(x fun.Pair[[]bool, fun.Pair[int, Option[string]]]) string {
		res := ""
		for range x.K {
			res += "v"
		}
		res += ":" + strconvItoa(x.V.K)
		if x.V.V.Valid {
			res += ":" + x.V.V.Value
		}
		return res
	}).ToOptions()

	for name, tc := range map[string]struct {
		args     []string
		expected string
	}{
		"all":     {args: []string{"-vvv", "-s", "10", "tail"}, expected: "vvv:10:tail"},
		"default": {args: []string{"tail"}, expected: "v:42:tail"},
	} {
		t.Run(name, func(t *testing.T) {
			v, err := parser.RunE(append([]string{"prog"}, tc.args...))
			require.NoError(t, err)
			require.Equal(t, tc.expected, v)
		})
	}
}

// Count/Some return parsers usable directly in combinators.
func TestFluentCountSome(t *testing.T) {
	parser := Pair(
		ReqFlag[struct{}](Short('c'), struct{}{}).Parser().Count(),
		Some(Positional[uint32]("N", parseUint32), "want N", false),
	).Parser().ToOptions()

	v, err := parser.RunE([]string{"prog", "-c", "-c", "-c", "7"})
	require.NoError(t, err)
	require.Equal(t, 3, v.K)
	require.Equal(t, []uint32{7}, v.V)
}

// FallbackDisplay shows the default in help output.
func TestFluentFallbackDisplay(t *testing.T) {
	parser := Argument[int](Short('s'), "SPEED", parseInt).
		Parser().
		FallbackDisplay(42).
		ToOptions()

	_, err := parser.RunE([]string{"prog", "--help"})
	var ferr FailureError
	require.ErrorAs(t, err, &ferr)
	require.Contains(t, ferr.Error(), "[default: 42]")
}

// Adjacent on a constructed pair chains into further combinators.
func TestFluentAdjacent(t *testing.T) {
	parser := Map(Pair(
		Argument[uint32](Short('w'), "W", parseUint32).Parser(),
		Argument[uint32](Short('h'), "H", parseUint32).Parser(),
	).Adjacent().Hide(), func(x fun.Pair[uint32, uint32]) uint32 { return x.K * x.V }).ToOptions()

	v, err := parser.RunE([]string{"prog", "-w", "2", "-h", "3"})
	require.NoError(t, err)
	require.Equal(t, uint32(6), v)
}

func strconvItoa(x int) string {
	res := ""
	if x == 0 {
		return "0"
	}
	neg := x < 0
	if neg {
		x = -x
	}
	for x > 0 {
		res = string(rune('0'+x%10)) + res
		x /= 10
	}
	if neg {
		res = "-" + res
	}
	return res
}

// Map and Parse are generic methods on Parser: they chain after any combinator
// and infer their type parameters, including on ParseCon results from Pair.
func TestFluentGenericMethods(t *testing.T) {
	parser := Some(Positional[uint32]("N", parseUint32), "want N", false).
		Map(func(xs []uint32) uint32 {
			res := uint32(0)
			for _, x := range xs {
				res += x
			}
			return res
		}).
		ToOptions()
	assertSuccess(t, uint32(3), parser, "1", "2")

	parser2 := Pair(Positional[uint32]("W", parseUint32), Positional[uint32]("H", parseUint32)).
		Map(func(x fun.Pair[uint32, uint32]) uint32 { return x.K * x.V }).
		ToOptions()
	assertSuccess(t, uint32(20), parser2, "4", "5")

	parser3 := Positional[uint32]("N", parseUint32).Parser().
		Parse(func(x uint32) (string, error) { return strconvItoa(int(x)), nil }).
		ToOptions()
	assertSuccess(t, "42", parser3, "42")
}
