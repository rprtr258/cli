package flags

import (
	"fmt"
	"testing"

	"github.com/rprtr258/fun"
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
	return fun.IF(m == "true", "yes", "no"), nil
}

type marshalledError bool

func (m marshalledError) MarshalFlag() (string, error) {
	return "", newErrorf(ErrMarshal, "Failed to marshal")
}

func TestUnmarshal(t *testing.T) {
	opts := struct {
		Value marshalled `short:"v"`
	}{}

	assertParserSuccess(t, &opts, "-v=yes")
	assert.Equal(t, marshalled("true"), opts.Value)
}

func TestUnmarshalDefault(t *testing.T) {
	opts := struct {
		Value marshalled `short:"v" default:"yes"`
	}{}

	assertParserSuccess(t, &opts)
	assert.Equal(t, marshalled("true"), opts.Value)
}

func TestUnmarshalOptional(t *testing.T) {
	opts := struct {
		Value marshalled `short:"v" optional:"yes" optional-value:"yes"`
	}{}

	assertParserSuccess(t, &opts, "-v")
	assert.Equal(t, marshalled("true"), opts.Value)
}

func TestUnmarshalError(t *testing.T) {
	opts := struct {
		Value marshalled `short:"v"`
	}{}

	assertParseFail(t, newErrorf(
		ErrMarshal,
		"invalid argument for flag `%cv' (expected flags.marshalled): `invalid' is not a valid value, please specify `yes' or `no'",
		_defaultOptDelimiterShort,
	), &opts, "-vinvalid")
}

func TestUnmarshalPositionalError(t *testing.T) {
	opts := struct {
		Args struct {
			Value marshalled
		} `positional-args:"yes"`
	}{}

	err := Run(&opts, "invalid")
	assert.Error(t, err)
	assert.EqualError(t, err, "`invalid' is not a valid value, please specify `yes' or `no'")
}

func TestMarshalError(t *testing.T) {
	opts := struct {
		Value marshalledError `short:"v"`
	}{}

	o := newParser("", &opts).Flags.flags[0]
	_, err := convertToString(o.value, o.tag)
	assert.Equal(t, &Error{
		Kind:    ErrMarshal,
		Message: "Failed to marshal",
	}, err)
}
