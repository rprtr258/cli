package cli_test

import (
	"testing"

	"github.com/rprtr258/fun"

	. "github.com/rprtr258/cli"
)

func TestParse_anywhere_positional(t *testing.T) {
	a := Any("X", func(h string) (string, bool) { return h, h != "--help" }, true).
		Help("all the things")

	b := Short('b').Help("batch mode").Switch()
	parser := Pair(a, b).Parser().ToOptions()

	const expected = `Usage: X [-b]

Available options:
    X           all the things
    -b          batch mode
    -h, --help  Prints help information
`

	assertHelp(t, parser, expected)
	// this should be allowed because "anywhere" prevents anything inside from being positional
	parser.CheckInvariants(true)
}

func TestParse_anywhere_no_catch(t *testing.T) {
	a := ReqFlag(Short('a'), struct{}{})
	b := Positional("X", parseUint)
	ab := Pair(a, b).Adjacent()
	c := Short('c').Switch()
	parser := Pair(ab, c).Parser().ToOptions()

	// Usage: -a <x> [-c],

	assertFailure(t, parser, "expected `X`, pass `--help` for usage information", "3", "-a")
	assertFailure(t, parser, "expected `X`, pass `--help` for usage information", "-a")
	assertFailure(t, parser, "couldn't parse `221b`: invalid digit found in string", "-a", "221b")
	assertFailure(t, parser, "expected `X`, pass `--help` for usage information", "-c", "-a")
	assertFailure(t, parser, "couldn't parse `221b`: invalid digit found in string", "-c", "-a", "221b")
	assertFailure(t, parser, "expected `X`, pass `--help` for usage information", "-a", "-c")
	assertFailure(t, parser, "couldn't parse `221b`: invalid digit found in string", "-a", "221b", "-c")
}

func TestAnywhere_catch_optional(t *testing.T) {
	a := ReqFlag(Short('a'), struct{}{})
	b := Positional("x", parseUint)
	ab := Optional(Pair(a, b).Adjacent().Parser(), true)
	bc := Short('a').Switch()
	parser := Pair(ab, bc).Parser().ToOptions()

	type P = fun.Pair[struct{}, uint]
	type R = fun.Pair[Option[P], bool]
	assertSuccess(t, R{Valid(P{struct{}{}, 10}), false}, parser, "-a", "10")
	assertSuccess(t, R{Invalid[P](), true}, parser, "-a")
	assertSuccess(t, R{Invalid[P](), false}, parser)
}

func TestAnywhere_catch_many(t *testing.T) {
	a := ReqFlag(Short('a'), struct{}{})
	b := Positional("x", parseUint)
	ab := Many(Pair(a, b).Adjacent(), true)
	bc := Short('a').Switch()
	parser := Pair(ab, bc).Parser().ToOptions()

	type P = fun.Pair[struct{}, uint]
	type R = fun.Pair[[]fun.Pair[struct{}, uint], bool]
	assertSuccess(t, R{[]P{}, true}, parser, "-a")
	assertSuccess(t, R{[]P{{struct{}{}, 10}}, false}, parser, "-a", "10")
	assertSuccess(t, R{[]P{}, false}, parser)
}

func TestAnywhere_catch_fallback(t *testing.T) {
	type P = fun.Pair[struct{}, uint]
	type R = fun.Pair[P, bool]
	a := ReqFlag(Short('a'), struct{}{})
	b := Positional("x", parseUint)
	ab := Pair(a, b).Adjacent().Parser().Fallback(P{struct{}{}, 10})
	bc := Short('a').Switch()
	parser := Pair(ab, bc).Parser().ToOptions()

	assertSuccess(t, R{P{struct{}{}, 12}, false}, parser, "-a", "12")
	assertSuccess(t, R{P{struct{}{}, 10}, false}, parser, "-a")
	assertSuccess(t, R{P{struct{}{}, 10}, false}, parser)
}

func TestParse_anywhere_catch_optional(t *testing.T) {
	a := ReqFlag(Short('a'), struct{}{})
	b := Positional("x", parseUint)

	// optional + catch makes it so parser succeeds without consuming anything
	// usually leaving `-a` untouched to be consumed by something else
	ab := Optional(Pair(a, b).Adjacent().Parser(), true)
	c := Short('c').Switch()
	parser := Pair(ab, c).Parser().ToOptions()

	assertFailure(t, parser, "`-a` is not expected in this context", "-a", "221b")
	assertFailure(t, parser, "`3` is not expected in this context", "3", "-a")
	assertFailure(t, parser, "`-a` is not expected in this context", "-a")
	assertFailure(t, parser, "`-a` is not expected in this context", "-c", "-a")
	assertFailure(t, parser, "`-a` is not expected in this context", "-c", "-a", "221b")
	assertFailure(t, parser, "`-a` is not expected in this context", "-a", "-c")
	assertFailure(t, parser, "`-a` is not expected in this context", "-a", "221b", "-c")
}

func TestAnywhere_literal(t *testing.T) {
	tag := Any(
		"-mode",
		func(x string) (struct{}, bool) { return struct{}{}, x == "-mode" },
		false,
	)
	mode := Positional("value", parseUint)
	a := Many(Pair(tag, mode).Adjacent(), true)
	b := Short('b').Switch()
	type P = fun.Pair[struct{}, uint]
	type R = fun.Pair[[]P, bool]
	parser := Pair(a, b).Parser().ToOptions()

	assertSuccess(t, R{[]P{{struct{}{}, 12}}, true}, parser, "-b", "-mode", "12")
	assertSuccess(t, R{[]P{{struct{}{}, 12}}, true}, parser, "-mode", "12", "-b")
	assertSuccess(t, R{[]P{{struct{}{}, 12}}, false}, parser, "-mode", "12")
}
