package flags

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type optsSimpleSubcommand struct {
	testCommander[optsSimpleSubcommand]

	G bool `short:"g"`
}

type optsFirst struct {
	testCommander[optsFirst]

	Value   bool                 `short:"v"`
	Command optsSimpleSubcommand `command:"cmd"`
}

func TestRun_first_subcommand(t *testing.T) {
	assertSuccess[optsFirst](t,
		func(opts optsSimpleSubcommand) {
			assert.True(t, opts.G)
		},
		"x", "cmd", "-g",
	)
}

func TestRun_first_command(t *testing.T) {
	assertSuccess[optsFirst](t,
		func(opts optsFirst) {
			assert.True(t, opts.Value)
		},
		"x", "-v",
	)
}

func TestRun_first_commandExtraneousPositionals(t *testing.T) {
	assertFail[optsFirst](t,
		`extraneous positional arguments found: ["cmd", "-g"]`,
		"x", "-v", "cmd", "-g",
	)
}

func TestNewCommand_CommandInline(t *testing.T) {
	cmd, err := newCommand[optsFirst]()
	assert.NoError(t, err)

	// null to be able to compare the rest
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

type optsSecond struct {
	Value bool                 `short:"v"`
	C1    struct{}             `command:"c1"`
	C2    optsSimpleSubcommand `command:"c2"`
}

// func Test_second(t *testing.T) {
// 	assertSuccess[optsSecond](t, &opts, "-v", "c2", "-g")
// 	assert.True(t, opts.Value)
// 	assert.True(t, opts.C2.G)
// }

// func TestCommandFlagOrder1(t *testing.T) {
// 	opts := struct {
// 		Value   bool `short:"v"`
// 		Command struct {
// 			G bool `short:"g"`
// 		} `command:"cmd"`
// 	}{}

// 	assertFail(t, &Error{
// 		Kind:    ErrUnknownFlag,
// 		Message: "unknown flag `g'",
// 	}, &opts, "-v", "-g", "cmd")
// }

// func TestCommandFlagOrder2(t *testing.T) {
// 	opts := struct {
// 		Value   bool `short:"v"`
// 		Command struct {
// 			G bool `short:"g"`
// 		} `command:"cmd"`
// 	}{}

// 	assertSuccess(t, &opts, "cmd", "-v", "-g")
// 	assert.True(t, opts.Value)
// 	assert.True(t, opts.Command.G)
// }

// func TestCommandFlagOrderSub(t *testing.T) {
// 	opts := struct {
// 		Value   bool `short:"v"`
// 		Command struct {
// 			G          bool `short:"g"`
// 			SubCommand struct {
// 				B bool `short:"b"`
// 			} `command:"sub"`
// 		} `command:"cmd"`
// 	}{}

// 	assertSuccess(t, &opts, "cmd", "sub", "-v", "-g", "-b")
// 	assert.True(t, opts.Value)
// 	assert.True(t, opts.Command.G)
// 	assert.True(t, opts.Command.SubCommand.B)
// }

// func TestCommandFlagOverride1(t *testing.T) {
// 	opts := struct {
// 		Value   bool `short:"v"`
// 		Command struct {
// 			Value bool `short:"v"`
// 		} `command:"cmd"`
// 	}{}

// 	assertSuccess(t, &opts, "-v", "cmd")
// 	assert.True(t, opts.Value)
// 	assert.False(t, opts.Command.Value)
// }

// func TestCommandFlagOverride2(t *testing.T) {
// 	opts := struct {
// 		Value   bool `short:"v"`
// 		Command struct {
// 			Value bool `short:"v"`
// 		} `command:"cmd"`
// 	}{}

// 	assertSuccess(t, &opts, "cmd", "-v")
// 	assert.False(t, opts.Value)
// 	assert.True(t, opts.Command.Value)
// }

// func TestCommandFlagOverrideSub(t *testing.T) {
// 	opts := struct {
// 		Value   bool `short:"v"`
// 		Command struct {
// 			Value      bool `short:"v"`
// 			SubCommand struct {
// 				Value bool `short:"v"`
// 			} `command:"sub"`
// 		} `command:"cmd"`
// 	}{}

// 	assertSuccess(t, &opts, "cmd", "sub", "-v")
// 	assert.False(t, opts.Value)
// 	assert.False(t, opts.Command.Value)
// 	assert.True(t, opts.Command.SubCommand.Value)
// }

// func TestCommandFlagOverrideSub2(t *testing.T) {
// 	opts := struct {
// 		Value   bool `short:"v"`
// 		Command struct {
// 			Value      bool `short:"v"`
// 			SubCommand struct {
// 				G bool `short:"g"`
// 			} `command:"sub"`
// 		} `command:"cmd"`
// 	}{}

// 	assertSuccess(t, &opts, "cmd", "sub", "-v")
// 	assert.False(t, opts.Value)
// 	assert.True(t, opts.Command.Value)
// }

// func TestCommandEstimate(t *testing.T) {
// 	opts := struct {
// 		Value bool     `short:"v"`
// 		Cmd1  struct{} `command:"remove"`
// 		Cmd2  struct{} `command:"add"`
// 	}{}

// 	err := Run(&opts)
// 	assert.Equal(t, &Error{
// 		Kind:    ErrCommandRequired,
// 		Message: "Please specify one command of: remove or add",
// 	}, err)
// }

// func TestCommandEstimate2(t *testing.T) {
// 	opts := struct {
// 		Value bool     `short:"v"`
// 		Cmd1  struct{} `command:"remove"`
// 		Cmd2  struct{} `command:"add"`
// 	}{}

// 	err := Run(&opts, "rmive")
// 	assert.Equal(t, &Error{
// 		Kind:    ErrUnknownCommand,
// 		Message: "Unknown command `rmive', did you mean `remove'?",
// 	}, err)
// }

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
