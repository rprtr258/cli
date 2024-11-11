package cli

import (
	"strconv"
	"testing"

	"github.com/rprtr258/fun"
	"github.com/stretchr/testify/require"
)

func assertSuccess[T any](t *testing.T, expected T) func(T, ParseFailure) {
	t.Helper()
	return func(actual T, err ParseFailure) {
		t.Helper()
		require.Nil(t, err)
		require.Equal(t, expected, actual)
	}
}

func wtf_shenanigans_1(t *testing.T) {
	for _, tc := range []struct {
		i_c    ArgType
		prefix string
	}{
		{ArgTypeShort, "f"},
		{ArgTypeLong, "foo"},
		{ArgTypeLong, "口水鸡"},
	} {
		i_c, prefix := tc.i_c, tc.prefix
		i_prefix := prefix
		i_suffix := string([]byte{0x66, 0x6f, 0xD8, 0x6f})

		os_string := ""
		switch i_c {
		case ArgTypeShort:
			os_string += "-"
		case ArgTypeLong:
			os_string += "--"
		}
		os_string += i_prefix
		os_string += "="
		os_string += i_suffix

		o_c, o_prefix, o_suffix, ok := split_os_argument(os_string)
		require.True(t, ok)
		require.Equal(t, i_c, o_c)
		require.Equal(t, i_prefix, o_prefix)
		require.Equal(t, ArgArgWord(i_suffix), o_suffix.unwrap())
	}
}

func wtf_shenanigans_2(t *testing.T) {
	for _, tc := range []struct {
		i_c    ArgType
		prefix string
	}{
		{ArgTypeShort, "f"},
		{ArgTypeLong, "foo"},
		{ArgTypeLong, "口水鸡"},
	} {
		i_c, prefix := tc.i_c, tc.prefix
		i_prefix := prefix
		i_suffix := "口水鸡"

		os_string := ""
		switch i_c {
		case ArgTypeShort:
			os_string += "-"
		case ArgTypeLong:
			os_string += "--"
		}
		os_string += i_prefix
		os_string += "="
		os_string += i_suffix

		o_c, o_prefix, o_suffix, ok := split_os_argument(os_string)
		require.True(t, ok)
		require.Equal(t, i_c, o_c)
		require.Equal(t, i_prefix, o_prefix)
		require.Equal(t, ArgArgWord(i_suffix), o_suffix.unwrap())
	}
}

func de_yoda(t *testing.T) {
	parser := Pair(Short('a').Switch(), Short('b').Switch()).Parser().ToOptions()
	assertSuccess(t, fun.Pair[bool, bool]{false, false})(parser.run_inner(newArgs()))
	assertSuccess(t, fun.Pair[bool, bool]{true, true})(parser.run_inner(newArgs("-a", "-b")))
}

func TestConstructWithFn(t *testing.T) {
	type Opts struct {
		a, b, c bool
	}

	a := Short('a').Switch()
	b := Short('b').Switch()
	c := Short('c').Switch()
	parser := Map3(a, b, c, func(a bool, b bool, c bool) Opts { return Opts{a, b, c} }).ToOptions()

	_, err := parser.run_inner(newArgs("--help"))
	help := unwrap_stdout(err)
	expected_help := `Usage: [-a] [-b] [-c]

Available options:
    -a
    -b
    -c
    -h, --help  Prints help information
`
	require.Equal(t, expected_help, help)

	x, err := parser.run_inner(newArgs("-b", "-c"))
	requireNoParseFailure(t, err)
	require.Equal(t, Opts{false, true, true}, x)
}

func TestSimpleTwoOptionalFlags(t *testing.T) {
	a := Short('a').Long("AAAAA").Switch()
	b := Short('b').Switch()
	x := Pair(a, b)
	decorated := x.Parser().ToOptions().Descr("this is a test")

	t.Run("no version information given - no version field generated", func(t *testing.T) {
		_, err := decorated.run_inner(newArgs("-a", "-V"))
		require.Equal(t, "`-V` is not expected in this context", unwrap_stderr(err))
	})

	t.Run("accept only one copy of -a", func(t *testing.T) {
		_, err := decorated.run_inner(newArgs("-a", "-a"))
		require.Equal(t,
			"argument `-a` cannot be used multiple times in this context",
			unwrap_stderr(err),
		)
	})

	t.Run("help", func(t *testing.T) {
		_, err := decorated.run_inner(newArgs("-h"))
		const expected_help = `this is a test

Usage: [-a] [-b]

Available options:
    -a, --AAAAA
    -b
    -h, --help   Prints help information
`
		require.Equal(t, expected_help, unwrap_stdout(err))
	})
}

func TestSimpleTwoOptionalFlagsWithOneHidden(t *testing.T) {
	a := Short('a').Long("AAAAA").Switch()
	b := Short('b').Switch().Hide()
	decorated := Pair(a, b).Parser().ToOptions().Descr("this is a test")

	// no version information given - no version field generated
	{
		_, err := decorated.run_inner(newArgs("-a", "-V"))
		require.Equal(t, "`-V` is not expected in this context", unwrap_stderr(err))
	}

	// accepts only one copy of -a
	{
		_, err := decorated.run_inner(newArgs("-a", "-a"))
		require.Equal(t,
			"argument `-a` cannot be used multiple times in this context",
			unwrap_stderr(err),
		)
	}

	{
		_, err := decorated.run_inner(newArgs("-h"))
		const expected_help = `this is a test

Usage: [-a]

Available options:
    -a, --AAAAA
    -h, --help   Prints help information
`
		require.Equal(t, expected_help, unwrap_stdout(err))
	}
}

func TestEitherOfThreeRequiredFlags(t *testing.T) {
	a := ReqFlag(Short('a'), struct{}{})
	b := ReqFlag(Short('b'), struct{}{})
	c := ReqFlag(Short('c'), struct{}{})
	p := a.Parser().or_else(b.Parser()).Parser().or_else(c.Parser()).Parser()
	decorated := p.ToOptions().version("1.0")

	// version help requires version meta
	{
		_, ver := decorated.run_inner(newArgs("-V"))
		require.Equal(t, "Version: 1.0\n", unwrap_stdout(ver))
	}

	// help is always generated
	{
		_, help := decorated.run_inner(newArgs("-h"))
		const expected_help = `Usage: (-a | -b | -c)

Available options:
    -a
    -b
    -c
    -h, --help     Prints help information
    -V, --version  Prints version information
`
		require.Equal(t, expected_help, unwrap_stdout(help))
	}

	// must specify one of the required flags
	{
		_, err := decorated.run_inner(newArgs())
		require.Equal(t,
			"expected `-a`, `-b`, or more, pass `--help` for usage information",
			unwrap_stderr(err),
		)
	}
}

func TestEitherOfThreeRequiredFlags2(t *testing.T) {
	a := ReqFlag(Short('a'), struct{}{}).Parser()
	b := ReqFlag(Short('b'), struct{}{}).Parser()
	c := ReqFlag(Short('c'), struct{}{}).Parser()
	p := Choice(a, b, c)
	decorated := p.ToOptions().version("1.0")

	_, ver := decorated.run_inner(newArgs("-V"))
	require.Equal(t, "Version: 1.0\n", unwrap_stdout(ver))

	// help is always generated
	_, help := decorated.run_inner(newArgs("-h"))
	const expected_help = `Usage: (-a | -b | -c)

Available options:
    -a
    -b
    -c
    -h, --help     Prints help information
    -V, --version  Prints version information
`
	require.Equal(t, expected_help, unwrap_stdout(help))

	// must specify one of the required flags
	_, err := decorated.run_inner(newArgs())
	require.Equal(t,
		"expected `-a`, `-b`, or more, pass `--help` for usage information",
		unwrap_stderr(err),
	)
}

