package flags

// import (
// 	"errors"
// 	"runtime"
// 	"strconv"
// 	"strings"
// 	"testing"
// 	"time"

// 	"github.com/stretchr/testify/assert"
// )

// type defaultOptions struct {
// 	Int        int `long:"i"`
// 	IntDefault int `long:"id" default:"1"`

// 	Float64        float64 `long:"f"`
// 	Float64Default float64 `long:"fd" default:"-3.14"`

// 	NumericFlag bool `short:"3"`

// 	String            string `long:"str"`
// 	StringDefault     string `long:"strd" default:"abc"`
// 	StringNotUnquoted string `long:"strnot" unquote:"false"`

// 	Time        time.Duration `long:"t"`
// 	TimeDefault time.Duration `long:"td" default:"1m"`

// 	Map        map[string]int `long:"m"`
// 	MapDefault map[string]int `long:"md" default:"a:1"`

// 	Slice        []int `long:"s"`
// 	SliceDefault []int `long:"sd" default:"1" default:"2"`
// }

// func TestDefaults(t *testing.T) {
// 	for _, test := range []struct {
// 		msg      string
// 		args     []string
// 		expected defaultOptions
// 	}{
// 		{
// 			msg:  "no arguments, expecting default values",
// 			args: []string{},
// 			expected: defaultOptions{
// 				Int:        0,
// 				IntDefault: 1,

// 				Float64:        0.0,
// 				Float64Default: -3.14,

// 				NumericFlag: false,

// 				String:        "",
// 				StringDefault: "abc",

// 				Time:        0,
// 				TimeDefault: time.Minute,

// 				Map:        map[string]int{},
// 				MapDefault: map[string]int{"a": 1},

// 				Slice:        []int{},
// 				SliceDefault: []int{1, 2},
// 			},
// 		},
// 		{
// 			msg:  "non-zero value arguments, expecting overwritten arguments",
// 			args: []string{"--i=3", "--id=3", "--f=-2.71", "--fd=2.71", "-3", "--str=def", "--strd=def", "--t=3ms", "--td=3ms", "--m=c:3", "--md=c:3", "--s=3", "--sd=3"},
// 			expected: defaultOptions{
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
// 		{
// 			msg:  "zero value arguments, expecting overwritten arguments",
// 			args: []string{"--i=0", "--id=0", "--f=0", "--fd=0", "--str", "", "--strd=\"\"", "--t=0ms", "--td=0s", "--m=:0", "--md=:0", "--s=0", "--sd=0"},
// 			expected: defaultOptions{
// 				Int:        0,
// 				IntDefault: 0,

// 				Float64:        0,
// 				Float64Default: 0,

// 				String:        "",
// 				StringDefault: "",

// 				Time:        0,
// 				TimeDefault: 0,

// 				Map:        map[string]int{"": 0},
// 				MapDefault: map[string]int{"": 0},

// 				Slice:        []int{0},
// 				SliceDefault: []int{0},
// 			},
// 		},
// 	} {
// 		var opts defaultOptions
// 		err := Run(&opts, test.args...)
// 		assert.NoError(t, err)
// 		assert.Equal(t, test.expected, opts)
// 	}
// }

// func TestNoDefaultsForBools(t *testing.T) {
// 	var opts struct {
// 		DefaultBool bool `short:"d" default:"true"`
// 	}

// 	var errMessage string
// 	if runtime.GOOS == "windows" {
// 		errMessage = "boolean flag `/d' may not have default values, they always default to `false' and can only be turned on"
// 	} else {
// 		errMessage = "boolean flag `-d' may not have default values, they always default to `false' and can only be turned on"
// 	}
// 	assertEqual(t, &Error{
// 		Kind:    ErrInvalidTag,
// 		Message: errMessage,
// 	}, &opts)
// }

// func TestUnquoting(t *testing.T) {
// 	tests := []struct {
// 		arg   string
// 		err   error
// 		value string
// 	}{
// 		{
// 			arg:   "\"abc",
// 			err:   strconv.ErrSyntax,
// 			value: "",
// 		},
// 		{
// 			arg:   "\"\"abc\"",
// 			err:   strconv.ErrSyntax,
// 			value: "",
// 		},
// 		{
// 			arg:   "\"abc\"",
// 			err:   nil,
// 			value: "abc",
// 		},
// 		{
// 			arg:   "\"\\\"abc\\\"\"",
// 			err:   nil,
// 			value: "\"abc\"",
// 		},
// 		{
// 			arg:   "\"\\\"abc\"",
// 			err:   nil,
// 			value: "\"abc",
// 		},
// 	}

