package cli

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCommandInline(t *testing.T) {
	var opts struct {
		Value bool `short:"v"`

		Command struct {
			G bool `short:"g"`
		} `command:"cmd"`
	}

	p, ret := useParser(t, &opts, "-v", "cmd", "-g")

	assert.Equal(t, ret, []string{})

	assert.NotNil(t, p.Active)

	assert.True(t, opts.Value)

	assert.True(t, opts.Command.G)

	assert.Equal(t, p.Active, p.Command.Find("cmd"))
}

func TestCommandInlineMulti(t *testing.T) {
	var opts struct {
		Value bool `short:"v"`

		C1 struct{} `command:"c1"`

		C2 struct {
			G bool `short:"g"`
		} `command:"c2"`
	}

	p, ret := useParser(t, &opts, "-v", "c2", "-g")

	assert.Equal(t, ret, []string{})

	assert.NotNil(t, p.Active)

	assert.True(t, opts.Value)

	assert.True(t, opts.C2.G)

	assert.NotNil(t, p.Command.Find("c1"))

	c2 := p.Command.Find("c2")
	assert.NotNil(t, c2)
	assert.Equal(t, p.Active, c2)
}

func TestCommandFlagOrder1(t *testing.T) {
	var opts struct {
		Value bool `short:"v"`

		Command struct {
			G bool `short:"g"`
		} `command:"cmd"`
	}

	assertParseFailed(t, ErrUnknownFlag, "unknown flag `g'", &opts, "-v", "-g", "cmd")
}

func TestCommandFlagOrder2(t *testing.T) {
	var opts struct {
		Value bool `short:"v"`

		Command struct {
			G bool `short:"g"`
		} `command:"cmd"`
	}

	useParsedArgs(t, &opts, "cmd", "-v", "-g")

	assert.True(t, opts.Value)

	assert.True(t, opts.Command.G)
}

func TestCommandFlagOrderSub(t *testing.T) {
	var opts struct {
		Value bool `short:"v"`

		Command struct {
			G bool `short:"g"`

			SubCommand struct {
				B bool `short:"b"`
			} `command:"sub"`
		} `command:"cmd"`
	}

	useParsedArgs(t, &opts, "cmd", "sub", "-v", "-g", "-b")

	assert.True(t, opts.Value)

	assert.True(t, opts.Command.G)

	assert.True(t, opts.Command.SubCommand.B)
}

func TestCommandFlagOverride1(t *testing.T) {
	var opts struct {
		Value bool `short:"v"`

		Command struct {
			Value bool `short:"v"`
		} `command:"cmd"`
	}

	useParsedArgs(t, &opts, "-v", "cmd")

	assert.True(t, opts.Value)

	if opts.Command.Value {
		t.Errorf("Expected Command.Value to be false")
	}
}

func TestCommandFlagOverride2(t *testing.T) {
	var opts struct {
		Value bool `short:"v"`

		Command struct {
			Value bool `short:"v"`
		} `command:"cmd"`
	}

	useParsedArgs(t, &opts, "cmd", "-v")

	if opts.Value {
		t.Errorf("Expected Value to be false")
	}

	assert.True(t, opts.Command.Value)
}

func TestCommandFlagOverrideSub(t *testing.T) {
	var opts struct {
		Value   bool `short:"v"`
		Command struct {
			Value      bool `short:"v"`
			SubCommand struct {
				Value bool `short:"v"`
			} `command:"sub"`
		} `command:"cmd"`
	}

	useParsedArgs(t, &opts, "cmd", "sub", "-v")
	if opts.Value {
		t.Errorf("Expected Value to be false")
	}
	if opts.Command.Value {
		t.Errorf("Expected Command.Value to be false")
	}

	assert.True(t, opts.Command.SubCommand.Value)
}

func TestCommandFlagOverrideSub2(t *testing.T) {
	var opts struct {
		Value   bool `short:"v"`
		Command struct {
			Value      bool `short:"v"`
			SubCommand struct {
				G bool `short:"g"`
			} `command:"sub"`
		} `command:"cmd"`
	}

	useParsedArgs(t, &opts, "cmd", "sub", "-v")
	if opts.Value {
		t.Errorf("Expected Value to be false")
	}

	assert.True(t, opts.Command.Value)
}

