package cli

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"runtime"
	"strings"
	"unicode/utf8"

	"github.com/rprtr258/cli/internal"
)

type alignmentInfo struct {
	maxLongLen      int
	hasShort        bool
	hasValueName    bool
	terminalColumns int
	indent          bool
}

const (
	paddingBeforeOption                 = 2
	distanceBetweenOptionAndDescription = 2
)

// b01 maps true to 1 and false to 0
func b01(b bool) int {
	if b {
		return 1
	}
	return 0
}

func (a *alignmentInfo) descriptionStart() int {
	return a.maxLongLen +
		distanceBetweenOptionAndDescription +
		b01(a.hasShort)*2 +
		b01(a.maxLongLen > 0)*4 +
		b01(a.hasValueName)*3
}

func (a *alignmentInfo) updateLen(name string, indent bool) {
	l := utf8.RuneCountInString(name) + b01(indent)*4
	a.maxLongLen = max(a.maxLongLen, l)
}

func (p *Parser) getAlignmentInfo() alignmentInfo {
	ret := alignmentInfo{
		maxLongLen:      0,
		hasShort:        false,
		hasValueName:    false,
		terminalColumns: getTerminalColumns(),
	}

	if ret.terminalColumns <= 0 {
		ret.terminalColumns = 80
	}

	var prevcmd *Command
	p.eachActiveGroup(func(c *Command, grp *Group) {
		if c != prevcmd {
			for _, arg := range c.Args {
				ret.updateLen(arg.Name, c != p.Command)
			}
			prevcmd = c
		}
		if !grp.showInHelp() {
			return
		}
		for _, info := range grp.options {
			if !info.showInHelp() {
				continue
			}

			if info.ShortName != 0 {
				ret.hasShort = true
			}

			if len(info.ValueName) > 0 {
				ret.hasValueName = true
			}

			l := info.LongNameWithNamespace() + info.ValueName

			if len(info.Choices) != 0 {
				l += "[" + strings.Join(info.Choices, "|") + "]"
			}

			ret.updateLen(l, c != p.Command)
		}
	})
	return ret
}

func wrapText(s string, l int, prefix string) string {
	l = max(l, 10)

	var ret string
	// Basic text wrapping of s at spaces to fit in l
	for _, line := range strings.Split(s, "\n") {
		var retline string
		line = strings.TrimSpace(line)
		for len(line) > l {
			// Try to split on space
			suffix := ""

			pos := strings.LastIndex(line[:l], " ")

			if pos < 0 {
				pos = l - 1
				suffix = "-\n"
			}

			if retline != "" {
				retline += "\n" + prefix
			}

			retline += strings.TrimSpace(line[:pos]) + suffix
			line = strings.TrimSpace(line[pos:])
		}

		if len(line) > 0 {
			if retline != "" {
				retline += "\n" + prefix
			}

			retline += line
		}

		if len(ret) > 0 {
			ret += "\n"

			if len(retline) > 0 {
				ret += prefix
			}
		}

		ret += retline
	}
	return ret
}

