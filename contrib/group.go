package flags

import (
	"errors"
	"reflect"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/rprtr258/cli/internal"
	"github.com/rprtr258/fun"
	"github.com/rprtr258/fun/iter"
)

func ToDict[K comparable, V, T any](seq iter.Seq[T], f func(T) (K, V)) map[K]V {
	ret := make(map[K]V)
	seq(func(t T) bool {
		k, v := f(t)
		ret[k] = v
		return true
	})
	return ret
}

// Command represents an application command. Commands can be added to the
// parser (which itself is a command) and are selected/executed when its name
// is specified on the command line. The Command type embeds a Group and
// therefore also carries a set of command specific options.
type Command struct {
	// Embedded, see Group for more information
	*Group

	// The name by which the command can be invoked
	Name string

	// The active sub command (set by parsing) or nil
	Active *Command

	// Whether subcommands are optional
	SubcommandsOptional bool

	// Aliases for the command
	Aliases []string

	// Whether positional arguments are required
	ArgsRequired bool

	commands            []*Command
	hasBuiltinHelpGroup bool
	args                []*Arg
}

// Commander is an interface which can be implemented by any command added in
// the options. When implemented, the Execute method will be called for the last
// specified (sub)command providing the remaining command line arguments.
type Commander interface {
	// Execute will be called for the last active (sub)command. The
	// args argument contains the remaining command line arguments. The
	// error that Execute returns will be eventually passed out of the
	// Parse method of the Parser.
	Execute(args []string) error
}

// Usage is an interface which can be implemented to show a custom usage string
// in the help message shown for a command.
type Usage interface {
	// Usage is called for commands to allow customized printing of command
	// usage in the generated help message.
	Usage() string
}

type lookup struct {
	shortNames map[string]*Option
	longNames  map[string]*Option
	commands   map[string]*Command
}

// AddCommand adds a new command to the parser with the given name and data. The
// data needs to be a pointer to a struct from which the fields indicate which
// options are in the command. The provided data can implement the Command and
// Usage interfaces.
func (c *Command) AddCommand(command, shortDescription, longDescription string, data any) (*Command, error) {
	cmd := newCommand(command, shortDescription, longDescription, data)

	cmd.parent = c

	if err := cmd.scan(); err != nil {
		return nil, err
	}

	c.commands = append(c.commands, cmd)
	return cmd, nil
}

// AddGroup adds a new group to the command with the given name and data. The
// data needs to be a pointer to a struct from which the fields indicate which
// options are in the group.
func (c *Command) AddGroup(shortDescription, longDescription string, data any) (*Group, error) {
	group := &Group{
		ShortDescription: shortDescription,
		LongDescription:  longDescription,
		data:             data,
		parent:           c,
	}

	if err := group.scanType(c.scanSubcommandHandler(group)); err != nil {
		return nil, err
	}

	c.groups = append(c.groups, group)
	return group, nil
}

// Commands returns a list of subcommands of this command.
func (c *Command) Commands() []*Command {
	return c.commands
}

// Find locates the subcommand with the given name and returns it. If no such
// command can be found Find will return nil.
func (c *Command) Find(name string) *Command {
	for _, cc := range c.commands {
		if cc.match(name) {
			return cc
		}
	}

	return nil
}

// FindOptionByLongName finds an option that is part of the command, or any of
// its parent commands, by matching its long name (including the option
// namespace).
func (c *Command) FindOptionByLongName(longName string) (option *Option) {
	for option == nil && c != nil {
		option = c.Group.findOptionByLongName(longName)

		c, _ = c.parent.(*Command)
	}

	return option
}

// FindOptionByShortName finds an option that is part of the command, or any of
// its parent commands, by matching its long name (including the option
// namespace).
func (c *Command) FindOptionByShortName(shortName rune) (option *Option) {
	for option == nil && c != nil {
		option = c.Group.findOptionByShortName(shortName)

		c, _ = c.parent.(*Command)
	}

	return option
}

