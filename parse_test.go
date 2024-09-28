package cli

// import (
// 	"errors"
// 	"fmt"
// 	"strings"
// 	"testing"
// 	"time"

// 	"github.com/stretchr/testify/require"
// )

// func TestDuplicateShortFlags(t *testing.T) {
// 	type cmd struct {
// 		Verbose   bool     `short:"v"`
// 		Variables []string `short:"v"`
// 	}

// 	assertFail[cmd](t, `flag shortname 'v' is duplicated`)
// }

// func TestDuplicateLongFlags(t *testing.T) {
// 	type cmd struct {
// 		Test1 bool     `long:"testing"`
// 		Test2 []string `long:"testing"`
// 	}

// 	assertFail[cmd](t, `flag longname "testing" is duplicated`)
// }

// type cmdOptions struct {
// 	Int        int `long:"ii"`
// 	IntDefault int `long:"id" default:"1"`

// 	Float64        float64 `long:"ff"`
// 	Float64Default float64 `long:"fd" default:"-3.14"`

// 	NumericFlag bool `short:"3"`

// 	String        string `long:"str"`
// 	StringDefault string `long:"strd" default:"abc"`

// 	Time        time.Duration `long:"tt"`
// 	TimeDefault time.Duration `long:"td" default:"1m"`

// 	Map        map[string]int `long:"mm"`
// 	MapDefault map[string]int `long:"md" default:"a:1"`

// 	Slice        []int `long:"ss"`
// 	SliceDefault []int `long:"sd" default:"1" default:"2"`
// }

// func TestDefaults(t *testing.T) {
// 	for name, test := range map[string]struct {
// 		args     []string
// 		expected cmdOptions
// 	}{
// 		"no arguments for flags with default": {
// 			args: []string{
// 				"--ii", "3",
// 				"--ff", "-2.71",
// 				"--str", "def",
// 				"--tt", "3ms",
// 			},
// 			expected: cmdOptions{
// 				Int:        3,
// 				IntDefault: 1,

// 				Float64:        -2.71,
// 				Float64Default: -3.14,

// 				String:        "def",
// 				StringDefault: "abc",

// 				Time:        3 * time.Millisecond,
// 				TimeDefault: time.Minute,

// 				Map:        nil,
// 				MapDefault: map[string]int{"a": 1},

// 				Slice:        nil,
// 				SliceDefault: []int{1, 2},
// 			},
// 		},
// 		"overwriting default values": {
// 			args: []string{
// 				"--ii", "3",
// 				"--id", "3",
// 				"--ff", "-2.71",
// 				"--fd", "2.71",
// 				"-3",
// 				"--str", "def",
// 				"--strd", "def",
// 				"--tt", "3ms",
// 				"--td", "3ms",
// 				"--mm", "c:3",
// 				"--md", "c:3",
// 				"--ss", "3",
// 				"--sd", "3",
// 			},
// 			expected: cmdOptions{
// 				Int:        3,
// 				IntDefault: 3,

// 				Float64:        -2.71,
// 				Float64Default: 2.71,

// 				NumericFlag: true,

// 				String:        "def",
// 				StringDefault: "def",

// 				Time:        3 * time.Millisecond,
// 				TimeDefault: 3 * time.Millisecond,

// 				Map:        map[string]int{"c": 3},
// 				MapDefault: map[string]int{"c": 3},

// 				Slice:        []int{3},
// 				SliceDefault: []int{3},
// 			},
// 		},
// 		"overwriting default values with zero values": {
// 			args: []string{
// 				"--ii", "0",
// 				"--id", "0",
// 				"--ff", "0",
// 				"--fd", "0",
// 				"--str", "",
// 				"--strd", "",
// 				"--tt", "0ms",
// 				"--td", "0s",
// 				"--mm", ":0",
// 				"--md", ":0",
// 				"--ss", "0",
// 				"--sd", "0",
// 			},
// 			expected: cmdOptions{
// 				Map:        map[string]int{"": 0},
// 				MapDefault: map[string]int{"": 0},

// 				Slice:        []int{0},
// 				SliceDefault: []int{0},
// 			},
// 		},
// 	} {
// 		test := test
// 		t.Run(name, func(t *testing.T) {
// 			assertSuccess[cmdOptions](t, func(x cmdOptions) {
// 				require.Equal(t, test.expected, x)
// 			}, test.args...)
// 		})
// 	}
// }

// func TestNoDefaultsForBools(t *testing.T) {
// 	type cmd struct {
// 		DefaultBool bool `short:"d" default:"true"`
// 	}

// 	assertFail[cmd](t, `parse flag "DefaultBool" bool flag cannot have default value, since then it always will be true`)
// }

// type flagString struct {
// 	Value string
// }

