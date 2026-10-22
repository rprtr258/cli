package cli

import (
	"testing"

	"github.com/rprtr258/fun"
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

	assertFailure(t, parser, "`-c` cannot be used at the same time as `-b`", "-a", "-b", "-c")
}

func TestNo_argument(t *testing.T) {
	a := Argument(Short('a'), "N", parseInt32)
	b := Short('2').Switch()
	parser := Pair(a, b).ToOptions()

	assertSuccess(t, PPair(int32(-42), false), parser, "-a", "-42")

	{
		//    r, err := parser.RunInner(newArgs("-a", "-4"]))
		//    require.Nil(t, err)
		//    require.Equal(t, r, (-4, flse));
	}

	// can we restore this?
	// "`-a` requires an argument `N`, got a flag `-2`, try `-a=-2` to use it as an argument"
	assertFailure(t, parser, "`-a` requires an argument `N`", "-a", "-2")
}

func TestCannot_be_used_partial_arg(t *testing.T) {
	a := ReqFlag[uint](Short('a'), 10)
	b := Argument(Short('b'), "ARG", parseUint)
	parser := Choice(a, b).ToOptions()

	// TODO - error message can be improved...
	assertFailure(t, parser, "`-b` is not expected in this context", "-b", "-a")

	assertFailure(t, parser, "`-b` is not expected in this context", "-a", "-b")
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
//         RunInner(newArgs("--alpha", "--beta"]).
//         unwrap_err().
//         UnwrapStderr();
//     require.Equal(t, res, "`--beta` cannot be used at the same time as `--alpha`");

//     res := foo().
//         RunInner(newArgs("--alpha", "--gamma"]).
//         unwrap_err().
//         UnwrapStderr();
//     require.Equal(t,
//         res,
//         "`--gamma` cannot be used at the same time as `--alpha`"
//     );

//     res := foo().
//         RunInner(newArgs("--beta", "--gamma"]).
//         unwrap_err().
//         UnwrapStderr();
//     require.Equal(t, res, "`--gamma` cannot be used at the same time as `--beta`");

//     res := foo().
//         RunInner(newArgs("--alpha", "--beta", "--gamma"]).
//         unwrap_err().
//         UnwrapStderr();
//     require.Equal(t, res, "`--beta` cannot be used at the same time as `--alpha`");
// }

func TestGuard_message(t *testing.T) {
	parser := Argument(Short('a'),
		"N", parseUint32).
		Parser().Guard(func(n uint32) bool { return n <= 10 }, "too high").
		Parser().ToOptions()

	assertFailure(t, parser, "`30`: too high", "-a", "30")
}

func TestStrict_positional_argument(t *testing.T) {
	a := Argument(Short('a'), "N", parseUint)
	parser := a.ToOptions()

	assertFailure(t, parser, "`-a` requires an argument `N`", "-a", "--", "10")
}

func TestNot_expected_at_all(t *testing.T) {
	a := Short('a').Switch()
	parser := a.ToOptions()

	assertFailure(t, parser, "`--megapotato` is not expected in this context", "--megapotato")

	assertFailure(t, parser, "`megapotato` is not expected in this context", "megapotato")
}

func TestCannot_be_used_twice(t *testing.T) {
	a := Short('a').Switch()
	b := Short('b').Switch().Many()
	parser := Pair(a, b).ToOptions()

	assertFailure(t, parser, "argument `-a` cannot be used multiple times in this context", "-a", "-b", "-a")

	assertFailure(t, parser, "argument `-a` cannot be used multiple times in this context", "-a", "-a")

	assertFailure(t, parser, "argument `-a` cannot be used multiple times in this context", "-abba")
}

func TestShould_not_split_adjacent_options(t *testing.T) {
	a := ReqFlag(Short('a'), 0)
	b := Pure(struct{}{}).ToOptions().Command("hello")
	parser := Pair(a, b).ToOptions()
	// can probably suggest splitting here too: `-a` `hello`
	const expected = "no such flag: `-ahello`, did you mean `hello`?"
	assertFailure(t, parser, expected, "-ahello")
}

func TestShould_not_split_adjacent_ambig_options(t *testing.T) {
	a1 := ReqFlag[uint](Short('a'), 0)
	a2 := Argument(Short('a'), "x", parseUint)
	a := Choice(a2, a1)
	c := Pure(struct{}{}).ToOptions().Command("hello")
	parser := Pair(a, c).ToOptions()

	// const expected = "expected `COMMAND ...`, got `hello`. Pass `--help` for usage information"
	assertSuccess(t, fun.Pair[uint, struct{}]{0, struct{}{}}, parser, "-a=hello")

	{
		const expected = "app supports `-a` as both an option and an option-argument, try to split `-ahello` into individual\noptions (-a -h ..) or use `-a=hello` syntax to disambiguate"
		assertFailure(t, parser, expected, "-ahello")
	}

	// this one is okay, try to parse -a as argument - it fails because "hello" is not a number, then
	// try to parse -a as a flag - this works
	assertSuccess(t, PPair(uint(0), struct{}{}), parser, "-a", "hello")
}

func TestAdjacent_option_complains_to(t *testing.T) {
	parser := Argument(Short('a'), "A", parseUint).ToOptions()

	// TODO - this should point to the whole "-ayam" thing
	assertFailure(t, parser, "couldn't parse `yam`: invalid digit found in string", "-ayam")
}

func TestSome_pos_with_invalid_flag(t *testing.T) {
	a := Short('a').Switch()
	b := Positional("B", parseUint).Parser().Some("Want B")
	parser := Pair(a, b).ToOptions()

	assertFailure(t, parser, "`-c` is not expected in this context", "-c", "12")

	assertFailure(t, parser, "`-c` is not expected in this context", "12", "-c")
}

func TestPos_with_invalid_arg(t *testing.T) {
	a := Argument(Short('a'), "A", parseUint).Parser().Optional()
	b := Positional("B", parseUint)
	parser := Pair(a, b).ToOptions()

	assertFailure(t, parser, "`-c` is not expected in this context", "-c", "12")

	assertFailure(t, parser, "`-c` is not expected in this context", "12", "-c")

	assertFailure(t, parser, "couldn't parse `t`: invalid digit found in string", "-c", "t")

	assertFailure(t, parser, "couldn't parse `t`: invalid digit found in string", "t", "-c")
}

func TestStrictly_positional_help(t *testing.T) {
	parser := Long("hhhh").Switch().ToOptions()

	assertFailure(t, parser, "`--help` is not expected in this context", "--", "--help")
}

func TestHidden_required_field_is_valid_but_strange(t *testing.T) {
	parser := ReqFlag(Short('a'), struct{}{}).Hide().ToOptions()

	assertFailure(t, parser, "parser requires an extra flag, argument or parameter, but its name is hidden by the author")
}

func TestGuard_on_fallback(t *testing.T) {
	parser := Argument(Short('a'),
		"A", parseUint).
		Parser().Fallback(10).
		Parser().Guard(func(a uint) bool { return a < 10 }, "too big").
		Parser().ToOptions()
	assertFailure(t, parser, "check failed: too big")
}

func TestTwo_required_fields_first_missing(t *testing.T) {
	a := Argument(Long("a"), "A", parseUint32)
	b := Argument(Long("b"), "B", parseUint32)
	parser := Pair(a, b).ToOptions()
	assertFailure(t, parser, "expected `--a=A`, pass `--help` for usage information", "--b", "1")
}

func TestUsed_only_once_is_more_important_error(t *testing.T) {
	format := Long("format").Switch()
	sort := Long("sort").Switch()
	filter := Long("filter").Switch()
	opts := Map3(format, sort, filter, func(format, sort, filter bool) struct{} { return struct{}{} }).ToOptions()

	assertFailure(t, opts, "argument `--filter` cannot be used multiple times in this context", "--filter", "--filter")

	assertFailure(t, opts, "argument `--sort` cannot be used multiple times in this context", "--sort", "--sort")
}