func TestEitherOfTwoRequiredFlagsAndOneOptional(t *testing.T) {
	a := ReqFlag(Short('a'), true).Parser()
	b := ReqFlag(Short('b'), false).Parser()
	c := Short('c').Switch()
	p := a.or_else(b).Parser().or_else(c).Parser()
	decorated := p.ToOptions().version("1.0")

	_, ver := decorated.run_inner(newArgs("-V"))
	require.Equal(t, "Version: 1.0\n", unwrap_stdout(ver))

	// help is always generated
	_, help := decorated.run_inner(newArgs("-h"))
	const expected_help = `Usage: (-a | -b | [-c])

Available options:
    -a
    -b
    -c
    -h, --help     Prints help information
    -V, --version  Prints version information
`
	require.Equal(t, expected_help, unwrap_stdout(help))

	// fallback to default
	res, err := decorated.run_inner(newArgs())
	require.Nil(t, err)
	require.True(t, !res)
}

func TestFallbackWithOk(t *testing.T) {
	parser := Argument(Short('a'),
		"ARG", parseUint32).Parser().
		fallback_with(func() (uint32, Error) { return 10, nil }).Parser().
		ToOptions()

	{
		r, err := parser.run_inner(newArgs("-a", "1"))
		require.Nil(t, err)
		require.Equal(t, uint32(1), r)
	}

	{
		r, err := parser.run_inner(newArgs())
		require.Nil(t, err)
		require.Equal(t, uint32(10), r)
	}
}

func TestFallbackWithErr(t *testing.T) {
	parser := Argument(Short('a'),
		"ARG", parseUint32).Parser().
		fallback_with(func() (uint32, Error) { return 0, ErrorParseFail("nope") }).Parser().
		ToOptions()

	{
		r, err := parser.run_inner(newArgs("-a", "1"))
		require.Nil(t, err)
		require.Equal(t, uint32(1), r)
	}

	{
		_, r := parser.run_inner(newArgs("-a", "x"))
		require.Equal(t, "couldn't parse `x`: invalid digit found in string", unwrap_stderr(r))
	}

	{
		_, r := parser.run_inner(newArgs())
		require.Equal(t, "nope", unwrap_stderr(r))
	}
}

func TestDefaultArguments(t *testing.T) {
	a := Argument(Short('a'), "ARG", parseInt32).Parser().Fallback(42).Parser()
	decorated := a.ToOptions()

	{
		_, help := decorated.run_inner(newArgs("-h"))
		const expected_help = `Usage: [-a=ARG]

Available options:
    -a=ARG
    -h, --help  Prints help information
`
		require.Equal(t, expected_help, unwrap_stdout(help))
	}

	{
		_, err := decorated.run_inner(newArgs("-a", "x12"))
		const expected_err = "couldn't parse `x12`: invalid digit found in string"
		require.Equal(t, expected_err, unwrap_stderr(err))
	}

	{
		_, err := decorated.run_inner(newArgs("-a"))
		const expected_err = "`-a` requires an argument `ARG`"
		require.Equal(t, expected_err, unwrap_stderr(err))
	}
}

func TestParseErrors(t *testing.T) {
	decorated := Argument(Short('a'), "ARG", parseInt32).Parser().ToOptions()

	{
		_, err := decorated.run_inner(newArgs("-a", "123x"))
		const expected_err = "couldn't parse `123x`: invalid digit found in string"
		require.Equal(t, expected_err, unwrap_stderr(err))
	}

	{
		_, err := decorated.run_inner(newArgs("-b", "123x"))
		const expected_err = "expected `-a=ARG`, got `-b`. Pass `--help` for usage information"
		require.Equal(t, expected_err, unwrap_stderr(err))
	}

	{
		_, err := decorated.run_inner(newArgs("-a", "123", "-b"))
		const expected_err = "`-b` is not expected in this context"
		require.Equal(t, expected_err, unwrap_stderr(err))
	}
}

func TestCustomUsage(t *testing.T) {
	a := Argument(Short('a').Long("long"), "ARG", parseString).Parser()
	parser := a.ToOptions().Usage(newDocFromString("Usage: -a <ARG> or --long <ARG>"))
	_, help := parser.run_inner(newArgs("--help"))
	const expected_help = `Usage: -a <ARG> or --long <ARG>

Available options:
    -a, --long=ARG
    -h, --help      Prints help information
`
	require.Equal(t, expected_help, unwrap_stdout(help))
}

func TestLongUsageString(t *testing.T) {
	a := Argument(Short('a').
		Long("a-very-long-flag-with"),
		"ARG", parseString)
	b := Argument(Short('b').
		Long("b-very-long-flag-with"),
		"ARG", parseString)
	c := Argument(Short('c').
		Long("c-very-long-flag-with"),
		"ARG", parseString)
	d := Argument(Short('d').
		Long("d-very-long-flag-with"),
		"ARG", parseString)
	e := Argument(Short('e').
		Long("e-very-long-flag-with"),
		"ARG", parseString)
	f := Argument(Short('f').
		Long("f-very-long-flag-with"),
		"ARG", parseString)

	parser := Pair(Pair(Pair(Pair(Pair(a, b), c), d), e), f).Parser().ToOptions()

	_, help := parser.run_inner(newArgs("--help"))

	const expected_help = `Usage: -a=ARG -b=ARG -c=ARG -d=ARG -e=ARG -f=ARG

Available options:
    -a, --a-very-long-flag-with=ARG
    -b, --b-very-long-flag-with=ARG
    -c, --c-very-long-flag-with=ARG
    -d, --d-very-long-flag-with=ARG
    -e, --e-very-long-flag-with=ARG
    -f, --f-very-long-flag-with=ARG
    -h, --help  Prints help information
`

	require.Equal(t, expected_help, unwrap_stdout(help))
	_, err := parser.run_inner(newArgs("-a", "-b"))
	require.Equal(t,
		"expected `-f=ARG`, got `-a`. Pass `--help` for usage information",
		// TODO: // "`-a` requires an argument `ARG`, got a flag `-b`, try `-a=-b` to use it as an argument",
		unwrap_stderr(err),
	)
}

func TestGroupHelpArgs(t *testing.T) {
	t.Skip() // TODO: fix
	a := Short('a').Help("flag A, related to B").Switch()
	b := Short('b').Help("flag B, related to A").Switch()
	c := Short('c').Help("flag C, unrelated").Switch()
	ab := Pair(a, b).Parser().GroupHelp("Explanation applicable for both A and B:")
	parser := Pair(ab, c).Parser().ToOptions()

	_, help := parser.run_inner(newArgs("--help"))
	const expected_help = `Usage: [-a] [-b] [-c]

Explanation applicable for both A and B:
    -a          flag A, related to B
    -b          flag B, related to A

Available options:
    -c          flag C, unrelated
    -h, --help  Prints help information
`
	require.Equal(t, expected_help, unwrap_stdout(help))
}

func TestGroupHelpCommands(t *testing.T) {
	t.Skip() // TODO: fix
	a := Short('a').
		Switch().
		ToOptions().
		Command("cmd_a").
		Help("command that does A").Parser()
	b := Short('a').
		Switch().
		ToOptions().
		Command("cmd_b").
		Help("command that does B").Parser()
	c := Short('a').
		Switch().
		ToOptions().
		Command("cmd_c").
		Help("command that does C").Parser()
	parser := Choice(Choice(a, b).GroupHelp("Explanation applicable for both A and B:").Parser(), c).Parser().ToOptions()

	_, help := parser.run_inner(newArgs("--help"))
	const expected_help = `Usage: COMMAND ...

Explanation applicable for both A and B:
    cmd_a       command that does A
    cmd_b       command that does B

Available options:
    -h, --help  Prints help information

Available commands:
    cmd_c       command that does C
`
	require.Equal(t, expected_help, unwrap_stdout(help))
}

