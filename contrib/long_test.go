package flags

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLong(t *testing.T) {
	opts := struct {
		Value bool `long:"value"`
	}{}

	ret := assertParseSuccess(t, &opts, "--value")
	assert.Equal(t, []string{}, ret)
	assert.True(t, opts.Value)
}

func TestLongArg(t *testing.T) {
	opts := struct {
		Value string `long:"value"`
	}{}

	ret := assertParseSuccess(t, &opts, "--value", "value")
	assert.Equal(t, []string{}, ret)
	assert.Equal(t, "value", opts.Value)
}

func TestLongArgEqual(t *testing.T) {
	opts := struct {
		Value string `long:"value"`
	}{}

	ret := assertParseSuccess(t, &opts, "--value=value")
	assert.Equal(t, []string{}, ret)
	assert.Equal(t, "value", opts.Value)
}

func TestLongDefault(t *testing.T) {
	opts := struct {
		Value string `long:"value" default:"value"`
	}{}

	ret := assertParseSuccess(t, &opts)
	assert.Equal(t, []string{}, ret)
	assert.Equal(t, "value", opts.Value)
}

func TestLongOptional(t *testing.T) {
	opts := struct {
		Value string `long:"value" optional:"yes" optional-value:"value"`
	}{}

	ret := assertParseSuccess(t, &opts, "--value")
	assert.Equal(t, []string{}, ret)
	assert.Equal(t, "value", opts.Value)
}

func TestLongOptionalArg(t *testing.T) {
	opts := struct {
		Value string `long:"value" optional:"yes" optional-value:"value"`
	}{}

	ret := assertParseSuccess(t, &opts, "--value", "no")
	assert.Equal(t, []string{"no"}, ret)
	assert.Equal(t, "value", opts.Value)
}

func TestLongOptionalArgEqual(t *testing.T) {
	opts := struct {
		Value string `long:"value" optional:"yes" optional-value:"value"`
	}{}

	ret := assertParseSuccess(t, &opts, "--value=value", "no")
	assert.Equal(t, []string{"no"}, ret)
	assert.Equal(t, "value", opts.Value)
}
