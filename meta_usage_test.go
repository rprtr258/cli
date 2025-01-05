package cli_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	. "github.com/rprtr258/cli"
)

func assert_usage[T any](t *testing.T, parser *OptionParser[T], expected string) {
	t.Helper()

	_, output := parser.RunInner(NewArgs("--help"))
	usage := strings.TrimPrefix(strings.Split(UnwrapStdout(output), "\n")[0], "Usage: ")
	require.Equal(t, expected, usage)
}

func TestOptional_group_meta(t *testing.T) {
	a := Argument(Short('a'), "A", ParseString)
	b := Argument(Short('b'), "B", ParseString)
	parser := Optional(Pair(a, b).Parser(), false).Parser().ToOptions()

	assert_usage(t, parser, "[-a=A -b=B]")
}

func TestSensors_many(t *testing.T) {
	a := Argument(Short('a'), "A", ParseString)
	b := Argument(Short('b'), "B", ParseString)
	parser := Many(Pair(a, b), false).Parser().ToOptions()
	assert_usage(t, parser, "[-a=A -b=B]...")
}

func TestSensors_some(t *testing.T) {
	a := Argument(Short('a'), "A", ParseString)
	b := Argument(Short('b'), "B", ParseString)
	parser := ParseSomeParser(Pair(a, b).Parser().Some("want some sensors", false)).ToOptions()
	assert_usage(t, parser, "(-a=A -b=B)...")
}

func TestMany_arg(t *testing.T) {
	parser := Many(Argument(Short('a'), "A", ParseString), false).Parser().ToOptions()
	assert_usage(t, parser, "[-a=A]...")
}

func TestSome_arg(t *testing.T) {
	parser := ParseSomeParser(Argument(Short('a'), "A", ParseString).Parser().Some("ARG", false)).ToOptions()
	assert_usage(t, parser, "(-a=A)...")
}

func TestMany_switch(t *testing.T) {
	t.SkipNow() // TODO: fix
	parser := Many(Short('a').Switch(), false).Parser().ToOptions()
	assert_usage(t, parser, "[-a]...")
}

func TestSome_switch(t *testing.T) {
	t.SkipNow() // TODO: fix
	parser := ParseSomeParser(Short('a').Switch().Parser().Some("ARG", false)).ToOptions()
	assert_usage(t, parser, "[-a]...")
}

func TestOptional_req_select(t *testing.T) {
	t.SkipNow() // TODO: fix
	a := ReqFlag(Short('a'), struct{}{})
	b := ReqFlag(Short('b'), struct{}{})
	parser := Optional(Choice(a, b), false).Parser().ToOptions()

	assert_usage(t, parser, "[-a | -b]")
}

func TestSingle_optional_req_select(t *testing.T) {
	a := ReqFlag(Short('a'), struct{}{})
	parser := Optional(Choice(a), false).Parser().ToOptions()

	assert_usage(t, parser, "[-a]")
}

func TestFallback_req_select(t *testing.T) {
	t.SkipNow() // TODO: fix
	a := ReqFlag(Short('a'), struct{}{})
	b := ReqFlag(Short('b'), struct{}{})
	parser := Choice(a, b).Parser().Fallback(struct{}{}).Parser().ToOptions()

	assert_usage(t, parser, "[-a | -b]")
}

func TestSingle_fallback_req_select(t *testing.T) {
	a := ReqFlag(Short('a'), struct{}{})
	parser := Choice(a).Parser().Fallback(struct{}{}).Parser().ToOptions()

	assert_usage(t, parser, "[-a]")
}

func TestOptional_argument_select(t *testing.T) {
	t.SkipNow() // TODO: fix
	a := Argument(Short('a'), "A", ParseString)
	b := Argument(Short('b'), "B", ParseString)
	parser := Optional(Choice(a, b), false).Parser().ToOptions()

	assert_usage(t, parser, "[-a=A | -b=B]")
}

func TestCommands_no_fallback_meta(t *testing.T) {
	a := Pure(struct{}{}).Parser().ToOptions().Command("a")
	b := Pure(struct{}{}).Parser().ToOptions().Command("b")
	parser := Choice(a, b).Parser().ToOptions()

	assert_usage(t, parser, "COMMAND ...")
}

func TestCommands_and_fallback_meta(t *testing.T) {
	t.SkipNow() // TODO: fix
	a := Pure(struct{}{}).Parser().ToOptions().Command("a")
	b := Pure(struct{}{}).Parser().ToOptions().Command("b")
	parser := Choice(a, b).Parser().Fallback(struct{}{}).Parser().ToOptions()

	assert_usage(t, parser, "[COMMAND ...]")
}

