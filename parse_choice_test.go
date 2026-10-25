package cli

import (
	"testing"
)

func TestChoiceAccept(t *testing.T) {
	parser := Long("animal").Argument("ANIMAL", ParseString).Choice("dog", "cat").ToOptions()

	assertSuccess(t, "dog", parser, "--animal dog")
	assertSuccess(t, "cat", parser, "--animal cat")
}

func TestChoiceReject(t *testing.T) {
	parser := Long("animal").Argument("ANIMAL", ParseString).Choice("dog", "cat").ToOptions()

	assertFailure(t, parser, "couldn't parse `zebra`: expected one of `cat`, `dog`", "--animal zebra")
}

func TestChoiceRepeated(t *testing.T) {
	// Choice calls accumulate and the value list stays sorted
	parser := Long("animal").Argument("ANIMAL", ParseString).Choice("dog").Choice("ant").Choice("cat").
		ToOptions()

	assertSuccess(t, "ant", parser, "--animal ant")
	assertFailure(t, parser, "couldn't parse `yak`: expected one of `ant`, `cat`, `dog`", "--animal yak")
}

func TestChoiceHelp(t *testing.T) {
	parser := Long("animal").Help("A pet").Argument("ANIMAL", ParseString).Choice("dog", "cat").ToOptions()

	assertHelp(t, parser, `Usage: --animal=ANIMAL

Available options:
        --animal=ANIMAL  A pet  (one of: cat, dog)
    -h, --help           Prints help information
`, "")
}

func TestChoiceCompletion(t *testing.T) {
	parser := Long("animal").Argument("ANIMAL", ParseString).Choice("dog", "cat").ToOptions()

	assertCompletion(t, parser, "cat\tcat\t\t\ndog\tdog\t\t\n\n", `--animal ""`)
}
