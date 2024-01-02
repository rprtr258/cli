package main

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"

	flags "github.com/rprtr258/cli/contrib"
)

type EditorOptions struct {
	Input  flags.Filename `short:"i" long:"input" description:"Input file" default:"-"`
	Output flags.Filename `short:"o" long:"output" description:"Output file" default:"-"`
}

type Point struct {
	X, Y int
}

func (p *Point) UnmarshalFlag(value string) error {
	parts := strings.Split(value, ",")
	if len(parts) != 2 {
		return errors.New("expected two numbers separated by a ,")
	}

	x, err := strconv.ParseInt(parts[0], 10, 32)
	if err != nil {
		return err
	}

	y, err := strconv.ParseInt(parts[1], 10, 32)
	if err != nil {
		return err
	}

	p.X = int(x)
	p.Y = int(y)

	return nil
}

func (p Point) MarshalFlag() (string, error) {
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
	EditorOptions
	// Example of custom type Marshal/Unmarshal
	Point Point `long:"point" description:"A x,y point" default:"1,2"`

	Add AddCommand `command:"add" description:"Add a file" long-description:"The add command adds a file to the repository. Use -a to add all files."`
	Rm  RmCommand  `command:"rm" description:"Remove a file" long-description:"The rm command removes a file to the repository. Use -f to force removal of files."`
}

func main() {
	parser := flags.NewParser(&Options{})
	if err := parser.Run(os.Args...); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
}
