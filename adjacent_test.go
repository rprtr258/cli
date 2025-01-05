package cli_test

import (
	"testing"

	"github.com/rprtr258/fun"
	"github.com/stretchr/testify/require"

	. "github.com/rprtr258/cli"
)

func TestTest_adjacent(t *testing.T) {
	a := ReqFlag(Short('a'), struct{}{})
	b := Short('b').Switch()
	c := Short('c').Switch()
	parser := Many(Triple(a.Parser(), b, c).Adjacent(), false).Parser().ToOptions()

	type R = Truple[struct{}, bool, bool]
	{
		r, err := parser.RunInner(NewArgs("-a", "-c", "-a", "-b", "-a", "-b", "-c"))
		require.Nil(t, err)
		// adjacent groups here argument
		// -a [-b] -c  | -a -b [-c] | -a -b -c
		require.Equal(t, r, []R{{struct{}{}, false, true}, {struct{}{}, true, false}, {struct{}{}, true, true}})
	}
	{
		r, err := parser.RunInner(NewArgs("-a"))
		require.Nil(t, err)
		require.Equal(t, r, []R{{struct{}{}, false, false}})
	}
	{
		r, err := parser.RunInner(NewArgs("-a", "-c"))
		require.Nil(t, err)
		require.Equal(t, r, []R{{struct{}{}, false, true}})
	}
	{
		r, err := parser.RunInner(NewArgs())
		require.Nil(t, err)
		require.Equal(t, r, []R{})
	}
}

func TestTest_adjacent_prefix(t *testing.T) {
	type P = fun.Pair[struct{}, uint]
	type R = fun.Pair[Option[P], bool]
	a := ReqFlag(Short('a'), struct{}{})
	b := Positional("X", ParseUint)
	ab := Optional(Pair(a, b).Adjacent().Parser(), false)
	c := Short('c').Switch()
	parser := Pair(ab, c).Parser().ToOptions()

	AssertSuccess(t, R{Invalid[P](), true}, parser, "-c")
	AssertSuccess(t, R{Valid(P{struct{}{}, 1}), false}, parser, "-a", "1")
	AssertSuccess(t, R{Valid(P{struct{}{}, 1}), true}, parser, "-c", "-a", "1")
	AssertSuccess(t, R{Valid(P{struct{}{}, 1}), true}, parser, "-a", "1", "-c")
}

func TestAdjacent_error_message_pos_single(t *testing.T) {
	a := ReqFlag(Short('a'), struct{}{})
	b := Positional("B", ParseUint)
	c := Positional("C", ParseUint)
	d := Short('d').Switch()
	adj := Triple(a, b, c).Adjacent()
	parser := Pair(adj, d).Parser().ToOptions()

	_, r := parser.RunInner(NewArgs("-a", "10"))
	require.Equal(t, UnwrapStderr(r), "expected `C`, pass `--help` for usage information")
}

func TestAdjacent_error_message_arg_single0(t *testing.T) {
	a := ReqFlag(Short('a'), struct{}{})
	b := Argument(Short('b'), "B", ParseUint)
	parser := Pair(a, b).Adjacent().Parser().ToOptions()

	_, r := parser.RunInner(NewArgs("-a", "10"))
	require.Equal(t, "expected `-b=B`, got `10`. Pass `--help` for usage information", UnwrapStderr(r))
}

func TestAdjacent_error_message_arg_single(t *testing.T) {
	a := ReqFlag(Short('a'), struct{}{})
	b := Argument(Short('b'), "B", ParseUint)
	c := Argument(Short('c'), "C", ParseUint)
	d := Short('d').Switch()
	adj := Triple(a, b, c).Adjacent()
	parser := Pair(adj, d).Parser().ToOptions()

	_, r := parser.RunInner(NewArgs("-a", "10"))
	require.Equal(t, "expected `-b=B`, got `10`. Pass `--help` for usage information", UnwrapStderr(r))
}

func TestAdjacent_error_message_pos_many(t *testing.T) {
	a := ReqFlag(Short('a'), struct{}{})
	b := Positional("B", ParseUint)
	c := Positional("C", ParseUint)
	d := Short('d').Switch()
	adj := Many(Triple(a, b, c).Adjacent(), false)
	parser := Pair(adj, d).Parser().ToOptions()

	_, r := parser.RunInner(NewArgs("-a", "10"))
	require.Equal(t, UnwrapStderr(r), "expected `C`, pass `--help` for usage information")
}

