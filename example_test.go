package cli

import (
	"context"
	"fmt"

	"github.com/rprtr258/cli/flags"
)

type cmdRoot struct {
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
	Filename flags.Filename `long:"filename" description:"A filename"`
	// Example of positional arguments
	Args struct {
		ID   string
		Num  int
		Rest []string
	} `positional-args:"yes" required:"yes"`
}

func (x cmdRoot) Execute(context.Context) error {
	fmt.Println("Verbosity:", x.Verbose)
	fmt.Println("Offset:", x.Offset)
	fmt.Println("Name:", x.Name)
	fmt.Println("Ptr:", *x.Ptr)
	fmt.Println("StringSlice:", x.StringSlice)
	fmt.Println("PtrSlice:", *x.PtrSlice[0], *x.PtrSlice[1])
	fmt.Println("IntMap:", x.IntMap)
	fmt.Println("Filename:", x.Filename)
	fmt.Println("Args.ID:", x.Args.ID)
	fmt.Println("Args.Num:", x.Args.Num)
	fmt.Println("Args.Rest:", x.Args.Rest)
	return nil

	// Output: Verbosity: [true true]
	// Offset: 5
	// Name: Me
	// Ptr: 3
	// StringSlice: [hello world]
	// PtrSlice: [hello world]
	// IntMap: [a:1 b:5]
	// Filename: hello.go
	// Args.ID: id
	// Args.Num: 10
	// Args.Rest: [remaining1 remaining2]
}

func Example() {
	err := RunContext[cmdRoot](context.Background(),
		// some fake arguments to parse
		"-vv",
		"--offset=5",
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
	)
	if err != nil {
		panic(err)
	}
}
