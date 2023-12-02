package main

import (
	"fmt"
	"os"

	"github.com/davecgh/go-spew/spew"

	"github.com/rprtr258/cli"
)

type Filename string

func main() {
	type Opts struct {
		Verbose     bool           `short:"v" long:"verbose" description:"Show verbose debug information"`
		Offset      uint           `long:"offset" description:"Offset"`
		Name        string         `short:"n" long:"name" description:"A name" required:"true"`
		File        string         `short:"f" long:"file" description:"A file" value-name:"FILE"`
		Ptr         *int           `short:"p" description:"A pointer to an integer"`
		StringSlice []string       `short:"s" description:"A slice of strings"`
		PtrSlice    []*string      `long:"ptrslice" description:"A slice of pointers to string"`
		IntMap      map[string]int `long:"intmap" description:"A map from string to int"`
		Filename    Filename       `long:"filename" description:"A filename"`
		Args        []string       `positional`
	}
	got, err := cli.Parse[Opts](os.Args...)
	spew.Dump("got", got)
	if err != nil {
		fmt.Println("err:", err.Error())
	}
}
