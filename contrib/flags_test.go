package flags

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

type optsSimpleSubcommand struct {
	G bool `short:"g"`
}

func (x optsSimpleSubcommand) Execute(ctx context.Context) error { return fromCtx(ctx)(x) }

type optsFirst struct {
	Value   bool                 `short:"v"`
	Command optsSimpleSubcommand `command:"cmd"`
}

func (x optsFirst) Execute(ctx context.Context) error { return fromCtx(ctx)(x) }

func TestRun_first_subcommand(t *testing.T) {
	assertSuccess[optsFirst](t,
		func(opts optsSimpleSubcommand) {
			assert.True(t, opts.G)
		},
		"cmd", "-g",
	)
}

func TestRun_first_command(t *testing.T) {
	assertSuccess[optsFirst](t,
		func(opts optsFirst) {
			assert.True(t, opts.Value)
		},
		"-v",
	)
}

func TestRun_first_commandExtraneousPositionals(t *testing.T) {
	assertFail[optsFirst](t,
		`extraneous positional arguments found: ["cmd", "-g"]`,
		"-v", "cmd", "-g",
	)
}

func TestRun_extraFlag(t *testing.T) {
	assertFail[optsFirst](t,
		`extraneous positional arguments found: ["-g", "cmd"]`,
		"-v", "-g", "cmd",
	)
}

func TestNewCommand_CommandInline(t *testing.T) {
	cmd, err := newCommand[optsFirst]()
	assert.NoError(t, err)

	// zero out function pointers to be able to compare the rest
	for i := range cmd.flags {
		cmd.flags[i].valuesetter = nil
	}
	cmd.posArgsRest.valuesetter = nil
	cmd.execute = nil
	for i, subcmd := range cmd.subcommands {
		for i := range subcmd.flags {
			subcmd.flags[i].valuesetter = nil
		}
		cmd.subcommands[i].posArgsRest.valuesetter = nil
		subcmd.execute = nil
	}

	want := &Command{
		subcommands: map[string]*Command{
			"cmd": {
				subcommands: map[string]*Command{},
				flags: []Flag{
					{
						shortname: 'g',
						longname:  "",
						isbool:    true,
						minCount:  0,
						maxCount:  1,
					},
				},
				posArgs:     []PositionalArg{},
				posArgsRest: PositionalArgsRest{},
			},
		},
		flags: []Flag{
			{
				shortname: 'v',
				longname:  "",
				isbool:    true,
				minCount:  0,
				maxCount:  1,
			},
		},
		posArgs:     []PositionalArg{},
		posArgsRest: PositionalArgsRest{},
	}
	assert.Equal(t, want, cmd)
}

func TestRun_leafHasNoExecute(t *testing.T) {
	type optsSecond struct {
		C1 struct{} `command:"c1"`
	}

	assertFail[optsSecond](t,
		`parse subcommand c1: command struct {} is leaf command, but no Execute function defined`,
		"-v",
	)
}

func TestRun_subcommand_extraArgs(t *testing.T) {
	assertFail[optsFirst](t,
		`extraneous positional arguments found: ["-v", "-g"]`,
		"cmd", "-v", "-g",
	)
}

func TestRun_subcommand_extraArg(t *testing.T) {
	assertFail[optsFirst](t,
		`extraneous positional arguments found: ["-v"]`,
		"cmd", "-g", "-v",
	)
}

type optsThirdCmdSub struct {
	B bool `short:"b"`
}

func (x optsThirdCmdSub) Execute(ctx context.Context) error { return fromCtx(ctx)(x) }

type optsThirdCmd struct {
	G          bool            `short:"g"`
	SubCommand optsThirdCmdSub `command:"sub"`
}

func (x optsThirdCmd) Execute(ctx context.Context) error { return fromCtx(ctx)(x) }

type optsThird struct {
	Value   bool         `short:"v"`
	Command optsThirdCmd `command:"cmd"`
}

func (x optsThird) Execute(ctx context.Context) error { return fromCtx(ctx)(x) }

func Test_third_subsubcommandExtraArgs(t *testing.T) {
	assertFail[optsThird](t,
		`extraneous positional arguments found: ["-v", "-g"]`,
		"cmd", "sub", "-b", "-v", "-g",
	)
}

type optsFourthSub struct {
	Value bool `short:"v"`
}

func (x optsFourthSub) Execute(ctx context.Context) error { return fromCtx(ctx)(x) }

type optsFourthCmd struct {
	Value   bool          `short:"v"`
	Command optsFourthSub `command:"sub"`
}

