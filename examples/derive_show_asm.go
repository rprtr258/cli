// `derive_show_asm.go` - parsing snippet from cargo-show-asm.
// Derive + typed fallback + external both with and without name.
package examples

import (
	"fmt"
	"iter"
	"os"
	"path/filepath"
	"strings"

	"github.com/rprtr258/cli"
	"github.com/rprtr258/fun"
)

type asmSyntax uint8

const (
	// Generate assembly using Intel style
	asmSyntaxIntel asmSyntax = iota
	// Generate assembly using AT&T style
	asmSyntaxAtt
)

// Select artifact to use for analysis, only one is valid
type asmFocusKind uint8

const (
	// Show results from library code
	asmFocusLib asmFocusKind = iota
	// Show results from a test
	asmFocusTest
	// Show results from a benchmark
	asmFocusBench
	// Show results from an example
	asmFocusExample
	// Show results from a binary
	asmFocusBin
)

type asmFocus struct {
	kind asmFocusKind
	arg  string
}

// How to render output
type asmFormat struct {
	rust     bool
	color    bool
	fullName bool
}

// Item to pick from the output
type asmSelectedFunction struct {
	// Complete or partial function name to filter
	function cli.Opt[string]
	// Select nth item from a filtered list
	nth uint
}

type asmOptions struct {
	manifestPath     string
	targetDir        cli.Opt[string]
	pkg              cli.Opt[string]
	focus            cli.Opt[asmFocus]
	dry              bool
	frozen           bool
	locked           bool
	offline          bool
	format           asmFormat
	syntax           asmSyntax
	selectedFunction asmSelectedFunction
}

func asmParseManifestPath() cli.Parser[string] {
	return cli.Long("manifest-path").
		Help("Path to Cargo.toml").
		Argument("PATH", cli.ParseString).
		CompleteShell(cli.ShellCompFile(cli.Valid("*.toml"))).
		Map(func(p string) string {
			if filepath.IsAbs(p) {
				return p
			}
			return p // canonicalization is up to the caller, see os.Getwd
		}).
		FallbackWith(func() (string, error) {
			wd, err := os.Getwd()
			if err != nil {
				return "", err
			}
			// cargo-metadata wants to see the full path
			return filepath.Join(wd, "Cargo.toml"), nil
		})
}

func asmColorDetection() cli.Parser[bool] {
	yes := cli.Long("color").Help("Enable color highlighting").ReqFlag(true)
	no := cli.Long("no-color").Help("Disable color highlighting").ReqFlag(false)
	return cli.Choice(yes, no).FallbackWith(func() (bool, error) {
		// we could call into a supports-color library here
		return true, nil
	})
}

func asmCompExamples(prefix string) iter.Seq2[string, cli.Opt[string]] {
	// in the actual app we can ask cargo-metadata for this info
	examples := []string{"derive_show_asm", "coreutils", "comonad"}
	return func(yield func(string, cli.Opt[string]) bool) {
		for _, e := range examples {
			if !strings.HasPrefix(e, prefix) {
				continue
			}
			if !yield(e, cli.Invalid[string]()) {
				return
			}
		}
	}
}

func asmParseFocus() cli.Parser[asmFocus] {
	lib := cli.Long("lib").Help("Show results from library code").ReqFlag(asmFocus{kind: asmFocusLib})
	test := cli.Long("test").
		Help("Show results from a test").
		Argument("TEST", cli.ParseString).
		Map(func(s string) asmFocus { return asmFocus{kind: asmFocusTest, arg: s} })
	bench := cli.Long("bench").
		Help("Show results from a benchmark").
		Argument("BENCH", cli.ParseString).
		Map(func(s string) asmFocus { return asmFocus{kind: asmFocusBench, arg: s} })
	example := cli.Long("example").
		Help("Show results from an example").
		Argument("EXAMPLE", cli.ParseString).
		Complete(asmCompExamples).
		Map(func(s string) asmFocus { return asmFocus{kind: asmFocusExample, arg: s} })
	bin := cli.Long("bin").
		Help("Show results from a binary").
		Argument("BIN", cli.ParseString).
		Map(func(s string) asmFocus { return asmFocus{kind: asmFocusBin, arg: s} })

	return cli.Choice(lib, test, bench, example, bin)
}

