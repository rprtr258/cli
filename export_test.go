package cli

import (
	"errors"
	"strconv"
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

// Check the invariants `bpaf` relies on for normal operations
//
// Takes a parameter whether to check for cosmetic invariants or not
// (max help width exceeding 120 symbols, etc), currently not in use
//
// Best used as part of your test suite:
// ```rust
// #[test]
// func check_options() {
// # fun.Options := || short('p').Switch().to_options();
//
//	    fun.Options().check_invariants(false)
//	}
//
// ```
//
// # Panics
//
// `check_invariants` indicates problems with panic
func (self OptionParser[T]) CheckInvariants(bool) {
	positional_invariant_check(self.inner.meta, true)
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

func MMap3[A, B, C any](_ A, _ B, _ C) struct{} {
	return struct{}{}
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

// TODO: move similar funcs to batteries and use
func ParseUint32(s string) (uint32, error) {
	u, err := strconv.ParseUint(s, 10, 32)
	if err != nil {
		// TODO: %q
		return 0, errors.New("invalid digit found in string")
	}
	return uint32(u), nil
}

// TODO: move similar funcs to batteries and use
func ParseInt32(s string) (int32, error) {
	u, err := strconv.ParseInt(s, 10, 32)
	if err != nil {
		return 0, errors.New("invalid digit found in string")
	}
	return int32(u), nil
}

// TODO: move similar funcs to batteries and use
func ParseUint(s string) (uint, error) {
	u, err := strconv.ParseUint(s, 10, 32)
	if err != nil {
		return 0, errors.New("invalid digit found in string")
	}
	return uint(u), nil
}

// TODO: move similar funcs to batteries and use
func ParseInt(s string) (int, error) {
	u, err := strconv.Atoi(s)
	if err != nil {
		return 0, errors.New("invalid digit found in string")
	}
	return u, nil
}

// TODO: move similar funcs to batteries and use
func ParseString(s string) (string, error) {
	return s, nil
}

// TODO: move similar funcs to batteries and use
func ParseBool(s string) (bool, error) {
	b, err := strconv.ParseBool(s)
	if err != nil {
		return false, err
	}
	return b, nil
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
