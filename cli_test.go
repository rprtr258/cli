package cli

import (
	"testing"

	"github.com/rprtr258/fun"
	"github.com/stretchr/testify/require"
)

// func assertSuccess[T any](t *testing.T, parser FlagParser[T], expected T, args ...string) {
// 	t.Helper()
// 	ctx := newContext(args...)
// 	got, err := parser.run(ctx).unpack()
// 	require.Nil(t, err)
// 	require.Equal(t, expected, got)
// }

// func assertFail[T any](t *testing.T, parser FlagParser[T], expected string, args ...string) {
// 	t.Helper()
// 	ctx := newContext(args...)
// 	_, err := parser.run(ctx).unpack()
// 	require.Equal(t, err, expected)
// }

// var optsSimpleSubcommand = Bool("g")

// func TestRun_first_subcommand(t *testing.T) {
// 	assertSuccess(t, optsSimpleSubcommand, true, "-g")
// }

// type optsFirst struct {
// 	Value   bool                 `short:"v"`
// 	Command optsSimpleSubcommand `command:"cmd"`
// }

// func TestRun_first_command(t *testing.T) {
// 	assertSuccess(t, optsFirst, true, "-v")
// }

// func TestRun_first_commandExtraneousPositionals(t *testing.T) {
// 	assertFail(t, optsFirst,
// 		`extraneous positional arguments found: ["cmd", "-g"]`,
// 		"-v", "cmd", "-g",
// 	)
// }

// func TestRun_extraFlag(t *testing.T) {
// 	assertFail(t, optsFirst,
// 		`extraneous positional arguments found: ["-g", "cmd"]`,
// 		"-v", "-g", "cmd",
// 	)
// }

// func TestNewCommand_CommandInline(t *testing.T) {
// 	cmd, err := newCommand[optsFirst]()
// 	require.NoError(t, err)

// 	// zero out function pointers to be able to compare the rest
// 	for i := range cmd.flags {
// 		cmd.flags[i].valuesetter = nil
// 	}
// 	cmd.posArgsRest.Value.valuesetter = nil
// 	cmd.execute = nil
// 	for i, subcmd := range cmd.subcommands {
// 		for i := range subcmd.flags {
// 			subcmd.flags[i].valuesetter = nil
// 		}
// 		cmd.subcommands[i].posArgsRest.Value.valuesetter = nil
// 		subcmd.execute = nil
// 	}

// 	want := &command{
// 		subcommands: map[string]*command{
// 			"cmd": {
// 				subcommands: map[string]*command{},
// 				flags: []flag{
// 					{
// 						shortname:    'g',
// 						longname:     "",
// 						description:  Option[string]{},
// 						defaultValue: nil,
// 						isbool:       true,
// 						minCount:     0,
// 						maxCount:     1,
// 					},
// 				},
// 				posArgs:     []positionalArg{},
// 				posArgsRest: Option[positionalArgsRest]{},
// 			},
// 		},
// 		flags: []flag{
// 			{
// 				shortname:    'v',
// 				longname:     "",
// 				description:  Option[string]{},
// 				defaultValue: nil,
// 				isbool:       true,
// 				minCount:     0,
// 				maxCount:     1,
// 			},
// 		},
// 		posArgs:     []positionalArg{},
// 		posArgsRest: Option[positionalArgsRest]{},
// 	}
// 	require.Equal(t, want, cmd)
// }

// func TestRun_leafHasNoExecute(t *testing.T) {
// 	type optsSecond struct {
// 		C1 struct{} `command:"c1"`
// 	}

// 	assertFail[optsSecond](t,
// 		`parse subcommand c1: command struct {} is leaf command, but no Execute method defined`,
// 		"-v",
// 	)
// }

// func TestRun_subcommand_extraArgs(t *testing.T) {
// 	assertFail[optsFirst](t,
// 		`extraneous positional arguments found: ["-v", "-g"]`,
// 		"cmd", "-v", "-g",
// 	)
// }

// func TestRun_subcommand_extraArg(t *testing.T) {
// 	assertFail[optsFirst](t,
// 		`extraneous positional arguments found: ["-v"]`,
// 		"cmd", "-g", "-v",
// 	)
// }

// type optsThirdCmdSub struct {
// 	B bool `short:"b"`
// }

// func (x optsThirdCmdSub) Execute(ctx context.Context) error { return fromCtx(ctx)(x) }

// type optsThirdCmd struct {
// 	G          bool            `short:"g"`
// 	SubCommand optsThirdCmdSub `command:"sub"`
// }

// func (x optsThirdCmd) Execute(ctx context.Context) error { return fromCtx(ctx)(x) }

// type optsThird struct {
// 	Value   bool         `short:"v"`
// 	Command optsThirdCmd `command:"cmd"`
// }

// func (x optsThird) Execute(ctx context.Context) error { return fromCtx(ctx)(x) }

// func Test_third_subsubcommandExtraArgs(t *testing.T) {
// 	assertFail[optsThird](t,
// 		`extraneous positional arguments found: ["-v", "-g"]`,
// 		"cmd", "sub", "-b", "-v", "-g",
// 	)
// }

// type optsFourthSub struct {
// 	Value bool `short:"v"`
// }

// func (x optsFourthSub) Execute(ctx context.Context) error { return fromCtx(ctx)(x) }

// type optsFourthCmd struct {
// 	Value   bool          `short:"v"`
// 	Command optsFourthSub `command:"sub"`
// }

// func (x optsFourthCmd) Execute(ctx context.Context) error { return fromCtx(ctx)(x) }

// func Test_fourth_flagThenPresumablySubcommand(t *testing.T) {
// 	assertFail[optsFourthCmd](t,
// 		`extraneous positional arguments found: ["sub"]`,
// 		"-v", "sub",
// 	)
// }

// func Test_fourth_flag(t *testing.T) {
// 	assertSuccess[optsFourthCmd](t,
// 		func(x optsFourthCmd) {
// 			require.True(t, x.Value)
// 		},
// 		"-v",
// 	)
// }

// type optsFourth struct {
// 	Value   bool          `short:"v"`
// 	Command optsFourthCmd `command:"cmd"`
// }

// func (x optsFourth) Execute(ctx context.Context) error { return fromCtx(ctx)(x) }

// func Test_fourth_altSubcommandFlag(t *testing.T) {
// 	assertSuccess[optsFourth](t,
// 		func(x optsFourthSub) {
// 			require.True(t, x.Value)
// 		},
// 		"cmd", "sub", "-v",
// 	)
// }

// func TestCommandEstimate(t *testing.T) {
// 	type optsSix struct {
// 		Cmd1 optsSimpleSubcommand `command:"remove"`
// 		Cmd2 optsSimpleSubcommand `command:"add"`
// 	}

// 	assertFail[optsSix](t,
// 		"Please specify one command of: add or remove\nUsage: \x1b[94mx\x1b[0m \x1b[91m[COMMAND]\x1b[0m \x1b[93m[FLAGS]\x1b[0m\n\n\n\nCommands:\n\t\x1b[91madd\x1b[0m      \n\t\x1b[91mremove\x1b[0m   \n\nFlags:\n\t\x1b[93m-h\x1b[0m, \x1b[93m--help\x1b[0m    Print this help message\n",
// 	)
// }

// func TestCommandEstimate2(t *testing.T) {
// 	type optsSix struct {
// 		Cmd1 optsSimpleSubcommand `command:"remove"`
// 		Cmd2 optsSimpleSubcommand `command:"add"`
// 	}

// 	assertFail[optsSix](t,
// 		`Unknown command "rmive", did you mean remove?`+"\nUsage: \x1b[94mx\x1b[0m \x1b[91m[COMMAND]\x1b[0m \x1b[93m[FLAGS]\x1b[0m\n\n\n\nCommands:\n\t\x1b[91madd\x1b[0m      \n\t\x1b[91mremove\x1b[0m   \n\nFlags:\n\t\x1b[93m-h\x1b[0m, \x1b[93m--help\x1b[0m    Print this help message\n",
// 		"rmive",
// 	)
// }

