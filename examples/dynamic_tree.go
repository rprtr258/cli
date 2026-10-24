// `dynamic-tree.go` - this example constructs a dynamic command tree from a structure.
package examples

import (
	"fmt"
	"os"

	"github.com/rprtr258/cli"
	"github.com/rprtr258/fun"
)

type cog struct {
	help      string
	name      string
	operation string
	nested    []cog
}

var cogConfig = cog{
	name: "commands",
	help: "contains all the commands, can be flattened with choose but effort :)",
	nested: []cog{
		{help: "First echo command", name: "echoCommand", operation: "echo 'some text'"},
		{help: "Second echo command", name: "anotherEchoCmd", operation: "echo 'another text'"},
		{name: "commandGroup", help: "contains a single sleep", nested: []cog{
			{name: "sleepCommand", help: "sleep for a bit", operation: "sleep 5"},
		}},
	},
}

func cogChoose(parsers []cli.Parser[string]) cli.Parser[string] {
	res := parsers[0]
	for _, next := range parsers[1:] {
		res = cli.Choice(res, next)
	}
	return res
}

func cogMakeParser(item cog) cli.Parser[string] {
	switch len(item.nested) {
	case 0:
		return cli.Pure(item.operation).
			ToOptions().
			Command(item.name).
			Help(item.help).
			Parser()
	default:
		return cogChoose(fun.Map[cli.Parser[string]](cogMakeParser, item.nested...))
	}
}

func DynamicTree() {
	fmt.Println(cogMakeParser(cogConfig).ToOptions().Run(os.Args))
}
