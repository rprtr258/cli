package flags

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCommandInline(t *testing.T) {
	opts := struct {
		Value bool `short:"v"`

		Command struct {
			G bool `short:"g"`
		} `command:"cmd"`
	}{}

	p, ret := assertParserSuccess(t, &opts, "-v", "cmd", "-g")

	assert.Equal(t, ret, []string{})

	if p.Active == nil {
		t.Errorf("Expected active command")
	}

	if !opts.Value {
		t.Errorf("Expected Value to be true")
	}

	if !opts.Command.G {
		t.Errorf("Expected Command.G to be true")
	}

	if p.Command.Find("cmd") != p.Active {
		t.Errorf("Expected to find command `cmd' to be active")
	}
}

func TestCommandInlineMulti(t *testing.T) {
	opts := struct {
		Value bool `short:"v"`

		C1 struct{} `command:"c1"`

		C2 struct {
			G bool `short:"g"`
		} `command:"c2"`
	}{}

	p, ret := assertParserSuccess(t, &opts, "-v", "c2", "-g")

	assert.Equal(t, ret, []string{})

	if p.Active == nil {
		t.Errorf("Expected active command")
	}

	if !opts.Value {
		t.Errorf("Expected Value to be true")
	}

	if !opts.C2.G {
		t.Errorf("Expected C2.G to be true")
	}

	if p.Command.Find("c1") == nil {
		t.Errorf("Expected to find command `c1'")
	}

	if c2 := p.Command.Find("c2"); c2 == nil {
		t.Errorf("Expected to find command `c2'")
	} else if c2 != p.Active {
		t.Errorf("Expected to find command `c2' to be active")
	}
}

func TestCommandFlagOrder1(t *testing.T) {
	opts := struct {
		Value bool `short:"v"`

		Command struct {
			G bool `short:"g"`
		} `command:"cmd"`
	}{}

	assertParseFail(t, ErrUnknownFlag, "unknown flag `g'", &opts, "-v", "-g", "cmd")
}

func TestCommandFlagOrder2(t *testing.T) {
	opts := struct {
		Value bool `short:"v"`

		Command struct {
			G bool `short:"g"`
		} `command:"cmd"`
	}{}

	assertParseSuccess(t, &opts, "cmd", "-v", "-g")

	if !opts.Value {
		t.Errorf("Expected Value to be true")
	}

	if !opts.Command.G {
		t.Errorf("Expected Command.G to be true")
	}
}

func TestCommandFlagOrderSub(t *testing.T) {
	opts := struct {
		Value bool `short:"v"`

		Command struct {
			G bool `short:"g"`

			SubCommand struct {
				B bool `short:"b"`
			} `command:"sub"`
		} `command:"cmd"`
	}{}

	assertParseSuccess(t, &opts, "cmd", "sub", "-v", "-g", "-b")

	if !opts.Value {
		t.Errorf("Expected Value to be true")
	}

	if !opts.Command.G {
		t.Errorf("Expected Command.G to be true")
	}

	if !opts.Command.SubCommand.B {
		t.Errorf("Expected Command.SubCommand.B to be true")
	}
}

func TestCommandFlagOverride1(t *testing.T) {
	opts := struct {
		Value bool `short:"v"`

		Command struct {
			Value bool `short:"v"`
		} `command:"cmd"`
	}{}

	assertParseSuccess(t, &opts, "-v", "cmd")

	if !opts.Value {
		t.Errorf("Expected Value to be true")
	}

	if opts.Command.Value {
		t.Errorf("Expected Command.Value to be false")
	}
}

func TestCommandFlagOverride2(t *testing.T) {
	opts := struct {
		Value bool `short:"v"`

		Command struct {
			Value bool `short:"v"`
		} `command:"cmd"`
	}{}

	assertParseSuccess(t, &opts, "cmd", "-v")

	if opts.Value {
		t.Errorf("Expected Value to be false")
	}

	if !opts.Command.Value {
		t.Errorf("Expected Command.Value to be true")
	}
}

func TestCommandFlagOverrideSub(t *testing.T) {
	opts := struct {
		Value bool `short:"v"`

		Command struct {
			Value bool `short:"v"`

			SubCommand struct {
				Value bool `short:"v"`
			} `command:"sub"`
		} `command:"cmd"`
	}{}

	assertParseSuccess(t, &opts, "cmd", "sub", "-v")

	if opts.Value {
		t.Errorf("Expected Value to be false")
	}

	if opts.Command.Value {
		t.Errorf("Expected Command.Value to be false")
	}

	if !opts.Command.SubCommand.Value {
		t.Errorf("Expected Command.Value to be true")
	}
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

	assertParseSuccess(t, &opts, "cmd", "sub", "-v")
	assert.False(t, opts.Value)
	assert.True(t, opts.Command.Value)
}

func TestCommandEstimate(t *testing.T) {
	opts := struct {
		Value bool     `short:"v"`
		Cmd1  struct{} `command:"remove"`
		Cmd2  struct{} `command:"add"`
	}{}

	p := NewParser(&opts, None)
	_, err := p.ParseArgs()
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

	p := NewParser(&opts, None)
	_, err := p.ParseArgs("rmive")
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

	assertParseSuccess(t, &opts, "-v", "cmd", "-g", "a", "b")
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

	args := assertParseFail(t, ErrUnknownCommand, "Unknown command `addd', did you mean `add'?", &opts, "-v", "addd")
	assert.Equal(t, []string{"addd"}, args)
}

