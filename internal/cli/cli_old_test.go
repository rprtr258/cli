package cli

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

// func Test_third_subsubcommandExtraArgs(t *testing.T) {
// 	type optsThird struct {
// 		Value   bool `short:"v"`
// 		Command struct {
// 			G          bool `short:"g"`
// 			SubCommand struct {
// 				B bool `short:"b"`
// 			} `command:"sub"`
// 		} `command:"cmd"`
// 	}
// 	assertFail[optsThird](t,
// 		`extraneous positional arguments found: ["-v", "-g"]`,
// 		"cmd", "sub", "-b", "-v", "-g",
// 	)
// }

// type optsFourthSub struct {
// 	Value bool `short:"v"`
// }

// type optsFourthCmd struct {
// 	Value   bool          `short:"v"`
// 	Command optsFourthSub `command:"sub"`
// }

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

// func Test_fourth_altSubcommandFlag(t *testing.T) {
// 	type optsFourth struct {
// 		Value   bool          `short:"v"`
// 		Command optsFourthCmd `command:"cmd"`
// 	}
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
//	assertFail[optsSix](t, "Please specify one command of: add or remove\nUsage: \x1b[94mx\x1b[0m \x1b[91m[COMMAND]\x1b[0m \x1b[93m[FLAGS]\x1b[0m\n\n\n\nCommands:\n\t\x1b[91madd\x1b[0m      \n\t\x1b[91mremove\x1b[0m   \n\nFlags:\n\t\x1b[93m-h\x1b[0m, \x1b[93m--help\x1b[0m    Print this help message\n")
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

// func TestGroupInline(t *testing.T) {
// 	type cmdTestGroupInline struct {
// 		Value bool `short:"v"`
// 		G     bool `short:"g"`
// 	}
// 	assertSuccess[cmdTestGroupInline](t, func(x cmdTestGroupInline) {
// 		require.True(t, x.Value)
// 		require.True(t, x.G)
// 	}, "-v", "-g")
// }

// func TestGroupAdd(t *testing.T) {
// 	type cmdTestGroupAdd struct {
// 		Value bool     `short:"v"`
// 		G     bool     `short:"g"`
// 		Args  []string `positional-args:"yes"`
// 	}
// 	assertSuccess[cmdTestGroupAdd](t, func(x cmdTestGroupAdd) {
// 		require.Equal(t, []string{"rest"}, x.Args)
// 		require.True(t, x.Value)
// 		require.True(t, x.G)
// 	}, "-v", "-g", "rest")
// }

// func TestGroupNestedInline(t *testing.T) {
// 	type cmdTestGroupNestedInline struct {
// 		Value bool     `short:"v"`
// 		G     bool     `short:"g"`
// 		Nah   string   `long:"nah"`
// 		Args  []string `positional-args:"yes"`
// 	}
// 	assertSuccess[cmdTestGroupNestedInline](t, func(x cmdTestGroupNestedInline) {
// 		require.Equal(t, []string{"rest"}, x.Args)
// 		require.True(t, x.Value)
// 		require.True(t, x.G)
// 		require.Equal(t, x.Nah, "n")
// 	}, "-v", "-g", "--nah", "n", "rest")
// }

// func TestGroupNestedInlineNamespace(t *testing.T) {
// 	type cmdTestNestedInlineNamespace struct {
// 		Opt       string   `long:"opt"`
// 		SipOpt    string   `long:"sip.opt"`
// 		SipSapOpt string   `long:"sip.sap.opt"`
// 		Args      []string `positional-args:"yes"`
// 	}
// 	assertSuccess[cmdTestNestedInlineNamespace](t, func(x cmdTestNestedInlineNamespace) {
// 		require.Equal(t, []string{"rest"}, x.Args)
// 		require.Equal(t, x.Opt, "a")
// 		require.Equal(t, x.SipOpt, "b")
// 		require.Equal(t, x.SipSapOpt, "c")
// 	}, "--opt", "a", "--sip.opt", "b", "--sip.sap.opt", "c", "rest")
// }

// func TestBoolFlag(t *testing.T) {
// 	type cmdBoolFlag struct {
// 		Test bool `long:"test"`
// 	}
// 	assertSuccess[cmdBoolFlag](t, func(x cmdBoolFlag) {
// 		require.False(t, x.Test)
// 	})
// 	assertSuccess[cmdBoolFlag](t, func(x cmdBoolFlag) {
// 		require.True(t, x.Test)
// 	}, "--test")
// }

// func TestStringFlag(t *testing.T) {
// 	type cmdStringFlag struct {
// 		Test string `long:"value"`
// 	}
// 	assertSuccess[cmdStringFlag](t, func(x cmdStringFlag) {
// 		require.Equal(t, "value", x.Test)
// 	}, "--value", "value")
// }

// type cmdStringDefaultFlag struct {
// 	Test string `long:"value" default:"uuh"`
// }

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

// func TestFlagsInPositionals(t *testing.T) {
// 	type cmdFlagsInPositionals struct {
// 		Verbose bool     `short:"v"`
// 		Args    []string `positional-args:"yes"`
// 	}
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
//	assertFail[cmd](t, `the required argument "Rest (at least 1 argument)" was not provided`)
// }

// func TestRest1(t *testing.T) {
// 	type cmdRest1 struct {
// 		Verbose    bool `short:"v"`
// 		Positional struct {
// 			Rest []string
// 		} `positional-args:"yes"`
// 	}
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
// 	assertFail[cmd](t, `the required argument "Rest (at least 2 arguments, but got only 1)" was not provided`, "rest1")
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
// 	require.Fail(t, `the required argument "Rest (at most 2 arguments, but got 3)" was not provided`, "rest1", "rest2", "rest3")
// }

// func TestShortTooLong(t *testing.T) {
// 	type opts struct {
// 		Value bool `short:"vv"`
// 	}
// 	assertFail[opts](t, `parse flag "Value": scan short name: must be single character, but was "vv"`)
// }
