package cli

import (
	"fmt"
	"testing"

	"github.com/rprtr258/fun"
	"github.com/stretchr/testify/require"
)

func TestHelp_after_switch(t *testing.T) {
	parser := Short('a').Switch().help(newDocFromString("this is help")).Parser().ToOptions()
	_, r := parser.run_inner(newArgs("--help"))
	const expected = "Usage: [-a]\n\nAvailable options:\n    -a          this is help\n    -h, --help  Prints help information\n"
	require.Equal(t, expected, unwrap_stdout(r))
}

func TestFallback_to_usage(t *testing.T) {
	a := Argument(Short('a'),
		"A", parseUint).
		Parser().ToOptions().
		FallbackToUsage()

	_, r := a.run_inner(newArgs())
	const expected = "Usage: -a=A\n\nAvailable options:\n    -a=A\n    -h, --help  Prints help information\n"
	require.Equal(t, expected, unwrap_stdout(r))
}

func TestFallback_to_usage_nested(t *testing.T) {
	a := Argument(Short('a'),
		"A", parseUint).
		Parser().ToOptions().
		FallbackToUsage().
		Command("cmd").
		Parser().ToOptions()

	_, r := a.run_inner(newArgs("cmd"))
	const expected = "Usage: cmd -a=A\n\nAvailable options:\n    -a=A\n    -h, --help  Prints help information\n"
	require.Equal(t, expected, unwrap_stdout(r))
}

func TestFancy_meta(t *testing.T) {
	a := Argument(Long("trailing-comma"), "all|es5|none", parseString)
	b := Argument(Long("stdin-file-path"), "PATH", parseString)
	parser := Pair(a, b).Parser().ToOptions()

	_, r := parser.run_inner(newArgs("--help"))

	const expected = `Usage: --trailing-comma=<all|es5|none> --stdin-file-path=PATH

Available options:
    --trailing-comma=<all|es5|none>
    --stdin-file-path=PATH
    -h, --help              Prints help information
`
	require.Equal(t, expected, unwrap_stdout(r))
}

func TestDecorations(t *testing.T) {
	p := Argument(Short('p').
		Long("parser").
		Env("BPAF_VARIABLE"), "ARG", parseString).
		Parser().ToOptions().
		Descr("descr\n descr").
		Header("header\n header").
		Footer("footer\n footer").
		version("version").
		Usage(newDocFromString("custom usage"))

	_, r := p.run_inner(newArgs("--help"))

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

	require.Equal(t, expected, unwrap_stdout(r))
}

func TestVery_long_switch(t *testing.T) {
	a := Argument(Short('p').
		Long("ppppppppppppppppppppppppppppppppppppp").
		Help("this is help for megapotato"),
		"MEGAPOTATO", parseUint)
	b := Short('b').Long("batman").Help("help for batman").Switch()
	p := Pair(a, b).Parser().ToOptions()

	_, r := p.run_inner(newArgs("--help"))
	const expected = `Usage: -p=MEGAPOTATO [-b]

Available options:
    -p, --ppppppppppppppppppppppppppppppppppppp=MEGAPOTATO  this is help for megapotato
    -b, --batman  help for batman
    -h, --help    Prints help information
`
	require.Equal(t, expected, unwrap_stdout(r))
}

func TestDuplicate_items_same_help(t *testing.T) {
	a := ReqFlag(Short('a'), struct{}{})
	b := ReqFlag(Short('b'), struct{}{})
	c1 := Short('c').Help("c").Switch()
	c2 := Short('c').Help("c").Switch()
	ac := Pair(a, c1)
	bc := Pair(b, c2)
	parser := Choice(ac, bc).Parser().ToOptions()

	_, r := parser.run_inner(newArgs("--help"))

	const expected = `Usage: (-a [-c] | -b [-c])

Available options:
    -a
    -c          c
    -b
    -h, --help  Prints help information
`

	require.Equal(t, expected, unwrap_stdout(r))
}

func TestDuplicate_items_dif_help(t *testing.T) {
	a := ReqFlag(Short('a'), struct{}{})
	b := ReqFlag(Short('b'), struct{}{})
	c1 := Short('c').Help("c1").Switch()
	c2 := Short('c').Help("c2").Switch()
	ac := Pair(a, c1)
	bc := Pair(b, c2)
	parser := Choice(ac, bc).Parser().ToOptions()

	_, r := parser.run_inner(newArgs("--help"))

	const expected = `Usage: (-a [-c] | -b [-c])

Available options:
    -a
    -c          c1
    -b
    -c          c2
    -h, --help  Prints help information
`

	require.Equal(t, expected, unwrap_stdout(r))
}

