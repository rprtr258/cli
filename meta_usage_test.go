package cli

import (
	"testing"
)

func TestOptional_group_meta(t *testing.T) {
	a := Argument(Short('a'), "A", parseString)
	b := Argument(Short('b'), "B", parseString)
	parser := Pair(a, b).Parser().Optional().ToOptions()

	assert_usage(t, parser, "[-a=A -b=B]")
}

func TestSensors_many(t *testing.T) {
	a := Argument(Short('a'), "A", parseString)
	b := Argument(Short('b'), "B", parseString)
	parser := Pair(a, b).Parser().Many().ToOptions()
	assert_usage(t, parser, "[-a=A -b=B]...")
}

func TestSensors_some(t *testing.T) {
	a := Argument(Short('a'), "A", parseString)
	b := Argument(Short('b'), "B", parseString)
	parser := Pair(a, b).Parser().Some("want some sensors").ToOptions()
	assert_usage(t, parser, "(-a=A -b=B)...")
}

func TestMany_arg(t *testing.T) {
	parser := Argument(Short('a'), "A", parseString).Parser().Many().ToOptions()
	assert_usage(t, parser, "[-a=A]...")
}

func TestSome_arg(t *testing.T) {
	parser := Argument(Short('a'), "A", parseString).Parser().Some("ARG").ToOptions()
	assert_usage(t, parser, "(-a=A)...")
}

func TestMany_switch(t *testing.T) {
	parser := Short('a').Switch().Many().ToOptions()
	assert_usage(t, parser, "-a...")
}

func TestSome_switch(t *testing.T) {
	t.SkipNow() // TODO: fix
	parser := Short('a').Switch().Some("ARG").ToOptions()
	assert_usage(t, parser, "[-a]...")
}

func TestOptional_req_select(t *testing.T) {
	t.SkipNow() // TODO: fix
	a := ReqFlag(Short('a'), struct{}{})
	b := ReqFlag(Short('b'), struct{}{})
	parser := Choice(a, b).Optional().ToOptions()

	assert_usage(t, parser, "[-a | -b]")
}

func TestSingle_optional_req_select(t *testing.T) {
	a := ReqFlag(Short('a'), struct{}{})
	parser := Choice(a).Optional().ToOptions()

	assert_usage(t, parser, "[-a]")
}

func TestFallback_req_select(t *testing.T) {
	t.SkipNow() // TODO: fix
	a := ReqFlag(Short('a'), struct{}{})
	b := ReqFlag(Short('b'), struct{}{})
	parser := Choice(a, b).Fallback(struct{}{}).ToOptions()

	assert_usage(t, parser, "[-a | -b]")
}

func TestSingle_fallback_req_select(t *testing.T) {
	a := ReqFlag(Short('a'), struct{}{})
	parser := Choice(a).Fallback(struct{}{}).ToOptions()

	assert_usage(t, parser, "[-a]")
}

func TestOptional_argument_select(t *testing.T) {
	t.SkipNow() // TODO: fix
	a := Argument(Short('a'), "A", parseString)
	b := Argument(Short('b'), "B", parseString)
	parser := Choice(a, b).Optional().ToOptions()

	assert_usage(t, parser, "[-a=A | -b=B]")
}

func TestCommands_no_fallback_meta(t *testing.T) {
	a := Pure(struct{}{}).ToOptions().Command("a")
	b := Pure(struct{}{}).ToOptions().Command("b")
	parser := Choice(a, b).ToOptions()

	assert_usage(t, parser, "COMMAND ...")
}

func TestCommands_and_fallback_meta(t *testing.T) {
	t.SkipNow() // TODO: fix
	a := Pure(struct{}{}).ToOptions().Command("a")
	b := Pure(struct{}{}).ToOptions().Command("b")
	parser := Choice(a, b).Fallback(struct{}{}).ToOptions()

	assert_usage(t, parser, "[COMMAND ...]")
}

func TestCommand_no_fallback(t *testing.T) {
	a := Pure(struct{}{}).ToOptions().Command("a")
	parser := a.ToOptions()

	assert_usage(t, parser, "COMMAND ...")
}

func TestCommand_and_fallback_meta(t *testing.T) {
	a := Pure(struct{}{}).ToOptions().Command("a")
	parser := a.Fallback(struct{}{}).ToOptions()

	assert_usage(t, parser, "[COMMAND ...]")
}

func TestRequierd_select(t *testing.T) {
	a := ReqFlag(Short('a'), struct{}{})
	b := ReqFlag(Short('b'), struct{}{})
	parser := Choice(a, b).ToOptions()

	assert_usage(t, parser, "(-a | -b)")
}

func TestRequierd_and(t *testing.T) {
	a := ReqFlag(Short('a'), struct{}{})
	b := ReqFlag(Short('b'), struct{}{})
	parser := Pair(a, b).ToOptions()

	assert_usage(t, parser, "-a -b")
}

func TestRequired_or_and(t *testing.T) {
	a := ReqFlag(Short('a'), struct{}{})
	b := ReqFlag(Short('b'), struct{}{})
	c := ReqFlag(Short('c'), struct{}{})
	d := ReqFlag(Short('d'), struct{}{})
	ab := Pair(a, b)
	cd := Pair(c, d)
	parser := Choice(ab, cd).ToOptions()
	assert_usage(t, parser, "(-a -b | -c -d)")
}

func TestRequired_one_many(t *testing.T) {
	a := ReqFlag(Short('a'), struct{}{})
	b := ReqFlag(Short('b'), struct{}{})
	parser := Pair(a, b).Parser().Many().ToOptions()
	assert_usage(t, parser, "[-a -b]...")
}

