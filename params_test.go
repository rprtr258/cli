package cli_test

import (
	"testing"

	"github.com/rprtr258/fun"
	"github.com/stretchr/testify/require"

	. "github.com/rprtr258/cli"
)

func some(s string) (string, bool) { return s, true }

func TestGet_any_simple(t *testing.T) {
	a := Short('a').Switch()
	b := Any("REST", some, false).Help("any help")
	parser := Pair(a, b).Parser().ToOptions()

	type R = fun.Pair[bool, string]
	AssertSuccess(t, R{true, "-b"}, parser, "-a", "-b")
	AssertSuccess(t, R{true, "-b"}, parser, "-b", "-a")
	AssertSuccess(t, R{true, "-b=foo"}, parser, "-b=foo", "-a")
}

func TestGet_any_many(t *testing.T) {
	a := Short('a').Switch()
	b := Many(Any("REST", some, false).Help("any help"), false)
	parser := Pair(a, b).Parser().ToOptions()

	type R = fun.Pair[bool, []string]
	AssertSuccess(t, R{true, []string{"-b"}}, parser, "-a", "-b")
	AssertSuccess(t, R{true, []string{"-b"}}, parser, "-b", "-a")
	AssertSuccess(t, R{true, []string{"-b", "-b"}}, parser, "-b", "-a", "-b")
}

func TestGet_any_many2(t *testing.T) {
	parser := Many(Any("REST", some, false), false).Parser().ToOptions()
	AssertSuccess(t, []string{"-vvv"}, parser, "-vvv")
}

func TestGet_any_magic(t *testing.T) {
	parser := Map(Optional(Argument(Short('b'),
		"anana", parseString).
		Adjacent().Parser().
		Guard(func(b string) bool { return b == "anana" }, "not anana").Parser(), true).
		Parser(),
		func(b Option[string]) bool { return b.Valid }).
		ToOptions()

	// -b anana - isn't a -banana
	_, r := parser.RunInner(NewArgs("-b", "anana"))
	require.Equal(t, UnwrapStderr(r), "`-b` is not expected in this context")

	// close enough :)
	AssertSuccess(t, true, parser, "-b=anana")
	AssertSuccess(t, true, parser, "-banana")
	AssertSuccess(t, false, parser)
}

func TestFrom_str_works_with_parse(t *testing.T) {
	parser := Parse(Positional("A", parseString).Parser(),
		parseUint).Parser().
		ToOptions()
	AssertSuccess(t, uint(42), parser, "42")
}

func TestSquashed_short_names(t *testing.T) {
	a := Short('a').Switch()
	b := Argument(Short('b'), "B", parseString)
	parser := Pair(a, b).Parser().ToOptions()

	type R = fun.Pair[bool, string]
	// For now instead of supporting I parse "-Obits=2048" instead as "-O" with "bits=2048" body
	// https://github.com/pacak/bpaf/issues/121
	//
	// r := parser.run_inner(newArgs("-ab=foo")).unwrap();
	// require.Equal(t,r.V, "foo");

	// an odd one, short flag a is squashed with short argument b
	AssertSuccess(t, R{true, "foo"}, parser, "-ab", "foo")

	// same, but body for b is adjacent as well
	AssertSuccess(t, R{true, "foo"}, parser, "-abfoo")

	// normal
	AssertSuccess(t, R{true, "foo"}, parser, "-a", "-bfoo")
}

// func Testcommand_alias(t *testing.T) {
//     #[derive(Debug, Bpaf, Clone)]
//     #[bpaf(options)]
//     enum Groups {
//         #[bpaf(command("top"), long("top-alias"))]
//         Command,
//     }

//     groups().run_inner(newArgs("top")).unwrap();
//     groups().run_inner(newArgs("top-alias")).unwrap();
// }
