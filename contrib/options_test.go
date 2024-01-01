package flags

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPassDoubleDash(t *testing.T) {
	opts := struct {
		Value bool `short:"v"`
	}{}

	ret, err := NewParser(&opts, PassDoubleDash).ParseArgs("-v", "--", "-v", "-g")
	assert.NoError(t, err)
	assert.True(t, opts.Value)
	assert.Equal(t, []string{"-v", "-g"}, ret)
}

func TestPassAfterNonOption(t *testing.T) {
	opts := struct {
		Value bool `short:"v"`
	}{}

	p := NewParser(&opts, PassAfterNonOption)
	ret, err := p.ParseArgs("-v", "arg", "-v", "-g")
	assert.NoError(t, err)
	if !opts.Value {
		t.Errorf("Expected Value to be true")
	}

	assert.Equal(t, ret, []string{"arg", "-v", "-g"})
}

func TestPassAfterNonOptionWithPositional(t *testing.T) {
	opts := struct {
		Value bool `short:"v"`

		Positional struct {
			Rest []string `required:"yes"`
		} `positional-args:"yes"`
	}{}

	p := NewParser(&opts, PassAfterNonOption)
	ret, err := p.ParseArgs("-v", "arg", "-v", "-g")
	assert.NoError(t, err)
	if !opts.Value {
		t.Errorf("Expected Value to be true")
	}

	assert.Equal(t, ret, []string{})
	assert.Equal(t, opts.Positional.Rest, []string{"arg", "-v", "-g"})
}

func TestPassAfterNonOptionWithPositionalIntPass(t *testing.T) {
	opts := struct {
		Value bool `short:"v"`

		Positional struct {
			Rest []int `required:"yes"`
		} `positional-args:"yes"`
	}{}

	p := NewParser(&opts, PassAfterNonOption)
	ret, err := p.ParseArgs("-v", "1", "2", "3")
	assert.NoError(t, err)
	if !opts.Value {
		t.Errorf("Expected Value to be true")
	}

	assert.Equal(t, ret, []string{})
	for i, rest := range opts.Positional.Rest {
		if rest != i+1 {
			assertErrorf(t, "Expected %v got %v", i+1, rest)
		}
	}
}

func TestPassAfterNonOptionWithPositionalIntFail(t *testing.T) {
	opts := struct {
		Value bool `short:"v"`

		Positional struct {
			Rest []int `required:"yes"`
		} `positional-args:"yes"`
	}{}

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
		ret, err := p.ParseArgs(test.opts...)
		assert.Error(t, err)
		if !strings.Contains(err.Error(), test.errContains) {
			assertErrorf(t, "Expected the first illegal argument in the error")
		}

		assert.Equal(t, ret, test.ret)
	}
}
