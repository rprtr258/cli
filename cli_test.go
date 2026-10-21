package cli

import (
	"errors"
	"strconv"
	"testing"

	"github.com/rprtr258/fun"
	"github.com/stretchr/testify/require"
)

func TestWtfShenanigans1(t *testing.T) {
	f := func(
		i_c argType,
		i_prefix string,
	) {
		t.Helper()

		i_suffix := string([]byte{0x66, 0x6f, 0xD8, 0x6f})

		dashes := ""
		switch i_c {
		case argTypeShort:
			dashes += "-"
		case argTypeLong:
			dashes += "--"
		}

		o_c, o_prefix, o_suffix, ok := splitOsArgument(dashes + i_prefix + "=" + i_suffix)
		require.True(t, ok)
		require.Equal(t, i_c, o_c)
		require.Equal(t, i_prefix, o_prefix)
		require.Equal(t, argWord(i_suffix), o_suffix.Unwrap())
	}

	f(argTypeShort, "f")
	f(argTypeLong, "foo")
	f(argTypeLong, "口水鸡")
}

func TestWtfShenanigans2(t *testing.T) {
	f := func(
		i_c argType,
		flagName string,
	) {
		t.Helper()

		const value = "口水鸡"

		dashes := ""
		switch i_c {
		case argTypeShort:
			dashes += "-"
		case argTypeLong:
			dashes += "--"
		}

		o_c, gotFlagName, gotValue, ok := splitOsArgument(dashes + flagName + "=" + value)
		require.True(t, ok)
		require.Equal(t, i_c, o_c)
		require.Equal(t, flagName, gotFlagName)
		require.Equal(t, argWord(value), gotValue.Unwrap())
	}

	f(argTypeShort, "f")
	f(argTypeLong, "foo")
	f(argTypeLong, "口水鸡")
}

func TestDeYoda(t *testing.T) {
	parser := Pair(Short('a').Switch(), Short('b').Switch()).Parser().ToOptions()
	assertSuccess(t, PPair(false, false), parser)
	assertSuccess(t, PPair(true, true), parser, "-a", "-b")
}

func TestConstructWithFn(t *testing.T) {
	type Opts struct {
		a, b, c bool
	}

	a := Short('a').Switch().Parser()
	b := Short('b').Switch().Parser()
	c := Short('c').Switch().Parser()
	parser := Map3(a, b, c, func(a bool, b bool, c bool) Opts { return Opts{a, b, c} }).ToOptions()

	_, err := parser.RunInner(newArgList("--help"))
	help := UnwrapStdout(err)
	expected_help := `Usage: [-a] [-b] [-c]

Available options:
    -a
    -b
    -c
    -h, --help  Prints help information
`
	require.Equal(t, expected_help, help)

	x, err := parser.RunInner(newArgList("-b", "-c"))
	RequireNoParseFailure(t, err)
	require.Equal(t, Opts{false, true, true}, x)
}

func TestSimpleTwoOptionalFlags(t *testing.T) {
	a := Short('a').Long("AAAAA").Switch()
	b := Short('b').Switch()
	x := Pair(a, b)
	decorated := x.Parser().ToOptions().Description("this is a test")

	t.Run("no version information given - no version field generated", func(t *testing.T) {
		assertFailure(t, decorated, "`-V` is not expected in this context", "-a", "-V")
	})

	t.Run("accept only one copy of -a", func(t *testing.T) {
		assertFailure(t, decorated, "argument `-a` cannot be used multiple times in this context", "-a", "-a")
	})

	t.Run("help", func(t *testing.T) {
		const expected_help = `this is a test

Usage: [-a] [-b]

Available options:
    -a, --AAAAA
    -b
    -h, --help   Prints help information
`
		assertHelp(t, decorated, expected_help, "-h")
	})
}

func TestSimpleTwoOptionalFlagsWithOneHidden(t *testing.T) {
	a := Short('a').Long("AAAAA").Switch()
	b := Short('b').Switch().Parser().Hide()
	decorated := Pair(a, b).Parser().ToOptions().Description("this is a test")

	// no version information given - no version field generated
	assertFailure(t, decorated, "`-V` is not expected in this context", "-a", "-V")

	// accepts only one copy of -a
	assertFailure(t, decorated, "argument `-a` cannot be used multiple times in this context", "-a", "-a")

	{
		const expected_help = `this is a test

Usage: [-a]

Available options:
    -a, --AAAAA
    -h, --help   Prints help information
`
		assertHelp(t, decorated, expected_help, "-h")
	}
}

func TestEitherOfThreeRequiredFlags(t *testing.T) {
	a := ReqFlag(Short('a'), struct{}{})
	b := ReqFlag(Short('b'), struct{}{})
	c := ReqFlag(Short('c'), struct{}{})
	p := a.Parser().OrElse(b.Parser()).Parser().OrElse(c.Parser()).Parser()
	decorated := p.ToOptions().Version("1.0")

	// version help requires version meta
	assertHelp(t, decorated, "Version: 1.0\n", "-V")

	// help is always generated
	{
		const expected_help = `Usage: (-a | -b | -c)

Available options:
    -a
    -b
    -c
    -h, --help     Prints help information
    -V, --version  Prints version information
`
		assertHelp(t, decorated, expected_help, "-h")
	}

	// must specify one of the required flags
	assertFailure(t, decorated, "expected `-a`, `-b`, or more, pass `--help` for usage information")
}

