package main

import (
	"fmt"
	"os"

	flags "github.com/rprtr258/cli"
)

type EditorOptions struct {
	Input  flags.Filename `short:"i" long:"input" description:"Input file" default:"-"`
	Output flags.Filename `short:"o" long:"output" description:"Output file" default:"-"`
}

type Point struct {
	X, Y int
}

func (p *Point) UnmarshalFlag(value string) error {
	if _, err := fmt.Sscanf(value, "%d,%d", &p.X, &p.Y); err != nil {
		return fmt.Errorf(`expected two numbers separated by a comma, e.g. "10,20", but got %s`, value)
	}

	return nil
}

func (p Point) MarshalFlag() (string, error) {
	return fmt.Sprintf("%d,%d", p.X, p.Y), nil
}

type AddCommand struct {
	All bool `short:"a" long:"all" description:"Add all files"`
}

func (x *AddCommand) Execute(args []string) error {
	fmt.Printf("Adding (all=%v): %#v\n", x.All, args)
	return nil
}

type RmCommand struct {
	Force bool `short:"f" long:"force" description:"Force removal of files"`
}

func (x *RmCommand) Execute(args []string) error {
	fmt.Printf("Removing (force=%v): %#v\n", x.Force, args)
	return nil
}

type Options struct {
	// Example of verbosity with level
	Verbose []bool `short:"v" long:"verbose" description:"Verbose output"`
	// Example of optional value
	User string `short:"u" long:"user" description:"User name" optional:"yes" optional-value:"pancake"`
	// Example of map with multiple default values
	Users map[string]string `long:"users" description:"User e-mail map" default:"system:system@example.org" default:"admin:admin@example.org"`
	// Example of option group
	EditorOptions
	// Example of custom type Marshal/Unmarshal
	Point Point `long:"point" description:"A x,y point" default:"1,2"`

	Add AddCommand `command:"add" description:"Add a file" long-description:"The add command adds a file to the repository. Use -a to add all files."`
	Rm  RmCommand  `command:"rm" description:"Remove a file" long-description:"The rm command removes a file to the repository. Use -f to force removal of files."`
}

func main() {
	if err := flags.Run[Options](os.Args...); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
}
