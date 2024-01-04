package cli

import (
	"fmt"

	"github.com/rprtr258/cli/internal"
	"github.com/rprtr258/scuf"
)

func helpFlag(f Flag) string {
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

func (c *Command) writeHelp(fullCommandName string) string {
	c.posArgs = append(c.posArgs, PositionalArg{
		name:        "name",
		description: "Who we say hello to",
	})
	c.subcommands["get"] = &Command{
		description: "Get entity",
	}

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
		if c.posArgsRest.defined {
			b.SPC().String("[REST...]", scuf.FgHiMagenta)
		}

		// description
		b.
			NL().
			NL().
			String("Print a friendly greeting")

		// subcommands
		if len(c.subcommands) > 0 {
			subcommandsPreLen := 0
			for subcommandName := range c.subcommands {
				subcommandsPreLen = max(subcommandsPreLen, len(subcommandName))
			}
			subcommandsPreLen += 4
			b.
				NL().
				NL().
				String("Commands:").
				Iter(func(yield func(func(scuf.Buffer)) bool) bool {
					for name, subc := range c.subcommands {
						yield(func(b scuf.Buffer) {
							b.
								NL().
								TAB().
								String(name, scuf.FgHiRed).
								SPC().
								RepeatByte(' ', subcommandsPreLen-len(name)).
								String(subc.description)
						})
					}
					return true
				})
		}

		// positional args
		if len(c.posArgs) > 0 {
			argsPreLen := internal.MaxBy(func(a PositionalArg) int {
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
			flagsPreLen := internal.MaxBy(func(f Flag) int {
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
								String(f.description)
							if f.defaultValue != "" {
								b.
									String(", (default: ").
									String(f.defaultValue).
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