func TestEitherOfThreeRequiredFlags2(t *testing.T) {
	a := ReqFlag(Short('a'), struct{}{}).Parser()
	b := ReqFlag(Short('b'), struct{}{}).Parser()
	c := ReqFlag(Short('c'), struct{}{}).Parser()
	p := Choice(a, b, c)
	decorated := p.ToOptions().Version("1.0")

	assertHelp(t, decorated, "Version: 1.0\n", "-V")

	// help is always generated
	const expected_help = `Usage: (-a | -b | -c)

Available options:
    -a
    -b
    -c
    -h, --help     Prints help information
    -V, --version  Prints version information
`
	assertHelp(t, decorated, expected_help, "-h")

	// must specify one of the required flags
	assertFailure(t, decorated, "expected `-a`, `-b`, or more, pass `--help` for usage information")
}

func TestEitherOfTwoRequiredFlagsAndOneOptional(t *testing.T) {
	a := ReqFlag(Short('a'), true).Parser()
	b := ReqFlag(Short('b'), false).Parser()
	c := Short('c').Switch()
	p := a.OrElse(b).Parser().OrElse(c.Parser()).Parser()
	decorated := p.ToOptions().Version("1.0")

	assertHelp(t, decorated, "Version: 1.0\n", "-V")

	// help is always generated
	const expected_help = `Usage: (-a | -b | [-c])

Available options:
    -a
    -b
    -c
    -h, --help     Prints help information
    -V, --version  Prints version information
`
	assertHelp(t, decorated, expected_help, "-h")

	// fallback to default
	res, err := decorated.RunInner(newArgList())
	require.Nil(t, err)
	require.True(t, !res)
}

func TestFallbackWithOk(t *testing.T) {
	parser := Argument(Short('a'),
		"ARG", parseUint32).Parser().
		FallbackWith(func() (uint32, error) { return 10, nil }).Parser().
		ToOptions()

	assertSuccess(t, uint32(1), parser, "-a", "1")

	assertSuccess(t, uint32(10), parser)
}

func TestFallbackWithErr(t *testing.T) {
	parser := Argument(Short('a'),
		"ARG", parseUint32).Parser().
		FallbackWith(func() (uint32, error) { return 0, errors.New("nope") }).Parser().
		ToOptions()

	assertSuccess(t, uint32(1), parser, "-a", "1")

	assertFailure(t, parser, "couldn't parse `x`: invalid digit found in string", "-a", "x")

	assertFailure(t, parser, "nope")
}

func TestDefaultArguments(t *testing.T) {
	a := Argument(Short('a'), "ARG", parseInt32).Parser().Fallback(42).Parser()
	decorated := a.ToOptions()

	{
		const expected_help = `Usage: [-a=ARG]

Available options:
    -a=ARG
    -h, --help  Prints help information
`
		assertHelp(t, decorated, expected_help, "-h")
	}

	{
		const expected_err = "couldn't parse `x12`: invalid digit found in string"
		assertFailure(t, decorated, expected_err, "-a", "x12")
	}

	{
		const expected_err = "`-a` requires an argument `ARG`"
		assertFailure(t, decorated, expected_err, "-a")
	}
}

func TestParseErrors(t *testing.T) {
	decorated := Argument(Short('a'), "ARG", parseInt32).Parser().ToOptions()

	{
		const expected_err = "couldn't parse `123x`: invalid digit found in string"
		assertFailure(t, decorated, expected_err, "-a", "123x")
	}

	{
		const expected_err = "expected `-a=ARG`, got `-b`. Pass `--help` for usage information"
		assertFailure(t, decorated, expected_err, "-b", "123x")
	}

	{
		const expected_err = "`-b` is not expected in this context"
		assertFailure(t, decorated, expected_err, "-a", "123", "-b")
	}
}

func TestCustomUsage(t *testing.T) {
	a := Argument(Short('a').Long("long"), "ARG", parseString).Parser()
	parser := a.ToOptions().Usage(NewDoc("Usage: -a <ARG> or --long <ARG>"))
	const expected_help = `Usage: -a <ARG> or --long <ARG>

Available options:
    -a, --long=ARG
    -h, --help      Prints help information
`
	assertHelp(t, parser, expected_help)
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

	assertHelp(t, parser, expected_help)
	assertFailure(t, parser, "`-a` requires an argument `ARG`, got a flag `-b`, try `-a=-b` to use it as an argument", "-a", "-b")
}

func TestGroupHelpArgs(t *testing.T) {
	a := Short('a').Help("flag A, related to B").Switch()
	b := Short('b').Help("flag B, related to A").Switch()
	c := Short('c').Help("flag C, unrelated").Switch()
	ab := Pair(a, b).Parser().GroupHelp("Explanation applicable for both A and B:")
	parser := Pair(ab, c).Parser().ToOptions()

	const expected_help = `Usage: [-a] [-b] [-c]

Explanation applicable for both A and B:
    -a          flag A, related to B
    -b          flag B, related to A

Available options:
    -c          flag C, unrelated
    -h, --help  Prints help information
`
	assertHelp(t, parser, expected_help)
}

func TestGroupHelpCommands(t *testing.T) {
	a := Short('a').
		Switch().Parser().
		ToOptions().
		Command("cmd_a").
		Help("command that does A").Parser()
	b := Short('a').
		Switch().Parser().
		ToOptions().
		Command("cmd_b").
		Help("command that does B").Parser()
	c := Short('a').
		Switch().Parser().
		ToOptions().
		Command("cmd_c").
		Help("command that does C").Parser()
	parser := Choice(Choice(a, b).GroupHelp("Explanation applicable for both A and B:").Parser(), c).Parser().ToOptions()

	const expected_help = `Usage: COMMAND ...

Explanation applicable for both A and B:
    cmd_a       command that does A
    cmd_b       command that does B

Available options:
    -h, --help  Prints help information

Available commands:
    cmd_c       command that does C
`
	assertHelp(t, parser, expected_help)
}