// Args returns a list of positional arguments associated with this command.
func (c *Command) Args() []*Arg {
	ret := make([]*Arg, len(c.args))
	copy(ret, c.args)

	return ret
}

func newCommand(name, shortDescription, longDescription string, data any) *Command {
	return &Command{
		Group: &Group{
			ShortDescription: shortDescription,
			LongDescription:  longDescription,
			data:             data,
		},
		Name: name,
	}
}

func (c *Command) scanSubcommandHandler(parentg *Group) scanHandler {
	f := func(realval reflect.Value, sfield *reflect.StructField) (bool, error) {
		mtag, err := ParseMultiTag(string(sfield.Tag))
		if err != nil {
			return true, err
		}

		positional := mtag.Get("positional-args")

		if positional != "" {
			stype := realval.Type()

			for i := 0; i < stype.NumField(); i++ {
				field := stype.Field(i)

				m, err := ParseMultiTag(string(field.Tag))
				if err != nil {
					return true, err
				}

				name := internal.Or(m.Get("positional-arg-name"), field.Name)

				required := -1
				requiredMaximum := -1

				if sreq := m.Get("required"); sreq != "" {
					required = 1

					if rng := strings.SplitN(sreq, "-", 2); len(rng) > 1 {
						if preq, err := strconv.ParseInt(rng[0], 10, 32); err == nil {
							required = int(preq)
						}

						if preq, err := strconv.ParseInt(rng[1], 10, 32); err == nil {
							requiredMaximum = int(preq)
						}
					} else {
						if preq, err := strconv.ParseInt(sreq, 10, 32); err == nil {
							required = int(preq)
						}
					}
				}

				arg := &Arg{
					Name:            name,
					Description:     m.Get("description"),
					Required:        required,
					RequiredMaximum: requiredMaximum,

					value: realval.Field(i),
					tag:   m,
				}

				c.args = append(c.args, arg)

				if mtag.Get("required") != "" {
					c.ArgsRequired = true
				}
			}

			return true, nil
		}

		subcommand := mtag.Get("command")
		if subcommand == "" {
			return parentg.scanSubGroupHandler(realval, sfield)
		}

		var ptrval reflect.Value

		if realval.Kind() == reflect.Ptr {
			ptrval = realval

			if ptrval.IsNil() {
				ptrval.Set(reflect.New(ptrval.Type().Elem()))
			}
		} else {
			ptrval = realval.Addr()
		}

		shortDescription := mtag.Get("description")
		longDescription := mtag.Get("long-description")
		subc, err := c.AddCommand(subcommand, shortDescription, longDescription, ptrval.Interface())
		if err != nil {
			return true, err
		}

		subc.Hidden = mtag.Get("hidden") != ""

		subcommandsOptional := mtag.Get("subcommands-optional")
		if subcommandsOptional != "" {
			subc.SubcommandsOptional = true
		}

		aliases := mtag.GetMany("alias")
		if len(aliases) > 0 {
			subc.Aliases = aliases
		}

		return true, nil
	}

	return f
}

func (c *Command) scan() error {
	return c.scanType(c.scanSubcommandHandler(c.Group))
}

func (c *Command) eachOption(yield func(*Command, *Group, *Option)) {
	c.eachCommand(func(c *Command) bool {
		return c.eachGroup(func(g *Group) bool {
			for _, option := range g.options {
				yield(c, g, option)
			}
			return true
		})
	})
}

func (c *Command) eachCommand(yield func(*Command) bool) bool {
	return yield(c) && fun.All(c.commands, func(c *Command) bool {
		return c.eachCommand(yield)
	})
}

func (c *Command) addHelpGroups(showHelp func() error) {
	if !c.hasBuiltinHelpGroup {
		c.addHelpGroup(showHelp)
		c.hasBuiltinHelpGroup = true
	}

	for _, cc := range c.commands {
		cc.addHelpGroups(showHelp)
	}
}

