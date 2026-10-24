// `derive_rudo.rs` - parser inspired by
// https://github.com/hood/rudo/blob/e448942b752c56dd2be2e2bb5026ced45e215ed5/src/main.rs
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

func rudoCommand(p cli.Parser[rudoAction], name, help string) cli.Parser[rudoAction] {
	return p.
		Parser().
		ToOptions().
		Command(name).
		Help(help).
		Parser()
}

func DeriveRudo() {
	add := rudoCommand(
		cli.Positional("ITEM", cli.ParseString).
			Map(func(item string) rudoAction { return rudoAction{kind: rudoActionAdd, item: item} }),
		"add", "Add a new TODO item",
	)
	mark := rudoCommand(
		cli.Positional("N", cli.ParseUint).
			Map(func(n uint) rudoAction { return rudoAction{kind: rudoActionMark, n: n} }),
		"mark", "Mark nth item as done",
	)
	read := rudoCommand(
		cli.Positional("N", cli.ParseUint).
			Map(func(n uint) rudoAction { return rudoAction{kind: rudoActionRead, n: n} }),
		"read", "Read nth item",
	)
	// name argument for command is optional
	list := rudoCommand(
		cli.Pure(rudoAction{kind: rudoActionList}),
		"list", "Lists everything",
	)

	opts := cli.Choice(add, mark, read, list).
		// fall back to listing when no command is present
		Fallback(rudoAction{kind: rudoActionList}).
		ToOptions().
		Run(os.Args)

	fmt.Println(opts)
}
