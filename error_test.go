package cli

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestThis_or_that_odd(t *testing.T) {
	ab := Pair(
		ReqFlag(Short('a'), struct{}{}),
		ReqFlag(Short('b'), struct{}{}),
	)
	cd := Pair(
		ReqFlag(Short('a'), struct{}{}),
		ReqFlag(Short('c'), struct{}{}),
	)
	parser := Choice(ab, cd).ToOptions()

	_, res := parser.run_inner(newArgs("-a", "-b", "-c"))

	require.Equal(t, unwrap_stderr(res), "`-c` cannot be used at the same time as `-b`")
}

func TestNo_argument(t *testing.T) {
	a := Argument(Short('a'), "N", parseInt32)
	b := Short('2').Switch()
	parser := Pair(a, b).Parser().ToOptions()

	{
		r, err := parser.run_inner(newArgs("-a", "-42"))
		require.Nil(t, err)
		require.Equal(t, r, pair(-42, false))
	}

	{
		//    r, err := parser.run_inner(newArgs("-a", "-4"]))
		//    require.Nil(t, err)
		//    require.Equal(t, r, (-4, flse));
	}

	{
		_, r := parser.run_inner(newArgs("-a", "-2"))
		// can we restore this?
		// "`-a` requires an argument `N`, got a flag `-2`, try `-a=-2` to use it as an argument"
		require.Equal(t, unwrap_stderr(r), "`-a` requires an argument `N`")
	}
}

func TestCannot_be_used_partial_arg(t *testing.T) {
	a := ReqFlag[uint](Short('a'), 10)
	b := Argument(Short('b'), "ARG", parseUint)
	parser := Choice(a, b).ToOptions()

	{
		// TODO - error message can be improved...
		_, res := parser.run_inner(newArgs("-b", "-a"))
		require.Equal(t, unwrap_stderr(res), "`-b` is not expected in this context")
	}

	{
		_, res := parser.run_inner(newArgs("-a", "-b"))
		require.Equal(t, unwrap_stderr(res), "`-b` is not expected in this context")
	}
}

// func TestBetter_error_with_enum(t*testing.T) {
//     #[derive(Bpaf)]
//     #[bpaf(options)]
//     enum Foo {
//         Alpha,
//         Beta,
//         Gamma,
//     }

//     res := foo().
//         run_inner(newArgs("--alpha", "--beta"]).
//         unwrap_err().
//         unwrap_stderr();
//     require.Equal(t, res, "`--beta` cannot be used at the same time as `--alpha`");

//     res := foo().
//         run_inner(newArgs("--alpha", "--gamma"]).
//         unwrap_err().
//         unwrap_stderr();
//     require.Equal(t,
//         res,
//         "`--gamma` cannot be used at the same time as `--alpha`"
//     );

//     res := foo().
//         run_inner(newArgs("--beta", "--gamma"]).
//         unwrap_err().
//         unwrap_stderr();
//     require.Equal(t, res, "`--gamma` cannot be used at the same time as `--beta`");

//     res := foo().
//         run_inner(newArgs("--alpha", "--beta", "--gamma"]).
//         unwrap_err().
//         unwrap_stderr();
//     require.Equal(t, res, "`--beta` cannot be used at the same time as `--alpha`");
// }

func TestGuard_message(t *testing.T) {
	parser := Argument(Short('a'),
		"N", parseUint32).
		Parser().Guard(func(n uint32) bool { return n <= 10 }, "too high").
		Parser().ToOptions()

	_, res := parser.run_inner(newArgs("-a", "30"))
	require.Equal(t, unwrap_stderr(res), "`30`: too high")
}

func TestStrict_positional_argument(t *testing.T) {
	a := Argument(Short('a'), "N", parseUint)
	parser := a.Parser().ToOptions()

	_, r := parser.run_inner(newArgs("-a", "--", "10"))
	require.Equal(t, unwrap_stderr(r), "`-a` requires an argument `N`")
}

func TestNot_expected_at_all(t *testing.T) {
	a := Short('a').Switch()
	parser := a.Parser().ToOptions()

	{
		_, r := parser.run_inner(newArgs("--megapotato"))
		require.Equal(t, unwrap_stderr(r), "`--megapotato` is not expected in this context")
	}

	{
		_, r := parser.run_inner(newArgs("megapotato"))
		require.Equal(t, unwrap_stderr(r), "`megapotato` is not expected in this context")
	}
}

func TestCannot_be_used_twice(t *testing.T) {
	a := Short('a').Switch()
	b := Many(Short('b').Switch().Parser())
	parser := Pair(a, b).Parser().ToOptions()

	{
		_, r := parser.run_inner(newArgs("-a", "-b", "-a"))
		require.Equal(t, "argument `-a` cannot be used multiple times in this context", unwrap_stderr(r))
	}

	{
		_, r := parser.run_inner(newArgs("-a", "-a"))
		require.Equal(t, "argument `-a` cannot be used multiple times in this context", unwrap_stderr(r))
	}

	{
		_, r := parser.run_inner(newArgs("-abba"))
		require.Equal(t, "argument `-a` cannot be used multiple times in this context", unwrap_stderr(r))
	}
}

func TestShould_not_split_adjacent_options(t *testing.T) {
	a := ReqFlag(Short('a'), 0)
	b := Pure(struct{}{}).Parser().ToOptions().Command("hello")
	parser := Pair(a, b).Parser().ToOptions()
	_, r := parser.run_inner(newArgs("-ahello"))
	// can probably suggest splitting here too: `-a` `hello`
	const expected = "no such flag: `-ahello`, did you mean `hello`?"
	require.Equal(t, unwrap_stderr(r), expected)
}

