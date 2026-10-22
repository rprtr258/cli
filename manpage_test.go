package cli

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func write_updated(t *testing.T, new_val, path string) {
	t.Helper()

	current_val, err := os.ReadFile(path)
	require.NoError(t, err)
	if string(current_val) != new_val {
		t.FailNow()
	}
}

func TestSimple(t *testing.T) {
	// unskipped
	kraken := Short('d').
		Long("kraken").
		Help("Unleash the kraken").
		Switch()

	user := Argument(Long("user").
		Env("USER").
		Help("Log in as this user"),
		"USER", ParseString)

	options := Pair(kraken, user).
		Parser().ToOptions().
		Description("I am a program and I do things").
		Header("Sometimes they even work.").
		Footer("Beware `-d`, dragons be here")
	roff := options.RenderManpage(
		"simple",
		Section{Kind: SectionGeneral},
		Valid("Aug 2022"),
		Valid("Michael Baykov <manpacket@gmail.com>"),
		Valid("asdf"),
	)

	write_updated(t, roff, "tests/simple.1")
}

func TestNested(t *testing.T) {
	// unskipped
	a := Argument(Short('d').
		Help("dragon"),
		"y", ParseString).
		Parser().ToOptions().
		Description("I am a program and I do things").
		Header("Sometimes they even work. 1").
		Footer("Beware `-d`, dragons be here 1").
		Command("cmd")

	b := Argument(Short('k').
		Help("kraken"),
		"x", ParseString).
		Parser().ToOptions().
		Description("I am a program and I do things 2").
		Header("Sometimes they even work. 2").
		Footer("Beware `-d`, dragons be here 2").
		Command("dmc").
		Short('d')

	c := Positional("C", ParseString).Help("Mystery file")

	d := Argument(Short('d').
		Long("ddd").
		Help("mystery arg"),
		"D", ParseString)

	// e := Short('e').Long("eee").help("the e in the room").argument::<String>("E");
	a_or_b := Choice(a, b)

	roff := Map3(d, c, a_or_b, map3).
		ToOptions().
		Description("I am a program and I do things 3").
		Header("Sometimes they even work. 3").
		Footer("Beware `-d`, dragons be here 3").
		RenderManpage(
			"simple",
			Section{Kind: SectionGeneral},
			Valid("Aug 2022"),
			Valid("Michael Baykov <manpacket@gmail.com>"),
			Valid("asdf"),
		)

	write_updated(t, roff, "tests/nested.1")
}

func TestVery_nested(t *testing.T) {
	// unskipped
	options := Argument(Short('k').
		Help("Unleash the Kraken"),
		"NAME", ParseString).
		Parser().ToOptions().
		Description("lvl 4 description").
		Command("lvl4").
		Parser().ToOptions().
		Description("lvl 3 description").
		Command("lvl3").
		Parser().ToOptions().
		Description("lvl 2 description").
		Command("lvl2").
		Parser().ToOptions().
		Description("lvl 1 description").
		Command("lvl1").
		Parser().ToOptions()
	roff := options.RenderManpage(
		"simple",
		Section{Kind: SectionGeneral},
		Valid("Aug 2022"),
		Valid("Michael Baykov <manpacket@gmail.com>"),
		Valid("asdf"),
	)

	write_updated(t, roff, "tests/very_nested.1")
}
