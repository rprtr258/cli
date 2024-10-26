// parsing argument value into enum. You can use crate `strum`'s `EnumString` for this purposes as well
package main

import (
	"errors"
	"fmt"

	"github.com/rprtr258/cli"
)

type Baz uint8

const (
	BazFoo Baz = iota
	BazBar
	BazFooBar
)

func parseBaz(s string) (Baz, error) {
	switch s {
	case "foo":
		return BazFoo, nil
	case "bar":
		return BazBar, nil
	case "foobar":
		return BazFooBar, nil
	default:
		return *new(Baz), errors.New("Expected foo|bar|foobar")
	}
}

func main() {
	opt := cli.Argument(cli.Long("baz").
		Short('b').
		Help("choose between foo, bar or foobar"), "CMD", parseBaz).
		Parser().
		ToOptions().
		Run()

	fmt.Println(opt)
}
