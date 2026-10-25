package cli

import (
	"testing"
)

// --- groups ---

type groupAuth struct {
	Token string `cli:"long=token,metavar=TOKEN,help='Auth token'"`
	Skip  bool   `cli:"long=skip-auth,help='Skip authentication'"`
}

type withGroup struct {
	Verbose bool      `cli:"short=v,help='Verbose output'"`
	Auth    groupAuth `cli:"group=Authentication"`
	Limit   int       `cli:"long=limit,default=10"`
}

func TestGroupParse(t *testing.T) {
	parser := Construct[withGroup]().ToOptions()

	// group options are parsed as usual parent options
	assertSuccess(t, withGroup{Verbose: true, Auth: groupAuth{Token: "secret"}, Limit: 3},
		parser, "-v --token secret --limit 3")
	// options inside a group keep their own required-ness: Token has no default
	assertFailure(t, parser, "expected `--token=TOKEN`, pass `--help` for usage information", "--limit", "3")
}

func TestGroupHelp(t *testing.T) {
	parser := Construct[withGroup]().ToOptions()

	assertHelp(t, parser, `Usage: [-v] --token=TOKEN [--skip-auth] [--limit=LIMIT]

Authentication
        --token=TOKEN  Auth token
        --skip-auth    Skip authentication

Available options:
    -v                 Verbose output
        --limit=LIMIT
                       [default: 10]
    -h, --help         Prints help information
`)
}

// --- hidden ---

type withHidden struct {
	Token  string `cli:"long=token,hidden"`
	Secret bool   `cli:"long=secret,short=s,hidden"`
}

func TestHiddenParse(t *testing.T) {
	parser := Construct[withHidden]().ToOptions()

	// hidden fields parse normally
	assertSuccess(t, withHidden{Token: "x", Secret: true}, parser, "--token x -s")
	assertSuccess(t, withHidden{}, parser, "")
}

func TestHiddenHelp(t *testing.T) {
	parser := Construct[withHidden]().ToOptions()

	assertHelp(t, parser, `Usage: 

Available options:
    -h, --help  Prints help information
`)
}

// --- commands ---

type cmdAdd struct {
	Name string `cli:"long=name,metavar=NAME"`
}

type cmdRm struct {
	Force bool `cli:"short=f,long=force"`
}

type withCommands struct {
	Verbose bool   `cli:"short=v"`
	Add     cmdAdd `cli:"command=add,help='Add an item'"`
	Rm      cmdRm  `cli:"command=rm,alias=remove,help='Remove an item'"`
}

func TestCommandDispatch(t *testing.T) {
	parser := Construct[withCommands]().ToOptions()

	assertSuccess(t, withCommands{Add: cmdAdd{Name: "x"}}, parser, "add --name x")
	assertSuccess(t, withCommands{Rm: cmdRm{Force: true}}, parser, "rm -f")
}

func TestCommandAlias(t *testing.T) {
	parser := Construct[withCommands]().ToOptions()

	assertSuccess(t, withCommands{Rm: cmdRm{Force: true}}, parser, "remove -f")
}

func TestCommandAbsent(t *testing.T) {
	parser := Construct[withCommands]().ToOptions()

	// commands are optional: absent command leaves the field zero-valued
	assertSuccess(t, withCommands{Verbose: true}, parser, "-v")
}

func TestCommandInnerFailure(t *testing.T) {
	parser := Construct[withCommands]().ToOptions()

	assertFailure(t, parser, "expected `--name=NAME`, pass `--help` for usage information", "add")
}

func TestCommandHelp(t *testing.T) {
	parser := Construct[withCommands]().ToOptions()

	assertHelp(t, parser, `Usage: [-v] COMMAND ... COMMAND ...

Available options:
    -v
    -h, --help  Prints help information

Available commands:
    add         Add an item
    rm          Remove an item
`)
}

// --- maps ---

type withMaps struct {
	Labels map[string]string `cli:"long=label,metavar=LABEL:VALUE,help='Labels'"`
	Counts map[string]int    `cli:"long=count,default=a:1,default=b:2"`
}

func TestMapFieldParse(t *testing.T) {
	parser := Construct[withMaps]().ToOptions()

	assertSuccess(t, withMaps{
		Labels: map[string]string{"x": "1", "y": "2"},
		Counts: map[string]int{"a": 1, "b": 2},
	}, parser, "--label x:1 --label y:2 --count a:1 --count b:2")
	// map fields are optional: absence gives empty maps
	assertSuccess(t, withMaps{Counts: map[string]int{"a": 1, "b": 2}}, parser, "")
}

func TestMapFieldFailureNoColon(t *testing.T) {
	parser := Construct[withMaps]().ToOptions()

	assertFailure(t, parser, "couldn't parse `x`: expected `key:value`", "--label", "x")
}

func TestMapFieldFailureBadValue(t *testing.T) {
	parser := Construct[withMaps]().ToOptions()

	assertFailure(t, parser, "couldn't parse `x:abc`: invalid syntax", "--count", "x:abc")
}

func TestMapFieldHelp(t *testing.T) {
	parser := Construct[withMaps]().ToOptions()

	assertHelp(t, parser, `Usage: --label=<LABEL:VALUE> [--count=<KEY:VALUE>]

Available options:
        --label=<LABEL:VALUE>  Labels
        --count=<KEY:VALUE>
                             [default: map[a:1 b:2]]
    -h, --help               Prints help information
`)
}

