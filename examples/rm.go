package main

import (
	"fmt"
)

type RmCommand struct {
	Force bool `short:"f" long:"force" description:"Force removal of files"`
}

func (x *RmCommand) Execute(args []string) error {
	fmt.Printf("Removing (force=%v): %#v\n", x.Force, args)
	return nil
}
