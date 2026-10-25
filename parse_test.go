package cli

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/rprtr258/fun"
	"github.com/stretchr/testify/require"
)

// A port of the go-arg parse tests from the pre-fluent-API `parse_test.go`,
// adapted to the current `Construct` API: `cli` tags, named string types with
// `UnmarshalText` instead of struct types with `UnmarshalFlag`, `*E` and
// `cli.Opt[E]` for optional fields and panics for misconfiguration which used
// to be parse failures.
//
// Old expectations that no longer hold are annotated; assertions follow the
// current behavior.

func TestDuplicateShortFlags(t *testing.T) {
	// Duplicate short names between fields are currently not detected: the
	// second field is silently shadowed by the first one.
	type cmd struct {
		Verbose   bool     `cli:"short=v"`
		Variables []string `cli:"short=v"`
	}

	parser := Construct[cmd]().ToOptions()
	v, err := parser.RunE([]string{"prog", "-v", "a", "-v", "b"})
	require.EqualError(t, err, "`a` is not expected in this context")
	require.Equal(t, cmd{}, v)
}

func TestDuplicateLongFlags(t *testing.T) {
	// Duplicate long names between fields are currently not detected: the
	// second field is silently shadowed by the first one.
	type cmd struct {
		Test1 bool     `cli:"long=testing"`
		Test2 []string `cli:"long=testing"`
	}

	parser := Construct[cmd]().ToOptions()
	v, err := parser.RunE([]string{"prog", "--testing", "a", "--testing", "b"})
	require.EqualError(t, err, "`a` is not expected in this context")
	require.Equal(t, cmd{}, v)
}

type cmdOptions struct {
	Int        int `cli:"long=ii"`
	IntDefault int `cli:"long=id,default=1"`

	Float64        float64 `cli:"long=ff"`
	Float64Default float64 `cli:"long=fd,default=-3.14"`

	NumericFlag bool `cli:"short=3"`

	String        string `cli:"long=str"`
	StringDefault string `cli:"long=strd,default=abc"`

	Time        time.Duration `cli:"long=tt"`
	TimeDefault time.Duration `cli:"long=td,default=1m"`

	Map        map[string]int `cli:"long=mm"`
	MapDefault map[string]int `cli:"long=md,default=a:1"`

	// Slice defaults are not supported anymore, see TestNoSliceDefaults.
	Slice        []int `cli:"long=ss"`
	SliceDefault []int `cli:"long=sd"`
}

func TestDefaults(t *testing.T) {
	for name, test := range map[string]struct {
		command  string
		expected cmdOptions
	}{
		"no arguments for flags with default": {
			command: "--ii 3 --ff -2.71 --str def --tt 3ms",
			expected: cmdOptions{
				Int:        3,
				IntDefault: 1,

				Float64:        -2.71,
				Float64Default: -3.14,

				String:        "def",
				StringDefault: "abc",

				Time:        3 * time.Millisecond,
				TimeDefault: time.Minute,

				Map:          nil,
				MapDefault:   map[string]int{"a": 1},
				Slice:        []int{},
				SliceDefault: []int{},
			},
		},
		"overwriting default values": {
			command: "--ii 3 --id 3 --ff -2.71 --fd 2.71 -3 --str def --strd def" +
				" --tt 3ms --td 3ms --mm c:3 --md c:3 --ss 3 --sd 3",
			expected: cmdOptions{
				Int:        3,
				IntDefault: 3,

				Float64:        -2.71,
				Float64Default: 2.71,

				NumericFlag: true,

				String:        "def",
				StringDefault: "def",

				Time:        3 * time.Millisecond,
				TimeDefault: 3 * time.Millisecond,

				Map:        map[string]int{"c": 3},
				MapDefault: map[string]int{"c": 3},

				Slice:        []int{3},
				SliceDefault: []int{3},
			},
		},
		"overwriting default values with zero values": {
			command: "--ii 0 --id 0 --ff 0 --fd 0 --str '' --strd ''" +
				" --tt 0ms --td 0s --mm :0 --md :0 --ss 0 --sd 0",
			expected: cmdOptions{
				Map:          map[string]int{"": 0},
				MapDefault:   map[string]int{"": 0},
				Slice:        []int{0},
				SliceDefault: []int{0},
			},
		},
	} {
		t.Run(name, func(t *testing.T) {
			parser := Construct[cmdOptions]().ToOptions()
			assertSuccess(t, test.expected, parser, test.command)
		})
	}
}

func TestNoBoolDefaults(t *testing.T) {
	type cmd struct {
		DefaultBool bool `cli:"short=d,default=true"`
	}

	require.Equal(t,
		`cli.Construct: field "DefaultBool": bool field cannot have a default`,
		catchPanic[string](t, func() { Construct[cmd]() }))
}

