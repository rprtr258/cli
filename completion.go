package cli

import (
	"fmt"
	"strings"

	"github.com/rprtr258/cli/internal"
)

// Completion variant, namely completion and description
type Completion struct {
	Item        string // Item is completion
	Description string // Description of the completion variant
}

func printCompletions(items []Completion) {
	maxl := internal.MaxBy(func(v Completion) int {
		return len(v.Item)
	}, items...)

	for _, v := range items {
		// TODO: problems with colored description
		// TODO: description get into completion wtf
		fmt.Printf("%s%s  # %s\n", v.Item, strings.Repeat(" ", maxl-len(v.Item)), v.Description)
	}
}
