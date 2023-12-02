package cli

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"reflect"
	"sort"
	"strconv"
	"strings"

	"github.com/samber/lo"

	"github.com/rprtr258/cli/utils"
)

// IniError contains location information on where an error occurred.
type IniError struct {
	Message    string // Message - error message.
	File       string // File name of the file in which the error occurred.
	LineNumber uint   // LineNumber at which the error occurred.
}

// Error provides a "file:line: message" formatted message of the ini error.
func (x *IniError) Error() string {
	return fmt.Sprintf(
		"%s:%d: %s",
		x.File,
		x.LineNumber,
		x.Message,
	)
}

// IniOptions for writing
type IniOptions uint

const (
	IniNone            IniOptions           = 1 << iota >> 1 // IniNone indicates no options.
	IniIncludeDefaults                                       // IniIncludeDefaults indicates that default values should be written.
	IniCommentDefaults                                       // IniCommentDefaults indicates that if IniIncludeDefaults is used options with default values are written but commented out.
	IniIncludeComments                                       // IniIncludeComments indicates that comments containing the description of an option should be written.
	IniDefault         = IniIncludeComments                  // IniDefault provides a default set of options.
)

// IniParser is a utility to read and write flags options from and to ini formatted strings.
type IniParser struct {
	ParseAsDefaults bool // override default flags
	parser          *Parser
}

type iniValue struct {
	Name       string
	Value      string
	Quoted     bool
	LineNumber uint
}

type iniSection []iniValue

type ini struct {
	File     string
	Sections map[string]iniSection
}

// NewIniParser creates a new ini parser for a given Parser.
func NewIniParser(p *Parser) *IniParser {
	return &IniParser{
		parser: p,
	}
}

// IniParse is a convenience function to parse command line options with default
// settings from an ini formatted file. The provided data is a pointer to a struct
// representing the default option group (named "Application Options"). For
// more control, use cli.NewParser.
func IniParse(filename string, data any) error {
	p := NewParser(data, Default)
	return NewIniParser(p).ParseFile(filename)
}

// ParseFile parses flags from an ini formatted file. See Parse for more
// information on the ini file format. The returned errors can be of the type
// cli.Error or cli.IniError.
func (i *IniParser) ParseFile(filename string) error {
	ini, err := readIniFromFile(filename)
	if err != nil {
		return err
	}

	return i.parse(ini)
}

// Parse parses flags from an ini format. You can use ParseFile as a
// convenience function to parse from a filename instead of a general
// io.Reader.
//
// The format of the ini file is as follows:
//
//	[Option group name]
//	option = value
//
// Each section in the ini file represents an option group or command in the
// flags parser. The default flags parser option group (i.e. when using
// cli.Parse) is named 'Application Options'. The ini option name is matched
// in the following order:
//
//  1. Compared to the ini-name tag on the option struct field (if present)
//  2. Compared to the struct field name
//  3. Compared to the option long name (if present)
//  4. Compared to the option short name (if present)
//
// Sections for nested groups and commands can be addressed using a dot `.'
// namespacing notation (i.e [subcommand.Options]). Group section names are
// matched case insensitive.
//
// The returned errors can be of the type cli.Error or cli.IniError.
func (i *IniParser) Parse(reader io.Reader) error {
	ini, err := readIni(reader, "")
	if err != nil {
		return err
	}

	return i.parse(ini)
}

// WriteFile writes the flags as ini format into a file. See Write
// for more information. The returned error occurs when the specified file
// could not be opened for writing.
func (i *IniParser) WriteFile(filename string, options IniOptions) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	writeIni(i, file, options)

	return nil
}

// Write writes the current values of all the flags to an ini format.
// See Parse for more information on the ini file format. You typically
// call this only after settings have been parsed since the default values of each
// option are stored just before parsing the flags (this is only relevant when
// IniIncludeDefaults is _not_ set in options).
func (i *IniParser) Write(w io.Writer, options IniOptions) {
	writeIni(i, w, options)
}

func readFullLine(reader *bufio.Reader) (string, error) {
	var line []byte
	for {
		l, more, err := reader.ReadLine()
		if err != nil {
			return "", err
		}

		if line == nil && !more {
			return string(l), nil
		}

		line = append(line, l...)

		if !more {
			return string(line), nil
		}
	}
}

func optionIniName(option *Option) string {
	if name := option.tag.Get("_read-ini-name"); name != "" {
		return name
	}

	if name := option.tag.Get("ini-name"); name != "" {
		return name
	}

	return option.field.Name
}

