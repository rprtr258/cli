package flags

import (
	"reflect"
	"strconv"
	"strings"

	"github.com/rprtr258/cli/internal"
	"github.com/rprtr258/fun"
	"github.com/rprtr258/fun/iter"
)

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

func (c *Command) eachActiveGroup(yield func(cc *Command, g *Group)) {
	c.eachGroup(func(g *Group) bool {
		yield(c, g)
		return true
	})

	if c.Active != nil {
		c.Active.eachActiveGroup(yield)
	}
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

func (c *Command) makeLookup() lookup {
	var parents []*Command
	for parent := c.parent; parent != nil; {
		cmd, ok := parent.(*Command)
		if !ok {
			break
		}

		parents = append(parents, cmd)
		parent = cmd.parent
	}

	ret := lookup{
		shortNames: make(map[string]*Option),
		longNames:  make(map[string]*Option),
		commands:   make(map[string]*Command),
	}
	for i := len(parents) - 1; i >= 0; i-- {
		parents[i].eachGroup(fillLookup(ret))
	}
	c.eachGroup(fillLookup(ret))
	for _, subcommand := range c.commands {
		ret.commands[subcommand.Name] = subcommand
		for _, a := range subcommand.Aliases {
			ret.commands[a] = subcommand
		}
	}
	return ret
}

func fillLookup(ret lookup) func(g *Group) bool {
	return func(g *Group) bool {
		iter.
			FromMany(g.options...).
			Filter(func(o *Option) bool {
				return o.ShortName != 0
			}).
			ForEach(func(o *Option) {
				ret.shortNames[string(o.ShortName)] = o
			})
		iter.
			FromMany(g.options...).
			Filter(func(o *Option) bool {
				return o.LongName != ""
			}).
			ForEach(func(o *Option) {
				ret.longNames[o.LongNameWithNamespace()] = o
			})
		return true
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

func fillParseState(c *Command) ([]*Arg, lookup, *Command) {
	lookup := c.makeLookup()
	return append([]*Arg(nil), c.args...), lookup, c
}
