package cli_test

import (
	"testing"

	"github.com/rprtr258/fun"

	"github.com/rprtr258/cli"
)

func newArgs(args ...string) cli.Args {
	return cli.NewArgs(args...)
}

func split_os_argument(input string) (cli.ArgType, string, cli.Option[cli.Arg], bool) {
	return cli.SplitOsArgument(input)
}

func unwrap_stdout(err cli.ParseFailure) string {
	return cli.UnwrapStdout(err)
}

func unwrap_stderr(err cli.ParseFailure) string {
	return cli.UnwrapStderr(err)
}

func pair[T, R any](a T, b R) fun.Pair[T, R] {
	return cli.PPair(a, b)
}

func requireNoParseFailure(t *testing.T, err cli.ParseFailure) {
	cli.RequireNoParseFailure(t, err)
}

func parseInt32(s string) (int32, error) {
	return cli.ParseInt32(s)
}

func parseUint32(s string) (uint32, error) {
	return cli.ParseUint32(s)
}

func parseUint(s string) (uint, error) {
	return cli.ParseUint(s)
}

func parseString(s string) (string, error) {
	return cli.ParseString(s)
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
