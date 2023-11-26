package cli

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLong(t *testing.T) {
	var opts struct {
		Value bool `long:"value"`
	}

	ret := assertParseSuccess(t, &opts, "--value")

	assert.Equal(t, ret, []string{})

	assert.True(t, opts.Value)
}

func TestLongArg(t *testing.T) {
	var opts struct {
		Value string `long:"value"`
	}

	ret := assertParseSuccess(t, &opts, "--value", "value")

	assert.Equal(t, ret, []string{})
	assert.Equal(t, opts.Value, "value")
}

func TestLongArgEqual(t *testing.T) {
	var opts struct {
		Value string `long:"value"`
	}

	ret := assertParseSuccess(t, &opts, "--value=value")

	assert.Equal(t, ret, []string{})
	assert.Equal(t, opts.Value, "value")
}

func TestLongDefault(t *testing.T) {
	var opts struct {
		Value string `long:"value" default:"value"`
	}

	ret := assertParseSuccess(t, &opts)

	assert.Equal(t, ret, []string{})
	assert.Equal(t, opts.Value, "value")
}

func TestLongOptional(t *testing.T) {
	var opts struct {
		Value string `long:"value" optional:"yes" optional-value:"value"`
	}

	ret := assertParseSuccess(t, &opts, "--value")

	assert.Equal(t, ret, []string{})
	assert.Equal(t, opts.Value, "value")
}

func TestLongOptionalArg(t *testing.T) {
	var opts struct {
		Value string `long:"value" optional:"yes" optional-value:"value"`
	}

	ret := assertParseSuccess(t, &opts, "--value", "no")

	assert.Equal(t, ret, []string{"no"})
	assert.Equal(t, opts.Value, "value")
}

func TestLongOptionalArgEqual(t *testing.T) {
	var opts struct {
		Value string `long:"value" optional:"yes" optional-value:"value"`
	}

	ret := assertParseSuccess(t, &opts, "--value=value", "no")

	assert.Equal(t, ret, []string{"no"})
	assert.Equal(t, opts.Value, "value")
}
