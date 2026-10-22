package cli

import (
	"errors"
	"iter"
	"strconv"
	"strings"
	"testing"

	"github.com/rprtr258/fun"
	funiter "github.com/rprtr258/fun/iter"
	"github.com/stretchr/testify/require"
)

func map3[A, B, C any](A, B, C) struct{} {
	return struct{}{}
}

// TODO: move similar funcs to batteries and use
func parseUint32(s string) (uint32, error) {
	u, err := strconv.ParseUint(s, 10, 32)
	if err != nil {
		// TODO: %q
		return 0, errors.New("invalid digit found in string")
	}
	return uint32(u), nil
}

// TODO: move similar funcs to batteries and use
func parseInt32(s string) (int32, error) {
	u, err := strconv.ParseInt(s, 10, 32)
	if err != nil {
		return 0, errors.New("invalid digit found in string")
	}
	return int32(u), nil
}

// TODO: move similar funcs to batteries and use
func parseUint(s string) (uint, error) {
	u, err := strconv.ParseUint(s, 10, 32)
	if err != nil {
		return 0, errors.New("invalid digit found in string")
	}
	return uint(u), nil
}

// TODO: move similar funcs to batteries and use
func parseString(s string) (string, error) {
	return s, nil
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
func cargo_helper[T any](cmd string, parser parser[T]) Parser[T] {
	skip := Literal(cmd).Optional().Hide()
	return Map2(skip, parser, func(_ Opt[struct{}], x T) T { return x })
}

// Adapt a sequence of pairs into a two-value sequence.
func seq2[K, V any](seq funiter.Seq[fun.Pair[K, V]]) iter.Seq2[K, V] {
	return func(yield func(K, V) bool) {
		for v := range seq {
			if !yield(v.K, v.V) {
				return
			}
		}
	}
}

func assertSuccess[T any](t *testing.T, expected T, parser *OptionParser[T], args ...string) {
	t.Helper()

	AssertSuccess(t, expected, parser, args...)
}

// Parser must fail with the given stderr message for the given args.
func assertFailure[T any](t *testing.T, parser *OptionParser[T], expected string, args ...string) {
	t.Helper()

	_, err := parser.RunInner(newArgList(args...))
	require.Equal(t, expected, UnwrapStderr(err))
}

// Parser must print the expected stdout for the given args, `--help` by default.
func assertHelp[T any](t *testing.T, parser *OptionParser[T], expected string, args ...string) {
	t.Helper()

	if len(args) == 0 {
		args = []string{"--help"}
	}
	_, err := parser.RunInner(newArgList(args...))
	require.Equal(t, expected, UnwrapStdout(err))
}

// Parser must produce the expected completion output for the given args.
func assertCompletion[T any](t *testing.T, parser *OptionParser[T], expected string, args ...string) {
	t.Helper()

	_, err := parser.RunInner(newArgList(args...).setComp(0))
	require.Equal(t, expected, UnwrapStdout(err))
}

// first line of `--help` output without the `Usage: ` prefix must be the expected usage.
func assert_usage[T any](t *testing.T, parser *OptionParser[T], expected string) {
	t.Helper()

	_, output := parser.RunInner(newArgList("--help"))
	usage := strings.TrimPrefix(strings.Split(UnwrapStdout(output), "\n")[0], "Usage: ")
	require.Equal(t, expected, usage)
}

// Catch a panic raised inside f and return its value, fail the test if f returns normally.
func catchPanic[T any](t *testing.T, f func()) (res T) {
	t.Helper()

	defer func() {
		if r := recover(); r != nil {
			res = r.(T)
		} else {
			t.FailNow()
		}
	}()
	f()
	return
}
