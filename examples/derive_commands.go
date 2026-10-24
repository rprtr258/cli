// `derive_commands.rs` - snippet from cargo-hackerman crate, shows how to parse
// commands and conditional skip for options.
//
// Command explain takes 3 parameters: required crate name and optional feature and
// crate version, user is allowed to omit either field. This example uses simplified
// isVersion, in practice you would use semver crate.
//
// End user would be able to run commands like
//
//	$ app explain random 314
//	> krate: "random", feature: None, version: Some(314)
//	$ app explain serde derive
//	> krate: "serde", feature: Some("derive"), version: None
package examples

import (
	"fmt"
	"os"

	"github.com/rprtr258/cli"
)

type deriveCommandsExplain struct {
	krate   string
	feature cli.Opt[string]
	version cli.Opt[string]
}

func deriveCommandsIsVersion(v string) bool {
	for _, c := range v {
		if c < '0' || c > '9' {
			return false
		}
	}
	return true
}

func deriveCommandsFeatureIf() cli.Parser[cli.Opt[string]] {
	// here feature starts as any string on a command line that does not start with a dash
	return cli.Positional("FEATURE", cli.ParseString).
		// guard restricts it such that it can't be a valid version
		Guard(func(s string) bool { return !deriveCommandsIsVersion(s) }, "").
		Parser().
		// OptionalCatch allows the parser to represent an ignored value with None
		// and to handle parse failures coming from guard
		OptionalCatch()
}

func deriveCommandsVersionIf() cli.Parser[cli.Opt[string]] {
	return cli.Positional("VERSION", cli.ParseString).
		Guard(deriveCommandsIsVersion, "").
		Parser().
		OptionalCatch()
}

func DeriveCommands() {
	explain := cli.Map3(
		cli.Positional("CRATE", cli.ParseString).Parser(),
		deriveCommandsFeatureIf(),
		deriveCommandsVersionIf(),
		func(krate string, feature, version cli.Opt[string]) deriveCommandsExplain {
			return deriveCommandsExplain{krate: krate, feature: feature, version: version}
		},
	).
		Parser().
		ToOptions().
		Command("explain").
		Parser()

	opts := explain.ToOptions().FallbackToUsage().Run(os.Args)
	fmt.Println(opts)
}