func TestFromSeveralAlternativesPickMoreMeaningful(t *testing.T) {
	t.Skip() // TODO: fix

	a := ReqFlag(Short('a'), struct{}{}).Parser()
	b := ReqFlag(Short('b'), struct{}{}).Parser()
	c := ReqFlag(Short('c'), struct{}{}).Parser()
	parser := Choice(a, b, c).ToOptions()

	_, err1 := parser.run_inner(newArgs("-a", "-b"))
	require.Equal(t, unwrap_stderr(err1), "`-b` cannot be used at the same time as `-a`")

	_, err2 := parser.run_inner(newArgs("-b", "-a"))
	require.Equal(t, unwrap_stderr(err2), "`-a` cannot be used at the same time as `-b`")

	_, err3 := parser.run_inner(newArgs("-c", "-a"))
	require.Equal(t, unwrap_stderr(err3), "`-a` cannot be used at the same time as `-c`")

	_, err4 := parser.run_inner(newArgs("-a", "-c"))
	require.Equal(t, unwrap_stderr(err4), "`-c` cannot be used at the same time as `-a`")

	_, err5 := parser.run_inner(newArgs("-c", "-b", "-a"))
	require.Equal(t, unwrap_stderr(err5), "`-b` cannot be used at the same time as `-c`")
}

func TestSubcommands(t *testing.T) {
	t.Skip() // TODO: fix
	bar := Short('b').Switch()

	bar_cmd := bar.ToOptions().Descr("This is local info").Command("bar")

	parser := bar_cmd.Parser().ToOptions().Descr("This is global info")

	{
		_, help := parser.run_inner(newArgs("--help"))
		const expected_help = `This is global info

Usage: COMMAND ...

Available options:
    -h, --help  Prints help information

Available commands:
    bar         This is local info
`
		require.Equal(t, expected_help, unwrap_stdout(help))
	}

	{
		_, help := parser.run_inner(newArgs("bar", "--help"))
		const expected_help = `This is local info

Usage: bar [-b]

Available options:
    -b
    -h, --help  Prints help information
`
		require.Equal(t, expected_help, unwrap_stdout(help))
	}
}

func TestMultipleAliases(t *testing.T) {
	t.Skip() // TODO: fix
	a := ReqFlag(Short('a').Short('b').Short('c'), struct{}{}).Parser()
	parser := a.ToOptions()

	_, help := parser.run_inner(newArgs("--help"))
	const expected_help = `Usage: -a

	Available options:

		-a
		-h, --help  Prints help information

	`

	require.Equal(t, expected_help, unwrap_stdout(help))
	{
		_, err := parser.run_inner(newArgs("-a"))
		require.Nil(t, err)
	}
	{
		_, err := parser.run_inner(newArgs("-b"))
		require.Nil(t, err)
	}
	{
		_, err := parser.run_inner(newArgs("-c"))
		require.Nil(t, err)
	}
}

// mod git {
//     use std::path::PathBuf;

//     use super::*;

//     #[derive(Debug, Clone)]
//     #[allow(dead_code)]
//     enum Opt {
//         Fetch {
//             dry_run: bool,
//             all: bool,
//             repository: String,
//         },
//         Add {
//             interactive: bool,
//             all: bool,
//             files: Vec<PathBuf>,
//         },
//     }

//     func setup() -> OptionParser<Opt> {
//         dry_run := long("dry_run").switch();
//         all := long("all").switch();
//         repository := positional::<String>("SRC").fallback("origin");
//         fetch := construct(Opt::Fetch {
//             dry_run,
//             all,
//             repository
//         });
//         fetch_inner := fetch
//             .ToOptions()
//             .descr("fetches branches from remote repository");
//         fetch_cmd := fetch_inner.command("fetch");

//         interactive := short('i').switch();
//         all := long("all").switch();
//         files := positional::<PathBuf>("FILE").many();
//         add := construct(Opt::Add {
//             interactive,
//             all,
//             files
//         });
//         add_inner := add.ToOptions().descr("add files to the staging area");
//         add_cmd := add_inner.command("add");

//         construct([fetch_cmd, add_cmd))
//             .ToOptions()
//             .descr("The stupid content tracker")
//     }

//     func no_command() {
//         parser := setup();

//         expected_err := "expected `COMMAND ...`, pass `--help` for usage information";
//         require.Equal(t,
//             expected_err,
//             parser.run_inner(newArgs()).unwrap_err().unwrap_stderr()
//         );
//     }

//     func root_help() {
//         parser := setup();
//         expected_help := "\
// The stupid content tracker

// Usage: COMMAND ...

// Available options:
//     -h, --help  Prints help information

// Available commands:
//     fetch       fetches branches from remote repository
//     add         add files to the staging area
// ";

//         require.Equal(t,
//             expected_help,
//             parser.run_inner(newArgs("--help")).unwrap_err().unwrap_stdout()
//         );
//     }

//     func fetch_help() {
//         parser := setup();
//         expected_help := "\
// fetches branches from remote repository

// Usage: fetch [--dry_run] [--all] [SRC]

// Available options:
//         --dry_run
//         --all
//     -h, --help     Prints help information
// ";
//         require.Equal(t,
//             expected_help,
//             parser
//                 .run_inner(newArgs("fetch", "--help"))
//                 .unwrap_err()
//                 .unwrap_stdout()
//         );
//     }

//     func add_help() {
//         parser := setup();
//         expected_help := "\
// add files to the staging area

// Usage: add [-i] [--all] [FILE]...

// Available options:
//     -i
//         --all
//     -h, --help  Prints help information
// ";
//         require.Equal(t,
//             expected_help,
//             parser
//                 .run_inner(newArgs("add", "--help"))
//                 .unwrap_err()
//                 .unwrap_stdout()
//         );
//     }
// }

// func arg_bench() {
//     use std::path::PathBuf;

//     #[derive(Clone, Debug, PartialEq, Eq)]
//     struct AppArgs {
//         number: uint32,
//         opt_number: Option<uint32>,
//         width: uint32,
//         input: Vec<PathBuf>,
//     }

//     number := long("number")
//         .help("Sets a number\nin two lines")
//         .argument::<uint32>("number");

//     opt_number := long("opt-number")
//         .help("Sets an optional number")
//         .argument::<uint32>("opt-number")
//         .optional();

//     width := long("width")
//         .help("Sets width")
//         .argument::<uint32>("width")
//         .guard(|n| *n > 0, "Width must be positive")
//         .fallback(10);

//     input := positional::<PathBuf>("INPUT").many();

//     parser := construct(AppArgs {
//         number,
//         opt_number,
//         width,
//         input
//     })
//     .ToOptions();

//     require.Equal(t,
//         AppArgs {
//             number: 42,
//             opt_number: Invalid(),
//             width: 10,
//             input: vec[PathBuf::from("foo"), PathBuf::from("foo2")],
//         },
//         parser
//             .run_inner(newArgs("--number", "42", "foo", "foo2"))
//             .unwrap()
//     );

//     require.Equal(t,
//         AppArgs {
//             number: 42,
//             opt_number: Invalid(),
//             width: 10,
//             input: Vec::new()
//         },
//         parser.run_inner(newArgs("--number", "42")).unwrap()
//     );

//     drop(parser);
// }

