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

func TestPositional(t *testing.T) {
	opts := struct {
		Value      bool `short:"v"`
		Positional struct {
			Command  int
			Filename string
			Rest     []string
		} `positional-args:"yes" required:"yes"`
	}{}

	err := NewParser(&opts).Run("10", "arg_test.go", "a", "b")
	assert.NoError(t, err)
	assert.Equal(t, 10, opts.Positional.Command)
	assert.Equal(t, "arg_test.go", opts.Positional.Filename)
	assert.Equal(t, []string{"a", "b"}, opts.Positional.Rest)
}

func TestPositionalRequired(t *testing.T) {
	opts := struct {
		Value      bool `short:"v"`
		Positional struct {
			Command  int
			Filename string
			Rest     []string
		} `positional-args:"yes" required:"yes"`
	}{}

	err := NewParser(&opts).Run("10")
	assert.Equal(t, &Error{
		Kind:    ErrRequired,
		Message: "the required argument `Filename` was not provided",
	}, err)
}

func TestPositionalRequiredRest1Fail(t *testing.T) {
	opts := struct {
		Value      bool `short:"v"`
		Positional struct {
			Rest []string `required:"yes"`
		} `positional-args:"yes"`
	}{}

	err := NewParser(&opts).Run()
	assert.Equal(t, &Error{
		Kind:    ErrRequired,
		Message: "the required argument `Rest (at least 1 argument)` was not provided",
	}, err)
}

func TestPositionalRequiredRest1Pass(t *testing.T) {
	opts := struct {
		Value      bool `short:"v"`
		Positional struct {
			Rest []string `required:"yes"`
		} `positional-args:"yes"`
	}{}

	p := NewParser(&opts)
	err := p.Run("rest1")
	assert.NoError(t, err)
	assert.Equal(t, []string{"rest1"}, opts.Positional.Rest)
}

func TestPositionalRequiredRest2Fail(t *testing.T) {
	opts := struct {
		Value      bool `short:"v"`
		Positional struct {
			Rest []string `required:"2"`
		} `positional-args:"yes"`
	}{}

	err := NewParser(&opts).Run("rest1")
	assert.Equal(t, &Error{
		Kind:    ErrRequired,
		Message: "the required argument `Rest (at least 2 arguments, but got only 1)` was not provided",
	}, err)
}

func TestPositionalRequiredRest2Pass(t *testing.T) {
	opts := struct {
		Value      bool `short:"v"`
		Positional struct {
			Rest []string `required:"2"`
		} `positional-args:"yes"`
	}{}

	err := NewParser(&opts).Run("rest1", "rest2", "rest3")
	assert.NoError(t, err)
	assert.Equal(t, []string{"rest1", "rest2", "rest3"}, opts.Positional.Rest)
}

func TestPositionalRequiredRestRangeFail(t *testing.T) {
	opts := struct {
		Value      bool `short:"v"`
		Positional struct {
			Rest []string `required:"1-2"`
		} `positional-args:"yes"`
	}{}

	err := NewParser(&opts).Run("rest1", "rest2", "rest3")
	assert.Equal(t, &Error{
		Kind:    ErrRequired,
		Message: "the required argument `Rest (at most 2 arguments, but got 3)` was not provided",
	}, err)
}

func TestPositionalRequiredRestRangeEmptyFail(t *testing.T) {
	opts := struct {
		Value      bool `short:"v"`
		Positional struct {
			Rest []string `required:"0-0"`
		} `positional-args:"yes"`
	}{}

	err := NewParser(&opts).Run("some", "thing")
	assert.Equal(t, &Error{
		Kind:    ErrRequired,
		Message: "the required argument `Rest (zero arguments)` was not provided",
	}, err)
}
