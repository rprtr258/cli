package cli

import (
	"errors"
	"strconv"
	"testing"

	"github.com/stretchr/testify/require"
)

// RunE must not print or exit, so it has to be usable from external packages.
func TestRunEReturnsParsedValue(t *testing.T) {
	parser := Long("verbose").Switch().ToOptions()

	verbose, err := parser.RunE([]string{"app", "--verbose"})

	require.NoError(t, err)
	require.True(t, verbose)
}

func TestRunEUsesDefaults(t *testing.T) {
	parser := Long("verbose").Switch().ToOptions()

	verbose, err := parser.RunE([]string{"app"})

	require.NoError(t, err)
	require.False(t, verbose)
}

func TestRunEParsesArgument(t *testing.T) {
	parser := Long("number").Argument("N", strconv.Atoi).ToOptions()

	number, err := parser.RunE([]string{"app", "--number", "42"})

	require.NoError(t, err)
	require.Equal(t, 42, number)
}

func TestRunEErrorInsteadOfExit(t *testing.T) {
	parser := Long("number").Argument("N", strconv.Atoi).ToOptions()

	// If RunE behaved like Run and called os.Exit, this test would terminate here.
	_, err := parser.RunE([]string{"app", "--number", "not-a-number"})

	require.Error(t, err)
	failure, ok := errors.AsType[FailureError](err)
	require.True(t, ok)
	require.True(t, failure.IsError())
	require.Equal(t, 1, failure.ExitCode())
	require.Contains(t, failure.Error(), "couldn't parse `not-a-number`")
}

func TestRunEHelpIsNotAnErrorDiagnostic(t *testing.T) {
	parser := Long("number").Argument("N", strconv.Atoi).ToOptions()

	_, err := parser.RunE([]string{"app", "--help"})

	failure, ok := errors.AsType[FailureError](err)
	require.True(t, ok)
	require.False(t, failure.IsError())
	require.Equal(t, 0, failure.ExitCode())
	require.Contains(t, failure.Error(), "Usage")
}

// The program name passed as args[0] must show up in the usage line.
func TestRunEUsageContainsProgramName(t *testing.T) {
	parser := Long("number").Argument("N", strconv.Atoi).ToOptions()

	_, err := parser.RunE([]string{"app", "--help"})

	failure, ok := errors.AsType[FailureError](err)
	require.True(t, ok)
	require.Contains(t, failure.Error(), "Usage: app --number=N")
}