func TestSimpleCargoHelper(t *testing.T) {
	a := Short('a').Long("AAAAA").Help("two lines\nof help").Switch()
	b := Short('b').Switch()
	parser := Pair(a, b).Parser()
	decorated := cargo_helper("simple", parser).
		ToOptions().
		Descr("this is a test")

	// cargo run variant
	{
		ok, err := decorated.run_inner(newArgs("-a"))
		require.Nil(t, err)
		require.Equal(t, fun.Pair[bool, bool]{true, false}, ok)
	}

	// cargo simple variant
	{
		ok, err := decorated.run_inner(newArgs("simple", "-b"))
		require.Nil(t, err)
		require.Equal(t, fun.Pair[bool, bool]{false, true}, ok)
	}

	{
		_, err := decorated.
			run_inner(newArgs("-a", "-a"))
		require.Equal(t,
			"argument `-a` cannot be used multiple times in this context",
			unwrap_stderr(err),
		)
	}

	{
		_, help := decorated.run_inner(newArgs("-h"))
		const expected_help = `this is a test

Usage: [-a] [-b]

Available options:
    -a, --AAAAA  two lines of help
    -b
    -h, --help   Prints help information
`
		require.Equal(t, expected_help, unwrap_stdout(help))
	}
}

// func TestLongPathInConstruct(t*testing.T) {
//     a := Short('a').Switch();
//     _ = construct(std::option::Option::Some(a));

//     b := Short('b').Switch();
//     _ = construct!(::std::option::Option::Some(b));
// }

func TestHiddenEnv(t *testing.T) {
	name := "BPAF_SECRET_API_KEY2"
	visible := Argument(Long("key").
		Help("use this secret key\n two lines"),
		"KEY", parseString).Parser()
	hidden := Argument(Env(name), "KEY", parseString).Parser()
	parser := Choice(visible, hidden).ToOptions()

	_, help := parser.run_inner(newArgs("-h"))
	const expected = `Usage: --key=KEY

Available options:
    --key=KEY   use this secret key
                two lines
    -h, --help  Prints help information
`
	require.Equal(t, expected, unwrap_stdout(help))

	_, r := parser.run_inner(newArgs())
	require.Equal(t, "environment variable `BPAF_SECRET_API_KEY2` is not set", unwrap_stderr(r))
}

func TestEnvVariable(t *testing.T) {
	const name = "BPAF_SECRET_API_KEY"
	parser := Argument(Long("key").
		Env(name).
		Help("use this secret key\ntwo lines"), "KEY", parseString).
		Parser().
		ToOptions()

	{
		_, help := parser.run_inner(newArgs("-h"))
		const expected_help = `Usage: --key=KEY

Available options:
    --key=KEY   use this secret key two lines
                [env:BPAF_SECRET_API_KEY: N/A]
    -h, --help  Prints help information
`
		require.Equal(t, expected_help, unwrap_stdout(help))
	}
	t.Setenv(name, "top s3cr3t")

	{
		_, help := parser.run_inner(newArgs("-h"))
		const expected_help = `Usage: --key=KEY

Available options:
    --key=KEY   use this secret key two lines
                [env:BPAF_SECRET_API_KEY = "top s3cr3t"]
    -h, --help  Prints help information
`
		require.Equal(t, expected_help, unwrap_stdout(help))
	}

	{
		res, err := parser.run_inner(newArgs("--key", "secret"))
		require.Nil(t, err)
		require.Equal(t, res, "secret")
	}

	{
		res, err := parser.run_inner(newArgs())
		require.Nil(t, err)
		require.Equal(t, res, "top s3cr3t")
	}
}

func TestDefaultPlaysNicelyWithCommand(t *testing.T) {
	t.Skip() // TODO: fix
	type Foo bool
	const (
		FooFoo Foo = true
		FooBar Foo = false
	)

	cmd := Pure(FooFoo).Parser().
		ToOptions().
		Descr("inner").
		Command("foo").
		Help("foo").Parser().
		Fallback(*new(Foo)).Parser()

	parser := cmd.ToOptions().Descr("outer")

	{
		_, help := parser.run_inner(newArgs("foo", "--help"))
		const expected_help = `inner

Usage: foo

Available options:
    -h, --help  Prints help information
`
		require.Equal(t, expected_help, unwrap_stdout(help))
	}

	{
		_, help := parser.run_inner(newArgs("--help"))
		const expected_help = `outer

Usage: [COMMAND ...]

Available options:
    -h, --help  Prints help information

Available commands:
    foo         foo
`
		require.Equal(t, expected_help, unwrap_stdout(help))
	}
}

func TestCommandWithAliases(t *testing.T) {
	t.Skip() // TODO: fix
	inner := Pure(struct{}{}).Parser().ToOptions().Descr("inner descr")
	cmd := inner.Command("foo").long("bar").short('f').short('b')
	parser := cmd.Parser().ToOptions().Descr("outer")

	{
		_, help := parser.run_inner(newArgs("--help"))
		const expected_help = `outer

Usage: COMMAND ...

Available options:
    -h, --help  Prints help information

Available commands:
    foo, f      inner descr
`
		require.Equal(t, expected_help, unwrap_stdout(help))
	}

	{
		_, help := parser.
			run_inner(newArgs("f", "--help"))
		const expected_help = `inner descr

Usage: foo

Available options:
    -h, --help  Prints help information
`
		require.Equal(t, expected_help, unwrap_stdout(help))
	}

	// hidden and visible aliases are working
	for _, alias := range []string{"foo", "f", "bar", "b"} {
		_, err := parser.run_inner(newArgs(alias))
		require.Nil(t, err)
	}

	// and "k" isn't a thing
	_, err := parser.run_inner(newArgs("k"))
	require.NotNil(t, err)
}

func TestHelpForOptions(t *testing.T) {
	a := Short('a').Help("help for\na").Switch()
	b := Argument(Short('c').
		Env("BbBbB").
		Help("help for\nb"), "B", parseString).Parser()
	c := Argument(Long("bbbbb").
		Env("ccccCCccc").
		Help("help for\nccc"), "CCC", parseString).Parser()
	parser := Triple(a, b, c).Parser().ToOptions()
	_, help := parser.run_inner(newArgs("--help"))

	const expected_help = `Usage: [-a] -c=B --bbbbb=CCC

Available options:
    -a           help for a
    -c=B         help for b
                 [env:BbBbB: N/A]
    --bbbbb=CCC  help for ccc
                 [env:ccccCCccc: N/A]
    -h, --help   Prints help information
`

	require.Equal(t, expected_help, unwrap_stdout(help))
}

func TestHelpForCommands(t *testing.T) {
	d := Pure(struct{}{}).Parser().
		ToOptions().
		Command("thing_d").
		Help("help for d\ntwo lines").Parser()
	e := Pure(struct{}{}).Parser().
		ToOptions().
		Command("thing_e").
		short('e').
		Help("help for e\ntwo lines").Parser()
	h := Pure(struct{}{}).Parser().ToOptions().Command("thing_h").Parser()
	parser := Choice(d, e, h).ToOptions()
	_, help := parser.run_inner(newArgs("--help"))

	const expected_help = `Usage: COMMAND ...

Available options:
    -h, --help  Prints help information

Available commands:
    thing_d     help for d two lines
    thing_e, e  help for e two lines
    thing_h
`
	require.Equal(t, expected_help, unwrap_stdout(help))
}

func TestManyDoesntPanic(t *testing.T) {
	parser := Map(Many(Short('a').Switch()), func(m []bool) int { return len(m) }).ToOptions()
	r, err := parser.run_inner(newArgs("-aaa"))
	require.Nil(t, err)
	require.Equal(t, r, 3)
}

func TestSomeDoesntPanic(t *testing.T) {
	parser := Map(ParseSomeParser(Short('a').Switch().Some("")), func(m []bool) int { return len(m) }).ToOptions()
	r, err := parser.run_inner(newArgs("-aaa"))
	require.Nil(t, err)
	require.Equal(t, r, 3)
}

