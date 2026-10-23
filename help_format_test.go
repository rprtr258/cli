package cli

import (
	"fmt"
	"testing"

	"github.com/rprtr258/fun"
)

func TestHelp_after_switch(t *testing.T) {
	parser := Short('a').Help("this is help").Switch().ToOptions()
	const expected = "Usage: [-a]\n\nAvailable options:\n    -a          this is help\n    -h, --help  Prints help information\n"
	assertHelp(t, parser, expected)
}

func TestFallback_to_usage(t *testing.T) {
	a := Argument(Short('a'),
		"A", ParseUint).
		Parser().ToOptions().
		FallbackToUsage()

	const expected = "Usage: -a=A\n\nAvailable options:\n    -a=A\n    -h, --help  Prints help information\n"
	assertHelp(t, a, expected)
}

func TestFallback_to_usage_nested(t *testing.T) {
	a := Argument(Short('a'),
		"A", ParseUint).
		Parser().ToOptions().
		FallbackToUsage().
		Command("cmd").
		Parser().ToOptions()

	const expected = "Usage: cmd -a=A\n\nAvailable options:\n    -a=A\n    -h, --help  Prints help information\n"
	assertHelp(t, a, expected, "cmd")
}

func TestFancy_meta(t *testing.T) {
	a := Argument(Long("trailing-comma"), "all|es5|none", ParseString)
	b := Argument(Long("stdin-file-path"), "PATH", ParseString)
	parser := Pair(a, b).ToOptions()

	const expected = `Usage: --trailing-comma=<all|es5|none> --stdin-file-path=PATH

Available options:
        --trailing-comma=<all|es5|none>
        --stdin-file-path=PATH
    -h, --help  Prints help information
`
	assertHelp(t, parser, expected)
}

func TestDecorations(t *testing.T) {
	p := Argument(Short('p').
		Long("parser").
		Env("BPAF_VARIABLE"), "ARG", ParseString).
		Parser().ToOptions().
		Description("descr\n descr").
		Header("header\n header").
		Footer("footer\n footer").
		Version("version").
		Usage(NewDoc("custom usage"))

	const expected = `descr
descr

custom usage

header
header

Available options:
    -p, --parser=ARG  [env:BPAF_VARIABLE: N/A]
    -h, --help        Prints help information
    -V, --version     Prints version information

footer
footer
`

	assertHelp(t, p, expected)
}

func TestVery_long_switch(t *testing.T) {
	a := Argument(Short('p').
		Long("ppppppppppppppppppppppppppppppppppppp").
		Help("this is help for megapotato"),
		"MEGAPOTATO", ParseUint)
	b := Short('b').Long("batman").Help("help for batman").Switch()
	p := Pair(a, b).ToOptions()

	const expected = `Usage: -p=MEGAPOTATO [-b]

Available options:
    -p, --ppppppppppppppppppppppppppppppppppppp=MEGAPOTATO  this is help for megapotato
    -b, --batman  help for batman
    -h, --help    Prints help information
`
	assertHelp(t, p, expected)
}

func TestDuplicate_items_same_help(t *testing.T) {
	a := ReqFlag(Short('a'), struct{}{})
	b := ReqFlag(Short('b'), struct{}{})
	c1 := Short('c').Help("c").Switch()
	c2 := Short('c').Help("c").Switch()
	ac := Pair(a, c1)
	bc := Pair(b, c2)
	parser := Choice(ac, bc).ToOptions()

	const expected = `Usage: (-a [-c] | -b [-c])

Available options:
    -a
    -c          c
    -b
    -h, --help  Prints help information
`

	assertHelp(t, parser, expected)
}

