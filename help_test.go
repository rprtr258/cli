package cli

import (
	"testing"
)

// A port of the go-flags `TestHelp` golden test, adapted to this library's
// declaration style: `cli` tags, groups, struct-side commands with aliases,
// hidden fields, map flags, pointer slices, choices and positionals.

type helpBommand struct {
	Hidden bool `cli:"long=hidden,help='A hidden goflags.Option'"`
}

type helpCommand struct {
	ExtraVerbose []bool `cli:"long=extra-verbose,help='Use for extra verbosity'"`
}

type helpParentSub struct {
	Opt string `cli:"long=opt,help='This is a sub command goflags.Option'"`
}

type helpParent struct {
	Opt        string        `cli:"long=opt,help='This is a parent command goflags.Option'"`
	SubCommand helpParentSub `cli:"command=sub,help='A sub command'"`
}

type cmdHelp struct {
	Verbose          []bool    `cli:"short=v,long=verbose,help='Show verbose debug information'"`
	PtrSlice         []*string `cli:"long=ptrslice,help='A slice of pointers to string'"`
	EmptyDescription bool      `cli:"long=empty-description"`

	Default           string            `cli:"long=default,default='Some value',help='Test default value'"`
	DefaultMap        map[string]string `cli:"long=default-map,default=some:value,default=another:value,help='Testdefault map value'"`
	EnvDefault1       string            `cli:"long=env-default1,default='Some value',help='Test env-default1 value'"`
	EnvDefault2       string            `cli:"long=env-default2,help='Test env-default2 value'"`
	OptionWithArgName string            `cli:"long=opt-with-arg-name,metavar=something,help='Opt with named argument'"`
	OptionWithChoices string            `cli:"long=opt-with-choices,metavar=choice,choice=dog,choice=cat,help='Opt with choices'"`
	Hidden            string            `cli:"long=hidden,help='Hidden goflags.Option'"`

	HiddenOptionWithVeryLongName bool `cli:"long=this-hidden-option-has-a-ridiculously-long-name"`

	StringSlice []string       `cli:"short=s,help='A slice of strings'"`
	IntMap      map[string]int `cli:"long=intmap,default=a:1,help='A map from string to int'"`

	// positional items
	Filename     string  `cli:"metavar=filename,help='A filename with a long description to trigger line wrapping'"`
	Number       int     `cli:"metavar=num,help='A number'"`
	HiddenInHelp float32 `cli:"metavar=hidden-in-help"`

	HelpBommand       helpBommand `cli:"command=bommand,help='A command with only hidden goflags.Options'"`
	HelpCommand       helpCommand `cli:"command=command,alias=cm,alias=cmd,help='A command'"`
	HelpHiddenCmd     helpBommand `cli:"command=hidden-command,help='A hidden command'"`
	HelpParentCommand helpParent  `cli:"command=parent,help='A parent command'"`
}

func TestHelp(t *testing.T) {
	parser := Construct[cmdHelp]().ToOptions()

	assertHelp(t, parser, `Usage: [-v]... [--ptrslice=PTRSLICE]... [--empty-description] [--default=DEFAULT] [--default-map=
<KEY:VALUE>] [--env-default1=ENVDEFAULT1] --env-default2=ENVDEFAULT2 --opt-with-arg-name=<something>
--opt-with-choices=<choice> --hidden=HIDDEN [--this-hidden-option-has-a-ridiculously-long-name] [-s=
STRINGSLICE]... [--intmap=<KEY:VALUE>] COMMAND ... COMMAND ... COMMAND ... COMMAND ... <filename>
<num> <hidden-in-help>

Available positional items:
    <filename>                A filename with a long description to trigger line wrapping
    <num>                     A number

Available options:
    -v, --verbose             Show verbose debug information
        --ptrslice=PTRSLICE   A slice of pointers to string
        --empty-description
        --default=DEFAULT     Test default value
                              [default: Some value]
        --default-map=<KEY:VALUE>  Testdefault map value
                              [default: map[another:value some:value]]
        --env-default1=ENVDEFAULT1  Test env-default1 value
                              [default: Some value]
        --env-default2=ENVDEFAULT2  Test env-default2 value
        --opt-with-arg-name=<something>  Opt with named argument
        --opt-with-choices=<choice>  Opt with choices  (one of: cat, dog)
        --hidden=HIDDEN       Hidden goflags.Option
        --this-hidden-option-has-a-ridiculously-long-name
    -s=STRINGSLICE            A slice of strings
        --intmap=<KEY:VALUE>  A map from string to int
                              [default: map[a:1]]
    -h, --help                Prints help information

Available commands:
    bommand                   A command with only hidden goflags.Options
    command                   A command
    hidden-command            A hidden command
    parent                    A parent command
`)
}

type cmdHelpCommands struct {
	EnvDefault2       string `cli:"long=env-default2,help='Test env-default2 value'"`
	OptionWithArgName string `cli:"long=opt-with-arg-name,metavar=something,help='Opt with named argument'"`
	OptionWithChoices string `cli:"long=opt-with-choices,metavar=choice,choice=dog,choice=cat,help='Opt with choices'"`
	Hidden            string `cli:"long=hidden,help='Hidden goflags.Option'"`

	HelpBommand       helpBommand `cli:"command=bommand,help='A command with only hidden goflags.Options'"`
	HelpCommand       helpCommand `cli:"command=command,alias=cm,alias=cmd,help='A command'"`
	HelpHiddenCmd     helpBommand `cli:"command=hidden-command,help='A hidden command'"`
	HelpParentCommand helpParent  `cli:"command=parent,help='A parent command'"`
}