func TestCommandResetsLeftHeadState(t *testing.T) {
	t.Skip() // TODO: fix
	type Foo fun.Option[uint32]

	a := Argument(Short('a'), "A", parseUint32).Parser().Fallback(0).Parser()
	b := ReqFlag(Short('b'), struct{}{}).Parser()

	p1 := Map(a, func(a uint32) Foo { return Foo{a, true} })
	p2 := Map(b, func(struct{}) Foo { return Foo{0, false} })
	cmd := Choice(p1, p2).
		ToOptions().
		Command("cmd").Parser().
		ToOptions()

	xx, err := cmd.run_inner(newArgs("cmd", "-b"))
	require.Nil(t, err)
	require.Equal(t, xx, Foo{})
}

func TestCommand_preserves_custom_failure_message(t *testing.T) {
	t.Skip() // TODO: fix
	const msg = "need more cheese"
	inner := Fail[struct{}](msg).Parser().ToOptions()

	_, err := inner.run_inner(newArgs())
	require.Equal(t, unwrap_stderr(err), "need more cheese")

	outer := inner.Command("feed").Parser().ToOptions()

	_, err = outer.run_inner(newArgs("feed"))
	require.Equal(t, unwrap_stderr(err), "need more cheese")
}

func TestOptional_error_handling(t *testing.T) {
	p := Optional(Argument(Short('p'), "P", parseUint32).Parser()).Parser().ToOptions()

	{
		res, err := p.run_inner(newArgs())
		require.Nil(t, err)
		require.Equal(t, res, Invalid[uint32]())
	}

	{
		res, err := p.run_inner(newArgs("-p", "3"))
		require.Nil(t, err)
		require.Equal(t, res, Valid[uint32](3))
	}

	{
		_, res := p.run_inner(newArgs("-p", "pi"))
		require.Equal(t, unwrap_stderr(res), "couldn't parse `pi`: invalid digit found in string")
	}
}

func TestMany_error_handling(t *testing.T) {
	p := Many(Argument(Short('p'), "P", parseUint32).Parser()).ToOptions()

	{
		res, err := p.run_inner(newArgs())
		require.Nil(t, err)
		require.Equal(t, res, []uint32{})
	}

	{
		res, err := p.run_inner(newArgs("-p", "3"))
		require.Nil(t, err)
		require.Equal(t, res, []uint32{3})
	}

	{
		_, res := p.run_inner(newArgs("-p", "pi"))
		require.Equal(t, unwrap_stderr(res), "couldn't parse `pi`: invalid digit found in string")
	}
}

func TestFailure_is_not_stupid_1(t *testing.T) {
	a := Argument(Short('a'), "A", parseUint32)
	b := Parse(Pure(struct{}{}).Parser(), func(struct{}) (string, Error) { return "", ErrorParseFail("nope") })
	parser := Pair(a, b).Parser().ToOptions()

	_, res := parser.run_inner(newArgs("-a", "42"))
	require.Equal(t, unwrap_stderr(res), "couldn't parse: nope")
}

func TestFailure_is_not_stupid_2(t *testing.T) {
	a := Argument(Short('a'), "A", parseUint32)
	b := Argument(Short('b'), "B", parseUint32)
	parser := Parse(Pair(a, b).Parser(), func(fun.Pair[uint32, uint32]) (string, Error) { return "", ErrorParseFail("nope") }).Parser().ToOptions()

	_, res := parser.run_inner(newArgs("-a", "42", "-b", "42"))
	require.Equal(t, unwrap_stderr(res), "couldn't parse: nope")
}

func TestNo_fallback_out_of_command_parser(t *testing.T) {
	t.Skip() // TODO: fix
	alt1 := Positional("NAME", parseString).Parser().ToOptions().Command("cmd")
	alt2 := Pure("")
	parser := Choice(alt1.Parser(), alt2.Parser()).ToOptions()

	{
		_, res := parser.run_inner(newArgs("cmd"))
		require.Equal(t, unwrap_stderr(res), "expected `NAME`, pass `--help` for usage information")
	}

	{
		res, err := parser.run_inner(newArgs("cmd", "a"))
		require.Nil(t, err)
		require.Equal(t, res, "a")
	}

	{
		res, err := parser.run_inner(newArgs())
		require.Nil(t, err)
		require.Equal(t, res, "")
	}
}

func TestDid_you_mean_switch(t *testing.T) {
	a := Short('f').Long("flag").Switch()
	b := Short('p').Long("plag").Switch()
	parser := Choice(a.Parser(), b.Parser()).ToOptions()

	{
		_, res := parser.run_inner(newArgs("--fla"))
		require.Equal(t, unwrap_stderr(res), "no such flag: `--fla`, did you mean `--flag`?")
	}

	{
		_, res := parser.run_inner(newArgs("flag"))
		require.Equal(t, unwrap_stderr(res), "no such command or positional: `flag`, did you mean `--flag`?")
	}

	{
		_, res := parser.run_inner(newArgs("--pla"))
		require.Equal(t, unwrap_stderr(res), "no such flag: `--pla`, did you mean `--plag`?")
	}

	{
		_, res := parser.run_inner(newArgs("--p"))
		require.Equal(t, unwrap_stderr(res), "no such flag: `--p` (with two dashes), did you mean `-p`?")
	}
}

func TestDid_you_mean_req_flag(t *testing.T) {
	parser := ReqFlag(Long("flag"), struct{}{}).Parser().ToOptions()
	_, res := parser.run_inner(newArgs("--fla"))
	require.Equal(t, unwrap_stderr(res), "no such flag: `--fla`, did you mean `--flag`?")
}

func TestDid_you_mean_argument(t *testing.T) {
	t.Skip() // TODO: fix
	parser := Argument(Long("flag"), "VAL", parseString).Parser().ToOptions()

	{
		_, res := parser.run_inner(newArgs("--fla"))
		require.Equal(t, unwrap_stderr(res), "no such flag: `--fla`, did you mean `--flag`?")
	}

	{
		_, res := parser.run_inner(newArgs("--flg=hellop"))
		require.Equal(t, unwrap_stderr(res), "no such flag: `--flg`, did you mean `--flag`?")
	}
}

func TestDid_you_mean_command(t *testing.T) {
	parser := Pure(struct{}{}).Parser().
		ToOptions().
		Command("command").
		short('c').Parser().
		ToOptions()

	{
		_, res := parser.run_inner(newArgs("comman"))
		require.Equal(t, unwrap_stderr(res), "no such command or positional: `comman`, did you mean `command`?")
	}

	{
		_, res := parser.run_inner(newArgs("--comman"))
		require.Equal(t, unwrap_stderr(res), "no such flag: `--comman`, did you mean `command`?")
	}
}

func TestDid_you_mean_two_and_arguments(t *testing.T) {
	t.Skip() // TODO: fix
	a := Long("flag").Switch()
	b := Long("parameter").Switch()
	parser := cargo_helper("cmd", Pair(a, b).Parser()).ToOptions()

	{
		_, r := parser.run_inner(newArgs("--flag", "--parametr"))
		require.Equal(t, unwrap_stderr(r), "no such flag: `--parametr`, did you mean `--parameter`?")
	}

	{
		_, r := parser.run_inner(newArgs("--flag", "--paramet=value"))
		require.Equal(t, unwrap_stderr(r), "no such flag: `--paramet`, did you mean `--parameter`?")
	}

	{
		_, r := parser.run_inner(newArgs("--parameter", "--flg"))
		require.Equal(t, unwrap_stderr(r), "no such flag: `--flg`, did you mean `--flag`?")
	}

	{
		_, r := parser.run_inner(newArgs("--fla"))
		require.Equal(t, unwrap_stderr(r), "no such flag: `--fla`, did you mean `--flag`?")
	}
}