// func (c *flagString) UnmarshalFlag(s string) error {
// 	if s != "-1" && s != "-foo" {
// 		return errors.New("invalid flag value, expected -1 or -foo")
// 	}

// 	c.Value = s
// 	return nil
// }

// type cmdCustomFlag struct {
// 	Custom flagString `long:"custom" short:"c"`
// }

// func TestCustomFlagFail(t *testing.T) {
// 	for name, test := range map[string]struct {
// 		args    []string
// 		errWant string
// 	}{
// 		"not provided": {
// 			args:    []string{},
// 			errWant: "invalid -c/--custom flag value: missing value",
// 		},
// 		"invalid value": {
// 			args:    []string{"--custom", "-2"},
// 			errWant: "invalid -c/--custom flag value: invalid flag value, expected -1 or -foo",
// 		},
// 	} {
// 		test := test
// 		t.Run(name, func(t *testing.T) {
// 			assertFail[cmdCustomFlag](t, test.errWant, test.args...)
// 		})
// 	}
// }

// func TestCustomFlagSuccess(t *testing.T) {
// 	for _, args := range map[string][]string{
// 		"valid value #1": {
// 			"--custom", "-foo",
// 		},
// 		"valid value #2": {
// 			"--custom", "-1",
// 		},
// 	} {
// 		assertSuccess[cmdCustomFlag](t, func(x cmdCustomFlag) {}, args...)
// 	}
// }

// // TODO: maybe add map with all unknown flag left
// type cmdUnknownFlags struct {
// 	Flag1 string `long:"flag1"`
// 	Flag2 string `long:"flag2"`
// }

// func TestUnknownFlag(t *testing.T) {
// 	assertFail[cmdUnknownFlags](t,
// 		`extraneous positional arguments found: ["--unknownFlagX", "blah", "--flag2", "foo"]`,
// 		"--flag1", "stuff", "--unknownFlagX", "blah", "--flag2", "foo")
// }

// type flagEnum int

// const (
// 	flagEnumV1 flagEnum = iota
// 	flagEnumV2
// )

// func (f *flagEnum) UnmarshalFlag(s string) error {
// 	switch s {
// 	case "v1":
// 		*f = flagEnumV1
// 	case "v2":
// 		*f = flagEnumV2
// 	default:
// 		return fmt.Errorf("allowed values: v1, v2")
// 	}
// 	return nil
// }

// type cmdEnum struct {
// 	Choice flagEnum `long:"choose"`
// }

// func TestEnum(t *testing.T) {
// 	assertFail[cmdEnum](t,
// 		`invalid --choose flag value: allowed values: v1, v2`,
// 		// "Invalid value `invalid' for fun.Option `--choose'. Allowed values are: v1 or v2",
// 		"--choose", "invalid")
// 	assertSuccess[cmdEnum](t,
// 		func(x cmdEnum) {
// 			require.Equal(t, flagEnumV2, x.Choice)
// 		}, "--choose", "v2")
// }

// type embedded struct {
// 	V bool `short:"v"`
// }
// type cmdEmbedded struct {
// 	embedded
// }

// func TestEmbedded(t *testing.T) {
// 	assertSuccess[cmdEmbedded](t, func(x cmdEmbedded) {
// 		require.True(t, x.V)
// 	}, "-v")
// }

// type cmdArgs struct {
// 	Value bool     `short:"v"`
// 	Args  []string `positional-args:"yes"`
// }

// func TestCommandHandlerNoCommand(t *testing.T) {
// 	assertSuccess[cmdArgs](t, func(x cmdArgs) {
// 		require.False(t, x.Value)
// 		require.Equal(t, []string{"arg1", "arg2"}, x.Args)
// 	}, "arg1", "arg2")
// }

// type cmdSubcommand struct {
// 	Value bool     `short:"v"`
// 	Args  []string `positional-args:"yes"`
// }

// func TestCommandHandler(t *testing.T) {
// 	type cmd struct {
// 		Command cmdSubcommand `command:"cmd"`
// 	}

// 	assertSuccess[cmd](t, func(x cmdSubcommand) {
// 		require.False(t, x.Value)
// 		require.Equal(t, []string{"arg1", "arg2"}, x.Args)
// 	}, "cmd", "arg1", "arg2")
// }

// type marshalled string

// func (m *marshalled) UnmarshalFlag(value string) error {
// 	switch value {
// 	case "yes":
// 		*m = "true"
// 	case "no":
// 		*m = "false"
// 	default:
// 		return fmt.Errorf(`%q is not a valid value, please specify "yes" or "no"`, value)
// 	}
// 	return nil
// }

// type cmdCustomFlag2 struct {
// 	Value marshalled `short:"v"`
// }

