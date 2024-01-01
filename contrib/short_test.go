package flags

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestShort(t *testing.T) {
	opts := struct {
		Value bool `short:"v"`
	}{}

	ret := assertParseSuccess(t, &opts, "-v")

	assert.Equal(t, ret, []string{})

	if !opts.Value {
		t.Errorf("Expected Value to be true")
	}
}

func TestShortTooLong(t *testing.T) {
	opts := struct {
		Value bool `short:"vv"`
	}{}

	assertParseFail(t, ErrShortNameTooLong, "short names can only be 1 character long, not `vv'", &opts)
}

func TestShortRequired(t *testing.T) {
	opts := struct {
		Value bool `short:"v" required:"true"`
	}{}

	assertParseFail(t, ErrRequired, fmt.Sprintf("the required flag `%cv' was not specified", defaultShortOptDelimiter), &opts)
}

func TestShortRequiredFalsy1(t *testing.T) {
	opts := struct {
		Value bool `short:"v" required:"false"`
	}{}

	assertParseSuccess(t, &opts)
}

func TestShortRequiredFalsy2(t *testing.T) {
	opts := struct {
		Value bool `short:"v" required:"no"`
	}{}

	assertParseSuccess(t, &opts)
}

func TestShortMultiConcat(t *testing.T) {
	opts := struct {
		V bool `short:"v"`
		O bool `short:"o"`
		F bool `short:"f"`
	}{}

	ret := assertParseSuccess(t, &opts, "-vo", "-f")

	assert.Equal(t, ret, []string{})

	if !opts.V {
		t.Errorf("Expected V to be true")
	}

	if !opts.O {
		t.Errorf("Expected O to be true")
	}

	if !opts.F {
		t.Errorf("Expected F to be true")
	}
}

func TestShortMultiRequiredConcat(t *testing.T) {
	opts := struct {
		V bool `short:"v" required:"true"`
		O bool `short:"o" required:"true"`
		F bool `short:"f" required:"true"`
	}{}

	ret := assertParseSuccess(t, &opts, "-vo", "-f")

	assert.Equal(t, ret, []string{})

	if !opts.V {
		t.Errorf("Expected V to be true")
	}

	if !opts.O {
		t.Errorf("Expected O to be true")
	}

	if !opts.F {
		t.Errorf("Expected F to be true")
	}
}

func TestShortMultiSlice(t *testing.T) {
	opts := struct {
		Values []bool `short:"v"`
	}{}

	ret := assertParseSuccess(t, &opts, "-v", "-v")
	assert.Equal(t, []string{}, ret)
	assert.Equal(t, []bool{true, true}, opts.Values)
}

func TestShortMultiSliceConcat(t *testing.T) {
	opts := struct {
		Values []bool `short:"v"`
	}{}

	ret := assertParseSuccess(t, &opts, "-vvv")
	assert.Equal(t, []string{}, ret)
	assert.Equal(t, []bool{true, true, true}, opts.Values)
}

func TestShortWithEqualArg(t *testing.T) {
	opts := struct {
		Value string `short:"v"`
	}{}

	ret := assertParseSuccess(t, &opts, "-v=value")
	assert.Equal(t, []string{}, ret)
	assert.Equal(t, "value", opts.Value)
}

func TestShortWithArg(t *testing.T) {
	opts := struct {
		Value string `short:"v"`
	}{}

	ret := assertParseSuccess(t, &opts, "-vvalue")
	assert.Equal(t, []string{}, ret)
	assert.Equal(t, "value", opts.Value)
}

func TestShortArg(t *testing.T) {
	opts := struct {
		Value string `short:"v"`
	}{}

	ret := assertParseSuccess(t, &opts, "-v", "value")
	assert.Equal(t, []string{}, ret)
	assert.Equal(t, "value", opts.Value)
}

func TestShortMultiWithEqualArg(t *testing.T) {
	opts := struct {
		F     []bool `short:"f"`
		Value string `short:"v"`
	}{}

	assertParseFail(t, ErrExpectedArgument, fmt.Sprintf("expected argument for flag `%cv'", defaultShortOptDelimiter), &opts, "-ffv=value")
}

func TestShortMultiArg(t *testing.T) {
	opts := struct {
		F     []bool `short:"f"`
		Value string `short:"v"`
	}{}

	ret := assertParseSuccess(t, &opts, "-ffv", "value")
	assert.Equal(t, []string{}, ret)
	assert.Equal(t, []bool{true, true}, opts.F)
	assert.Equal(t, "value", opts.Value)
}

func TestShortMultiArgConcatFail(t *testing.T) {
	opts := struct {
		F     []bool `short:"f"`
		Value string `short:"v"`
	}{}

	assertParseFail(t, ErrExpectedArgument, fmt.Sprintf("expected argument for flag `%cv'", defaultShortOptDelimiter), &opts, "-ffvvalue")
}

func TestShortMultiArgConcat(t *testing.T) {
	opts := struct {
		F     []bool `short:"f"`
		Value string `short:"v"`
	}{}

	ret := assertParseSuccess(t, &opts, "-vff")
	assert.Equal(t, []string{}, ret)
	assert.Equal(t, "ff", opts.Value)
}

func TestShortOptional(t *testing.T) {
	opts := struct {
		F     []bool `short:"f"`
		Value string `short:"v" optional:"yes" optional-value:"value"`
	}{}

	ret := assertParseSuccess(t, &opts, "-fv", "f")
	assert.Equal(t, []string{"f"}, ret)
	assert.Equal(t, "value", opts.Value)
}

func TestShortOptionalFalsy1(t *testing.T) {
	opts := struct {
		F     []bool `short:"f"`
		Value string `short:"v" optional:"false" optional-value:"value"`
	}{}

	ret := assertParseSuccess(t, &opts, "-fv", "f")
	assert.Equal(t, []string{}, ret)
	assert.Equal(t, "f", opts.Value)
}

func TestShortOptionalFalsy2(t *testing.T) {
	opts := struct {
		F     []bool `short:"f"`
		Value string `short:"v" optional:"no" optional-value:"value"`
	}{}

	ret := assertParseSuccess(t, &opts, "-fv", "f")
	assert.Equal(t, []string{}, ret)
	assert.Equal(t, "f", opts.Value)
}