func TestNoSliceDefaults(t *testing.T) {
	type cmd struct {
		Slice []int `cli:"long=sd,default=1"`
	}

	require.Equal(t,
		`cli.Construct: field "Slice": []E and cli.Opt[E] fields cannot have a default`,
		catchPanic[string](t, func() { Construct[cmd]() }))
}

// flagString is a custom flag type parsed with UnmarshalText.
type flagString string

func (c *flagString) UnmarshalText(s []byte) error {
	if string(s) != "-1" && string(s) != "-foo" {
		return fmt.Errorf("invalid flag value, expected -1 or -foo")
	}
	*c = flagString(s)
	return nil
}

type cmdCustomFlag struct {
	Custom flagString `cli:"long=custom,short=c"`
}

func TestCustomFlagFail(t *testing.T) {
	for name, test := range map[string]struct {
		command string
		errWant string
	}{
		"not provided": {
			command: "",
			errWant: "expected `-c=CUSTOM`, pass `--help` for usage information",
		},
		"invalid value": {
			command: "--custom=-2",
			errWant: "couldn't parse `-2`: invalid flag value, expected -1 or -foo",
		},
	} {
		t.Run(name, func(t *testing.T) {
			parser := Construct[cmdCustomFlag]().ToOptions()
			assertFailure(t, parser, test.errWant, test.command)
		})
	}
}

func TestCustomFlagSuccess(t *testing.T) {
	for name, test := range map[string]struct {
		command string
		custom  flagString
	}{
		"valid value #1": {"--custom=-foo", flagString("-foo")},
		"valid value #2": {"--custom=-1", flagString("-1")},
	} {
		t.Run(name, func(t *testing.T) {
			parser := Construct[cmdCustomFlag]().ToOptions()
			assertSuccess(t, cmdCustomFlag{Custom: test.custom}, parser, test.command)
		})
	}
}

type cmdUnknownFlags struct {
	Flag1 string `cli:"long=flag1"`
	Flag2 string `cli:"long=flag2"`
}

func TestUnknownFlag(t *testing.T) {
	parser := Construct[cmdUnknownFlags]().ToOptions()
	assertFailure(t, parser,
		"`--unknownFlagX` is not expected in this context",
		"--flag1 stuff --unknownFlagX blah --flag2 foo")
}

// flagEnum is a custom enum flag type parsed with UnmarshalText.
type flagEnum int

const (
	flagEnumV1 flagEnum = iota
	flagEnumV2
)

func (f *flagEnum) UnmarshalText(s []byte) error {
	switch string(s) {
	case "v1":
		*f = flagEnumV1
	case "v2":
		*f = flagEnumV2
	default:
		return fmt.Errorf("allowed values: v1, v2")
	}
	return nil
}

type cmdEnum struct {
	Choice flagEnum `cli:"long=choose"`
}

func TestEnum(t *testing.T) {
	parser := Construct[cmdEnum]().ToOptions()
	assertFailure(t, parser,
		"couldn't parse `invalid`: allowed values: v1, v2",
		"--choose invalid")
	assertSuccess(t, cmdEnum{Choice: flagEnumV2}, parser, "--choose v2")
}

type cmdEnumFallback struct {
	Choice flagEnum `cli:"long=choose,default=v2"`
}

func TestEnumFallback(t *testing.T) {
	parser := Construct[cmdEnumFallback]().ToOptions()
	assertSuccess(t, cmdEnumFallback{Choice: flagEnumV2}, parser, "")
}

// Unexported embedded structs are silently skipped: they are unexported fields
// without a tag, and only exported fields with tags become parsers.
type cmdEmbedded struct {
	embedded
}

type embedded struct {
	V bool `cli:"short=v"`
}

func TestEmbeddedSkipped(t *testing.T) {
	parser := Construct[cmdEmbedded]().ToOptions()
	assertFailure(t, parser, "`-v` is not expected in this context", "-v")
}

type cmdArgs struct {
	Value   bool `cli:"short=v"`
	ArgList []string
}

func TestCommandHandlerNoCommand(t *testing.T) {
	parser := Construct[cmdArgs]().ToOptions()
	assertSuccess(t, cmdArgs{ArgList: []string{"arg1", "arg2"}}, parser, "arg1 arg2")
}

type cmdSubcommand struct {
	Value   bool `cli:"short=v"`
	ArgList []string
}

type cmdWithCommand struct {
	Command cmdSubcommand `cli:"command=cmd"`
}