// type optsSeven struct {
// 	G    bool     `short:"g"`
// 	Args []string `positional-args:"yes"`
// }

// func (x optsSeven) Execute(ctx context.Context) error { return fromCtx(ctx)(x) }

// func Test_subcommands_posArgs(t *testing.T) {
// 	type Opts struct {
// 		Command optsSeven `command:"cmd"`
// 	}

// 	assertSuccess[Opts](t,
// 		func(x optsSeven) {
// 			require.True(t, x.G)
// 			require.Equal(t, []string{"a", "b"}, x.Args)
// 		},
// 		"cmd", "-g", "a", "b",
// 	)
// }

// type optsParse struct {
// 	Verbose bool     `short:"v" long:"verbose" description:"Show verbose debug information"`
// 	Name    string   `short:"n" long:"name" description:"A name"`
// 	Args    []string `positional-args:"yes"`
// }

// func (x optsParse) Execute(ctx context.Context) error { return fromCtx(ctx)(x) }

// func TestParse(t *testing.T) {
// 	assertSuccess[optsParse](t,
// 		func(x optsParse) {
// 			require.Equal(t, optsParse{
// 				Verbose: true,
// 				Name:    "aboba",
// 				Args:    []string{"xdd"},
// 			}, x)
// 		},
// 		"--name", "aboba", "--verbose", "xdd",
// 	)
// }

// type cmdTestGroupInline struct {
// 	Value bool `short:"v"`
// 	G     bool `short:"g"`
// }

// func (x cmdTestGroupInline) Execute(ctx context.Context) error { return fromCtx(ctx)(x) }

// func TestGroupInline(t *testing.T) {
// 	assertSuccess[cmdTestGroupInline](t, func(x cmdTestGroupInline) {
// 		require.True(t, x.Value)
// 		require.True(t, x.G)
// 	}, "-v", "-g")
// }

// type cmdTestGroupAdd struct {
// 	Value bool     `short:"v"`
// 	G     bool     `short:"g"`
// 	Args  []string `positional-args:"yes"`
// }

// func (x cmdTestGroupAdd) Execute(ctx context.Context) error { return fromCtx(ctx)(x) }

// func TestGroupAdd(t *testing.T) {
// 	assertSuccess[cmdTestGroupAdd](t, func(x cmdTestGroupAdd) {
// 		require.Equal(t, []string{"rest"}, x.Args)
// 		require.True(t, x.Value)
// 		require.True(t, x.G)
// 	}, "-v", "-g", "rest")
// }

// type cmdTestGroupNestedInline struct {
// 	Value bool     `short:"v"`
// 	G     bool     `short:"g"`
// 	Nah   string   `long:"nah"`
// 	Args  []string `positional-args:"yes"`
// }

// func (x cmdTestGroupNestedInline) Execute(ctx context.Context) error { return fromCtx(ctx)(x) }

// func TestGroupNestedInline(t *testing.T) {
// 	assertSuccess[cmdTestGroupNestedInline](t, func(x cmdTestGroupNestedInline) {
// 		require.Equal(t, []string{"rest"}, x.Args)
// 		require.True(t, x.Value)
// 		require.True(t, x.G)
// 		require.Equal(t, x.Nah, "n")
// 	}, "-v", "-g", "--nah", "n", "rest")
// }

// type cmdTestNestedInlineNamespace struct {
// 	Opt       string   `long:"opt"`
// 	SipOpt    string   `long:"sip.opt"`
// 	SipSapOpt string   `long:"sip.sap.opt"`
// 	Args      []string `positional-args:"yes"`
// }

// func (x cmdTestNestedInlineNamespace) Execute(ctx context.Context) error { return fromCtx(ctx)(x) }

// func TestGroupNestedInlineNamespace(t *testing.T) {
// 	assertSuccess[cmdTestNestedInlineNamespace](t, func(x cmdTestNestedInlineNamespace) {
// 		require.Equal(t, []string{"rest"}, x.Args)
// 		require.Equal(t, x.Opt, "a")
// 		require.Equal(t, x.SipOpt, "b")
// 		require.Equal(t, x.SipSapOpt, "c")
// 	}, "--opt", "a", "--sip.opt", "b", "--sip.sap.opt", "c", "rest")
// }

// type cmdBoolFlag struct {
// 	Test bool `long:"test"`
// }

// func (x cmdBoolFlag) Execute(ctx context.Context) error { return fromCtx(ctx)(x) }

// func TestBoolFlag(t *testing.T) {
// 	assertSuccess[cmdBoolFlag](t, func(x cmdBoolFlag) {
// 		require.False(t, x.Test)
// 	})
// 	assertSuccess[cmdBoolFlag](t, func(x cmdBoolFlag) {
// 		require.True(t, x.Test)
// 	}, "--test")
// }

// type cmdStringFlag struct {
// 	Test string `long:"value"`
// }

// func (x cmdStringFlag) Execute(ctx context.Context) error { return fromCtx(ctx)(x) }

// func TestStringFlag(t *testing.T) {
// 	assertSuccess[cmdStringFlag](t, func(x cmdStringFlag) {
// 		require.Equal(t, "value", x.Test)
// 	}, "--value", "value")
// }

// type cmdStringDefaultFlag struct {
// 	Test string `long:"value" default:"uuh"`
// }

// func (x cmdStringDefaultFlag) Execute(ctx context.Context) error { return fromCtx(ctx)(x) }

// func TestStringDefaultFlagValueProvided(t *testing.T) {
// 	assertSuccess[cmdStringDefaultFlag](t, func(x cmdStringDefaultFlag) {
// 		require.Equal(t, "value", x.Test)
// 	}, "--value", "value")
// }

// func TestStringDefaultFlagValueNotProvided(t *testing.T) {
// 	assertSuccess[cmdStringDefaultFlag](t, func(x cmdStringDefaultFlag) {
// 		require.Equal(t, "uuh", x.Test)
// 	})
// }

// type cmdStringOptionalFlag struct {
// 	Test *string `long:"value"`
// }

// func (x cmdStringOptionalFlag) Execute(ctx context.Context) error { return fromCtx(ctx)(x) }

// func TestStringOptionalFlagValueProvided(t *testing.T) {
// 	assertSuccess[cmdStringOptionalFlag](t, func(x cmdStringOptionalFlag) {
// 		require.Equal(t, "value", *x.Test)
// 	}, "--value", "value")
// }

// func TestStringOptionalFlagValueNotProvided(t *testing.T) {
// 	assertSuccess[cmdStringOptionalFlag](t, func(x cmdStringOptionalFlag) {
// 		require.Nil(t, x.Test)
// 	})
// }

// func TestUnknown(t *testing.T) {
// 	assertFail[cmdStringOptionalFlag](t,
// 		`extraneous positional arguments found: ["-f"]`,
// 		"-f")
// }

// type cmdFlagsInPositionals struct {
// 	Verbose bool     `short:"v"`
// 	Args    []string `positional-args:"yes"`
// }

// func (x cmdFlagsInPositionals) Execute(ctx context.Context) error { return fromCtx(ctx)(x) }

// func TestFlagsInPositionals(t *testing.T) {
// 	assertSuccess[cmdFlagsInPositionals](t, func(x cmdFlagsInPositionals) {
// 		require.False(t, x.Verbose)
// 		require.Equal(t, []string{"hello", "world", "-v", "--foo", "bar", "--verbose", "-f"}, x.Args)
// 	}, "hello", "world", "-v", "--foo", "bar", "--verbose", "-f")
// }

// type cmdPositionals struct {
// 	Verbose    bool `short:"v"`
// 	Positional struct {
// 		Command  int
// 		Filename string
// 		Rest     []string
// 	} `positional-args:"yes"`
// }

// func (x cmdPositionals) Execute(ctx context.Context) error { return fromCtx(ctx)(x) }

