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

	ret := assertParserSuccess(t, &opts, "-v")
	assert.Equal(t, []string{}, ret)
	assert.True(t, opts.Value)
}

func TestShortTooLong(t *testing.T) {
	opts := struct {
		Value bool `short:"vv"`
	}{}

	assertParseFail(t, &Error{
		Kind:    ErrShortNameTooLong,
		Message: "short names can only be 1 character long, not `vv'",
	}, &opts)
}

func TestShortRequired(t *testing.T) {
	opts := struct {
		Value bool `short:"v" required:"true"`
	}{}

	assertParseFail(t, &Error{
		Kind:    ErrRequired,
		Message: fmt.Sprintf("the required flag `%cv' was not specified", _defaultOptDelimiterShort),
	}, &opts)
}

func TestShortRequiredFalsy1(t *testing.T) {
	opts := struct {
		Value bool `short:"v" required:"false"`
	}{}

	assertParserSuccess(t, &opts)
}

func TestShortRequiredFalsy2(t *testing.T) {
	opts := struct {
		Value bool `short:"v" required:"no"`
	}{}

	assertParserSuccess(t, &opts)
}

func TestShortMultiConcat(t *testing.T) {
	opts := struct {
		V bool `short:"v"`
		O bool `short:"o"`
		F bool `short:"f"`
	}{}

	ret := assertParserSuccess(t, &opts, "-vo", "-f")
	assert.Equal(t, []string{}, ret)
	assert.True(t, opts.V)
	assert.True(t, opts.O)
	assert.True(t, opts.F)
}

func TestShortMultiRequiredConcat(t *testing.T) {
	opts := struct {
		V bool `short:"v" required:"true"`
		O bool `short:"o" required:"true"`
		F bool `short:"f" required:"true"`
	}{}

	ret := assertParserSuccess(t, &opts, "-vo", "-f")
	assert.Equal(t, []string{}, ret)
	assert.True(t, opts.V)
	assert.True(t, opts.O)
	assert.True(t, opts.F)
}

func TestShortMultiSlice(t *testing.T) {
	opts := struct {
		Values []bool `short:"v"`
	}{}

	ret := assertParserSuccess(t, &opts, "-v", "-v")
	assert.Equal(t, []string{}, ret)
	assert.Equal(t, []bool{true, true}, opts.Values)
}

func TestShortMultiSliceConcat(t *testing.T) {
	opts := struct {
		Values []bool `short:"v"`
	}{}

	ret := assertParserSuccess(t, &opts, "-vvv")
	assert.Equal(t, []string{}, ret)
	assert.Equal(t, []bool{true, true, true}, opts.Values)
}

func TestShortWithEqualArg(t *testing.T) {
	opts := struct {
		Value string `short:"v"`
	}{}

	ret := assertParserSuccess(t, &opts, "-v=value")
	assert.Equal(t, []string{}, ret)
	assert.Equal(t, "value", opts.Value)
}

func TestShortWithArg(t *testing.T) {
	opts := struct {
		Value string `short:"v"`
	}{}

	ret := assertParserSuccess(t, &opts, "-vvalue")
	assert.Equal(t, []string{}, ret)
	assert.Equal(t, "value", opts.Value)
}

func TestShortArg(t *testing.T) {
	opts := struct {
		Value string `short:"v"`
	}{}

	ret := assertParserSuccess(t, &opts, "-v", "value")
	assert.Equal(t, []string{}, ret)
	assert.Equal(t, "value", opts.Value)
}

func TestShortMultiWithEqualArg(t *testing.T) {
	opts := struct {
		F     []bool `short:"f"`
		Value string `short:"v"`
	}{}

	assertParseFail(t, &Error{
		Kind:    ErrExpectedArgument,
		Message: fmt.Sprintf("expected argument for flag `%cv'", _defaultOptDelimiterShort),
	}, &opts, "-ffv=value")
}

func TestShortMultiArg(t *testing.T) {
	opts := struct {
		F     []bool `short:"f"`
		Value string `short:"v"`
	}{}

	ret := assertParserSuccess(t, &opts, "-ffv", "value")
	assert.Equal(t, []string{}, ret)
	assert.Equal(t, []bool{true, true}, opts.F)
	assert.Equal(t, "value", opts.Value)
}

func TestShortMultiArgConcatFail(t *testing.T) {
	opts := struct {
		F     []bool `short:"f"`
		Value string `short:"v"`
	}{}

	assertParseFail(t, &Error{
		Kind:    ErrExpectedArgument,
		Message: fmt.Sprintf("expected argument for flag `%cv'", _defaultOptDelimiterShort),
	}, &opts, "-ffvvalue")
}

func TestShortMultiArgConcat(t *testing.T) {
	opts := struct {
		F     []bool `short:"f"`
		Value string `short:"v"`
	}{}

	ret := assertParserSuccess(t, &opts, "-vff")
	assert.Equal(t, []string{}, ret)
	assert.Equal(t, "ff", opts.Value)
}

func TestShortOptional(t *testing.T) {
	opts := struct {
		F     []bool `short:"f"`
		Value string `short:"v" optional:"yes" optional-value:"value"`
	}{}

	ret := assertParserSuccess(t, &opts, "-fv", "f")
	assert.Equal(t, []string{"f"}, ret)
	assert.Equal(t, "value", opts.Value)
}

func TestShortOptionalFalsy1(t *testing.T) {
	opts := struct {
		F     []bool `short:"f"`
		Value string `short:"v" optional:"false" optional-value:"value"`
	}{}

	ret := assertParserSuccess(t, &opts, "-fv", "f")
	assert.Equal(t, []string{}, ret)
	assert.Equal(t, "f", opts.Value)
}

func TestShortOptionalFalsy2(t *testing.T) {
	opts := struct {
		F     []bool `short:"f"`
		Value string `short:"v" optional:"no" optional-value:"value"`
	}{}

	ret := assertParserSuccess(t, &opts, "-fv", "f")
	assert.Equal(t, []string{}, ret)
	assert.Equal(t, "f", opts.Value)
}