func TestDuplicate_items_dif_help(t *testing.T) {
	a := ReqFlag(Short('a'), struct{}{})
	b := ReqFlag(Short('b'), struct{}{})
	c1 := Short('c').Help("c1").Switch()
	c2 := Short('c').Help("c2").Switch()
	ac := Pair(a, c1)
	bc := Pair(b, c2)
	parser := Choice(ac, bc).ToOptions()

	const expected = `Usage: (-a [-c] | -b [-c])

Available options:
    -a
    -c          c1
    -b
    -c          c2
    -h, --help  Prints help information
`

	assertHelp(t, parser, expected)
}

func TestDuplicate_pos_items_same_help(t *testing.T) {
	a := ReqFlag(Short('a'), struct{}{})
	b := ReqFlag(Short('b'), struct{}{})
	c1 := Positional("C", ParseString).Help("C")
	c2 := Positional("C", ParseString).Help("C")
	ac := Pair(a, c1)
	bc := Pair(b, c2)
	parser := Choice(ac, bc).ToOptions()

	const expected = `Usage: (-a C | -b C)

Available positional items:
    C           C

Available options:
    -a
    -b
    -h, --help  Prints help information
`

	assertHelp(t, parser, expected)
}

func TestDuplicate_pos_items_diff_help(t *testing.T) {
	a := ReqFlag(Short('a'), struct{}{})
	b := ReqFlag(Short('b'), struct{}{})
	c1 := Positional("C", ParseString).Help("C1")
	c2 := Positional("C", ParseString).Help("C2")
	ac := Pair(a, c1)
	bc := Pair(b, c2)
	parser := Choice(ac, bc).ToOptions()

	const expected = `Usage: (-a C | -b C)

Available positional items:
    C           C1
    C           C2

Available options:
    -a
    -b
    -h, --help  Prints help information
`

	assertHelp(t, parser, expected)
}

func TestEnum_with_docs(t *testing.T) {
	// builder equivalent of bpaf's derive expansion (top_comment_is_group_help_enum):
	// /// Pick mode:
	// enum Mode { /// help
	//     Intel, /// help
	//     Att }
	// construct!([long("intel").help("help").req_flag(...),
	//              long("att").help("help").req_flag(...)]).group_help("Pick mode:")
	intel := ReqFlag(Long("intel").Help("help"), false)
	att := ReqFlag(Long("att").Help("help"), true)
	parser := Choice(intel, att).GroupHelp("Pick mode:").ToOptions()

	const expected = `Usage: (--intel | --att)

Pick mode:
        --intel  help
        --att    help

Available options:
    -h, --help   Prints help information
`
	assertHelp(t, parser, expected)
}

func TestAnywhere_invariant_check(t *testing.T) {
	// builder equivalent of bpaf's derive expansion:
	// #[bpaf(adjacent)]
	// struct Fooo { tag: (), name: String, val: String }
	// construct!(long("tag").req_flag(()), positional("NAME").help("help for name"),
	//            positional("VAL").help("help for val")).adjacent()
	tag := ReqFlag(Long("tag"), struct{}{})
	name := Positional("NAME", ParseString).Help("help for name")
	val := Positional("VAL", ParseString).Help("help for val")
	fooo := Pair(Pair(tag, name), val).Adjacent()
	a := Short('a').Help("help for a").Switch()
	b := Short('b').Help("help for b").Switch()
	parser := Pair(Pair(a, fooo), b).ToOptions()

	const expected = `Usage: [-a] --tag NAME VAL [-b]

Available options:
    -a          help for a
  --tag NAME VAL
    NAME        help for name
    VAL         help for val

    -b          help for b
    -h, --help  Prints help information
`
	assertHelp(t, parser, expected)

	// this shouldn't crash
	parser.CheckInvariants(true)
}

