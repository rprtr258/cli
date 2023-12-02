package cli

import (
	"bytes"
	"fmt"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWriteIni(t *testing.T) {
	t.Setenv("ENV_DEFAULT", "env-def")

	tests := []struct {
		args     []string
		options  IniOptions
		expected string
	}{
		{
			[]string{"-vv", "--intmap=a:2", "--intmap", "b:3", "filename", "0", "3.14", "command"},
			IniDefault,
			`[Application Options]
; Show verbose debug information
verbose = true
verbose = true

; Test env-default1 value
EnvDefault1 = env-def

; Test env-default2 value
EnvDefault2 = env-def

[Other Options]
; A map from string to int
int-map = a:2
int-map = b:3

`,
		},
		{
			[]string{"-vv", "--intmap=a:2", "--intmap", "b:3", "filename", "0", "3.14", "command"},
			IniDefault | IniIncludeDefaults,
			`[Application Options]
; Show verbose debug information
verbose = true
verbose = true

; A slice of pointers to string
; PtrSlice =

EmptyDescription = false

; Test default value
Default = "Some\nvalue"

; Test default array value
DefaultArray = Some value
DefaultArray = "Other\tvalue"

; Testdefault map value
DefaultMap = another:value
DefaultMap = some:value

; Test env-default1 value
EnvDefault1 = env-def

; Test env-default2 value
EnvDefault2 = env-def

; Option with named argument
OptionWithArgName =

; Option with choices
OptionWithChoices =

; Option only available in ini
only-ini =

[Other Options]
; A slice of strings
StringSlice = some
StringSlice = value

; A map from string to int
int-map = a:2
int-map = b:3

[Subgroup]
; This is a subgroup option
Opt =

; Not hidden inside group
NotHiddenInsideGroup =

[Subsubgroup]
; This is a subsubgroup option
Opt =

[command]
; Use for extra verbosity
; ExtraVerbose =

[parent]
; This is a parent command option
Opt =

[parent.sub]
; This is a sub command option
Opt =

`,
		},
		{
			[]string{"filename", "0", "3.14", "command"},
			IniDefault | IniIncludeDefaults | IniCommentDefaults,
			`[Application Options]
; Show verbose debug information
; verbose =

; A slice of pointers to string
; PtrSlice =

; EmptyDescription = false

; Test default value
; Default = "Some\nvalue"

; Test default array value
; DefaultArray = Some value
; DefaultArray = "Other\tvalue"

; Testdefault map value
; DefaultMap = another:value
; DefaultMap = some:value

; Test env-default1 value
EnvDefault1 = env-def

; Test env-default2 value
EnvDefault2 = env-def

; Option with named argument
; OptionWithArgName =

; Option with choices
; OptionWithChoices =

; Option only available in ini
; only-ini =

[Other Options]
; A slice of strings
; StringSlice = some
; StringSlice = value

; A map from string to int
; int-map = a:1

[Subgroup]
; This is a subgroup option
; Opt =

; Not hidden inside group
; NotHiddenInsideGroup =

[Subsubgroup]
; This is a subsubgroup option
; Opt =

[command]
; Use for extra verbosity
; ExtraVerbose =

[parent]
; This is a parent command option
; Opt =

[parent.sub]
; This is a sub command option
; Opt =

`,
		},
		{
			[]string{"--default=New value", "--default-array=New value", "--default-map=new:value", "filename", "0", "3.14", "command"},
			IniDefault | IniIncludeDefaults | IniCommentDefaults,
			`[Application Options]
; Show verbose debug information
; verbose =

; A slice of pointers to string
; PtrSlice =

; EmptyDescription = false

; Test default value
Default = New value

; Test default array value
DefaultArray = New value

; Testdefault map value
DefaultMap = new:value

; Test env-default1 value
EnvDefault1 = env-def

; Test env-default2 value
EnvDefault2 = env-def

; Option with named argument
; OptionWithArgName =

; Option with choices
; OptionWithChoices =

; Option only available in ini
; only-ini =

[Other Options]
; A slice of strings
; StringSlice = some
; StringSlice = value

; A map from string to int
; int-map = a:1

[Subgroup]
; This is a subgroup option
; Opt =

; Not hidden inside group
; NotHiddenInsideGroup =

[Subsubgroup]
; This is a subsubgroup option
; Opt =

[command]
; Use for extra verbosity
; ExtraVerbose =

[parent]
; This is a parent command option
; Opt =

[parent.sub]
; This is a sub command option
; Opt =

`,
		},
		{
			[]string{"-vv", "filename", "0", "3.14", "parent", "--opt=p", "sub", "--opt=s"},
			IniDefault,
			`[Application Options]
; Show verbose debug information
verbose = true
verbose = true

; Test env-default1 value
EnvDefault1 = env-def

; Test env-default2 value
EnvDefault2 = env-def

[parent]
; This is a parent command option
Opt = p

[parent.sub]
; This is a sub command option
Opt = s

`,
		},
	}

	for _, test := range tests {
		var opts helpOptions

		p := NewNamedParser("TestIni", Default)
		_, _ = p.AddGroup("Application Options", "The application options", &opts)

		_, err := p.ParseArgs(test.args)
		assert.NoError(t, err)

		inip := NewIniParser(p)

		var b bytes.Buffer
		inip.Write(&b, test.options)

		got := b.String()
		expected := test.expected

		msg := fmt.Sprintf("with arguments %+v and ini options %b", test.args, test.options)
		assert.Equal(t, got, expected, msg)
	}
}

func TestReadIni_flagEquivalent(t *testing.T) {
	type options struct {
		Opt1 bool `long:"opt1"`

		Group1 struct {
			Opt2 bool `long:"opt2"`
		} `group:"group1"`

		Group2 struct {
			Opt3 bool `long:"opt3"`
		} `group:"group2" namespace:"ns1"`

		Cmd1 struct {
			Opt4 bool `long:"opt4"`
			Opt5 bool `long:"foo.opt5"`

			Group1 struct {
				Opt6 bool `long:"opt6"`
				Opt7 bool `long:"foo.opt7"`
			} `group:"group1"`

			Group2 struct {
				Opt8 bool `long:"opt8"`
			} `group:"group2" namespace:"ns1"`
		} `command:"cmd1"`
	}

	a := `
opt1=true

[group1]
opt2=true

[group2]
ns1.opt3=true

[cmd1]
opt4=true
foo.opt5=true

[cmd1.group1]
opt6=true
foo.opt7=true

[cmd1.group2]
ns1.opt8=true
`
	b := `
opt1=true
opt2=true
ns1.opt3=true

[cmd1]
opt4=true
foo.opt5=true
opt6=true
foo.opt7=true
ns1.opt8=true
`

	parse := func(readIni string) (opts options, writeIni string) {
		p := NewNamedParser("TestIni", Default)
		_, _ = p.AddGroup("Application Options", "The application options", &opts)

		inip := NewIniParser(p)
		err := inip.Parse(strings.NewReader(readIni))
		assert.NoError(t, err)

		var b bytes.Buffer
		inip.Write(&b, Default)

		return opts, b.String()
	}

	aOpt, aIni := parse(a)
	bOpt, bIni := parse(b)

	assert.Equal(t, aIni, bIni, "")
	assert.True(t, reflect.DeepEqual(aOpt, bOpt))
}

func TestReadIni(t *testing.T) {
	var opts helpOptions

	p := NewNamedParser("TestIni", Default)
	_, _ = p.AddGroup("Application Options", "The application options", &opts)

	inip := NewIniParser(p)

	inic := `
; Show verbose debug information
verbose = true
verbose = true

DefaultMap = another:"value\n1"
DefaultMap = some:value 2

[Application Options]
; A slice of pointers to string
; PtrSlice =

; Test default value
Default = "New\nvalue"

; Test env-default1 value
EnvDefault1 = New value

[Other Options]
# A slice of strings
StringSlice = "some\nvalue"
StringSlice = another value

; A map from string to int
int-map = a:2
int-map = b:3

`

	b := strings.NewReader(inic)
	err := inip.Parse(b)
	assert.NoError(t, err)

	assert.Equal(t, opts.Verbose, []bool{true, true})

	if v := map[string]string{"another": "value\n1", "some": "value 2"}; !reflect.DeepEqual(opts.DefaultMap, v) {
		t.Fatalf("Expected %#v for DefaultMap but got %#v", v, opts.DefaultMap)
	}

	assert.Equal(t, opts.Default, "New\nvalue")

	assert.Equal(t, opts.EnvDefault1, "New value")

	assert.Equal(t, opts.Other.StringSlice, []string{"some\nvalue", "another value"})

	v, ok := opts.Other.IntMap["a"]
	assert.True(t, ok)
	assert.Equal(t, 2, v)

	v, ok = opts.Other.IntMap["b"]
	assert.True(t, ok)
	assert.Equal(t, 3, v)
}

func TestReadAndWriteIni(t *testing.T) {
	tests := []struct {
		options IniOptions
		read    string
		write   string
	}{
		{
			IniIncludeComments,
			`[Application Options]
; Show verbose debug information
verbose = true
verbose = true

; Test default value
Default = "quote me"

; Test default array value
DefaultArray = 1
DefaultArray = "2"
DefaultArray = 3

; Testdefault map value
; DefaultMap =

; Test env-default1 value
EnvDefault1 = env-def

; Test env-default2 value
EnvDefault2 = env-def

[Other Options]
; A slice of strings
; StringSlice =

; A map from string to int
int-map = a:2
int-map = b:"3"

`,
			`[Application Options]
; Show verbose debug information
verbose = true
verbose = true

; Test default value
Default = "quote me"

; Test default array value
DefaultArray = 1
DefaultArray = 2
DefaultArray = 3

; Testdefault map value
; DefaultMap =

; Test env-default1 value
EnvDefault1 = env-def

; Test env-default2 value
EnvDefault2 = env-def

[Other Options]
; A slice of strings
; StringSlice =

; A map from string to int
int-map = a:2
int-map = b:3

`,
		},
		{
			IniIncludeComments,
			`[Application Options]
; Show verbose debug information
verbose = true
verbose = true

; Test default value
Default = "quote me"

; Test default array value
DefaultArray = "1"
DefaultArray = "2"
DefaultArray = "3"

; Testdefault map value
; DefaultMap =

; Test env-default1 value
EnvDefault1 = env-def

; Test env-default2 value
EnvDefault2 = env-def

[Other Options]
; A slice of strings
; StringSlice =

; A map from string to int
int-map = a:"2"
int-map = b:"3"

`,
			`[Application Options]
; Show verbose debug information
verbose = true
verbose = true

; Test default value
Default = "quote me"

; Test default array value
DefaultArray = "1"
DefaultArray = "2"
DefaultArray = "3"

; Testdefault map value
; DefaultMap =

; Test env-default1 value
EnvDefault1 = env-def

; Test env-default2 value
EnvDefault2 = env-def

[Other Options]
; A slice of strings
; StringSlice =

; A map from string to int
int-map = a:"2"
int-map = b:"3"

`,
		},
	}

	for _, test := range tests {
		var opts helpOptions

		p := NewNamedParser("TestIni", Default)
		_, _ = p.AddGroup("Application Options", "The application options", &opts)

		inip := NewIniParser(p)

		read := strings.NewReader(test.read)
		err := inip.Parse(read)
		assert.NoError(t, err)

		var write bytes.Buffer
		inip.Write(&write, test.options)

		got := write.String()

		msg := fmt.Sprintf("with ini options %b", test.options)
		assert.Equal(t, got, test.write, msg)
	}
}

func TestReadIniWrongQuoting(t *testing.T) {
	tests := []struct {
		iniFile    string
		lineNumber uint
	}{
		{
			iniFile:    `Default = "New\nvalue`,
			lineNumber: 1,
		},
		{
			iniFile:    `StringSlice = "New\nvalue`,
			lineNumber: 1,
		},
		{
			iniFile: `StringSlice = "New\nvalue"
			StringSlice = "Second\nvalue`,
			lineNumber: 2,
		},
		{
			iniFile:    `DefaultMap = some:"value`,
			lineNumber: 1,
		},
		{
			iniFile: `DefaultMap = some:value
			DefaultMap = another:"value`,
			lineNumber: 2,
		},
	}

	for _, test := range tests {
		var opts helpOptions

		p := NewNamedParser("TestIni", Default)
		_, _ = p.AddGroup("Application Options", "The application options", &opts)

		inip := NewIniParser(p)

		inic := test.iniFile

		b := strings.NewReader(inic)
		err := inip.Parse(b)

		assert.Error(t, err)

		iniError := err.(*IniError)

		if iniError.LineNumber != test.lineNumber {
			t.Fatalf("Expect error on line %d", test.lineNumber)
		}
	}
}

func TestIniCommands(t *testing.T) {
	var opts struct {
		Value string `short:"v" long:"value"`

		Add struct {
			Name int `short:"n" long:"name" ini-name:"AliasName"`

			Other struct {
				O string `short:"o" long:"other"`
			} `group:"Other Options"`
		} `command:"add"`
	}

	p := NewNamedParser("TestIni", Default)
	_, _ = p.AddGroup("Application Options", "The application options", &opts)

	inip := NewIniParser(p)

	inic := `[Application Options]
value = some value

[add]
AliasName = 5

[add.Other Options]
other = subgroup

`

	b := strings.NewReader(inic)
	err := inip.Parse(b)
	assert.NoError(t, err)

	assert.Equal(t, opts.Value, "some value")

	assert.Equal(t, 5, opts.Add.Name)

	assert.Equal(t, opts.Add.Other.O, "subgroup")

	// Test writing it back
	buf := &bytes.Buffer{}

	inip.Write(buf, IniDefault)

	assert.Equal(t, buf.String(), inic, "ini contents")
}

func TestIniNoIni(t *testing.T) {
	var opts struct {
		NoValue string `short:"n" long:"novalue" no-ini:"yes"`
		Value   string `short:"v" long:"value"`
	}

	p := NewNamedParser("TestIni", Default)
	_, _ = p.AddGroup("Application Options", "The application options", &opts)

	inip := NewIniParser(p)

	// read INI
	inic := `[Application Options]
novalue = some value
value = some other value
`

	b := strings.NewReader(inic)
	err := inip.Parse(b)
	assert.Equal(t, &IniError{
		LineNumber: 2,
		Message:    "unknown option: novalue",
	}, err)

	// write INI
	opts.NoValue = "some value"
	opts.Value = "some other value"

	file, err := os.CreateTemp("", "")
	assert.NoError(t, err)
	defer os.Remove(file.Name())

	err = inip.WriteFile(file.Name(), IniIncludeDefaults)
	assert.NoError(t, err)

	found, err := os.ReadFile(file.Name())
	assert.NoError(t, err)

	expected := "[Application Options]\nValue = some other value\n\n"

	assert.Equal(t, string(found), expected, "ini content")
}

func TestIniParse(t *testing.T) {
	file, err := os.CreateTemp("", "")
	assert.NoError(t, err)
	defer os.Remove(file.Name())

	_, err = file.WriteString("value = 123")
	assert.NoError(t, err)

	file.Close()

	var opts struct {
		Value int `long:"value"`
	}

	err = IniParse(file.Name(), &opts)
	assert.NoError(t, err)

	assert.Equal(t, 123, opts.Value)
}

func TestIniCliOverrides(t *testing.T) {
	file, err := os.CreateTemp("", "")
	assert.NoError(t, err)
	defer os.Remove(file.Name())
	_, err = file.WriteString("values = 123\n" + "values = 456\n")
	assert.NoError(t, err)
	file.Close()

	var opts struct {
		Values []int `long:"values"`
	}

	p := NewParser(&opts, Default)
	err = NewIniParser(p).ParseFile(file.Name())

	assert.NoError(t, err)

	_, err = p.ParseArgs([]string{"--values", "111", "--values", "222"})

	assert.NoError(t, err)

	assert.Equal(t, 2, len(opts.Values))

	assert.Equal(t, 111, opts.Values[0])

	assert.Equal(t, 222, opts.Values[1])
}

func TestIniOverrides(t *testing.T) {
	file, err := os.CreateTemp("", "")
	assert.NoError(t, err)
	defer os.Remove(file.Name())
	_, err = file.WriteString("value-with-default = \"ini-value\"\n" + "value-with-default-override-cli = \"ini-value\"\n")
	assert.NoError(t, err)
	file.Close()

	var opts struct {
		ValueWithDefault            string `long:"value-with-default" default:"value"`
		ValueWithDefaultOverrideCli string `long:"value-with-default-override-cli" default:"value"`
	}

	p := NewParser(&opts, Default)
	err = NewIniParser(p).ParseFile(file.Name())

	assert.NoError(t, err)

	_, err = p.ParseArgs([]string{"--value-with-default-override-cli", "cli-value"})

	assert.NoError(t, err)

	assert.Equal(t, opts.ValueWithDefault, "ini-value")
	assert.Equal(t, opts.ValueWithDefaultOverrideCli, "cli-value")
}

func TestIniOverridesFromConfigFlag(t *testing.T) {
	file, err := os.CreateTemp("", "")
	assert.NoError(t, err)
	defer os.Remove(file.Name())
	_, err = file.WriteString("value-with-default = \"ini-value\"\n" + "value-with-default-override-cli = \"ini-value\"\n")
	assert.NoError(t, err)
	file.Close()

	var opts struct {
		Config                      func(filename string) `long:"config"`
		ValueWithDefault            string                `long:"value-with-default" default:"value"`
		ValueWithDefaultOverrideCli string                `long:"value-with-default-override-cli" default:"value"`
	}

	p := NewParser(&opts, Default)

	opt := p.FindOptionByLongName("config")
	opt.Default = []string{file.Name()}

	opts.Config = func(filename string) {
		parser := NewIniParser(p)
		parser.ParseAsDefaults = true
		_ = parser.ParseFile(filename)
	}

	_, err = p.ParseArgs([]string{"--value-with-default-override-cli", "cli-value"})

	assert.NoError(t, err)

	assert.Equal(t, opts.ValueWithDefault, "ini-value")
	assert.Equal(t, opts.ValueWithDefaultOverrideCli, "cli-value")
}

func TestIniRequired(t *testing.T) {
	var opts struct {
		Required string               `short:"r" required:"yes" description:"required"`
		Config   func(s string) error `long:"config" default:"no-ini-file" no-ini:"true"`
	}

	p := NewParser(&opts, Default)

	opts.Config = func(s string) error {
		inip := NewIniParser(p)
		inip.ParseAsDefaults = true
		return inip.Parse(strings.NewReader("Required = ini-value\n"))
	}

	_, err := p.ParseArgs([]string{"-r", "cli-value"})
	assert.NoError(t, err)

	assert.Equal(t, opts.Required, "cli-value")
}

func TestIniRequiredSlice_ShouldNotNeedToBeSpecifiedOnCli(t *testing.T) {
	type options struct {
		Items []string `long:"item" required:"true"`
	}
	var opts options
	ini := `
[Application Options]
item=abc`
	args := []string{}

	parser := NewParser(&opts, Default)
	inip := NewIniParser(parser)

	_ = inip.Parse(strings.NewReader(ini))

	_, err := parser.ParseArgs(args)
	assert.NoError(t, err)

	assert.Equal(t, opts.Items[0], "abc")
}

func TestWriteFile(t *testing.T) {
	file, err := os.CreateTemp("", "")
	assert.NoError(t, err)
	defer os.Remove(file.Name())

	var opts struct {
		Value int `long:"value"`
	}

	opts.Value = 123

	p := NewParser(&opts, Default)
	ini := NewIniParser(p)

	err = ini.WriteFile(file.Name(), IniIncludeDefaults)
	assert.NoError(t, err)

	found, err := os.ReadFile(file.Name())
	assert.NoError(t, err)

	expected := "[Application Options]\nValue = 123\n\n"

	assert.Equal(t, string(found), expected, "ini content")
}

func TestOverwriteRequiredOptions(t *testing.T) {
	tests := []struct {
		args     []string
		expected []string
	}{
		{
			args: []string{"--value", "from CLI"},
			expected: []string{
				"from CLI",
				"from default",
			},
		},
		{
			args: []string{"--value", "from CLI", "--default", "from CLI"},
			expected: []string{
				"from CLI",
				"from CLI",
			},
		},
		{
			args: []string{"--config", "no file name"},
			expected: []string{
				"from INI",
				"from INI",
			},
		},
		{
			args: []string{"--value", "from CLI before", "--default", "from CLI before", "--config", "no file name"},
			expected: []string{
				"from INI",
				"from INI",
			},
		},
		{
			args: []string{"--value", "from CLI before", "--default", "from CLI before", "--config", "no file name", "--value", "from CLI after", "--default", "from CLI after"},
			expected: []string{
				"from CLI after",
				"from CLI after",
			},
		},
	}

	for _, test := range tests {
		var opts struct {
			Config  func(s string) error `long:"config" no-ini:"true"`
			Value   string               `long:"value" required:"true"`
			Default string               `long:"default" required:"true" default:"from default"`
		}

		p := NewParser(&opts, Default)

		opts.Config = func(s string) error {
			ini := NewIniParser(p)

			return ini.Parse(bytes.NewBufferString("value = from INI\ndefault = from INI"))
		}

		_, err := p.ParseArgs(test.args)
		assert.NoError(t, err)

		if opts.Value != test.expected[0] {
			t.Fatalf("Expected Value to be \"%s\" but was \"%s\" with args %+v", test.expected[0], opts.Value, test.args)
		}

		if opts.Default != test.expected[1] {
			t.Fatalf("Expected Default to be \"%s\" but was \"%s\" with args %+v", test.expected[1], opts.Default, test.args)
		}
	}
}

func TestIniOverwriteOptions(t *testing.T) {
	tests := []struct {
		args     []string
		expected string
		toggled  bool
	}{
		{
			args:     []string{},
			expected: "from default",
		},
		{
			args:     []string{"--value", "from CLI"},
			expected: "from CLI",
		},
		{
			args:     []string{"--config", "no file name"},
			expected: "from INI",
			toggled:  true,
		},
		{
			args:     []string{"--value", "from CLI before", "--config", "no file name"},
			expected: "from CLI before",
			toggled:  true,
		},
		{
			args:     []string{"--config", "no file name", "--value", "from CLI after"},
			expected: "from CLI after",
			toggled:  true,
		},
		{
			args:     []string{"--toggle"},
			toggled:  true,
			expected: "from default",
		},
	}

	for _, test := range tests {
		var opts struct {
			Config string `long:"config" no-ini:"true"`
			Value  string `long:"value" default:"from default"`
			Toggle bool   `long:"toggle"`
		}

		p := NewParser(&opts, Default)

		_, err := p.ParseArgs(test.args)
		assert.NoError(t, err)

		if opts.Config != "" {
			inip := NewIniParser(p)
			inip.ParseAsDefaults = true

			err = inip.Parse(bytes.NewBufferString("value = from INI\ntoggle = true"))
			assert.NoError(t, err)
		}

		if opts.Value != test.expected {
			t.Fatalf("Expected Value to be \"%s\" but was \"%s\" with args %+v", test.expected, opts.Value, test.args)
		}

		if opts.Toggle != test.toggled {
			t.Fatalf("Expected Toggle to be \"%v\" but was \"%v\" with args %+v", test.toggled, opts.Toggle, test.args)
		}

	}
}
