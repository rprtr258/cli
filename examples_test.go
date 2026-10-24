package cli

import (
	"errors"
	"strconv"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSplitExampleArgs(t *testing.T) {
	for name, test := range map[string]struct {
		command  string
		expected []string
	}{
		"plain":          {"src.txt dst.txt", []string{"src.txt", "dst.txt"}},
		"extra spaces":   {"  a   b  ", []string{"a", "b"}},
		"single quotes":  {"--name 'my file.txt'", []string{"--name", "my file.txt"}},
		"double quotes":  {`--name "my file.txt"`, []string{"--name", "my file.txt"}},
		"glued quotes":   {"a'b c'd", []string{"ab cd"}},
		"newlines/tabs":  {"a\nb\tc", []string{"a", "b", "c"}},
		"empty":          {"", []string{}},
		"only spaces":    {"   ", []string{}},
		"empty argument": {"''", []string{""}},
	} {
		t.Run(name, func(t *testing.T) {
			words, err := split_example_args(test.command)

			require.NoError(t, err)
			require.Equal(t, test.expected, words)
		})
	}
}

func TestSplitExampleArgsUnbalancedQuotes(t *testing.T) {
	for _, command := range []string{"'", `"`, "a 'b", `a "b`} {
		_, err := split_example_args(command)

		require.Error(t, err)
		require.Contains(t, err.Error(), "unbalanced")
	}
}

func TestExamplesRenderedInHelp(t *testing.T) {
	parser := Positional[string]("FILE", ParseString).
		Help("File to copy").
		ToOptions().
		Description("Copy a file").
		Example("Copy the file to the default destination", "src.txt").
		Example("Copy a file with spaces in the name", "'my file.txt'")

	_, err := parser.RunE([]string{"prog", "--help"})

	failure, ok := errors.AsType[FailureError](err)
	require.True(t, ok)
	require.Contains(t, failure.Error(), "Examples:")
	require.Contains(t, failure.Error(), "Copy the file to the default destination")
	require.Contains(t, failure.Error(), "$ prog src.txt")
	require.Contains(t, failure.Error(), "Copy a file with spaces in the name")
	require.Contains(t, failure.Error(), "$ prog 'my file.txt'")
}

func TestExamplesNotRenderedInHelpWhenEmpty(t *testing.T) {
	parser := Positional[string]("FILE", ParseString).ToOptions()

	_, err := parser.RunE([]string{"prog", "--help"})

	failure, ok := errors.AsType[FailureError](err)
	require.True(t, ok)
	require.NotContains(t, failure.Error(), "Examples:")
}

func TestExamplesRenderedInManpage(t *testing.T) {
	parser := Positional[string]("FILE", ParseString).
		ToOptions().
		Example("Copy the file to the default destination", "src.txt")

	manpage := parser.RenderManpage("prog", Section{Kind: SectionGeneral}, Invalid[string](), Invalid[string](), Invalid[string]())

	require.Contains(t, manpage, ".SH EXAMPLES")
	require.Contains(t, manpage, ".SS EXAMPLES:")
	require.Contains(t, manpage, "Copy the file to the default destination")
	require.Contains(t, manpage, "$ prog")
	require.Contains(t, manpage, "src.txt")
}

func TestExamplesNotRenderedInManpageWhenEmpty(t *testing.T) {
	parser := Positional[string]("FILE", ParseString).ToOptions()

	manpage := parser.RenderManpage("prog", Section{Kind: SectionGeneral}, Invalid[string](), Invalid[string](), Invalid[string]())

	require.NotContains(t, manpage, "EXAMPLES")
}

// Each example must parse with the parser it is declared on.
func TestExamplesValid(t *testing.T) {
	parser := Short('n').Long("number").Argument("N", strconv.Atoi).ToOptions().
		Example("Parse a number", "--number 42").
		Example("Two examples can be declared at once", "--number=42")

	require.NotPanics(t, func() {
		_, err := parser.RunE([]string{"prog", "--number", "1"})
		require.NoError(t, err)
	})
}

func TestExamplesInvalidPanicsOnRun(t *testing.T) {
	parser := Short('n').Long("number").Argument("N", strconv.Atoi).ToOptions().
		Example("Parse a number", "--number 42").
		// --back is not a flag of this parser
		Example("Pass a backup flag", "--back")

	require.Panics(t, func() { parser.RunE([]string{"prog", "--number", "1"}) })
	assertPanicsContains(t, func() { parser.RunE([]string{"prog", "--number", "1"}) },
		"cli usage BUG: example \"--back\" doesn't parse")
}

func TestExamplesInvalidPanicsOnCheckInvariants(t *testing.T) {
	parser := Positional[string]("FILE", ParseString).ToOptions().
		Example(
			"Copy the file",
			// two positionals given, one expected
			"src.txt dst.txt",
		)

	require.Panics(t, func() { parser.CheckInvariants() })
}

func TestExamplesOnCommandValidated(t *testing.T) {
	show := Positional[string]("FILE", ParseString).Help("File to show").
		Parser().ToOptions().
		Example(
			"Show a file",
			"src.txt",
		)
	parser := show.Command("show").ToOptions()

	// The command name is prepended to the example before parsing.
	require.NotPanics(t, func() {
		_, err := parser.RunE([]string{"prog", "show", "src.txt"})
		require.NoError(t, err)
	})
}

func TestExamplesOnCommandInvalidPanics(t *testing.T) {
	show := Positional[string]("FILE", ParseString).Help("File to show").
		Parser().ToOptions().
		Example(
			"Show a file",
			// the positional belongs to the `show` command, not to the root
			"src.txt show",
		)
	parser := show.Command("show").ToOptions()

	require.Panics(t, func() { parser.RunE([]string{"prog"}) })
}

func TestExamplesRenderedInSubcommandHelp(t *testing.T) {
	show := Positional[string]("FILE", ParseString).Help("File to show").
		Parser().ToOptions().
		Example("Show a file", "src.txt")
	parser := show.Command("show").ToOptions()

	_, err := parser.RunE([]string{"prog", "show", "--help"})

	failure, ok := errors.AsType[FailureError](err)
	require.True(t, ok)
	require.Contains(t, failure.Error(), "$ prog show src.txt")
}

func TestExampleRenderedInHelp(t *testing.T) {
	parser := Positional[string]("FILE", ParseString).
		ToOptions().
		Example("Copy the file to the default destination", "src.txt")

	_, err := parser.RunE([]string{"prog", "--help"})

	failure, ok := errors.AsType[FailureError](err)
	require.True(t, ok)
	require.Contains(t, failure.Error(), "Copy the file to the default destination")
	require.Contains(t, failure.Error(), "$ prog src.txt")
}

func TestExampleInvalidPanicsOnRun(t *testing.T) {
	parser := Positional[string]("FILE", ParseString).
		ToOptions().
		// two positionals given, one expected
		Example("Copy the file", "src.txt dst.txt")

	assertPanicsContains(t, func() { parser.RunE([]string{"prog", "src.txt"}) },
		"cli usage BUG: example \"src.txt dst.txt\" doesn't parse")
}

func TestExampleDeclarationPanics(t *testing.T) {
	for name, args := range map[string][2]string{
		"empty description": {"", "cmd"},
		"empty command":     {"descr", ""},
		"unbalanced quotes": {"descr", "'unclosed"},
	} {
		t.Run(name, func(t *testing.T) {
			parser := Positional[string]("FILE", ParseString).ToOptions()

			require.Panics(t, func() { parser.Example(args[0], args[1]) })
		})
	}
}

func TestExamplesDeclarationPanics(t *testing.T) {
	for name, examples := range map[string][2]string{
		"odd count":         {"descr"},
		"empty description": {"", "cmd"},
		"empty command":     {"descr", ""},
		"unbalanced quotes": {"descr", "'unclosed"},
	} {
		t.Run(name, func(t *testing.T) {
			parser := Positional[string]("FILE", ParseString).ToOptions()
			require.Panics(t, func() { parser.Example(examples[0], examples[1]) })
		})
	}
}

func TestExamplesAppend(t *testing.T) {
	parser := Positional[string]("FILE", ParseString).ToOptions().
		Example("First", "a.txt").
		Example("Second", "b.txt")

	_, err := parser.RunE([]string{"prog", "--help"})

	failure, ok := errors.AsType[FailureError](err)
	require.True(t, ok)
	require.Contains(t, failure.Error(), "$ prog a.txt")
	require.Contains(t, failure.Error(), "$ prog b.txt")
}