// func TestPositional(t *testing.T) {
// 	assertSuccess[cmdPositionals](t, func(x cmdPositionals) {
// 		require.Equal(t, 10, x.Positional.Command)
// 		require.Equal(t, "arg_test.go", x.Positional.Filename)
// 		require.Equal(t, []string{"a", "b"}, x.Positional.Rest)
// 	}, "10", "arg_test.go", "a", "b")
// }

// func TestPositionalMissing(t *testing.T) {
// 	assertFail[cmdPositionals](t,
// 		`positional arguments are missing: Filename`,
// 		"10")
// }

// func TestPositionalRequiredRest1Fail(t *testing.T) {
// 	t.Skip()

// 	type cmd struct {
// 		Value      bool `short:"v"`
// 		Positional struct {
// 			// TODO: test using custom field for rest, that checks that at least 1 arg is provided
// 			Rest []string
// 		} `positional-args:"yes"`
// 	}

// 	assertFail[cmd](t,
// 		`the required argument "Rest (at least 1 argument)" was not provided`,
// 	)
// }

// type cmdRest1 struct {
// 	Verbose    bool `short:"v"`
// 	Positional struct {
// 		Rest []string
// 	} `positional-args:"yes"`
// }

// func (x cmdRest1) Execute(ctx context.Context) error { return fromCtx(ctx)(x) }

// func TestRest1(t *testing.T) {
// 	assertSuccess[cmdRest1](t, func(x cmdRest1) {
// 		require.Equal(t, []string{"rest1"}, x.Positional.Rest)
// 		require.False(t, x.Verbose)
// 	}, "rest1")
// }

// func TestPositionalRequiredRest2Fail(t *testing.T) {
// 	t.Skip()

// 	type cmd struct {
// 		Value      bool `short:"v"`
// 		Positional struct {
// 			// TODO: test using custom field for rest, that checks that at least 2 args are provided
// 			Rest []string
// 		} `positional-args:"yes"`
// 	}

// 	assertFail[cmd](t,
// 		`the required argument "Rest (at least 2 arguments, but got only 1)" was not provided`,
// 		"rest1")
// 	assertSuccess[cmd](t, func(x cmd) {
// 		require.Equal(t, []string{"rest1", "rest2", "rest3"}, x.Positional.Rest)
// 	}, "rest1", "rest2", "rest3")
// }

// func TestPositionalRequiredRestRangeFail(t *testing.T) {
// 	t.Skip()

// 	type cmd struct {
// 		Value      bool `short:"v"`
// 		Positional struct {
// 			// TODO: test using custom field for rest, that checks that at 1-2 args are provided
// 			Rest []string
// 		} `positional-args:"yes"`
// 	}

// 	require.Fail(t,
// 		`the required argument "Rest (at most 2 arguments, but got 3)" was not provided`,
// 		"rest1", "rest2", "rest3")
// }

// func TestShortTooLong(t *testing.T) {
// 	type opts struct {
// 		Value bool `short:"vv"`
// 	}

// 	assertFail[opts](t, `parse flag "Value": scan short name: must be single character, but was "vv"`)
// }

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

	{
		r, err := parser.run_inner(newArgs())
		require.Nil(t, err)
		require.Equal(t, r, fun.Pair[bool, bool]{false, false})
	}

	{
		r, err := parser.run_inner(newArgs("-a", "-b"))
		require.Nil(t, err)
		require.Equal(t, r, fun.Pair[bool, bool]{true, true})
	}
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

// func group_help_args() {
//     let a = short('a').help("flag A, related to B").switch();
//     let b = short('b').help("flag B, related to A").switch();
//     let c = short('c').help("flag C, unrelated").switch();
//     let ab = construct(a, b).group_help("Explanation applicable for both A and B:");
//     let parser = construct(ab, c).ToOptions();

//     let help = parser.run_inner(newArgs("--help")).unwrap_err().unwrap_stdout();
//     let expected_help = "\
// Usage: [-a] [-b] [-c]

// Explanation applicable for both A and B:
//     -a          flag A, related to B
//     -b          flag B, related to A

// Available options:
//     -c          flag C, unrelated
//     -h, --help  Prints help information
// ";

//     require.Equal(t, expected_help, help);
// }

// func group_help_commands() {
//     let a = short('a')
//         .switch()
//         .ToOptions()
//         .command("cmd_a")
//         .help("command that does A");
//     let b = short('a')
//         .switch()
//         .ToOptions()
//         .command("cmd_b")
//         .help("command that does B")
//         .boxed();
//     let c = short('a')
//         .switch()
//         .ToOptions()
//         .command("cmd_c")
//         .help("command that does C");
//     let parser = construct([a, b]).group_help("Explanation applicable for both A and B:");

//     let parser = construct([parser, c]).ToOptions();

//     let help = parser.run_inner(newArgs("--help")).unwrap_err().unwrap_stdout();
//     let expected_help = "\
// Usage: COMMAND ...

// Explanation applicable for both A and B:
//     cmd_a       command that does A
//     cmd_b       command that does B

// Available options:
//     -h, --help  Prints help information

// Available commands:
//     cmd_c       command that does C
// ";
//     require.Equal(t, expected_help, help);
// }

// func from_several_alternatives_pick_more_meaningful() {
//     let a = short('a').req_flag(());
//     let b = short('b').req_flag(());
//     let c = short('c').req_flag(());
//     let parser = construct([a, b, c]).ToOptions();

//     let err1 = parser.run_inner(newArgs("-a", "-b")).unwrap_err().unwrap_stderr();
//     require.Equal(t, err1, "`-b` cannot be used at the same time as `-a`");

//     let err2 = parser.run_inner(newArgs("-b", "-a")).unwrap_err().unwrap_stderr();
//     require.Equal(t, err2, "`-a` cannot be used at the same time as `-b`");

//     let err3 = parser.run_inner(newArgs("-c", "-a")).unwrap_err().unwrap_stderr();
//     require.Equal(t, err3, "`-a` cannot be used at the same time as `-c`");

//     let err4 = parser.run_inner(newArgs("-a", "-c")).unwrap_err().unwrap_stderr();
//     require.Equal(t, err4, "`-c` cannot be used at the same time as `-a`");

//     let err5 = parser
//         .run_inner(newArgs("-c", "-b", "-a"])
//         .unwrap_err()
//         .unwrap_stderr();
//     require.Equal(t, err5, "`-b` cannot be used at the same time as `-c`");
// }

// func subcommands() {
//     let bar = short('b').switch();

//     let bar_cmd = bar.ToOptions().descr("This is local info").command("bar");

//     let parser = bar_cmd.ToOptions().descr("This is global info");

//     let help = parser.run_inner(newArgs("--help")).unwrap_err().unwrap_stdout();
//     let expected_help = "\
// This is global info

// Usage: COMMAND ...

// Available options:
//     -h, --help  Prints help information

// Available commands:
//     bar         This is local info
// ";
//     require.Equal(t, expected_help, help);

//     let help = parser
//         .run_inner(newArgs("bar", "--help"])
//         .unwrap_err()
//         .unwrap_stdout();
//     let expected_help = "\
// This is local info

// Usage: bar [-b]

// Available options:
//     -b
//     -h, --help  Prints help information
// ";
//     require.Equal(t, expected_help, help);
// }

// func multiple_aliases() {
//     let a = short('a').short('b').short('c').req_flag(());
//     let parser = a.ToOptions();

//     let help = parser.run_inner(newArgs("--help")).unwrap_err().unwrap_stdout();
//     let expected_help = "\
// Usage: -a

// Available options:
//     -a
//     -h, --help  Prints help information
// ";
//     require.Equal(t, expected_help, help);
//     parser.run_inner(newArgs("-a")).unwrap();
//     parser.run_inner(newArgs("-b")).unwrap();
//     parser.run_inner(newArgs("-c")).unwrap();
// }

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
//         let dry_run = long("dry_run").switch();
//         let all = long("all").switch();
//         let repository = positional::<String>("SRC").fallback("origin".to_string());
//         let fetch = construct(Opt::Fetch {
//             dry_run,
//             all,
//             repository
//         });
//         let fetch_inner = fetch
//             .ToOptions()
//             .descr("fetches branches from remote repository");
//         let fetch_cmd = fetch_inner.command("fetch");

