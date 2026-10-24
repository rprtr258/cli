package cli

import (
	"testing"

	"github.com/rprtr258/fun"
)

func TestParse_anywhere_positional(t *testing.T) {
	a := Any("X", func(h string) (string, bool) { return h, h != "--help" }, true).
		Help("all the things")

	b := Short('b').Help("batch mode").Switch()
	parser := Pair(a, b).ToOptions()

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
	b := Positional("X", ParseUint)
	ab := Pair(a, b).Adjacent()
	c := Short('c').Switch()
	parser := Pair(ab, c).ToOptions()

	// Usage: -a <x> [-c],

	assertFailure(t, parser, "expected `X`, pass `--help` for usage information", "3", "-a")
	assertFailure(t, parser, "expected `X`, pass `--help` for usage information", "-a")
	assertFailure(t, parser, "couldn't parse `221b`: invalid syntax", "-a", "221b")
	assertFailure(t, parser, "expected `X`, pass `--help` for usage information", "-c", "-a")
	assertFailure(t, parser, "couldn't parse `221b`: invalid syntax", "-c", "-a", "221b")
	assertFailure(t, parser, "expected `X`, pass `--help` for usage information", "-a", "-c")
	assertFailure(t, parser, "couldn't parse `221b`: invalid syntax", "-a", "221b", "-c")
}

func TestAnywhere_catch_optional(t *testing.T) {
	a := ReqFlag(Short('a'), struct{}{})
	b := Positional("x", ParseUint)
	ab := Pair(a, b).Adjacent().OptionalCatch()
	bc := Short('a').Switch()
	parser := Pair(ab, bc).ToOptions()

	type P = fun.Pair[struct{}, uint]
	type R = fun.Pair[Opt[P], bool]
	assertSuccess(t, R{Valid(P{struct{}{}, 10}), false}, parser, "-a", "10")
	assertSuccess(t, R{Invalid[P](), true}, parser, "-a")
	assertSuccess(t, R{Invalid[P](), false}, parser)
}

func TestAnywhere_catch_many(t *testing.T) {
	a := ReqFlag(Short('a'), struct{}{})
	b := Positional("x", ParseUint)
	ab := Pair(a, b).Adjacent().ManyCatch()
	bc := Short('a').Switch()
	parser := Pair(ab, bc).ToOptions()

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
	b := Positional("x", ParseUint)
	ab := Pair(a, b).Adjacent().Fallback(P{struct{}{}, 10})
	bc := Short('a').Switch()
	parser := Pair(ab, bc).ToOptions()

	assertSuccess(t, R{P{struct{}{}, 12}, false}, parser, "-a", "12")
	assertSuccess(t, R{P{struct{}{}, 10}, false}, parser, "-a")
	assertSuccess(t, R{P{struct{}{}, 10}, false}, parser)
}

func TestParse_anywhere_catch_optional(t *testing.T) {
	a := ReqFlag(Short('a'), struct{}{})
	b := Positional("x", ParseUint)

	// optional + catch makes it so parser succeeds without consuming anything
	// usually leaving `-a` untouched to be consumed by something else
	ab := Pair(a, b).Adjacent().OptionalCatch()
	c := Short('c').Switch()
	parser := Pair(ab, c).ToOptions()

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
	mode := Positional("value", ParseUint)
	a := Pair(tag, mode).Adjacent().ManyCatch()
	b := Short('b').Switch()
	type P = fun.Pair[struct{}, uint]
	type R = fun.Pair[[]P, bool]
	parser := Pair(a, b).ToOptions()

	assertSuccess(t, R{[]P{{struct{}{}, 12}}, true}, parser, "-b", "-mode", "12")
	assertSuccess(t, R{[]P{{struct{}{}, 12}}, true}, parser, "-mode", "12", "-b")
	assertSuccess(t, R{[]P{{struct{}{}, 12}}, false}, parser, "-mode", "12")
}