func TestCommand_no_fallback(t *testing.T) {
	a := Pure(struct{}{}).Parser().ToOptions().Command("a")
	parser := a.Parser().ToOptions()

	assert_usage(t, parser, "COMMAND ...")
}

func TestCommand_and_fallback_meta(t *testing.T) {
	a := Pure(struct{}{}).Parser().ToOptions().Command("a")
	parser := a.Parser().Fallback(struct{}{}).Parser().ToOptions()

	assert_usage(t, parser, "[COMMAND ...]")
}

func TestRequierd_select(t *testing.T) {
	a := ReqFlag(Short('a'), struct{}{})
	b := ReqFlag(Short('b'), struct{}{})
	parser := Choice(a, b).Parser().ToOptions()

	assert_usage(t, parser, "(-a | -b)")
}

func TestRequierd_and(t *testing.T) {
	a := ReqFlag(Short('a'), struct{}{})
	b := ReqFlag(Short('b'), struct{}{})
	parser := Pair(a, b).Parser().ToOptions()

	assert_usage(t, parser, "-a -b")
}

func TestRequired_or_and(t *testing.T) {
	a := ReqFlag(Short('a'), struct{}{})
	b := ReqFlag(Short('b'), struct{}{})
	c := ReqFlag(Short('c'), struct{}{})
	d := ReqFlag(Short('d'), struct{}{})
	ab := Pair(a, b)
	cd := Pair(c, d)
	parser := Choice(ab, cd).Parser().ToOptions()
	assert_usage(t, parser, "(-a -b | -c -d)")
}

func TestRequired_one_many(t *testing.T) {
	a := ReqFlag(Short('a'), struct{}{})
	b := ReqFlag(Short('b'), struct{}{})
	parser := Many(Pair(a, b), false).Parser().ToOptions()
	assert_usage(t, parser, "[-a -b]...")
}

func TestOptional_one_many(t *testing.T) {
	t.SkipNow() // TODO: fix
	a := ReqFlag(Short('a'), struct{}{})
	b := ReqFlag(Short('b'), struct{}{})
	parser := Many(Optional(Pair(a, b).Parser(), false), false).Parser().ToOptions()
	assert_usage(t, parser, "[-a -b]...")
}

func TestRequired_or_many(t *testing.T) {
	a := ReqFlag(Short('a'), struct{}{})
	b := ReqFlag(Short('b'), struct{}{})
	c := ReqFlag(Short('c'), struct{}{})
	d := ReqFlag(Short('d'), struct{}{})
	ab := Pair(a, b)
	cd := Pair(c, d)
	e := Pure(PPair(struct{}{}, struct{}{}))
	f := Pure(PPair(struct{}{}, struct{}{}))
	ef := Choice(e, f)
	parser := Many(Choice(ab, cd, ef), false).Parser().ToOptions()
	assert_usage(t, parser, "[-a -b | -c -d]...")
}

func TestNo_actual_arguments_also_works(t *testing.T) {
	t.SkipNow() // TODO: fix
	parser := Pure(true).Parser().ToOptions()

	_, r := parser.RunInner(NewArgs("--help"))
	require.Equal(t,
		UnwrapStdout(r),
		"Usage: \n\nAvailable options:\n    -h, --help  Prints help information\n",
	)
}

func TestA_or_b(t *testing.T) {
	a := Argument(Short('a').Long("aaa"), "A", ParseString)
	b := Argument(Short('b').Long("bbb"), "B", ParseString)
	parser := Choice(a, b).Parser().ToOptions()
	assert_usage(t, parser, "(-a=A | -b=B)")
}

func TestSome_req_flag(t *testing.T) {
	t.SkipNow() // TODO: fix
	parser := ParseSomeParser(ReqFlag(Short('a'), struct{}{}).Parser().Some("some", false)).ToOptions()
	assert_usage(t, parser, "-a...")
}

func TestA_or_b_and_c(t *testing.T) {
	a := Argument(Short('a').Long("aaa"), "A", ParseString)
	b := Argument(Short('b').Long("bbb"), "B", ParseString)
	ab := Choice(a, b)
	c := Positional("C", ParseString)
	parser := Pair(ab, c).Parser().ToOptions()
	assert_usage(t, parser, "(-a=A | -b=B) C")
}

func TestA_or_b_opt(t *testing.T) {
	t.SkipNow() // TODO: fix
	a := Argument(Short('a').Long("aaa"), "A", ParseString)
	b := Argument(Short('b').Long("bbb"), "B", ParseString)
	parser := Optional(Choice(a, b).Parser(), false).Parser().ToOptions()
	assert_usage(t, parser, "[-a=A | -b=B]")
}

