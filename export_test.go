package cli

import (
	"testing"

	"github.com/kr/pretty"
	"github.com/rprtr258/fun"
	"github.com/stretchr/testify/require"
)

func (self State) peek() (Arg, bool) {
	next := self.items_iter()
	_, a, ok := next()
	return a, ok
}

func (parser *OptionParser[T]) RunInner(args Args) (T, ParseFailure) {
	return parser.run(args)
}

func (o Option[T]) Unwrap() T {
	return o.unwrap()
}

func SplitOsArgument(input string) (ArgType, string, Option[Arg], bool) {
	return split_os_argument(input)
}

func UnwrapStdout(err ParseFailure) string {
	return unwrap_stdout(err)
}

func UnwrapStderr(err ParseFailure) string {
	return unwrap_stderr(err)
}

func PPair[T, R any](a T, b R) fun.Pair[T, R] {
	return fun.Pair[T, R]{a, b}
}

func RequireNoParseFailure(t *testing.T, err ParseFailure) {
	t.Helper()

	if err != nil {
		switch self := err.(type) {
		case ParseFailureStdout:
			t.Log("Unexpected ParseFailureStdout:", pretty.Sprint(self.msg), self.full)
		case ParseFailureCompletion:
			t.Log("Unexpected ParseFailureCompletion:", string(self))
		case ParseFailureStderr:
			t.Log("Unexpected ParseFailureStderr:", pretty.Sprint(Doc(self)))
		}
		t.FailNow()
	}
}

func AssertSuccess[T any](
	t *testing.T,
	expected T,
	parser *OptionParser[T],
	args ...string,
) {
	t.Helper()

	actual, err := parser.RunInner(NewArgs(args...))
	require.Nil(t, err)
	require.Equal(t, expected, actual)
}
