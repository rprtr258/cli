package cli

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPointerBool(t *testing.T) {
	var opts struct {
		Value *bool `short:"v"`
	}

	ret := assertParseSuccess(t, &opts, "-v")
	assert.Equal(t, ret, []string{})
	assert.True(t, *opts.Value)
}

func TestPointerString(t *testing.T) {
	var opts struct {
		Value *string `short:"v"`
	}

	ret := assertParseSuccess(t, &opts, "-v", "value")
	assert.Equal(t, ret, []string{})
	assert.Equal(t, *opts.Value, "value")
}

func TestPointerSlice(t *testing.T) {
	var opts struct {
		Value *[]string `short:"v"`
	}

	ret := assertParseSuccess(t, &opts, "-v", "value1", "-v", "value2")
	assert.Equal(t, ret, []string{})
	assert.Equal(t, *opts.Value, []string{"value1", "value2"})
}

func TestPointerMap(t *testing.T) {
	var opts struct {
		Value *map[string]int `short:"v"`
	}

	ret := assertParseSuccess(t, &opts, "-v", "k1:2", "-v", "k2:-5")
	assert.Equal(t, ret, []string{})
	assert.Equal(t, map[string]int{
		"k1": 2,
		"k2": -5,
	}, *opts.Value)
}

type marshalledString string

func (m *marshalledString) UnmarshalFlag(value string) error {
	*m = marshalledString(value)
	return nil
}

func (m marshalledString) MarshalFlag() (string, error) {
	return string(m), nil
}

func TestPointerStringMarshalled(t *testing.T) {
	var opts struct {
		Value *marshalledString `short:"v"`
	}

	ret := assertParseSuccess(t, &opts, "-v", "value")
	assert.Equal(t, ret, []string{})
	assert.NotNil(t, opts.Value)
	assert.Equal(t, string(*opts.Value), "value")
}

type marshalledStruct struct {
	Value string
}

func (m *marshalledStruct) UnmarshalFlag(value string) error {
	m.Value = value
	return nil
}

func (m marshalledStruct) MarshalFlag() (string, error) {
	return m.Value, nil
}

func TestPointerStructMarshalled(t *testing.T) {
	var opts struct {
		Value *marshalledStruct `short:"v"`
	}

	ret := assertParseSuccess(t, &opts, "-v", "value")
	assert.Equal(t, ret, []string{})
	assert.NotNil(t, opts.Value)
	assert.Equal(t, opts.Value.Value, "value")
}

type PointerGroup struct {
	Value bool `short:"v"`
}

func TestPointerGroup(t *testing.T) {
	var opts struct {
		Group *PointerGroup `group:"Group Options"`
	}

	ret := assertParseSuccess(t, &opts, "-v")
	assert.Equal(t, ret, []string{})
	assert.True(t, opts.Group.Value)
}

func TestDoNotChangeNonTaggedFields(t *testing.T) {
	var opts struct {
		A struct {
			Pointer *int
		}
		B *struct {
			Pointer *int
		}
	}

	ret := assertParseSuccess(t, &opts)
	assert.Equal(t, ret, []string{})
	assert.Nil(t, opts.A.Pointer)
	assert.Nil(t, opts.B)
}
