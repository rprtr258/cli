package flags

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

func (a *alignmentInfo) descriptionStart() int {
	ret := a.maxLongLen + distanceBetweenOptionAndDescription

	if a.hasShort {
		ret += 2
	}

	if a.maxLongLen > 0 {
		ret += 4
	}

	if a.hasValueName {
		ret += 3
	}

	return ret
}

func (a *alignmentInfo) updateLen(name string, indent bool) {
	l := utf8.RuneCountInString(name)

	if indent {
		l = l + 4
	}

	if l > a.maxLongLen {
		a.maxLongLen = l
	}
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
			for _, arg := range c.args {
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
	var ret string

	if l < 10 {
		l = 10
	}

	// Basic text wrapping of s at spaces to fit in l
	lines := strings.Split(s, "\n")

	for _, line := range lines {
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

			if len(retline) != 0 {
				retline += "\n" + prefix
			}

			retline += strings.TrimSpace(line[:pos]) + suffix
			line = strings.TrimSpace(line[pos:])
		}

		if len(line) > 0 {
			if len(retline) != 0 {
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

	line.WriteString(strings.Repeat(" ", prefix))

	if option.ShortName != 0 {
		line.WriteRune(defaultShortOptDelimiter)
		line.WriteRune(option.ShortName)
	} else if info.hasShort {
		line.WriteString("  ")
	}

	descstart := info.descriptionStart() + paddingBeforeOption

	if len(option.LongName) > 0 {
		if option.ShortName != 0 {
			line.WriteString(", ")
		} else if info.hasShort {
			line.WriteString("  ")
		}

		line.WriteString(defaultLongOptDelimiter)
		line.WriteString(option.LongNameWithNamespace())
	}

	if option.canArgument() {
		line.WriteRune(defaultNameArgDelimiter)

		if len(option.ValueName) > 0 {
			line.WriteString(option.ValueName)
		}

		if len(option.Choices) > 0 {
			line.WriteString("[" + strings.Join(option.Choices, "|") + "]")
		}
	}

	written := line.Len()
	line.WriteTo(writer)

	if option.Description != "" {
		dw := descstart - written
		writer.WriteString(strings.Repeat(" ", dw))

		var def string

		if len(option.DefaultMask) != 0 {
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

		writer.WriteString(wrapText(desc,
			info.terminalColumns-descstart,
			strings.Repeat(" ", descstart)))
	}

	writer.WriteString("\n")
}

func maxCommandLength(s []*Command) int {
	ret := 0
	for _, v := range s {
		aliasesTotalLen := 0
		for _, alias := range v.Aliases {
			aliasesTotalLen += len(alias)
		}
		ret = max(ret, len(v.Name)+len(v.Aliases)+aliasesTotalLen)
	}
	return ret
}

// WriteHelp writes a help message containing all the possible options and
// their descriptions to the provided writer. Note that the HelpFlag parser
// option provides a convenient way to add a -h/--help option group to the
// command line parser which will automatically show the help messages using
// this method.
func (p *Parser) WriteHelp(w io.Writer) {
	if w == nil {
		return
	}

	bw := bufio.NewWriter(w)
	defer bw.Flush()

	aligninfo := p.getAlignmentInfo()

	cmd := p.Command

	for cmd.Active != nil {
		cmd = cmd.Active
	}

	if p.Name != "" {
		bw.WriteString("Usage:\n")
		bw.WriteString(" ")

		allcmd := p.Command

		for allcmd != nil {
			var usage string

			if allcmd == p.Command {
				if len(p.Usage) != 0 {
					usage = p.Usage
				} else if p.Options&HelpFlag != 0 {
					usage = "[OPTIONS]"
				}
			} else if us, ok := allcmd.data.(Usage); ok {
				usage = us.Usage()
			} else if allcmd.hasHelpOptions() {
				usage = fmt.Sprintf("[%s-OPTIONS]", allcmd.Name)
			}

			if len(usage) != 0 {
				fmt.Fprintf(bw, " %s %s", allcmd.Name, usage)
			} else {
				fmt.Fprintf(bw, " %s", allcmd.Name)
			}

			if len(allcmd.args) > 0 {
				fmt.Fprintf(bw, " ")
			}

			for i, arg := range allcmd.args {
				if i != 0 {
					fmt.Fprintf(bw, " ")
				}

				name := arg.Name

				if arg.isRemaining() {
					name = name + "..."
				}

				if !allcmd.ArgsRequired {
					if arg.Required > 0 {
						fmt.Fprintf(bw, "%s", name)
					} else {
						fmt.Fprintf(bw, "[%s]", name)
					}
				} else {
					fmt.Fprintf(bw, "%s", name)
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
					fmt.Fprintf(bw, " %scommand%s", co, cc)
				} else {
					subcommands := allcmd.sortedVisibleCommands()
					names := make([]string, len(subcommands))

					for i, subc := range subcommands {
						names[i] = subc.Name
					}

					fmt.Fprintf(bw, " %s%s%s", co, strings.Join(names, " | "), cc)
				}
			}

			allcmd = allcmd.Active
		}

		fmt.Fprintln(bw)

		if len(cmd.LongDescription) != 0 {
			fmt.Fprintln(bw)

			t := wrapText(cmd.LongDescription,
				aligninfo.terminalColumns,
				"")

			fmt.Fprintln(bw, t)
		}
	}

	c := p.Command

	for c != nil {
		printcmd := c != p.Command

		c.eachGroup(func(grp *Group) {
			first := true

			// Skip built-in help group for all commands except the top-level parser
			if grp.Hidden || grp.isBuiltinHelp && c != p.Command {
				return
			}

			for _, info := range grp.options {
				if !info.showInHelp() {
					continue
				}

				if printcmd {
					fmt.Fprintf(bw, "\n[%s command options]\n", c.Name)
					aligninfo.indent = true
					printcmd = false
				}

				if first && cmd.Group != grp {
					fmt.Fprintln(bw)

					if aligninfo.indent {
						bw.WriteString("    ")
					}

					fmt.Fprintf(bw, "%s:\n", grp.ShortDescription)
					first = false
				}

				p.writeHelpOption(bw, info, aligninfo)
			}
		})

		if args := internal.Filter(func(arg *Arg) bool {
			// TODO: separate Hidden field
			return arg.Description != ""
		}, c.args...); len(args) > 0 {
			if c == p.Command {
				fmt.Fprintf(bw, "\nArguments:\n")
			} else {
				fmt.Fprintf(bw, "\n[%s command arguments]\n", c.Name)
			}

			descStart := aligninfo.descriptionStart() + paddingBeforeOption

			for _, arg := range args {
				argPrefix := strings.Repeat(" ", paddingBeforeOption) + arg.Name
				if len(arg.Description) > 0 {
					argPrefix += ":"
					bw.WriteString(argPrefix)

					// Space between "arg:" and the description start
					bw.WriteString(strings.Repeat(" ", descStart-len(argPrefix)))
					bw.WriteString(wrapText(
						arg.Description,
						// How much space the description gets before wrapping
						aligninfo.terminalColumns-1-descStart,
						// Whitespace to which we can indent new description lines
						strings.Repeat(" ", descStart),
					))
				} else {
					bw.WriteString(argPrefix)
				}

				fmt.Fprintln(bw)
			}
		}

		c = c.Active
	}

	if scommands := cmd.sortedVisibleCommands(); len(scommands) > 0 {
		maxnamelen := maxCommandLength(scommands)

		fmt.Fprintln(bw)
		fmt.Fprintln(bw, "Available commands:")

		for _, c := range scommands {
			fullNameRepresentation := c.Name
			if len(c.Aliases) > 0 {
				fullNameRepresentation += "," + strings.Join(c.Aliases, ",")
			}
			fmt.Fprint(bw, "  "+fullNameRepresentation)

			if len(c.ShortDescription) > 0 {
				pad := strings.Repeat(" ", maxnamelen-len(fullNameRepresentation))
				fmt.Fprintf(bw, "%s  %s", pad, c.ShortDescription)
			}

			fmt.Fprintln(bw)
		}
	}
}

// WroteHelp is a helper to test the error from ParseArgs() to
// determine if the help message was written. It is safe to
// call without first checking that error is nil.
func WroteHelp(err error) bool {
	if err == nil { // No error
		return false
	}

	flagError, ok := err.(*Error)
	if !ok { // Not a go-flag error
		return false
	}

	if flagError.Type != ErrHelp { // Did not print the help message
		return false
	}

	return true
}
