package flags

import (
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/rprtr258/cli/internal"
	"github.com/rprtr258/fun/iter"
)

func manQuoteLines(s string) string {
	lines := strings.Split(s, "\n")
	parts := []string{}

	for _, line := range lines {
		parts = append(parts, manQuote(line))
	}

	return strings.Join(parts, "\n")
}

func manQuote(s string) string {
	return strings.ReplaceAll(s, "\\", "\\\\")
}

func formatForMan(w io.Writer, s string) {
	for {
		idx := strings.IndexRune(s, '`')

		if idx < 0 {
			fmt.Fprintf(w, "%s", manQuoteLines(s))
			break
		}

		fmt.Fprintf(w, "%s", manQuoteLines(s[:idx]))

		s = s[idx+1:]
		idx = strings.IndexRune(s, '\'')

		if idx < 0 {
			fmt.Fprintf(w, "%s", manQuoteLines(s))
			break
		}

		fmt.Fprintf(w, "\\fB%s\\fP", manQuoteLines(s[:idx]))
		s = s[idx+1:]
	}
}

func writeManPageOptions(w io.Writer, grp *Group) {
	grp.eachGroup(func(group *Group) bool {
		if !group.showInHelp() {
			return true
		}

		// If the parent (grp) has any subgroups, display their descriptions as
		// subsection headers similar to the output of --help.
		if group.ShortDescription != "" && len(grp.groups) > 0 {
			fmt.Fprintf(w, ".SS %s\n", group.ShortDescription)

			if group.LongDescription != "" {
				formatForMan(w, group.LongDescription)
				fmt.Fprintln(w, "")
			}
		}

		for _, opt := range group.options {
			if !opt.showInHelp() {
				continue
			}

			fmt.Fprintln(w, ".TP")
			fmt.Fprintf(w, "\\fB")

			if opt.ShortName != 0 {
				fmt.Fprintf(w, "\\fB\\-%c\\fR", opt.ShortName)
			}

			if opt.LongName != "" {
				if opt.ShortName != 0 {
					fmt.Fprintf(w, ", ")
				}

				fmt.Fprintf(w, "\\fB\\-\\-%s\\fR", manQuote(opt.LongNameWithNamespace()))
			}

			if opt.ValueName != "" || opt.IsOptional {
				if opt.IsOptional {
					fmt.Fprintf(w, " [\\fI%s=%s\\fR]", manQuote(opt.ValueName), manQuote(strings.Join(internal.Map(strconv.Quote, opt.OptionalValue...), ", ")))
				} else {
					fmt.Fprintf(w, " \\fI%s\\fR", manQuote(opt.ValueName))
				}
			}

			if len(opt.Default) != 0 {
				fmt.Fprintf(w, " <default: \\fI%s\\fR>", manQuote(strings.Join(internal.Map(strconv.Quote, opt.Default...), ", ")))
			}

			if opt.Required {
				fmt.Fprintf(w, " (\\fIrequired\\fR)")
			}

			fmt.Fprintln(w, "\\fP")

			if opt.Description != "" {
				formatForMan(w, opt.Description)
				fmt.Fprintln(w, "")
			}
		}
		return true
	})
}

func writeManPageSubcommands(w io.Writer, name, usagePrefix string, commands []*Command) {
	iter.
		FromMany(commands...).
		Filter(func(c *Command) bool {
			return !c.Group.Hidden
		}).
		ForEach(func(command *Command) {
			name := internal.Iverson(name != "", name+" ") + command.Name
			fmt.Fprintf(w, ".SS %s\n", name)
			fmt.Fprintln(w, command.Group.ShortDescription)

			if command.Group.LongDescription != "" {
				fmt.Fprintln(w, "")

				cmdstart := fmt.Sprintf("The %s command", manQuote(command.Name))

				if strings.HasPrefix(command.Group.LongDescription, cmdstart) {
					fmt.Fprintf(w, "The \\fI%s\\fP command", manQuote(command.Name))

					formatForMan(w, command.Group.LongDescription[len(cmdstart):])
				} else {
					formatForMan(w, command.Group.LongDescription)
				}
				fmt.Fprintln(w)
			}

			var usage string
			if us, ok := command.Group.data.(Usage); ok {
				usage = us.Usage()
			} else if command.hasHelpOptions() {
				usage = fmt.Sprintf("[%s-OPTIONS]", command.Name)
			}

			nextPrefix := usagePrefix + " " + command.Name
			if usage != "" {
				fmt.Fprintf(w, "\n\\fBUsage\\fP: %s %s\n.TP\n", manQuote(nextPrefix), manQuote(usage))
				nextPrefix += " " + usage
			}

			if len(command.Aliases) > 0 {
				fmt.Fprintf(w, "\n\\fBAliases\\fP: %s\n\n", manQuote(strings.Join(command.Aliases, ", ")))
			}

			writeManPageOptions(w, command.Group)
			writeManPageSubcommands(w, name, nextPrefix, visibleCommands(command.commands))
		})
}

// WriteManPage writes a basic man page in groff format to the specified writer.
func (p *Parser) WriteManPage(w io.Writer, t time.Time) {
	// TODO: move out
	if sourceDateEpoch := os.Getenv("SOURCE_DATE_EPOCH"); sourceDateEpoch != "" {
		sde, err := strconv.ParseInt(sourceDateEpoch, 10, 64)
		if err != nil {
			panic(fmt.Sprintf("Invalid SOURCE_DATE_EPOCH: %s", err))
		}

		t = time.Unix(sde, 0)
	}

	fmt.Fprintf(w, ".TH %s 1 \"%s\"\n", manQuote(p.Command.Name), t.Format("2 January 2006"))
	fmt.Fprintln(w, ".SH NAME")
	fmt.Fprintf(w, "%s \\- %s\n", manQuote(p.Command.Name), manQuoteLines(p.Command.Group.ShortDescription))
	fmt.Fprintln(w, ".SH SYNOPSIS")

	usage := internal.Or(p.Usage, "[OPTIONS]")

	fmt.Fprintf(w, "\\fB%s\\fP %s\n", manQuote(p.Command.Name), manQuote(usage))
	fmt.Fprintln(w, ".SH DESCRIPTION")

	formatForMan(w, p.Command.Group.LongDescription)
	fmt.Fprintln(w)

	fmt.Fprintln(w, ".SH OPTIONS")

	writeManPageOptions(w, p.Command.Group)

	if len(visibleCommands(p.Command.commands)) > 0 {
		fmt.Fprintln(w, ".SH COMMANDS")
		writeManPageSubcommands(w, "", p.Command.Name+" "+usage, visibleCommands(p.Command.commands))
	}
}