func TestA_or_b_opt_and_c(t *testing.T) {
	t.SkipNow() // TODO: fix
	a := Argument(Short('a').Long("aaa"), "A", ParseString)
	b := Argument(Short('b').Long("bbb"), "B", ParseString)
	ab := Optional(Choice(a, b).Parser(), false)
	c := Positional("C", ParseString)
	parser := Pair(ab, c).Parser().ToOptions()
	assert_usage(t, parser, "[-a=A | -b=B] C")
}

func TestAny_in_adjacent(t *testing.T) {
	a := ReqFlag(Short('a'), struct{}{})
	b := Any("A", func(s string) (string, bool) { return s, true }, false)
	parser := Pair(a, b).Adjacent().Parser().ToOptions()

	assert_usage(t, parser, "-a A")
}

func TestPositionals_in_branches_are_okay(t *testing.T) {
	a := Argument(Short('a'), "A", ParseString)
	b := Argument(Short('b'), "B", ParseString)
	c := Positional("C", ParseString)
	d := Positional("D", ParseString)

	ac := Pair(a, c)
	bd := Pair(b, d)
	parser := Choice(ac, bd).Parser().ToOptions()
	assert_usage(t, parser, "(-a=A C | -b=B D)")
}

func TestOptional_last_of_many(t *testing.T) {
	t.SkipNow() // TODO: fix
	a := ReqFlag(Short('a'), true)
	b := ReqFlag(Short('b'), false)
	ab := Choice(a, b).Last().Parser().Fallback(false)
	parser := ab.Parser().ToOptions()

	assert_usage(t, parser, "[-a | -b]...")
}

// func TestHidden_fallback_branch(t*testing.T) {
//     #[derive(Bpaf)]
//     struct Fallback {
//         #[bpaf(positional("COMMAND"))]
//         name: String,
//     }

//     #[derive(Bpaf)]
//     #[bpaf(options)]
//     enum Commands {
//         #[bpaf(command)]
//         Build {},
//         Fallback(#[bpaf(external(fallback), hide)] Fallback),
//     }

//     assert_usage(t,commands(), "COMMAND ...");
// }

func TestCustom_usage(t *testing.T) {
	a := Short('a').Switch()

	var buf Doc
	buf.Text("<also takes flag b>")
	b := Short('b').Switch().Parser().CustomUsage(buf)

	parser := Pair(a, b).Parser().ToOptions()

	assert_usage(t, parser, "[-a] <also takes flag b>")
}

func TestDouble_strict(t *testing.T) {
	t.SkipNow() // TODO: fix
	a := Short('a').Switch()
	b := Positional("B", ParseUint).Strict()
	c := Positional("C", ParseUint).Strict()
	parser := Map3(a, b, c, MMap3).Parser().ToOptions()
	assert_usage(t, parser, "[-a] -- B C")
}

func TestOptional_strict(t *testing.T) {
	t.SkipNow() // TODO: fix
	parser := Optional(Positional("A", ParseUint).Strict().Parser(), false).Parser().ToOptions()
	assert_usage(t, parser, "-- [A]")
}

func TestMany_strict(t *testing.T) {
	t.SkipNow() // TODO: fix
	a := Short('a').Switch()
	b := Many(Positional("B", ParseUint).Strict(), false)
	parser := Pair(a, b).Parser().ToOptions()
	assert_usage(t, parser, "[-a] -- [B]...")
}

func TestQuadruple_strict(t *testing.T) {
	t.SkipNow() // TODO: fix
	a := Short('a').Switch()
	b := Positional("B", ParseUint).Strict()
	c := Positional("C", ParseUint).Strict()
	abc := Map3(a, b, c, MMap3)

	d := Short('d').Switch()
	e := Positional("E", ParseUint).Strict()
	f := Positional("F", ParseUint).Strict()
	def := Map3(d, e, f, MMap3)

	parser := Choice(abc, def).Parser().ToOptions()
	assert_usage(t, parser, "([-a] -- B C | [-d] -- E F)")
}

func TestQuadruple_strict_many(t *testing.T) {
	t.SkipNow() // TODO: fix
	a := Short('a').Switch()
	b := Many(Positional("B", ParseUint).Strict(), false)
	ab := Pair(a, b)

	d := Short('d').Switch()
	f := Many(Positional("F", ParseUint).Strict(), false)
	df := Pair(d, f)

	parser := Choice(ab, df).Parser().ToOptions()
	assert_usage(t, parser, "([-a] -- [B]... | [-d] -- [F]...)")
}
