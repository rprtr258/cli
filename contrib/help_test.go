package flags

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"runtime"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type helpOptions struct {
	Verbose          []bool       `short:"v" long:"verbose" description:"Show verbose debug information"`
	Call             func(string) `short:"c" description:"Call phone number"`
	PtrSlice         []*string    `long:"ptrslice" description:"A slice of pointers to string"`
	EmptyDescription bool         `long:"empty-description"`

	Default           string            `long:"default" default:"Some\nvalue" description:"Test default value"`
	DefaultArray      []string          `long:"default-array" default:"Some value" default:"Other\tvalue" description:"Test default array value"`
	DefaultMap        map[string]string `long:"default-map" default:"some:value" default:"another:value" description:"Testdefault map value"`
	EnvDefault1       string            `long:"env-default1" default:"Some value" description:"Test env-default1 value"`
	EnvDefault2       string            `long:"env-default2" description:"Test env-default2 value"`
	OptionWithArgName string            `long:"opt-with-arg-name" value-name:"something" description:"Option with named argument"`
	OptionWithChoices string            `long:"opt-with-choices" value-name:"choice" choice:"dog" choice:"cat" description:"Option with choices"`
	Hidden            string            `long:"hidden" description:"Hidden option" hidden:"yes"`

	HiddenOptionWithVeryLongName bool `long:"this-hidden-option-has-a-ridiculously-long-name" hidden:"yes"`

	Other struct {
		StringSlice []string       `short:"s" default:"some" default:"value" description:"A slice of strings"`
		IntMap      map[string]int `long:"intmap" default:"a:1" description:"A map from string to int"`
	} `group:"Other Options"`

	HiddenGroup struct {
		InsideHiddenGroup string `long:"inside-hidden-group" description:"Inside hidden group"`
		Padder            bool   `long:"this-option-in-a-hidden-group-has-a-ridiculously-long-name"`
	} `group:"Hidden group" hidden:"yes"`

	GroupWithOnlyHiddenOptions struct {
		SecretFlag bool `long:"secret" description:"Hidden flag in a non-hidden group" hidden:"yes"`
	} `group:"Non-hidden group with only hidden options"`

	Group struct {
		Opt                  string `long:"opt" description:"This is a subgroup option"`
		HiddenInsideGroup    string `long:"hidden-inside-group" description:"Hidden inside group" hidden:"yes"`
		NotHiddenInsideGroup string `long:"not-hidden-inside-group" description:"Not hidden inside group" hidden:"false"`

		Group struct {
			Opt string `long:"opt" description:"This is a subsubgroup option"`
		} `group:"Subsubgroup" namespace:"sap"`
	} `group:"Subgroup" namespace:"sip"`

	Bommand struct {
		Hidden bool `long:"hidden" description:"A hidden option" hidden:"yes"`
	} `command:"bommand" description:"A command with only hidden options"`

	Command struct {
		ExtraVerbose []bool `long:"extra-verbose" description:"Use for extra verbosity"`
	} `command:"command" alias:"cm" alias:"cmd" description:"A command"`

	HiddenCommand struct {
		ExtraVerbose []bool `long:"extra-verbose" description:"Use for extra verbosity"`
	} `command:"hidden-command" description:"A hidden command" hidden:"yes"`

	ParentCommand struct {
		Opt        string `long:"opt" description:"This is a parent command option"`
		SubCommand struct {
			Opt string `long:"opt" description:"This is a sub command option"`
		} `command:"sub" description:"A sub command"`
	} `command:"parent" description:"A parent command"`

	Args struct {
		Filename     string  `positional-arg-name:"filename" description:"A filename with a long description to trigger line wrapping"`
		Number       int     `positional-arg-name:"num" description:"A number"`
		HiddenInHelp float32 `positional-arg-name:"hidden-in-help" required:"yes"`
	} `positional-args:"yes"`
}