func TestFromSeveralAlternativesPickMoreMeaningful(t *testing.T) {
	a := ReqFlag(Short('a'), struct{}{}).Parser()
	b := ReqFlag(Short('b'), struct{}{}).Parser()
	c := ReqFlag(Short('c'), struct{}{}).Parser()
	parser := Choice(a, b, c).ToOptions()

	assertFailure(t, parser, "`-b` cannot be used at the same time as `-a`", "-a", "-b")

	assertFailure(t, parser, "`-a` cannot be used at the same time as `-b`", "-b", "-a")

	assertFailure(t, parser, "`-a` cannot be used at the same time as `-c`", "-c", "-a")

	assertFailure(t, parser, "`-c` cannot be used at the same time as `-a`", "-a", "-c")

	assertFailure(t, parser, "`-b` cannot be used at the same time as `-c`", "-c", "-b", "-a")
}

func TestSubcommands(t *testing.T) {
	bar := Short('b').Switch()
	bar_cmd := bar.Parser().ToOptions().Description("This is local info").Command("bar")
	parser := bar_cmd.Parser().ToOptions().Description("This is global info")

	{
		const expected_help = `This is local info

Usage: bar [-b]

Available options:
    -b
    -h, --help  Prints help information
`
		assertHelp(t, parser, expected_help, "bar", "--help")
	}

	{
		const expected_help = `This is global info

Usage: COMMAND ...

Available options:
    -h, --help  Prints help information

Available commands:
    bar         This is local info
`
		assertHelp(t, parser, expected_help)
	}
}

func TestMultipleAliases(t *testing.T) {
	a := ReqFlag(Short('a').Short('b').Short('c'), struct{}{}).Parser()
	parser := a.ToOptions()

	const expected_help = `Usage: -a

Available options:
    -a
    -h, --help  Prints help information
`
	assertHelp(t, parser, expected_help)

	{
		_, err := parser.RunInner(newArgList("-a"))
		require.Nil(t, err)
	}
	{
		_, err := parser.RunInner(newArgList("-b"))
		require.Nil(t, err)
	}
	{
		_, err := parser.RunInner(newArgList("-c"))
		require.Nil(t, err)
	}
}

func TestGit(t *testing.T) {
	type Opt func(
		Fetch func(
			dry_run bool,
			all bool,
			repository string,
		),
		Add func(
			interactive bool,
			all bool,
			files []string,
		),
	)
	OptFetch := func(dry_run bool, all bool, repository string) Opt {
		return func(
			Fetch func(
				dry_run bool,
				all bool,
				repository string,
			),
			Add func(
				interactive bool,
				all bool,
				files []string,
			),
		) {
			Fetch(dry_run, all, repository)
		}
	}
	OptAdd := func(interactive bool, all bool, files []string) Opt {
		return func(
			Fetch func(
				dry_run bool,
				all bool,
				repository string,
			),
			Add func(
				interactive bool,
				all bool,
				files []string,
			),
		) {
			Add(interactive, all, files)
		}
	}

	parser := func() *OptionParser[Opt] {
		fetch := Map3(
			Long("dry_run").Switch().Parser(),
			Long("all").Switch().Parser(),
			Positional("SRC", parseString).Parser().Fallback("origin").Parser(),
			OptFetch)
		fetch_inner := fetch.
			ToOptions().
			Description("fetches branches from remote repository")
		fetch_cmd := fetch_inner.Command("fetch")

		add := Map3(
			Short('i').Switch().Parser(),
			Long("all").Switch().Parser(),
			Many(Positional("FILE", parseString)).Parser(),
			OptAdd)
		add_inner := add.ToOptions().Description("add files to the staging area")
		add_cmd := add_inner.Command("add")

		return Choice(fetch_cmd, add_cmd).Parser().
			ToOptions().
			Description("The stupid content tracker")
	}()

	t.Run("no_command", func(t *testing.T) {
		const expected_err = "expected `COMMAND ...`, pass `--help` for usage information"
		assertFailure(t, parser, expected_err)
	})

	t.Run("root_help", func(t *testing.T) {
		const expected_help = `The stupid content tracker

Usage: COMMAND ...

Available options:
    -h, --help  Prints help information

Available commands:
    fetch       fetches branches from remote repository
    add         add files to the staging area
`

		assertHelp(t, parser, expected_help)
	})

	t.Run("fetch_help", func(t *testing.T) {
		const expected_help = `fetches branches from remote repository

Usage: fetch [--dry_run] [--all] [SRC]

Available options:
    --dry_run
    --all
    -h, --help  Prints help information
`
		assertHelp(t, parser, expected_help, "fetch", "--help")
	})

	t.Run("add_help", func(t *testing.T) {
		expected_help := `add files to the staging area

Usage: add [-i] [--all] [FILE]...

Available options:
    -i
    --all
    -h, --help  Prints help information
`
		assertHelp(t, parser, expected_help, "add", "--help")
	})
}

func TestArg_bench(t *testing.T) {
	type AppArgs struct {
		number     uint32
		opt_number Opt[uint32]
		width      uint32
		input      []string
	}

	number := Argument[uint32](Long("number").
		Help("Sets a number\nin two lines"),
		"number", parseUint32)

	opt_number := Optional(Argument[uint32](Long("opt-number").
		Help("Sets an optional number"),
		"opt-number", parseUint32).Parser())

	width := Argument[uint32](Long("width").
		Help("Sets width"),
		"width", parseUint32).Parser().
		Guard(func(n uint32) bool { return n > 0 }, "Width must be positive").Parser().
		Fallback(10)

	input := Many(Positional("INPUT", parseString))

	parser := Map4(
		number.Parser(),
		opt_number.Parser(),
		width.Parser(),
		input.Parser(),
		func(number uint32, opt_number Opt[uint32], width uint32, input []string) AppArgs {
			return AppArgs{
				number,
				opt_number,
				width,
				input,
			}
		},
	).ToOptions()

	assertSuccess(t, AppArgs{
		number:     42,
		opt_number: Invalid[uint32](),
		width:      10,
		input:      []string{"foo", "foo2"},
	}, parser, "--number", "42", "foo", "foo2")
	assertSuccess(t, AppArgs{
		number:     42,
		opt_number: Invalid[uint32](),
		width:      10,
		input:      []string{},
	}, parser, "--number", "42")
}