func TestShould_not_split_adjacent_ambig_options(t *testing.T) {
	a1 := ReqFlag[uint](Short('a'), 0)
	a2 := Argument(Short('a'), "x", parseUint)
	a := Choice(a2, a1)
	c := Pure(struct{}{}).Parser().ToOptions().Command("hello")
	parser := Pair(a, c).Parser().ToOptions()

	t.Run("TODO: fix", func(t *testing.T) {
		t.Skip() // TODO: fix
		_, r := parser.run_inner(newArgs("-a=hello"))
		const expected = "expected `COMMAND ...`, got `hello`. Pass `--help` for usage information"
		require.Equal(t, unwrap_stderr(r), expected)
	})

	{
		_, r := parser.run_inner(newArgs("-ahello"))
		const expected = "app supports `-a` as both an option and an option-argument, try to split `-ahello` into individual\noptions (-a -h ..) or use `-a=hello` syntax to disambiguate"
		require.Equal(t, expected, unwrap_stderr(r))
	}

	{
		// this one is okay, try to parse -a as argument - it fails because "hello" is not a number, then
		// try to parse -a as a flag - this works
		r, err := parser.run_inner(newArgs("-a", "hello"))
		require.Nil(t, err)
		require.Equal(t, r, pair(0, struct{}{}))
	}
}

func TestAdjacent_option_complains_to(t *testing.T) {
	parser := Argument(Short('a'), "A", parseUint).Parser().ToOptions()

	_, r := parser.run_inner(newArgs("-ayam"))
	// TODO - this should point to the whole "-ayam" thing
	require.Equal(t, unwrap_stderr(r), "couldn't parse `yam`: invalid digit found in string")
}

func TestSome_pos_with_invalid_flag(t *testing.T) {
	a := Short('a').Switch()
	b := ParseSomeParser(Positional("B", parseUint).Parser().Some("Want B"))
	parser := Pair(a, b).Parser().ToOptions()

	{
		_, r := parser.run_inner(newArgs("-c", "12"))
		require.Equal(t, unwrap_stderr(r), "`-c` is not expected in this context")
	}

	{
		_, r := parser.run_inner(newArgs("12", "-c"))
		require.Equal(t, unwrap_stderr(r), "`-c` is not expected in this context")
	}
}

func TestPos_with_invalid_arg(t *testing.T) {
	a := Optional(Argument(Short('a'), "A", parseUint).Parser())
	b := Positional("B", parseUint)
	parser := Pair(a, b).Parser().ToOptions()

	{
		_, r := parser.run_inner(newArgs("-c", "12"))
		require.Equal(t, unwrap_stderr(r), "`-c` is not expected in this context")
	}

	{
		_, r := parser.run_inner(newArgs("12", "-c"))
		require.Equal(t, unwrap_stderr(r), "`-c` is not expected in this context")
	}

	{
		_, r := parser.run_inner(newArgs("-c", "t"))
		require.Equal(t, unwrap_stderr(r), "couldn't parse `t`: invalid digit found in string")
	}

	{
		_, r := parser.run_inner(newArgs("t", "-c"))
		require.Equal(t, unwrap_stderr(r), "couldn't parse `t`: invalid digit found in string")
	}
}

func TestStrictly_positional_help(t *testing.T) {
	parser := Long("hhhh").Switch().Parser().ToOptions()

	_, r := parser.run_inner(newArgs("--", "--help"))
	require.Equal(t, unwrap_stderr(r), "`--help` is not expected in this context")
}

func TestHidden_required_field_is_valid_but_strange(t *testing.T) {
	parser := ReqFlag(Short('a'), struct{}{}).Parser().Hide().ToOptions()

	_, r := parser.run_inner(newArgs())
	require.Equal(t, unwrap_stderr(r), "parser requires an extra flag, argument or parameter, but its name is hidden by the author")
}

func TestGuard_on_fallback(t *testing.T) {
	parser := Argument(Short('a'),
		"A", parseUint).
		Parser().Fallback(10).
		Parser().Guard(func(a uint) bool { return a < 10 }, "too big").
		Parser().ToOptions()
	_, r := parser.run_inner(newArgs())
	require.Equal(t, unwrap_stderr(r), "check failed: too big")
}

func TestTwo_required_fields_first_missing(t *testing.T) {
	a := Argument(Long("a"), "A", parseUint32)
	b := Argument(Long("b"), "B", parseUint32)
	parser := Pair(a, b).Parser().ToOptions()
	_, r := parser.run_inner(newArgs("--b", "1"))
	require.Equal(t, unwrap_stderr(r), "expected `--a=A`, pass `--help` for usage information")
}

func TestUsed_only_once_is_more_important_error(t *testing.T) {
	format := Long("format").Switch().Parser()
	sort := Long("sort").Switch().Parser()
	filter := Long("filter").Switch().Parser()
	opts := Map3(format, sort, filter, func(format, sort, filter bool) struct{} { return struct{}{} }).ToOptions()

	{
		_, err := opts.run_inner(newArgs("--filter", "--filter"))
		require.Equal(t,
			unwrap_stderr(err),
			"argument `--filter` cannot be used multiple times in this context",
		)
	}

	{
		_, err := opts.run_inner(newArgs("--sort", "--sort"))
		require.Equal(t,
			unwrap_stderr(err),
			"argument `--sort` cannot be used multiple times in this context",
		)
	}
}
