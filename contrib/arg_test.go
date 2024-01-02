package flags

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPositional(t *testing.T) {
	opts := struct {
		Value      bool `short:"v"`
		Positional struct {
			Command  int
			Filename string
			Rest     []string
		} `positional-args:"yes" required:"yes"`
	}{}

	ret, err := NewParser(&opts, Default).ParseArgs("10", "arg_test.go", "a", "b")
	assert.NoError(t, err)
	assert.Equal(t, 10, opts.Positional.Command)
	assert.Equal(t, "arg_test.go", opts.Positional.Filename)
	assert.Equal(t, []string{"a", "b"}, opts.Positional.Rest)
	assert.Equal(t, []string{}, ret)
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

	_, err := NewParser(&opts, None).ParseArgs("10")
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

	_, err := NewParser(&opts, None).ParseArgs()
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

	p := NewParser(&opts, None)
	_, err := p.ParseArgs("rest1")
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

	_, err := NewParser(&opts, None).ParseArgs("rest1")
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

	_, err := NewParser(&opts, None).ParseArgs("rest1", "rest2", "rest3")
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

	_, err := NewParser(&opts, None).ParseArgs("rest1", "rest2", "rest3")
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

	_, err := NewParser(&opts, None).ParseArgs("some", "thing")
	assert.Equal(t, &Error{
		Kind:    ErrRequired,
		Message: "the required argument `Rest (zero arguments)` was not provided",
	}, err)
}