func TestSimpleCargoHelper(t *testing.T) {
	a := Short('a').Long("AAAAA").Help("two lines\nof help").Switch()
	b := Short('b').Switch()
	parser := Pair(a, b).Parser()
	decorated := cargo_helper("simple", parser).
		ToOptions().
		Description("this is a test")

	// cargo run variant
	assertSuccess(t, PPair(true, false), decorated, "-a")

	// cargo simple variant
	assertSuccess(t, PPair(false, true), decorated, "simple", "-b")

	assertFailure(t, decorated, "argument `-a` cannot be used multiple times in this context", "-a", "-a")

	{
		const expected_help = `this is a test

Usage: [-a] [-b]

Available options:
    -a, --AAAAA  two lines of help
    -b
    -h, --help   Prints help information
`
		assertHelp(t, decorated, expected_help, "-h")
	}
}

// func TestLongPathInConstruct(t*testing.T) {
//     a := Short('a').Switch();
//     _ = construct(std::option::Opt::Some(a));

//     b := Short('b').Switch();
//     _ = construct(::std::option::Opt::Some(b));
// }

func TestHiddenEnv(t *testing.T) {
	name := "BPAF_SECRET_API_KEY2"
	visible := Argument(Long("key").
		Help("use this secret key\n two lines"),
		"KEY", parseString).Parser()
	hidden := Argument(Env(name), "KEY", parseString).Parser()
	parser := Choice(visible, hidden).ToOptions()

	const expected = `Usage: --key=KEY

Available options:
    --key=KEY   use this secret key
                two lines
    -h, --help  Prints help information
`
	assertHelp(t, parser, expected, "-h")

	assertFailure(t, parser, "environment variable `BPAF_SECRET_API_KEY2` is not set")
}

func TestEnvVariable(t *testing.T) {
	const name = "BPAF_SECRET_API_KEY"
	parser := Argument(Long("key").
		Env(name).
		Help("use this secret key\ntwo lines"), "KEY", parseString).
		Parser().
		ToOptions()

	{
		const expected_help = `Usage: --key=KEY

Available options:
    --key=KEY   use this secret key two lines
                [env:BPAF_SECRET_API_KEY: N/A]
    -h, --help  Prints help information
`
		assertHelp(t, parser, expected_help, "-h")
	}
	t.Setenv(name, "top s3cr3t")

	{
		const expected_help = `Usage: --key=KEY

Available options:
    --key=KEY   use this secret key two lines
                [env:BPAF_SECRET_API_KEY = "top s3cr3t"]
    -h, --help  Prints help information
`
		assertHelp(t, parser, expected_help, "-h")
	}

	assertSuccess(t, "secret", parser, "--key", "secret")
	assertSuccess(t, "top s3cr3t", parser)
}

func TestDefaultPlaysNicelyWithCommand(t *testing.T) {
	type Foo bool
	const (
		FooFoo Foo = true
		FooBar Foo = false
	)

	cmd := Pure(FooFoo).Parser().
		ToOptions().
		Description("inner").
		Command("foo").
		Help("foo").Parser().
		Fallback(*new(Foo)).Parser()

	parser := cmd.ToOptions().Description("outer")

	{
		const expected_help = `inner

Usage: foo

Available options:
    -h, --help  Prints help information
`
		assertHelp(t, parser, expected_help, "foo", "--help")
	}

	{
		const expected_help = `outer

Usage: [COMMAND ...]

Available options:
    -h, --help  Prints help information

Available commands:
    foo         foo
`
		assertHelp(t, parser, expected_help)
	}
}

func TestCommandWithAliases(t *testing.T) {
	inner := Pure(struct{}{}).Parser().ToOptions().Description("inner descr")
	cmd := inner.Command("foo").Long("bar").Short('f').Short('b')
	parser := cmd.Parser().ToOptions().Description("outer")

	{
		const expected_help = `outer

Usage: COMMAND ...

Available options:
    -h, --help  Prints help information

Available commands:
    foo, f      inner descr
`
		assertHelp(t, parser, expected_help)
	}

	{
		const expected_help = `inner descr

Usage: foo

Available options:
    -h, --help  Prints help information
`
		assertHelp(t, parser, expected_help, "f", "--help")
	}

	// hidden and visible aliases are working
	for _, alias := range []string{"foo", "f", "bar", "b"} {
		_, err := parser.RunInner(newArgList(alias))
		require.Nil(t, err)
	}

	// and "k" isn't a thing
	_, err := parser.RunInner(newArgList("k"))
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

	const expected_help = `Usage: [-a] -c=B --bbbbb=CCC

Available options:
    -a           help for a
    -c=B         help for b
                 [env:BbBbB: N/A]
    --bbbbb=CCC  help for ccc
                 [env:ccccCCccc: N/A]
    -h, --help   Prints help information
`

	assertHelp(t, parser, expected_help)
}

func TestHelpForCommands(t *testing.T) {
	d := Pure(struct{}{}).Parser().
		ToOptions().
		Command("thing_d").
		Help("help for d\ntwo lines").Parser()
	e := Pure(struct{}{}).Parser().
		ToOptions().
		Command("thing_e").
		Short('e').
		Help("help for e\ntwo lines").Parser()
	h := Pure(struct{}{}).Parser().ToOptions().Command("thing_h").Parser()
	parser := Choice(d, e, h).ToOptions()

	const expected_help = `Usage: COMMAND ...

Available options:
    -h, --help  Prints help information

Available commands:
    thing_d     help for d two lines
    thing_e, e  help for e two lines
    thing_h
`
	assertHelp(t, parser, expected_help)
}