func (p *Parser) writeHelpOption(writer *bufio.Writer, option *Option, info alignmentInfo) {
	line := &bytes.Buffer{}

	prefix := paddingBeforeOption

	if info.indent {
		prefix += 4
	}

	if option.Hidden {
		return
	}

	_, _ = line.WriteString(strings.Repeat(" ", prefix))

	if option.ShortName != 0 {
		line.WriteByte(defaultShortOptDelimiter)
		line.WriteRune(option.ShortName)
	} else if info.hasShort {
		_, _ = line.WriteString("  ")
	}

	descstart := info.descriptionStart() + paddingBeforeOption

	if len(option.LongName) > 0 {
		if option.ShortName != 0 {
			_, _ = line.WriteString(", ")
		} else if info.hasShort {
			_, _ = line.WriteString("  ")
		}

		_, _ = line.WriteString(defaultLongOptDelimiter)
		_, _ = line.WriteString(option.LongNameWithNamespace())
	}

	if option.canArgument() {
		line.WriteByte(defaultNameArgDelimiter)

		if len(option.ValueName) > 0 {
			_, _ = line.WriteString(option.ValueName)
		}

		if len(option.Choices) > 0 {
			_, _ = line.WriteString("[" + strings.Join(option.Choices, "|") + "]")
		}
	}

	written := line.Len()
	_, _ = line.WriteTo(writer)

	if option.Description != "" {
		dw := descstart - written
		_, _ = writer.WriteString(strings.Repeat(" ", dw))

		var def string

		if option.DefaultMask != "" {
			if option.DefaultMask != "-" {
				def = option.DefaultMask
			}
		} else {
			def = option.defaultLiteral
		}

		var envDef string
		if option.EnvKeyWithNamespace() != "" {
			var envPrintable string
			if runtime.GOOS == "windows" {
				envPrintable = "%" + option.EnvKeyWithNamespace() + "%"
			} else {
				envPrintable = "$" + option.EnvKeyWithNamespace()
			}
			envDef = fmt.Sprintf(" [%s]", envPrintable)
		}

		var desc string

		if def != "" {
			desc = fmt.Sprintf("%s (default: %v)%s", option.Description, def, envDef)
		} else {
			desc = option.Description + envDef
		}

		_, _ = writer.WriteString(wrapText(desc,
			info.terminalColumns-descstart,
			strings.Repeat(" ", descstart)))
	}

	_, _ = writer.WriteString("\n")
}

func maxCommandLength(s []*Command) int {
	ret := 0
	for _, v := range s {
		ret = max(ret, len(v.Name))
	}
	return ret
}