//         let interactive = short('i').switch();
//         let all = long("all").switch();
//         let files = positional::<PathBuf>("FILE").many();
//         let add = construct(Opt::Add {
//             interactive,
//             all,
//             files
//         });
//         let add_inner = add.ToOptions().descr("add files to the staging area");
//         let add_cmd = add_inner.command("add");

//         construct([fetch_cmd, add_cmd])
//             .ToOptions()
//             .descr("The stupid content tracker")
//     }

//     func no_command() {
//         let parser = setup();

//         let expected_err = "expected `COMMAND ...`, pass `--help` for usage information";
//         require.Equal(t,
//             expected_err,
//             parser.run_inner(newArgs()).unwrap_err().unwrap_stderr()
//         );
//     }

//     func root_help() {
//         let parser = setup();
//         let expected_help = "\
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
//         let parser = setup();
//         let expected_help = "\
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
//                 .run_inner(newArgs("fetch", "--help"])
//                 .unwrap_err()
//                 .unwrap_stdout()
//         );
//     }

//     func add_help() {
//         let parser = setup();
//         let expected_help = "\
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
//                 .run_inner(newArgs("add", "--help"])
//                 .unwrap_err()
//                 .unwrap_stdout()
//         );
//     }
// }

// func arg_bench() {
//     use std::path::PathBuf;

//     #[derive(Clone, Debug, PartialEq, Eq)]
//     struct AppArgs {
//         number: u32,
//         opt_number: Option<u32>,
//         width: u32,
//         input: Vec<PathBuf>,
//     }

//     let number = long("number")
//         .help("Sets a number\nin two lines")
//         .argument::<u32>("number");

//     let opt_number = long("opt-number")
//         .help("Sets an optional number")
//         .argument::<u32>("opt-number")
//         .optional();

//     let width = long("width")
//         .help("Sets width")
//         .argument::<u32>("width")
//         .guard(|n| *n > 0, "Width must be positive")
//         .fallback(10);

//     let input = positional::<PathBuf>("INPUT").many();

//     let parser = construct(AppArgs {
//         number,
//         opt_number,
//         width,
//         input
//     })
//     .ToOptions();

//     require.Equal(t,
//         AppArgs {
//             number: 42,
//             opt_number: None,
//             width: 10,
//             input: vec[PathBuf::from("foo"), PathBuf::from("foo2")],
//         },
//         parser
//             .run_inner(newArgs("--number", "42", "foo", "foo2"])
//             .unwrap()
//     );

//     require.Equal(t,
//         AppArgs {
//             number: 42,
//             opt_number: None,
//             width: 10,
//             input: Vec::new()
//         },
//         parser.run_inner(newArgs("--number", "42")).unwrap()
//     );

//     drop(parser);
// }

// func simple_cargo_helper() {
//     let a = short('a').long("AAAAA").help("two lines\nof help").switch();
//     let b = short('b').switch();
//     let parser = construct(a, b);
//     let decorated = cargo_helper("simple", parser)
//         .ToOptions()
//         .descr("this is a test");

//     // cargo run variant
//     let ok = decorated.run_inner(newArgs("-a")).unwrap();
//     require.Equal(t, (true, false), ok);

//     // cargo simple variant
//     let ok = decorated.run_inner(newArgs("simple", "-b")).unwrap();
//     require.Equal(t, (false, true), ok);

//     let err = decorated
//         .run_inner(newArgs("-a", "-a"])
//         .unwrap_err()
//         .unwrap_stderr();
//     require.Equal(t,
//         "argument `-a` cannot be used multiple times in this context",
//         err
//     );

//     let help = decorated.run_inner(newArgs("-h")).unwrap_err().unwrap_stdout();

//     let expected_help = "\
// this is a test

// Usage: [-a] [-b]

// Available options:
//     -a, --AAAAA  two lines of help
//     -b
//     -h, --help   Prints help information
// ";
//     require.Equal(t, expected_help, help);
// }

// func long_path_in_construct() {
//     let a = short('a').switch();
//     let _ = construct(std::option::Option::Some(a));

//     let b = short('b').switch();
//     let _ = construct!(::std::option::Option::Some(b));
// }

// func hidden_env() {
//     let name = "BPAF_SECRET_API_KEY2";
//     let visible = long("key")
//         .help("use this secret key\n two lines")
//         .argument::<String>("KEY");
//     let hidden = env(name).argument("KEY");
//     let parser = construct!([visible, hidden]).ToOptions();

//     let help = parser.run_inner(newArgs("-h")).unwrap_err().unwrap_stdout();

//     let expected = "\
// Usage: --key=KEY

// Available options:
//         --key=KEY  use this secret key
//                    two lines
//     -h, --help     Prints help information
// ";

//     require.Equal(t, help, expected);

//     let r = parser.run_inner(newArgs()).unwrap_err().unwrap_stderr();
//     require.Equal(t, r, "environment variable `BPAF_SECRET_API_KEY2` is not set");
// }

// func env_variable() {
//     let name = "BPAF_SECRET_API_KEY";
//     let parser = long("key")
//         .env(name)
//         .help("use this secret key\ntwo lines")
//         .argument::<String>("KEY")
//         .ToOptions();

//     let help = parser.run_inner(newArgs("-h")).unwrap_err().unwrap_stdout();
//     let expected_help = "\
// Usage: --key=KEY

// Available options:
//         --key=KEY  use this secret key two lines
//                    [env:BPAF_SECRET_API_KEY: N/A]
//     -h, --help     Prints help information
// ";
//     require.Equal(t, expected_help, help);
//     std::env::set_var(name, "top s3cr3t");

//     let help = parser.run_inner(newArgs("-h")).unwrap_err().unwrap_stdout();
//     let expected_help = "\
// Usage: --key=KEY

// Available options:
//         --key=KEY  use this secret key two lines
//                    [env:BPAF_SECRET_API_KEY = \"top s3cr3t\"]
//     -h, --help     Prints help information
// ";
//     require.Equal(t, expected_help, help);

//     let res = parser.run_inner(newArgs("--key", "secret")).unwrap();
//     require.Equal(t, res, "secret");

//     let res = parser.run_inner(newArgs()).unwrap();
//     require.Equal(t, res, "top s3cr3t");
// }

// func default_plays_nicely_with_command() {
//     #[derive(Debug, Clone)]
//     enum Foo {
//         Foo,
//         Bar,
//     }
//     impl Default for Foo {
//         func default() -> Self {
//             Foo::Bar
//         }
//     }

//     let cmd = pure(Foo::Foo)
//         .ToOptions()
//         .descr("inner")
//         .command("foo")
//         .help("foo")
//         .fallback(Default::default());

//     let parser = cmd.ToOptions().descr("outer");

//     let help = parser
//         .run_inner(newArgs("foo", "--help"])
//         .unwrap_err()
//         .unwrap_stdout();

//     let expected_help =
//         "inner\n\nUsage: foo \n\nAvailable options:\n    -h, --help  Prints help information\n";

//     require.Equal(t, expected_help, help);

//     let help = parser.run_inner(newArgs("--help")).unwrap_err().unwrap_stdout();

//     let expected_help = "\
// outer

// Usage: [COMMAND ...]

// Available options:
//     -h, --help  Prints help information

// Available commands:
//     foo         foo
// ";

//     require.Equal(t, expected_help, help);
// }

// func command_with_aliases() {
//     let inner = pure(()).ToOptions().descr("inner descr");
//     let cmd = inner.command("foo").long("bar").short('f').short('b');
//     let parser = cmd.ToOptions().descr("outer");

//     let help = parser.run_inner(newArgs("--help")).unwrap_err().unwrap_stdout();

//     let expected_help = "\
// outer

// Usage: COMMAND ...

// Available options:
//     -h, --help  Prints help information

// Available commands:
//     foo, f      inner descr
// ";
//     require.Equal(t, expected_help, help);