func TestDid_you_mean_two_or_arguments(t *testing.T) {
	a := Long("flag").Switch()
	b := Long("parameter").Switch()
	parser := cargo_helper("cmd", Choice(a, b)).ToOptions()

	{
		_, r := parser.run_inner(newArgs("--fla"))
		require.Equal(t, unwrap_stderr(r), "no such flag: `--fla`, did you mean `--flag`?")
	}

	{
		_, r := parser.run_inner(newArgs("--flag", "--parametr"))
		require.Equal(t, unwrap_stderr(r), "no such flag: `--parametr`, did you mean `--parameter`?")
	}

	{
		_, r := parser.run_inner(newArgs("--parametr", "--flag"))
		require.Equal(t, unwrap_stderr(r), "no such flag: `--parametr`, did you mean `--parameter`?")
	}

	t.Skip() // TODO: fix
	{
		_, r := parser.run_inner(newArgs("--parameter", "--flag"))
		require.Equal(t, unwrap_stderr(r), "`--flag` cannot be used at the same time as `--parameter`")
	}

	{
		_, r := parser.run_inner(newArgs("--flag", "--parameter"))
		require.Equal(t, unwrap_stderr(r), "`--parameter` cannot be used at the same time as `--flag`")
	}
}

// problematic steps look something like this:
// - "asm" is parsed as expected
// - "-t x" is consumed as expected
// - parsing of "x" fails
// - ParseWith rollbacks the arguments state - "asm" is back
// - suggestion looks for something it can complain at and finds "asm"
//
// parse/guard failures should "taint" the arguments and disable the suggestion logic

func TestCargo_show_asm_issue_guard(t *testing.T) {
	t.Skip() // TODO: fix
	target_dir := Argument(Short('t'), "T", parseString).Parser().Guard(func(string) bool { return false }, "nope")
	verbosity := Short('v').Switch()
	inner := Pair(target_dir, verbosity)
	parser := cargo_helper("asm", inner.Parser()).ToOptions()

	{
		_, res := parser.run_inner(newArgs("asm", "-t", "x"))
		require.Equal(t, unwrap_stderr(res), "`x`: nope")
	}

	{
		_, res := parser.run_inner(newArgs("-t", "x"))
		require.Equal(t, unwrap_stderr(res), "`x`: nope")
	}
}

func TestCargo_show_asm_issue_from_str(t *testing.T) {
	t.Skip() // TODO: fix
	target_dir := Argument(Short('t'), "T", parseUint)
	verbosity := Short('v').Switch()
	inner := Pair(target_dir, verbosity)
	parser := cargo_helper("asm", inner.Parser()).ToOptions()

	{
		_, res := parser.run_inner(newArgs("asm", "-t", "x"))
		require.Equal(t, unwrap_stderr(res), "couldn't parse `x`: invalid digit found in string")
	}

	{
		_, res := parser.run_inner(newArgs("-t", "x"))
		require.Equal(t, unwrap_stderr(res), "couldn't parse `x`: invalid digit found in string")
	}
}

func TestCargo_show_asm_issue_parse(t *testing.T) {
	t.Skip() // TODO: fix
	target_dir := Parse(Argument(Short('t'), "T", parseString).Parser(),
		func(string) (struct{}, Error) { return struct{}{}, ErrorParseFail("nope") })

	verbosity := Short('v').Switch()
	inner := Pair(target_dir, verbosity)
	parser := cargo_helper("asm", inner.Parser()).ToOptions()

	{
		_, res := parser.run_inner(newArgs("asm", "-t", "x"))
		require.Equal(t, unwrap_stderr(res), "couldn't parse `x`: nope")
	}

	{
		_, res := parser.run_inner(newArgs("-t", "x"))
		require.Equal(t, unwrap_stderr(res), "couldn't parse `x`: nope")
	}
}

// problematic case looks something like this:
// "asm" is consumed
// "--fla" is not consumed, but the output is not tainted
// parser fails and arguments are restored.

func TestCargo_show_asm_issue_unknown_switch(t *testing.T) {
	target := Long("flag").Switch()

	verbosity := Short('v').Switch()
	parser := cargo_helper("asm", Pair(target, verbosity).Parser()).ToOptions()

	{
		_, res := parser.run_inner(newArgs("asm", "--fla"))
		require.Equal(t, unwrap_stderr(res), "no such flag: `--fla`, did you mean `--flag`?")
	}

	{
		_, res := parser.run_inner(newArgs("--fla"))
		require.Equal(t, unwrap_stderr(res), "no such flag: `--fla`, did you mean `--flag`?")
	}
}

func TestDid_you_mean_inside_command(t *testing.T) {
	t.Skip() // TODO: fix
	a := Long("flag").Switch()
	b := Long("parameter").Switch()
	parser := Choice(a, b).ToOptions().Command("cmd").Parser().ToOptions()

	{
		_, r := parser.run_inner(newArgs("--fla"))
		require.Equal(t, unwrap_stderr(r), "expected `COMMAND ...`, got `--fla`. Pass `--help` for usage information")
	}

	{
		_, r := parser.run_inner(newArgs("cmd", "--fla"))
		require.Equal(t, unwrap_stderr(r), "no such flag: `--fla`, did you mean `--flag`?")
	}

	{
		_, r := parser.run_inner(newArgs("cmd", "--flag", "--parametr"))
		require.Equal(t, unwrap_stderr(r), "no such flag: `--parametr`, did you mean `--parameter`?")
	}

	{
		_, r := parser.run_inner(newArgs("cmd", "--parametr", "--flag"))
		require.Equal(t, unwrap_stderr(r), "no such flag: `--parametr`, did you mean `--parameter`?")
	}

	{
		_, r := parser.run_inner(newArgs("cmd", "--parameter", "--flag"))
		require.Equal(t, unwrap_stderr(r), "`--flag` cannot be used at the same time as `--parameter`")
	}

	{
		_, r := parser.run_inner(newArgs("cmd", "--flag", "--parameter"))
		require.Equal(t, unwrap_stderr(r), "`--parameter` cannot be used at the same time as `--flag`")
	}
}

func TestCombine_flags_by_order(t *testing.T) {
	t.Skip() // TODO: fix
	a := ReqFlag(Short('a'), true)
	b := ReqFlag(Short('A'), false)
	parser := Many(Choice(a.Parser(), b.Parser())).ToOptions()

	r, err := parser.run_inner(newArgs("-a", "-A", "-A", "-A", "-a"))
	require.Nil(t, err)
	require.Equal(t, r, []bool{true, false, false, false, true})
}

func TestParse_many_errors_positional(t *testing.T) {
	t.Skip() // TODO: fix
	p := Many(Positional("N", parseUint32).Parser()).ToOptions()

	{
		r, err := p.run_inner(newArgs("1", "2", "3"))
		require.Nil(t, err)
		require.Equal(t, r, []uint32{1, 2, 3})
	}

	{
		_, r := p.run_inner(newArgs("1", "2", "x"))
		require.Equal(t, unwrap_stderr(r), "couldn't parse `x`: invalid digit found in string")
	}
}

func TestParse_collect_flag(t *testing.T) {
	p := ParseCollectParser(Argument(Short('p'), "N", parseUint32).Parser().collect()).ToOptions()

	{
		r, err := p.run_inner(newArgs("-p", "1", "-p", "2"))
		require.Nil(t, err)
		require.Equal(t, r, []uint32{1, 2})
	}

	{
		_, r := p.run_inner(newArgs("-p", "1", "-p", "x"))
		require.Equal(t, unwrap_stderr(r), "couldn't parse `x`: invalid digit found in string")
	}
}

func TestParse_many_errors_flag(t *testing.T) {
	p := Many(Argument(Short('p'), "N", parseUint32).Parser()).ToOptions()

	{
		r, err := p.run_inner(newArgs("-p", "1", "-p", "2"))
		require.Nil(t, err)
		require.Equal(t, r, []uint32{1, 2})
	}

	{
		_, r := p.run_inner(newArgs("-p", "1", "-p", "x"))
		require.Equal(t, unwrap_stderr(r), "couldn't parse `x`: invalid digit found in string")
	}
}

