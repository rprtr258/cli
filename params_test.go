package cli

import (
	"testing"

	"github.com/rprtr258/fun"
)

func some(s string) (string, bool) { return s, true }

func TestGet_any_simple(t *testing.T) {
	a := Short('a').Switch()
	b := Any("REST", some, false).Help("any help")
	parser := Pair(a, b).ToOptions()

	type R = fun.Pair[bool, string]
	assertSuccess(t, R{true, "-b"}, parser, "-a -b")
	assertSuccess(t, R{true, "-b"}, parser, "-b -a")
	assertSuccess(t, R{true, "-b=foo"}, parser, "-b=foo -a")
}

func TestGet_any_many(t *testing.T) {
	a := Short('a').Switch()
	b := Any("REST", some, false).Help("any help").Parser().Many()
	parser := Pair(a, b).ToOptions()

	type R = fun.Pair[bool, []string]
	assertSuccess(t, R{true, []string{"-b"}}, parser, "-a -b")
	assertSuccess(t, R{true, []string{"-b"}}, parser, "-b -a")
	assertSuccess(t, R{true, []string{"-b", "-b"}}, parser, "-b -a -b")
}

func TestGet_any_many2(t *testing.T) {
	parser := Any("REST", some, false).Parser().Many().ToOptions()
	assertSuccess(t, []string{"-vvv"}, parser, "-vvv")
}

func TestGet_any_magic(t *testing.T) {
	parser := Short('b').Argument(
		"anana", ParseString).
		Adjacent().
		Parser().
		Guard(func(b string) bool { return b == "anana" }, "not anana").
		Parser().
		OptionalCatch().
		Map(func(b Opt[string]) bool { return b.Valid }).
		ToOptions()

	// -b anana - isn't a -banana
	assertFailure(t, parser, "`-b` is not expected in this context", "-b", "anana")

	// close enough :)
	assertSuccess(t, true, parser, "-b=anana")
	assertSuccess(t, true, parser, "-banana")
	assertSuccess(t, false, parser, "")
}

func TestFrom_str_works_with_parse(t *testing.T) {
	parser := Parse(Positional("A", ParseString),
		ParseUint).ToOptions()
	assertSuccess(t, uint(42), parser, "42")
}

func TestSquashed_short_names(t *testing.T) {
	a := Short('a').Switch()
	b := Short('b').Argument("B", ParseString)
	parser := Pair(a, b).ToOptions()

	type R = fun.Pair[bool, string]

	// an odd one, short flag a is squashed with short argument b
	assertSuccess(t, R{true, "foo"}, parser, "-ab foo")

	// same, but body for b is adjacent as well
	assertSuccess(t, R{true, "foo"}, parser, "-abfoo")

	// normal
	assertSuccess(t, R{true, "foo"}, parser, "-a -bfoo")
}
