package main

import (
	"context"
	"fmt"
	"os"

	"github.com/kr/pretty"

	"github.com/rprtr258/cli"
)

type Filename string

type Opts struct {
	Verbose     bool           `short:"v" long:"verbose" description:"Show verbose debug information"`
	Offset      uint           `long:"offset" description:"Offset"`
	Name        string         `short:"n" long:"name" description:"A name" required:"true"`
	Ptr         *int           `short:"p" description:"A pointer to an integer"`
	StringSlice []string       `short:"s" description:"A slice of strings"`
	PtrSlice    []*string      `long:"ptrslice" description:"A slice of pointers to string"`
	IntMap      map[string]int `long:"intmap" description:"A map from string to int"`
	File        string         `short:"f" long:"file" description:"A file" value-name:"FILE"`
	Filename    Filename       `long:"filename" description:"A filename"`
	Args        []string       `positional-args:"yes"`
}

func (x Opts) Execute(context.Context) error {
	pretty.Println(x)
	return nil
}

func main() {
	err := cli.Run[Opts](os.Args...)
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
	}
}