func TestCommand_with_req_parameters(t *testing.T) {
	t.Skip() // TODO: fix
	p := Positional("X", parseString).Parser().
		ToOptions().
		Command("cmd").Parser().
		Fallback("").Parser().
		ToOptions()

	_, r := p.run_inner(newArgs("cmd"))
	require.Equal(t, unwrap_stderr(r), "expected `X`, pass `--help` for usage information")
}

func TestSuggestion_for_equals_1(t *testing.T) {
	parser := Argument(Short('p').Long("par"), "P", parseString).Parser().ToOptions()

	{
		_, r := parser.run_inner(newArgs("-p", "--bar"))
		require.Equal(t, unwrap_stderr(r), "`-p` requires an argument `P`, got a flag `--bar`, try `-p=--bar` to use it as an argument")
	}

	{
		_, r := parser.run_inner(newArgs("--par", "--bar"))
		require.Equal(t, unwrap_stderr(r), "`--par` requires an argument `P`, got a flag `--bar`, try `--par=--bar` to use it as an argument")
	}

	{
		_, r := parser.run_inner(newArgs("--par", "--bar=baz"))
		require.Equal(t, unwrap_stderr(r), "`--par` requires an argument `P`, got a flag `--bar=baz`, try `--par=--bar=baz` to use it as an\nargument")
	}
}

func TestDouble_dash_is_pos_only_just_once(t *testing.T) {
	parser := Many(Positional("POS", parseString).Parser()).ToOptions()

	{
		r, err := parser.run_inner(newArgs("--"))
		require.Nil(t, err)
		require.Equal(t, r, []string{})
	}

	{
		r, err := parser.run_inner(newArgs("--", "--"))
		require.Nil(t, err)
		require.Equal(t, r, []string{"--"})
	}
}

func TestReject_fbar(t *testing.T) {
	parser := Argument(Short('f'), "F", parseString).Parser().ToOptions()

	{
		_, r := parser.run_inner(newArgs("-fbar", "baz"))
		require.Equal(t, unwrap_stderr(r), "`baz` is not expected in this context")
	}

	{
		r, err := parser.run_inner(newArgs("-fbar"))
		require.Nil(t, err)
		require.Equal(t, r, "bar")
	}
}

func TestCustom_usage_override_fixed(t *testing.T) {
	parser := Short('p').Switch().ToOptions().Usage(newDocFromString("Usage: hey [-p]"))
	_, r := parser.run_inner(newArgs("--help"))
	require.Equal(t, unwrap_stdout(r), "Usage: hey [-p]\n\nAvailable options:\n    -p\n    -h, --help  Prints help information\n")
}

func TestCustom_usage_override_with_fn(t *testing.T) {
	parser := Short('p').Switch().ToOptions().WithUsage(func(b Doc) Doc {
		buf := Doc{}
		buf.Text("Usage: hey ")
		buf.Doc(b)
		return buf
	})
	_, r := parser.run_inner(newArgs("--help"))
	require.Equal(t, unwrap_stdout(r), "Usage: hey [-p]\n\nAvailable options:\n    -p\n    -h, --help  Prints help information\n")
}

func TestCatch_works(t *testing.T) {
	type A struct {
		isNum bool
		Num   uint
		Str   string
	}
	ANum := func(a uint) A { return A{true, a, ""} }
	AStr := func(s string) A { return A{false, 0, s} }
	a_n := Optional(Map(Argument(Short('a'),
		"A", parseUint).Parser(),
		ANum)).
		Catch()
	a_s := Optional(Map(Argument(Short('a'),
		"A", parseString).Parser(),
		AStr).
		Hide())
	parser := Choice(a_n.Parser(), a_s.Parser()).ToOptions()

	{
		r, err := parser.run_inner(newArgs("-a", "1"))
		require.Nil(t, err)
		require.Equal(t, r, Valid(ANum(1)))
	}

	t.Skip() // TODO: fix
	{
		r, err := parser.run_inner(newArgs("-a", "x1"))
		require.Nil(t, err)
		require.Equal(t, r, Valid(AStr("x1")))
	}
}

func TestSneaky_command(t *testing.T) {
	t.Skip() // TODO: fix
	type Cmd struct {
		isA bool
		x   bool
	}
	CmdA := func(x bool) Cmd { return Cmd{true, x} }
	CmdB := func(x bool) Cmd { return Cmd{false, x} }
	cmd := Map(Short('f').
		Switch().
		ToOptions().
		Command("hello").Parser(),
		CmdA)
	b := Map(Short('f').Switch(), CmdB)

	sneaky := true // TODO: ???

	maybe_cmd := cmd
	if !sneaky {
		maybe_cmd = Fail[Cmd]("no sneaky").Parser()
	}
	parser := Choice(maybe_cmd, b).ToOptions()

	r, err := parser.run_inner(newArgs("hello"))
	require.Nil(t, err)
	require.Equal(t, r, CmdA(false))
}

func TestDefault_for_many(t *testing.T) {
	parser := Map(Many(Positional("ROOTS", parseString).Parser()),
		func(items []string) []string {
			if len(items) == 0 {
				return []string{"."}
			}
			return items
		}).
		ToOptions()

	r, err := parser.run_inner(newArgs())
	require.Nil(t, err)
	require.Equal(t, r, []string{"."})
}

func TestParse_option_catch(t *testing.T) {
	type A struct {
		isU32 bool
		U32   uint32
		S     string
	}
	AU32 := func(a uint32) A { return A{true, a, ""} }
	AS := func(s string) A { return A{false, 0, s} }
	a1 := Optional(Map(Argument(Short('a'), "N", parseUint32).Parser(), AU32)).Catch()
	a2 := Optional(Map(Argument(Short('a'), "S", parseString).Parser(), AS)).Catch()
	parser := Choice(a1.Parser(), a2.Parser()).ToOptions()

	{
		r, err := parser.run_inner(newArgs("-a", "10"))
		require.Nil(t, err)
		require.Equal(t, r, Valid(AU32(10)))
	}

	t.Skip() // TODO: fix
	{
		r, err := parser.run_inner(newArgs("-a", "x"))
		require.Nil(t, err)
		require.Equal(t, r, Valid(AS("x")))
	}

	{
		r, err := parser.run_inner(newArgs())
		require.Nil(t, err)
		require.Equal(t, r, Invalid[A]())
	}
}

func TestParse_some_catch(t *testing.T) {
	type A struct {
		isU32 bool
		U32   uint32
		S     string
	}
	AU32 := func(a uint32) A { return A{true, a, ""} }
	AS := func(s string) A { return A{false, 0, s} }
	a1 := ParseSomeParser(Map(Argument(Short('a'),
		"N", parseUint32).Parser(),
		AU32).
		Some("A").
		Catch()).
		Hide()
	a2 := ParseSomeParser(Map(Argument(Short('a'), "S", parseString).Parser(), AS).Some("A").Catch()).Hide()
	parser := Choice(a1, a2).ToOptions()

	{
		r, err := parser.run_inner(newArgs("-a", "10"))
		require.Nil(t, err)
		require.Equal(t, r, []A{AU32(10)})
	}

	t.Skip() // TODO: fix
	{
		r, err := parser.run_inner(newArgs("-a", "x"))
		require.Nil(t, err)
		require.Equal(t, r, []A{AS("x")})
	}

	{
		_, r := parser.run_inner(newArgs())
		require.Equal(t, unwrap_stderr(r), "A")
	}
}

