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
	assert.True(t, opts.Value)
	assert.True(t, opts.Group.G)
	assert.NotNil(t, p.Command.Group.Find("Grouped Options"))
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
	assert.NoError(t, err)

	ret, err := p.ParseArgs("-v", "-g", "rest")
	assert.NoError(t, err)

	assertStringArray(t, ret, []string{"rest"})
	assert.True(t, opts.Value)
	assert.True(t, grp.G)
	assert.Equal(t, g, p.Command.Group.Find("Grouped Options"))
	assert.Equal(t, g, p.Groups()[1])
	assert.Equal(t, 'g', g.Options()[0].ShortName)
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
	assert.True(t, opts.Value)
	assert.True(t, opts.Group.G)
	assert.Equal(t, opts.Group.Nested.N, "n")
	assert.NotNil(t, p.Command.Group.Find("Grouped Options"))
	assert.NotNil(t, p.Command.Group.Find("Nested Options"))
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
		assert.NotNil(t, p.Command.Group.Find(name))
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
	assert.Error(t, err)
	err2 := err.(*Error)
	assert.Equal(t, ErrDuplicatedFlag, err2.Kind)
}

func TestDuplicateLongFlags(t *testing.T) {
	var opts struct {
		Test1 []bool   `short:"a" long:"testing" description:"Test 1"`
		Test2 []string `short:"b" long:"testing" description:"Test 2."`
	}

	_, err := ParseArgs(&opts, "--testing")
	assert.Error(t, err)
	err2 := err.(*Error)
	assert.Equal(t, ErrDuplicatedFlag, err2.Kind)
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
	assert.NoError(t, err)
	assert.True(t, opts.Test)
}
