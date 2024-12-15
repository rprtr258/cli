package cli

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func write_updated(t *testing.T, new_val string, path string) {
	t.Helper()

	current_val, err := os.ReadFile(path)
	require.NoError(t, err)
	if string(current_val) != new_val {
		// os.WriteFile(path,new_val)
		t.FailNow()
	}
}

func TestSimple(t *testing.T) {
	kraken := Short('d').
		Long("kraken").
		Help("Unleash the kraken").
		Switch()

	user := Argument(Long("user").
		Env("USER").
		Help("Log in as this user"),
		"USER", parseString)

	options := Pair(kraken, user).
		Parser().ToOptions().
		Descr("I am a program and I do things").
		Header("Sometimes they even work.").
		Footer("Beware `-d`, dragons be here")
	roff := options.render_manpage(
		"simple",
		Section{kind: SectionGeneral},
		Valid("Aug 2022"),
		Valid(os.Getenv("CARGO_PKG_AUTHORS")),
		Valid("asdf"),
	)

	write_updated(t, roff, "tests/simple.1")
}

func TestNested(t *testing.T) {
	a := Argument(Short('d').
		Help("dragon"),
		"y", parseString).
		Parser().ToOptions().
		Descr("I am a program and I do things").
		Header("Sometimes they even work. 1").
		Footer("Beware `-d`, dragons be here 1").
		Command("cmd")

	b := Argument(Short('k').
		Help("kraken"),
		"x", parseString).
		Parser().ToOptions().
		Descr("I am a program and I do things 2").
		Header("Sometimes they even work. 2").
		Footer("Beware `-d`, dragons be here 2").
		Command("dmc").
		short('d')

	c := Positional("C", parseString).Help("Mystery file")

	d := Argument(Short('d').
		Long("ddd").
		Help("mystery arg"),
		"D", parseString)

	// e := Short('e').Long("eee").help("the e in the room").argument::<String>("E");
	a_or_b := Choice(a, b)

	roff := Map3(d, c, a_or_b, map3).
		ToOptions().
		Descr("I am a program and I do things 3").
		Header("Sometimes they even work. 3").
		Footer("Beware `-d`, dragons be here 3").
		render_manpage(
			"simple",
			Section{kind: SectionGeneral},
			Valid("Aug 2022"),
			Valid(os.Getenv("CARGO_PKG_AUTHORS")),
			Valid("asdf"),
		)

	write_updated(t, roff, "tests/nested.1")
}

func TestVery_nested(t *testing.T) {
	options := Argument(Short('k').
		Help("Unleash the Kraken"),
		"NAME", parseString).
		Parser().ToOptions().
		Descr("lvl 4 description").
		Command("lvl4").
		Parser().ToOptions().
		Descr("lvl 3 description").
		Command("lvl3").
		Parser().ToOptions().
		Descr("lvl 2 description").
		Command("lvl2").
		Parser().ToOptions().
		Descr("lvl 1 description").
		Command("lvl1").
		Parser().ToOptions()
	roff := options.render_manpage(
		"simple",
		Section{kind: SectionGeneral},
		Valid("Aug 2022"),
		Valid(os.Getenv("CARGO_PKG_AUTHORS")),
		Valid("asdf"),
	)

	write_updated(t, roff, "tests/very_nested.1")
}
