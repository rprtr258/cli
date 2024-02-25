package cli

import (
	"fmt"
	"sort"
	"strings"

	"github.com/rprtr258/cli/internal"
	"github.com/rprtr258/fun"
	"github.com/rprtr258/scuf"
)

func helpFlag(f flag) string {
	switch {
	case f.shortname != 0 && f.longname != "":
		return fmt.Sprintf("-%c, --%s", f.shortname, f.longname)
	case f.shortname != 0:
		return fmt.Sprintf("-%c", f.shortname)
	case f.longname != "":
		return fmt.Sprintf("--%s", f.longname)
	default:
		return ""
	}
}

func (c *command) writeHelp(fullCommandName string) string {
	return scuf.NewString(func(b scuf.Buffer) {
		// usage
		b.String("Usage: ")
		b.String(fullCommandName, scuf.FgHiBlue)
		if len(c.subcommands) > 0 {
			b.SPC().String("[COMMAND]", scuf.FgHiRed)
		}
		if len(c.flags) > 0 {
			b.SPC().String("[FLAGS]", scuf.FgHiYellow)
		}
		if len(c.posArgs) > 0 {
			b.SPC().String("[ARGUMENTS]", scuf.FgHiMagenta)
		}
		if c.posArgsRest.Valid {
			b.SPC().String("[REST...]", scuf.FgHiMagenta)
		}

		// description
		b.
			NL().
			NL().
			String(c.description)

		// subcommands
		if len(c.subcommands) > 0 {
			subcommandsPreLen := 0
			for subcommandName, subc := range c.subcommands {
				aliasesLen := 0
				for _, alias := range subc.aliases {
					aliasesLen += len(alias) + 1
				}

				subcommandsPreLen = max(subcommandsPreLen, len(subcommandName)+aliasesLen)
			}
			subcommandsPreLen += 2
			b.
				NL().
				NL().
				String("Commands:").
				Iter(func(yield func(func(scuf.Buffer)) bool) bool {
					keys := fun.Keys(c.subcommands)
					sort.Strings(keys)

					for _, name := range keys {
						subc := c.subcommands[name]

						aliasesLen := 0
						for _, alias := range subc.aliases {
							aliasesLen += len(alias) + 1
						}

						yield(func(b scuf.Buffer) {
							b.
								NL().
								TAB().
								String(name, scuf.FgHiRed).
								Iter(func(yield func(func(scuf.Buffer)) bool) bool {
									for _, alias := range subc.aliases {
										yield(func(b scuf.Buffer) {
											b.Bytes(',').String(alias, scuf.FgRed)
										})
									}
									return true
								}).
								SPC().
								RepeatByte(' ', subcommandsPreLen-len(name)-aliasesLen).
								String(subc.description)
						})
					}
					return true
				})
		}

		// positional args
		if len(c.posArgs) > 0 {
			argsPreLen := internal.MaxBy(func(a positionalArg) int {
				return 1 + len(a.name) + 1 + 1 + len("string")
			}, c.posArgs...) + 4
			b.
				NL().
				NL().
				String("Arguments:").
				Iter(func(yield func(func(scuf.Buffer)) bool) bool {
					for _, a := range c.posArgs {
						yield(func(b scuf.Buffer) {
							b.
								NL().
								TAB().
								Styled(func(b scuf.Buffer) {
									b.InBytePair('[', ']', func(b scuf.Buffer) {
										b.String(a.name)
									})
								}, scuf.FgHiMagenta).
								SPC().
								String("string", scuf.FgHiGreen).
								RepeatByte(' ', argsPreLen-1-len(a.name)-1-1-len("string")).
								String(a.description)
						})
					}
					return true
				})
		}

		// flags
		if len(c.flags) > 0 {
			flagsPreLen := internal.MaxBy(func(f flag) int {
				return len(helpFlag(f))
			}, c.flags...) + 4
			b.
				NL().
				NL().
				String("Flags:").
				NL().
				Iter(func(yield func(func(scuf.Buffer)) bool) bool {
					for _, f := range c.flags {
						yield(func(b scuf.Buffer) {
							b.
								TAB()
							if f.shortname != 0 {
								b.String("-"+string(f.shortname), scuf.FgHiYellow)
							}
							if f.shortname != 0 && f.longname != "" {
								b.String(", ")
							}
							if f.longname != "" {
								b.String("--"+f.longname, scuf.FgHiYellow)
							}
							b.
								RepeatByte(' ', flagsPreLen-len(helpFlag(f))).
								String(f.description.Value)
							if len(f.defaultValue) > 0 {
								b.
									String(" (default: ").
									String(strings.Join(f.defaultValue, ", ")).
									String(")")
							}
							b.NL()
						})
					}
					return true
				})
		}
	})
}
