// You can open files as part of parsing process too, might not be the best idea though
// because depending on a context the library might need to evaluate some parsers multiple times.
//
// Main motivation for this example is that you can open a file as part of the argument parsing
// and give a reader directly to user. In practice to replicate `cat`'s behavior you'd accept
// multiple files with `many` and open them one by one in your code.
package examples

import (
	"bufio"
	"fmt"
	"log"
	"os"

	"github.com/rprtr258/cli"
)

func Cat() {
	run := func() error {
		file := cli.Parse(
			cli.Optional(
				cli.Positional("FILE", func(s string) (string, error) { return s, nil }).
					Help("File name to concatenate, with no FILE or when FILE is -, read standard input").
					Parser()).Parser(),
			func(pathOpt cli.Opt[string]) (*os.File, error) {
				path, ok := pathOpt.Unpack()
				if !ok {
					return os.Stdin, nil
				}
				f, err := os.Open(string(path))
				if err != nil {
					return nil, err
				}

				return f, nil
			}).
			Parser().
			ToOptions().
			Description("Concatenate a file to standard output").
			Run(os.Args)
		defer file.Close()

		scanner := bufio.NewScanner(file)
		scanner.Split(bufio.ScanLines)
		for scanner.Scan() {
			line := scanner.Text()
			fmt.Println(line)
		}
		return scanner.Err()
	}

	if err := run(); err != nil {
		log.Fatal(err)
	}
}