func TestOptional_one_many(t *testing.T) {
	t.SkipNow() // TODO: fix
	a := ReqFlag(Short('a'), struct{}{})
	b := ReqFlag(Short('b'), struct{}{})
	parser := Pair(a, b).Parser().Optional().Many().ToOptions()
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
	parser := Choice(ab, cd, ef).Many().ToOptions()
	assert_usage(t, parser, "-a -b | -c -d...")
}

func TestNo_actual_arguments_also_works(t *testing.T) {
	t.SkipNow() // TODO: fix
	parser := Pure(true).ToOptions()

	assertHelp(t, parser, "Usage: \n\nAvailable options:\n    -h, --help  Prints help information\n")
}

func TestA_or_b(t *testing.T) {
	a := Argument(Short('a').Long("aaa"), "A", parseString)
	b := Argument(Short('b').Long("bbb"), "B", parseString)
	parser := Choice(a, b).ToOptions()
	assert_usage(t, parser, "(-a=A | -b=B)")
}

func TestSome_req_flag(t *testing.T) {
	t.SkipNow() // TODO: fix
	parser := ReqFlag(Short('a'), struct{}{}).Some("some").ToOptions()
	assert_usage(t, parser, "-a...")
}

func TestA_or_b_and_c(t *testing.T) {
	a := Argument(Short('a').Long("aaa"), "A", parseString)
	b := Argument(Short('b').Long("bbb"), "B", parseString)
	ab := Choice(a, b)
	c := Positional("C", parseString)
	parser := Pair(ab, c).ToOptions()
	assert_usage(t, parser, "(-a=A | -b=B) C")
}

func TestA_or_b_opt(t *testing.T) {
	t.SkipNow() // TODO: fix
	a := Argument(Short('a').Long("aaa"), "A", parseString)
	b := Argument(Short('b').Long("bbb"), "B", parseString)
	parser := Choice(a, b).Optional().ToOptions()
	assert_usage(t, parser, "[-a=A | -b=B]")
}

func TestA_or_b_opt_and_c(t *testing.T) {
	t.SkipNow() // TODO: fix
	a := Argument(Short('a').Long("aaa"), "A", parseString)
	b := Argument(Short('b').Long("bbb"), "B", parseString)
	ab := Choice(a, b).Optional()
	c := Positional("C", parseString)
	parser := Pair(ab, c).ToOptions()
	assert_usage(t, parser, "[-a=A | -b=B] C")
}

func TestAny_in_adjacent(t *testing.T) {
	a := ReqFlag(Short('a'), struct{}{})
	b := Any("A", func(s string) (string, bool) { return s, true }, false)
	parser := Pair(a, b).Adjacent().ToOptions()

	assert_usage(t, parser, "-a A")
}

func TestPositionals_in_branches_are_okay(t *testing.T) {
	a := Argument(Short('a'), "A", parseString)
	b := Argument(Short('b'), "B", parseString)
	c := Positional("C", parseString)
	d := Positional("D", parseString)

	ac := Pair(a, c)
	bd := Pair(b, d)
	parser := Choice(ac, bd).ToOptions()
	assert_usage(t, parser, "(-a=A C | -b=B D)")
}

func TestOptional_last_of_many(t *testing.T) {
	t.SkipNow() // TODO: fix
	a := ReqFlag(Short('a'), true)
	b := ReqFlag(Short('b'), false)
	ab := Choice(a, b).Last().Fallback(false)
	parser := ab.ToOptions()

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
	b := Short('b').Switch().CustomUsage(buf)

	parser := Pair(a, b).ToOptions()

	assert_usage(t, parser, "[-a] <also takes flag b>")
}

func TestDouble_strict(t *testing.T) {
	t.SkipNow() // TODO: fix
	a := Short('a').Switch()
	b := Positional("B", parseUint).Strict()
	c := Positional("C", parseUint).Strict()
	parser := Map3(a, b, c, map3).ToOptions()
	assert_usage(t, parser, "[-a] -- B C")
}

func TestOptional_strict(t *testing.T) {
	t.SkipNow() // TODO: fix
	parser := Positional("A", parseUint).Strict().Parser().Optional().ToOptions()
	assert_usage(t, parser, "-- [A]")
}

func TestMany_strict(t *testing.T) {
	t.SkipNow() // TODO: fix
	a := Short('a').Switch()
	b := Positional("B", parseUint).Strict().Parser().Many()
	parser := Pair(a, b).ToOptions()
	assert_usage(t, parser, "[-a] -- [B]...")
}

func TestQuadruple_strict(t *testing.T) {
	t.SkipNow() // TODO: fix
	a := Short('a').Switch()
	b := Positional("B", parseUint).Strict()
	c := Positional("C", parseUint).Strict()
	abc := Map3(a, b, c, map3)

	d := Short('d').Switch()
	e := Positional("E", parseUint).Strict()
	f := Positional("F", parseUint).Strict()
	def := Map3(d, e, f, map3)

	parser := Choice(abc, def).ToOptions()
	assert_usage(t, parser, "([-a] -- B C | [-d] -- E F)")
}

func TestQuadruple_strict_many(t *testing.T) {
	t.SkipNow() // TODO: fix
	a := Short('a').Switch()
	b := Positional("B", parseUint).Strict().Parser().Many()
	ab := Pair(a, b)

	d := Short('d').Switch()
	f := Positional("F", parseUint).Strict().Parser().Many()
	df := Pair(d, f)

	parser := Choice(ab, df).ToOptions()
	assert_usage(t, parser, "([-a] -- [B]... | [-d] -- [F]...)")
}
