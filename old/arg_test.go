package cli

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPositional(t *testing.T) {
	var opts struct {
		Value      bool `short:"v"`
		Positional struct {
			Command  int
			Filename string
			Rest     []string
		} `positional-args:"yes" required:"yes"`
	}

	p := NewParser(&opts, Default)
	ret, err := p.ParseArgs([]string{"10", "arg_test.go", "a", "b"})
	assert.NoError(t, err)
	assert.Equal(t, 10, opts.Positional.Command)
	assert.Equal(t, "arg_test.go", opts.Positional.Filename)
	assert.Equal(t, opts.Positional.Rest, []string{"a", "b"})
	assert.Equal(t, ret, []string{})
}

func TestPositionalRequired(t *testing.T) {
	var opts struct {
		Value      bool `short:"v"`
		Positional struct {
			Command  int
			Filename string
			Rest     []string
		} `positional-args:"yes" required:"yes"`
	}

	p := NewParser(&opts, None)
	_, err := p.ParseArgs([]string{"10"})
	assert.Error(t, err, &Error{
		Kind:    ErrRequired,
		Message: "the required argument `Filename` was not provided",
	})
}

func TestPositionalRequiredRest1Fail(t *testing.T) {
	var opts struct {
		Value      bool `short:"v"`
		Positional struct {
			Rest []string `required:"yes"`
		} `positional-args:"yes"`
	}

	p := NewParser(&opts, None)
	_, err := p.ParseArgs([]string{})
	assert.Error(t, err, ErrRequired, "the required argument `Rest (at least 1 argument)` was not provided")
}

func TestPositionalRequiredRest1Pass(t *testing.T) {
	var opts struct {
		Value      bool `short:"v"`
		Positional struct {
			Rest []string `required:"yes"`
		} `positional-args:"yes"`
	}

	p := NewParser(&opts, None)
	_, err := p.ParseArgs([]string{"rest1"})
	assert.NoError(t, err)
	assert.Equal(t, 1, len(opts.Positional.Rest))
	assert.Equal(t, opts.Positional.Rest[0], "rest1")
}

func TestPositionalRequiredRest2Fail(t *testing.T) {
	var opts struct {
		Value      bool `short:"v"`
		Positional struct {
			Rest []string `required:"2"`
		} `positional-args:"yes"`
	}

	p := NewParser(&opts, None)
	_, err := p.ParseArgs([]string{"rest1"})
	assert.Error(t, err, ErrRequired, "the required argument `Rest (at least 2 arguments, but got only 1)` was not provided")
}

func TestPositionalRequiredRest2Pass(t *testing.T) {
	var opts struct {
		Value      bool `short:"v"`
		Positional struct {
			Rest []string `required:"2"`
		} `positional-args:"yes"`
	}

	p := NewParser(&opts, None)
	_, err := p.ParseArgs([]string{"rest1", "rest2", "rest3"})
	assert.NoError(t, err)
	assert.Equal(t, 3, len(opts.Positional.Rest))
	assert.Equal(t, opts.Positional.Rest[0], "rest1")
	assert.Equal(t, opts.Positional.Rest[1], "rest2")
	assert.Equal(t, opts.Positional.Rest[2], "rest3")
}

func TestPositionalRequiredRestRangeFail(t *testing.T) {
	var opts struct {
		Value      bool `short:"v"`
		Positional struct {
			Rest []string `required:"1-2"`
		} `positional-args:"yes"`
	}

	p := NewParser(&opts, None)
	_, err := p.ParseArgs([]string{"rest1", "rest2", "rest3"})
	assert.Error(t, err, &Error{
		Kind:    ErrRequired,
		Message: "the required argument `Rest (at most 2 arguments, but got 3)` was not provided",
	})
}

func TestPositionalRequiredRestRangeEmptyFail(t *testing.T) {
	var opts struct {
		Value      bool `short:"v"`
		Positional struct {
			Rest []string `required:"0-0"`
		} `positional-args:"yes"`
	}

	p := NewParser(&opts, None)
	_, err := p.ParseArgs([]string{"some", "thing"})
	assert.Error(t, err, ErrRequired, "the required argument `Rest (zero arguments)` was not provided")
}
