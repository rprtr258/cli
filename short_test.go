package cli

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestShort(t *testing.T) {
	var opts struct {
		Value bool `short:"v"`
	}

	ret := useParsedArgs(t, &opts, "-v")
	assert.Equal(t, ret, []string{})
	assert.True(t, opts.Value)
}

func TestShortTooLong(t *testing.T) {
	var opts struct {
		Value bool `short:"vv"`
	}

	assertParseFailed(t, ErrShortNameTooLong, "short names can only be 1 character long, not `vv'", &opts)
}

func TestShortRequired(t *testing.T) {
	var opts struct {
		Value bool `short:"v" required:"true"`
	}

	assertParseFailed(t, ErrRequired, fmt.Sprintf("the required flag `%cv' was not specified", defaultShortOptDelimiter), &opts)
}

func TestShortRequiredFalsy1(t *testing.T) {
	var opts struct {
		Value bool `short:"v" required:"false"`
	}

	useParsedArgs(t, &opts)
}

func TestShortRequiredFalsy2(t *testing.T) {
	var opts struct {
		Value bool `short:"v" required:"no"`
	}

	useParsedArgs(t, &opts)
}

func TestShortMultiConcat(t *testing.T) {
	var opts struct {
		V bool `short:"v"`
		O bool `short:"o"`
		F bool `short:"f"`
	}

	ret := useParsedArgs(t, &opts, "-vo", "-f")
	assert.Equal(t, ret, []string{})
	assert.True(t, opts.V)
	assert.True(t, opts.O)
	assert.True(t, opts.F)
}

func TestShortMultiRequiredConcat(t *testing.T) {
	var opts struct {
		V bool `short:"v" required:"true"`
		O bool `short:"o" required:"true"`
		F bool `short:"f" required:"true"`
	}

	ret := useParsedArgs(t, &opts, "-vo", "-f")
	assert.Equal(t, ret, []string{})
	assert.True(t, opts.V)
	assert.True(t, opts.O)
	assert.True(t, opts.F)
}

func TestShortMultiSlice(t *testing.T) {
	var opts struct {
		Values []bool `short:"v"`
	}

	ret := useParsedArgs(t, &opts, "-v", "-v")
	assert.Equal(t, ret, []string{})
	assert.Equal(t, opts.Values, []bool{true, true})
}

func TestShortMultiSliceConcat(t *testing.T) {
	var opts struct {
		Values []bool `short:"v"`
	}

	ret := useParsedArgs(t, &opts, "-vvv")
	assert.Equal(t, ret, []string{})
	assert.Equal(t, opts.Values, []bool{true, true, true})
}

func TestShortWithEqualArg(t *testing.T) {
	var opts struct {
		Value string `short:"v"`
	}

	ret := useParsedArgs(t, &opts, "-v=value")
	assert.Equal(t, ret, []string{})
	assert.Equal(t, opts.Value, "value")
}

func TestShortWithArg(t *testing.T) {
	var opts struct {
		Value string `short:"v"`
	}

	ret := useParsedArgs(t, &opts, "-vvalue")
	assert.Equal(t, ret, []string{})
	assert.Equal(t, opts.Value, "value")
}

func TestShortArg(t *testing.T) {
	var opts struct {
		Value string `short:"v"`
	}

	ret := useParsedArgs(t, &opts, "-v", "value")
	assert.Equal(t, ret, []string{})
	assert.Equal(t, opts.Value, "value")
}

func TestShortMultiWithEqualArg(t *testing.T) {
	var opts struct {
		F     []bool `short:"f"`
		Value string `short:"v"`
	}

	assertParseFailed(t, ErrExpectedArgument, fmt.Sprintf("expected argument for flag `%cv'", defaultShortOptDelimiter), &opts, "-ffv=value")
}

func TestShortMultiArg(t *testing.T) {
	var opts struct {
		F     []bool `short:"f"`
		Value string `short:"v"`
	}

	ret := useParsedArgs(t, &opts, "-ffv", "value")
	assert.Equal(t, ret, []string{})
	assert.Equal(t, opts.F, []bool{true, true})
	assert.Equal(t, opts.Value, "value")
}

func TestShortMultiArgConcatFail(t *testing.T) {
	var opts struct {
		F     []bool `short:"f"`
		Value string `short:"v"`
	}

	assertParseFailed(t, ErrExpectedArgument, fmt.Sprintf("expected argument for flag `%cv'", defaultShortOptDelimiter), &opts, "-ffvvalue")
}

func TestShortMultiArgConcat(t *testing.T) {
	var opts struct {
		F     []bool `short:"f"`
		Value string `short:"v"`
	}

	ret := useParsedArgs(t, &opts, "-vff")
	assert.Equal(t, ret, []string{})
	assert.Equal(t, opts.Value, "ff")
}

func TestShortOptional(t *testing.T) {
	var opts struct {
		F     []bool `short:"f"`
		Value string `short:"v" optional:"yes" optional-value:"value"`
	}

	ret := useParsedArgs(t, &opts, "-fv", "f")
	assert.Equal(t, ret, []string{"f"})
	assert.Equal(t, opts.Value, "value")
}

func TestShortOptionalFalsy1(t *testing.T) {
	var opts struct {
		F     []bool `short:"f"`
		Value string `short:"v" optional:"false" optional-value:"value"`
	}

	ret := useParsedArgs(t, &opts, "-fv", "f")
	assert.Equal(t, ret, []string{})
	assert.Equal(t, opts.Value, "f")
}

func TestShortOptionalFalsy2(t *testing.T) {
	var opts struct {
		F     []bool `short:"f"`
		Value string `short:"v" optional:"no" optional-value:"value"`
	}

	ret := useParsedArgs(t, &opts, "-fv", "f")
	assert.Equal(t, ret, []string{})
	assert.Equal(t, opts.Value, "f")
}
