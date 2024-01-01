package flags

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPositional(t *testing.T) {
	opts := struct {
		Value bool `short:"v"`

		Positional struct {
			Command  int
			Filename string
			Rest     []string
		} `positional-args:"yes" required:"yes"`
	}{}

	p := NewParser(&opts, Default)
	ret, err := p.ParseArgs("10", "arg_test.go", "a", "b")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
		return
	}

	if opts.Positional.Command != 10 {
		t.Fatalf("Expected opts.Positional.Command to be 10, but got %v", opts.Positional.Command)
	}

	if opts.Positional.Filename != "arg_test.go" {
		t.Fatalf("Expected opts.Positional.Filename to be \"arg_test.go\", but got %v", opts.Positional.Filename)
	}

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
	assertError(t, err, ErrRequired, "the required argument `Filename` was not provided")
}

func TestPositionalRequiredRest1Fail(t *testing.T) {
	opts := struct {
		Value      bool `short:"v"`
		Positional struct {
			Rest []string `required:"yes"`
		} `positional-args:"yes"`
	}{}

	_, err := NewParser(&opts, None).ParseArgs()
	assertError(t, err, ErrRequired, "the required argument `Rest (at least 1 argument)` was not provided")
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

	p := NewParser(&opts, None)
	_, err := p.ParseArgs("rest1")
	assertError(t, err, ErrRequired, "the required argument `Rest (at least 2 arguments, but got only 1)` was not provided")
}

func TestPositionalRequiredRest2Pass(t *testing.T) {
	opts := struct {
		Value bool `short:"v"`

		Positional struct {
			Rest []string `required:"2"`
		} `positional-args:"yes"`
	}{}

	p := NewParser(&opts, None)
	_, err := p.ParseArgs("rest1", "rest2", "rest3")
	assert.NoError(t, err)
	assert.Equal(t, []string{"rest1", "rest2", "rest3"}, opts.Positional.Rest)
}

func TestPositionalRequiredRestRangeFail(t *testing.T) {
	opts := struct {
		Value bool `short:"v"`

		Positional struct {
			Rest []string `required:"1-2"`
		} `positional-args:"yes"`
	}{}

	p := NewParser(&opts, None)
	_, err := p.ParseArgs("rest1", "rest2", "rest3")

	assertError(t, err, ErrRequired, "the required argument `Rest (at most 2 arguments, but got 3)` was not provided")
}

func TestPositionalRequiredRestRangeEmptyFail(t *testing.T) {
	opts := struct {
		Value bool `short:"v"`

		Positional struct {
			Rest []string `required:"0-0"`
		} `positional-args:"yes"`
	}{}

	p := NewParser(&opts, None)
	_, err := p.ParseArgs("some", "thing")

	assertError(t, err, ErrRequired, "the required argument `Rest (zero arguments)` was not provided")
}
