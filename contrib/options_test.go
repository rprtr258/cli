package flags

import (
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

	ret, err := NewParser(&opts, PassAfterNonOption).ParseArgs("-v", "arg", "-v", "-g")
	assert.NoError(t, err)
	assert.True(t, opts.Value)
	assert.Equal(t, []string{"arg", "-v", "-g"}, ret)
}

func TestPassAfterNonOptionWithPositional(t *testing.T) {
	opts := struct {
		Value bool `short:"v"`

		Positional struct {
			Rest []string `required:"yes"`
		} `positional-args:"yes"`
	}{}

	ret, err := NewParser(&opts, PassAfterNonOption).ParseArgs("-v", "arg", "-v", "-g")
	assert.NoError(t, err)
	assert.True(t, opts.Value)
	assert.Equal(t, []string{}, ret)
	assert.Equal(t, []string{"arg", "-v", "-g"}, opts.Positional.Rest)
}

func TestPassAfterNonOptionWithPositionalIntPass(t *testing.T) {
	opts := struct {
		Value bool `short:"v"`

		Positional struct {
			Rest []int `required:"yes"`
		} `positional-args:"yes"`
	}{}

	ret, err := NewParser(&opts, PassAfterNonOption).ParseArgs("-v", "1", "2", "3")
	assert.NoError(t, err)
	assert.True(t, opts.Value)
	assert.Equal(t, []string{}, ret)
	assert.Equal(t, []int{1, 2, 3}, opts.Positional.Rest)
}

func TestPassAfterNonOptionWithPositionalIntFail(t *testing.T) {
	opts := struct {
		Value bool `short:"v"`

		Positional struct {
			Rest []int `required:"yes"`
		} `positional-args:"yes"`
	}{}

	for _, test := range []struct {
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
	} {
		p := NewParser(&opts, PassAfterNonOption)
		ret, err := p.ParseArgs(test.opts...)
		assert.Error(t, err)
		assert.ErrorContains(t, err, test.errContains)
		assert.Equal(t, ret, test.ret)
	}
}
