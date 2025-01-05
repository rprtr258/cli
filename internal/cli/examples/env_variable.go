// Consume a named argument with fallback to environment variable
package examples

import (
	"fmt"
	"os"

	"github.com/rprtr258/cli/internal/cli"
)

func EnvVariable() {
	type Opts struct {
		key string
	}

	key := cli.Argument[string](cli.Long("key").
		Env("ACCESS_KEY").
		Help("access key to use"),
		"KEY", func(s string) (string, error) {
			return s, nil
		}).Parser()

	opts := cli.Map(key, func(key string) Opts { return Opts{key} }).ToOptions().Run(os.Args)

	fmt.Println(opts)
}