//     let help = parser
//         .run_inner(newArgs("f", "--help"])
//         .unwrap_err()
//         .unwrap_stdout();

//     let expected_help =
//         "inner descr\n\nUsage: foo \n\nAvailable options:\n    -h, --help  Prints help information\n";
//     require.Equal(t, expected_help, help);

//     // hidden and visible aliases are working
//     parser.run_inner(newArgs("foo")).unwrap();
//     parser.run_inner(newArgs("f")).unwrap();
//     parser.run_inner(newArgs("bar")).unwrap();
//     parser.run_inner(newArgs("b")).unwrap();

//     // and "k" isn't a thing
//     parser.run_inner(newArgs("k")).unwrap_err();
// }

// func help_for_options() {
//     let a = short('a').help("help for\na").switch();
//     let b = short('c')
//         .env("BbBbB")
//         .help("help for\nb")
//         .argument::<String>("B");
//     let c = long("bbbbb")
//         .env("ccccCCccc")
//         .help("help for\nccc")
//         .argument::<String>("CCC");
//     let parser = construct!(a, b, c).ToOptions();
//     let help = parser.run_inner(newArgs("--help")).unwrap_err().unwrap_stdout();

//     let expected_help = "\
// Usage: [-a] -c=B --bbbbb=CCC

// Available options:
//     -a               help for a
//     -c=B             help for b
//                      [env:BbBbB: N/A]
//         --bbbbb=CCC  help for ccc
//                      [env:ccccCCccc: N/A]
//     -h, --help       Prints help information
// ";

//     require.Equal(t, expected_help, help);
// }

// func help_for_commands() {
//     let d = pure(())
//         .ToOptions()
//         .command("thing_d")
//         .help("help for d\ntwo lines");
//     let e = pure(())
//         .ToOptions()
//         .command("thing_e")
//         .short('e')
//         .help("help for e\ntwo lines");
//     let h = pure(()).ToOptions().command("thing_h");
//     let parser = construct!([d, e, h]).ToOptions();
//     let help = parser.run_inner(newArgs("--help")).unwrap_err().unwrap_stdout();

//     let expected_help = "\
// Usage: COMMAND ...

// Available options:
//     -h, --help  Prints help information

// Available commands:
//     thing_d     help for d two lines
//     thing_e, e  help for e two lines
//     thing_h
// ";
//     require.Equal(t, expected_help, help);
// }

// func many_doesnt_panic() {
//     let parser = short('a').switch().many().map(|m| m.len()).ToOptions();
//     let r = parser.run_inner(newArgs("-aaa")).unwrap();
//     require.Equal(t, r, 3);
// }

// func some_doesnt_panic() {
//     let parser = short('a').switch().some("").map(|m| m.len()).ToOptions();
//     let r = parser.run_inner(newArgs("-aaa")).unwrap();
//     require.Equal(t, r, 3);
// }

// func command_resets_left_head_state() {
//     #[derive(Debug, Eq, PartialEq)]
//     enum Foo {
//         Bar1 { a: u32 },
//         Bar2 { b: () },
//     }

//     let a = short('a').argument::<u32>("A").fallback(0);
//     let b = short('b').req_flag(());

//     let p1 = construct!(Foo::Bar1 { a });
//     let p2 = construct!(Foo::Bar2 { b });
//     let cmd = construct!([p1, p2])
//         .ToOptions()
//         .command("cmd")
//         .ToOptions();

//     let xx = cmd.run_inner(newArgs("cmd", "-b")).unwrap();
//     require.Equal(t, xx, Foo::Bar2 { b: () });
// }

// func command_preserves_custom_failure_message() {
//     let msg = "need more cheese";
//     let inner = fail::<()>(msg).ToOptions();

//     let err = inner.run_inner(newArgs()).unwrap_err().unwrap_stderr();
//     require.Equal(t, err, "need more cheese");

//     let outer = inner.command("feed").ToOptions();

//     let err = outer.run_inner(newArgs("feed")).unwrap_err().unwrap_stderr();
//     require.Equal(t, err, "need more cheese");
// }

// func optional_error_handling() {
//     let p = short('p').argument::<u32>("P").optional().ToOptions();

//     let res = p.run_inner(newArgs()).unwrap();
//     require.Equal(t, res, None);

//     let res = p.run_inner(newArgs("-p", "3")).unwrap();
//     require.Equal(t, res, Some(3));

//     let res = p.run_inner(newArgs("-p", "pi")).unwrap_err().unwrap_stderr();
//     require.Equal(t, res, "couldn't parse `pi`: invalid digit found in string");
// }

// func many_error_handling() {
//     let p = short('p').argument::<u32>("P").many().ToOptions();

//     let res = p.run_inner(newArgs()).unwrap();
//     require.Equal(t, res, Vec::new());

//     let res = p.run_inner(newArgs("-p", "3")).unwrap();
//     require.Equal(t, res, vec![3]);

//     let res = p.run_inner(newArgs("-p", "pi")).unwrap_err().unwrap_stderr();
//     require.Equal(t, res, "couldn't parse `pi`: invalid digit found in string");
// }

// func failure_is_not_stupid_1() {
//     let a = short('a').argument::<u32>("A");
//     let b = pure(()).parse::<_, _, String>(|_| Err("nope".to_string()));
//     let parser = construct!(a, b).ToOptions();

//     let res = parser.run_inner(newArgs("-a", "42")).unwrap_err().unwrap_stderr();
//     require.Equal(t, res, "couldn't parse: nope");
// }

// func failure_is_not_stupid_2() {
//     let a = short('a').argument::<u32>("A");
//     let b = short('b').argument::<u32>("B");
//     let parser = construct!(a, b)
//         .parse::<_, (), String>(|_| Err("nope".to_string()))
//         .ToOptions();

//     let res = parser
//         .run_inner(newArgs("-a", "42", "-b", "42"])
//         .unwrap_err()
//         .unwrap_stderr();
//     require.Equal(t, res, "couldn't parse: nope");
// }

// func no_fallback_out_of_command_parser() {
//     let alt1 = positional::<String>("NAME").ToOptions().command("cmd");
//     let alt2 = pure(String::new());
//     let parser = construct!([alt1, alt2]).ToOptions();

//     let res = parser.run_inner(newArgs("cmd")).unwrap_err().unwrap_stderr();
//     require.Equal(t, res, "expected `NAME`, pass `--help` for usage information");

//     let res = parser.run_inner(newArgs("cmd", "a")).unwrap();
//     require.Equal(t, res, "a");

//     let res = parser.run_inner(newArgs()).unwrap();
//     require.Equal(t, res, "");
// }

// func did_you_mean_switch() {
//     let a = short('f').long("flag").switch();
//     let b = short('p').long("plag").switch();
//     let parser = construct!([a, b]).ToOptions();

//     let res = parser.run_inner(newArgs("--fla")).unwrap_err().unwrap_stderr();
//     require.Equal(t, res, "no such flag: `--fla`, did you mean `--flag`?");

//     let res = parser.run_inner(newArgs("flag")).unwrap_err().unwrap_stderr();
//     require.Equal(t,
//         res,
//         "no such command or positional: `flag`, did you mean `--flag`?"
//     );

//     let res = parser.run_inner(newArgs("--pla")).unwrap_err().unwrap_stderr();
//     require.Equal(t, res, "no such flag: `--pla`, did you mean `--plag`?");

//     let res = parser.run_inner(newArgs("--p")).unwrap_err().unwrap_stderr();
//     require.Equal(t,
//         res,
//         "no such flag: `--p` (with two dashes), did you mean `-p`?"
//     );
// }

// func did_you_mean_req_flag() {
//     let parser = long("flag").req_flag(()).ToOptions();
//     let res = parser.run_inner(newArgs("--fla")).unwrap_err().unwrap_stderr();
//     require.Equal(t, res, "no such flag: `--fla`, did you mean `--flag`?");
// }

// func did_you_mean_argument() {
//     let parser = long("flag").argument::<String>("VAL").ToOptions();

