package main

import (
	"fmt"
)

type AddCommand struct {
	All bool `short:"a" long:"all" description:"Add all files"`
}

func (x *AddCommand) Execute(args []string) error {
	fmt.Printf("Adding (all=%v): %#v\n", x.All, args)
	return nil
}