func writeGroupIni(cmd *Command, group *Group, namespace string, w io.Writer, options IniOptions) {
	sname := namespace
	if cmd.Group != group && group.ShortDescription != "" {
		if sname != "" {
			sname += "."
		}

		sname += group.ShortDescription
	}

	sectionwritten := false
	comments := (options & IniIncludeComments) != 0
	for _, option := range group.options {
		if option.isFunc() || option.Hidden {
			continue
		}

		if option.tag.Get("no-ini") != "" {
			continue
		}

		val := option.value

		if (options&IniIncludeDefaults) == IniNone && option.valueIsDefault() {
			continue
		}

		if !sectionwritten {
			fmt.Fprintf(w, "[%s]\n", sname)
			sectionwritten = true
		}

		if comments && option.Description != "" {
			fmt.Fprintf(w, "; %s\n", option.Description)
		}

		oname := optionIniName(option)

		commentOption := (options&(IniIncludeDefaults|IniCommentDefaults)) == IniIncludeDefaults|IniCommentDefaults && option.valueIsDefault()

		kind := val.Type().Kind()
		switch kind {
		case reflect.Slice:
			kind = val.Type().Elem().Kind()

			if val.Len() == 0 {
				writeOption(w, oname, kind, "", "", true, option.iniQuote)
			} else {
				for i := 0; i < val.Len(); i++ {
					v, _ := convertToString(val.Index(i), option.tag)
					writeOption(w, oname, kind, "", v, commentOption, option.iniQuote)
				}
			}
		case reflect.Map:
			kind = val.Type().Elem().Kind()

			if val.Len() == 0 {
				writeOption(w, oname, kind, "", "", true, option.iniQuote)
			} else {
				keys := make([]string, len(val.MapKeys()))
				kkmap := make(map[string]reflect.Value)
				for i, k := range val.MapKeys() {
					keys[i], _ = convertToString(k, option.tag)
					kkmap[keys[i]] = k
				}
				sort.Strings(keys)

				for _, k := range keys {
					v, _ := convertToString(val.MapIndex(kkmap[k]), option.tag)
					writeOption(w, oname, kind, k, v, commentOption, option.iniQuote)
				}
			}
		default:
			v, _ := convertToString(val, option.tag)
			writeOption(w, oname, kind, "", v, commentOption, option.iniQuote)
		}

		if comments {
			fmt.Fprintln(w)
		}
	}

	if sectionwritten && !comments {
		fmt.Fprintln(w)
	}
}

func writeOption(
	w io.Writer,
	optionName string,
	optionType reflect.Kind,
	optionKey, optionValue string,
	commentOption, forceQuote bool,
) {
	if forceQuote || optionType == reflect.String && !isPrint(optionValue) {
		optionValue = strconv.Quote(optionValue)
	}

	comment := lo.Ternary(commentOption, "; ", "")

	fmt.Fprintf(w, "%s%s =", comment, optionName)
	if optionKey != "" {
		fmt.Fprintf(w, " %s:%s", optionKey, optionValue)
	} else if optionValue != "" {
		fmt.Fprintf(w, " %s", optionValue)
	}
	fmt.Fprintln(w)
}

func writeCommandIni(command *Command, namespace string, w io.Writer, options IniOptions) {
	command.eachGroup(func(group *Group) {
		if !group.Hidden {
			writeGroupIni(command, group, namespace, w, options)
		}
	})

	for _, c := range command.commands {
		if c.Hidden {
			continue
		}

		fqn := namespace + lo.Ternary(namespace != "", ".", "") + c.Name
		writeCommandIni(c, fqn, w, options)
	}
}

func writeIni(parser *IniParser, w io.Writer, options IniOptions) {
	writeCommandIni(parser.parser.Command, "", w, options)
}

func readIniFromFile(filename string) (*ini, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	return readIni(file, filename)
}

