package flags

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCommandInline(t *testing.T) {
	opts := struct {
		Value   bool `short:"v"`
		Command struct {
			G bool `short:"g"`
		} `command:"cmd"`
	}{}

	p := assertParserSuccess(t, &opts, "-v", "cmd", "-g")
	assert.NotNil(t, p.Command.Active)
	assert.True(t, opts.Value)
	assert.True(t, opts.Command.G)
	assert.Equal(t, p.Command.find("cmd"), p.Command.Active)
}

func TestCommandInlineMulti(t *testing.T) {
	opts := struct {
		Value bool     `short:"v"`
		C1    struct{} `command:"c1"`
		C2    struct {
			G bool `short:"g"`
		} `command:"c2"`
	}{}

	p := assertParserSuccess(t, &opts, "-v", "c2", "-g")
	assert.NotNil(t, p.Command.Active)
	assert.True(t, opts.Value)
	assert.True(t, opts.C2.G)
	assert.NotNil(t, p.Command.find("c1"))

	c2 := p.Command.find("c2")
	assert.NotNil(t, c2)
	assert.Equal(t, c2, p.Command.Active)
}

func TestCommandFlagOrder1(t *testing.T) {
	opts := struct {
		Value   bool `short:"v"`
		Command struct {
			G bool `short:"g"`
		} `command:"cmd"`
	}{}

	assertParseFail(t, &Error{
		Kind:    ErrUnknownFlag,
		Message: "unknown flag `g'",
	}, &opts, "-v", "-g", "cmd")
}

func TestCommandFlagOrder2(t *testing.T) {
	opts := struct {
		Value   bool `short:"v"`
		Command struct {
			G bool `short:"g"`
		} `command:"cmd"`
	}{}

	assertParserSuccess(t, &opts, "cmd", "-v", "-g")
	assert.True(t, opts.Value)
	assert.True(t, opts.Command.G)
}

func TestCommandFlagOrderSub(t *testing.T) {
	opts := struct {
		Value   bool `short:"v"`
		Command struct {
			G          bool `short:"g"`
			SubCommand struct {
				B bool `short:"b"`
			} `command:"sub"`
		} `command:"cmd"`
	}{}

	assertParserSuccess(t, &opts, "cmd", "sub", "-v", "-g", "-b")
	assert.True(t, opts.Value)
	assert.True(t, opts.Command.G)
	assert.True(t, opts.Command.SubCommand.B)
}

func TestCommandFlagOverride1(t *testing.T) {
	opts := struct {
		Value   bool `short:"v"`
		Command struct {
			Value bool `short:"v"`
		} `command:"cmd"`
	}{}

	assertParserSuccess(t, &opts, "-v", "cmd")
	assert.True(t, opts.Value)
	assert.False(t, opts.Command.Value)
}

func TestCommandFlagOverride2(t *testing.T) {
	opts := struct {
		Value   bool `short:"v"`
		Command struct {
			Value bool `short:"v"`
		} `command:"cmd"`
	}{}

	assertParserSuccess(t, &opts, "cmd", "-v")
	assert.False(t, opts.Value)
	assert.True(t, opts.Command.Value)
}

func TestCommandFlagOverrideSub(t *testing.T) {
	opts := struct {
		Value   bool `short:"v"`
		Command struct {
			Value      bool `short:"v"`
			SubCommand struct {
				Value bool `short:"v"`
			} `command:"sub"`
		} `command:"cmd"`
	}{}

	assertParserSuccess(t, &opts, "cmd", "sub", "-v")
	assert.False(t, opts.Value)
	assert.False(t, opts.Command.Value)
	assert.True(t, opts.Command.SubCommand.Value)
}

func TestCommandFlagOverrideSub2(t *testing.T) {
	opts := struct {
		Value   bool `short:"v"`
		Command struct {
			Value      bool `short:"v"`
			SubCommand struct {
				G bool `short:"g"`
			} `command:"sub"`
		} `command:"cmd"`
	}{}

	assertParserSuccess(t, &opts, "cmd", "sub", "-v")
	assert.False(t, opts.Value)
	assert.True(t, opts.Command.Value)
}

func TestCommandEstimate(t *testing.T) {
	opts := struct {
		Value bool     `short:"v"`
		Cmd1  struct{} `command:"remove"`
		Cmd2  struct{} `command:"add"`
	}{}

	p := NewParser(&opts)
	err := p.Run()
	assert.Equal(t, &Error{
		Kind:    ErrCommandRequired,
		Message: "Please specify one command of: remove or add",
	}, err)
}

func TestCommandEstimate2(t *testing.T) {
	opts := struct {
		Value bool     `short:"v"`
		Cmd1  struct{} `command:"remove"`
		Cmd2  struct{} `command:"add"`
	}{}

	p := NewParser(&opts)
	err := p.Run("rmive")
	assert.Equal(t, &Error{
		Kind:    ErrUnknownCommand,
		Message: "Unknown command `rmive', did you mean `remove'?",
	}, err)
}

type testCommand struct {
	G        bool `short:"g"`
	Executed bool
	EArgs    []string
}

func (c *testCommand) Execute(args []string) error {
	c.Executed = true
	c.EArgs = args
	return nil
}

func TestCommandExecute(t *testing.T) {
	opts := struct {
		Value   bool        `short:"v"`
		Command testCommand `command:"cmd"`
	}{}

	assertParserSuccess(t, &opts, "-v", "cmd", "-g", "a", "b")
	assert.True(t, opts.Value)
	assert.True(t, opts.Command.Executed)
	assert.True(t, opts.Command.G)
	assert.Equal(t, []string{"a", "b"}, opts.Command.EArgs)
}

