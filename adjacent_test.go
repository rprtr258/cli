package cli

import (
	"testing"

	"github.com/rprtr258/fun"
)

func TestTest_adjacent(t *testing.T) {
	a := Short('a').ReqFlag(struct{}{})
	b := Short('b').Switch()
	c := Short('c').Switch()
	parser := Triple(a, b, c).Adjacent().Many().ToOptions()

	type R = Tuple[struct{}, bool, bool]
	assertSuccess(t, []R{{struct{}{}, false, true}, {struct{}{}, true, false}, {struct{}{}, true, true}}, parser, "-a -c -a -b -a -b -c")
	assertSuccess(t, []R{{struct{}{}, false, false}}, parser, "-a")
	assertSuccess(t, []R{{struct{}{}, false, true}}, parser, "-a -c")
	assertSuccess(t, []R{}, parser, "")
}

func TestTest_adjacent_prefix(t *testing.T) {
	type P = fun.Pair[struct{}, uint]
	type R = fun.Pair[Opt[P], bool]
	a := Short('a').ReqFlag(struct{}{})
	b := Positional("X", ParseUint)
	ab := Pair(a, b).Adjacent().Optional()
	c := Short('c').Switch()
	parser := Pair(ab, c).ToOptions()

	assertSuccess(t, R{Invalid[P](), true}, parser, "-c")
	assertSuccess(t, R{Valid(P{struct{}{}, 1}), false}, parser, "-a 1")
	assertSuccess(t, R{Valid(P{struct{}{}, 1}), true}, parser, "-c -a 1")
	assertSuccess(t, R{Valid(P{struct{}{}, 1}), true}, parser, "-a 1 -c")
}

func TestAdjacent_error_message_pos_single(t *testing.T) {
	a := Short('a').ReqFlag(struct{}{})
	b := Positional("B", ParseUint)
	c := Positional("C", ParseUint)
	d := Short('d').Switch()
	adj := Triple(a, b, c).Adjacent()
	parser := Pair(adj, d).ToOptions()

	assertFailure(t, parser, "expected `C`, pass `--help` for usage information", "-a", "10")
}

func TestAdjacent_error_message_arg_single0(t *testing.T) {
	a := Short('a').ReqFlag(struct{}{})
	b := Short('b').Argument("B", ParseUint)
	parser := Pair(a, b).Adjacent().ToOptions()

	assertFailure(t, parser, "expected `-b=B`, got `10`. Pass `--help` for usage information", "-a", "10")
}

func TestAdjacent_error_message_arg_single(t *testing.T) {
	a := Short('a').ReqFlag(struct{}{})
	b := Short('b').Argument("B", ParseUint)
	c := Short('c').Argument("C", ParseUint)
	d := Short('d').Switch()
	adj := Triple(a, b, c).Adjacent()
	parser := Pair(adj, d).ToOptions()

	assertFailure(t, parser, "expected `-b=B`, got `10`. Pass `--help` for usage information", "-a", "10")
}

func TestAdjacent_error_message_pos_many(t *testing.T) {
	a := Short('a').ReqFlag(struct{}{})
	b := Positional("B", ParseUint)
	c := Positional("C", ParseUint)
	d := Short('d').Switch()
	adj := Triple(a, b, c).Adjacent().Many()
	parser := Pair(adj, d).ToOptions()

	assertFailure(t, parser, "expected `C`, pass `--help` for usage information", "-a", "10")
}

func TestAdjacent_error_message_arg_many(t *testing.T) {
	a := Short('a').ReqFlag(struct{}{})
	b := Short('b').Argument("B", ParseUint)
	c := Short('c').Argument("C", ParseUint)
	d := Short('d').Switch()
	adj := Triple(a, b, c).Adjacent().Many()
	parser := Pair(adj, d).ToOptions()

	// this should ask for -b or -c and complain on 10...
	assertFailure(t, parser, "expected `-b=B`, got `10`. Pass `--help` for usage information", "-a", "10")
}

func TestAdjacent_is_adjacent(t *testing.T) {
	a := Short('a').ReqFlag(struct{}{})
	b := Positional("B", ParseUint)
	parser := Pair(a, b).Adjacent().Map(func(t fun.Pair[struct{}, uint]) uint { return t.V }).Many().ToOptions()

	assertFailure(t, parser, "expected `B`, pass `--help` for usage information", "-a", "-a", "10", "20")

	assertSuccess(t, []uint{10, 20}, parser, "-a 10 -a 20")
}

func TestAdjacent_with_switch(t *testing.T) {
	a := Short('a').ReqFlag(struct{}{})
	b := Positional("B", ParseUint)
	ab := Pair(a, b).Adjacent().Map(func(t fun.Pair[struct{}, uint]) uint { return t.V }).Many()
	c := Short('c').Switch()
	parser := Pair(ab, c).ToOptions()

	type R = fun.Pair[[]uint, bool]
	assertSuccess(t, R{[]uint{10}, true}, parser, "-a 10 -c")
	assertSuccess(t, R{[]uint{10, 20}, true}, parser, "-a 10 -c -a 20")
	assertSuccess(t, R{[]uint{10, 20}, true}, parser, "-c -a 10 -a 20")
}

func TestAdjacent_limits_commands(t *testing.T) {
	x := Pure(struct{}{}).ToOptions().Command("a").Adjacent()
	s := Short('s').Switch()
	parser := Pair(s, x).ToOptions()

	assertSuccess(t, ppair(true, struct{}{}), parser, "a -s")
}

func TestCommands_and_adjacent(t *testing.T) {
	eat := Positional("FOOD", ParseString).ToOptions().
		Command("eat").
		Help("eat something").
		Adjacent()

	sleep := Long("time").Argument(
		"HOURS", ParseString).ToOptions().
		Command("sleep").
		Help("sleep for a bit").
		Adjacent()

	cmds := Choice(eat, sleep)
	switch_ := Short('s').Switch()

	parser := Pair(switch_, cmds).ToOptions()

	assertSuccess(t, ppair(true, "12"), parser, "sleep --time 12 -s")

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
	x := Short('x').Argument("X", ParseInt)
	y := Short('y').Argument("Y", ParseInt)
	c := Short('c').Switch()
	point := Pair(x, y).Adjacent()
	parser := Pair(point, c).ToOptions()

	assertSuccess(t, ppair(ppair(3, 4), true), parser, "-x 3 -y 4 -c")
	// NOTE: adjacent groups must currently be adjacent to the first element; a
	// proper fix is to split "adjacent" into "adjacent to" and "adjacent block".
	// TODO: get back to allowing "-y 3 -x 4 -c" here.
	assertFailure(t, parser, "expected `-y=Y`, pass `--help` for usage information", "-y", "3", "-c", "-x", "4")
}