// 	for _, test := range tests {
// 		var opts defaultOptions
// 		for _, args := range [][]string{
// 			{"--str=" + test.arg, "--strnot=" + test.arg},
// 			{"--str", test.arg, "--strnot", test.arg},
// 		} {
// 			err := Run(&opts, args...)
// 			if test.err == nil {
// 				assert.NoError(t, err)
// 				assert.Equal(t, test.value, opts.String)
// 				assert.Equal(t, strconv.Quote(test.value), opts.StringNotUnquoted)
// 			} else {
// 				assert.Error(t, err)
// 				if e, ok := err.(*Error); ok {
// 					assert.False(t, strings.HasPrefix(e.Message, test.err.Error()))
// 				}
// 			}
// 		}
// 	}
// }

// type CustomFlag struct {
// 	Value string
// }

// func (c *CustomFlag) UnmarshalFlag(s string) error {
// 	c.Value = s
// 	return nil
// }

// func (c *CustomFlag) IsValidValue(s string) error {
// 	if s != "-1" && s != "-foo" {
// 		return errors.New("invalid flag value")
// 	}
// 	return nil
// }

// func TestOptionAsArgument(t *testing.T) {
// 	var opts struct {
// 		StringSlice []string   `long:"string-slice"`
// 		IntSlice    []int      `long:"int-slice"`
// 		Int16       int16      `long:"int16"`
// 		Float32     float32    `long:"float32"`
// 		Float32Ptr  *float32   `long:"float32ptr"`
// 		OtherOption bool       `long:"other-option" short:"o"`
// 		Custom      CustomFlag `long:"custom-flag" short:"c"`
// 	}

// 	for _, test := range []struct {
// 		args        []string
// 		expectError bool
// 		errWant     *Error
// 	}{
// 		{
// 			// short option must not be accepted as argument
// 			args:        []string{"--string-slice", "foobar", "--string-slice", "-o"},
// 			expectError: true,
// 			errWant: &Error{
// 				Kind:    ErrExpectedArgument,
// 				Message: "expected argument for flag `" + _defaultOptDelimiterLong + "string-slice', but got option `-o'",
// 			},
// 		},
// 		{
// 			// long option must not be accepted as argument
// 			args:        []string{"--string-slice", "foobar", "--string-slice", "--other-option"},
// 			expectError: true,
// 			errWant: &Error{
// 				Kind:    ErrExpectedArgument,
// 				Message: "expected argument for flag `" + _defaultOptDelimiterLong + "string-slice', but got option `--other-option'",
// 			},
// 		},
// 		{
// 			// long option must not be accepted as argument
// 			args:        []string{"--string-slice", "--"},
// 			expectError: true,
// 			errWant: &Error{
// 				Kind:    ErrExpectedArgument,
// 				Message: "expected argument for flag `" + _defaultOptDelimiterLong + "string-slice', but got double dash `--'",
// 			},
// 		},
// 		{
// 			// quoted and appended option should be accepted as argument (even if it looks like an option)
// 			args: []string{"--string-slice", "foobar", "--string-slice=\"--other-option\""},
// 		},
// 		{
// 			// Accept any single character arguments including '-'
// 			args: []string{"--string-slice", "-"},
// 		},
// 		{
// 			// Do not accept arguments which start with '-' even if the next character is a digit
// 			args:        []string{"--string-slice", "-3.14"},
// 			expectError: true,
// 			errWant: &Error{
// 				Kind:    ErrExpectedArgument,
// 				Message: "expected argument for flag `" + _defaultOptDelimiterLong + "string-slice', but got option `-3.14'",
// 			},
// 		},
// 		{
// 			// Do not accept arguments which start with '-' if the next character is not a digit
// 			args:        []string{"--string-slice", "-character"},
// 			expectError: true,
// 			errWant: &Error{
// 				Kind:    ErrExpectedArgument,
// 				Message: "expected argument for flag `" + _defaultOptDelimiterLong + "string-slice', but got option `-character'",
// 			},
// 		},
// 		{
// 			args: []string{"-o", "-", "-"},
// 			// rest: []string{"-", "-"},
// 		},
// 		{
// 			// Accept arguments which start with '-' if the next character is a digit
// 			args: []string{"--int-slice", "-3"},
// 		},
// 		{
// 			// Accept arguments which start with '-' if the next character is a digit
// 			args: []string{"--int16", "-3"},
// 		},
// 		{
// 			// Accept arguments which start with '-' if the next character is a digit
// 			args: []string{"--float32", "-3.2"},
// 		},
// 		{
// 			// Accept arguments which start with '-' if the next character is a digit
// 			args: []string{"--float32ptr", "-3.2"},
// 		},
// 		{
// 			// Accept arguments for values that pass the IsValidValue fuction for value validators
// 			args: []string{"--custom-flag", "-foo"},
// 		},
// 		{
// 			// Accept arguments for values that pass the IsValidValue fuction for value validators
// 			args: []string{"--custom-flag", "-1"},
// 		},
// 		{
// 			// Rejects arguments for values that fail the IsValidValue fuction for value validators
// 			args:        []string{"--custom-flag", "-2"},
// 			expectError: true,
// 			errWant: &Error{
// 				Kind:    ErrExpectedArgument,
// 				Message: "invalid flag value",
// 			},
// 		},
// 	} {
// 		if test.expectError {
// 			assertEqual(t, test.errWant, &opts, test.args...)
// 		} else {
// 			assertNoError(t, &opts, test.args...)
// 		}
// 	}
// }