func TestCommandHandler(t *testing.T) {
	parser := Construct[cmdWithCommand]().ToOptions()
	assertSuccess(t, cmdWithCommand{Command: cmdSubcommand{ArgList: []string{"arg1", "arg2"}}}, parser, "cmd arg1 arg2")
}

// marshalled is a custom type parsed with UnmarshalText.
type marshalled string

func (m *marshalled) UnmarshalText(value []byte) error {
	switch string(value) {
	case "yes":
		*m = "true"
	case "no":
		*m = "false"
	default:
		return fmt.Errorf(`%q is not a valid value, please specify "yes" or "no"`, string(value))
	}
	return nil
}

type cmdCustomFlag2 struct {
	Value marshalled `cli:"short=v"`
}

func TestUnmarshal(t *testing.T) {
	parser := Construct[cmdCustomFlag2]().ToOptions()
	assertSuccess(t, cmdCustomFlag2{Value: "true"}, parser, "-v yes")
	assertFailure(t, parser,
		"`-v` requires an argument `VALUE`",
		"-v")
	assertFailure(t, parser,
		"couldn't parse `invalid`: \"invalid\" is not a valid value, please specify \"yes\" or \"no\"",
		"-v invalid")
}

type cmdCustomFlag2Default struct {
	Value marshalled `cli:"short=v,default=yes"`
}

func TestUnmarshalDefault(t *testing.T) {
	parser := Construct[cmdCustomFlag2Default]().ToOptions()
	assertSuccess(t, cmdCustomFlag2Default{Value: "true"}, parser, "")
}

type cmdCustomPositional struct {
	Value marshalled
}

// Custom positional parsers are implemented now: the old test was skipped
// until "custom positionals are implemented".
func TestUnmarshalPositionalError(t *testing.T) {
	parser := Construct[cmdCustomPositional]().ToOptions()
	assertFailure(t, parser,
		"couldn't parse `invalid`: \"invalid\" is not a valid value, please specify \"yes\" or \"no\"",
		"invalid")
}

type cmdRootHelp struct {
	Subcommand cmdSubcommand `cli:"command=sub"`
}

func TestRootHelp(t *testing.T) {
	parser := Construct[cmdRootHelp]().ToOptions()
	assertHelp(t, parser, strings.Join([]string{
		"Usage: COMMAND ...",
		"",
		"Available options:",
		"    -h, --help  Prints help information",
		"",
		"Available commands:",
		"    sub",
		"",
	}, "\n"), "--help")
}

type cmdPositionalsV0 struct {
	Value   bool `cli:"short=v"`
	ArgList []string
}

func TestPassDoubleDash(t *testing.T) {
	parser := Construct[cmdPositionalsV0]().ToOptions()
	assertSuccess(t, cmdPositionalsV0{Value: true, ArgList: []string{"-v", "-g"}}, parser, "-v -- -v -g")
}

// Positionals no longer switch the parser into a pass-through mode: options
// are still parsed after the first positional, only `--` enables pass-through
// (see TestPassDoubleDash).
func TestPassAfterNonOption(t *testing.T) {
	parser := Construct[cmdPositionalsV0]().ToOptions()
	assertSuccess(t, cmdPositionalsV0{Value: true, ArgList: []string{"arg"}}, parser, "arg -v")
	assertFailure(t, parser, "argument `-v` cannot be used multiple times in this context", "-v arg -v -g")
}

type cmdPositionalsV struct {
	Value bool `cli:"short=v"`
	Rest  []string
}

func TestPassAfterNonOptionWithPositional(t *testing.T) {
	parser := Construct[cmdPositionalsV]().ToOptions()
	assertSuccess(t, cmdPositionalsV{Value: true, Rest: []string{"arg"}}, parser, "arg -v")
	assertFailure(t, parser, "argument `-v` cannot be used multiple times in this context", "-v arg -v -g")
}

type cmdPositionalsVInts struct {
	Value   bool `cli:"short=v"`
	ArgList []int
}

func TestPassAfterNonOptionWithPositionalIntPass(t *testing.T) {
	parser := Construct[cmdPositionalsVInts]().ToOptions()
	assertSuccess(t, cmdPositionalsVInts{Value: true, ArgList: []int{1, 2, 3}}, parser, "-v 1 2 3")
}

func TestPassAfterNonOptionWithPositionalIntFail(t *testing.T) {
	parser := Construct[cmdPositionalsVInts]().ToOptions()
	assertFailure(t, parser,
		"couldn't parse `notint1`: invalid syntax",
		"-v notint1 notint2 notint3")
	assertFailure(t, parser,
		"couldn't parse `notint2`: invalid syntax",
		"-v 1 notint2 notint3")
}

// Optional flags use `cli.Opt[E]` instead of pointers to bool: `*bool` fields
// are rejected.
type cmdPtrBool struct {
	Value Opt[bool] `cli:"short=v"`
}