func visibleCommands(commands []*Command) []*Command {
	return internal.Filter(func(cmd *Command) bool {
		return !cmd.Hidden
	}, commands...)
}

func (c *Command) match(name string) bool {
	return c.Name == name || internal.OneOf(name, c.Aliases...)
}

func (c *Command) hasHelpOptions() bool {
	ret := false
	c.eachGroup(func(g *Group) bool {
		if g.isBuiltinHelp {
			return true
		}

		for _, opt := range g.options {
			if opt.showInHelp() {
				ret = true
			}
		}
		return true
	})
	return ret
}

func fillParseState(c *Command) lookup {
	var parents []*Command
	for parent := c.parent; parent != nil; {
		if cmd, ok := parent.(*Command); ok {
			parent = cmd.parent
			parents = append(parents, cmd)
		} else {
			break
		}
	}

	opts := iter.FlatMap(
		iter.FlatMap(iter.
			Seq[*Command](
			func(yield func(*Command) bool) bool {
				for i := len(parents) - 1; i >= 0; i-- {
					if !yield(parents[i]) {
						return false
					}
				}
				return yield(c)
			}),
			func(c *Command) iter.Seq[*Group] {
				return c.eachGroup
			}),
		func(g *Group) iter.Seq[*Option] {
			return iter.FromMany(g.options...)
		})

	return lookup{
		shortNames: ToDict(
			opts.Filter(func(o *Option) bool {
				return o.ShortName != 0
			}),
			func(o *Option) (string, *Option) {
				return string(o.ShortName), o
			}),
		longNames: ToDict(
			opts.Filter(func(o *Option) bool {
				return o.LongName != ""
			}),
			func(o *Option) (string, *Option) {
				return o.LongNameWithNamespace(), o
			}),
		commands: ToDict(
			iter.FlatMap(
				iter.FromMany(c.commands...),
				func(c *Command) iter.Seq[fun.Pair[string, *Command]] {
					return iter.Concat(
						iter.FromMany(fun.Pair[string, *Command]{K: c.Name, V: c}),
						iter.Map(
							iter.FromMany(c.Aliases...),
							func(s string) fun.Pair[string, *Command] {
								return fun.Pair[string, *Command]{K: s, V: c}
							}),
					)
				}),
			func(kv fun.Pair[string, *Command]) (string, *Command) {
				return kv.K, kv.V
			}),
	}
}

// ErrNotPointerToStruct indicates that a provided data container is not
// a pointer to a struct. Only pointers to structs are valid data containers
// for options.
var ErrNotPointerToStruct = errors.New("provided data is not a pointer to struct")

// Group represents an option group. Option groups can be used to logically
// group options together under a description. Groups are only used to provide
// more structure to options both for the user (as displayed in the help message)
// and for you, since groups can be nested.
type Group struct {
	// A short description of the group. The
	// short description is primarily used in the built-in generated help
	// message
	ShortDescription string

	// A long description of the group. The long
	// description is primarily used to present information on commands
	// (Command embeds Group) in the built-in generated help and man pages.
	LongDescription string

	// The namespace of the group
	Namespace string

	// The environment namespace of the group
	EnvNamespace string

	// If true, the group is not displayed in the help or man page
	Hidden bool

	// The parent of the group or nil if it has no parent
	parent any

	// All the options in the group
	options []*Option

	// All the subgroups
	groups []*Group

	// Whether the group represents the built-in help group
	isBuiltinHelp bool

	data any
}

type scanHandler func(reflect.Value, *reflect.StructField) (bool, error)

// AddGroup adds a new group to the command with the given name and data. The
// data needs to be a pointer to a struct from which the fields indicate which
// options are in the group.
func (g *Group) AddGroup(shortDescription, longDescription string, data any) (*Group, error) {
	group := &Group{
		ShortDescription: shortDescription,
		LongDescription:  longDescription,
		data:             data,
		parent:           g,
	}

	if err := group.scang(); err != nil {
		return nil, err
	}

	g.groups = append(g.groups, group)
	return group, nil
}

