// Consume a named argument with fallback to environment variable
package examples

import (
	"fmt"
	"os"

	"github.com/rprtr258/cli"
)

func EnvVariable() {
	type Opts struct {
		key string
	}

	opts := cli.Long("key").
		Env("ACCESS_KEY").
		Help("access key to use").
		Argument("KEY", func(s string) (string, error) { return s, nil }).
		Map(func(key string) Opts { return Opts{key} }).
		ToOptions().
		Run(os.Args)

	fmt.Println(opts)
}
