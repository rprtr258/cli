package cli

import (
	"testing"

	"github.com/rprtr258/fun"
)

func some(s string) (string, bool) { return s, true }

func TestGet_any_simple(t *testing.T) {
	a := Short('a').Switch()
	b := Any("REST", some).Help("any help")
	var parser *OptionParser[fun.Pair[bool, string]] = Pair(a, b).Parser().ToOptions()

	type R = fun.Pair[bool, string]
	assertSuccess(t, R{true, "-b"}, parser, "-a", "-b")
	assertSuccess(t, R{true, "-b"}, parser, "-b", "-a")
	assertSuccess(t, R{true, "-b=foo"}, parser, "-b=foo", "-a")
}

func TestGet_any_many(t *testing.T) {
	var a = Short('a').Switch()
	var b = Many(Any("REST", some).Help("any help").Parser())
	var parser *OptionParser[fun.Pair[bool, []string]] = Pair(a, b).Parser().ToOptions()

	type R = fun.Pair[bool, []string]
	assertSuccess(t, R{true, []string{"-b"}}, parser, "-a", "-b")
	assertSuccess(t, R{true, []string{"-b"}}, parser, "-b", "-a")
	assertSuccess(t, R{true, []string{"-b", "-b"}}, parser, "-b", "-a", "-b")
}

// func Testget_any_many2(t *testing.T) {
// 	var parser OptionParser[[]string] = Any("REST", some).many().ToOptions()

// 	var r = parser.run_inner(newArgs("-vvv")).unwrap()
// 	require.Equal(t,r[0], "-vvv")
// }

// func Testget_any_magic(t *testing.T) {
// 	var parser = Short('b').
// 		Argument[string]("anana").
// 		Adjacent().
// 		Guard(func(b) b == "anana", "not anana").
// 		Optional().
// 		Catch().
// 		Map((func(b) b.is_some)()).
// 		ToOptions()

// 	// -b anana - isn't a -banana
// 	var r = parser.
// 		run_inner(newArgs("-b", "anana")).
// 		unwrap_err().
// 		unwrap_stderr()
// 	require.Equal(t,r, "`-b` is not expected in this context")

// 	// close enough :)
// 	assert(parser.run_inner(newArgs("-b=anana")).unwrap())

// 	assert(parser.run_inner(newArgs("-banana")).unwrap())
// 	assert(parser.run_inner(newArgs()).unwrap())
// }

// func Testfrom_str_works_with_parse(t *testing.T) {
// 	var parser = positional[string]("A").
// 		parse(parseUint).
// 		ToOptions()

// 	var r = parser.run_inner(newArgs("42")).unwrap()
// 	require.Equal(t,r, 42)
// }

// func Testsquashed_short_names(t *testing.T) {
// 	var a = Short('a').Switch()
// 	var b = Short('b').Argument[string]("B")
// 	var parser = Pair(a, b).ToOptions()

// 	// For now instead of supporting I parse "-Obits=2048" instead as "-O" with "bits=2048" body
// 	// https://github.com/pacak/bpaf/issues/121
// 	//
// 	// var r = parser.run_inner(newArgs("-ab=foo")).unwrap();
// 	// require.Equal(t,r.V, "foo");

// 	// an odd one, short flag a is squashed with short argument b
// 	var r = parser.run_inner(newArgs("-ab", "foo")).unwrap()
// 	require.Equal(t,r.V, "foo")

// 	// same, but body for b is adjacent as well
// 	var r = parser.run_inner(newArgs("-abfoo")).unwrap()
// 	require.Equal(t,r.V, "foo")

// 	// normal
// 	var r = parser.run_inner(newArgs("-a", "-bfoo")).unwrap()
// 	require.Equal(t,r.V, "foo")
// }

// // func Testcommand_alias(t *testing.T) {
// //     #[derive(Debug, Bpaf, Clone)]
// //     #[bpaf(options)]
// //     enum Groups {
// //         #[bpaf(command("top"), long("top-alias"))]
// //         Command,
// //     }

// //     groups().run_inner(newArgs("top")).unwrap();
// //     groups().run_inner(newArgs("top-alias")).unwrap();
// // }
