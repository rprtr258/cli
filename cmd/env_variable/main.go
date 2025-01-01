// Consume a named argument with fallback to environment variable
package main

import (
	"fmt"
	"os"

	"github.com/rprtr258/cli"
)

type Opts struct {
	key string
}

func main() {
	key := cli.Argument[string](cli.Long("key").
		Env("ACCESS_KEY").
		Help("access key to use"),
		"KEY", func(s string) (string, error) {
			return s, nil
		}).Parser()

	opts := cli.Map(key, func(key string) Opts { return Opts{key} }).ToOptions().Run(os.Args)

	fmt.Println(opts)
}
