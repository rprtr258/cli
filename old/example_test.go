package cli

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func ptr[T any](v T) *T {
	return &v
}

func TestExample(t *testing.T) {
	t.Skip()

	type Args struct {
		ID   string
		Num  int
		Rest []string
	}
	type Opts struct {
		// Slice of bool will append 'true' each time the option
		// is encountered (can be set multiple times, like -vvv)
		Verbose []bool `short:"v" long:"verbose" description:"Show verbose debug information"`

		// Example of automatic marshalling to desired type (uint)
		Offset uint `long:"offset" description:"Offset"`

		// Example of a callback, called each time the option is found.
		Call func(string) `short:"c" description:"Call phone number"`

		// Example of a required flag
		Name string `short:"n" long:"name" description:"A name" required:"true"`

		// Example of a value name
		File string `short:"f" long:"file" description:"A file" value-name:"FILE"`

		// Example of a pointer
		Ptr *int `short:"p" description:"A pointer to an integer"`

		// Example of a slice of strings
		StringSlice []string `short:"s" description:"A slice of strings"`

		// Example of a slice of pointers
		PtrSlice []*string `long:"ptrslice" description:"A slice of pointers to string"`

		// Example of a map
		IntMap map[string]int `long:"intmap" description:"A map from string to int"`

		// Example of a filename (useful for completion)
		Filename Filename `long:"filename" description:"A filename"`

		// Example of positional arguments
		Args Args `positional-args:"yes" required:"yes"`
	}
	var opts Opts

	// Callback which will invoke callto:<argument> to call a number.
	// Note that this works just on OS X (and probably only with
	// Skype) but it shows the idea.
	called := false
	opts.Call = func(num string) {
		assert.Equal(t, "123-456-7890", num)
		called = true
	}

	// Make some fake arguments to parse.
	args := []string{
		"-vv",
		"--offset=5",
		"-c", "123-456-7890",
		"-n", "Me",
		"-p", "3",
		"-s", "hello",
		"-s", "world",
		"--ptrslice", "hello",
		"--ptrslice", "world",
		"--intmap", "a:1",
		"--intmap", "b:5",
		"--filename", "hello.go",
		"id",
		"10",
		"remaining1",
		"remaining2",
	}

	// Parse flags from `args'. Note that here we use cli.ParseArgs for
	// the sake of making a working example. Normally, you would simply use
	// cli.Parse(&opts) which uses os.Args
	_, err := ParseArgs(&opts, args)
	assert.NoError(t, err)

	assert.Equal(t, Opts{
		Verbose:     []bool{true, true},
		Offset:      5,
		Name:        "Me",
		Ptr:         ptr(3),
		StringSlice: []string{"hello", "world"},
		PtrSlice:    []*string{ptr("hello"), ptr("world")},
		IntMap:      map[string]int{"a": 1, "b": 5},
		Filename:    "hello.go",
		Args: Args{
			ID:   "id",
			Num:  10,
			Rest: []string{"remaining1", "remaining2"},
		},

		Call: opts.Call,
		File: "",
	}, opts)
	assert.True(t, called)
}