// func TestUnmarshal(t *testing.T) {
// 	assertSuccess[cmdCustomFlag2](t, func(x cmdCustomFlag2) {
// 		require.Equal(t, marshalled("true"), x.Value)
// 	}, "-v", "yes")
// 	assertFail[cmdCustomFlag2](t,
// 		`expected argument for flag -v`,
// 		"-v")
// 	assertFail[cmdCustomFlag2](t,
// 		`invalid -v flag value: "invalid" is not a valid value, please specify "yes" or "no"`,
// 		"-v", "invalid")
// }

// type cmdCustomFlag2Default struct {
// 	Value marshalled `short:"v" default:"yes"`
// }

// func TestUnmarshalDefault(t *testing.T) {
// 	assertSuccess[cmdCustomFlag2Default](t, func(x cmdCustomFlag2Default) {
// 		require.Equal(t, marshalled("true"), x.Value)
// 	})
// }

// type cmdCustomPositional struct {
// 	Args struct {
// 		Value marshalled
// 	} `positional-args:"yes"`
// }

// func TestUnmarshalPositionalError(t *testing.T) {
// 	// TODO: get back when custom positionals are implemented
// 	t.Skip()

// 	assertFail[cmdCustomPositional](t,
// 		`"invalid" is not a valid value, please specify "yes" or "no"`,
// 		"invalid",
// 	)
// }

// type cmdRootHelp struct {
// 	Subcommand cmdSubcommand `command:"sub"`
// }

// func TestRootHelp(t *testing.T) {
// 	assertFail[cmdRootHelp](t, strings.Join([]string{
// 		"Usage: \x1b[94mx\x1b[0m \x1b[91m[COMMAND]\x1b[0m \x1b[93m[FLAGS]\x1b[0m",
// 		"",
// 		"",
// 		"",
// 		"Commands:",
// 		"\t\x1b[91msub\x1b[0m   ",
// 		"",
// 		"Flags:",
// 		"\t\x1b[93m-h\x1b[0m, \x1b[93m--help\x1b[0m    Print this help message",
// 		"",
// 	}, "\n"), "--help")
// }

// type cmdPositionalsV0 struct {
// 	Value bool     `short:"v"`
// 	Args  []string `positional-args:"yes"`
// }

// func TestPassDoubleDash(t *testing.T) {
// 	assertSuccess[cmdPositionalsV0](t, func(x cmdPositionalsV0) {
// 		require.True(t, x.Value)
// 		require.Equal(t, []string{"-v", "-g"}, x.Args)
// 	}, "-v", "--", "-v", "-g")
// }

// func TestPassAfterNonOption(t *testing.T) {
// 	assertSuccess[cmdPositionalsV0](t, func(x cmdPositionalsV0) {
// 		require.True(t, x.Value)
// 		require.Equal(t, []string{"arg", "-v", "-g"}, x.Args)
// 	}, "-v", "arg", "-v", "-g")
// }

// type cmdPositionalsV struct {
// 	Value      bool `short:"v"`
// 	Positional struct {
// 		Rest []string
// 	} `positional-args:"yes"`
// }

// func TestPassAfterNonOptionWithPositional(t *testing.T) {
// 	assertSuccess[cmdPositionalsV](t, func(x cmdPositionalsV) {
// 		require.True(t, x.Value)
// 		require.Equal(t, []string{"arg", "-v", "-g"}, x.Positional.Rest)
// 	}, "-v", "arg", "-v", "-g")
// }

// type cmdPositionalsVInts struct {
// 	Value bool  `short:"v"`
// 	Args  []int `positional-args:"yes"`
// }

// func TestPassAfterNonOptionWithPositionalIntPass(t *testing.T) {
// 	assertSuccess[cmdPositionalsVInts](t, func(x cmdPositionalsVInts) {
// 		require.True(t, x.Value)
// 		require.Equal(t, []int{1, 2, 3}, x.Args)
// 	}, "-v", "1", "2", "3")
// }

// type cmdPositionalsVInts2 struct {
// 	Value      bool `short:"v"`
// 	Positional struct {
// 		Args []int
// 	} `positional-args:"yes"`
// }

// func TestPassAfterNonOptionWithPositionalIntPass2(t *testing.T) {
// 	assertSuccess[cmdPositionalsVInts2](t, func(x cmdPositionalsVInts2) {
// 		require.True(t, x.Value)
// 		require.Equal(t, []int{1, 2, 3}, x.Positional.Args)
// 	}, "-v", "1", "2", "3")
// }

// func TestPassAfterNonOptionWithPositionalIntFail(t *testing.T) {
// 	for _, test := range []struct {
// 		args  []string
// 		error string
// 	}{
// 		{
// 			[]string{"-v", "notint1", "notint2", "notint3"},
// 			`invalid rest of positional arguments: slice element 0: expected int, found "notint1"`,
// 		},
// 		{
// 			[]string{"-v", "1", "notint2", "notint3"},
// 			`invalid rest of positional arguments: slice element 1: expected int, found "notint2"`,
// 		},
// 	} {
// 		assertFail[cmdPositionalsVInts](t, test.error, test.args...)
// 	}
// }