//     let res = parser.run_inner(newArgs("--fla")).unwrap_err().unwrap_stderr();
//     require.Equal(t, res, "no such flag: `--fla`, did you mean `--flag`?");

//     let res = parser
//         .run_inner(newArgs("--flg=hellop"])
//         .unwrap_err()
//         .unwrap_stderr();
//     require.Equal(t, res, "no such flag: `--flg`, did you mean `--flag`?");
// }

// func did_you_mean_command() {
//     let parser = pure(())
//         .ToOptions()
//         .command("command")
//         .short('c')
//         .ToOptions();

//     let res = parser.run_inner(newArgs("comman")).unwrap_err().unwrap_stderr();
//     require.Equal(t,
//         res,
//         "no such command or positional: `comman`, did you mean `command`?"
//     );

//     let res = parser.run_inner(newArgs("--comman")).unwrap_err().unwrap_stderr();
//     require.Equal(t, res, "no such flag: `--comman`, did you mean `command`?");
// }

// func did_you_mean_two_and_arguments() {
//     let a = long("flag").switch();
//     let b = long("parameter").switch();
//     let parser = cargo_helper("cmd", construct!(a, b)).ToOptions();

//     let r = parser
//         .run_inner(newArgs("--flag", "--parametr"])
//         .unwrap_err()
//         .unwrap_stderr();
//     require.Equal(t, r, "no such flag: `--parametr`, did you mean `--parameter`?");

//     let r = parser
//         .run_inner(newArgs("--flag", "--paramet=value"])
//         .unwrap_err()
//         .unwrap_stderr();
//     require.Equal(t, r, "no such flag: `--paramet`, did you mean `--parameter`?");

//     let r = parser
//         .run_inner(newArgs("--parameter", "--flg"])
//         .unwrap_err()
//         .unwrap_stderr();
//     require.Equal(t, r, "no such flag: `--flg`, did you mean `--flag`?");

//     let r = parser.run_inner(newArgs("--fla")).unwrap_err().unwrap_stderr();
//     require.Equal(t, r, "no such flag: `--fla`, did you mean `--flag`?");
// }

// func did_you_mean_two_or_arguments() {
//     let a = long("flag").switch();
//     let b = long("parameter").switch();
//     let parser = cargo_helper("cmd", construct!([a, b])).ToOptions();

//     let r = parser.run_inner(newArgs("--fla")).unwrap_err().unwrap_stderr();
//     require.Equal(t, r, "no such flag: `--fla`, did you mean `--flag`?");

//     let r = parser
//         .run_inner(newArgs("--flag", "--parametr"])
//         .unwrap_err()
//         .unwrap_stderr();
//     require.Equal(t, r, "no such flag: `--parametr`, did you mean `--parameter`?");

//     let r = parser
//         .run_inner(newArgs("--parametr", "--flag"])
//         .unwrap_err()
//         .unwrap_stderr();
//     require.Equal(t, r, "no such flag: `--parametr`, did you mean `--parameter`?");

//     let r = parser
//         .run_inner(newArgs("--parameter", "--flag"])
//         .unwrap_err()
//         .unwrap_stderr();
//     require.Equal(t,
//         r,
//         "`--flag` cannot be used at the same time as `--parameter`"
//     );

//     let r = parser
//         .run_inner(newArgs("--flag", "--parameter"])
//         .unwrap_err()
//         .unwrap_stderr();
//     require.Equal(t,
//         r,
//         "`--parameter` cannot be used at the same time as `--flag`"
//     );
// }

// // problematic steps look something like this:
// // - "asm" is parsed as expected
// // - "-t x" is consumed as expected
// // - parsing of "x" fails
// // - ParseWith rollbacks the arguments state - "asm" is back
// // - suggestion looks for something it can complain at and finds "asm"
// //
// // parse/guard failures should "taint" the arguments and disable the suggestion logic

// func cargo_show_asm_issue_guard() {
//     let target_dir = short('t').argument::<String>("T").guard(|_| false, "nope");
//     let verbosity = short('v').switch();
//     let inner = construct!(target_dir, verbosity);
//     let parser = cargo_helper("asm", inner).ToOptions();

//     let res = parser
//         .run_inner(newArgs("asm", "-t", "x"])
//         .unwrap_err()
//         .unwrap_stderr();
//     require.Equal(t, res, "`x`: nope");

//     let res = parser.run_inner(newArgs("-t", "x")).unwrap_err().unwrap_stderr();
//     require.Equal(t, res, "`x`: nope");
// }

// func cargo_show_asm_issue_from_str() {
//     let target_dir = short('t').argument::<usize>("T");
//     let verbosity = short('v').switch();
//     let inner = construct!(target_dir, verbosity);
//     let parser = cargo_helper("asm", inner).ToOptions();

//     let res = parser
//         .run_inner(newArgs("asm", "-t", "x"])
//         .unwrap_err()
//         .unwrap_stderr();
//     require.Equal(t, res, "couldn't parse `x`: invalid digit found in string");

//     let res = parser.run_inner(newArgs("-t", "x")).unwrap_err().unwrap_stderr();
//     require.Equal(t, res, "couldn't parse `x`: invalid digit found in string");
// }

// func cargo_show_asm_issue_parse() {
//     let target_dir = short('t')
//         .argument::<String>("T")
//         .parse::<_, (), String>(|_| Err("nope".to_string()));

//     let verbosity = short('v').switch();
//     let inner = construct!(target_dir, verbosity);
//     let parser = cargo_helper("asm", inner).ToOptions();

//     let res = parser
//         .run_inner(newArgs("asm", "-t", "x"])
//         .unwrap_err()
//         .unwrap_stderr();
//     require.Equal(t, res, "couldn't parse `x`: nope");

//     let res = parser.run_inner(newArgs("-t", "x")).unwrap_err().unwrap_stderr();
//     require.Equal(t, res, "couldn't parse `x`: nope");
// }

// // problematic case looks something like this:
// // "asm" is consumed
// // "--fla" is not consumed, but the output is not tainted
// // parser fails and arguments are restored.

// func cargo_show_asm_issue_unknown_switch() {
//     let target = long("flag").switch();

//     let verbosity = short('v').switch();
//     let parser = cargo_helper("asm", construct!(target, verbosity)).ToOptions();

//     let res = parser
//         .run_inner(newArgs("asm", "--fla"])
//         .unwrap_err()
//         .unwrap_stderr();
//     require.Equal(t, res, "no such flag: `--fla`, did you mean `--flag`?");

//     let res = parser.run_inner(newArgs("--fla")).unwrap_err().unwrap_stderr();
//     require.Equal(t, res, "no such flag: `--fla`, did you mean `--flag`?");
// }

// func did_you_mean_inside_command() {
//     let a = long("flag").switch();
//     let b = long("parameter").switch();
//     let parser = construct!([a, b]).ToOptions().command("cmd").ToOptions();

//     let r = parser.run_inner(newArgs("--fla")).unwrap_err().unwrap_stderr();
//     require.Equal(t,
//         r,
//         "expected `COMMAND ...`, got `--fla`. Pass `--help` for usage information"
//     );

//     let r = parser
//         .run_inner(newArgs("cmd", "--fla"])
//         .unwrap_err()
//         .unwrap_stderr();
//     require.Equal(t, r, "no such flag: `--fla`, did you mean `--flag`?");

//     let r = parser
//         .run_inner(newArgs("cmd", "--flag", "--parametr"])
//         .unwrap_err()
//         .unwrap_stderr();
//     require.Equal(t, r, "no such flag: `--parametr`, did you mean `--parameter`?");

//     let r = parser
//         .run_inner(newArgs("cmd", "--parametr", "--flag"])
//         .unwrap_err()
//         .unwrap_stderr();
//     require.Equal(t, r, "no such flag: `--parametr`, did you mean `--parameter`?");

//     let r = parser
//         .run_inner(newArgs("cmd", "--parameter", "--flag"])
//         .unwrap_err()
//         .unwrap_stderr();
//     require.Equal(t,
//         r,
//         "`--flag` cannot be used at the same time as `--parameter`"
//     );

