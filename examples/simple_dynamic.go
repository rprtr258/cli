// `simple_dynamic.rs` - simple dynamic completion example.
package examples

import (
	"fmt"
	"iter"
	"os"
	"strings"

	"github.com/rprtr258/cli"
)

type simpleDynamicFormat uint8

const (
	simpleDynamicFormatText simpleDynamicFormat = iota
	simpleDynamicFormatJson
	simpleDynamicFormatXml
)

func simpleDynamicCrates(input string) iter.Seq2[string, cli.Opt[string]] {
	crates := [][2]string{
		{"cargo-hackerman", "Workspace hack management and package/feature query"},
		{"cargo-prebuilt", "Download prebuilt crate binaries"},
		{"cargo-show-asm", "Display generated assembly"},
		{"cargo-supply-chain", "Gather author, contributor, publisher data on crates"},
		{"chezmoi_modify_manager", "Chezmoi addon to patch ini files"},
		{"xvf", "Easy archive extraction"},
		{"newdoc", "Generate pre-populated module files"},
		{"nust64", "Tools for compiling a Rust project into an N64 ROM"},
		{"uggo", "CLI tool to query builds from u.gg"},
	}

	return func(yield func(string, cli.Opt[string]) bool) {
		for _, crate := range crates {
			if !strings.HasPrefix(crate[0], input) {
				continue
			}
			if !yield(crate[0], cli.Valid(crate[1])) {
				return
			}
		}
	}
}

type simpleDynamicOptions struct {
	name         string
	dependencies bool
	format       simpleDynamicFormat
	upload       cli.Opt[string]
}

func SimpleDynamic() {
	// Select crate for analysis
	name := cli.Long("crate").
		Help("Select crate for analysis").
		Argument("NAME", cli.ParseString).
		Complete(simpleDynamicCrates).
		Parser()

	// Include dependencies into report
	dependencies := cli.Long("dependencies").
		Help("Include dependencies into report").Switch()

	// Format for generated report, fallbacks to plaintext
	format := cli.Choice(
		cli.ReqFlag(cli.Long("json").Help("Generate report in JSON format"), simpleDynamicFormatJson),
		cli.ReqFlag(cli.Long("xml").Help("Generate report in XML format"), simpleDynamicFormatXml),
	).Fallback(simpleDynamicFormatText)

	// Upload report to a url
	upload := cli.Positional("URL", cli.ParseString).
		Help("Upload report to a url").
		Parser().
		Optional()

	opts := cli.Map4(
		name, dependencies, format, upload,
		func(name string, dependencies bool, format simpleDynamicFormat, upload cli.Opt[string]) simpleDynamicOptions {
			return simpleDynamicOptions{
				name:         name,
				dependencies: dependencies,
				format:       format,
				upload:       upload,
			}
		},
	).ToOptions()

	fmt.Println(opts.Run(os.Args))
}
