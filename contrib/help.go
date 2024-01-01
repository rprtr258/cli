package flags

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
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
	_paddingBeforeOption                 = 2
	_distanceBetweenOptionAndDescription = 2
)

func (a *alignmentInfo) descriptionStart() int {
	return a.maxLongLen + _distanceBetweenOptionAndDescription +
		internal.Iverson(a.hasShort, 2) +
		internal.Iverson(a.maxLongLen > 0, 4) +
		internal.Iverson(a.hasValueName, 3)
}

func (p *Parser) getAlignmentInfo() alignmentInfo {
	ret := alignmentInfo{
		maxLongLen:      0,
		hasShort:        false,
		hasValueName:    false,
		terminalColumns: getTerminalColumns(),
	}
	var prevcmd *Command
	p.eachActiveGroup(func(c *Command, grp *Group) {
		if c != prevcmd {
			for _, arg := range c.args {
				ret.maxLongLen = max(ret.maxLongLen, utf8.RuneCountInString(arg.Name)+internal.Iverson(c != p.Command, 4))
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

			if info.ValueName != "" {
				ret.hasValueName = true
			}

			l := info.LongNameWithNamespace() + info.ValueName +
				internal.Iverson(len(info.Choices) != 0, "["+strings.Join(info.Choices, "|")+"]")
			ret.maxLongLen = max(ret.maxLongLen, utf8.RuneCountInString(l)+internal.Iverson(c != p.Command, 4))
		}
	})
	return ret
}

func wrapText(s string, l int, prefix string) string {
	l = max(l, 10)
	var ret string
	for _, line := range strings.Split(s, "\n") {
		var retline string

		for line = strings.TrimSpace(line); len(line) > l; {
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

		if line != "" {
			if retline != "" {
				retline += "\n" + prefix
			}

			retline += line
		}

		if ret != "" {
			ret += "\n"

			if retline != "" {
				ret += prefix
			}
		}

		ret += retline
	}
	return ret
}

func (p *Parser) writeHelpOption(writer *bufio.Writer, option *Option, info alignmentInfo) {
	if option.Hidden {
		return
	}

	line := &bytes.Buffer{}

	prefixLen := _paddingBeforeOption + internal.Iverson(info.indent, 4)
	line.WriteString(strings.Repeat(" ", prefixLen))

	if option.ShortName != 0 {
		line.WriteByte(defaultShortOptDelimiter)
		line.WriteRune(option.ShortName)
	} else if info.hasShort {
		line.WriteString("  ")
	}

	descstart := info.descriptionStart() + _paddingBeforeOption

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
		line.WriteByte(defaultNameArgDelimiter)

		if option.ValueName != "" {
			line.WriteString(option.ValueName)
		}

		if len(option.Choices) > 0 {
			line.WriteString("[" + strings.Join(option.Choices, "|") + "]")
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

		desc := option.Description
		if def != "" {
			desc = fmt.Sprintf("%s (default: %v)", option.Description, def)
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
		_, _ = bw.WriteString("Usage:\n ")
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

			fmt.Fprintf(bw, " %s", allcmd.Name)
			if usage != "" {
				fmt.Fprintf(bw, " %s", usage)
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
					name += "..."
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

				if visibleCommandss := visibleCommands(allcmd.commands); len(visibleCommandss) > 3 {
					fmt.Fprintf(bw, " %scommand%s", co, cc)
				} else {
					fmt.Fprintf(bw, " "+co+strings.Join(internal.Map(func(subc *Command) string {
						return subc.Name
					}, visibleCommands(allcmd.commands)...), " | ")+cc)
				}
			}
		}

		fmt.Fprintln(bw)

		if cmd.LongDescription != "" {
			fmt.Fprintln(bw)
			fmt.Fprintln(bw, wrapText(cmd.LongDescription,
				aligninfo.terminalColumns,
				""))
		}
	}

	for c := p.Command; c != nil; c = c.Active {
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
						_, _ = bw.WriteString("    ")
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

			descStart := aligninfo.descriptionStart() + _paddingBeforeOption

			for _, arg := range args {
				argPrefix := strings.Repeat(" ", _paddingBeforeOption) + arg.Name
				if arg.Description != "" {
					argPrefix += ":"
					_, _ = bw.WriteString(argPrefix)

					// Space between "arg:" and the description start
					_, _ = bw.WriteString(strings.Repeat(" ", descStart-len(argPrefix)))
					_, _ = bw.WriteString(wrapText(
						arg.Description,
						// How much space the description gets before wrapping
						aligninfo.terminalColumns-1-descStart,
						// Whitespace to which we can indent new description lines
						strings.Repeat(" ", descStart),
					))
				} else {
					_, _ = bw.WriteString(argPrefix)
				}

				fmt.Fprintln(bw)
			}
		}
	}

	if scommands := visibleCommands(cmd.commands); len(scommands) > 0 {
		fmt.Fprintln(bw)
		fmt.Fprintln(bw, "Available commands:")

		maxnamelen := maxCommandLength(scommands)
		for _, c := range scommands {
			fullNameRepresentation := c.Name
			if len(c.Aliases) > 0 {
				fullNameRepresentation += "," + strings.Join(c.Aliases, ",")
			}
			fmt.Fprint(bw, "  "+fullNameRepresentation)

			if c.ShortDescription != "" {
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
	return ok && flagError.Kind == ErrHelp
}
