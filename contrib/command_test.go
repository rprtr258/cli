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

// type testCommand struct {
// 	G        bool `short:"g"`
// 	Executed bool
// 	EArgs    []string
// }

// func (c *testCommand) Execute(args []string) error {
// 	c.Executed = true
// 	c.EArgs = args
// 	return nil
// }

// func TestCommandExecute(t *testing.T) {
// 	opts := struct {
// 		Value   bool        `short:"v"`
// 		Command testCommand `command:"cmd"`
// 	}{}

// 	assertSuccess(t, &opts, "-v", "cmd", "-g", "a", "b")
// 	assert.True(t, opts.Value)
// 	assert.True(t, opts.Command.Executed)
// 	assert.True(t, opts.Command.G)
// 	assert.Equal(t, []string{"a", "b"}, opts.Command.EArgs)
// }

// func TestCommandClosest(t *testing.T) {
// 	opts := struct {
// 		Value bool     `short:"v"`
// 		Cmd1  struct{} `command:"remove"`
// 		Cmd2  struct{} `command:"add"`
// 	}{}

// 	assertFail(t, &Error{
// 		Kind:    ErrUnknownCommand,
// 		Message: "Unknown command `addd', did you mean `add'?",
// 	}, &opts, "-v", "addd")
// }

// func TestCommandAdd(t *testing.T) {
// 	opts := struct {
// 		Value bool `short:"v"`
// 		Cmd   struct {
// 			G    bool     `short:"g"`
// 			Args []string `args`
// 		} `command:"cmd"`
// 	}{}

// 	err := Run(&opts, "-v", "cmd", "-g", "rest")
// 	assert.NoError(t, err)
// 	assert.Equal(t, []string{"rest"}, opts.Cmd.Args)
// 	assert.True(t, opts.Value)
// 	assert.True(t, opts.Cmd.G)
// }

// func TestCommandNestedInline(t *testing.T) {
// 	opts := struct {
// 		Value   bool `short:"v"`
// 		Command struct {
// 			G      bool `short:"g"`
// 			Nested struct {
// 				N    string   `long:"n"`
// 				Args []string `args`
// 			} `command:"nested"`
// 		} `command:"cmd"`
// 	}{}

// 	assertSuccess(t, &opts, "-v", "cmd", "-g", "nested", "--n", "n", "rest")
// 	assert.Equal(t, []string{"rest"}, opts.Command.Nested.Args)
// 	assert.True(t, opts.Value)
// 	assert.True(t, opts.Command.G)
// 	assert.Equal(t, "n", opts.Command.Nested.N)
// }

// func TestRequiredOnCommand(t *testing.T) {
// 	type Opts struct {
// 		Value   bool `short:"v" required:"true"`
// 		Command struct {
// 			G bool `short:"g"`
// 		} `command:"cmd"`
// 	}

// 	assertFail[Opts](t, newErrorf(ErrRequired, "the required flag `%cv' was not specified", _defaultOptDelimiterShort), "cmd")
// }

// func TestRequiredAllOnCommand(t *testing.T) {
// 	opts := struct {
// 		Value   bool `short:"v" required:"true"`
// 		Missing bool `long:"missing" required:"true"`
// 		Command struct {
// 			G bool `short:"g"`
// 		} `command:"cmd"`
// 	}{}

// 	assertFail(t, &Error{
// 		Kind:    ErrRequired,
// 		Message: fmt.Sprintf("the required flags `%smissing' and `%cv' were not specified", _defaultOptDelimiterLong, _defaultOptDelimiterShort),
// 	}, &opts, "cmd")
// }

// func TestDefaultOnCommand(t *testing.T) {
// 	opts := struct {
// 		Command struct {
// 			G string `short:"g" default:"value"`
// 		} `command:"cmd"`
// 	}{}

// 	assertSuccess(t, &opts, "cmd")
// 	assert.Equal(t, "value", opts.Command.G)
// }

// func TestAfterNonCommand(t *testing.T) {
// 	opts := struct {
// 		Value bool     `short:"v"`
// 		Cmd1  struct{} `command:"remove"`
// 		Cmd2  struct{} `command:"add"`
// 	}{}

// 	assertFail(t, &Error{
// 		Kind:    ErrUnknownCommand,
// 		Message: "Unknown command `nocmd'. Please specify one command of: remove or add",
// 	}, &opts, "nocmd", "remove")
// }

// func TestSubcommandsOptional(t *testing.T) {
// 	opts := struct {
// 		Value bool     `short:"v"`
// 		Cmd1  struct{} `command:"remove"`
// 		Cmd2  struct{} `command:"add"`
// 	}{}

// 	err := Run(&opts, "-v")
// 	assert.NoError(t, err)
// 	assert.True(t, opts.Value)
// }

// func TestSubcommandsOptionalAfterNonCommand(t *testing.T) {
// 	opts := struct {
// 		Value bool     `short:"v"`
// 		Cmd1  struct{} `command:"remove"`
// 		Cmd2  struct{} `command:"add"`
// 	}{}

// 	err := Run(&opts, "nocmd", "remove")
// 	assert.NoError(t, err)
// 	// assert.Equal(t, []string{"nocmd", "remove"}, retargs)
// }

// func TestCommandAlias(t *testing.T) {
// 	opts := struct {
// 		Command struct {
// 			G string `short:"g" default:"value"`
// 		} `command:"cmd" alias:"cm"`
// 	}{}

// 	assertSuccess(t, &opts, "cm")
// 	assert.Equal(t, "value", opts.Command.G)
// }

// type fooCmd struct {
// 	Flag bool `short:"f"`
// 	args []string
// }

// func (foo *fooCmd) Execute(s []string) error {
// 	foo.args = s
// 	return nil
// }

// func TestCommandPassAfterNonOption(t *testing.T) {
// 	opts := struct {
// 		Value bool   `short:"v"`
// 		Foo   fooCmd `command:"foo"`
// 	}{}
// 	err := Run(&opts, "-v", "foo", "-f", "bar", "-v", "-g")
// 	assert.NoError(t, err)
// 	assert.True(t, opts.Value)
// 	assert.True(t, opts.Foo.Flag)
// 	assert.Equal(t, []string{"bar", "-v", "-g"}, opts.Foo.args)
// }

// type barCmd struct {
// 	fooCmd
// 	Positional struct {
// 		Args []string
// 	} `positional-args:"yes"`
// }

// func TestCommandPassAfterNonOptionWithPositional(t *testing.T) {
// 	opts := struct {
// 		Value bool   `short:"v"`
// 		Bar   barCmd `command:"bar"`
// 	}{}
// 	err := Run(&opts, "-v", "bar", "-f", "baz", "-v", "-g")
// 	assert.NoError(t, err)
// 	assert.True(t, opts.Value)
// 	assert.True(t, opts.Bar.Flag)
// 	assert.Equal(t, []string{}, opts.Bar.args)
// 	assert.Equal(t, []string{"baz", "-v", "-g"}, opts.Bar.Positional.Args)
// }