func TestHelpCommandDispatch(t *testing.T) {
	parser := Construct[cmdHelpCommands]().ToOptions()

	// `--env-default2` and `--opt-with-arg-name` have no defaults: required
	assertSuccess(t, cmdHelpCommands{
		HelpCommand:       helpCommand{ExtraVerbose: []bool{true, true}},
		EnvDefault2:       "e2",
		OptionWithArgName: "an",
		OptionWithChoices: "dog",
		Hidden:            "h",
	}, parser, "command", "--extra-verbose", "--extra-verbose", "--env-default2", "e2",
		"--opt-with-arg-name", "an", "--opt-with-choices", "dog", "--hidden", "h")
	// hidden long aliases
	assertSuccess(t, cmdHelpCommands{
		HelpCommand:       helpCommand{ExtraVerbose: []bool{true}},
		EnvDefault2:       "e2",
		OptionWithArgName: "an",
		OptionWithChoices: "dog",
		Hidden:            "h",
	}, parser, "cm", "--extra-verbose", "--env-default2", "e2", "--opt-with-arg-name", "an",
		"--opt-with-choices", "dog", "--hidden", "h")
	assertSuccess(t, cmdHelpCommands{
		HelpCommand:       helpCommand{ExtraVerbose: []bool{true}},
		EnvDefault2:       "e2",
		OptionWithArgName: "an",
		OptionWithChoices: "dog",
		Hidden:            "h",
	}, parser, "cmd", "--extra-verbose", "--env-default2", "e2", "--opt-with-arg-name", "an",
		"--opt-with-choices", "dog", "--hidden", "h")
	// nested subcommand
	assertSuccess(t, cmdHelpCommands{
		EnvDefault2:       "e2",
		OptionWithArgName: "an",
		OptionWithChoices: "dog",
		Hidden:            "h",
		HelpParentCommand: helpParent{
			Opt:        "parent-opt",
			SubCommand: helpParentSub{Opt: "sub-opt"},
		},
	}, parser, "parent", "--opt", "parent-opt", "sub", "--opt", "sub-opt",
		"--env-default2", "e2", "--opt-with-arg-name", "an", "--opt-with-choices", "dog", "--hidden", "h")
}

func TestHelpCommandsOptional(t *testing.T) {
	parser := Construct[cmdHelp]().ToOptions()

	// commands are optional: absent command leaves the field zero-valued
	assertSuccess(t, cmdHelp{
		Verbose:           []bool{false},
		PtrSlice:          []*string{},
		Default:           "Some value",
		DefaultMap:        map[string]string{"some": "value", "another": "value"},
		EnvDefault1:       "Some value",
		EnvDefault2:       "e2",
		OptionWithArgName: "an",
		OptionWithChoices: "dog",
		Hidden:            "h",
		StringSlice:       []string{},
		IntMap:            map[string]int{"a": 1},
		Filename:          "file",
		Number:            3,
		HiddenInHelp:      1.5,
	}, parser, "file", "3", "1.5", "--env-default2", "e2", "--opt-with-arg-name", "an",
		"--opt-with-choices", "dog", "--hidden", "h")
}

func TestHelpFlagsAndPositionals(t *testing.T) {
	parser := Construct[cmdHelp]().ToOptions()

	assertSuccess(t, cmdHelp{
		Verbose:           []bool{true, true},
		PtrSlice:          []*string{strptr("a")},
		Default:           "Some value",
		DefaultMap:        map[string]string{"some": "value", "another": "value"},
		EnvDefault1:       "Some value",
		EnvDefault2:       "e2",
		OptionWithArgName: "an",
		OptionWithChoices: "dog",
		Hidden:            "h",
		StringSlice:       []string{"x"},
		IntMap:            map[string]int{"k": 1},
		Filename:          "name",
		Number:            2,
		HiddenInHelp:      1.5,
	}, parser, "-v", "-v", "--ptrslice", "a", "--opt-with-choices", "dog",
		"--hidden", "h", "-s", "x", "--intmap", "k:1", "name", "2", "1.5",
		"--env-default2", "e2", "--opt-with-arg-name", "an")

	// default map entries can be overridden
	assertSuccess(t, cmdHelp{
		Verbose:           []bool{false},
		PtrSlice:          []*string{},
		Default:           "Some value",
		DefaultMap:        map[string]string{"x": "1"},
		EnvDefault1:       "Some value",
		EnvDefault2:       "e2",
		OptionWithArgName: "an",
		OptionWithChoices: "dog",
		Hidden:            "h",
		StringSlice:       []string{},
		IntMap:            map[string]int{"a": 1},
		Filename:          "name",
		Number:            2,
		HiddenInHelp:      1.5,
	}, parser, "--default-map", "x:1", "--env-default2", "e2", "--opt-with-arg-name", "an",
		"--opt-with-choices", "dog", "--hidden", "h", "name", "2", "1.5")
}

func TestHelpChoices(t *testing.T) {
	parser := Construct[cmdHelp]().ToOptions()

	assertFailure(t, parser, "couldn't parse `zebra`: expected one of `cat`, `dog`",
		"--opt-with-choices", "zebra")
}