func TestDuplicate_pos_items_same_help(t *testing.T) {
	a := ReqFlag(Short('a'), struct{}{})
	b := ReqFlag(Short('b'), struct{}{})
	c1 := Positional("C", parseString).Help("C")
	c2 := Positional("C", parseString).Help("C")
	ac := Pair(a, c1)
	bc := Pair(b, c2)
	parser := Choice(ac, bc).Parser().ToOptions()

	_, r := parser.run_inner(newArgs("--help"))

	const expected = `Usage: (-a C | -b C)

Available positional items:
    C           C

Available options:
    -a
    -b
    -h, --help  Prints help information
`

	require.Equal(t, expected, unwrap_stdout(r))
}

func TestDuplicate_pos_items_diff_help(t *testing.T) {
	a := ReqFlag(Short('a'), struct{}{})
	b := ReqFlag(Short('b'), struct{}{})
	c1 := Positional("C", parseString).Help("C1")
	c2 := Positional("C", parseString).Help("C2")
	ac := Pair(a, c1)
	bc := Pair(b, c2)
	parser := Choice(ac, bc).Parser().ToOptions()

	_, r := parser.run_inner(newArgs("--help"))

	const expected = `Usage: (-a C | -b C)

Available positional items:
    C           C1
    C           C2

Available options:
    -a
    -b
    -h, --help  Prints help information
`

	require.Equal(t, expected, unwrap_stdout(r))
}

// TODO: fix
// func TestEnum_with_docs(t *testing.T) {
//  #derive(Bpaf)
// 	type Mode bool
// 	// Pick mode:
// 	const (
// 		// help
// 		//
// 		// absent
// 		Intel Mode = false
// 		// help
// 		//
// 		// Hidden
// 		Att Mode = true
// 	)

// 	r := mode().
// 		Parser().ToOptions().
// 		run_inner(newArgs("--help")).
// 		unwrap_err().
// 		unwrap_stdout()

// 	const expected=`// Usage: (--intel | --att)

// Pick mode:
//         --intel  help
//         --att    help

// Available options:
//     -h, --help   Prints help information
// `
// 	require.Equal(t, expected, r)
// }

// TODO: fix
// func TestAnywhere_invariant_check(t*testing.T) {
//     #[bpaf(adjacent)]
//     struct Fooo {
//         tag: (),
//         #[bpaf(positional("NAME"))]
//         /// help for name
//         name: String,
//         #[bpaf(positional("VAL"))]
//         /// help for val
//         val: String,
//     }

//     a := Short('a').help("help for a").Switch();
//     b := Short('b').help("help for b").Switch();
//     parser := Pair(a, fooo(), b).Parser().ToOptions();

//     const expected=`// Usage: [-a] --tag NAME VAL [-b]

// Available options:
//     -a          help for a
//   --tag NAME VAL
//     NAME        help for name
//     VAL         help for val

//     -b          help for b
//     -h, --help  Prints help information
// `

//     r := parser.run_inner(newArgs("--help")).unwrap_err().unwrap_stdout();
//     require.Equal(t, expected, r);

//     // this shouldn't crash
//     parser.check_invariants(true);
// }

func TestMulti_arg_help(t *testing.T) {
	a := ReqFlag(Short('f').Long("flag").Help("flag help"), struct{}{}).Parser()
	b := Short('e').Long("extra").Help("extra strange").Switch()
	c := Positional("NAME", parseString).Help("pos1 help").Parser()
	d := Positional("STATE", parseBool).Help("pos2 help").Parser()
	combo := Optional(Pair(Pair(a, b), Pair(c, d)).Adjacent().Parser()).Parser()
	verbose := Short('v').Long("verbose").Help("verbose").Switch().Parser()
	detailed := Long("detailed").Short('d').Help("detailed").Switch().Parser()
	parser := Map3(verbose, combo, detailed, func(_ bool, _ Option[fun.Pair[fun.Pair[struct{}, bool], fun.Pair[string, bool]]], _ bool) struct{} {
		return struct{}{}
	}).Parser().ToOptions()

	_, r := parser.run_inner(newArgs("--help"))

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

	require.Equal(t, expected, unwrap_stdout(r))
}

