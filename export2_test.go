package cli_test

import (
	"errors"
	"strconv"
	"testing"

	"github.com/rprtr258/cli"
)

func map3[A, B, C any](A, B, C) struct{} {
	return struct{}{}
}

// TODO: move similar funcs to batteries and use
func parseInt(s string) (int, error) {
	u, err := strconv.Atoi(s)
	if err != nil {
		return 0, errors.New("invalid digit found in string")
	}
	return u, nil
}

// TODO: move similar funcs to batteries and use
func parseBool(s string) (bool, error) {
	b, err := strconv.ParseBool(s)
	if err != nil {
		return false, err
	}
	return b, nil
}

// Strip a command name if present at the front when used as a `cargo` command
//
// When implementing a cargo subcommand parser needs to be able to skip the first argument which
// is always the same as the executable name without `cargo-` prefix. For example if executable name is
// `cargo-cmd` so first argument would be `cmd`. `cargo_helper` helps to support both invocations:
// with name present when used via cargo and without it when used locally.
//
// You can read the code of this function as this approximate sequence of statements:
// 1. Try to parse a string literal that corresponds to a command name
// 2. It's okay if it's missing
// 3. And don't show anything to the user in `--help` or completion
// 4. Parse this word and then everything else as a tuple, return that second item.
func cargo_helper[T any](cmd string, parser cli.Parser[T]) cli.Parser[T] {
	skip := cli.Optional(cli.Literal(cmd), false).Parser().Hide()
	return cli.Map2(skip, parser, func(_ cli.Option[struct{}], x T) T { return x })
}

func assertSuccess[T any](
	t *testing.T,
	expected T,
	parser *cli.OptionParser[T],
	args ...string,
) {
	cli.AssertSuccess(t, expected, parser, args...)
}