// AddOption adds a new option to this group.
func (g *Group) AddOption(option *Option, data any) {
	option.value = reflect.ValueOf(data)
	option.group = g
	g.options = append(g.options, option)
}

// Find locates the subgroup with the given short description and returns it.
// If no such group can be found Find will return nil.
// Note that the description is matched case insensitively.
func (g *Group) Find(shortDescription string) *Group {
	lshortDescription := strings.ToLower(shortDescription)
	ret, _ := iter.
		Seq[*Group](g.eachGroup).
		Filter(func(gg *Group) bool {
			return gg != g && strings.EqualFold(gg.ShortDescription, lshortDescription)
		}).
		Head()
	return ret
}

func (g *Group) findOption(matcher func(*Option) bool) *Option {
	option, _ := iter.
		FlatMap(iter.
			Seq[*Group](g.eachGroup),
			func(gg *Group) iter.Seq[*Option] {
				return iter.FromMany(gg.options...)
			}).
		Filter(matcher).
		Head()
	return option
}

// findOptionByLongName finds an option that is part of the group, or any of its
// subgroups, by matching its long name (including the option namespace).
func (g *Group) findOptionByLongName(longName string) *Option {
	return g.findOption(func(option *Option) bool {
		return option.LongNameWithNamespace() == longName
	})
}

// findOptionByShortName finds an option that is part of the group, or any of
// its subgroups, by matching its short name.
func (g *Group) findOptionByShortName(shortName rune) *Option {
	return g.findOption(func(option *Option) bool {
		return option.ShortName == shortName
	})
}

func (g *Group) showInHelp() bool {
	return !g.Hidden && fun.Any(g.options, func(opt *Option) bool {
		return opt.showInHelp()
	})
}

func (g *Group) eachGroup(yield func(*Group) bool) bool {
	return yield(g) && fun.All(g.groups, func(gg *Group) bool {
		return gg.eachGroup(yield)
	})
}

func isStringFalsy(s string) bool {
	return s == "" || s == "false" || s == "no" || s == "0"
}

func (g *Group) scanStruct(realval reflect.Value, sfield *reflect.StructField, handler scanHandler) error {
	stype := realval.Type()

	if sfield != nil {
		if ok, err := handler(realval, sfield); err != nil {
			return err
		} else if ok {
			return nil
		}
	}

	for i := 0; i < stype.NumField(); i++ {
		field := stype.Field(i)

		// PkgName is set only for non-exported fields, which we ignore
		if field.PkgPath != "" && !field.Anonymous {
			continue
		}

		mtag, err := ParseMultiTag(string(field.Tag))
		if err != nil {
			return err
		}

		// Skip fields with the no-flag tag
		if mtag.Get("no-flag") != "" {
			continue
		}

		// Dive deep into structs or pointers to structs
		kind := field.Type.Kind()
		fld := realval.Field(i)

		if kind == reflect.Struct {
			if err := g.scanStruct(fld, &field, handler); err != nil {
				return err
			}
		} else if kind == reflect.Ptr && field.Type.Elem().Kind() == reflect.Struct {
			flagCountBefore := len(g.options) + len(g.groups)

			if fld.IsNil() {
				fld = reflect.New(fld.Type().Elem())
			}

			if err := g.scanStruct(reflect.Indirect(fld), &field, handler); err != nil {
				return err
			}

			if len(g.options)+len(g.groups) != flagCountBefore {
				realval.Field(i).Set(fld)
			}
		}

		longname := mtag.Get("long")
		shortname := mtag.Get("short")
		// Need at least either a short or long name
		if longname == "" && shortname == "" {
			continue
		}

		short := rune(0)
		rc := utf8.RuneCountInString(shortname)

		if rc > 1 {
			return newErrorf(ErrShortNameTooLong,
				"short names can only be 1 character long, not `%s'",
				shortname)
		} else if rc == 1 {
			short, _ = utf8.DecodeRuneInString(shortname)
		}

		description := mtag.Get("description")
		def := mtag.GetMany("default")

		optionalValue := mtag.GetMany("optional-value")
		valueName := mtag.Get("value-name")
		defaultMask := mtag.Get("default-mask")

		optional := !isStringFalsy(mtag.Get("optional"))
		required := !isStringFalsy(mtag.Get("required"))
		choices := mtag.GetMany("choice")
		hidden := !isStringFalsy(mtag.Get("hidden"))

		option := &Option{
			Description:      description,
			ShortName:        short,
			LongName:         longname,
			Default:          def,
			OptionalArgument: optional,
			OptionalValue:    optionalValue,
			Required:         required,
			ValueName:        valueName,
			DefaultMask:      defaultMask,
			Choices:          choices,
			Hidden:           hidden,

			group: g,

			field: field,
			value: realval.Field(i),
			tag:   mtag,
		}

		if option.isBool() && option.Default != nil {
			return newErrorf(ErrInvalidTag,
				"boolean flag `%s' may not have default values, they always default to `false' and can only be turned on",
				shortAndLongName(option.ShortName, option.LongName))
		}

		g.options = append(g.options, option)
	}

	return nil
}

