// `--help` output customizations
//
// help header, help footer, a short description and a custom usage line
package main

import (
	"fmt"
	"os"

	"github.com/rprtr258/cli"
)

func main() {
	opt := cli.Short('d').
		Help("Release the dragon").
		Switch().Parser().
		ToOptions().
		Descr("I am a program and I do things").
		Header("Sometimes they even work.").
		Footer("Beware `-d`, dragons be here").
		WithUsage(func(doc cli.Doc) cli.Doc {
			u := cli.Doc{}
			u.Emphasis("You can call it with following flags:")
			u.Text(" ")
			u.Doc(doc)
			return u
		}).
		Run(os.Args)

	fmt.Println(opt)
}