func TestMulti_arg_help(t *testing.T) {
	a := ReqFlag(Short('f').Long("flag").Help("flag help"), struct{}{})
	b := Short('e').Long("extra").Help("extra strange").Switch()
	c := Positional("NAME", ParseString).Help("pos1 help")
	d := Positional("STATE", ParseBool).Help("pos2 help")
	combo := Pair(Pair(a, b), Pair(c, d)).Adjacent().Optional()
	verbose := Short('v').Long("verbose").Help("verbose").Switch()
	detailed := Long("detailed").Short('d').Help("detailed").Switch()
	parser := Map3(verbose, combo, detailed, func(_ bool, _ Opt[fun.Pair[fun.Pair[struct{}, bool], fun.Pair[string, bool]]], _ bool) struct{} {
		return struct{}{}
	}).ToOptions()

	const expected = `Usage: [-v] [-f [-e] NAME STATE] [-d]

Available options:
    -v, --verbose   verbose
  -f [-e] NAME STATE
    -f, --flag      flag help
    -e, --extra     extra strange
    NAME            pos1 help
    STATE           pos2 help

    -d, --detailed  detailed
    -h, --help      Prints help information
`

	assertHelp(t, parser, expected)
}

func TestMulti_pos_help(t *testing.T) {
	a := Positional("NAME", ParseString).Help("name help")
	b := Positional("VAL", ParseString).Help("val help")
	combo := Pair(a, b).Adjacent()
	verbose := Short('v').Long("verbose").Switch()
	parser := Pair(verbose, combo).ToOptions()

	const expected = `Usage: [-v] NAME VAL

Available positional items:
  NAME VAL
    NAME           name help
    VAL            val help

Available options:
    -v, --verbose
    -h, --help     Prints help information
`
	assertHelp(t, parser, expected)
}

func TestFallback_display_simple_arg(t *testing.T) {
	parser := Argument(Long("a").
		Help("help for a"),
		"NUM", ParseUint).
		Parser().FallbackDisplay(42).
		ToOptions()

	const expected = `Usage: [--a=NUM]

Available options:
        --a=NUM  help for a
                 [default: 42]
    -h, --help   Prints help information
`

	assertHelp(t, parser, expected)
}

func TestFallback_display_simple_pos(t *testing.T) {
	parser := Positional("NUM", ParseUint).
		Help("help for pos").
		Parser().FallbackDisplay(42).
		ToOptions()

	const expected = `Usage: [NUM]

Available positional items:
    NUM         help for pos
                [default: 42]

Available options:
    -h, --help  Prints help information
`

	assertHelp(t, parser, expected)
}

type Pair2 fun.Pair[uint32, uint32]

func (self Pair2) String() string {
	return fmt.Sprintf("Pair %d, %d", self.K, self.V)
}

func TestFallback_display_tuple(t *testing.T) {
	a := Argument(Long("a").Help("help for a"), "NUM", ParseUint)
	b := Argument(Long("b").Help("help for b"), "NUM", ParseUint)
	parser := Pair(a, b).
		Map(func(ab fun.Pair[uint, uint]) Pair2 { return Pair2{uint32(ab.K), uint32(ab.V)} }).
		FallbackDisplay(Pair2{42, 333}).ToOptions()

	const expected = `Usage: [--a=NUM --b=NUM]

Available options:
        --a=NUM  help for a
        --b=NUM  help for b
                 [default: Pair 42, 333]
    -h, --help   Prints help information
`

	assertHelp(t, parser, expected)
}

func TestFallback_display_no_help(t *testing.T) {
	parser := Argument(Long("a"),
		"NUM", ParseUint).
		Parser().FallbackDisplay(42).
		ToOptions()

	const expected = `Usage: [--a=NUM]

Available options:
        --a=NUM
                 [default: 42]
    -h, --help   Prints help information
`

	assertHelp(t, parser, expected)
}

