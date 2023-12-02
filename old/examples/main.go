package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"strings"

	cli "github.com/rprtr258/cli/old"
)

type CommandAdd struct {
	All bool `short:"a" long:"all" description:"Add all files"`
}

func (x *CommandAdd) Execute(args []string) error {
	fmt.Printf("Adding (all=%v): %#v\n", x.All, args)
	return nil
}

type CommandRm struct {
	Force bool `short:"f" long:"force" description:"Force removal of files"`
}

func (x *CommandRm) Execute(args []string) error {
	fmt.Printf("Removing (force=%v): %#v\n", x.Force, args)
	return nil
}

type EditorOptions struct {
	Input  cli.Filename `short:"i" long:"input" description:"Input file" default:"-"`
	Output cli.Filename `short:"o" long:"output" description:"Output file" default:"-"`
}

type FlagPoint struct {
	X, Y int
}

func (p *FlagPoint) UnmarshalFlag(value string) error {
	if parts := strings.Split(value, ","); len(parts) != 2 {
		return errors.New("expected two numbers separated by a ,")
	}

	if _, err := fmt.Sscanf(value, "%d,%d", &p.X, &p.Y); err != nil {
		return err
	}

	return nil
}

func (p FlagPoint) MarshalFlag() (string, error) {
	return fmt.Sprintf("%d,%d", p.X, p.Y), nil
}

type Options struct {
	// Example of verbosity with level
	Verbose []bool `short:"v" long:"verbose" description:"Verbose output"`
	// Example of optional value
	User string `short:"u" long:"user" description:"User name" optional:"yes" optional-value:"pancake"`
	// Example of map with multiple default values
	Users map[string]string `long:"users" description:"User e-mail map" default:"system:system@example.org" default:"admin:admin@example.org"`
	// Example of option group
	Editor EditorOptions `group:"Editor Options"`
	// Example of custom type Marshal/Unmarshal
	Point FlagPoint `long:"point" description:"A x,y point" default:"1,2"`
}

func main() {
	var opts Options
	var parser = cli.NewParser(&opts, cli.Default)
	parser.AddCommand(
		"add",
		"Add a file",
		"The add command adds a file to the repository. Use -a to add all files.",
		&CommandAdd{},
	)
	parser.AddCommand(
		"rm",
		"Remove a file",
		"The rm command removes a file to the repository. Use -f to force removal of files.",
		&CommandRm{},
	)

	if _, err := parser.ParseArgs(os.Args[1:]); err != nil {
		if flagsErr, ok := err.(cli.ErrorKind); ok && flagsErr != cli.ErrHelp {
			log.Fatal(flagsErr.Error())
		}
		return
	}

	fmt.Print(opts)
}
