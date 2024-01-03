package flags

// import (
// 	"testing"

// 	"github.com/stretchr/testify/assert"
// )

// func TestPointerBool(t *testing.T) {
// 	opts := struct {
// 		Value *bool `short:"v"`
// 	}{}

// 	assertNoError(t, &opts, "-v")
// 	assert.True(t, *opts.Value)
// }

// func TestPointerString(t *testing.T) {
// 	opts := struct {
// 		Value *string `short:"v"`
// 	}{}

// 	assertNoError(t, &opts, "-v", "value")
// 	assert.Equal(t, "value", *opts.Value)
// }

// func TestPointerSlice(t *testing.T) {
// 	opts := struct {
// 		Value *[]string `short:"v"`
// 	}{}

// 	assertNoError(t, &opts, "-v", "value1", "-v", "value2")
// 	assert.Equal(t, []string{"value1", "value2"}, *opts.Value)
// }

// func TestPointerMap(t *testing.T) {
// 	opts := struct {
// 		Value *map[string]int `short:"v"`
// 	}{}

// 	assertNoError(t, &opts, "-v", "k1:2", "-v", "k2:-5")
// 	assert.Equal(t, 2, (*opts.Value)["k1"])
// 	assert.Equal(t, -5, (*opts.Value)["k2"])
// }

// type marshalledString string

// func (m *marshalledString) UnmarshalFlag(value string) error {
// 	*m = marshalledString(value)
// 	return nil
// }

// func (m marshalledString) MarshalFlag() (string, error) {
// 	return string(m), nil
// }

// func TestPointerStringMarshalled(t *testing.T) {
// 	opts := struct {
// 		Value *marshalledString `short:"v"`
// 	}{}

// 	assertNoError(t, &opts, "-v", "value")
// 	assert.NotNil(t, opts.Value)
// 	assert.Equal(t, "value", string(*opts.Value))
// }

// type marshalledStruct struct {
// 	Value string
// }

// func (m *marshalledStruct) UnmarshalFlag(value string) error {
// 	m.Value = value
// 	return nil
// }

// func (m marshalledStruct) MarshalFlag() (string, error) {
// 	return m.Value, nil
// }

// func TestPointerStructMarshalled(t *testing.T) {
// 	opts := struct {
// 		Value *marshalledStruct `short:"v"`
// 	}{}

// 	assertNoError(t, &opts, "-v", "value")
// 	assert.NotNil(t, opts.Value)
// 	assert.Equal(t, "value", opts.Value.Value)
// }

// type PointerGroup struct {
// 	Value bool `short:"v"`
// }

// func TestPointerGroup(t *testing.T) {
// 	opts := struct {
// 		PointerGroup
// 	}{}

// 	assertNoError(t, &opts, "-v")
// 	assert.True(t, opts.Value)
// }

// func TestDoNotChangeNonTaggedFields(t *testing.T) {
// 	var opts struct {
// 		A struct {
// 			Pointer *int
// 		}
// 		B *struct {
// 			Pointer *int
// 		}
// 	}

// 	assertNoError(t, &opts)
// 	assert.Nil(t, opts.A.Pointer)
// 	assert.Nil(t, opts.B)
// }
