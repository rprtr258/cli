package cli

import (
	"fmt"
	"strings"

	"github.com/rprtr258/cli/internal"
	"github.com/rprtr258/fun"
)

// Completion variant, namely completion and description
type Completion struct {
	Item        string             // Item is completion
	Description fun.Option[string] // Description of the completion variant
}

func (c *command) complete(args []string) []Completion {
	return nil
}

func (p *command) printCompletions(args ...string) {
	items := p.complete(args)
	maxl := internal.MaxBy(func(v Completion) int {
		return len(v.Item)
	}, items...)

	for _, v := range items {
		fmt.Printf(v.Item)
		// TODO: problems with colored description
		// TODO: description get into completion wtf
		if description, ok := v.Description.Unpack(); ok {
			fmt.Printf("%s  # %s", strings.Repeat(" ", maxl-len(v.Item)), description)
		}
		fmt.Println()
	}
}
