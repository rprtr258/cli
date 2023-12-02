package cli

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

type marshalled string

func (m *marshalled) UnmarshalFlag(value string) error {
	switch value {
	case "yes":
		*m = "true"
	case "no":
		*m = "false"
	default:
		return fmt.Errorf("`%s' is not a valid value, please specify `yes' or `no'", value)
	}

	return nil
}

func (m marshalled) MarshalFlag() (string, error) {
	if m == "true" {
		return "yes", nil
	}

	return "no", nil
}

type marshalledError bool

func (m marshalledError) MarshalFlag() (string, error) {
	return "", newErrorf(ErrMarshal, "Failed to marshal")
}

func TestUnmarshal(t *testing.T) {
	var opts struct {
		Value marshalled `short:"v"`
	}

	ret := useParsedArgs(t, &opts, "-v=yes")

	assert.Equal(t, ret, []string{})
	assert.Equal(t, marshalled("true"), opts.Value)
}

func TestUnmarshalDefault(t *testing.T) {
	var opts struct {
		Value marshalled `short:"v" default:"yes"`
	}

	ret := useParsedArgs(t, &opts)
	assert.Equal(t, ret, []string{})
	assert.Equal(t, marshalled("true"), opts.Value)
}

func TestUnmarshalOptional(t *testing.T) {
	var opts struct {
		Value marshalled `short:"v" optional:"yes" optional-value:"yes"`
	}

	ret := useParsedArgs(t, &opts, "-v")
	assert.Equal(t, ret, []string{})
	assert.Equal(t, marshalled("true"), opts.Value)
}

func TestUnmarshalError(t *testing.T) {
	var opts struct {
		Value marshalled `short:"v"`
	}

	assertParseFailed(t, ErrMarshal, fmt.Sprintf("invalid argument for flag `%cv' (expected cli.marshalled): `invalid' is not a valid value, please specify `yes' or `no'", defaultShortOptDelimiter), &opts, "-vinvalid")
}

func TestUnmarshalPositionalError(t *testing.T) {
	var opts struct {
		Args struct {
			Value marshalled
		} `positional-args:"yes"`
	}

	parser := NewParser(&opts, Default&^PrintErrors)
	_, err := parser.ParseArgs([]string{"invalid"})
	assert.EqualError(t, err, "`invalid' is not a valid value, please specify `yes' or `no'")
}

func TestMarshalError(t *testing.T) {
	var opts struct {
		Value marshalledError `short:"v"`
	}

	p := NewParser(&opts, Default)
	o := p.Command.Groups()[0].Options()[0]

	_, err := convertToString(o.value, o.tag)
	assert.Equal(t, err, &Error{
		Type:    ErrMarshal,
		Message: "Failed to marshal",
	})
}
