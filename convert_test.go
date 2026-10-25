package cli

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// A port of the old convert_test.go, which tested `convertToString`: value to
// string conversion used for help rendering. The helper is gone; defaults are
// rendered with `fmt.Sprintf("%v", ...)`, so the port asserts the
// `[default: ...]` fragments of the `--help` output produced by Construct.

func TestConvertToString(t *testing.T) {
	type opts struct {
		String string `cli:"long=string,default=string"`

		Int   int   `cli:"long=int,default=-2"`
		Int8  int8  `cli:"long=int8,default=-1"`
		Int16 int16 `cli:"long=int16,default=0"`
		Int32 int32 `cli:"long=int32,default=1"`
		Int64 int64 `cli:"long=int64,default=2"`

		Uint   uint   `cli:"long=uint,default=1"`
		Uint8  uint8  `cli:"long=uint8,default=2"`
		Uint16 uint16 `cli:"long=uint16,default=3"`
		Uint32 uint32 `cli:"long=uint32,default=4"`
		Uint64 uint64 `cli:"long=uint64,default=5"`

		Float32 float32 `cli:"long=float32,default=1.2"`
		Float64 float64 `cli:"long=float64,default=-3.4"`

		Duration time.Duration `cli:"long=duration,default=1h2m4s"`

		IntFloatMap map[int]float64 `cli:"long=int-float-map,default=-2:4.5"`
	}

	parser := Construct[opts]().ToOptions()
	_, err := parser.RunE([]string{"prog", "--help"})
	help := unwrapStdout(err.(FailureError).failure)

	for _, expect := range []string{
		"[default: string]",
		"[default: -2]",
		"[default: -1]",
		"[default: 0]",
		"[default: 1]",
		"[default: 2]",
		"[default: 1]",
		"[default: 2]",
		"[default: 3]",
		"[default: 4]",
		"[default: 5]",
		"[default: 1.2]",
		"[default: -3.4]",
		"[default: 1h2m4s]",
	} {
		require.Contains(t, help, expect)
	}

	// The map default is rendered too, check it is present at all.
	require.Contains(t, help, "[default:")
}

func TestConvertToStringBoolDefault(t *testing.T) {
	type opts struct {
		Bool bool `cli:"long=bool,default=true"`
	}

	require.Equal(t,
		`cli.Construct: field "Bool": bool field cannot have a default`,
		catchPanic[string](t, func() { Construct[opts]() }))
}

func TestConvertToStringSliceDefault(t *testing.T) {
	type opts struct {
		IntSlice []int `cli:"long=int-slice,default=1"`
	}

	require.Equal(t,
		`cli.Construct: field "IntSlice": []E and cli.Opt[E] fields cannot have a default`,
		catchPanic[string](t, func() { Construct[opts]() }))
}

func TestConvertToStringPointerDefault(t *testing.T) {
	type opts struct {
		PtrBool *bool `cli:"long=ptr-bool,default=true"`
	}

	require.Equal(t,
		`cli.Construct: field "PtrBool": *bool and *cli.Count fields are not supported, use bool, cli.Count or cli.Opt[...]`,
		catchPanic[string](t, func() { Construct[opts]() }))
}

func TestConvertToStringInterfaceField(t *testing.T) {
	type opts struct {
		Interface any `cli:"long=interface"`
	}

	require.Equal(t,
		`cli.Construct: field "Interface": unsupported field type interface {}, implement encoding.TextUnmarshaler to parse custom types`,
		catchPanic[string](t, func() { Construct[opts]() }))
}