func TestManyDoesntPanic(t *testing.T) {
	parser := Map(Many(Short('a').Switch()).Parser(), lenOf).ToOptions()
	assertSuccess(t, 3, parser, "-aaa")
}

func TestSomeDoesntPanic(t *testing.T) {
	parser := Map(Some(Short('a').Switch(), ""), lenOf).ToOptions()
	assertSuccess(t, 3, parser, "-aaa")
}

func TestCommandResetsLeftHeadState(t *testing.T) {
	type Foo fun.Option[uint32]

	a := Argument(Short('a'), "A", parseUint32).Parser().Fallback(0).Parser()
	b := ReqFlag(Short('b'), struct{}{}).Parser()

	p1 := Map(a, func(a uint32) Foo { return Foo{a, true} })
	p2 := Map(b, func(struct{}) Foo { return Foo{0, false} })
	cmd := Choice(p1, p2).
		ToOptions().
		Command("cmd").Parser().
		ToOptions()
	assertSuccess(t, Foo{}, cmd, "cmd", "-b")
}

func TestCommand_preserves_custom_failure_message(t *testing.T) {
	const msg = "need more cheese"
	inner := Fail[struct{}](msg).Parser().ToOptions()

	_, err := inner.RunInner(newArgList())
	require.Equal(t, UnwrapStderr(err), "need more cheese")

	outer := inner.Command("feed").Parser().ToOptions()

	_, err = outer.RunInner(newArgList("feed"))
	require.Equal(t, UnwrapStderr(err), "need more cheese")
}

func TestOptional_error_handling(t *testing.T) {
	p := Optional(Argument(Short('p'), "P", parseUint32).Parser()).Parser().ToOptions()

	assertSuccess(t, Invalid[uint32](), p)
	assertSuccess(t, Valid[uint32](3), p, "-p", "3")
	assertFailure(t, p, "couldn't parse `pi`: invalid digit found in string", "-p", "pi")
}

func TestMany_error_handling(t *testing.T) {
	p := Many(Argument(Short('p'), "P", parseUint32)).Parser().ToOptions()

	assertSuccess(t, []uint32{}, p)
	assertSuccess(t, []uint32{3}, p, "-p", "3")
	assertFailure(t, p, "couldn't parse `pi`: invalid digit found in string", "-p", "pi")
}

func TestFailure_is_not_stupid_1(t *testing.T) {
	a := Argument(Short('a'), "A", parseUint32)
	b := Parse(Pure(struct{}{}).Parser(), func(struct{}) (string, error) { return "", errors.New("nope") })
	parser := Pair(a, b).Parser().ToOptions()

	assertFailure(t, parser, "couldn't parse: nope", "-a", "42")
}

func TestFailure_is_not_stupid_2(t *testing.T) {
	a := Argument(Short('a'), "A", parseUint32)
	b := Argument(Short('b'), "B", parseUint32)
	parser := Parse(Pair(a, b).Parser(), func(fun.Pair[uint32, uint32]) (string, error) { return "", errors.New("nope") }).Parser().ToOptions()

	assertFailure(t, parser, "couldn't parse: nope", "-a", "42", "-b", "42")
}

func TestNo_fallback_out_of_command_parser(t *testing.T) {
	alt1 := Positional("NAME", parseString).Parser().ToOptions().Command("cmd")
	alt2 := Pure("")
	parser := Choice(alt1.Parser(), alt2.Parser()).ToOptions()

	assertFailure(t, parser, "expected `NAME`, pass `--help` for usage information", "cmd")

	assertSuccess(t, "a", parser, "cmd", "a")
	assertSuccess(t, "", parser)
}

func TestDid_you_mean_switch(t *testing.T) {
	a := Short('f').Long("flag").Switch()
	b := Short('p').Long("plag").Switch()
	parser := Choice(a.Parser(), b.Parser()).ToOptions()

	assertFailure(t, parser, "no such flag: `--fla`, did you mean `--flag`?", "--fla")

	assertFailure(t, parser, "no such command or positional: `flag`, did you mean `--flag`?", "flag")

	assertFailure(t, parser, "no such flag: `--pla`, did you mean `--plag`?", "--pla")

	assertFailure(t, parser, "no such flag: `--p` (with two dashes), did you mean `-p`?", "--p")
}

func TestDid_you_mean_req_flag(t *testing.T) {
	parser := ReqFlag(Long("flag"), struct{}{}).Parser().ToOptions()
	assertFailure(t, parser, "no such flag: `--fla`, did you mean `--flag`?", "--fla")
}

func TestDid_you_mean_argument(t *testing.T) {
	parser := Argument(Long("flag"), "VAL", parseString).Parser().ToOptions()

	assertFailure(t, parser, "no such flag: `--flg`, did you mean `--flag`?", "--flg=hellop")

	assertFailure(t, parser, "no such flag: `--fla`, did you mean `--flag`?", "--fla")
}

func TestDid_you_mean_command(t *testing.T) {
	parser := Pure(struct{}{}).Parser().
		ToOptions().
		Command("command").
		Short('c').Parser().
		ToOptions()

	assertFailure(t, parser, "no such command or positional: `comman`, did you mean `command`?", "comman")

	assertFailure(t, parser, "no such flag: `--comman`, did you mean `command`?", "--comman")
}