func TestPointerBool(t *testing.T) {
	parser := Construct[cmdPtrBool]().ToOptions()
	assertSuccess(t, cmdPtrBool{Value: Valid(true)}, parser, "-v")
}

type cmdPtrString struct {
	Value Opt[string] `cli:"short=v"`
}

func TestPointerString(t *testing.T) {
	parser := Construct[cmdPtrString]().ToOptions()
	assertSuccess(t, cmdPtrString{Value: Valid("value")}, parser, "-v value")
}

func TestPointerSlice(t *testing.T) {
	type cmd struct {
		Value *[]string `cli:"short=v"`
	}

	require.Equal(t,
		`cli.Construct: field "Value": nested []string is not supported, use a single []E, []*E, *E or cli.Opt[E] wrapper`,
		catchPanic[string](t, func() { Construct[cmd]() }))
}

func TestPointerMap(t *testing.T) {
	type cmd struct {
		Value *map[string]int `cli:"short=v"`
	}

	require.Equal(t,
		`cli.Construct: field "Value": unsupported field type map[string]int, implement encoding.TextUnmarshaler to parse custom types`,
		catchPanic[string](t, func() { Construct[cmd]() }))
}

// marshalledString is a custom type parsed with UnmarshalText.
type marshalledString string

func (m *marshalledString) UnmarshalText(value []byte) error {
	*m = marshalledString(value)
	return nil
}

type cmdPtrCustomString struct {
	Value Opt[marshalledString] `cli:"short=v"`
}

func TestPointerStringMarshalled(t *testing.T) {
	parser := Construct[cmdPtrCustomString]().ToOptions()
	assertSuccess(t, cmdPtrCustomString{Value: Valid(marshalledString("value"))}, parser, "-v value")
}

// flagStruct is a custom struct type parsed with UnmarshalText.
type flagStruct struct {
	Value string
}

func (m *flagStruct) UnmarshalText(value []byte) error {
	m.Value = string(value)
	return nil
}

// Pointer to a struct type with UnmarshalText works as a leaf parser.
type cmdPtrCustomStruct struct {
	Value *flagStruct `cli:"short=v"`
}

func TestPointerStructMarshalled(t *testing.T) {
	parser := Construct[cmdPtrCustomStruct]().ToOptions()
	v, err := parser.RunE([]string{"prog", "-v", "value"})
	require.Nil(t, err)
	require.Equal(t, "value", v.Value.Value)
}

// A struct type without UnmarshalText cannot be a leaf parser.
func TestPointerGroup(t *testing.T) {
	type PointerGroup struct {
		Value bool `cli:"short=v"`
	}

	type cmd struct {
		Value *PointerGroup `cli:"short=v"`
	}

	require.Equal(t,
		`cli.Construct: field "Value": unsupported field type cli.PointerGroup, implement encoding.TextUnmarshaler to parse custom types`,
		catchPanic[string](t, func() { Construct[cmd]() }))
}

// Untagged scalar fields become positionals: they no longer fail with
// "flag must have short or long name".
func TestDoNotChangeNonTaggedFields(t *testing.T) {
	type opts struct {
		A *int
	}

	parser := Construct[opts]().ToOptions()
	assertSuccess(t, opts{A: fun.Ptr(42)}, parser, "42")
	assertSuccess(t, opts{}, parser, "")
}

type cmdShortBool struct {
	Value bool `cli:"short=v"`
}

func TestShort(t *testing.T) {
	parser := Construct[cmdShortBool]().ToOptions()
	assertSuccess(t, cmdShortBool{Value: true}, parser, "-v")
	assertSuccess(t, cmdShortBool{}, parser, "")
}

type cmdShorts struct {
	V   bool   `cli:"short=v"`
	O   bool   `cli:"short=o"`
	F   bool   `cli:"short=f"`
	Str string `cli:"short=s,default=aboba"`
}

// Short flags can be concatenated, see also TestShortWithEqualArg.
func TestShortMultiConcat(t *testing.T) {
	parser := Construct[cmdShorts]().ToOptions()
	assertSuccess(t, cmdShorts{V: true, O: true, F: true, Str: "aboba"}, parser, "-vo -f")
	assertSuccess(t, cmdShorts{V: true, Str: "f"}, parser, "-vs f")
}

type cmdShortString struct {
	Value string `cli:"short=v"`
}

func TestShortWithEqualArg(t *testing.T) {
	parser := Construct[cmdShortString]().ToOptions()
	assertSuccess(t, cmdShortString{Value: "value"}, parser, "-v value")
	assertSuccess(t, cmdShortString{Value: "value"}, parser, "-vvalue")
}
