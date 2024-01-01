package flags

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGroupInline(t *testing.T) {
	opts := struct {
		Value bool `short:"v"`

		Group struct {
			G bool `short:"g"`
		} `group:"Grouped Options"`
	}{}

	p, ret := assertParserSuccess(t, &opts, "-v", "-g")

	assertStringArray(t, ret, []string{})

	if !opts.Value {
		t.Errorf("Expected Value to be true")
	}

	if !opts.Group.G {
		t.Errorf("Expected Group.G to be true")
	}

	if p.Command.Group.Find("Grouped Options") == nil {
		t.Errorf("Expected to find group `Grouped Options'")
	}
}

func TestGroupAdd(t *testing.T) {
	opts := struct {
		Value bool `short:"v"`
	}{}

	grp := struct {
		G bool `short:"g"`
	}{}

	p := NewParser(&opts, Default)
	g, err := p.AddGroup("Grouped Options", "", &grp)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
		return
	}

	ret, err := p.ParseArgs("-v", "-g", "rest")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
		return
	}

	assertStringArray(t, ret, []string{"rest"})

	if !opts.Value {
		t.Errorf("Expected Value to be true")
	}

	if !grp.G {
		t.Errorf("Expected Group.G to be true")
	}

	if p.Command.Group.Find("Grouped Options") != g {
		t.Errorf("Expected to find group `Grouped Options'")
	}

	if p.Groups()[1] != g {
		t.Errorf("Expected group %#v, but got %#v", g, p.Groups()[0])
	}

	if g.Options()[0].ShortName != 'g' {
		t.Errorf("Expected short name `g' but got %v", g.Options()[0].ShortName)
	}
}

func TestGroupNestedInline(t *testing.T) {
	opts := struct {
		Value bool `short:"v"`

		Group struct {
			G bool `short:"g"`

			Nested struct {
				N string `long:"n"`
			} `group:"Nested Options"`
		} `group:"Grouped Options"`
	}{}

	p, ret := assertParserSuccess(t, &opts, "-v", "-g", "--n", "n", "rest")

	assertStringArray(t, ret, []string{"rest"})

	if !opts.Value {
		t.Errorf("Expected Value to be true")
	}

	if !opts.Group.G {
		t.Errorf("Expected Group.G to be true")
	}

	assert.Equal(t, opts.Group.Nested.N, "n")

	if p.Command.Group.Find("Grouped Options") == nil {
		t.Errorf("Expected to find group `Grouped Options'")
	}

	if p.Command.Group.Find("Nested Options") == nil {
		t.Errorf("Expected to find group `Nested Options'")
	}
}

func TestGroupNestedInlineNamespace(t *testing.T) {
	opts := struct {
		Opt string `long:"opt"`

		Group struct {
			Opt   string `long:"opt"`
			Group struct {
				Opt string `long:"opt"`
			} `group:"Subsubgroup" namespace:"sap"`
		} `group:"Subgroup" namespace:"sip"`
	}{}

	p, ret := assertParserSuccess(t, &opts, "--opt", "a", "--sip.opt", "b", "--sip.sap.opt", "c", "rest")

	assertStringArray(t, ret, []string{"rest"})

	assert.Equal(t, opts.Opt, "a")
	assert.Equal(t, opts.Group.Opt, "b")
	assert.Equal(t, opts.Group.Group.Opt, "c")

	for _, name := range []string{"Subgroup", "Subsubgroup"} {
		if p.Command.Group.Find(name) == nil {
			t.Errorf("Expected to find group '%s'", name)
		}
	}
}

func TestDuplicateShortFlags(t *testing.T) {
	var opts struct {
		Verbose   []bool   `short:"v" long:"verbose" description:"Show verbose debug information"`
		Variables []string `short:"v" long:"variable" description:"Set a variable value."`
	}

	_, err := ParseArgs(&opts,
		"--verbose",
		"-v", "123",
		"-v", "456",
	)

	if err == nil {
		t.Errorf("Expected an error with type ErrDuplicatedFlag")
	} else {
		err2 := err.(*Error)
		if err2.Kind != ErrDuplicatedFlag {
			t.Errorf("Expected an error with type ErrDuplicatedFlag")
		}
	}
}

func TestDuplicateLongFlags(t *testing.T) {
	var opts struct {
		Test1 []bool   `short:"a" long:"testing" description:"Test 1"`
		Test2 []string `short:"b" long:"testing" description:"Test 2."`
	}

	_, err := ParseArgs(&opts, "--testing")

	if err == nil {
		t.Errorf("Expected an error with type ErrDuplicatedFlag")
	} else {
		err2 := err.(*Error)
		if err2.Kind != ErrDuplicatedFlag {
			t.Errorf("Expected an error with type ErrDuplicatedFlag")
		}
	}
}

func TestFindOptionByLongFlag(t *testing.T) {
	var opts struct {
		Testing bool `long:"testing" description:"Testing"`
	}

	p := NewParser(&opts, Default)
	opt := p.FindOptionByLongName("testing")
	assert.NotNil(t, opt)
	assert.Equal(t, "testing", opt.LongName)
}

func TestFindOptionByShortFlag(t *testing.T) {
	var opts struct {
		Testing bool `short:"t" description:"Testing"`
	}

	p := NewParser(&opts, Default)
	opt := p.FindOptionByShortName('t')
	assert.NotNil(t, opt)
	assert.Equal(t, 't', opt.ShortName)
}

func TestFindOptionByLongFlagInSubGroup(t *testing.T) {
	var opts struct {
		Group struct {
			Testing bool `long:"testing" description:"Testing"`
		} `group:"sub-group"`
	}

	p := NewParser(&opts, Default)
	opt := p.FindOptionByLongName("testing")
	assert.NotNil(t, opt)
	assert.Equal(t, "testing", opt.LongName)
}

func TestFindOptionByShortFlagInSubGroup(t *testing.T) {
	var opts struct {
		Group struct {
			Testing bool `short:"t" description:"Testing"`
		} `group:"sub-group"`
	}

	p := NewParser(&opts, Default)
	opt := p.FindOptionByShortName('t')
	assert.NotNil(t, opt)
	assert.Equal(t, 't', opt.ShortName)
}

func TestAddOptionNonOptional(t *testing.T) {
	var opts struct {
		Test bool
	}
	p := NewParser(&opts, Default)
	p.AddOption(&Option{
		LongName: "test",
	}, &opts.Test)
	_, err := p.ParseArgs("--test")
	if err != nil {
		t.Errorf("unexpected error: %s", err)
	} else if !opts.Test {
		t.Errorf("option not set")
	}
}