func (x optsFourthCmd) Execute(ctx context.Context) error { return fromCtx(ctx)(x) }

func Test_fourth_flagThenPresumablySubcommand(t *testing.T) {
	assertFail[optsFourthCmd](t,
		`extraneous positional arguments found: ["sub"]`,
		"-v", "sub",
	)
}

func Test_fourth_flag(t *testing.T) {
	assertSuccess[optsFourthCmd](t,
		func(x optsFourthCmd) {
			assert.True(t, x.Value)
		},
		"-v",
	)
}

type optsFourth struct {
	Value   bool          `short:"v"`
	Command optsFourthCmd `command:"cmd"`
}

func (x optsFourth) Execute(ctx context.Context) error { return fromCtx(ctx)(x) }

func Test_fourth_altSubcommandFlag(t *testing.T) {
	assertSuccess[optsFourth](t,
		func(x optsFourthSub) {
			assert.True(t, x.Value)
		},
		"cmd", "sub", "-v",
	)
}

func TestCommandEstimate(t *testing.T) {
	type optsSix struct {
		Cmd1 optsSimpleSubcommand `command:"remove"`
		Cmd2 optsSimpleSubcommand `command:"add"`
	}

	assertFail[optsSix](t,
		"Please specify one command of: add or remove",
	)
}

func TestCommandEstimate2(t *testing.T) {
	type optsSix struct {
		Cmd1 optsSimpleSubcommand `command:"remove"`
		Cmd2 optsSimpleSubcommand `command:"add"`
	}

	assertFail[optsSix](t,
		"Unknown command `rmive', did you mean `remove'?",
		"rmive",
	)
}

type optsSeven struct {
	G    bool     `short:"g"`
	Args []string `positional-args:"yes"`
}

func (x optsSeven) Execute(ctx context.Context) error { return fromCtx(ctx)(x) }

func Test_subcommands_posArgs(t *testing.T) {
	type Opts struct {
		Command optsSeven `command:"cmd"`
	}

	assertSuccess[Opts](t,
		func(x optsSeven) {
			assert.True(t, x.G)
			assert.Equal(t, []string{"a", "b"}, x.Args)
		},
		"cmd", "-g", "a", "b",
	)
}

type optsParse struct {
	Verbose bool     `short:"v" long:"verbose" description:"Show verbose debug information"`
	Name    string   `short:"n" long:"name" description:"A name"`
	Args    []string `positional-args:"yes"`
}

func (x optsParse) Execute(ctx context.Context) error { return fromCtx(ctx)(x) }

func TestParse(t *testing.T) {
	assertSuccess[optsParse](t,
		func(x optsParse) {
			assert.Equal(t, optsParse{
				Verbose: true,
				Name:    "aboba",
				Args:    []string{"xdd"},
			}, x)
		},
		"--name", "aboba", "--verbose", "xdd",
	)
}

// func TestGroupInline(t *testing.T) {
// 	opts := struct {
// 		Value bool `short:"v"`
// 		G     bool `short:"g"`
// 	}{}

// 	assertNoError(t, &opts, "-v", "-g")
// 	assert.True(t, opts.Value)
// 	assert.True(t, opts.G)
// }

// func TestGroupAdd(t *testing.T) {
// 	opts := struct {
// 		Value bool     `short:"v"`
// 		G     bool     `short:"g"`
// 		Args  []string `args`
// 	}{}

// 	err := Run(&opts, "-v", "-g", "rest")
// 	assert.NoError(t, err)

// 	assert.Equal(t, []string{"rest"}, opts.Args)
// 	assert.True(t, opts.Value)
// 	assert.True(t, opts.G)
// }

// func TestGroupNestedInline(t *testing.T) {
// 	opts := struct {
// 		Value bool     `short:"v"`
// 		G     bool     `short:"g"`
// 		N     string   `long:"n"`
// 		Args  []string `args`
// 	}{}

// 	assertNoError(t, &opts, "-v", "-g", "--n", "n", "rest")
// 	assert.Equal(t, []string{"rest"}, opts.Args)
// 	assert.True(t, opts.Value)
// 	assert.True(t, opts.G)
// 	assert.Equal(t, opts.N, "n")
// }

// func TestGroupNestedInlineNamespace(t *testing.T) {
// 	opts := struct {
// 		Opt       string   `long:"opt"`
// 		SipOpt    string   `long:"sip.opt"`
// 		SipSapOpt string   `long:"sip.sap.opt"`
// 		Args      []string `args`
// 	}{}

