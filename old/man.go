package cli

import (
	"fmt"
	"io"
	"os"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/samber/lo"

	"github.com/rprtr258/cli/internal"
)

func manQuote(s string) string {
	return strings.ReplaceAll(s, `\`, `\\`)
}

func manQuoteLines(s string) string {
	return strings.Join(lo.Map(strings.Split(s, "\n"), func(line string, _ int) string {
		return manQuote(line)
	}), "\n")
}

func formatForMan(w io.Writer, s string, quoter func(s string) string) {
	for {
		idx := strings.IndexRune(s, '`')

		if idx < 0 {
			fmt.Fprintf(w, "%s", quoter(s))
			break
		}

		fmt.Fprintf(w, "%s", quoter(s[:idx]))

		s = s[idx+1:]
		idx = strings.IndexRune(s, '\'')

		if idx < 0 {
			fmt.Fprintf(w, "%s", quoter(s))
			break
		}

		fmt.Fprintf(w, "\\fB%s\\fP", quoter(s[:idx]))
		s = s[idx+1:]
	}
}

func quoteV(s []string) []string {
	return internal.Map(strconv.Quote, s...)
}

func writeManPageOptions(w io.Writer, grp *Group) {
	grp.eachGroup(func(group *Group) {
		if !group.showInHelp() {
			return
		}

		// If the parent (grp) has any subgroups, display their descriptions as
		// subsection headers similar to the output of --help.
		if group.ShortDescription != "" && len(grp.groups) > 0 {
			fmt.Fprintf(w, ".SS %s\n", group.ShortDescription)

			if group.LongDescription != "" {
				formatForMan(w, group.LongDescription, manQuoteLines)
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

			if opt.ValueName != "" || opt.OptionalArgument {
				if opt.OptionalArgument {
					fmt.Fprintf(w, " [\\fI%s=%s\\fR]", manQuote(opt.ValueName), manQuote(strings.Join(quoteV(opt.OptionalValue), ", ")))
				} else {
					fmt.Fprintf(w, " \\fI%s\\fR", manQuote(opt.ValueName))
				}
			}

			if len(opt.Default) != 0 {
				fmt.Fprintf(w, " <default: \\fI%s\\fR>", manQuote(strings.Join(quoteV(opt.Default), ", ")))
			} else if opt.EnvKeyWithNamespace() != "" {
				if runtime.GOOS == "windows" {
					fmt.Fprintf(w, " <default: \\fI%%%s%%\\fR>", manQuote(opt.EnvKeyWithNamespace()))
				} else {
					fmt.Fprintf(w, " <default: \\fI$%s\\fR>", manQuote(opt.EnvKeyWithNamespace()))
				}
			}

			if opt.Required {
				fmt.Fprintf(w, " (\\fIrequired\\fR)")
			}

			fmt.Fprintln(w, "\\fP")

			if opt.Description != "" {
				formatForMan(w, opt.Description, manQuoteLines)
				fmt.Fprintln(w, "")
			}
		}
	})
}

func writeManPageSubcommands(w io.Writer, name, usagePrefix string, root *Command) {
	commands := root.sortedVisibleCommands()

	for _, c := range commands {
		if c.Hidden {
			continue
		}

		nn := name + lo.Ternary(name != "", " ", "") + c.Name
		writeManPageCommand(w, nn, usagePrefix, c)
	}
}

func writeManPageCommand(w io.Writer, name, usagePrefix string, command *Command) {
	fmt.Fprintf(w, ".SS %s\n", name)
	fmt.Fprintln(w, command.ShortDescription)

	if len(command.LongDescription) > 0 {
		fmt.Fprintln(w, "")

		cmdstart := fmt.Sprintf("The %s command", manQuote(command.Name))

		if strings.HasPrefix(command.LongDescription, cmdstart) {
			fmt.Fprintf(w, "The \\fI%s\\fP command", manQuote(command.Name))

			formatForMan(w, command.LongDescription[len(cmdstart):], manQuoteLines)
			fmt.Fprintln(w, "")
		} else {
			formatForMan(w, command.LongDescription, manQuoteLines)
			fmt.Fprintln(w, "")
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
	if len(usage) > 0 {
		fmt.Fprintf(w, "\n\\fBUsage\\fP: %s %s\n.TP\n", manQuote(pre), manQuote(usage))
		nextPrefix = pre + " " + usage
	}

	if len(command.Aliases) > 0 {
		fmt.Fprintf(w, "\n\\fBAliases\\fP: %s\n\n", manQuote(strings.Join(command.Aliases, ", ")))
	}

	writeManPageOptions(w, command.Group)
	writeManPageSubcommands(w, name, nextPrefix, command)
}

// WriteManPage writes a basic man page in groff format to the specified writer.
func (p *Parser) WriteManPage(w io.Writer) {
	t := time.Now()
	source_date_epoch := os.Getenv("SOURCE_DATE_EPOCH")
	if source_date_epoch != "" {
		sde, err := strconv.ParseInt(source_date_epoch, 10, 64)
		if err != nil {
			panic(fmt.Sprintf("Invalid SOURCE_DATE_EPOCH: %s", err))
		}
		t = time.Unix(sde, 0)
	}

	fmt.Fprintf(w, ".TH %s 1 \"%s\"\n", manQuote(p.Name), t.Format("2 January 2006"))
	fmt.Fprintln(w, ".SH NAME")
	fmt.Fprintf(w, "%s \\- %s\n", manQuote(p.Name), manQuoteLines(p.ShortDescription))
	fmt.Fprintln(w, ".SH SYNOPSIS")

	usage := internal.Or(p.Usage, "[OPTIONS]")

	fmt.Fprintf(w, "\\fB%s\\fP %s\n", manQuote(p.Name), manQuote(usage))
	fmt.Fprintln(w, ".SH DESCRIPTION")

	formatForMan(w, p.LongDescription, manQuoteLines)
	fmt.Fprintln(w, "")

	fmt.Fprintln(w, ".SH OPTIONS")

	writeManPageOptions(w, p.Command.Group)

	if len(p.visibleCommands()) > 0 {
		fmt.Fprintln(w, ".SH COMMANDS")
		writeManPageSubcommands(w, "", p.Name+" "+usage, p.Command)
	}
}