func TestCommandEstimate(t *testing.T) {
	var opts struct {
		Value bool     `short:"v"`
		Cmd1  struct{} `command:"remove"`
		Cmd2  struct{} `command:"add"`
	}

	_, err := NewParser(&opts, None).ParseArgs([]string{})
	assert.Equal(t, err, &Error{
		Type:    ErrCommandRequired,
		Message: "Please specify one command of: add or remove",
	})
}

func TestCommandEstimate2(t *testing.T) {
	var opts struct {
		Value bool     `short:"v"`
		Cmd1  struct{} `command:"remove"`
		Cmd2  struct{} `command:"add"`
	}

	_, err := NewParser(&opts, None).ParseArgs([]string{"rmive"})
	assert.Equal(t, err, &Error{
		Type:    ErrUnknownCommand,
		Message: "Unknown command `rmive', did you mean `remove'?",
	})
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
	var opts struct {
		Value bool `short:"v"`

		Command testCommand `command:"cmd"`
	}

	useParsedArgs(t, &opts, "-v", "cmd", "-g", "a", "b")

	assert.True(t, opts.Value)

	assert.True(t, opts.Command.Executed)

	assert.True(t, opts.Command.G)

	assert.Equal(t, opts.Command.EArgs, []string{"a", "b"})
}

func TestCommandClosest(t *testing.T) {
	var opts struct {
		Value bool `short:"v"`

		Cmd1 struct{} `command:"remove"`

		Cmd2 struct{} `command:"add"`
	}

	args := assertParseFailed(t, ErrUnknownCommand, "Unknown command `addd', did you mean `add'?", &opts, "-v", "addd")

	assert.Equal(t, args, []string{"addd"})
}

func TestCommandAdd(t *testing.T) {
	var opts struct {
		Value bool `short:"v"`
	}

	var cmd struct {
		G bool `short:"g"`
	}

	p := NewParser(&opts, Default)
	c := p.AddCommand("cmd", "", "", &cmd)

	ret, err := p.ParseArgs([]string{"-v", "cmd", "-g", "rest"})
	assert.NoError(t, err)

	assert.Equal(t, ret, []string{"rest"})

	assert.True(t, opts.Value)

	assert.True(t, cmd.G)

	assert.Equal(t, c, p.Command.Find("cmd"))

	assert.Equal(t, c, p.Commands()[0])

	assert.Equal(t, 'g', c.Options()[0].ShortName)
}

func TestCommandNestedInline(t *testing.T) {
	var opts struct {
		Value bool `short:"v"`

		Command struct {
			G bool `short:"g"`

			Nested struct {
				N string `long:"n"`
			} `command:"nested"`
		} `command:"cmd"`
	}

	p, ret := useParser(t, &opts, "-v", "cmd", "-g", "nested", "--n", "n", "rest")
	assert.Equal(t, ret, []string{"rest"})
	assert.True(t, opts.Value)
	assert.True(t, opts.Command.G)
	assert.Equal(t, opts.Command.Nested.N, "n")

	c := p.Command.Find("cmd")
	assert.NotNil(t, c)
	assert.Equal(t, p.Active, c)

	nested := c.Find("nested")
	assert.NotNil(t, nested)
	assert.Equal(t, c.Active, nested)
}

func TestRequiredOnCommand(t *testing.T) {
	var opts struct {
		Value bool `short:"v" required:"true"`

		Command struct {
			G bool `short:"g"`
		} `command:"cmd"`
	}

	assertParseFailed(t, ErrRequired, fmt.Sprintf("the required flag `%cv' was not specified", defaultShortOptDelimiter), &opts, "cmd")
}

func TestRequiredAllOnCommand(t *testing.T) {
	var opts struct {
		Value   bool `short:"v" required:"true"`
		Missing bool `long:"missing" required:"true"`

		Command struct {
			G bool `short:"g"`
		} `command:"cmd"`
	}

	assertParseFailed(t, ErrRequired, fmt.Sprintf("the required flags `%smissing' and `%cv' were not specified", defaultLongOptDelimiter, defaultShortOptDelimiter), &opts, "cmd")
}

func TestDefaultOnCommand(t *testing.T) {
	var opts struct {
		Command struct {
			G string `short:"g" default:"value"`
		} `command:"cmd"`
	}

	useParsedArgs(t, &opts, "cmd")

	assert.Equal(t, "value", opts.Command.G)
}