// WriteHelp writes a help message containing all the possible options and
// their descriptions to the provided writer. Note that the HelpFlag parser
// option provides a convenient way to add a -h/--help option group to the
// command line parser which will automatically show the help messages using
// this method.
func (p *Parser) WriteHelp(writer io.Writer) {
	if writer == nil {
		return
	}

	wr := bufio.NewWriter(writer)
	aligninfo := p.getAlignmentInfo()

	cmd := p.Command
	for cmd.Active != nil {
		cmd = cmd.Active
	}

	if p.Name != "" {
		_, _ = wr.WriteString("Usage:\n")
		_, _ = wr.WriteString(" ")

		for allcmd := p.Command; allcmd != nil; allcmd = allcmd.Active {
			var usage string
			if allcmd == p.Command {
				if p.Usage != "" {
					usage = p.Usage
				} else if p.Options&HelpFlag != 0 {
					usage = "[OPTIONS]"
				}
			} else if us, ok := allcmd.data.(Usage); ok {
				usage = us.Usage()
			} else if allcmd.hasHelpOptions() {
				usage = fmt.Sprintf("[%s-OPTIONS]", allcmd.Name)
			}

			if usage != "" {
				fmt.Fprintf(wr, " %s %s", allcmd.Name, usage)
			} else {
				fmt.Fprintf(wr, " %s", allcmd.Name)
			}

			if len(allcmd.Args) > 0 {
				fmt.Fprintf(wr, " ")
			}

			for i, arg := range allcmd.Args {
				if i != 0 {
					fmt.Fprintf(wr, " ")
				}

				name := arg.Name
				if arg.isRemaining() {
					name += "..."
				}

				if !allcmd.ArgsRequired {
					if arg.Required > 0 {
						fmt.Fprintf(wr, "%s", name)
					} else {
						fmt.Fprintf(wr, "[%s]", name)
					}
				} else {
					fmt.Fprintf(wr, "%s", name)
				}
			}

			if allcmd.Active == nil && len(allcmd.commands) > 0 {
				var co, cc string
				if allcmd.SubcommandsOptional {
					co, cc = "[", "]"
				} else {
					co, cc = "<", ">"
				}

				visibleCommands := allcmd.visibleCommands()

				if len(visibleCommands) > 3 {
					fmt.Fprintf(wr, " %scommand%s", co, cc)
				} else {
					subcommands := allcmd.sortedVisibleCommands()
					names := internal.Map(func(subc *Command) string {
						return subc.Name
					}, subcommands...)

					fmt.Fprintf(wr, " %s%s%s", co, strings.Join(names, " | "), cc)
				}
			}
		}

		fmt.Fprintln(wr)

		if cmd.LongDescription != "" {
			fmt.Fprintln(wr)
			fmt.Fprintln(wr, wrapText(cmd.LongDescription, aligninfo.terminalColumns, ""))
		}
	}

	for c := p.Command; c != nil; c = c.Active {
		printcmd := c != p.Command

		c.eachGroup(func(grp *Group) {
			first := true

			// Skip built-in help group for all commands except the top-level
			// parser
			if grp.Hidden || (grp.isBuiltinHelp && c != p.Command) {
				return
			}

			for _, info := range grp.options {
				if !info.showInHelp() {
					continue
				}

				if printcmd {
					fmt.Fprintf(wr, "\n[%s command options]\n", c.Name)
					aligninfo.indent = true
					printcmd = false
				}

				if first && cmd.Group != grp {
					fmt.Fprintln(wr)

					if aligninfo.indent {
						_, _ = wr.WriteString("    ")
					}

					fmt.Fprintf(wr, "%s:\n", grp.ShortDescription)
					first = false
				}

				p.writeHelpOption(wr, info, aligninfo)
			}
		})

		args := internal.Filter(func(arg Arg) bool {
			return arg.Description != ""
		}, c.Args...)
		if len(args) > 0 {
			if c == p.Command {
				fmt.Fprintf(wr, "\nArguments:\n")
			} else {
				fmt.Fprintf(wr, "\n[%s command arguments]\n", c.Name)
			}

			descStart := aligninfo.descriptionStart() + paddingBeforeOption

			for _, arg := range args {
				argPrefix := strings.Repeat(" ", paddingBeforeOption)
				argPrefix += arg.Name

				if len(arg.Description) > 0 {
					argPrefix += ":"
					_, _ = wr.WriteString(argPrefix)

					// Space between "arg:" and the description start
					descPadding := strings.Repeat(" ", descStart-len(argPrefix))
					// How much space the description gets before wrapping
					descWidth := aligninfo.terminalColumns - 1 - descStart
					// Whitespace to which we can indent new description lines
					descPrefix := strings.Repeat(" ", descStart)

					_, _ = wr.WriteString(descPadding)
					_, _ = wr.WriteString(wrapText(arg.Description, descWidth, descPrefix))
				} else {
					_, _ = wr.WriteString(argPrefix)
				}

				fmt.Fprintln(wr)
			}
		}
	}

	if scommands := cmd.sortedVisibleCommands(); len(scommands) > 0 {
		maxnamelen := maxCommandLength(scommands)

		fmt.Fprintln(wr)
		fmt.Fprintln(wr, "Available commands:")

		for _, c := range scommands {
			fmt.Fprintf(wr, "  %s", c.Name)

			if len(c.ShortDescription) > 0 {
				pad := strings.Repeat(" ", maxnamelen-len(c.Name))
				fmt.Fprintf(wr, "%s  %s", pad, c.ShortDescription)

				if len(c.Aliases) > 0 {
					fmt.Fprintf(wr, " (aliases: %s)", strings.Join(c.Aliases, ", "))
				}

			}

			fmt.Fprintln(wr)
		}
	}

	wr.Flush()
}

// WroteHelp is a helper to test the error from ParseArgs() to
// determine if the help message was written. It is safe to
// call without first checking that error is nil.
func WroteHelp(err error) bool {
	if err == nil {
		return false
	}

	flagError, ok := err.(*Error)
	return ok && flagError.Type == ErrHelp
}