// type cmdPtrBool struct {
// 	Value *bool `short:"v"`
// }

// func TestPointerBool(t *testing.T) {
// 	assertSuccess[cmdPtrBool](t, func(x cmdPtrBool) {
// 		require.True(t, *x.Value)
// 	}, "-v")
// }

// type cmdPtrString struct {
// 	Value *string `short:"v"`
// }

// func TestPointerString(t *testing.T) {
// 	assertSuccess[cmdPtrString](t, func(x cmdPtrString) {
// 		require.Equal(t, "value", *x.Value)
// 	}, "-v", "value")
// }

// func TestPointerSlice(t *testing.T) {
// 	type cmd struct {
// 		Value *[]string `short:"v"`
// 	}

// 	assertFail[cmd](t,
// 		`parse flag "Value": pointer flag cannot be slice, use slice type directly instead`,
// 		"-v", "value1", "-v", "value2")
// }

// func TestPointerMap(t *testing.T) {
// 	type cmd struct {
// 		Value *map[string]int `short:"v"`
// 	}

// 	assertFail[cmd](t,
// 		`parse flag "Value": pointer flag cannot be map, use map type directly instead`,
// 		"-v", "k1:2", "-v", "k2:-5")
// }

// type marshalledString string

// func (m *marshalledString) UnmarshalFlag(value string) error {
// 	*m = marshalledString(value)
// 	return nil
// }

// type cmdPtrCustomString struct {
// 	Value *marshalledString `short:"v"`
// }

// func TestPointerStringMarshalled(t *testing.T) {
// 	assertSuccess[cmdPtrCustomString](t, func(x cmdPtrCustomString) {
// 		require.Equal(t, "value", string(*x.Value))
// 	}, "-v", "value")
// }

// type flagStruct struct {
// 	Value string
// }

// func (m *flagStruct) UnmarshalFlag(value string) error {
// 	m.Value = value
// 	return nil
// }

// type cmdPtrCustomStruct struct {
// 	Value *flagStruct `short:"v"`
// }

// func TestPointerStructMarshalled(t *testing.T) {
// 	assertSuccess[cmdPtrCustomStruct](t, func(x cmdPtrCustomStruct) {
// 		require.Equal(t, "value", x.Value.Value)
// 	}, "-v", "value")
// }

// type PointerGroup struct {
// 	Value bool `short:"v"`
// }

// type cmdPtrEmbedded struct {
// 	*PointerGroup
// }

// func TestPointerGroup(t *testing.T) {
// 	// TODO: embedded not-structs are unsupported now
// 	t.Skip()

// 	assertSuccess[cmdPtrEmbedded](t, func(x cmdPtrEmbedded) {
// 		require.True(t, x.Value)
// 	}, "-v")
// }

// func TestDoNotChangeNonTaggedFields(t *testing.T) {
// 	type opts struct {
// 		A *int
// 	}

// 	assertFail[opts](t, `parse flag "A": flag must have short or long name`)
// }

// type cmdShortBool struct {
// 	Value bool `short:"v"`
// }

// func TestShort(t *testing.T) {
// 	assertSuccess[cmdShortBool](t, func(x cmdShortBool) {
// 		require.True(t, x.Value)
// 	}, "-v")
// 	assertSuccess[cmdShortBool](t, func(x cmdShortBool) {
// 		require.False(t, x.Value)
// 	})
// }

// type cmdShorts struct {
// 	V   bool   `short:"v"`
// 	O   bool   `short:"o"`
// 	F   bool   `short:"f"`
// 	Str string `short:"s" default:"aboba"`
// }

// func TestShortMultiConcat(t *testing.T) {
// 	// TODO: get back when short flags parsing is fixed
// 	t.Skip()

// 	assertSuccess[cmdShorts](t, func(x cmdShorts) {
// 		require.True(t, x.V)
// 		require.True(t, x.O)
// 		require.True(t, x.F)
// 	}, "-vo", "-f")
// 	assertSuccess[cmdShorts](t, func(x cmdShorts) {
// 		require.Equal(t, "f", x.Str)
// 	}, "-vs", "f")
// }

// type cmdShortString struct {
// 	Value string `short:"v"`
// }

// func TestShortWithEqualArg(t *testing.T) {
// 	// TODO: get back when short flags parsing is fixed
// 	t.Skip()

// 	assertSuccess[cmdShortString](t, func(x cmdShortString) {
// 		require.Equal(t, "value", x.Value)
// 	}, "-v", "value")
// 	assertSuccess[cmdShortString](t, func(x cmdShortString) {
// 		require.Equal(t, "value", x.Value)
// 	}, "-vvalue")
// }
