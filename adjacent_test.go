package cli

import (
	"testing"

	"github.com/rprtr258/fun"
	"github.com/stretchr/testify/require"
)

func TestTest_adjacent(t *testing.T) {
	var a = ReqFlag(Short('a'), struct{}{})
	var b = Short('b').Switch()
	var c = Short('c').Switch()
	var parser = Many(Triple(a.Parser(), b, c).Adjacent(), false).Parser().ToOptions()

	type R = triple[struct{}, bool, bool]
	{
		r, err := parser.run_inner(newArgs("-a", "-c", "-a", "-b", "-a", "-b", "-c"))
		require.Nil(t, err)
		// adjacent groups here argument
		// -a [-b] -c  | -a -b [-c] | -a -b -c
		require.Equal(t, r, []R{{struct{}{}, false, true}, {struct{}{}, true, false}, {struct{}{}, true, true}})
	}
	{
		r, err := parser.run_inner(newArgs("-a"))
		require.Nil(t, err)
		require.Equal(t, r, []R{{struct{}{}, false, false}})
	}
	{
		r, err := parser.run_inner(newArgs("-a", "-c"))
		require.Nil(t, err)
		require.Equal(t, r, []R{{struct{}{}, false, true}})
	}
	{
		r, err := parser.run_inner(newArgs())
		require.Nil(t, err)
		require.Equal(t, r, []R{})
	}
}

func TestTest_adjacent_prefix(t *testing.T) {
	type P = fun.Pair[struct{}, uint]
	type R = fun.Pair[Option[P], bool]
	var a = ReqFlag(Short('a'), struct{}{})
	var b = Positional("X", parseUint)
	var ab = Optional(Pair(a, b).Adjacent().Parser(), false)
	var c = Short('c').Switch()
	var parser = Pair(ab, c).Parser().ToOptions()

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

	_, r := parser.run_inner(newArgs("-a", "10"))
	require.Equal(t, unwrap_stderr(r), "expected `C`, pass `--help` for usage information")
}

func TestAdjacent_error_message_arg_single0(t *testing.T) {
	a := ReqFlag(Short('a'), struct{}{})
	b := Argument(Short('b'), "B", parseUint)
	parser := Pair(a, b).Adjacent().Parser().ToOptions()

	t.Skip() // TODO: fix
	_, r := parser.run_inner(newArgs("-a", "10"))
	require.Equal(t, "expected `-b=B`, got `10`. Pass `--help` for usage information", unwrap_stderr(r))
}

func TestAdjacent_error_message_arg_single(t *testing.T) {
	a := ReqFlag(Short('a'), struct{}{})
	b := Argument(Short('b'), "B", parseUint)
	c := Argument(Short('c'), "C", parseUint)
	d := Short('d').Switch()
	adj := Triple(a, b, c).Adjacent()
	parser := Pair(adj, d).Parser().ToOptions()

	t.Skip() // TODO: fix
	_, r := parser.run_inner(newArgs("-a", "10"))
	require.Equal(t, "expected `-b=B`, got `10`. Pass `--help` for usage information", unwrap_stderr(r))
}

func TestAdjacent_error_message_pos_many(t *testing.T) {
	a := ReqFlag(Short('a'), struct{}{})
	b := Positional("B", parseUint)
	c := Positional("C", parseUint)
	d := Short('d').Switch()
	adj := Many(Triple(a, b, c).Adjacent(), false)
	parser := Pair(adj, d).Parser().ToOptions()

	_, r := parser.run_inner(newArgs("-a", "10"))
	require.Equal(t, unwrap_stderr(r), "expected `C`, pass `--help` for usage information")
}

func TestAdjacent_error_message_arg_many(t *testing.T) {
	a := ReqFlag(Short('a'), struct{}{})
	b := Argument(Short('b'), "B", parseUint)
	c := Argument(Short('c'), "C", parseUint)
	d := Short('d').Switch()
	adj := Many(Triple(a, b, c).Adjacent(), false)
	parser := Pair(adj, d).Parser().ToOptions()

	t.Skip() // TODO: fix
	_, r := parser.run_inner(newArgs("-a", "10"))
	// this should ask for -b or -c and complain on 10...
	require.Equal(t, "expected `-b=B`, got `10`. Pass `--help` for usage information", unwrap_stderr(r))
}