//     let r = parser
//         .run_inner(newArgs("cmd", "--flag", "--parameter"])
//         .unwrap_err()
//         .unwrap_stderr();
//     require.Equal(t,
//         r,
//         "`--parameter` cannot be used at the same time as `--flag`"
//     );
// }

// func combine_flags_by_order() {
//     let a = short('a').req_flag(true);
//     let b = short('A').req_flag(false);
//     let parser = construct!([a, b]).many().ToOptions();

//     let r = parser.run_inner(newArgs("-a", "-A", "-A", "-A", "-a")).unwrap();
//     require.Equal(t, r, vec![true, false, false, false, true]);
// }

// func parse_many_errors_positional() {
//     let p = positional::<u32>("N").many().ToOptions();

//     let r = p.run_inner(newArgs("1", "2", "3")).unwrap();
//     require.Equal(t, r, vec![1, 2, 3]);

//     let r = p.run_inner(newArgs("1", "2", "x")).unwrap_err().unwrap_stderr();
//     require.Equal(t, r, "couldn't parse `x`: invalid digit found in string");
// }

// func parse_collect_flag() {
//     let p = short('p')
//         .argument::<u32>("N")
//         .collect::<Vec<_>>()
//         .ToOptions();

//     let r = p.run_inner(newArgs("-p", "1", "-p", "2")).unwrap();
//     require.Equal(t, r, vec![1, 2]);

//     let r = p
//         .run_inner(newArgs("-p", "1", "-p", "x"])
//         .unwrap_err()
//         .unwrap_stderr();
//     require.Equal(t, r, "couldn't parse `x`: invalid digit found in string");
// }

// func parse_many_errors_flag() {
//     let p = short('p').argument::<u32>("N").many().ToOptions();

//     let r = p.run_inner(newArgs("-p", "1", "-p", "2")).unwrap();
//     require.Equal(t, r, vec![1, 2]);

//     let r = p
//         .run_inner(newArgs("-p", "1", "-p", "x"])
//         .unwrap_err()
//         .unwrap_stderr();
//     require.Equal(t, r, "couldn't parse `x`: invalid digit found in string");
// }

// func command_with_req_parameters() {
//     let p = positional::<String>("X")
//         .ToOptions()
//         .command("cmd")
//         .fallback(String::new())
//         .ToOptions();

//     let r = p.run_inner(newArgs("cmd")).unwrap_err().unwrap_stderr();
//     require.Equal(t, r, "expected `X`, pass `--help` for usage information");
// }

// func suggestion_for_equals_1() {
//     let parser = short('p').long("par").argument::<String>("P").ToOptions();

//     let r = parser
//         .run_inner(newArgs("-p", "--bar"])
//         .unwrap_err()
//         .unwrap_stderr();
//     require.Equal(t,
//         r,
//         "`-p` requires an argument `P`, got a flag `--bar`, try `-p=--bar` to use it as an argument"
//     );

//     let r = parser
//         .run_inner(newArgs("--par", "--bar"])
//         .unwrap_err()
//         .unwrap_stderr();
//     require.Equal(t,
//         r,
//         "`--par` requires an argument `P`, got a flag `--bar`, try `--par=--bar` to use it as an argument"
//     );

//     let r = parser
//         .run_inner(newArgs("--par", "--bar=baz"])
//         .unwrap_err()
//         .unwrap_stderr();
//     require.Equal(t,
//         r,
//         "`--par` requires an argument `P`, got a flag `--bar=baz`, try `--par=--bar=baz` to use it as an\nargument"
//     );
// }

// func double_dash_is_pos_only_just_once() {
//     let parser = positional::<String>("POS").many().ToOptions();

//     let r = parser.run_inner(newArgs("--")).unwrap();
//     require.Equal(t, r, Vec::<String>::new());

//     let r = parser.run_inner(newArgs("--", "--")).unwrap();
//     require.Equal(t, r, vec!["--".to_string()]);
// }

// func reject_fbar() {
//     let parser = short('f').argument::<String>("F").ToOptions();

//     let r = parser
//         .run_inner(newArgs("-fbar", "baz"])
//         .unwrap_err()
//         .unwrap_stderr();
//     require.Equal(t, r, "`baz` is not expected in this context");

//     let r = parser.run_inner(newArgs("-fbar")).unwrap();
//     require.Equal(t, r, "bar");
// }

// func custom_usage_override_fixed() {
//     let parser = short('p').switch().ToOptions().usage("Usage: hey [-p]");
//     let r = parser.run_inner(newArgs("--help")).unwrap_err().unwrap_stdout();
//     require.Equal(t,
//         r,
//         "Usage: hey [-p]\n\nAvailable options:\n    -p\n    -h, --help  Prints help information\n"
//     );
// }

// func custom_usage_override_with_fn() {
//     let parser = short('p').switch().ToOptions().with_usage(|b| {
//         let mut buf = Doc::default();
//         buf.text("Usage: hey ");
//         buf.doc(&b);
//         buf
//     });
//     let r = parser.run_inner(newArgs("--help")).unwrap_err().unwrap_stdout();
//     require.Equal(t,
//         r,
//         "Usage: hey [-p]\n\nAvailable options:\n    -p\n    -h, --help  Prints help information\n"
//     );
// }

// func catch_works() {
//     #[derive(Debug, Eq, PartialEq)]
//     enum A {
//         Num(usize),
//         Str(String),
//     }
//     let a_n = short('a')
//         .argument::<usize>("A")
//         .map(A::Num)
//         .optional()
//         .catch();
//     let a_s = short('a')
//         .argument::<String>("A")
//         .map(A::Str)
//         .hide()
//         .optional();
//     let parser = construct!([a_n, a_s]).ToOptions();

//     let r = parser.run_inner(newArgs("-a", "1")).unwrap();
//     require.Equal(t, r, Some(A::Num(1)));

//     let r = parser.run_inner(newArgs("-a", "x1")).unwrap();
//     require.Equal(t, r, Some(A::Str("x1".to_owned())));
// }

// func sneaky_command() {
//     #[derive(Debug, Eq, PartialEq)]
//     enum Cmd {
//         A(bool),
//         B(bool),
//     }
//     let cmd = short('f')
//         .switch()
//         .ToOptions()
//         .command("hello")
//         .map(Cmd::A);
//     let b = short('f').switch().map(Cmd::B);

//     let sneaky = true;

//     let maybe_cmd = if sneaky {
//         construct!(cmd)
//     } else {
//         let nope = fail("no sneaky");
//         construct!(nope)
//     };
//     let parser = construct!([maybe_cmd, b]).ToOptions();

//     let r = parser.run_inner(newArgs("hello")).unwrap();
//     require.Equal(t, r, Cmd::A(false));
// }

// func default_for_many() {
//     let parser = positional::<String>("ROOTS")
//         .many()
//         .map(|items| {
//             if items.is_empty() {
//                 vec![String::from(".")]
//             } else {
//                 items
//             }
//         })
//         .ToOptions();

//     let r = parser.run_inner(newArgs()).unwrap();
//     require.Equal(t, r, vec![String::from(".")]);
// }

// func parse_option_catch() {
//     #[derive(Debug, Clone, Eq, PartialEq)]
//     enum A {
//         U32(u32),
//         S(String),
//     }
//     let a1 = short('a').argument("N").map(A::U32).optional().catch();
//     let a2 = short('a').argument("S").map(A::S).optional().catch();
//     let parser = construct!([a1, a2]).ToOptions();

//     let r = parser.run_inner(newArgs("-a", "10")).unwrap();
//     require.Equal(t, r, Some(A::U32(10)));

//     let r = parser.run_inner(newArgs("-a", "x")).unwrap();
//     require.Equal(t, r, Some(A::S("x".to_string())));

//     let r = parser.run_inner(newArgs()).unwrap();
//     require.Equal(t, r, None);
// }