func TestEmpty_struct(t *testing.T) {
	type Foo struct{}
	parser := Pure(Foo{}).Parser().ToOptions()

	{
		r, err := parser.run_inner(newArgs())
		require.Nil(t, err)
		require.Equal(t, r, Foo{})
	}

	{
		_, r := parser.run_inner(newArgs("--help"))
		require.Equal(t, unwrap_stdout(r), "Usage: \n\nAvailable options:\n    -h, --help  Prints help information\n")
	}
}

func TestStrange_short_option(t *testing.T) {
	t.Skip() // TODO: fix
	parser := Argument(Short('O'), "ARG", parseString).Parser().ToOptions()
	r, err := parser.run_inner(newArgs("-Obits=2048"))
	require.Nil(t, err)
	require.Equal(t, r, "bits=2048")
}

func TestOptional_bool_states(t *testing.T) {
	parser := Optional(Short('a').Switch()).Parser().ToOptions()

	{
		r, err := parser.run_inner(newArgs("-a"))
		require.Nil(t, err)
		require.Equal(t, r, Valid(true))
	}

	{
		r, err := parser.run_inner(newArgs())
		require.Nil(t, err)
		require.Equal(t, r, Valid(false))
	}
}

func TestFancy_negative(t *testing.T) {
	t.Skip() // TODO: fix
	a := ReqFlag(Short('a'), struct{}{})
	b := Any("A", func(s string) (int32, bool) { res, err := strconv.ParseInt(s, 10, 32); return int32(res), err == nil })
	ab := Map(Pair(a, b).Adjacent().Parser(), func(x fun.Pair[struct{}, int32]) int32 { return x.V })

	c := Argument(Short('c'), "C", parseUint).Parser().Fallback(42)

	parser := Pair(ab, c).Parser().ToOptions()

	{
		r, err := parser.run_inner(newArgs("-a", "-10"))
		require.Nil(t, err)
		require.Equal(t, r, fun.Pair[int32, uint]{-10, 42})
	}

	{
		r, err := parser.run_inner(newArgs("-a=-20", "-c", "110"))
		require.Nil(t, err)
		require.Equal(t, r, fun.Pair[int32, uint]{-20, 110})
	}

	{
		_, r := parser.run_inner(newArgs("--help"))
		// TODO - rendering sucks once you start inventing fancy combinations and don't provide help...
		expected := `Usage: -a A [-c=C]

Available options:
  -a A

    -c=C
    -h, --help  Prints help information
`
		require.Equal(t, unwrap_stdout(r), expected)
	}
}

func TestMany_env(t *testing.T) {
	t.Setenv("USER1", "top s3cr3t")
	parser := Many(Argument(Short('v').
		Env("USER1"),
		"USER", parseString).Parser()).
		ToOptions()
	r, err := parser.run_inner(newArgs())
	require.Nil(t, err)
	require.Equal(t, r, []string{"top s3cr3t"})
}

func TestEnv_hidden_arg(t *testing.T) {
	t.Setenv("USER1", "top s3cr3t")
	parser := Argument(Env("USER1"), "USER", parseString).Parser().ToOptions()
	r, err := parser.run_inner(newArgs())
	require.Nil(t, err)
	require.Equal(t, r, "top s3cr3t")
}

func TestEnv_hidden_switch(t *testing.T) {
	t.Setenv("USER1", "top s3cr3t")
	parser := Env("USER1").Switch().ToOptions()
	r, err := parser.run_inner(newArgs())
	require.Nil(t, err)
	require.True(t, r)
}

func TestEnv_hidden_flag(t *testing.T) {
	t.Setenv("USER1", "top s3cr3t")
	parser := Flag(Env("USER1"), true, false).Parser().ToOptions()
	r, err := parser.run_inner(newArgs())
	require.Nil(t, err)
	require.True(t, r)
}

func TestSome_env(t *testing.T) {
	t.Setenv("USER1", "top s3cr3t")
	parser := ParseSomeParser(Argument(Short('v').
		Env("USER1"),
		"USER", parseString).Parser().
		Some("a")).
		ToOptions()
	r, err := parser.run_inner(newArgs())
	require.Nil(t, err)
	require.Equal(t, r, []string{"top s3cr3t"})
}

func TestOption_requires_other_option1(t *testing.T) {
	t.Skip() // TODO: fix
	a := Short('a').Switch()
	b := Argument(Short('b'), "B", parseString)
	parser := Optional(Pair(a, b).Parser()).Parser().ToOptions()

	_, r := parser.run_inner(newArgs("-a"))
	require.Equal(t, unwrap_stderr(r), "expected `-b=B`, pass `--help` for usage information")
}

func TestOption_requires_other_option2(t *testing.T) {
	t.Skip() // TODO: fix
	a := Short('a').Switch()
	b := Argument(Short('b'), "B", parseString)
	parser := Optional(Pair(b, a).Parser()).Parser().ToOptions()

	_, r := parser.run_inner(newArgs("-a"))
	require.Equal(t, unwrap_stderr(r), "expected `-b=B`, pass `--help` for usage information")
}

func TestDefault_for_some(t *testing.T) {
	parser := ParseSomeParser(Positional[uint32]("ROOTS", parseUint32).Parser().
		Some("msg")).
		fallback_with(func() ([]uint32, Error) { return []uint32{1, 2, 3}, nil }).Parser().
		ToOptions()

	r, err := parser.run_inner(newArgs())
	require.Nil(t, err)

	require.Equal(t, r, []uint32{1, 2, 3})
}

func TestAdjacent_anywhere_needs_to_consume_something(t *testing.T) {
	t.Skip() // TODO: fix
	a := Short('a').Switch()
	b := Short('b').Switch()
	parser := Pair(a, b).Adjacent().Parser().ToOptions()

	{
		r, err := parser.run_inner(newArgs("-a"))
		require.Nil(t, err)
		require.Equal(t, r, fun.Pair[bool, bool]{true, false})
	}

	{
		r, err := parser.run_inner(newArgs("-b"))
		require.Nil(t, err)
		require.Equal(t, r, fun.Pair[bool, bool]{false, true})
	}
}

func TestFallback_for_some(t *testing.T) {
	parser := ParseSomeParser(Argument(Short('a'),
		"ARG", parseUint32).Parser().
		Some("potatoes")).
		Fallback([]uint32{1, 2, 3}).Parser().ToOptions()

	{
		r, err := parser.run_inner(newArgs("-a", "1"))
		require.Nil(t, err)
		require.Equal(t, r, []uint32{1})
	}

	{
		r, err := parser.run_inner(newArgs())
		require.Nil(t, err)
		require.Equal(t, r, []uint32{1, 2, 3})
	}
}

func TestFlag_like_commands(t *testing.T) {
	a := ReqFlag(Short('a'), 1).Parser().ToOptions().Command("--add")
	b := ReqFlag(Short('b'), 2).Parser().ToOptions().Command("remove")
	parser := Choice(a.Parser(), b.Parser()).ToOptions()

	{
		r, err := parser.run_inner(newArgs("--add", "-a"))
		require.Nil(t, err)
		require.Equal(t, r, 1)
	}

	t.Skip() // TODO: fix
	{
		r, err := parser.run_inner(newArgs("remove", "-b"))
		require.Nil(t, err)
		require.Equal(t, r, 2)
	}

	{
		_, r := parser.run_inner(newArgs("--help"))
		expected := "Usage: COMMAND ...\n\nAvailable options:\n    -h, --help  Prints help information\n\nAvailable commands:\n    --add\n    remove\n"
		require.Equal(t, unwrap_stdout(r), expected)
	}

	{
		_, r := parser.run_inner(newArgs("--add", "--help"))
		const expected = "Usage: --add -a\n\nAvailable options:\n    -a\n    -h, --help  Prints help information\n"
		require.Equal(t, unwrap_stdout(r), expected)
	}
}
