package cli

import (
	"testing"

	"github.com/kr/pretty"
	"github.com/rprtr258/fun"
	"github.com/stretchr/testify/require"
)

func (self state) peek() (arg, bool) {
	next := self.items_iter()
	_, a, ok := next()
	return a, ok
}

func (parser *OptionParser[T]) RunInner(args argList) (T, parseFailure) {
	return parser.run(args)
}

func (o Opt[T]) Unwrap() T {
	return o.unwrap()
}

func splitOsArgument(input string) (argType, string, Opt[arg], bool) {
	return split_os_argument(input)
}

func UnwrapStdout(err parseFailure) string {
	return unwrap_stdout(err)
}

func UnwrapStderr(err parseFailure) string {
	return unwrap_stderr(err)
}

func PPair[T, R any](a T, b R) fun.Pair[T, R] {
	return fun.Pair[T, R]{a, b}
}

func RequireNoParseFailure(t *testing.T, err parseFailure) {
	t.Helper()

	if err != nil {
		switch self := err.(type) {
		case parseFailureStdout:
			t.Log("Unexpected parseFailureStdout:", pretty.Sprint(self.msg), self.full)
		case parseFailureCompletion:
			t.Log("Unexpected parseFailureCompletion:", string(self))
		case parseFailureStderr:
			t.Log("Unexpected parseFailureStderr:", pretty.Sprint(Doc(self)))
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

	actual, err := parser.RunInner(newArgList(args...))
	require.Nil(t, err)
	require.Equal(t, expected, actual)
}
