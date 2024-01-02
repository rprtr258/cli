package flags

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGroupInline(t *testing.T) {
	opts := struct {
		Value bool `short:"v"`
		G     bool `short:"g"`
	}{}

	assertParserSuccess(t, &opts, "-v", "-g")
	assert.True(t, opts.Value)
	assert.True(t, opts.G)
}

func TestGroupAdd(t *testing.T) {
	opts := struct {
		Value bool     `short:"v"`
		G     bool     `short:"g"`
		Args  []string `args`
	}{}

	p := NewParser(&opts)

	err := p.Run("-v", "-g", "rest")
	assert.NoError(t, err)

	assert.Equal(t, []string{"rest"}, opts.Args)
	assert.True(t, opts.Value)
	assert.True(t, opts.G)
}

func TestGroupNestedInline(t *testing.T) {
	opts := struct {
		Value bool     `short:"v"`
		G     bool     `short:"g"`
		N     string   `long:"n"`
		Args  []string `args`
	}{}

	assertParserSuccess(t, &opts, "-v", "-g", "--n", "n", "rest")
	assert.Equal(t, []string{"rest"}, opts.Args)
	assert.True(t, opts.Value)
	assert.True(t, opts.G)
	assert.Equal(t, opts.N, "n")
}

func TestGroupNestedInlineNamespace(t *testing.T) {
	opts := struct {
		Opt       string   `long:"opt"`
		SipOpt    string   `long:"sip.opt"`
		SipSapOpt string   `long:"sip.sap.opt"`
		Args      []string `args`
	}{}

	assertParserSuccess(t, &opts, "--opt", "a", "--sip.opt", "b", "--sip.sap.opt", "c", "rest")
	assert.Equal(t, []string{"rest"}, opts.Args)
	assert.Equal(t, opts.Opt, "a")
	assert.Equal(t, opts.SipOpt, "b")
	assert.Equal(t, opts.SipSapOpt, "c")
}

func TestDuplicateShortFlags(t *testing.T) {
	var opts struct {
		Verbose   []bool   `short:"v" long:"verbose" description:"Show verbose debug information"`
		Variables []string `short:"v" long:"variable" description:"Set a variable value."`
	}

	err := NewParser(&opts).Run(
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

	err := NewParser(&opts).Run("--testing")
	assert.Error(t, err)
	err2 := err.(*Error)
	assert.Equal(t, ErrDuplicatedFlag, err2.Kind)
}

func TestAddOptionNonOptional(t *testing.T) {
	var opts struct {
		Test bool
	}
	p := NewParser(&opts)
	p.Command.Flags.AddOption(&Flag{
		LongName: "test",
	}, &opts.Test)

	err := p.Run("--test")
	assert.NoError(t, err)
	assert.True(t, opts.Test)
}

func TestLong(t *testing.T) {
	t.Run("Long", func(t *testing.T) {
		opts := struct {
			Value bool `long:"value"`
		}{}

		ret := assertParserSuccess(t, &opts, "--value")
		assert.Equal(t, []string{}, ret)
		assert.True(t, opts.Value)
	})

	t.Run("LongArg", func(t *testing.T) {
		opts := struct {
			Value string `long:"value"`
		}{}

		ret := assertParserSuccess(t, &opts, "--value", "value")
		assert.Equal(t, []string{}, ret)
		assert.Equal(t, "value", opts.Value)
	})

	t.Run("LongArgEqual", func(t *testing.T) {
		opts := struct {
			Value string `long:"value"`
		}{}

		ret := assertParserSuccess(t, &opts, "--value=value")
		assert.Equal(t, []string{}, ret)
		assert.Equal(t, "value", opts.Value)
	})

	t.Run("LongDefault", func(t *testing.T) {
		opts := struct {
			Value string `long:"value" default:"value"`
		}{}

		ret := assertParserSuccess(t, &opts)
		assert.Equal(t, []string{}, ret)
		assert.Equal(t, "value", opts.Value)
	})

	t.Run("LongOptional", func(t *testing.T) {
		opts := struct {
			Value string `long:"value" optional:"yes" optional-value:"value"`
		}{}

		ret := assertParserSuccess(t, &opts, "--value")
		assert.Equal(t, []string{}, ret)
		assert.Equal(t, "value", opts.Value)
	})

	t.Run("LongOptionalArg", func(t *testing.T) {
		opts := struct {
			Value string `long:"value" optional:"yes" optional-value:"value"`
		}{}

		ret := assertParserSuccess(t, &opts, "--value", "no")
		assert.Equal(t, []string{"no"}, ret)
		assert.Equal(t, "value", opts.Value)
	})

	t.Run("LongOptionalArgEqual", func(t *testing.T) {
		opts := struct {
			Value string `long:"value" optional:"yes" optional-value:"value"`
		}{}

		ret := assertParserSuccess(t, &opts, "--value=value", "no")
		assert.Equal(t, []string{"no"}, ret)
		assert.Equal(t, "value", opts.Value)
	})
}

func TestUnknown(t *testing.T) {
	t.Run("UnknownFlags", func(t *testing.T) {
		opts := struct {
			Verbose []bool `short:"v" long:"verbose" description:"Verbose output"`
		}{}

		err := NewParser(&opts).Run("-f")
		assert.Error(t, err)
	})

	t.Run("IgnoreUnknownFlags", func(t *testing.T) {
		opts := struct {
			Verbose []bool   `short:"v" long:"verbose" description:"Verbose output"`
			Args    []string `args`
		}{}

		err := NewParser(&opts).Run(
			"hello",
			"world",
			"-v",
			"--foo=bar",
			"--verbose",
			"-f",
		)
		assert.NoError(t, err)
		assert.Len(t, opts.Verbose, 0)
		assert.Equal(t, []string{
			"hello",
			"world",
			"-v",
			"--foo=bar",
			"--verbose",
			"-f",
		}, opts.Args)
	})
}