// func parse_some_catch() {
//     #[derive(Debug, Clone, Eq, PartialEq)]
//     enum A {
//         U32(u32),
//         S(String),
//     }
//     let a1 = short('a')
//         .argument("N")
//         .map(A::U32)
//         .some("A")
//         .catch()
//         .hide();
//     let a2 = short('a').argument("S").map(A::S).some("A").catch().hide();
//     let parser = construct!([a1, a2]).ToOptions();

//     let r = parser.run_inner(newArgs("-a", "10")).unwrap();
//     require.Equal(t, r, vec![A::U32(10)]);

//     let r = parser.run_inner(newArgs("-a", "x")).unwrap();
//     require.Equal(t, r, vec![A::S("x".to_string())]);

//     let r = parser.run_inner(newArgs()).unwrap_err().unwrap_stderr();
//     require.Equal(t, r, "A");
// }

// func empty_struct() {
//     #[derive(Debug, Clone, Eq, PartialEq)]
//     struct Foo {}
//     let parser = construct!(Foo {}).ToOptions();

//     let r = parser.run_inner(newArgs()).unwrap();
//     require.Equal(t, r, Foo {});

//     let r = parser.run_inner(newArgs("--help")).unwrap_err().unwrap_stdout();
//     require.Equal(t,
//         r,
//         "Usage: \n\nAvailable options:\n    -h, --help  Prints help information\n"
//     );
// }

// func empty_tuple() {
//     #[derive(Debug, Clone, Eq, PartialEq)]
//     struct Foo();
//     let parser = construct!(Foo()).ToOptions();

//     let r = parser.run_inner(newArgs()).unwrap();
//     require.Equal(t, r, Foo());

//     let r = parser.run_inner(newArgs("--help")).unwrap_err().unwrap_stdout();
//     require.Equal(t,
//         r,
//         "Usage: \n\nAvailable options:\n    -h, --help  Prints help information\n"
//     );
// }

// func strange_short_option() {
//     let parser = short('O').argument::<String>("ARG").ToOptions();
//     let r = parser.run_inner(newArgs("-Obits=2048")).unwrap();
//     require.Equal(t, r, "bits=2048");
// }

// func optional_bool_states() {
//     let parser = short('a').switch().optional().ToOptions();

//     let r = parser.run_inner(newArgs("-a")).unwrap();
//     require.Equal(t, r, Some(true));

//     let r = parser.run_inner(newArgs()).unwrap();
//     require.Equal(t, r, Some(false));
// }

// func fancy_negative() {
//     let a = short('a').req_flag(());
//     #[allow(clippy::redundant_closure)]
//     let b = any("A", |i: i32| Some(i));
//     let ab = construct!(a, b).adjacent().map(|x| x.1);

//     let c = short('c').argument::<usize>("C").fallback(42);

//     let parser = construct!(ab, c).ToOptions();

//     let r = parser.run_inner(newArgs("-a", "-10")).unwrap();
//     require.Equal(t, r, (-10, 42));

//     let r = parser.run_inner(newArgs("-a=-20", "-c", "110")).unwrap();
//     require.Equal(t, r, (-20, 110));

//     let r = parser.run_inner(newArgs("--help")).unwrap_err().unwrap_stdout();

//     // TODO - rendering sucks once you start inventing fancy combinations and don't provide help...
//     let expected = "\
// Usage: -a A [-c=C]

// Available options:
//   -a A

//     -c=C
//     -h, --help  Prints help information
// ";
//     require.Equal(t, r, expected);
// }

// func many_env() {
//     std::env::set_var("USER1", "top s3cr3t");
//     let parser = short('v')
//         .env("USER1")
//         .argument::<String>("USER")
//         .many()
//         .ToOptions();
//     let r = parser.run_inner(newArgs()).unwrap();
//     require.Equal(t, r, vec!["top s3cr3t".to_owned()]);
// }

// func env_hidden_arg() {
//     std::env::set_var("USER1", "top s3cr3t");
//     let parser = env("USER1").argument::<String>("USER").ToOptions();
//     let r = parser.run_inner(newArgs()).unwrap();
//     require.Equal(t, r, "top s3cr3t");
// }

// func env_hidden_switch() {
//     std::env::set_var("USER1", "top s3cr3t");
//     let parser = env("USER1").switch().ToOptions();
//     let r = parser.run_inner(newArgs()).unwrap();
//     assert!(r);
// }

// func env_hidden_flag() {
//     std::env::set_var("USER1", "top s3cr3t");
//     let parser = env("USER1").flag(true, false).ToOptions();
//     let r = parser.run_inner(newArgs()).unwrap();
//     assert!(r);
// }

// func some_env() {
//     std::env::set_var("USER1", "top s3cr3t");
//     let parser = short('v')
//         .env("USER1")
//         .argument::<String>("USER")
//         .some("a")
//         .ToOptions();
//     let r = parser.run_inner(newArgs()).unwrap();
//     require.Equal(t, r, vec!["top s3cr3t".to_owned()]);
// }

// func option_requires_other_option1() {
//     let a = short('a').switch();
//     let b = short('b').argument::<String>("B");
//     let parser = construct!(a, b).optional().ToOptions();

//     let r = parser.run_inner(newArgs("-a")).unwrap_err().unwrap_stderr();
//     require.Equal(t, r, "expected `-b=B`, pass `--help` for usage information");
// }

// func option_requires_other_option2() {
//     let a = short('a').switch();
//     let b = short('b').argument::<String>("B");
//     let parser = construct!(b, a).optional().ToOptions();

//     let r = parser.run_inner(newArgs("-a")).unwrap_err().unwrap_stderr();
//     require.Equal(t, r, "expected `-b=B`, pass `--help` for usage information");
// }

// func default_for_some() {
//     let parser = bpaf::positional::<u32>("ROOTS")
//         .some("msg")
//         .fallback_with(|| Ok::<_, Infallible>(vec![1, 2, 3]))
//         .ToOptions();

//     let r = parser.run_inner(newArgs()).unwrap();

//     require.Equal(t, r, vec![1, 2, 3]);
// }

// func adjacent_anywhere_needs_to_consume_something() {
//     let a = short('a').switch();
//     let b = short('b').switch();
//     let parser = construct!(a, b).adjacent().ToOptions();

//     let r = parser.run_inner(newArgs("-a")).unwrap();
//     require.Equal(t, r, (true, false));

//     let r = parser.run_inner(newArgs("-b")).unwrap();
//     require.Equal(t, r, (false, true));
// }

// func fallback_for_some() {
//     let a = short('a')
//         .argument::<u32>("ARG")
//         .some("potatoes")
//         .fallback(vec![1, 2, 3]);
//     let parser = a.ToOptions();

//     let r = parser.run_inner(newArgs("-a", "1")).unwrap();
//     require.Equal(t, r, vec![1]);

//     let r = parser.run_inner(newArgs()).unwrap();
//     require.Equal(t, r, vec![1, 2, 3]);
// }

// func flag_like_commands() {
//     let a = short('a').req_flag(1).ToOptions().command("--add");
//     let b = short('b').req_flag(2).ToOptions().command("remove");
//     let parser = construct!([a, b]).ToOptions();

//     let r = parser.run_inner(newArgs("--add", "-a")).unwrap();
//     require.Equal(t, r, 1);

//     let r = parser.run_inner(newArgs("remove", "-b")).unwrap();
//     require.Equal(t, r, 2);

//     let r = parser.run_inner(newArgs("--help")).unwrap_err().unwrap_stdout();
//     let expected = "Usage: COMMAND ...\n\nAvailable options:\n    -h, --help  Prints help information\n\nAvailable commands:\n    --add\n    remove\n";
//     require.Equal(t, r, expected);

//     let r = parser
//         .run_inner(newArgs("--add", "--help"])
//         .unwrap_err()
//         .unwrap_stdout();
//     let expected =
//         "Usage: --add -a\n\nAvailable options:\n    -a\n    -h, --help  Prints help information\n";
//     require.Equal(t, r, expected);
// }