func TestAdjacent_error_message_arg_many(t *testing.T) {
	a := ReqFlag(Short('a'), struct{}{})
	b := Argument(Short('b'), "B", ParseUint)
	c := Argument(Short('c'), "C", ParseUint)
	d := Short('d').Switch()
	adj := Many(Triple(a, b, c).Adjacent(), false)
	parser := Pair(adj, d).Parser().ToOptions()

	_, r := parser.RunInner(NewArgs("-a", "10"))
	// this should ask for -b or -c and complain on 10...
	require.Equal(t, "expected `-b=B`, got `10`. Pass `--help` for usage information", UnwrapStderr(r))
}

func TestAdjacent_is_adjacent(t *testing.T) {
	var a = ReqFlag(Short('a'), struct{}{})
	var b = Positional("B", ParseUint)
	var parser = Many(Map(Pair(a, b).Adjacent().Parser(), func(t fun.Pair[struct{}, uint]) uint { return t.V }), false).Parser().ToOptions()

	_, r := parser.RunInner(NewArgs("-a", "-a", "10", "20"))
	require.Equal(t, UnwrapStderr(r), "expected `B`, pass `--help` for usage information")

	{
		r, err := parser.RunInner(NewArgs("-a", "10", "-a", "20"))
		require.Nil(t, err)
		require.Equal(t, r, []uint{10, 20})
	}
}

func TestAdjacent_with_switch(t *testing.T) {
	a := ReqFlag(Short('a'), struct{}{})
	b := Positional("B", ParseUint)
	ab := Many(Map(Pair(a, b).Adjacent().Parser(), func(t fun.Pair[struct{}, uint]) uint { return t.V }), false)
	c := Short('c').Switch()
	parser := Pair(ab, c).Parser().ToOptions()

	type R = fun.Pair[[]uint, bool]
	{
		r, err := parser.RunInner(NewArgs("-a", "10", "-c"))
		require.Nil(t, err)
		require.Equal(t, r, R{[]uint{10}, true})
	}
	{
		r, err := parser.RunInner(NewArgs("-a", "10", "-c", "-a", "20"))
		require.Nil(t, err)
		require.Equal(t, r, R{[]uint{10, 20}, true})
	}
	{
		r, err := parser.RunInner(NewArgs("-c", "-a", "10", "-a", "20"))
		require.Nil(t, err)
		require.Equal(t, r, R{[]uint{10, 20}, true})
	}
}

func TestAdjacent_limits_commands(t *testing.T) {
	x := Pure(struct{}{}).Parser().ToOptions().Command("a").Adjacent()
	s := Short('s').Switch()
	parser := Pair(s, x).Parser().ToOptions()

	r, err := parser.RunInner(NewArgs("a", "-s"))
	require.Nil(t, err)
	require.Equal(t, r, PPair(true, struct{}{}))
}

func TestCommands_and_adjacent(t *testing.T) {
	eat := Positional("FOOD", ParseString).Parser().
		ToOptions().
		Command("eat").
		Help("eat something").
		Adjacent()

	sleep := Argument(Long("time"),
		"HOURS", ParseString).Parser().
		ToOptions().
		Command("sleep").
		Help("sleep for a bit").
		Adjacent()

	cmds := Choice(eat, sleep)
	switch_ := Short('s').Switch()

	var parser = Pair(switch_, cmds).Parser().ToOptions()

	{
		r, err := parser.RunInner(NewArgs("sleep", "--time", "12", "-s"))
		require.Nil(t, err)
		require.Equal(t, r, PPair(true, "12"))
	}

	_, r := parser.RunInner(NewArgs("--help"))
	// TODO - this is ugly
	const expected = `Usage: [-s] COMMAND ...

Available options:
    -s
    -h, --help  Prints help information

Available commands:
    eat         eat something
    sleep       sleep for a bit
`
	require.Equal(t, UnwrapStdout(r), expected)
}

func TestTwo_adjacent_args(t *testing.T) {
	x := Argument(Short('x'), "X", parseInt)
	y := Argument(Short('y'), "Y", parseInt)
	c := Short('c').Switch()
	point := Pair(x, y).Adjacent()
	parser := Pair(point, c).Parser().ToOptions()

	{
		r, err := parser.RunInner(NewArgs("-x", "3", "-y", "4", "-c"))
		require.Nil(t, err)
		require.Equal(t, r, PPair(PPair(3, 4), true))
	}
	{
		// they are adjacent to each other, but the way it is coded currently - they must be adjacent
		// to the first element.
		// Proper fix is to split "adjacent" into "adjacent to" and "adjacent block"

		// TODO: get back
		// r, err := parser.RunInner(NewArgs("-y", "3", "-x", "4", "-c"))
		// require.Nil(t, err)
		// require.Equal(t, r, pair(pair(4, 3), true))
	}
	{
		_, r := parser.RunInner(NewArgs("-y", "3", "-c", "-x", "4"))
		require.Equal(t, UnwrapStderr(r), "expected `-y=Y`, pass `--help` for usage information")
	}
}