func TestAfterNonCommand(t *testing.T) {
	var opts struct {
		Value bool `short:"v"`

		Cmd1 struct{} `command:"remove"`

		Cmd2 struct{} `command:"add"`
	}

	assertParseFailed(t, ErrUnknownCommand, "Unknown command `nocmd'. Please specify one command of: add or remove", &opts, "nocmd", "remove")
}

func TestSubcommandsOptional(t *testing.T) {
	var opts struct {
		Value bool `short:"v"`

		Cmd1 struct{} `command:"remove"`

		Cmd2 struct{} `command:"add"`
	}

	p := NewParser(&opts, None)
	p.SubcommandsOptional = true

	_, err := p.ParseArgs([]string{"-v"})
	assert.NoError(t, err)

	assert.True(t, opts.Value)
}

func TestSubcommandsOptionalAfterNonCommand(t *testing.T) {
	var opts struct {
		Value bool `short:"v"`

		Cmd1 struct{} `command:"remove"`

		Cmd2 struct{} `command:"add"`
	}

	p := NewParser(&opts, None)
	p.SubcommandsOptional = true

	retargs, err := p.ParseArgs([]string{"nocmd", "remove"})
	assert.NoError(t, err)

	assert.Equal(t, retargs, []string{"nocmd", "remove"})
}

func TestCommandAlias(t *testing.T) {
	var opts struct {
		Command struct {
			G string `short:"g" default:"value"`
		} `command:"cmd" alias:"cm"`
	}

	useParsedArgs(t, &opts, "cm")

	assert.Equal(t, "value", opts.Command.G)
}

func TestSubCommandFindOptionByLongFlag(t *testing.T) {
	var opts struct {
		Testing bool `long:"testing" description:"Testing"`
	}

	var cmd struct {
		Other bool `long:"other" description:"Other"`
	}

	p := NewParser(&opts, Default)
	c := p.AddCommand("command", "Short", "Long", &cmd)

	opt := c.FindOptionByLongName("other")
	assert.NotNil(t, opt)

	assert.Equal(t, opt.LongName, "other")

	opt = c.FindOptionByLongName("testing")
	assert.NotNil(t, opt)

	assert.Equal(t, opt.LongName, "testing")
}

func TestSubCommandFindOptionByShortFlag(t *testing.T) {
	var opts struct {
		Testing bool `short:"t" description:"Testing"`
	}

	var cmd struct {
		Other bool `short:"o" description:"Other"`
	}

	p := NewParser(&opts, Default)
	c := p.AddCommand("command", "Short", "Long", &cmd)

	opt := c.FindOptionByShortName('o')
	assert.NotNil(t, opt)

	assert.Equal(t, 'o', opt.ShortName)

	opt = c.FindOptionByShortName('t')
	assert.NotNil(t, opt)

	assert.Equal(t, 't', opt.ShortName)
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
	var opts struct {
		Value bool   `short:"v"`
		Foo   fooCmd `command:"foo"`
	}
	p := NewParser(&opts, PassAfterNonOption)
	ret, err := p.ParseArgs([]string{"-v", "foo", "-f", "bar", "-v", "-g"})
	assert.NoError(t, err)

	assert.True(t, opts.Value)

	assert.True(t, opts.Foo.Flag)

	assert.Equal(t, ret, []string{"bar", "-v", "-g"})
	assert.Equal(t, opts.Foo.args, []string{"bar", "-v", "-g"})
}

type barCmd struct {
	fooCmd
	Positional struct {
		Args []string
	} `positional-args:"yes"`
}

func TestCommandPassAfterNonOptionWithPositional(t *testing.T) {
	var opts struct {
		Value bool   `short:"v"`
		Bar   barCmd `command:"bar"`
	}
	p := NewParser(&opts, PassAfterNonOption)
	ret, err := p.ParseArgs([]string{"-v", "bar", "-f", "baz", "-v", "-g"})
	assert.NoError(t, err)

	assert.True(t, opts.Value)

	assert.True(t, opts.Bar.Flag)

	assert.Equal(t, ret, []string{})
	assert.Equal(t, opts.Bar.args, []string{})
	assert.Equal(t, opts.Bar.Positional.Args, []string{"baz", "-v", "-g"})
}