// 	assertNoError(t, &opts, "--opt", "a", "--sip.opt", "b", "--sip.sap.opt", "c", "rest")
// 	assert.Equal(t, []string{"rest"}, opts.Args)
// 	assert.Equal(t, opts.Opt, "a")
// 	assert.Equal(t, opts.SipOpt, "b")
// 	assert.Equal(t, opts.SipSapOpt, "c")
// }

// func TestDuplicateShortFlags(t *testing.T) {
// 	var opts struct {
// 		Verbose   []bool   `short:"v" long:"verbose" description:"Show verbose debug information"`
// 		Variables []string `short:"v" long:"variable" description:"Set a variable value."`
// 	}

// 	err := Run(&opts,
// 		"--verbose",
// 		"-v", "123",
// 		"-v", "456",
// 	)
// 	assert.Error(t, err)
// 	err2 := err.(*Error)
// 	assert.Equal(t, ErrDuplicatedFlag, err2.Kind)
// }

// func TestDuplicateLongFlags(t *testing.T) {
// 	var opts struct {
// 		Test1 []bool   `short:"a" long:"testing" description:"Test 1"`
// 		Test2 []string `short:"b" long:"testing" description:"Test 2."`
// 	}

// 	err := Run(&opts, "--testing")
// 	assert.Error(t, err)
// 	err2 := err.(*Error)
// 	assert.Equal(t, ErrDuplicatedFlag, err2.Kind)
// }

// func TestAddOptionNonOptional(t *testing.T) {
// 	var opts struct {
// 		Test bool `long:"test"`
// 	}

// 	err := Run(&opts, "--test")
// 	assert.NoError(t, err)
// 	assert.True(t, opts.Test)
// }

// func TestLong(t *testing.T) {
// 	t.Run("Long", func(t *testing.T) {
// 		opts := struct {
// 			Value bool `long:"value"`
// 		}{}

// 		assertNoError(t, &opts, "--value")
// 		assert.True(t, opts.Value)
// 	})

// 	t.Run("LongArg", func(t *testing.T) {
// 		opts := struct {
// 			Value string `long:"value"`
// 		}{}

// 		assertNoError(t, &opts, "--value", "value")
// 		assert.Equal(t, "value", opts.Value)
// 	})

// 	t.Run("LongArgEqual", func(t *testing.T) {
// 		opts := struct {
// 			Value string `long:"value"`
// 		}{}

// 		assertNoError(t, &opts, "--value=value")
// 		assert.Equal(t, "value", opts.Value)
// 	})

// 	t.Run("LongDefault", func(t *testing.T) {
// 		opts := struct {
// 			Value string `long:"value" default:"value"`
// 		}{}

// 		assertNoError(t, &opts)
// 		assert.Equal(t, "value", opts.Value)
// 	})

// 	t.Run("LongOptional", func(t *testing.T) {
// 		opts := struct {
// 			Value string `long:"value" optional:"yes" optional-value:"value"`
// 		}{}

// 		assertNoError(t, &opts, "--value")
// 		assert.Equal(t, "value", opts.Value)
// 	})

// 	t.Run("LongOptionalArg", func(t *testing.T) {
// 		opts := struct {
// 			Value string `long:"value" optional:"yes" optional-value:"value"`
// 		}{}

// 		assertNoError(t, &opts, "--value", "no")
// 		// assert.Equal(t, []string{"no"}, ret)
// 		assert.Equal(t, "value", opts.Value)
// 	})

// 	t.Run("LongOptionalArgEqual", func(t *testing.T) {
// 		opts := struct {
// 			Value string `long:"value" optional:"yes" optional-value:"value"`
// 		}{}

// 		assertNoError(t, &opts, "--value=value", "no")
// 		// assert.Equal(t, []string{"no"}, ret)
// 		assert.Equal(t, "value", opts.Value)
// 	})
// }

// func TestUnknown(t *testing.T) {
// 	t.Run("UnknownFlags", func(t *testing.T) {
// 		opts := struct {
// 			Verbose []bool `short:"v" long:"verbose" description:"Verbose output"`
// 		}{}

// 		err := Run(&opts, "-f")
// 		assert.Error(t, err)
// 	})

// 	t.Run("IgnoreUnknownFlags", func(t *testing.T) {
// 		opts := struct {
// 			Verbose []bool   `short:"v" long:"verbose" description:"Verbose output"`
// 			Args    []string `args`
// 		}{}