func TestHelp(t *testing.T) {
	var opts helpOptions
	p := NewNamedParser("TestHelp", HelpFlag)
	_, _ = p.AddGroup("Application Options", "The application options", &opts)

	_, err := p.ParseArgs("--help")
	var expected string
	if runtime.GOOS == "windows" {
		expected = `Usage:
  TestHelp [OPTIONS] [filename] [num] hidden-in-help <bommand | command | parent>

Application Options:
  /v, /verbose                              Show verbose debug information
  /c:                                       Call phone number
      /ptrslice:                            A slice of pointers to string
      /empty-description
      /default:                             Test default value (default:
                                            "Some\nvalue")
      /default-array:                       Test default array value (default:
                                            Some value, "Other\tvalue")
      /default-map:                         Testdefault map value (default:
                                            some:value, another:value)
      /env-default1:                        Test env-default1 value (default:
                                            Some value)
      /env-default2:                        Test env-default2 value
      /opt-with-arg-name:something          Option with named argument
      /opt-with-choices:choice[dog|cat]     Option with choices

Other Options:
  /s:                                       A slice of strings (default: some,
                                            value)
      /intmap:                              A map from string to int (default:
                                            a:1)

Subgroup:
      /sip.opt:                             This is a subgroup option
      /sip.not-hidden-inside-group:         Not hidden inside group

Subsubgroup:
      /sip.sap.opt:                         This is a subsubgroup option

Help Options:
  /?                                        Show this help message
  /h, /help                                 Show this help message

Arguments:
  filename:                                 A filename with a long description
                                            to trigger line wrapping
  num:                                      A number

Available commands:
  bommand  A command with only hidden options
  command  A command (aliases: cm, cmd)
  parent   A command with a sub command
`
	} else {
		expected = `Usage:
  TestHelp [OPTIONS] [filename] [num] hidden-in-help <bommand | command | parent>

Application Options:
  -v, --verbose                             Show verbose debug information
  -c=                                       Call phone number
      --ptrslice=                           A slice of pointers to string
      --empty-description
      --default=                            Test default value (default:
                                            "Some\nvalue")
      --default-array=                      Test default array value (default:
                                            Some value, "Other\tvalue")
      --default-map=                        Testdefault map value (default:
                                            some:value, another:value)
      --env-default1=                       Test env-default1 value (default:
                                            Some value)
      --env-default2=                       Test env-default2 value
      --opt-with-arg-name=something         Option with named argument
      --opt-with-choices=choice[dog|cat]    Option with choices

Other Options:
  -s=                                       A slice of strings (default: some,
                                            value)
      --intmap=                             A map from string to int (default:
                                            a:1)

Subgroup:
      --sip.opt=                            This is a subgroup option
      --sip.not-hidden-inside-group=        Not hidden inside group

Subsubgroup:
      --sip.sap.opt=                        This is a subsubgroup option

Help Options:
  -h, --help                                Show this help message

Arguments:
  filename:                                 A filename with a long description
                                            to trigger line wrapping
  num:                                      A number

Available commands:
  bommand         A command with only hidden options
  command,cm,cmd  A command
  parent          A parent command
`
	}
	assert.Equal(t, &Error{
		Kind:    ErrHelp,
		Message: expected,
	}, err)
}

