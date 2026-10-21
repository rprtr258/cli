package cli_test

import (
	"strconv"
	"testing"

	"github.com/rprtr258/cli"
	"github.com/stretchr/testify/require"
)

// RunE must not print or exit, so it has to be usable from external packages.
func TestRunEReturnsParsedValue(t *testing.T) {
	parser := cli.Long("verbose").Switch().Parser().ToOptions()

	verbose, err := parser.RunE([]string{"app", "--verbose"})

	require.NoError(t, err)
	require.True(t, verbose)
}

func TestRunEUsesDefaults(t *testing.T) {
	parser := cli.Long("verbose").Switch().Parser().ToOptions()

	verbose, err := parser.RunE([]string{"app"})

	require.NoError(t, err)
	require.False(t, verbose)
}

func TestRunEParsesArgument(t *testing.T) {
	parser := cli.Argument(cli.Long("number"), "N", strconv.Atoi).Parser().ToOptions()

	number, err := parser.RunE([]string{"app", "--number", "42"})

	require.NoError(t, err)
	require.Equal(t, 42, number)
}

func TestRunEErrorInsteadOfExit(t *testing.T) {
	parser := cli.Argument(cli.Long("number"), "N", strconv.Atoi).Parser().ToOptions()

	// If RunE behaved like Run and called os.Exit, this test would terminate here.
	_, err := parser.RunE([]string{"app", "--number", "not-a-number"})

	require.Error(t, err)
	var failure cli.FailureError
	require.ErrorAs(t, err, &failure)
	require.True(t, failure.IsError())
	require.Equal(t, 1, failure.ExitCode())
	require.Contains(t, failure.Error(), "couldn't parse `not-a-number`")
}

func TestRunEHelpIsNotAnErrorDiagnostic(t *testing.T) {
	parser := cli.Argument(cli.Long("number"), "N", strconv.Atoi).Parser().ToOptions()

	_, err := parser.RunE([]string{"app", "--help"})

	var failure cli.FailureError
	require.ErrorAs(t, err, &failure)
	require.False(t, failure.IsError())
	require.Equal(t, 0, failure.ExitCode())
	require.Contains(t, failure.Error(), "Usage")
}