func TestAdjacent_is_adjacent(t *testing.T) {
	var a = ReqFlag(Short('a'), struct{}{})
	var b = Positional("B", parseUint)
	var parser = Many(Map(Pair(a, b).Adjacent().Parser(), func(t fun.Pair[struct{}, uint]) uint { return t.V }), false).Parser().ToOptions()

	_, r := parser.run_inner(newArgs("-a", "-a", "10", "20"))
	require.Equal(t, unwrap_stderr(r), "expected `B`, pass `--help` for usage information")

	{
		r, err := parser.run_inner(newArgs("-a", "10", "-a", "20"))
		require.Nil(t, err)
		require.Equal(t, r, []uint{10, 20})
	}
}

func TestAdjacent_with_switch(t *testing.T) {
	var a = ReqFlag(Short('a'), struct{}{})
	var b = Positional("B", parseUint)
	var ab = Many(Map(Pair(a, b).Adjacent().Parser(), func(t fun.Pair[struct{}, uint]) uint { return t.V }), false)
	var c = Short('c').Switch()
	var parser = Pair(ab, c).Parser().ToOptions()

	type R = fun.Pair[[]uint, bool]
	{
		r, err := parser.run_inner(newArgs("-a", "10", "-c"))
		require.Nil(t, err)
		require.Equal(t, r, R{[]uint{10}, true})
	}
	{
		r, err := parser.run_inner(newArgs("-a", "10", "-c", "-a", "20"))
		require.Nil(t, err)
		require.Equal(t, r, R{[]uint{10, 20}, true})
	}
	{
		r, err := parser.run_inner(newArgs("-c", "-a", "10", "-a", "20"))
		require.Nil(t, err)
		require.Equal(t, r, R{[]uint{10, 20}, true})
	}
}

func TestAdjacent_limits_commands(t *testing.T) {
	var x = Pure(struct{}{}).Parser().ToOptions().Command("a").Adjacent()
	var s = Short('s').Switch()
	var parser = Pair(s, x).Parser().ToOptions()

	r, err := parser.run_inner(newArgs("a", "-s"))
	require.Nil(t, err)
	require.Equal(t, r, pair(true, struct{}{}))
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

	{
		r, err := parser.run_inner(newArgs("sleep", "--time", "12", "-s"))
		require.Nil(t, err)
		require.Equal(t, r, pair(true, "12"))
	}

	_, r := parser.run_inner(newArgs("--help"))
	// TODO - this is ugly
	const expected = `Usage: [-s] COMMAND ...

Available options:
    -s
    -h, --help  Prints help information

Available commands:
    eat         eat something
    sleep       sleep for a bit
`
	require.Equal(t, unwrap_stdout(r), expected)
}

func TestTwo_adjacent_args(t *testing.T) {
	x := Argument(Short('x'), "X", parseUint)
	y := Argument(Short('y'), "Y", parseUint)
	c := Short('c').Switch()
	point := Pair(x, y).Adjacent()
	parser := Pair(point, c).Parser().ToOptions()

	{
		r, err := parser.run_inner(newArgs("-x", "3", "-y", "4", "-c"))
		require.Nil(t, err)
		require.Equal(t, r, pair(pair(uint(3), uint(4)), true))
	}
	{
		// they are adjacent to each other, but the way it is coded currently - they must be adjacent
		// to the first element.
		// Proper fix is to split "adjacent" into "adjacent to" and "adjacent block"

		// r,err:= parser.run_inner(newArgs("-y", "3", "-x", "4", "-c")).unwrap();
		// require.Equal(t, r, ((4, 3), true));
	}
	{
		_, r := parser.run_inner(newArgs("-y", "3", "-c", "-x", "4"))
		require.Equal(t, unwrap_stderr(r), "expected `-y=Y`, pass `--help` for usage information")
	}
}