func TestMan(t *testing.T) {
	p := NewNamedParser("TestMan", HelpFlag)
	p.ShortDescription = "Test manpage generation"
	p.LongDescription = "This is a somewhat `longer' description of what this does.\nWith multiple lines."
	_, _ = p.AddGroup("Application Options", "The application options", &helpOptions{})

	for _, cmd := range p.Commands() {
		cmd.LongDescription = fmt.Sprintf("Longer `%s' description", cmd.Name)
	}

	tt := time.Now()

	var buf bytes.Buffer
	p.WriteManPage(&buf, tt)

	got := buf.String()

	expected := fmt.Sprintf(`.TH TestMan 1 "%s"
.SH NAME
TestMan \- Test manpage generation
.SH SYNOPSIS
\fBTestMan\fP [OPTIONS]
.SH DESCRIPTION
This is a somewhat \fBlonger\fP description of what this does.
With multiple lines.
.SH OPTIONS
.SS Application Options
The application options
.TP
\fB\fB\-v\fR, \fB\-\-verbose\fR\fP
Show verbose debug information
.TP
\fB\fB\-c\fR\fP
Call phone number
.TP
\fB\fB\-\-ptrslice\fR\fP
A slice of pointers to string
.TP
\fB\fB\-\-empty-description\fR\fP
.TP
\fB\fB\-\-default\fR <default: \fI"Some\\nvalue"\fR>\fP
Test default value
.TP
\fB\fB\-\-default-array\fR <default: \fI"Some value", "Other\\tvalue"\fR>\fP
Test default array value
.TP
\fB\fB\-\-default-map\fR <default: \fI"some:value", "another:value"\fR>\fP
Testdefault map value
.TP
\fB\fB\-\-env-default1\fR <default: \fI"Some value"\fR>\fP
Test env-default1 value
.TP
\fB\fB\-\-env-default2\fR\fP
Test env-default2 value
.TP
\fB\fB\-\-opt-with-arg-name\fR \fIsomething\fR\fP
Option with named argument
.TP
\fB\fB\-\-opt-with-choices\fR \fIchoice\fR\fP
Option with choices
.SS Other Options
.TP
\fB\fB\-s\fR <default: \fI"some", "value"\fR>\fP
A slice of strings
.TP
\fB\fB\-\-intmap\fR <default: \fI"a:1"\fR>\fP
A map from string to int
.SS Subgroup
.TP
\fB\fB\-\-sip.opt\fR\fP
This is a subgroup option
.TP
\fB\fB\-\-sip.not-hidden-inside-group\fR\fP
Not hidden inside group
.SS Subsubgroup
.TP
\fB\fB\-\-sip.sap.opt\fR\fP
This is a subsubgroup option
.SH COMMANDS
.SS bommand
A command with only hidden options

Longer \fBbommand\fP description
.SS command
A command

Longer \fBcommand\fP description

\fBUsage\fP: TestMan [OPTIONS] command [command-OPTIONS]
.TP

\fBAliases\fP: cm, cmd

.TP
\fB\fB\-\-extra-verbose\fR\fP
Use for extra verbosity
.SS parent
A parent command

Longer \fBparent\fP description

\fBUsage\fP: TestMan [OPTIONS] parent [parent-OPTIONS]
.TP
.TP
\fB\fB\-\-opt\fR\fP
This is a parent command option
.SS parent sub
A sub command

\fBUsage\fP: TestMan [OPTIONS] parent [parent-OPTIONS] sub [sub-OPTIONS]
.TP
.TP
\fB\fB\-\-opt\fR\fP
This is a sub command option
`, tt.Format("2 January 2006"))
	assert.Equal(t, expected, got)
}

type helpCommandNoOptions struct {
	Command struct{} `command:"command" description:"A command"`
}

func TestHelpCommand(t *testing.T) {
	var opts helpCommandNoOptions
	p := NewNamedParser("TestHelpCommand", HelpFlag)
	_, _ = p.AddGroup("Application Options", "The application options", &opts)

	_, err := p.ParseArgs("command", "--help")
	var expected string
	if runtime.GOOS == "windows" {
		expected = `Usage:
  TestHelpCommand [OPTIONS] command

Help Options:
  /?              Show this help message
  /h, /help       Show this help message
`
	} else {
		expected = `Usage:
  TestHelpCommand [OPTIONS] command

Help Options:
  -h, --help      Show this help message
`
	}
	assert.Equal(t, &Error{
		Kind:    ErrHelp,
		Message: expected,
	}, err)
}

