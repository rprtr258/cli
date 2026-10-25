// `derive_rudo.go` - parser inspired by
// https://github.com/hood/rudo/blob/master/src/main.rs
package examples

import (
	"fmt"
	"os"

	"github.com/rprtr258/cli"
)

type rudoActionKind uint8

const (
	rudoActionAdd rudoActionKind = iota
	rudoActionMark
	rudoActionRead
	rudoActionList
)

type rudoAction struct {
	kind rudoActionKind
	item string // add
	n    uint   // mark, read
}

func DeriveRudo() {
	add := cli.Positional("ITEM", cli.ParseString).
		Map(func(item string) rudoAction { return rudoAction{kind: rudoActionAdd, item: item} }).
		ToOptions().
		Command("add").
		Help("Add a new TODO item")
	mark := cli.Positional("N", cli.ParseUint).
		Map(func(n uint) rudoAction { return rudoAction{kind: rudoActionMark, n: n} }).
		ToOptions().
		Command("mark").
		Help("Mark nth item as done")
	read := cli.Positional("N", cli.ParseUint).
		Map(func(n uint) rudoAction { return rudoAction{kind: rudoActionRead, n: n} }).
		ToOptions().
		Command("read").
		Help("Read nth item")
	// name argument for command is optional
	list := cli.Pure(rudoAction{kind: rudoActionList}).
		ToOptions().
		Command("list").
		Help("Lists everything")

	opts := cli.Choice(add, mark, read, list).
		// fall back to listing when no command is present
		Fallback(rudoAction{kind: rudoActionList}).
		ToOptions().
		Run(os.Args)

	fmt.Println(opts)
}