func TestDid_you_mean_two_and_arguments(t *testing.T) {
	a := Long("flag").Switch()
	b := Long("parameter").Switch()
	parser := cargo_helper("cmd", Pair(a, b).Parser()).ToOptions()

	assertFailure(t, parser, "no such flag: `--parametr`, did you mean `--parameter`?", "--flag", "--parametr")

	assertFailure(t, parser, "no such flag: `--paramet`, did you mean `--parameter`?", "--flag", "--paramet=value")

	assertFailure(t, parser, "no such flag: `--flg`, did you mean `--flag`?", "--parameter", "--flg")

	assertFailure(t, parser, "no such flag: `--fla`, did you mean `--flag`?", "--fla")
}

func TestDid_you_mean_two_or_arguments(t *testing.T) {
	a := Long("flag").Switch()
	b := Long("parameter").Switch()
	parser := cargo_helper("cmd", Choice(a, b)).ToOptions()

	assertFailure(t, parser, "no such flag: `--fla`, did you mean `--flag`?", "--fla")

	assertFailure(t, parser, "no such flag: `--parametr`, did you mean `--parameter`?", "--flag", "--parametr")

	assertFailure(t, parser, "no such flag: `--parametr`, did you mean `--parameter`?", "--parametr", "--flag")

	assertFailure(t, parser, "`--flag` cannot be used at the same time as `--parameter`", "--parameter", "--flag")

	assertFailure(t, parser, "`--parameter` cannot be used at the same time as `--flag`", "--flag", "--parameter")
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
	target_dir := Argument(Short('t'), "T", parseString).Parser().Guard(func(string) bool { return false }, "nope")
	verbosity := Short('v').Switch()
	inner := Pair(target_dir, verbosity)
	parser := cargo_helper("asm", inner.Parser()).ToOptions()

	assertFailure(t, parser, "`x`: nope", "asm", "-t", "x")

	assertFailure(t, parser, "`x`: nope", "-t", "x")
}

func TestCargo_show_asm_issue_from_str(t *testing.T) {
	target_dir := Argument(Short('t'), "T", parseUint)
	verbosity := Short('v').Switch()
	inner := Pair(target_dir, verbosity)
	parser := cargo_helper("asm", inner.Parser()).ToOptions()

	assertFailure(t, parser, "couldn't parse `x`: invalid digit found in string", "asm", "-t", "x")

	assertFailure(t, parser, "couldn't parse `x`: invalid digit found in string", "-t", "x")
}

func TestCargo_show_asm_issue_parse(t *testing.T) {
	target_dir := Parse(Argument(Short('t'), "T", parseString).Parser(),
		func(string) (struct{}, error) { return struct{}{}, errors.New("nope") })

	verbosity := Short('v').Switch()
	inner := Pair(target_dir, verbosity)
	parser := cargo_helper("asm", inner.Parser()).ToOptions()

	assertFailure(t, parser, "couldn't parse `x`: nope", "asm", "-t", "x")

	assertFailure(t, parser, "couldn't parse `x`: nope", "-t", "x")
}

// problematic case looks something like this:
// "asm" is consumed
// "--fla" is not consumed, but the output is not tainted
// parser fails and arguments are restored.

func TestCargo_show_asm_issue_unknown_switch(t *testing.T) {
	target := Long("flag").Switch()

	verbosity := Short('v').Switch()
	parser := cargo_helper("asm", Pair(target, verbosity).Parser()).ToOptions()

	assertFailure(t, parser, "no such flag: `--fla`, did you mean `--flag`?", "asm", "--fla")

	assertFailure(t, parser, "no such flag: `--fla`, did you mean `--flag`?", "--fla")
}

func TestDid_you_mean_inside_command(t *testing.T) {
	a := Long("flag").Switch()
	b := Long("parameter").Switch()
	parser := Choice(a, b).ToOptions().Command("cmd").Parser().ToOptions()

	assertFailure(t, parser, "expected `COMMAND ...`, got `--fla`. Pass `--help` for usage information", "--fla")

	assertFailure(t, parser, "no such flag: `--fla`, did you mean `--flag`?", "cmd", "--fla")

	assertFailure(t, parser, "no such flag: `--parametr`, did you mean `--parameter`?", "cmd", "--flag", "--parametr")

	assertFailure(t, parser, "no such flag: `--parametr`, did you mean `--parameter`?", "cmd", "--parametr", "--flag")

	assertFailure(t, parser, "`--flag` cannot be used at the same time as `--parameter`", "cmd", "--parameter", "--flag")

	assertFailure(t, parser, "`--parameter` cannot be used at the same time as `--flag`", "cmd", "--flag", "--parameter")
}

func TestCombine_flags_by_order(t *testing.T) {
	a := ReqFlag(Short('a'), true)
	b := ReqFlag(Short('A'), false)
	parser := Many(Choice(a, b)).Parser().ToOptions()

	assertSuccess(t, []bool{true, false, false, false, true}, parser, "-a", "-A", "-A", "-A", "-a")
}

func TestParse_many_errors_positional(t *testing.T) {
	p := Many(Positional("N", parseUint32)).Parser().ToOptions()

	assertFailure(t, p, "couldn't parse `x`: invalid digit found in string", "1", "2", "x") // TODO: replace `x` with %q // TODO: replace `x` with %q
	assertSuccess(t, []uint32{1, 2, 3}, p, "1", "2", "3")
}

func TestParse_collect_flag(t *testing.T) {
	p := Collect(Argument(Short('p'), "N", parseUint32)).ToOptions()

	assertSuccess(t, []uint32{1, 2}, p, "-p", "1", "-p", "2")
	assertFailure(t, p, "couldn't parse `x`: invalid digit found in string", "-p", "1", "-p", "x")
}