// --- pointers ---

type withPointers struct {
	Opt    *int      `cli:"long=opt"`
	Ptrs   []*string `cli:"long=ptr"`
	DefPtr *int      `cli:"long=def,default=7"`
}

func TestPointerParse(t *testing.T) {
	parser := Construct[withPointers]().ToOptions()

	one := 1
	assertSuccess(t, withPointers{
		Opt:    &one,
		Ptrs:   []*string{strptr("a"), strptr("b")},
		DefPtr: iptr(7),
	}, parser, "--opt 1 --ptr a --ptr b")

	// absent pointer fields stay nil, absent slices are empty
	assertSuccess(t, withPointers{Ptrs: []*string{}, DefPtr: iptr(7)}, parser, "")

	seven := 7
	assertSuccess(t, withPointers{Opt: &seven, Ptrs: []*string{}, DefPtr: iptr(7)}, parser, "--opt 7")
}

func strptr(s string) *string { return &s }
func iptr(i int) *int         { return &i }

// --- choices ---

type withChoices struct {
	Animal string `cli:"long=animal,metavar=ANIMAL,help='A pet',choice=dog,choice=cat"`
}

func TestChoiceTagParse(t *testing.T) {
	parser := Construct[withChoices]().ToOptions()

	assertSuccess(t, withChoices{Animal: "dog"}, parser, "--animal dog")
	assertFailure(t, parser, "couldn't parse `zebra`: expected one of `cat`, `dog`",
		"--animal", "zebra")
}

func TestChoiceTagHelp(t *testing.T) {
	parser := Construct[withChoices]().ToOptions()

	assertHelp(t, parser, `Usage: --animal=ANIMAL

Available options:
        --animal=ANIMAL  A pet  (one of: cat, dog)
    -h, --help           Prints help information
`)
}

func TestChoiceTagCompletion(t *testing.T) {
	parser := Construct[withChoices]().ToOptions()

	assertCompletion(t, parser, "cat\tcat\t\tA pet\ndog\tdog\t\tA pet\n\n", "--animal", "")
}

// --- panics ---

func TestConstructFeaturePanics(t *testing.T) {
	type PlainStruct struct {
		Inner groupAuth
	}
	assertPanicsContains(t, func() { Construct[PlainStruct]() },
		"nested struct requires either a command=NAME or a group=NAME tag")

	type GroupNonStruct struct {
		X int `cli:"group=G"`
	}
	assertPanicsContains(t, func() { Construct[GroupNonStruct]() },
		"command and group tags are only valid for struct fields")

	type BothGroupCommand struct {
		X groupAuth `cli:"command=x,group=g"`
	}
	assertPanicsContains(t, func() { Construct[BothGroupCommand]() },
		"cannot be both a command and a group")

	type AliasWithoutCommand struct {
		X groupAuth `cli:"group=g,alias=a"`
	}
	assertPanicsContains(t, func() { Construct[AliasWithoutCommand]() },
		"alias is only valid for command fields")

	type ChoicePositional struct {
		X string `cli:"choice=dog"`
	}
	assertPanicsContains(t, func() { Construct[ChoicePositional]() },
		"choice is only valid for named fields")

	type ChoiceBool struct {
		X bool `cli:"short=x,choice=dog"`
	}
	assertPanicsContains(t, func() { Construct[ChoiceBool]() },
		"choice is not valid for bool fields")

	type ChoiceCount struct {
		X Count `cli:"short=x,choice=1"`
	}
	assertPanicsContains(t, func() { Construct[ChoiceCount]() },
		"choice is not valid for Count fields")

	type MapPositional struct {
		X map[string]string
	}
	assertPanicsContains(t, func() { Construct[MapPositional]() },
		"map field requires a short or a long name")

	type MapWithEnv struct {
		X map[string]string `cli:"long=x,env=X"`
	}
	assertPanicsContains(t, func() { Construct[MapWithEnv]() },
		"env is not valid for map fields")

	type MapWithChoice struct {
		X map[string]string `cli:"long=x,choice=a"`
	}
	assertPanicsContains(t, func() { Construct[MapWithChoice]() },
		"choice is not valid for map fields")

	type MultipleDefaults struct {
		X string `cli:"long=x,default=a,default=b"`
	}
	assertPanicsContains(t, func() { Construct[MultipleDefaults]() },
		"multiple default values are only valid for map fields")

	type GroupWithMetavar struct {
		X groupAuth `cli:"group=g,metavar=X"`
	}
	assertPanicsContains(t, func() { Construct[GroupWithMetavar]() },
		"metavar is not valid for struct fields")

	type PtrBool struct {
		X *bool `cli:"short=x"`
	}
	assertPanicsContains(t, func() { Construct[PtrBool]() },
		"*bool and *cli.Count fields are not supported")

	type NestedPtr struct {
		X **int `cli:"long=x"`
	}
	assertPanicsContains(t, func() { Construct[NestedPtr]() }, "nested")

	type OptPtr struct {
		X Opt[*int] `cli:"long=x"`
	}
	assertPanicsContains(t, func() { Construct[OptPtr]() }, "nested")
}