func (g *Group) checkForDuplicateFlags() *Error {
	shortNames := make(map[rune]*Option)
	longNames := make(map[string]*Option)
	var duplicateError *Error
	g.eachGroup(func(g *Group) bool {
		for _, option := range g.options {
			if option.LongName != "" {
				longName := option.LongNameWithNamespace()

				if otherOption, ok := longNames[longName]; ok {
					duplicateError = newErrorf(ErrDuplicatedFlag, "option `%s' uses the same long name as option `%s'", option, otherOption)
					return true
				}

				longNames[longName] = option
			}
			if option.ShortName != 0 {
				if otherOption, ok := shortNames[option.ShortName]; ok {
					duplicateError = newErrorf(ErrDuplicatedFlag, "option `%s' uses the same short name as option `%s'", option, otherOption)
					return true
				}

				shortNames[option.ShortName] = option
			}
		}
		return true
	})
	return duplicateError
}

func (g *Group) scanSubGroupHandler(realval reflect.Value, sfield *reflect.StructField) (bool, error) {
	mtag, err := ParseMultiTag(string(sfield.Tag))
	if err != nil {
		return true, err
	}

	subgroup := mtag.Get("group")
	if subgroup == "" {
		return false, nil
	}

	var ptrval reflect.Value
	if realval.Kind() == reflect.Ptr {
		ptrval = realval

		if ptrval.IsNil() {
			ptrval.Set(reflect.New(ptrval.Type()))
		}
	} else {
		ptrval = realval.Addr()
	}

	description := mtag.Get("description")
	group, err := g.AddGroup(subgroup, description, ptrval.Interface())
	if err != nil {
		return true, err
	}

	group.Namespace = mtag.Get("namespace")
	group.EnvNamespace = mtag.Get("env-namespace")
	group.Hidden = mtag.Get("hidden") != ""

	return true, nil
}

func (g *Group) scanType(handler scanHandler) error {
	// Get all the public fields in the data struct
	ptrval := reflect.ValueOf(g.data)

	if ptrval.Type().Kind() != reflect.Ptr {
		panic(ErrNotPointerToStruct)
	}

	stype := ptrval.Type().Elem()

	if stype.Kind() != reflect.Struct {
		panic(ErrNotPointerToStruct)
	}

	realval := reflect.Indirect(ptrval)

	if err := g.scanStruct(realval, nil, handler); err != nil {
		return err
	}

	if err := g.checkForDuplicateFlags(); err != nil {
		return err
	}

	return nil
}

func (g *Group) scang() error {
	return g.scanType(g.scanSubGroupHandler)
}