func TestCommandAdd(t *testing.T) {
	opts := struct {
		Value bool `short:"v"`
	}{}

	cmd := struct {
		G bool `short:"g"`
	}{}

	p := NewParser(&opts, Default)
	c, err := p.AddCommand("cmd", "", "", &cmd)
	assert.NoError(t, err)

	ret, err := p.ParseArgs("-v", "cmd", "-g", "rest")
	assert.NoError(t, err)
	assert.Equal(t, []string{"rest"}, ret)
	assert.True(t, opts.Value)
	assert.True(t, cmd.G)
	assert.Equal(t, c, p.Command.Find("cmd"))
	assert.Equal(t, c, p.Commands()[0])
	assert.Equal(t, 'g', c.Options()[0].ShortName)
}

func TestCommandNestedInline(t *testing.T) {
	opts := struct {
		Value bool `short:"v"`

		Command struct {
			G bool `short:"g"`

			Nested struct {
				N string `long:"n"`
			} `command:"nested"`
		} `command:"cmd"`
	}{}

	p, ret := assertParserSuccess(t, &opts, "-v", "cmd", "-g", "nested", "--n", "n", "rest")
	assert.Equal(t, []string{"rest"}, ret)
	assert.True(t, opts.Value)
	assert.True(t, opts.Command.G)
	assert.Equal(t, "n", opts.Command.Nested.N)

	c := p.Command.Find("cmd")
	assert.NotNil(t, c)
	assert.Equal(t, c, p.Active)

	nested := c.Find("nested")
	assert.NotNil(t, nested)
	assert.Equal(t, nested, c.Active)
}

func TestRequiredOnCommand(t *testing.T) {
	opts := struct {
		Value   bool `short:"v" required:"true"`
		Command struct {
			G bool `short:"g"`
		} `command:"cmd"`
	}{}

	assertParseFail(t, ErrRequired, fmt.Sprintf("the required flag `%cv' was not specified", defaultShortOptDelimiter), &opts, "cmd")
}

func TestRequiredAllOnCommand(t *testing.T) {
	opts := struct {
		Value   bool `short:"v" required:"true"`
		Missing bool `long:"missing" required:"true"`
		Command struct {
			G bool `short:"g"`
		} `command:"cmd"`
	}{}

	assertParseFail(t, ErrRequired, fmt.Sprintf("the required flags `%smissing' and `%cv' were not specified", defaultLongOptDelimiter, defaultShortOptDelimiter), &opts, "cmd")
}

func TestDefaultOnCommand(t *testing.T) {
	opts := struct {
		Command struct {
			G string `short:"g" default:"value"`
		} `command:"cmd"`
	}{}

	assertParseSuccess(t, &opts, "cmd")
	assert.Equal(t, "value", opts.Command.G)
}

func TestAfterNonCommand(t *testing.T) {
	opts := struct {
		Value bool     `short:"v"`
		Cmd1  struct{} `command:"remove"`
		Cmd2  struct{} `command:"add"`
	}{}

	assertParseFail(t, ErrUnknownCommand, "Unknown command `nocmd'. Please specify one command of: remove or add", &opts, "nocmd", "remove")
}

func TestSubcommandsOptional(t *testing.T) {
	opts := struct {
		Value bool     `short:"v"`
		Cmd1  struct{} `command:"remove"`
		Cmd2  struct{} `command:"add"`
	}{}

	p := NewParser(&opts, None)
	p.SubcommandsOptional = true

	_, err := p.ParseArgs("-v")
	assert.NoError(t, err)
	assert.True(t, opts.Value)
}

func TestSubcommandsOptionalAfterNonCommand(t *testing.T) {
	opts := struct {
		Value bool     `short:"v"`
		Cmd1  struct{} `command:"remove"`
		Cmd2  struct{} `command:"add"`
	}{}

	p := NewParser(&opts, None)
	p.SubcommandsOptional = true

	retargs, err := p.ParseArgs("nocmd", "remove")
	assert.NoError(t, err)
	assert.Equal(t, []string{"nocmd", "remove"}, retargs)
}

func TestCommandAlias(t *testing.T) {
	opts := struct {
		Command struct {
			G string `short:"g" default:"value"`
		} `command:"cmd" alias:"cm"`
	}{}

	assertParseSuccess(t, &opts, "cm")
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
	c, _ := p.AddCommand("command", "Short", "Long", &cmd)

	opt := c.FindOptionByLongName("other")
	assert.NotNil(t, opt)
	assert.Equal(t, "other", opt.LongName)

	opt = c.FindOptionByLongName("testing")
	assert.NotNil(t, opt)
	assert.Equal(t, "testing", opt.LongName)
}

func TestSubCommandFindOptionByShortFlag(t *testing.T) {
	var opts struct {
		Testing bool `short:"t" description:"Testing"`
	}

	var cmd struct {
		Other bool `short:"o" description:"Other"`
	}

	p := NewParser(&opts, Default)
	c, _ := p.AddCommand("command", "Short", "Long", &cmd)

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
	opts := struct {
		Value bool   `short:"v"`
		Foo   fooCmd `command:"foo"`
	}{}
	ret, err := NewParser(&opts, PassAfterNonOption).ParseArgs("-v", "foo", "-f", "bar", "-v", "-g")
	assert.NoError(t, err)
	assert.True(t, opts.Value)
	assert.True(t, opts.Foo.Flag)
	assert.Equal(t, []string{"bar", "-v", "-g"}, ret)
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
	p := NewParser(&opts, PassAfterNonOption)
	ret, err := p.ParseArgs("-v", "bar", "-f", "baz", "-v", "-g")
	assert.NoError(t, err)
	assert.True(t, opts.Value)
	assert.True(t, opts.Bar.Flag)
	assert.Equal(t, []string{}, ret)
	assert.Equal(t, []string{}, opts.Bar.args)
	assert.Equal(t, []string{"baz", "-v", "-g"}, opts.Bar.Positional.Args)
}