func TestHiddenCommandNoBuiltinHelp(t *testing.T) {
	// no auto added help group
	p := NewNamedParser("TestHelpCommand", 0)
	// and no usage information either
	p.Usage = ""

	// add custom help group which is not listed in --help output
	var help struct {
		ShowHelp func() error `short:"h" long:"help"`
	}
	help.ShowHelp = func() error {
		return &Error{Kind: ErrHelp}
	}
	hlpgrp, err := p.AddGroup("Help Options", "", &help)
	assert.NoError(t, err)
	hlpgrp.Hidden = true
	hlp := p.FindOptionByLongName("help")
	hlp.Description = "Show this help message"
	// make sure the --help option is hidden
	hlp.Hidden = true

	// add a hidden command
	var hiddenCmdOpts struct {
		Foo        bool `short:"f" long:"very-long-foo-option" description:"Very long foo description"`
		Bar        bool `short:"b" description:"Option bar"`
		Positional struct {
			PositionalFoo string `positional-arg-name:"<positional-foo>" description:"positional foo"`
		} `positional-args:"yes"`
	}
	cmdHidden, err := p.Command.AddCommand("hidden", "Hidden command description", "Long hidden command description", &hiddenCmdOpts)
	assert.NoError(t, err)
	// make it hidden
	cmdHidden.Hidden = true
	assert.Len(t, cmdHidden.Options(), 2)

	// which help we ask for explicitly
	_, err = p.ParseArgs("hidden", "--help")
	assert.Equal(t, &Error{
		Kind:    ErrHelp,
		Message: "",
	}, err)
	h := &bytes.Buffer{}
	p.WriteHelp(h)
	const _expected = `Usage:
  TestHelpCommand hidden [hidden-OPTIONS] [<positional-foo>]

Long hidden command description

[hidden command arguments]
  <positional-foo>:         positional foo
`
	assert.Equal(t, _expected, h.String())
}

func TestHelpDefaults(t *testing.T) {
	var expected string
	if runtime.GOOS == "windows" {
		expected = `Usage:
  TestHelpDefaults [OPTIONS]

Application Options:
      /with-default:               With default (default: default-value)
      /without-default:            Without default
      /with-programmatic-default:  With programmatic default (default:
                                   default-value)

Help Options:
  /?                               Show this help message
  /h, /help                        Show this help message
`
	} else {
		expected = `Usage:
  TestHelpDefaults [OPTIONS]

Application Options:
      --with-default=              With default (default: default-value)
      --without-default=           Without default
      --with-programmatic-default= With programmatic default (default:
                                   default-value)

Help Options:
  -h, --help                       Show this help message
`
	}

	for _, test := range []struct {
		Args   []string
		Output string
	}{
		{
			Args:   []string{"-h"},
			Output: expected,
		},
		{
			Args:   []string{"--with-default", "other-value", "--with-programmatic-default", "other-value", "-h"},
			Output: expected,
		},
	} {
		var opts struct {
			WithDefault             string `long:"with-default" default:"default-value" description:"With default"`
			WithoutDefault          string `long:"without-default" description:"Without default"`
			WithProgrammaticDefault string `long:"with-programmatic-default" description:"With programmatic default"`
		}

		opts.WithProgrammaticDefault = "default-value"

		p := NewNamedParser("TestHelpDefaults", HelpFlag)
		_, _ = p.AddGroup("Application Options", "The application options", &opts)

		_, err := p.ParseArgs(test.Args...)
		assert.Equal(t, &Error{
			Kind:    ErrHelp,
			Message: test.Output,
		}, err)
	}
}

func TestHelpRestArgs(t *testing.T) {
	opts := struct {
		Verbose bool `short:"v"`
	}{}

	p := NewNamedParser("TestHelpDefaults", HelpFlag)
	_, _ = p.AddGroup("Application Options", "The application options", &opts)

	retargs, err := p.ParseArgs("-h", "-v", "rest")
	assert.Error(t, err)
	assert.Equal(t, retargs, []string{"-v", "rest"})
}

