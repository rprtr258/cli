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

func (p *Command) writeHelp() string {
	p.posArgs = append(p.posArgs, PositionalArg{
		name:        "name",
		description: "Who we say hello to",
	})
	return scuf.NewString(func(b scuf.Buffer) {
		// usage line
		b.
			String("Usage: ").
			String("kubectl", scuf.FgHiBlue).
			SPC().
			String("[OPTIONS]", scuf.FgHiYellow).
			SPC().
			String("[ARGUMENTS]", scuf.FgHiMagenta)
		// description
		b.
			NL().
			NL().
			String("Print a friendly greeting")
		// positional args
		argsPreLen := internal.MaxBy(func(a PositionalArg) int {
			return 1 + len(a.name) + 1 + 1 + len("string")
		}, p.posArgs...) + 4
		b.
			NL().
			NL().
			String("Arguments:").
			NL().
			Iter(func(yield func(func(scuf.Buffer)) bool) bool {
				for _, a := range p.posArgs {
					yield(func(b scuf.Buffer) {
						b.
							TAB().
							Styled(func(b scuf.Buffer) {
								b.InBytePair('[', ']', func(b scuf.Buffer) {
									b.String(a.name)
								})
							}, scuf.FgHiMagenta).
							SPC().
							String("string", scuf.FgHiGreen).
							RepeatByte(' ', argsPreLen-1-len(a.name)-1-1-len("string")).
							String(a.description).
							NL()
					})
				}
				return true
			})
		// flags
		flagsPreLen := internal.MaxBy(func(f Flag) int {
			return len(helpFlag(f))
		}, p.flags...) + 4
		b.
			NL().
			String("Options:").
			NL().
			Iter(func(yield func(func(scuf.Buffer)) bool) bool {
				for _, f := range p.flags {
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
							String("short is " + string(f.shortname) + ", long is " + f.longname).
							NL()
					})
				}
				return true
			})
	})
}
