package main

import (
	"context"
	"fmt"
	"os"

	"github.com/kr/pretty"

	"github.com/rprtr258/cli"
	"github.com/rprtr258/cli/flags"
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

type cmdAdd struct {
	All  bool     `short:"a" long:"all" description:"Add all files"`
	Args []string `positional-args:"yes"`
}

func (x cmdAdd) Execute(context.Context) error {
	fmt.Printf("Adding (all=%v): %#v\n", x.All, x.Args)
	return nil
}

type cmdRm struct {
	Force bool     `short:"f" long:"force" description:"Force removal of files"`
	Args  []string `positional-args:"yes"`
}

func (x cmdRm) Execute(context.Context) error {
	fmt.Printf("Removing (force=%v): %#v\n", x.Force, x.Args)
	return nil
}

type cmdExample struct {
	Verbose     bool           `short:"v" long:"verbose" description:"Show verbose debug information"`
	Offset      uint           `long:"offset" description:"Offset"`
	Name        string         `short:"n" long:"name" description:"A name" required:"true"`
	Ptr         *int           `short:"p" description:"A pointer to an integer"`
	StringSlice []string       `short:"s" description:"A slice of strings"`
	PtrSlice    []*string      `long:"ptrslice" description:"A slice of pointers to string"`
	IntMap      map[string]int `long:"intmap" description:"A map from string to int"`
	File        string         `short:"f" long:"file" description:"A file" value-name:"FILE"`
	Filename    flags.Filename `long:"filename" description:"A filename"`
	Args        []string       `positional-args:"yes"`
}

func (x cmdExample) Execute(context.Context) error {
	pretty.Println(x)
	return nil
}

type cmdImageList struct {
	Verbose bool     `short:"v" long:"verbose" description:"Verbose output"`
	Output  string   `short:"o" long:"output" description:"Output format" default:"json"`
	Names   []string `positional-args:"yes" description:"Names of images to list"`
}

func (x cmdImageList) Execute(context.Context) error {
	return nil
}

type cmdImage struct {
	List cmdImageList `command:"list" alias:"ls" description:"List images"`
}

type cmdDocker struct {
	Image cmdImage `command:"image" description:"Manage images"`
}

type cmdRoot struct {
	// Example of verbosity with level
	// Verbose []bool `short:"v" long:"verbose" description:"Verbose output"`
	// Example of fun.Optional value
	// User string `short:"u" long:"user" description:"User name" fun.Optional:"yes" fun.Optional-value:"pancake"`
	// Example of map with multiple default values
	// Users map[string]string `long:"users" description:"User e-mail map" default:"system:system@example.org" default:"admin:admin@example.org"`
	// Example of fun.Option group
	// EditorOptions
	// Example of custom type Marshal/Unmarshal
	// Point Point `long:"point" description:"A x,y point" default:"1,2"`

	Add     cmdAdd     `command:"add" description:"Add a file" long-description:"The add command adds a file to the repository. Use -a to add all files."`
	Rm      cmdRm      `command:"rm" description:"Remove a file" long-description:"The rm command removes a file to the repository. Use -f to force removal of files."`
	Example cmdExample `command:"example" description:"Command example with various flag types"`
	Docker  cmdDocker  `command:"docker" description:"Manage Docker"`
}

func main() {
	if err := cli.Run[cmdRoot](context.Background(), os.Args...); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
}