func TestWrapText(t *testing.T) {
	s := "Lorem ipsum dolor sit amet, consectetur adipisicing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat. Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla pariatur. Excepteur sint occaecat cupidatat non proident, sunt in culpa qui officia deserunt mollit anim id est laborum."
	got := wrapText(s, 60, "      ")
	const _expected = `Lorem ipsum dolor sit amet, consectetur adipisicing elit,
      sed do eiusmod tempor incididunt ut labore et dolore magna
      aliqua. Ut enim ad minim veniam, quis nostrud exercitation
      ullamco laboris nisi ut aliquip ex ea commodo consequat.
      Duis aute irure dolor in reprehenderit in voluptate velit
      esse cillum dolore eu fugiat nulla pariatur. Excepteur sint
      occaecat cupidatat non proident, sunt in culpa qui officia
      deserunt mollit anim id est laborum.`
	assert.Equal(t, got, _expected)
}

func TestWrapParagraph(t *testing.T) {
	s := "Lorem ipsum dolor sit amet, consectetur adipisicing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua.\n\n" +
		"Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat.\n\n" +
		"Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla pariatur.\n\n" +
		"Excepteur sint occaecat cupidatat non proident, sunt in culpa qui officia deserunt mollit anim id est laborum.\n"
	got := wrapText(s, 60, "      ")
	const _expected = `Lorem ipsum dolor sit amet, consectetur adipisicing elit,
      sed do eiusmod tempor incididunt ut labore et dolore magna
      aliqua.

      Ut enim ad minim veniam, quis nostrud exercitation ullamco
      laboris nisi ut aliquip ex ea commodo consequat.

      Duis aute irure dolor in reprehenderit in voluptate velit
      esse cillum dolore eu fugiat nulla pariatur.

      Excepteur sint occaecat cupidatat non proident, sunt in
      culpa qui officia deserunt mollit anim id est laborum.
`
	assert.Equal(t, got, _expected)
}

func TestHelpDefaultMask(t *testing.T) {
	tests := []struct {
		opts    any
		present string
	}{
		{
			opts: &struct {
				Value string `short:"v" default:"123" description:"V"`
			}{},
			present: "V (default: 123)\n",
		},
		{
			opts: &struct {
				Value string `short:"v" default:"123" default-mask:"abc" description:"V"`
			}{},
			present: "V (default: abc)\n",
		},
		{
			opts: &struct {
				Value string `short:"v" default:"123" default-mask:"-" description:"V"`
			}{},
			present: "V\n",
		},
		{
			opts: &struct {
				Value string `short:"v" description:"V"`
			}{Value: "123"},
			present: "V (default: 123)\n",
		},
		{
			opts: &struct {
				Value string `short:"v" default-mask:"abc" description:"V"`
			}{Value: "123"},
			present: "V (default: abc)\n",
		},
		{
			opts: &struct {
				Value string `short:"v" default-mask:"-" description:"V"`
			}{Value: "123"},
			present: "V\n",
		},
	}

	for _, test := range tests {
		p := NewParser(test.opts, HelpFlag)
		_, err := p.ParseArgs("-h")
		if flagsErr, ok := err.(*Error); ok && flagsErr.Kind == ErrHelp {
			err = nil
		}
		assert.NoError(t, err)
		h := &bytes.Buffer{}
		w := bufio.NewWriter(h)
		p.writeHelpOption(w, p.FindOptionByShortName('v'), p.getAlignmentInfo())
		w.Flush()
		assert.Contains(t, h.String(), test.present)
	}
}

func TestWroteHelp(t *testing.T) {
	for name, test := range map[string]struct {
		value  error
		isHelp bool
	}{
		"No error":    {value: nil, isHelp: false},
		"Plain error": {value: errors.New("an error"), isHelp: false},
		"ErrUnknown":  {value: newError(ErrUnknown, "an error"), isHelp: false},
		"ErrHelp":     {value: newError(ErrHelp, "an error"), isHelp: true},
	} {
		t.Run(name, func(t *testing.T) {
			res := WroteHelp(test.value)
			assert.Equal(t, test.isHelp, res)
		})
	}
}
