// `env_logger.go` - convert verbosity level from a count of -v flags into an enum
// from a logger crate of your choice.
package examples

import (
	"fmt"
	"os"

	"github.com/rprtr258/cli"
)

// generally you'd use this from the log crate itself
type levelFilter uint8

const (
	levelOff levelFilter = iota
	levelError
	levelWarn
	levelInfo
	levelDebug
	levelTrace
)

type envLoggerOptions struct {
	verbose levelFilter
	potato  uint
}

func EnvLogger() {
	verbose := cli.Short('v').
		Help("Verbosity level, use multiple times for more verbosity").
		ReqFlag(struct{}{}).
		Count().
		Map(func(l int) levelFilter {
			levels := []levelFilter{levelOff, levelError, levelWarn, levelInfo, levelDebug, levelTrace}
			if l > len(levels)-1 {
				l = len(levels) - 1
			}
			return levels[l]
		})

	potato := cli.Long("potato").
		Help("number of potatoes").
		Argument("NUM", cli.ParseUint).
		Fallback(3)

	opts := cli.Map2(verbose, potato, func(verbose levelFilter, potato uint) envLoggerOptions {
		return envLoggerOptions{verbose: verbose, potato: potato}
	}).ToOptions()

	fmt.Println(opts.Run(os.Args))
}