func TestMulti_pos_help(t *testing.T) {
	a := Positional("NAME", parseString).Help("name help")
	b := Positional("VAL", parseString).Help("val help")
	combo := Pair(a, b).Adjacent()
	verbose := Short('v').Long("verbose").Switch()
	parser := Pair(verbose, combo).Parser().ToOptions()
	_, r := parser.run_inner(newArgs("--help"))

	const expected = `Usage: [-v] NAME VAL

Available positional items:
  NAME VAL
    NAME           name help
    VAL            val help

Available options:
    -v, --verbose
    -h, --help     Prints help information
`
	require.Equal(t, expected, unwrap_stdout(r))
}

func TestFallback_display_simple_arg(t *testing.T) {
	parser := Argument(Long("a").
		Help("help for a"),
		"NUM", parseUint).
		Parser().Fallback(42).
		DisplayFallback().
		Parser().ToOptions()

	_, r := parser.run_inner(newArgs("--help"))
	const expected = `Usage: [--a=NUM]

Available options:
    --a=NUM     help for a
                [default: 42]
    -h, --help  Prints help information
`

	require.Equal(t, expected, unwrap_stdout(r))
}

func TestFallback_display_simple_pos(t *testing.T) {
	parser := Positional("NUM", parseUint).
		Help("help for pos").
		Parser().Fallback(42).
		DisplayFallback().
		Parser().ToOptions()

	_, r := parser.run_inner(newArgs("--help"))

	const expected = `Usage: [NUM]

Available positional items:
    NUM         help for pos
                [default: 42]

Available options:
    -h, --help  Prints help information
`

	require.Equal(t, expected, unwrap_stdout(r))
}

type Pair2 fun.Pair[uint32, uint32]

func (self Pair2) String() string {
	return fmt.Sprintf("Pair %d, %d", self.K, self.V)
}

func TestFallback_display_tuple(t *testing.T) {
	a := Argument(Long("a").Help("help for a"), "NUM", parseUint)
	b := Argument(Long("b").Help("help for b"), "NUM", parseUint)
	parser := Map(Pair(a, b).Parser(), func(ab fun.Pair[uint, uint]) Pair2 { return Pair2{uint32(ab.K), uint32(ab.V)} }).
		Parser().Fallback(Pair2{42, 333}).
		DisplayFallback().
		Parser().ToOptions()

	_, r := parser.run_inner(newArgs("--help"))

	const expected = `Usage: [--a=NUM --b=NUM]

Available options:
    --a=NUM     help for a
    --b=NUM     help for b
                [default: Pair 42, 333]
    -h, --help  Prints help information
`

	require.Equal(t, expected, unwrap_stdout(r))
}

func TestFallback_display_no_help(t *testing.T) {
	parser := Argument(Long("a"),
		"NUM", parseUint).
		Parser().Fallback(42).
		DisplayFallback().
		Parser().ToOptions()

	_, r := parser.run_inner(newArgs("--help"))
	const expected = `Usage: [--a=NUM]

Available options:
    --a=NUM
                [default: 42]
    -h, --help  Prints help information
`

	require.Equal(t, expected, unwrap_stdout(r))
}

func TestEnv_fallback_visible(t *testing.T) {
	fonts_dir := Optional(Argument(Long("fonts").
		Env("OIKOS_FONTS").
		Help("Load fonts from this directory"),
		"DIR", parseString).Parser())

	system_fonts := Long("system-fonts").
		Env("OIKOS_SYSTEM_FONTS").
		Help("Search for additional fonts in system directories").
		Switch()
	parser := Pair(fonts_dir, system_fonts).Parser().ToOptions()

	_, r := parser.run_inner(newArgs("--help"))

	const expected = `Usage: [--fonts=DIR] [--system-fonts]

Available options:
    --fonts=DIR     Load fonts from this directory
                    [env:OIKOS_FONTS: N/A]
    --system-fonts  Search for additional fonts in system directories
                    [env:OIKOS_SYSTEM_FONTS: not set]
    -h, --help      Prints help information
`
	require.Equal(t, expected, unwrap_stdout(r))
}