func readIni(r io.Reader, filename string) (*ini, error) {
	ret := &ini{
		File: filename,
		Sections: map[string]iniSection{
			// Empty global section
			"": make(iniSection, 0, 10),
		},
	}

	section := make(iniSection, 0, 10)
	br := bufio.NewReader(r)
	var lineno uint
	sectionname := ""
	for {
		line, err := readFullLine(br)

		if err == io.EOF {
			break
		} else if err != nil {
			return nil, err
		}

		lineno++
		line = strings.TrimSpace(line)

		// Skip empty lines and lines starting with ; (comments)
		if line == "" || utils.OneOf(line[0], ';', '#') {
			continue
		}

		if line[0] == '[' {
			if line[0] != '[' || line[len(line)-1] != ']' {
				return nil, &IniError{
					Message:    "malformed section header",
					File:       filename,
					LineNumber: lineno,
				}
			}

			name := strings.TrimSpace(line[1 : len(line)-1])
			if name == "" {
				return nil, &IniError{
					Message:    "empty section name",
					File:       filename,
					LineNumber: lineno,
				}
			}

			sectionname = name
			section = ret.Sections[name]

			if section == nil {
				section = make(iniSection, 0, 10)
				ret.Sections[name] = section
			}

			continue
		}

		// Parse option here
		keyval := strings.SplitN(line, "=", 2)

		if len(keyval) != 2 {
			return nil, &IniError{
				Message:    fmt.Sprintf("malformed key=value (%s)", line),
				File:       filename,
				LineNumber: lineno,
			}
		}

		name := strings.TrimSpace(keyval[0])
		value := strings.TrimSpace(keyval[1])
		quoted := false

		if value != "" && value[0] == '"' {
			if v, err := strconv.Unquote(value); err == nil {
				value = v

				quoted = true
			} else {
				return nil, &IniError{
					Message:    err.Error(),
					File:       filename,
					LineNumber: lineno,
				}
			}
		}

		section = append(section, iniValue{
			Name:       name,
			Value:      value,
			Quoted:     quoted,
			LineNumber: lineno,
		})

		ret.Sections[sectionname] = section
	}

	return ret, nil
}

func (i *IniParser) matchingGroups(name string) []*Group {
	if name == "" {
		var ret []*Group
		i.parser.eachGroup(func(g *Group) {
			ret = append(ret, g)
		})
		return ret
	}

	g := i.parser.groupByName(name)
	if g != nil {
		return []*Group{g}
	}

	return nil
}

func (i *IniParser) parse(ini *ini) error {
	p := i.parser

	p.eachOption(func(cmd *Command, group *Group, option *Option) {
		option.clearReferenceBeforeSet = true
	})

	quotesLookup := make(map[*Option]bool)

	for name, section := range ini.Sections {
		groups := i.matchingGroups(name)
		if len(groups) == 0 {
			if (p.Options & IgnoreUnknown) == None {
				return newErrorf(ErrUnknownGroup, "could not find option group `%s'", name)
			}

			continue
		}

		for _, inival := range section {
			var opt *Option

			for _, group := range groups {
				opt = group.optionByName(inival.Name, func(o *Option, n string) bool {
					return strings.EqualFold(o.tag.Get("ini-name"), n)
				})

				if opt != nil && opt.tag.Get("no-ini") != "" {
					opt = nil
				}

				if opt != nil {
					break
				}
			}

			if opt == nil {
				if (p.Options & IgnoreUnknown) == None {
					return &IniError{
						Message:    fmt.Sprintf("unknown option: %s", inival.Name),
						File:       ini.File,
						LineNumber: inival.LineNumber,
					}
				}

				continue
			}

			// ini value is ignored if parsed as default but defaults are prevented
			if i.ParseAsDefaults && opt.preventDefault {
				continue
			}

			pval := &inival.Value
			if !opt.canArgument() && inival.Value == "" {
				pval = nil
			} else if opt.value.Type().Kind() == reflect.Map {
				parts := strings.SplitN(inival.Value, ":", 2)

				// only handle unquoting
				if len(parts) == 2 && parts[1][0] == '"' {
					if v, err := strconv.Unquote(parts[1]); err == nil {
						parts[1] = v
						inival.Quoted = true
					} else {
						return &IniError{
							Message:    err.Error(),
							File:       ini.File,
							LineNumber: inival.LineNumber,
						}
					}

					pval = lo.ToPtr(parts[0] + ":" + parts[1])
				}
			}

			var err error
			if i.ParseAsDefaults {
				err = opt.setDefault(pval)
			} else {
				err = opt.Set(pval)
			}

			if err != nil {
				return &IniError{
					Message:    err.Error(),
					File:       ini.File,
					LineNumber: inival.LineNumber,
				}
			}

			// Defaults from ini files take precendence over defaults from parser
			opt.preventDefault = true

			// either all INI values are quoted or only values who need quoting
			if _, ok := quotesLookup[opt]; !inival.Quoted || !ok {
				quotesLookup[opt] = inival.Quoted
			}

			opt.tag.Set("_read-ini-name", inival.Name)
		}
	}

	for opt, quoted := range quotesLookup {
		opt.iniQuote = quoted
	}

	return nil
}
