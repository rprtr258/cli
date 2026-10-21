package cli

import (
	"testing"

	"github.com/rprtr258/fun"
)

func TestTest_adjacent(t *testing.T) {
	a := ReqFlag(Short('a'), struct{}{})
	b := Short('b').Switch()
	c := Short('c').Switch()
	parser := Many(Triple(a.Parser(), b, c).Adjacent()).Parser().ToOptions()

	type R = Tuple[struct{}, bool, bool]
	assertSuccess(t, []R{{struct{}{}, false, true}, {struct{}{}, true, false}, {struct{}{}, true, true}}, parser, "-a", "-c", "-a", "-b", "-a", "-b", "-c")
	assertSuccess(t, []R{{struct{}{}, false, false}}, parser, "-a")
	assertSuccess(t, []R{{struct{}{}, false, true}}, parser, "-a", "-c")
	assertSuccess(t, []R{}, parser)
}

func TestTest_adjacent_prefix(t *testing.T) {
	type P = fun.Pair[struct{}, uint]
	type R = fun.Pair[Opt[P], bool]
	a := ReqFlag(Short('a'), struct{}{})
	b := Positional("X", parseUint)
	ab := Optional(Pair(a, b).Adjacent().Parser())
	c := Short('c').Switch()
	parser := Pair(ab, c).Parser().ToOptions()

	assertSuccess(t, R{Invalid[P](), true}, parser, "-c")
	assertSuccess(t, R{Valid(P{struct{}{}, 1}), false}, parser, "-a", "1")
	assertSuccess(t, R{Valid(P{struct{}{}, 1}), true}, parser, "-c", "-a", "1")
	assertSuccess(t, R{Valid(P{struct{}{}, 1}), true}, parser, "-a", "1", "-c")
}

func TestAdjacent_error_message_pos_single(t *testing.T) {
	a := ReqFlag(Short('a'), struct{}{})
	b := Positional("B", parseUint)
	c := Positional("C", parseUint)
	d := Short('d').Switch()
	adj := Triple(a, b, c).Adjacent()
	parser := Pair(adj, d).Parser().ToOptions()

	assertFailure(t, parser, "expected `C`, pass `--help` for usage information", "-a", "10")
}

func TestAdjacent_error_message_arg_single0(t *testing.T) {
	a := ReqFlag(Short('a'), struct{}{})
	b := Argument(Short('b'), "B", parseUint)
	parser := Pair(a, b).Adjacent().Parser().ToOptions()

	assertFailure(t, parser, "expected `-b=B`, got `10`. Pass `--help` for usage information", "-a", "10")
}

func TestAdjacent_error_message_arg_single(t *testing.T) {
	a := ReqFlag(Short('a'), struct{}{})
	b := Argument(Short('b'), "B", parseUint)
	c := Argument(Short('c'), "C", parseUint)
	d := Short('d').Switch()
	adj := Triple(a, b, c).Adjacent()
	parser := Pair(adj, d).Parser().ToOptions()

	assertFailure(t, parser, "expected `-b=B`, got `10`. Pass `--help` for usage information", "-a", "10")
}

func TestAdjacent_error_message_pos_many(t *testing.T) {
	a := ReqFlag(Short('a'), struct{}{})
	b := Positional("B", parseUint)
	c := Positional("C", parseUint)
	d := Short('d').Switch()
	adj := Many(Triple(a, b, c).Adjacent())
	parser := Pair(adj, d).Parser().ToOptions()

	assertFailure(t, parser, "expected `C`, pass `--help` for usage information", "-a", "10")
}

func TestAdjacent_error_message_arg_many(t *testing.T) {
	a := ReqFlag(Short('a'), struct{}{})
	b := Argument(Short('b'), "B", parseUint)
	c := Argument(Short('c'), "C", parseUint)
	d := Short('d').Switch()
	adj := Many(Triple(a, b, c).Adjacent())
	parser := Pair(adj, d).Parser().ToOptions()

	// this should ask for -b or -c and complain on 10...
	assertFailure(t, parser, "expected `-b=B`, got `10`. Pass `--help` for usage information", "-a", "10")
}

func TestAdjacent_is_adjacent(t *testing.T) {
	var a = ReqFlag(Short('a'), struct{}{})
	var b = Positional("B", parseUint)
	var parser = Many(Map(Pair(a, b).Adjacent().Parser(), func(t fun.Pair[struct{}, uint]) uint { return t.V })).Parser().ToOptions()

	assertFailure(t, parser, "expected `B`, pass `--help` for usage information", "-a", "-a", "10", "20")

	assertSuccess(t, []uint{10, 20}, parser, "-a", "10", "-a", "20")
}

func TestAdjacent_with_switch(t *testing.T) {
	a := ReqFlag(Short('a'), struct{}{})
	b := Positional("B", parseUint)
	ab := Many(Map(Pair(a, b).Adjacent().Parser(), func(t fun.Pair[struct{}, uint]) uint { return t.V }))
	c := Short('c').Switch()
	parser := Pair(ab, c).Parser().ToOptions()

	type R = fun.Pair[[]uint, bool]
	assertSuccess(t, R{[]uint{10}, true}, parser, "-a", "10", "-c")
	assertSuccess(t, R{[]uint{10, 20}, true}, parser, "-a", "10", "-c", "-a", "20")
	assertSuccess(t, R{[]uint{10, 20}, true}, parser, "-c", "-a", "10", "-a", "20")
}

func TestAdjacent_limits_commands(t *testing.T) {
	x := Pure(struct{}{}).Parser().ToOptions().Command("a").Adjacent()
	s := Short('s').Switch()
	parser := Pair(s, x).Parser().ToOptions()

	assertSuccess(t, PPair(true, struct{}{}), parser, "a", "-s")
}

func TestCommands_and_adjacent(t *testing.T) {
	eat := Positional("FOOD", parseString).Parser().
		ToOptions().
		Command("eat").
		Help("eat something").
		Adjacent()

	sleep := Argument(Long("time"),
		"HOURS", parseString).Parser().
		ToOptions().
		Command("sleep").
		Help("sleep for a bit").
		Adjacent()

	cmds := Choice(eat, sleep)
	switch_ := Short('s').Switch()

	var parser = Pair(switch_, cmds).Parser().ToOptions()

	assertSuccess(t, PPair(true, "12"), parser, "sleep", "--time", "12", "-s")

	// TODO - this is ugly
	const expected = `Usage: [-s] COMMAND ...

Available options:
    -s
    -h, --help  Prints help information

Available commands:
    eat         eat something
    sleep       sleep for a bit
`
	assertHelp(t, parser, expected)
}

func TestTwo_adjacent_args(t *testing.T) {
	x := Argument(Short('x'), "X", parseInt)
	y := Argument(Short('y'), "Y", parseInt)
	c := Short('c').Switch()
	point := Pair(x, y).Adjacent()
	parser := Pair(point, c).Parser().ToOptions()

	assertSuccess(t, PPair(PPair(3, 4), true), parser, "-x", "3", "-y", "4", "-c")
	{
		// they are adjacent to each other, but the way it is coded currently - they must be adjacent
		// to the first element.
		// Proper fix is to split "adjacent" into "adjacent to" and "adjacent block"

		// TODO: get back
		// r, err := parser.RunInner(newArgList("-y", "3", "-x", "4", "-c"))
		// require.Nil(t, err)
		// require.Equal(t, r, pair(pair(4, 3), true))
	}
	assertFailure(t, parser, "expected `-y=Y`, pass `--help` for usage information", "-y", "3", "-c", "-x", "4")
}