func asmParseSyntax() cli.Parser[asmSyntax] {
	intel := cli.Long("intel").Help("Generate assembly using Intel style").ReqFlag(asmSyntaxIntel)
	att := cli.Long("att").Help("Generate assembly using AT&T style").ReqFlag(asmSyntaxAtt)
	return cli.Choice(intel, att).Fallback(asmSyntaxIntel)
}

func asmParseFormat() cli.Parser[asmFormat] {
	return cli.Map3(
		cli.Long("rust").Help("Print interleaved Rust code").Switch(),
		asmColorDetection(),
		cli.Long("full-name").Help("include full demangled name instead of just prefix").Switch(),
		func(rust, color, fullName bool) asmFormat {
			return asmFormat{rust: rust, color: color, fullName: fullName}
		},
	)
}

func asmParseSelectedFunction() cli.Parser[asmSelectedFunction] {
	return cli.Map2(
		cli.Positional("FUNCTION", cli.ParseString).
			Help("Complete or partial function name to filter").
			Optional(),
		cli.Positional("INDEX", cli.ParseUint).
			Help("Select nth item from a filtered list").
			Fallback(0),
		func(function cli.Opt[string], nth uint) asmSelectedFunction {
			return asmSelectedFunction{function: function, nth: nth}
		},
	)
}

type asmFlags struct {
	dry     bool
	frozen  bool
	locked  bool
	offline bool
}

func asmParseFlags() cli.Parser[asmFlags] {
	return cli.Map4(
		cli.Long("dry").Help("Produce a build plan instead of actually building").Switch(),
		cli.Long("frozen").Help("Requires Cargo.lock and cache are up to date").Switch(),
		cli.Long("locked").Help("Requires Cargo.lock is up to date").Switch(),
		cli.Long("offline").Help("Run without accessing the network").Switch(),
		func(dry, frozen, locked, offline bool) asmFlags {
			return asmFlags{dry: dry, frozen: frozen, locked: locked, offline: offline}
		},
	)
}

type asmHead struct {
	manifestPath string
	targetDir    cli.Opt[string]
	pkg          cli.Opt[string]
}

type asmMid struct {
	flags  asmFlags
	format asmFormat
	syntax asmSyntax
}

func DeriveShowAsm() {
	// Custom target directory for generated artifacts
	targetDir := cli.Long("target-dir").
		Help("Custom target directory for generated artifacts").
		Argument("DIR", cli.ParseString).
		Optional()
	// Package to use if ambigous
	pkg := cli.Short('p').Long("package").
		Help("Package to use if ambigous").
		Argument("SPEC", cli.ParseString).
		Optional()

	head := cli.Map3(
		asmParseManifestPath(), targetDir, pkg,
		func(manifestPath string, targetDir, pkg cli.Opt[string]) asmHead {
			return asmHead{manifestPath: manifestPath, targetDir: targetDir, pkg: pkg}
		},
	)
	mid := cli.Map3(
		asmParseFlags(), asmParseFormat(), asmParseSyntax(),
		func(flags asmFlags, format asmFormat, syntax asmSyntax) asmMid {
			return asmMid{flags: flags, format: format, syntax: syntax}
		},
	)
	tail := cli.Pair(asmParseFocus().Optional(), asmParseSelectedFunction())

	opts := cli.Map3(
		head, mid, tail,
		func(head asmHead, mid asmMid, tail fun.Pair[cli.Opt[asmFocus], asmSelectedFunction]) asmOptions {
			return asmOptions{
				manifestPath:     head.manifestPath,
				targetDir:        head.targetDir,
				pkg:              head.pkg,
				focus:            tail.K,
				dry:              mid.flags.dry,
				frozen:           mid.flags.frozen,
				locked:           mid.flags.locked,
				offline:          mid.flags.offline,
				format:           mid.format,
				syntax:           mid.syntax,
				selectedFunction: tail.V,
			}
		},
	).
		ToOptions().
		Run(os.Args)

	fmt.Println(opts)
}