func TestEnv_fallback_visible(t *testing.T) {
	fonts_dir := Argument(Long("fonts").
		Env("OIKOS_FONTS").
		Help("Load fonts from this directory"),
		"DIR", ParseString).
		Parser().
		Optional()

	system_fonts := Long("system-fonts").
		Env("OIKOS_SYSTEM_FONTS").
		Help("Search for additional fonts in system directories").
		Switch()
	parser := Pair(fonts_dir, system_fonts).ToOptions()

	const expected = `Usage: [--fonts=DIR] [--system-fonts]

Available options:
        --fonts=DIR     Load fonts from this directory
                        [env:OIKOS_FONTS: N/A]
        --system-fonts  Search for additional fonts in system directories
                        [env:OIKOS_SYSTEM_FONTS: not set]
    -h, --help          Prints help information
`
	assertHelp(t, parser, expected)
}

func TestNested_group_help(t *testing.T) {
	a := Short('a').Switch().GroupHelp("inner")
	b := Short('b').Switch()

	parser := Pair(a, b).GroupHelp("outer").ToOptions()

	const expected = `Usage: [-a] [-b]

outer
    -a
    -b

Available options:
    -h, --help  Prints help information
`
	assertHelp(t, parser, expected)
}

func TestWith_group_help(t *testing.T) {
	a := Short('a').Help("option a").Switch()
	b := Short('b').Help("option b").Switch()
	c := Short('c').Help("option c").Switch()

	ab := Pair(a, b).Parser().withGroupHelp(func(meta meta) Doc {
		var b Doc
		b.Emphasis("Uses either of those")
		b.meta(meta, false)
		return b
	})
	parser := Pair(ab, c).ToOptions()

	{
		const expected = `Usage: [-a] [-b] [-c]

Uses either of those
    -a          option a
    -b          option b

Available options:
    -c          option c
    -h, --help  Prints help information
`

		assertHelp(t, parser, expected)
	}

	assertSuccess(t, PPair(PPair(true, false), true), parser, "-a", "-c")
}

func TestCustom_help_and_version(t *testing.T) {
	h := Short('H').Long("halp").Help("halps you")
	v := Short('v').Long("release").Help("prints release id")
	a := Short('a').Switch()
	parser := a.ToOptions().HelpParser(h).VersionParser(v)

	assertFailure(t, parser, "`--help` is not expected in this context", "--help")

	{
		const expected = "Usage: [-a]\n\nAvailable options:\n    -a\n    -H, --halp  halps you\n"
		assertHelp(t, parser, expected, "--halp")
	}
}

func TestVarious_name_lengths_under(t *testing.T) {
	parser := Short('a').
		Long("123456789012345").
		Help("A").
		Switch().
		Parser().ToOptions()

	const expected = `Usage: [-a]

Available options:
    -a, --123456789012345  A
    -h, --help             Prints help information
`
	assertHelp(t, parser, expected)
}

func TestVarious_name_lengths_at(t *testing.T) {
	parser := Short('a').
		Long("1234567890123456").
		Help("A").
		Switch().
		Parser().ToOptions()

	const expected = `Usage: [-a]

Available options:
    -a, --1234567890123456  A
    -h, --help              Prints help information
`
	assertHelp(t, parser, expected)
}

func TestVarious_name_lengths_over1(t *testing.T) {
	{
		parser := Short('a').
			Long("12345678901234567").
			Help("A").
			Switch().
			Parser().ToOptions()

		const expected = `Usage: [-a]

Available options:
    -a, --12345678901234567  A
    -h, --help               Prints help information
`
		assertHelp(t, parser, expected)
	}

	{
		parser := Short('a').
			Long("1234567890123456789").
			Help("A").
			Switch().
			Parser().ToOptions()

		const expected = `Usage: [-a]

Available options:
    -a, --1234567890123456789  A
    -h, --help  Prints help information
`
		assertHelp(t, parser, expected)
	}

	{
		parser := Short('a').Switch().ToOptions().Version("1")

		assertHelp(t, parser, "Usage: [-a]\n\nAvailable options:\n    -a\n    -h, --help     Prints help information\n    -V, --version  Prints version information\n")

		assertHelp(t, parser, "Version: 1\n", "--version")
	}
}