func TestCommandClosest(t *testing.T) {
	opts := struct {
		Value bool     `short:"v"`
		Cmd1  struct{} `command:"remove"`
		Cmd2  struct{} `command:"add"`
	}{}

	assertParseFail(t, &Error{
		Kind:    ErrUnknownCommand,
		Message: "Unknown command `addd', did you mean `add'?",
	}, &opts, "-v", "addd")
}

func TestCommandAdd(t *testing.T) {
	opts := struct {
		Value bool `short:"v"`
		Cmd   struct {
			G    bool     `short:"g"`
			Args []string `args`
		} `command:"cmd"`
	}{}

	err := NewParser(&opts).Run("-v", "cmd", "-g", "rest")
	assert.NoError(t, err)
	assert.Equal(t, []string{"rest"}, opts.Cmd.Args)
	assert.True(t, opts.Value)
	assert.True(t, opts.Cmd.G)
}

func TestCommandNestedInline(t *testing.T) {
	opts := struct {
		Value   bool `short:"v"`
		Command struct {
			G      bool `short:"g"`
			Nested struct {
				N    string   `long:"n"`
				Args []string `args`
			} `command:"nested"`
		} `command:"cmd"`
	}{}

	assertParserSuccess(t, &opts, "-v", "cmd", "-g", "nested", "--n", "n", "rest")
	assert.Equal(t, []string{"rest"}, opts.Command.Nested.Args)
	assert.True(t, opts.Value)
	assert.True(t, opts.Command.G)
	assert.Equal(t, "n", opts.Command.Nested.N)
}

func TestRequiredOnCommand(t *testing.T) {
	opts := struct {
		Value   bool `short:"v" required:"true"`
		Command struct {
			G bool `short:"g"`
		} `command:"cmd"`
	}{}

	assertParseFail(t, &Error{
		Kind:    ErrRequired,
		Message: fmt.Sprintf("the required flag `%cv' was not specified", _defaultOptDelimiterShort),
	}, &opts, "cmd")
}

func TestRequiredAllOnCommand(t *testing.T) {
	opts := struct {
		Value   bool `short:"v" required:"true"`
		Missing bool `long:"missing" required:"true"`
		Command struct {
			G bool `short:"g"`
		} `command:"cmd"`
	}{}

	assertParseFail(t, &Error{
		Kind:    ErrRequired,
		Message: fmt.Sprintf("the required flags `%smissing' and `%cv' were not specified", _defaultOptDelimiterLong, _defaultOptDelimiterShort),
	}, &opts, "cmd")
}

func TestDefaultOnCommand(t *testing.T) {
	opts := struct {
		Command struct {
			G string `short:"g" default:"value"`
		} `command:"cmd"`
	}{}

	assertParserSuccess(t, &opts, "cmd")
	assert.Equal(t, "value", opts.Command.G)
}

func TestAfterNonCommand(t *testing.T) {
	opts := struct {
		Value bool     `short:"v"`
		Cmd1  struct{} `command:"remove"`
		Cmd2  struct{} `command:"add"`
	}{}

	assertParseFail(t, &Error{
		Kind:    ErrUnknownCommand,
		Message: "Unknown command `nocmd'. Please specify one command of: remove or add",
	}, &opts, "nocmd", "remove")
}

func TestSubcommandsOptional(t *testing.T) {
	opts := struct {
		Value bool     `short:"v"`
		Cmd1  struct{} `command:"remove"`
		Cmd2  struct{} `command:"add"`
	}{}

	p := NewParser(&opts)
	p.Command.SubcommandsOptional = true

	err := p.Run("-v")
	assert.NoError(t, err)
	assert.True(t, opts.Value)
}

func TestSubcommandsOptionalAfterNonCommand(t *testing.T) {
	opts := struct {
		Value bool     `short:"v"`
		Cmd1  struct{} `command:"remove"`
		Cmd2  struct{} `command:"add"`
	}{}

	p := NewParser(&opts)
	p.Command.SubcommandsOptional = true

	err := p.Run("nocmd", "remove")
	assert.NoError(t, err)
	// assert.Equal(t, []string{"nocmd", "remove"}, retargs)
}

func TestCommandAlias(t *testing.T) {
	opts := struct {
		Command struct {
			G string `short:"g" default:"value"`
		} `command:"cmd" alias:"cm"`
	}{}

	assertParserSuccess(t, &opts, "cm")
	assert.Equal(t, "value", opts.Command.G)
}

type fooCmd struct {
	Flag bool `short:"f"`
	args []string
}

func (foo *fooCmd) Execute(s []string) error {
	foo.args = s
	return nil
}

func TestCommandPassAfterNonOption(t *testing.T) {
	opts := struct {
		Value bool   `short:"v"`
		Foo   fooCmd `command:"foo"`
	}{}
	err := NewParser(&opts).Run("-v", "foo", "-f", "bar", "-v", "-g")
	assert.NoError(t, err)
	assert.True(t, opts.Value)
	assert.True(t, opts.Foo.Flag)
	assert.Equal(t, []string{"bar", "-v", "-g"}, opts.Foo.args)
}

type barCmd struct {
	fooCmd
	Positional struct {
		Args []string
	} `positional-args:"yes"`
}

func TestCommandPassAfterNonOptionWithPositional(t *testing.T) {
	opts := struct {
		Value bool   `short:"v"`
		Bar   barCmd `command:"bar"`
	}{}
	p := NewParser(&opts)
	err := p.Run("-v", "bar", "-f", "baz", "-v", "-g")
	assert.NoError(t, err)
	assert.True(t, opts.Value)
	assert.True(t, opts.Bar.Flag)
	assert.Equal(t, []string{}, opts.Bar.args)
	assert.Equal(t, []string{"baz", "-v", "-g"}, opts.Bar.Positional.Args)
}