func TestParse_many_errors_flag(t *testing.T) {
	p := Many(Argument(Short('p'), "N", parseUint32)).Parser().ToOptions()

	assertSuccess(t, []uint32{1, 2}, p, "-p", "1", "-p", "2")
	assertFailure(t, p, "couldn't parse `x`: invalid digit found in string", "-p", "1", "-p", "x")
}

func TestCommand_with_req_parameters(t *testing.T) {
	p := Positional("X", parseString).Parser().
		ToOptions().
		Command("cmd").Parser().
		Fallback("").Parser().
		ToOptions()

	assertFailure(t, p, "expected `X`, pass `--help` for usage information", "cmd")
}

func TestSuggestion_for_equals_1(t *testing.T) {
	parser := Argument(Short('p').Long("par"), "P", parseString).Parser().ToOptions()

	assertFailure(t, parser, "`-p` requires an argument `P`, got a flag `--bar`, try `-p=--bar` to use it as an argument", "-p", "--bar")

	assertFailure(t, parser, "`--par` requires an argument `P`, got a flag `--bar`, try `--par=--bar` to use it as an argument", "--par", "--bar")

	assertFailure(t, parser, "`--par` requires an argument `P`, got a flag `--bar=baz`, try `--par=--bar=baz` to use it as an\nargument", "--par", "--bar=baz")
}

func TestDouble_dash_is_pos_only_just_once(t *testing.T) {
	parser := Many(Positional("POS", parseString)).Parser().ToOptions()

	assertSuccess(t, []string{}, parser, "--")
	assertSuccess(t, []string{"--"}, parser, "--", "--")
}

func TestReject_fbar(t *testing.T) {
	parser := Argument(Short('f'), "F", parseString).Parser().ToOptions()

	assertFailure(t, parser, "`baz` is not expected in this context", "-fbar", "baz")
	assertSuccess(t, "bar", parser, "-fbar")
}

func TestCustom_usage_override_fixed(t *testing.T) {
	parser := Short('p').Switch().Parser().ToOptions().Usage(NewDoc("Usage: hey [-p]"))
	assertHelp(t, parser, "Usage: hey [-p]\n\nAvailable options:\n    -p\n    -h, --help  Prints help information\n")
}

func TestCustom_usage_override_with_fn(t *testing.T) {
	parser := Short('p').Switch().Parser().ToOptions().WithUsage(func(b Doc) Doc {
		buf := Doc{}
		buf.Text("Usage: hey ")
		buf.Doc(b)
		return buf
	})
	assertHelp(t, parser, "Usage: hey [-p]\n\nAvailable options:\n    -p\n    -h, --help  Prints help information\n")
}

func TestCatch_works(t *testing.T) {
	type A struct {
		isNum bool
		Num   uint
		Str   string
	}
	ANum := func(a uint) A { return A{true, a, ""} }
	AStr := func(s string) A { return A{false, 0, s} }
	a_n := OptionalCatch(Map(Argument(Short('a'),
		"A", parseUint).Parser(),
		ANum))
	a_s := Optional(Map(Argument(Short('a'),
		"A", parseString).Parser(),
		AStr).
		Hide())
	parser := Choice(a_n.Parser(), a_s.Parser()).ToOptions()

	assertSuccess(t, Valid(AStr("x1")), parser, "-a", "x1")
	assertSuccess(t, Valid(ANum(1)), parser, "-a", "1")
}

func TestSneaky_command(t *testing.T) {
	type Cmd struct {
		isA bool
		x   bool
	}
	CmdA := func(x bool) Cmd { return Cmd{true, x} }
	CmdB := func(x bool) Cmd { return Cmd{false, x} }
	cmd := Map(Short('f').
		Switch().Parser().
		ToOptions().
		Command("hello").Parser(),
		CmdA)
	b := Map(Short('f').Switch().Parser(), CmdB)

	sneaky := true // TODO: ???

	maybe_cmd := cmd
	if !sneaky {
		maybe_cmd = Fail[Cmd]("no sneaky").Parser()
	}
	parser := Choice(maybe_cmd, b).ToOptions()

	assertSuccess(t, CmdA(false), parser, "hello")
}

func TestDefault_for_many(t *testing.T) {
	parser := Map(Many(Positional("ROOTS", parseString)).Parser(),
		func(items []string) []string {
			if len(items) == 0 {
				return []string{"."}
			}
			return items
		}).
		ToOptions()

	assertSuccess(t, []string{"."}, parser)
}

func TestParse_option_catch(t *testing.T) {
	type A struct {
		isU32 bool
		U32   uint32
		S     string
	}
	AU32 := func(a uint32) A { return A{true, a, ""} }
	AS := func(s string) A { return A{false, 0, s} }
	a1 := OptionalCatch(Map(Argument(Short('a'), "N", parseUint32).Parser(), AU32))
	a2 := OptionalCatch(Map(Argument(Short('a'), "S", parseString).Parser(), AS))
	parser := Choice(a1.Parser(), a2.Parser()).ToOptions()

	assertSuccess(t, Valid(AU32(10)), parser, "-a", "10")
	assertSuccess(t, Valid(AS("x")), parser, "-a", "x")
	assertSuccess(t, Invalid[A](), parser)
}

func TestParse_some_catch(t *testing.T) {
	type A struct {
		isU32 bool
		U32   uint32
		S     string
	}
	AU32 := func(a uint32) A { return A{true, a, ""} }
	AS := func(s string) A { return A{false, 0, s} }
	a1 := SomeCatch(Map(Argument(Short('a'),
		"N", parseUint32).Parser(),
		AU32), "A").
		Hide()
	a2 := SomeCatch(Map(Argument(Short('a'), "S", parseString).Parser(), AS), "A").Hide()
	parser := Choice(a1, a2).ToOptions()

	assertSuccess(t, []A{AU32(10)}, parser, "-a", "10")
	assertSuccess(t, []A{AS("x")}, parser, "-a", "x")
	assertFailure(t, parser, "A")
}

