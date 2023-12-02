package cli

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPassDoubleDash(t *testing.T) {
	var opts struct {
		Value bool `short:"v"`
	}

	p := NewParser(&opts, PassDoubleDash)
	ret, err := p.ParseArgs([]string{"-v", "--", "-v", "-g"})
	assert.NoError(t, err)

	assert.True(t, opts.Value)

	assert.Equal(t, ret, []string{"-v", "-g"})
}

func TestPassAfterNonOption(t *testing.T) {
	var opts struct {
		Value bool `short:"v"`
	}

	p := NewParser(&opts, PassAfterNonOption)
	ret, err := p.ParseArgs([]string{"-v", "arg", "-v", "-g"})
	assert.NoError(t, err)

	assert.True(t, opts.Value)

	assert.Equal(t, ret, []string{"arg", "-v", "-g"})
}

func TestPassAfterNonOptionWithPositional(t *testing.T) {
	var opts struct {
		Value bool `short:"v"`

		Positional struct {
			Rest []string `required:"yes"`
		} `positional-args:"yes"`
	}

	p := NewParser(&opts, PassAfterNonOption)
	ret, err := p.ParseArgs([]string{"-v", "arg", "-v", "-g"})
	assert.NoError(t, err)

	assert.True(t, opts.Value)

	assert.Equal(t, ret, []string{})
	assert.Equal(t, opts.Positional.Rest, []string{"arg", "-v", "-g"})
}

func TestPassAfterNonOptionWithPositionalIntPass(t *testing.T) {
	var opts struct {
		Value bool `short:"v"`

		Positional struct {
			Rest []int `required:"yes"`
		} `positional-args:"yes"`
	}

	p := NewParser(&opts, PassAfterNonOption)
	ret, err := p.ParseArgs([]string{"-v", "1", "2", "3"})
	assert.NoError(t, err)
	assert.True(t, opts.Value)
	assert.Equal(t, ret, []string{})
	for i, rest := range opts.Positional.Rest {
		assert.Equal(t, i+1, rest)
	}
}

func TestPassAfterNonOptionWithPositionalIntFail(t *testing.T) {
	var opts struct {
		Value bool `short:"v"`

		Positional struct {
			Rest []int `required:"yes"`
		} `positional-args:"yes"`
	}

	tests := []struct {
		opts        []string
		errContains string
		ret         []string
	}{
		{
			[]string{"-v", "notint1", "notint2", "notint3"},
			"notint1",
			[]string{"notint1", "notint2", "notint3"},
		},
		{
			[]string{"-v", "1", "notint2", "notint3"},
			"notint2",
			[]string{"1", "notint2", "notint3"},
		},
	}

	for _, test := range tests {
		p := NewParser(&opts, PassAfterNonOption)
		ret, err := p.ParseArgs(test.opts)
		assert.ErrorContains(t, err, test.errContains)
		assert.Equal(t, ret, test.ret)
	}
}
