// You can open files as part of parsing process too, might not be the best idea though
// because depending on a context `bpaf` might need to evaluate some parsers multiple times.
//
// Main motivation for this example is that you can open a file as part of the argument parsing
// and give a reader directly to user. In practice to replicate `cat`'s behavior you'd accept
// multiple files with `many` and open them one by one in your code.
package main

import (
	"bufio"
	"fmt"
	"log"
	"os"

	"github.com/rprtr258/cli"
)

func main() {
	// TODO: get back optional with stdin as default
	file := cli.Parse(
		cli.Positional("FILE", func(s string) (string, error) { return s, nil }).
			Help("File name to concatenate, with no FILE or when FILE is -, read standard input").
			Parser(),
		func(path string) cli.Result[*os.File] {
			f, err := os.Open(string(path))
			if err != nil {
				return cli.Err[*os.File](cli.Error(cli.MessageParseFail(err.Error())))
			}

			return cli.Ok(f)
		}).
		Parser().
		ToOptions().
		Descr("Concatenate a file to standard output").
		Run()
	defer file.Close()

	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanLines)
	for scanner.Scan() {
		line := scanner.Text()
		fmt.Println(line)
	}
	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
}