func TestEmpty_struct(t *testing.T) {
	type Foo struct{}
	parser := Pure(Foo{}).Parser().ToOptions()

	assertSuccess(t, Foo{}, parser)
	assertHelp(t, parser, `Usage:

Available options:
    -h, --help  Prints help information
`)
}

func TestStrange_short_option(t *testing.T) {
	parser := Argument(Short('O'), "ARG", parseString).Parser().ToOptions()
	assertSuccess(t, "bits=2048", parser, "-Obits=2048")
}

func TestOptional_bool_states(t *testing.T) {
	parser := Optional(Short('a').Switch().Parser()).Parser().ToOptions()
	assertSuccess(t, Valid(true), parser, "-a")
	assertSuccess(t, Valid(false), parser)
}

func TestFancy_negative(t *testing.T) {
	a := ReqFlag(Short('a'), struct{}{})
	b := Any("A", func(s string) (int32, bool) { res, err := strconv.ParseInt(s, 10, 32); return int32(res), err == nil }, false)
	ab := Map(Pair(a, b).Adjacent().Parser(), func(x fun.Pair[struct{}, int32]) int32 { return x.V })

	c := Argument(Short('c'), "C", parseUint).Parser().Fallback(42)

	parser := Pair(ab, c).Parser().ToOptions()

	assertSuccess(t, PPair[int32, uint](-10, 42), parser, "-a", "-10")
	assertSuccess(t, PPair[int32, uint](-20, 110), parser, "-a=-20", "-c", "110")
	{
		// TODO - rendering sucks once you start inventing fancy combinations and don't provide help...
		const expected = `Usage: -a A [-c=C]

Available options:
  -a A

    -c=C
    -h, --help  Prints help information
`
		assertHelp(t, parser, expected)
	}
}

func TestMany_env(t *testing.T) {
	t.Setenv("USER1", "top s3cr3t")
	parser := Many(Argument(Short('v').
		Env("USER1"),
		"USER", parseString)).Parser().
		ToOptions()
	assertSuccess(t, []string{"top s3cr3t"}, parser)
}

func TestEnv_hidden_arg(t *testing.T) {
	t.Setenv("USER1", "top s3cr3t")
	parser := Argument(Env("USER1"), "USER", parseString).Parser().ToOptions()
	assertSuccess(t, "top s3cr3t", parser)
}

func TestEnv_hidden_switch(t *testing.T) {
	t.Setenv("USER1", "top s3cr3t")
	parser := Env("USER1").Switch().Parser().ToOptions()
	assertSuccess(t, true, parser)
}

func TestEnv_hidden_flag(t *testing.T) {
	t.Setenv("USER1", "top s3cr3t")
	parser := Flag(Env("USER1"), true, false).Parser().ToOptions()
	assertSuccess(t, true, parser)
}

func TestSome_env(t *testing.T) {
	t.Setenv("USER1", "top s3cr3t")
	parser := Some(Argument(Short('v').
		Env("USER1"),
		"USER", parseString), "a").
		ToOptions()
	assertSuccess(t, []string{"top s3cr3t"}, parser)
}

func TestOption_requires_other_option1(t *testing.T) {
	a := Short('a').Switch()
	b := Argument(Short('b'), "B", parseString)
	parser := Optional(Pair(a, b).Parser()).Parser().ToOptions()

	assertFailure(t, parser, "expected `-b=B`, pass `--help` for usage information", "-a")
}

func TestOption_requires_other_option2(t *testing.T) {
	a := Short('a').Switch()
	b := Argument(Short('b'), "B", parseString)
	parser := Optional(Pair(b, a).Parser()).Parser().ToOptions()

	assertFailure(t, parser, "expected `-b=B`, pass `--help` for usage information", "-a")
}

func TestDefault_for_some(t *testing.T) {
	parser := Some(Positional[uint32]("ROOTS", parseUint32), "msg").
		FallbackWith(func() ([]uint32, error) { return []uint32{1, 2, 3}, nil }).
		ToOptions()

	assertSuccess(t, []uint32{1, 2, 3}, parser)
}

func TestAdjacent_anywhere_needs_to_consume_something(t *testing.T) {
	a := Short('a').Switch()
	b := Short('b').Switch()
	parser := Pair(a, b).Adjacent().Parser().ToOptions()

	assertSuccess(t, PPair(true, false), parser, "-a")
	assertSuccess(t, PPair(false, true), parser, "-b")
}

func TestFallback_for_some(t *testing.T) {
	parser := Some(Argument(Short('a'),
		"ARG", parseUint32), "potatoes").
		Fallback([]uint32{1, 2, 3}).ToOptions()

	assertSuccess(t, []uint32{1}, parser, "-a", "1")
	assertSuccess(t, []uint32{1, 2, 3}, parser)
}

func TestFlag_like_commands(t *testing.T) {
	a := ReqFlag(Short('a'), 1).Parser().ToOptions().Command("--add")
	b := ReqFlag(Short('b'), 2).Parser().ToOptions().Command("remove")
	parser := Choice(a.Parser(), b.Parser()).ToOptions()

	assertSuccess(t, 1, parser, "--add", "-a")
	assertSuccess(t, 2, parser, "remove", "-b")

	{
		const expected = "Usage: COMMAND ...\n\nAvailable options:\n    -h, --help  Prints help information\n\nAvailable commands:\n    --add\n    remove\n"
		assertHelp(t, parser, expected)
	}

	{
		const expected = "Usage: --add -a\n\nAvailable options:\n    -a\n    -h, --help  Prints help information\n"
		assertHelp(t, parser, expected, "--add", "--help")
	}
}
