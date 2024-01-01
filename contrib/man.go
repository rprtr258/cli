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

func formatForMan(wr io.Writer, s string) {
	for {
		idx := strings.IndexRune(s, '`')

		if idx < 0 {
			fmt.Fprintf(wr, "%s", manQuoteLines(s))
			break
		}

		fmt.Fprintf(wr, "%s", manQuoteLines(s[:idx]))

		s = s[idx+1:]
		idx = strings.IndexRune(s, '\'')

		if idx < 0 {
			fmt.Fprintf(wr, "%s", manQuoteLines(s))
			break
		}

		fmt.Fprintf(wr, "\\fB%s\\fP", manQuoteLines(s[:idx]))
		s = s[idx+1:]
	}
}

func writeManPageOptions(wr io.Writer, grp *Group) {
	grp.eachGroup(func(group *Group) {
		if !group.showInHelp() {
			return
		}

		// If the parent (grp) has any subgroups, display their descriptions as
		// subsection headers similar to the output of --help.
		if group.ShortDescription != "" && len(grp.groups) > 0 {
			fmt.Fprintf(wr, ".SS %s\n", group.ShortDescription)

			if group.LongDescription != "" {
				formatForMan(wr, group.LongDescription)
				fmt.Fprintln(wr, "")
			}
		}

		for _, opt := range group.options {
			if !opt.showInHelp() {
				continue
			}

			fmt.Fprintln(wr, ".TP")
			fmt.Fprintf(wr, "\\fB")

			if opt.ShortName != 0 {
				fmt.Fprintf(wr, "\\fB\\-%c\\fR", opt.ShortName)
			}

			if opt.LongName != "" {
				if opt.ShortName != 0 {
					fmt.Fprintf(wr, ", ")
				}

				fmt.Fprintf(wr, "\\fB\\-\\-%s\\fR", manQuote(opt.LongNameWithNamespace()))
			}

			if opt.ValueName != "" || opt.OptionalArgument {
				if opt.OptionalArgument {
					fmt.Fprintf(wr, " [\\fI%s=%s\\fR]", manQuote(opt.ValueName), manQuote(strings.Join(internal.Map(strconv.Quote, opt.OptionalValue...), ", ")))
				} else {
					fmt.Fprintf(wr, " \\fI%s\\fR", manQuote(opt.ValueName))
				}
			}

			if len(opt.Default) != 0 {
				fmt.Fprintf(wr, " <default: \\fI%s\\fR>", manQuote(strings.Join(internal.Map(strconv.Quote, opt.Default...), ", ")))
			}

			if opt.Required {
				fmt.Fprintf(wr, " (\\fIrequired\\fR)")
			}

			fmt.Fprintln(wr, "\\fP")

			if opt.Description != "" {
				formatForMan(wr, opt.Description)
				fmt.Fprintln(wr, "")
			}
		}
	})
}

func writeManPageSubcommands(wr io.Writer, name, usagePrefix string, commands []*Command) {
	iter.
		FromMany(commands...).
		Filter(func(c *Command) bool {
			return !c.Hidden
		}).
		ForEach(func(c *Command) {
			nn := c.Name
			if name != "" {
				nn = name + " " + c.Name
			}

			writeManPageCommand(wr, nn, usagePrefix, c)
		})
}

func writeManPageCommand(wr io.Writer, name, usagePrefix string, command *Command) {
	fmt.Fprintf(wr, ".SS %s\n", name)
	fmt.Fprintln(wr, command.ShortDescription)

	if command.LongDescription != "" {
		fmt.Fprintln(wr, "")

		cmdstart := fmt.Sprintf("The %s command", manQuote(command.Name))

		if strings.HasPrefix(command.LongDescription, cmdstart) {
			fmt.Fprintf(wr, "The \\fI%s\\fP command", manQuote(command.Name))

			formatForMan(wr, command.LongDescription[len(cmdstart):])
			fmt.Fprintln(wr, "")
		} else {
			formatForMan(wr, command.LongDescription)
			fmt.Fprintln(wr, "")
		}
	}

	pre := usagePrefix + " " + command.Name

	var usage string
	if us, ok := command.data.(Usage); ok {
		usage = us.Usage()
	} else if command.hasHelpOptions() {
		usage = fmt.Sprintf("[%s-OPTIONS]", command.Name)
	}

	nextPrefix := pre
	if usage != "" {
		fmt.Fprintf(wr, "\n\\fBUsage\\fP: %s %s\n.TP\n", manQuote(pre), manQuote(usage))
		nextPrefix = pre + " " + usage
	}

	if len(command.Aliases) > 0 {
		fmt.Fprintf(wr, "\n\\fBAliases\\fP: %s\n\n", manQuote(strings.Join(command.Aliases, ", ")))
	}

	writeManPageOptions(wr, command.Group)
	writeManPageSubcommands(wr, name, nextPrefix, visibleCommands(command.commands))
}

// WriteManPage writes a basic man page in groff format to the specified writer.
func (p *Parser) WriteManPage(wr io.Writer, t time.Time) {
	// TODO: move out
	if sourceDateEpoch := os.Getenv("SOURCE_DATE_EPOCH"); sourceDateEpoch != "" {
		sde, err := strconv.ParseInt(sourceDateEpoch, 10, 64)
		if err != nil {
			panic(fmt.Sprintf("Invalid SOURCE_DATE_EPOCH: %s", err))
		}

		t = time.Unix(sde, 0)
	}

	fmt.Fprintf(wr, ".TH %s 1 \"%s\"\n", manQuote(p.Name), t.Format("2 January 2006"))
	fmt.Fprintln(wr, ".SH NAME")
	fmt.Fprintf(wr, "%s \\- %s\n", manQuote(p.Name), manQuoteLines(p.ShortDescription))
	fmt.Fprintln(wr, ".SH SYNOPSIS")

	usage := internal.Or(p.Usage, "[OPTIONS]")

	fmt.Fprintf(wr, "\\fB%s\\fP %s\n", manQuote(p.Name), manQuote(usage))
	fmt.Fprintln(wr, ".SH DESCRIPTION")

	formatForMan(wr, p.LongDescription)
	fmt.Fprintln(wr)

	fmt.Fprintln(wr, ".SH OPTIONS")

	writeManPageOptions(wr, p.Command.Group)

	if len(visibleCommands(p.commands)) > 0 {
		fmt.Fprintln(wr, ".SH COMMANDS")
		writeManPageSubcommands(wr, "", p.Name+" "+usage, visibleCommands(p.Command.commands))
	}
}