// func TestUnknownFlagHandler(t *testing.T) {
// 	var opts struct {
// 		Flag1 string `long:"flag1"`
// 		Flag2 string `long:"flag2"`
// 	}

// 	var unknownFlag1 string
// 	var unknownFlag2 bool
// 	var unknownFlag3 string

// 	// Parse args containing some unknown flags, verify that
// 	// our callback can handle all of them
// 	err := Run(&opts, "--flag1=stuff", "--unknownFlag1", "blah", "--unknownFlag2", "--unknownFlag3=baz", "--flag2=foo")
// 	assert.NoError(t, err)
// 	assert.Equal(t, opts.Flag1, "stuff")
// 	assert.Equal(t, opts.Flag2, "foo")
// 	assert.Equal(t, unknownFlag1, "blah")
// 	assert.Equal(t, unknownFlag3, "baz")
// 	assert.True(t, unknownFlag2)

// 	// Parse args with unknown flags that callback doesn't handle, verify it returns error
// 	err = Run(&opts, "--flag1=stuff", "--unknownFlagX", "blah", "--flag2=foo")
// 	assert.Error(t, err)
// }

// func TestChoices(t *testing.T) {
// 	var opts struct {
// 		Choice string `long:"choose" choice:"v1" choice:"v2"`
// 	}

// 	assertEqual(t, &Error{
// 		Kind:    ErrInvalidChoice,
// 		Message: "Invalid value `invalid' for option `" + _defaultOptDelimiterLong + "choose'. Allowed values are: v1 or v2",
// 	}, &opts, "--choose", "invalid")
// 	assertNoError(t, &opts, "--choose", "v2")
// 	assert.Equal(t, opts.Choice, "v2")
// }

// func TestEmbedded(t *testing.T) {
// 	type embedded struct {
// 		V bool `short:"v"`
// 	}
// 	var opts struct {
// 		embedded
// 	}

// 	assertNoError(t, &opts, "-v")
// 	assert.True(t, opts.V)
// }

// func TestCommandHandlerNoCommand(t *testing.T) {
// 	opts := struct {
// 		Value bool     `short:"v"`
// 		Args  []string `positional-args:"yes"`
// 	}{}

// 	// executed := false
// 	err := Run(&opts, "arg1", "arg2")
// 	assert.NoError(t, err)
// 	// assert.True(t, executed)
// 	assert.Equal(t, []string{"arg1", "arg2"}, opts.Args)
// }

// type command struct {
// 	Args []string `positional-args:"yes"`
// }

// func (c *command) Execute(args []string) error {
// 	return nil
// }

// func TestCommandHandler(t *testing.T) {
// 	opts := struct {
// 		Value   bool    `short:"v"`
// 		Command command `command:"cmd"`
// 	}{}

// 	executed := false
// 	err := Run(&opts, "cmd", "arg1", "arg2")
// 	assert.NoError(t, err)
// 	assert.True(t, executed)
// 	// assert.NotNil(t, executedCommand)
// 	assert.Equal(t, []string{"arg1", "arg2"}, opts.Command.Args)
// }