func TestNested_group_help(t *testing.T) {
	a := Short('a').Switch().Parser().GroupHelp("inner")
	b := Short('b').Switch()

	parser := Pair(a, b).Parser().GroupHelp("outer").Parser().ToOptions()

	_, r := parser.run_inner(newArgs("--help"))
	const expected = `Usage: [-a] [-b]

outer
    -a
    -b

Available options:
    -h, --help  Prints help information
`
	require.Equal(t, expected, unwrap_stdout(r))
}

func TestWith_group_help(t *testing.T) {
	t.Skip() // TODO: fix
	a := Short('a').Help("option a").Switch()
	b := Short('b').Help("option b").Switch()
	c := Short('c').Help("option c").Switch()

	ab := Pair(a, b).Parser().with_group_help(func(meta Meta) Doc {
		var b Doc
		b.Emphasis("Uses either of those ")
		b.Meta(meta, false)
		return b
	}).Parser()
	parser := Pair(ab, c).Parser().ToOptions()

	{
		_, r := parser.run_inner(newArgs("--help"))
		const expected = `Usage: [-a] [-b] [-c]

Uses either of those [-a] [-b]
    -a          option a
    -b          option b

Available options:
    -c          option c
    -h, --help  Prints help information
`

		require.Equal(t, expected, unwrap_stdout(r))
	}

	{
		r, err := parser.run_inner(newArgs("-a", "-c"))
		require.Nil(t, err)
		require.Equal(t, r, pair(pair(true, false), true))
	}
}

func TestCustom_help_and_version(t *testing.T) {
	h := Short('H').Long("halp").Help("halps you")
	v := Short('v').Long("release").Help("prints release id")
	a := Short('a').Switch()
	parser := a.Parser().ToOptions().help_parser(h).version_parser(v)

	{
		_, r := parser.run_inner(newArgs("--help"))
		require.Equal(t, unwrap_stderr(r), "`--help` is not expected in this context")
	}

	{
		_, r := parser.run_inner(newArgs("--halp"))
		const expected = "Usage: [-a]\n\nAvailable options:\n    -a\n    -H, --halp  halps you\n"
		require.Equal(t, expected, unwrap_stdout(r))
	}
}

func TestVarious_name_lengths_under(t *testing.T) {
	parser := Short('a').
		Long("123456789012345").
		Help("A").
		Switch().
		Parser().ToOptions()
	_, r := parser.run_inner(newArgs("--help"))

	const expected = `Usage: [-a]

Available options:
    -a, --123456789012345  A
    -h, --help             Prints help information
`
	require.Equal(t, expected, unwrap_stdout(r))
}

func TestVarious_name_lengths_at(t *testing.T) {
	parser := Short('a').
		Long("1234567890123456").
		Help("A").
		Switch().
		Parser().ToOptions()
	_, r := parser.run_inner(newArgs("--help"))

	const expected = `Usage: [-a]

Available options:
    -a, --1234567890123456  A
    -h, --help              Prints help information
`
	require.Equal(t, expected, unwrap_stdout(r))
}

func TestVarious_name_lengths_over1(t *testing.T) {
	{
		parser := Short('a').
			Long("12345678901234567").
			Help("A").
			Switch().
			Parser().ToOptions()
		_, r := parser.run_inner(newArgs("--help"))

		const expected = `Usage: [-a]

Available options:
    -a, --12345678901234567  A
    -h, --help               Prints help information
`
		require.Equal(t, expected, unwrap_stdout(r))
	}

	{
		parser := Short('a').
			Long("1234567890123456789").
			Help("A").
			Switch().
			Parser().ToOptions()
		_, r := parser.run_inner(newArgs("--help"))

		const expected = `Usage: [-a]

Available options:
    -a, --1234567890123456789  A
    -h, --help  Prints help information
`
		require.Equal(t, expected, unwrap_stdout(r))
	}

	{
		parser := Short('a').Switch().Parser().ToOptions().version("1")

		{
			_, r := parser.run_inner(newArgs("--help"))
			require.Equal(t, unwrap_stdout(r), "Usage: [-a]\n\nAvailable options:\n    -a\n    -h, --help     Prints help information\n    -V, --version  Prints version information\n")
		}

		{
			_, r := parser.run_inner(newArgs("--version"))
			require.Equal(t, unwrap_stdout(r), "Version: 1\n")
		}
	}
}