// 		err := Run(&opts,
// 			"hello",
// 			"world",
// 			"-v",
// 			"--foo=bar",
// 			"--verbose",
// 			"-f",
// 		)
// 		assert.NoError(t, err)
// 		assert.Len(t, opts.Verbose, 0)
// 		assert.Equal(t, []string{
// 			"hello",
// 			"world",
// 			"-v",
// 			"--foo=bar",
// 			"--verbose",
// 			"-f",
// 		}, opts.Args)
// 	})
// }

// func TestPositional(t *testing.T) {
// 	opts := struct {
// 		Value      bool `short:"v"`
// 		Positional struct {
// 			Command  int
// 			Filename string
// 			Rest     []string
// 		} `positional-args:"yes" required:"yes"`
// 	}{}

// 	err := Run(&opts, "10", "arg_test.go", "a", "b")
// 	assert.NoError(t, err)
// 	assert.Equal(t, 10, opts.Positional.Command)
// 	assert.Equal(t, "arg_test.go", opts.Positional.Filename)
// 	assert.Equal(t, []string{"a", "b"}, opts.Positional.Rest)
// }

// func TestPositionalRequired(t *testing.T) {
// 	opts := struct {
// 		Value      bool `short:"v"`
// 		Positional struct {
// 			Command  int
// 			Filename string
// 			Rest     []string
// 		} `positional-args:"yes" required:"yes"`
// 	}{}

// 	err := Run(&opts, "10")
// 	assert.Equal(t, &Error{
// 		Kind:    ErrRequired,
// 		Message: "the required argument `Filename` was not provided",
// 	}, err)
// }

// func TestPositionalRequiredRest1Fail(t *testing.T) {
// 	opts := struct {
// 		Value      bool `short:"v"`
// 		Positional struct {
// 			Rest []string `required:"yes"`
// 		} `positional-args:"yes"`
// 	}{}

// 	err := Run(&opts)
// 	assert.Equal(t, &Error{
// 		Kind:    ErrRequired,
// 		Message: "the required argument `Rest (at least 1 argument)` was not provided",
// 	}, err)
// }

// func TestPositionalRequiredRest1Pass(t *testing.T) {
// 	opts := struct {
// 		Value      bool `short:"v"`
// 		Positional struct {
// 			Rest []string `required:"yes"`
// 		} `positional-args:"yes"`
// 	}{}

// 	err := Run(&opts, "rest1")
// 	assert.NoError(t, err)
// 	assert.Equal(t, []string{"rest1"}, opts.Positional.Rest)
// }

// func TestPositionalRequiredRest2Fail(t *testing.T) {
// 	opts := struct {
// 		Value      bool `short:"v"`
// 		Positional struct {
// 			Rest []string `required:"2"`
// 		} `positional-args:"yes"`
// 	}{}

// 	err := Run(&opts, "rest1")
// 	assert.Equal(t, &Error{
// 		Kind:    ErrRequired,
// 		Message: "the required argument `Rest (at least 2 arguments, but got only 1)` was not provided",
// 	}, err)
// }

// func TestPositionalRequiredRest2Pass(t *testing.T) {
// 	opts := struct {
// 		Value      bool `short:"v"`
// 		Positional struct {
// 			Rest []string `required:"2"`
// 		} `positional-args:"yes"`
// 	}{}

// 	err := Run(&opts, "rest1", "rest2", "rest3")
// 	assert.NoError(t, err)
// 	assert.Equal(t, []string{"rest1", "rest2", "rest3"}, opts.Positional.Rest)
// }

// func TestPositionalRequiredRestRangeFail(t *testing.T) {
// 	opts := struct {
// 		Value      bool `short:"v"`
// 		Positional struct {
// 			Rest []string `required:"1-2"`
// 		} `positional-args:"yes"`
// 	}{}

// 	err := Run(&opts, "rest1", "rest2", "rest3")
// 	assert.Equal(t, &Error{
// 		Kind:    ErrRequired,
// 		Message: "the required argument `Rest (at most 2 arguments, but got 3)` was not provided",
// 	}, err)
// }

// func TestPositionalRequiredRestRangeEmptyFail(t *testing.T) {
// 	opts := struct {
// 		Value      bool `short:"v"`
// 		Positional struct {
// 			Rest []string `required:"0-0"`
// 		} `positional-args:"yes"`
// 	}{}

// 	err := Run(&opts, "some", "thing")
// 	assert.Equal(t, &Error{
// 		Kind:    ErrRequired,
// 		Message: "the required argument `Rest (zero arguments)` was not provided",
// 	}, err)
// }
