package flags

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/rprtr258/cli/internal"
	"github.com/rprtr258/fun"
	"github.com/rprtr258/fun/iter"
	"github.com/samber/lo"
)

// Arg represents a positional argument on the command line.
type Arg struct {
	Name            string // Name of the positional argument (used in the help)
	Description     string // Description of the positional argument (used in the help)
	Required        int    // minimal number of required positional arguments
	RequiredMaximum int    // maximum number of required positional arguments

	value reflect.Value
	tag   multiTag
}

func isRemainingArg(value reflect.Value) bool {
	return value.Type().Kind() == reflect.Slice
}

// Marshaler is the interface implemented by types that can marshal themselves
// to a string representation of the flag.
type Marshaler interface {
	// MarshalFlag marshals a flag value to its string representation.
	MarshalFlag() (string, error)
}

// Unmarshaler is the interface implemented by types that can unmarshal a flag argument to themselves.
// The provided value is directly passed from the command line.
type Unmarshaler interface {
	// UnmarshalFlag unmarshals a string value representation to the flag
	// value (which therefore needs to be a pointer receiver).
	UnmarshalFlag(value string) error
}

// ValueValidator is the interface implemented by types that can validate a flag argument themselves.
// The provided value is directly passed from the command line.
type ValueValidator interface {
	// IsValidValue returns an error if the provided string value is valid for the flag
	IsValidValue(value string) error
}

func getBase(base string, defaultBase int) (int, error) {
	if base != "" {
		ivbase, err := strconv.ParseInt(base, 10, 32)
		return int(ivbase), err
	}

	return defaultBase, nil
}

func convertMarshal(val reflect.Value) (val_ bool, ret_ string, err_ error) {
	// Check first for the Marshaler interface
	if val.IsValid() && val.Type().NumMethod() > 0 && val.CanInterface() {
		if marshaler, ok := val.Interface().(Marshaler); ok {
			ret, err := marshaler.MarshalFlag()
			return true, ret, err
		}
	}

	return false, "", nil
}

func convertToString(val reflect.Value, options multiTag) (string, error) {
	if ok, ret, err := convertMarshal(val); ok {
		return ret, err
	}

	if !val.IsValid() {
		return "", nil
	}

	tp := val.Type()

	// Support for time.Duration
	if tp == typeof[time.Duration]() {
		return val.Interface().(fmt.Stringer).String(), nil
	}

	switch tp.Kind() {
	case reflect.String:
		return val.String(), nil
	case reflect.Bool:
		return fun.IF(val.Bool(), "true", "false"), nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		base, err := getBase(options.Get("base"), 10)
		if err != nil {
			return "", err
		}

		return strconv.FormatInt(val.Int(), base), nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		base, err := getBase(options.Get("base"), 10)
		if err != nil {
			return "", err
		}

		return strconv.FormatUint(val.Uint(), base), nil
	case reflect.Float32, reflect.Float64:
		return strconv.FormatFloat(val.Float(), 'g', -1, tp.Bits()), nil
	case reflect.Slice:
		if val.Len() == 0 {
			return "", nil
		}

		ret := make([]string, val.Len())
		for i := 0; i < val.Len(); i++ {
			item, err := convertToString(val.Index(i), options)
			if err != nil {
				return "", err
			}

			ret[i] = item
		}
		return "[" + strings.Join(ret, ", ") + "]", nil
	case reflect.Map:
		ret := make([]string, val.Len())
		for i, key := range val.MapKeys() {
			keyitem, err := convertToString(key, options)
			if err != nil {
				return "", err
			}

			item, err := convertToString(val.MapIndex(key), options)
			if err != nil {
				return "", err
			}

			ret[i] = keyitem + ":" + item
		}
		return "{" + strings.Join(ret, ", ") + "}", nil
	case reflect.Ptr:
		return convertToString(reflect.Indirect(val), options)
	case reflect.Interface:
		if !val.IsNil() {
			return convertToString(val.Elem(), options)
		}
	}

	return "", nil
}

func convertUnmarshal(val string, retval reflect.Value) (bool, error) {
	if retval.Type().NumMethod() > 0 && retval.CanInterface() {
		if unmarshaler, ok := retval.Interface().(Unmarshaler); ok {
			if retval.IsNil() {
				retval.Set(reflect.New(retval.Type().Elem()))
				unmarshaler = retval.Interface().(Unmarshaler) // Re-assign from the new value
			}

			return true, unmarshaler.UnmarshalFlag(val)
		}
	}

	if retval.Type().Kind() != reflect.Ptr && retval.CanAddr() {
		return convertUnmarshal(val, retval.Addr())
	}

	if retval.Type().Kind() == reflect.Interface && !retval.IsNil() {
		return convertUnmarshal(val, retval.Elem())
	}

	return false, nil
}

func convert(val string, retval reflect.Value, multitag multiTag) error {
	if ok, err := convertUnmarshal(val, retval); ok {
		return err
	}

	tp := retval.Type()

	// Support for time.Duration
	if tp == typeof[time.Duration]() {
		parsed, err := time.ParseDuration(val)
		if err != nil {
			return err
		}

		retval.SetInt(int64(parsed))
		return nil
	}

	switch tp.Kind() {
	case reflect.String:
		retval.SetString(val)
	case reflect.Bool:
		if val == "" {
			retval.SetBool(true)
		} else {
			b, err := strconv.ParseBool(val)
			if err != nil {
				return err
			}

			retval.SetBool(b)
		}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		base, err := getBase(multitag.Get("base"), 10)
		if err != nil {
			return err
		}

		parsed, err := strconv.ParseInt(val, base, tp.Bits())
		if err != nil {
			return err
		}

		retval.SetInt(parsed)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		base, err := getBase(multitag.Get("base"), 10)
		if err != nil {
			return err
		}

		parsed, err := strconv.ParseUint(val, base, tp.Bits())
		if err != nil {
			return err
		}

		retval.SetUint(parsed)
	case reflect.Float32, reflect.Float64:
		parsed, err := strconv.ParseFloat(val, tp.Bits())
		if err != nil {
			return err
		}

		retval.SetFloat(parsed)
	case reflect.Slice:
		elemtp := tp.Elem()

		elemvalptr := reflect.New(elemtp)
		elemval := reflect.Indirect(elemvalptr)

		if err := convert(val, elemval, multitag); err != nil {
			return err
		}

		retval.Set(reflect.Append(retval, elemval))
	case reflect.Map:
		parts := strings.SplitN(val, ":", 2)

		var value string
		if len(parts) == 2 {
			value = parts[1]
		}

		keyval := reflect.New(tp.Key())
		key := parts[0]
		if err := convert(key, keyval, multitag); err != nil {
			return err
		}

		valueval := reflect.New(tp.Elem())
		if err := convert(value, valueval, multitag); err != nil {
			return err
		}

		if retval.IsNil() {
			retval.Set(reflect.MakeMap(tp))
		}

		retval.SetMapIndex(reflect.Indirect(keyval), reflect.Indirect(valueval))
	case reflect.Ptr:
		if retval.IsNil() {
			retval.Set(reflect.New(retval.Type().Elem()))
		}

		return convert(val, reflect.Indirect(retval), multitag)
	case reflect.Interface:
		if !retval.IsNil() {
			return convert(val, retval.Elem(), multitag)
		}
	}

	return nil
}

func toDict[K comparable, V, T any](seq iter.Seq[T], f func(T) (K, V)) map[K]V {
	ret := make(map[K]V)
	seq(func(t T) bool {
		k, v := f(t)
		ret[k] = v
		return true
	})
	return ret
}

// Flag information. Contains a description of the flag, short and
// long name as well as a default value and whether an argument for this
// flag is optional.
type Flag struct {
	// The description of the option flag. This description is shown automatically in the built-in help.
	Description string

	// The short name of the option (a single character).
	// If not 0, the option flag can be 'activated' using -<ShortName>.
	// Either ShortName or LongName needs to be non-empty.
	ShortName rune
	// The long name of the option.
	// If not "", the option flag can be activated using --<LongName>.
	// Either ShortName or LongName needs to be non-empty.
	LongName string

	// The default value of the option.
	Default []string

	// If true, specifies that the argument to an option flag is optional.
	// When no argument to the flag is specified on the command line, the
	// value of OptionalValue will be set in the field this option represents.
	// This is only valid for non-boolean options.
	IsOptional bool

	// The optional value of the option.
	// The optional value is used when the option flag is marked as having an OptionalArgument.
	// This means that when the flag is specified, but no option argument is given,
	// the value of the field this option represents will be set to OptionalValue.
	// This is only valid for non-boolean options.
	OptionalValue []string

	// If true, the option _must_ be specified on the command line.
	// If the option is not specified, the parser will generate an ErrRequired type error.
	Required bool

	// A name for the value of an option shown in the Help as --flag [ValueName]
	ValueName string

	// A mask value to show in the help instead of the default value.
	// This is useful for hiding sensitive information in the help, such as passwords.
	DefaultMask string

	// If non empty, only a certain set of values is allowed for an option.
	Choices []string

	// If true, the option is not displayed in the help or man page
	Hidden bool

	// The group which the option belongs to
	group *Group

	// The struct field which the option represents.
	field reflect.StructField

	// The struct field value which the option represents.
	value reflect.Value

	tag                     multiTag
	isSet                   bool
	isSetDefault            bool
	preventDefault          bool
	clearReferenceBeforeSet bool

	defaultLiteral string
}

// LongNameWithNamespace returns the option's long name with the group namespaces
// prepended by walking up the option's group tree. Namespaces and the long name
// itself are separated by the parser's namespace delimiter. If the long name is
// empty an empty string is returned.
func (option *Flag) LongNameWithNamespace() string {
	if option.LongName == "" {
		return ""
	}

	// fetch the namespace delimiter from the parser which is always at the
	// end of the group hierarchy
	var namespaceDelimiter string
	g := option.group
	for {
		if p, ok := g.parent.(*Parser); ok {
			namespaceDelimiter = p.NamespaceDelimiter
			break
		}

		switch i := g.parent.(type) {
		case *Command:
			g = i.Group
		case *Group:
			g = i
		}
	}

	// concatenate long name with namespace
	longName := option.LongName
	for g := option.group; g != nil; g = func() *Group {
		switch i := g.parent.(type) {
		case *Command:
			return i.Group
		case *Group:
			return i
		}
		return nil
	}() {
		if g.Namespace != "" {
			longName = g.Namespace + namespaceDelimiter + longName
		}
	}
	return longName
}

// String converts an option to a human friendly readable string describing the
// option.
func (option *Flag) String() string {
	if option.ShortName != 0 {
		data := make([]byte, utf8.RuneLen(option.ShortName))
		utf8.EncodeRune(data, option.ShortName)
		short := string(data)

		if option.LongName != "" {
			return fmt.Sprintf("%s, %s",
				string(_defaultOptDelimiterShort)+short,
				_defaultOptDelimiterLong+option.LongNameWithNamespace())
		}

		return string(_defaultOptDelimiterShort) + short
	}

	if option.LongName != "" {
		return _defaultOptDelimiterLong + option.LongNameWithNamespace()
	}

	return ""
}

// Value returns the option value as an any.
func (option *Flag) Value() any {
	return option.value.Interface()
}

// Set the value of an option to the specified value. An error will be returned
// if the specified value could not be converted to the corresponding option
// value type.
func (option *Flag) Set(value *string) error {
	kind := option.value.Type().Kind()

	if internal.OneOf(kind, reflect.Map, reflect.Slice) && option.clearReferenceBeforeSet {
		option.empty()
	}

	option.isSet = true
	option.preventDefault = true
	option.clearReferenceBeforeSet = false

	if len(option.Choices) != 0 {
		if fun.All(func(choice string) bool {
			return choice != *value
		}, option.Choices...) {
			allowed := strings.Join(option.Choices[0:len(option.Choices)-1], ", ")

			if len(option.Choices) > 1 {
				allowed += " or " + option.Choices[len(option.Choices)-1]
			}

			return newErrorf(ErrInvalidChoice,
				"Invalid value `%s' for option `%s'. Allowed values are: %s",
				*value, option, allowed)
		}
	}

	switch {
	case option.isFunc():
		return option.call(value)
	case value != nil:
		return convert(*value, option.value, option.tag)
	default:
		return convert("", option.value, option.tag)
	}
}

func (option *Flag) showInHelp() bool {
	return !option.Hidden && (option.ShortName != 0 || option.LongName != "")
}

func (option *Flag) canArgument() bool {
	return isInterface[Unmarshaler](option.value) != nil || !option.isBool()
}

func (option *Flag) empty() {
	if !option.isFunc() {
		tp := option.value.Type()
		zeroValue := fun.
			If(tp.Kind() != reflect.Map, reflect.Zero(tp)).
			ElseF(func() reflect.Value {
				return reflect.MakeMap(tp)
			})
		option.value.Set(zeroValue)
	}
}

func isInterface[I any](value reflect.Value) I {
	for v := value; v.CanInterface(); v = v.Addr() {
		if u, ok := v.Interface().(I); ok {
			return u
		}

		if !v.CanAddr() {
			break
		}
	}

	return fun.Zero[I]()
}

func (option *Flag) isBool() bool {
	for tp := option.value.Type(); ; tp = tp.Elem() {
		if !internal.OneOf(tp.Kind(), reflect.Slice, reflect.Ptr) {
			return tp.Kind() == reflect.Bool ||
				tp.Kind() == reflect.Func && tp.NumIn() == 0
		}
	}
}

func (option *Flag) isSignedNumber() bool {
	for tp := option.value.Type(); ; tp = tp.Elem() {
		if !internal.OneOf(tp.Kind(), reflect.Slice, reflect.Ptr) {
			return internal.OneOf(tp.Kind(),
				reflect.Int,
				reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
				reflect.Float32, reflect.Float64)
		}
	}
}

func (option *Flag) isFunc() bool {
	return option.value.Type().Kind() == reflect.Func
}

func typeof[T any]() reflect.Type {
	return reflect.TypeOf((*T)(nil)).Elem()
}

func (option *Flag) call(value *string) error {
	var retval []reflect.Value
	if value == nil {
		retval = option.value.Call(nil)
	} else {
		val := reflect.Indirect(reflect.New(option.value.Type().In(0)))
		if err := convert(*value, val, option.tag); err != nil {
			return err
		}

		retval = option.value.Call([]reflect.Value{val})
	}

	if len(retval) == 1 && retval[0].Type() == typeof[error]() {
		if retval[0].Interface() == nil {
			return nil
		}

		return retval[0].Interface().(error)
	}

	return nil
}

func (option *Flag) updateDefaultLiteral() string {
	defs := option.Default
	switch {
	case len(defs) == 0 && option.canArgument():
		kind := option.field.Type.Kind()
		if internal.OneOf(kind, reflect.Func, reflect.Ptr) && !option.value.IsNil() ||
			internal.OneOf(kind, reflect.Slice, reflect.String, reflect.Array, reflect.Map) && option.value.Len() > 0 ||
			!reflect.DeepEqual(reflect.Zero(option.field.Type).Interface(), option.value.Interface()) {
			def, _ := convertToString(option.value, option.tag)
			return def
		}

		return ""
	case len(defs) != 0:
		return internal.Join(func(s string) string {
			if !fun.All(strconv.IsPrint, []rune(s)...) {
				return strconv.Quote(s)
			}

			return s
		}, ", ", defs...)
	default:
		return ""
	}
}

func shortAndLongName(shortName rune, longName string) string {
	ret := &bytes.Buffer{}

	if shortName != 0 {
		ret.WriteByte(_defaultOptDelimiterShort)
		ret.WriteRune(shortName)
	}

	if longName != "" {
		if shortName != 0 {
			ret.WriteByte('/')
		}

		ret.WriteString(longName)
	}

	return ret.String()
}

func (option *Flag) isValidValue(arg string) error {
	if validator := isInterface[ValueValidator](option.value); validator != nil {
		return validator.IsValidValue(arg)
	}

	if isFlag(arg) && (!option.isSignedNumber() ||
		len(arg) <= 1 ||
		arg[0] != '-' ||
		arg[1] < '0' || arg[1] > '9') {
		return fmt.Errorf("expected argument for flag `%s', but got option `%s'", option, arg)
	}

	return nil
}

// Command represents an application command. Commands can be added to the
// parser (which itself is a command) and are selected/executed when its name
// is specified on the command line. The Command type embeds a Group and
// therefore also carries a set of command specific options.
type Command struct {
	// Embedded, see Group for more information
	Group *Group

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
	shortNames map[string]*Flag
	longNames  map[string]*Flag
	commands   map[string]*Command
}

// AddCommand adds a new command to the parser with the given name and data. The
// data needs to be a pointer to a struct from which the fields indicate which
// options are in the command. The provided data can implement the Command and
// Usage interfaces.
func (c *Command) AddCommand(command, shortDescription, longDescription string, data any) (*Command, error) {
	cmd := &Command{
		Name: command,
		Group: &Group{
			ShortDescription: shortDescription,
			LongDescription:  longDescription,
			data:             data,
			parent:           c,
		},
	}
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

	c.Group.groups = append(c.Group.groups, group)
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
func (c *Command) FindOptionByLongName(longName string) (option *Flag) {
	for option == nil && c != nil {
		option = c.Group.findOptionByLongName(longName)

		c, _ = c.Group.parent.(*Command)
	}

	return option
}

// FindOptionByShortName finds an option that is part of the command, or any of
// its parent commands, by matching its long name (including the option
// namespace).
func (c *Command) FindOptionByShortName(shortName rune) (option *Flag) {
	for option == nil && c != nil {
		option = c.Group.findOptionByShortName(shortName)

		c, _ = c.Group.parent.(*Command)
	}

	return option
}

// Args returns a list of positional arguments associated with this command.
func (c *Command) Args() []*Arg {
	ret := make([]*Arg, len(c.args))
	copy(ret, c.args)

	return ret
}

func (c *Command) scanSubcommandHandler(parentg *Group) scanHandler {
	return func(realval reflect.Value, sfield *reflect.StructField) (bool, error) {
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

		subc.Group.Hidden = mtag.Get("hidden") != ""

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
}

func (c *Command) scan() error {
	return c.Group.scanType(c.scanSubcommandHandler(c.Group))
}

func (c *Command) eachOption(yield func(*Command, *Group, *Flag)) {
	c.eachCommand(func(c *Command) bool {
		return c.Group.eachGroup(func(g *Group) bool {
			for _, option := range g.options {
				yield(c, g, option)
			}
			return true
		})
	})
}

func (c *Command) eachCommand(yield func(*Command) bool) bool {
	return yield(c) && fun.All(func(c *Command) bool {
		return c.eachCommand(yield)
	}, c.commands...)
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
		return !cmd.Group.Hidden
	}, commands...)
}

func (c *Command) match(name string) bool {
	return c.Name == name || internal.OneOf(name, c.Aliases...)
}

func (c *Command) hasHelpOptions() bool {
	ret := false
	c.Group.eachGroup(func(g *Group) bool {
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

func makeLookupTable(c *Command) lookup {
	var parents []*Command
	for parent := c.Group.parent; parent != nil; {
		if cmd, ok := parent.(*Command); ok {
			parent = cmd.Group.parent
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
				return c.Group.eachGroup
			}),
		func(g *Group) iter.Seq[*Flag] {
			return iter.FromMany(g.options...)
		})

	return lookup{
		shortNames: toDict(
			opts.Filter(func(o *Flag) bool {
				return o.ShortName != 0
			}),
			func(o *Flag) (string, *Flag) {
				return string(o.ShortName), o
			}),
		longNames: toDict(
			opts.Filter(func(o *Flag) bool {
				return o.LongName != ""
			}),
			func(o *Flag) (string, *Flag) {
				return o.LongNameWithNamespace(), o
			}),
		commands: toDict(
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
	// in the built-in generated help and man pages.
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
	options []*Flag

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

	if err := group.scanType(group.scanSubGroupHandler); err != nil {
		return nil, err
	}

	g.groups = append(g.groups, group)
	return group, nil
}

// AddOption adds a new option to this group.
func (g *Group) AddOption(option *Flag, data any) {
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

func (g *Group) findOption(matcher func(*Flag) bool) *Flag {
	option, _ := iter.
		FlatMap(iter.
			Seq[*Group](g.eachGroup),
			func(gg *Group) iter.Seq[*Flag] {
				return iter.FromMany(gg.options...)
			}).
		Filter(matcher).
		Head()
	return option
}

// findOptionByLongName finds an option that is part of the group, or any of its
// subgroups, by matching its long name (including the option namespace).
func (g *Group) findOptionByLongName(longName string) *Flag {
	return g.findOption(func(option *Flag) bool {
		return option.LongNameWithNamespace() == longName
	})
}

// findOptionByShortName finds an option that is part of the group, or any of
// its subgroups, by matching its short name.
func (g *Group) findOptionByShortName(shortName rune) *Flag {
	return g.findOption(func(option *Flag) bool {
		return option.ShortName == shortName
	})
}

func (g *Group) showInHelp() bool {
	return !g.Hidden && fun.Any(func(opt *Flag) bool {
		return opt.showInHelp()
	}, g.options...)
}

func (g *Group) eachGroup(yield func(*Group) bool) bool {
	return yield(g) && fun.All(func(gg *Group) bool {
		return gg.eachGroup(yield)
	}, g.groups...)
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

		option := &Flag{
			Description:   description,
			ShortName:     short,
			LongName:      longname,
			Default:       def,
			IsOptional:    optional,
			OptionalValue: optionalValue,
			Required:      required,
			ValueName:     valueName,
			DefaultMask:   defaultMask,
			Choices:       choices,
			Hidden:        hidden,

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

func (g *Group) checkForDuplicateFlags() error {
	shortNames := make(map[rune]*Flag)
	longNames := make(map[string]*Flag)
	var duplicateError error
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

	if ptrval.Type().Kind() != reflect.Ptr ||
		ptrval.Type().Elem().Kind() != reflect.Struct {
		panic(ErrNotPointerToStruct)
	}

	if err := g.scanStruct(reflect.Indirect(ptrval), nil, handler); err != nil {
		return err
	}

	return g.checkForDuplicateFlags()
}

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

// A Parser provides command line option parsing. It can contain several
// option groups each with their own set of options.
type Parser struct {
	// Embedded, see Command for more information
	Command *Command

	// A usage string to be displayed in the help message.
	Usage string

	// Option flags changing the behavior of the parser.
	Options Options

	// NamespaceDelimiter separates group namespaces and option long names
	NamespaceDelimiter string

	// EnvNamespaceDelimiter separates group env namespaces and env keys
	EnvNamespaceDelimiter string

	// UnknownOptionsHandler is a function which gets called when the parser
	// encounters an unknown option. The function receives the unknown option
	// name, a SplitArgument which specifies its value if set with an argument
	// separator, and the remaining command line arguments.
	// It should return a new list of remaining arguments to continue parsing,
	// or an error to indicate a parse failure.
	UnknownOptionHandler func(option string, arg *string, args []string) ([]string, error)

	// CompletionHandler is a function gets called to handle the completion of items.
	// By default, the items are printed and the application is exited.
	// You can override this default behavior by specifying a custom CompletionHandler.
	CompletionHandler func(items []Completion)

	// CommandHandler is a function that gets called to handle execution of a
	// command. By default, the command will simply be executed. This can be
	// overridden to perform certain actions (such as applying global flags)
	// just before the command is executed. Note that if you override the
	// handler it is your responsibility to call the command.Execute function.
	//
	// The command passed into CommandHandler may be nil in case there is no
	// command to be executed when parsing has finished.
	CommandHandler func(command Commander, args []string) error

	internalError error
}

func (p *Parser) getAlignmentInfo() alignmentInfo {
	ret := alignmentInfo{
		maxLongLen:      0,
		hasShort:        false,
		hasValueName:    false,
		terminalColumns: getTerminalColumns(),
	}
	var prevcmd *Command
	for c := p.Command; c != nil; c = c.Active {
		c.Group.eachGroup(func(grp *Group) bool {
			if c != prevcmd {
				ret.maxLongLen = max(ret.maxLongLen, internal.MaxBy(func(arg *Arg) int {
					return utf8.RuneCountInString(arg.Name) + internal.Iverson(c != p.Command, 4)
				}, c.args...))
				prevcmd = c
			}
			if !grp.showInHelp() {
				return true
			}

			for _, info := range grp.options {
				if !info.showInHelp() {
					continue
				}

				ret.hasShort = ret.hasShort || info.ShortName != 0
				ret.hasValueName = ret.hasValueName || info.ValueName != ""

				l := info.LongNameWithNamespace() + info.ValueName +
					internal.Iverson(len(info.Choices) != 0, "["+strings.Join(info.Choices, "|")+"]")
				ret.maxLongLen = max(ret.maxLongLen, utf8.RuneCountInString(l)+internal.Iverson(c != p.Command, 4))
			}
			return true
		})
	}
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

func (p *Parser) writeHelpOption(w *bytes.Buffer, option *Flag, info alignmentInfo) {
	if option.Hidden {
		return
	}

	line := &bytes.Buffer{}

	prefixLen := _paddingBeforeOption + internal.Iverson(info.indent, 4)
	line.WriteString(strings.Repeat(" ", prefixLen))

	if option.ShortName != 0 {
		line.WriteByte(_defaultOptDelimiterShort)
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

		line.WriteString(_defaultOptDelimiterLong)
		line.WriteString(option.LongNameWithNamespace())
	}

	if option.canArgument() {
		line.WriteByte(_defaultArgDelimiterName)

		if option.ValueName != "" {
			line.WriteString(option.ValueName)
		}

		if len(option.Choices) > 0 {
			line.WriteString("[" + strings.Join(option.Choices, "|") + "]")
		}
	}

	written := line.Len()
	_, _ = line.WriteTo(w)

	if option.Description != "" {
		dw := descstart - written
		_, _ = w.WriteString(strings.Repeat(" ", dw))

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

		_, _ = w.WriteString(wrapText(desc,
			info.terminalColumns-descstart,
			strings.Repeat(" ", descstart)))
	}

	_, _ = w.WriteString("\n")
}

// WriteHelp writes a help message containing all the possible options and
// their descriptions to the provided writer. Note that the HelpFlag parser
// option provides a convenient way to add a -h/--help option group to the
// command line parser which will automatically show the help messages using
// this method.
func (p *Parser) WriteHelp() string {
	aligninfo := p.getAlignmentInfo()

	cmd := p.Command
	for cmd.Active != nil {
		cmd = cmd.Active
	}

	w := &bytes.Buffer{}
	if p.Command.Name != "" {
		_, _ = w.WriteString("Usage:\n ")
		for allcmd := p.Command; allcmd != nil; allcmd = allcmd.Active {
			var usage string
			if allcmd == p.Command {
				if p.Usage != "" {
					usage = p.Usage
				} else if p.Options&HelpFlag != 0 {
					usage = "[OPTIONS]"
				}
			} else if us, ok := allcmd.Group.data.(Usage); ok {
				usage = us.Usage()
			} else if allcmd.hasHelpOptions() {
				usage = fmt.Sprintf("[%s-OPTIONS]", allcmd.Name)
			}

			fmt.Fprint(w, " "+allcmd.Name)
			if usage != "" {
				fmt.Fprint(w, " "+usage)
			}

			if len(allcmd.args) > 0 {
				fmt.Fprintf(w, " ")
			}

			for i, arg := range allcmd.args {
				if i != 0 {
					fmt.Fprintf(w, " ")
				}

				name := arg.Name + internal.Iverson(isRemainingArg(arg.value), "...")
				if !allcmd.ArgsRequired && arg.Required <= 0 {
					name = "[" + name + "]"
				}

				fmt.Fprint(w, name)
			}

			if allcmd.Active == nil && len(allcmd.commands) > 0 {
				scommands := "command"
				if visibleCommands := visibleCommands(allcmd.commands); len(visibleCommands) <= 3 {
					scommands = strings.Join(internal.Map(func(subc *Command) string {
						return subc.Name
					}, visibleCommands...), " | ")
				}

				co, cc := "<", ">"
				if allcmd.SubcommandsOptional {
					co, cc = "[", "]"
				}

				fmt.Fprint(w, " "+co+scommands+cc)
			}
		}

		fmt.Fprintln(w)

		if cmd.Group.LongDescription != "" {
			fmt.Fprintln(w)
			fmt.Fprintln(w, wrapText(cmd.Group.LongDescription, aligninfo.terminalColumns, ""))
		}
	}

	for c := p.Command; c != nil; c = c.Active {
		printcmd := c != p.Command

		c.Group.eachGroup(func(grp *Group) bool {
			// Skip built-in help group for all commands except the top-level parser
			if grp.Hidden || grp.isBuiltinHelp && c != p.Command {
				return true
			}

			first := true
			for _, info := range grp.options {
				if !info.showInHelp() {
					continue
				}

				if printcmd {
					fmt.Fprintf(w, "\n[%s command options]\n", c.Name)
					aligninfo.indent = true
					printcmd = false
				}

				if first && cmd.Group != grp {
					fmt.Fprintln(w)

					if aligninfo.indent {
						_, _ = w.WriteString("    ")
					}

					fmt.Fprint(w, grp.ShortDescription+":\n")
					first = false
				}

				p.writeHelpOption(w, info, aligninfo)
			}
			return true
		})

		if args := internal.Filter(func(arg *Arg) bool {
			// TODO: separate Hidden field
			return arg.Description != ""
		}, c.args...); len(args) > 0 {
			sarguments := "Arguments:"
			if c != p.Command {
				sarguments = fmt.Sprintf("[%s command arguments]", c.Name)
			}
			fmt.Fprint(w, "\n"+sarguments+"\n")

			descStart := aligninfo.descriptionStart() + _paddingBeforeOption
			for _, arg := range args {
				argPrefix := strings.Repeat(" ", _paddingBeforeOption) + arg.Name
				if arg.Description != "" {
					argPrefix += ":"
					_, _ = w.WriteString(argPrefix)

					// Space between "arg:" and the description start
					_, _ = w.WriteString(strings.Repeat(" ", descStart-len(argPrefix)))
					_, _ = w.WriteString(wrapText(
						arg.Description,
						// How much space the description gets before wrapping
						aligninfo.terminalColumns-1-descStart,
						// Whitespace to which we can indent new description lines
						strings.Repeat(" ", descStart),
					))
				} else {
					_, _ = w.WriteString(argPrefix)
				}

				fmt.Fprintln(w)
			}
		}
	}

	if scommands := visibleCommands(cmd.commands); len(scommands) > 0 {
		fmt.Fprintln(w)
		fmt.Fprintln(w, "Available commands:")

		maxnamelen := internal.MaxBy(func(v *Command) int {
			aliasesTotalLen := 0
			for _, alias := range v.Aliases {
				aliasesTotalLen += len(alias)
			}
			return len(v.Name) + len(v.Aliases) + aliasesTotalLen
		}, scommands...)

		for _, c := range scommands {
			fullNameRepresentation := c.Name + internal.Iverson(len(c.Aliases) > 0, ","+strings.Join(c.Aliases, ","))
			fmt.Fprint(w, "  "+fullNameRepresentation)

			if c.Group.ShortDescription != "" {
				pad := strings.Repeat(" ", maxnamelen-len(fullNameRepresentation))
				fmt.Fprintf(w, "%s  %s", pad, c.Group.ShortDescription)
			}

			fmt.Fprintln(w)
		}
	}

	return w.String()
}

// Options provides parser options that change the behavior of the option parser
type Options uint

const (
	// None indicates no options.
	None Options = 0

	// HelpFlag adds a default Help Options group to the parser containing
	// -h and --help options. When either -h or --help is specified on the
	// command line, the parser will return the special error of type
	// ErrHelp. When PrintErrors is also specified, then the help message
	// will also be automatically printed to os.Stdout.
	HelpFlag = 1 << iota

	// PassDoubleDash passes all arguments after a double dash, --, as
	// remaining command line arguments
	// (i.e. they will not be parsed for flags)
	PassDoubleDash

	// IgnoreUnknown ignores any unknown options and passes them as
	// remaining command line arguments instead of generating an error.
	IgnoreUnknown

	// PassAfterNonOption passes all arguments after the first non option
	// as remaining command line arguments.
	// This is equivalent to strict POSIX processing.
	PassAfterNonOption

	// Default is a convenient default set of options which should cover
	// most of the uses of the flags package.
	Default = HelpFlag | PassDoubleDash
)

type parseState struct {
	arg        string
	args       []string
	retargs    []string
	positional []*Arg
	err        error

	command *Command
	lookup  lookup
}

// NewParser creates a new parser. It uses os.Args[0] as the application
// name and then calls Parser.NewNamedParser (see Parser.NewNamedParser for
// more details). The provided data is a pointer to a struct representing the
// default option group (named "Application Options"), or nil if the default
// group should not be added. The options parameter specifies a set of options
// for the parser.
func NewParser(data any, options Options) *Parser {
	p := NewNamedParser(filepath.Base(os.Args[0]), options)

	if data != nil {
		g, err := p.Command.AddGroup("Application Options", "", data)
		if err == nil {
			g.parent = p
		}

		p.internalError = err
	}

	return p
}

// NewNamedParser creates a new parser. The appname is used to display the
// executable name in the built-in help message. Option groups and commands can
// be added to this parser by using AddGroup and AddCommand.
func NewNamedParser(appname string, options Options) *Parser {
	p := &Parser{
		Command: &Command{
			Name: appname,
			Group: &Group{
				ShortDescription: "",
				LongDescription:  "",
				data:             nil,
			},
		},
		Options:               options,
		NamespaceDelimiter:    ".",
		EnvNamespaceDelimiter: "_",
	}
	p.Command.Group.parent = p
	return p
}

// Run parses the command line arguments according to the option groups that
// were added to the parser. On successful parsing of the arguments, the
// remaining, non-option, arguments (if any) are returned. The returned error
// indicates a parsing error and can be used with PrintError to display
// contextual information on where the error occurred exactly.
//
// When the common help group has been added (AddHelp) and either -h or --help
// was specified in the command line arguments, a help message will be
// automatically printed if the PrintErrors option is enabled.
// Furthermore, the special error type ErrHelp is returned.
// It is up to the caller to exit the program if so desired.
func (p *Parser) Run(args ...string) ([]string, error) {
	if p.internalError != nil {
		return nil, p.internalError
	}

	p.Command.eachOption(func(c *Command, g *Group, option *Flag) {
		option.clearReferenceBeforeSet = true
		option.defaultLiteral = option.updateDefaultLiteral()
	})

	// Add built-in help group to all commands if necessary
	if p.Options&HelpFlag != 0 {
		p.Command.addHelpGroups(p.showBuiltinHelp)
	}

	if _, ok := os.LookupEnv("GO_FLAGS_COMPLETION"); ok {
		comp := &completion{parser: p}

		if p.CompletionHandler == nil {
			p.CompletionHandler = func(items []Completion) {
				comp.print(os.Stdout, items)
			}
		}

		p.CompletionHandler(comp.complete(args))
		return nil, nil
	}

	s := &parseState{
		args:       args,
		retargs:    make([]string, 0, len(args)),
		positional: p.Command.args,
		lookup:     makeLookupTable(p.Command),
		command:    p.Command,
	}
	for !s.noargs() {
		arg := s.pop()

		// When PassDoubleDash is set and we encounter a --, then
		// simply append all the rest as arguments and break out
		if p.Options&PassDoubleDash != 0 && arg == "--" {
			_ = s.addArgs(s.args...)
			break
		}

		if !isFlag(arg) {
			if p.Options&PassAfterNonOption != 0 && s.lookup.commands[arg] == nil {
				// If PassAfterNonOption is set then all remaining arguments are considered positional
				if !s.addArgs(s.arg) || !s.addArgs(s.args...) {
					break
				}

				break
			}

			// Note: this also sets s.err, so we can just check for nil here and use s.err later
			if !p.parseNonOption(s) {
				break
			}

			continue
		}

		prefix, optname, islong := stripFlagPrefix(arg)
		optname, argument := splitFlag(prefix, optname, islong)

		if err := fun.IF(islong, p.parseLong, p.parseShort)(s, optname, argument); err != nil {
			ignoreUnknown := (p.Options & IgnoreUnknown) != None

			if parseErr := wrapError(err); parseErr.Kind != ErrUnknownFlag || !ignoreUnknown && p.UnknownOptionHandler == nil {
				s.err = parseErr
				break
			}

			if ignoreUnknown {
				_ = s.addArgs(arg)
			} else if p.UnknownOptionHandler != nil {
				modifiedArgs, err := p.UnknownOptionHandler(optname, argument, s.args)
				if err != nil {
					s.err = err
					break
				}

				s.args = modifiedArgs
			}
		}
	}

	if s.err == nil {
		p.Command.eachOption(func(c *Command, g *Group, option *Flag) {
			if option.preventDefault {
				return
			}

			option.isSetDefault = true

			if len(option.Default) > 0 {
				option.empty()

				for _, d := range option.Default {
					if option.preventDefault {
						break
					}

					if err := option.Set(&d); err != nil {
						if err != nil {
							if _, ok := err.(*Error); !ok {
								err = marshalError(option, err)
							}
							s.err = err
						}
						break
					}

					option.isSetDefault = true
					option.preventDefault = false
				}
			} else if internal.OneOf(option.value.Type().Kind(), reflect.Map, reflect.Slice) && option.value.IsNil() {
				option.empty()
			}
		})

		s.checkRequired(p)
	}

	if reterr := s.reterr(p.CommandHandler); reterr != nil {
		return func() []string {
			if ourErr, ok := reterr.(*Error); !ok || ourErr.Kind != ErrHelp {
				return append([]string{s.arg}, s.args...)
			} else {
				return s.args
			}
		}(), reterr
	}

	return s.retargs, nil
}

func (p *parseState) noargs() bool {
	return len(p.args) == 0
}

func (p *parseState) pop() string {
	if p.noargs() {
		return ""
	}

	p.arg = p.args[0]
	p.args = p.args[1:]

	return p.arg
}

func (s *parseState) checkRequired(p *Parser) {
	var required []*Flag
	for c := p.Command; c != nil; c = c.Active {
		c.Group.eachGroup(func(g *Group) bool {
			for _, option := range g.options {
				if !option.isSet && option.Required {
					required = append(required, option)
				}
			}
			return true
		})
	}

	if len(required) == 0 {
		reqnames := iter.MapFilter(iter.FromMany(s.positional...),
			func(arg *Arg) (string, bool) {
				var desc string
				switch {
				case (isRemainingArg(arg.value) || !s.command.ArgsRequired) && arg.Required == -1 && arg.RequiredMaximum == -1:
					return "", false
				case !isRemainingArg(arg.value):
					desc = ""
				case arg.value.Len() < arg.Required:
					desc = " (at least " + fmt.Sprint(arg.Required) + " argument" + internal.Iverson(arg.Required > 1, "s, but got only "+fmt.Sprint(arg.value.Len())) + ")"
				case arg.RequiredMaximum != -1 && arg.value.Len() > arg.RequiredMaximum:
					desc = lo.If(arg.RequiredMaximum == 0, " (zero arguments)").ElseF(func() string {
						return " (at most " + fmt.Sprint(arg.RequiredMaximum) + " argument" + internal.Iverson(arg.RequiredMaximum > 1, "s, but got "+fmt.Sprint(arg.value.Len())) + ")"
					})
				default:
					return "", false
				}
				return "`" + arg.Name + desc + "`", true
			}).
			ToSlice()
		if len(reqnames) != 0 {
			s.err = &Error{
				Kind: ErrRequired,
				Message: func() string {
					if len(reqnames) == 1 {
						return fmt.Sprintf("the required argument %s was not provided", reqnames[0])
					} else {
						return fmt.Sprintf("the required arguments %s and %s were not provided",
							strings.Join(reqnames[:len(reqnames)-1], ", "), reqnames[len(reqnames)-1])
					}
				}(),
			}
		}

		return
	}

	names := internal.Map(func(k *Flag) string {
		return "`" + k.String() + "'"
	}, required...)
	sort.Strings(names)

	s.err = &Error{
		Kind: ErrRequired,
		Message: func() string {
			if len(required) == 1 {
				return fmt.Sprintf("the required flag %s was not specified", names[0])
			} else {
				return fmt.Sprintf("the required flags %s and %s were not specified",
					strings.Join(names[:len(names)-1], ", "), names[len(names)-1])
			}
		}(),
	}
}

func (p *parseState) reterr(commandHandler func(Commander, []string) error) error {
	if p.err != nil {
		return p.err
	}

	if len(p.command.commands) != 0 && !p.command.SubcommandsOptional {
		cmdnames := internal.Map(func(v *Command) string {
			return v.Name
		}, visibleCommands(p.command.commands)...)

		if len(p.retargs) != 0 {
			return newUnknownCommandError(p.retargs[0], cmdnames)
		}

		return newCommandRequiredError(cmdnames)
	}

	if cmd, ok := p.command.Group.data.(Commander); ok {
		if commandHandler != nil {
			return commandHandler(cmd, p.retargs)
		}

		return cmd.Execute(p.retargs)
	}

	if commandHandler != nil {
		return commandHandler(nil, p.retargs)
	}

	return nil
}

func (p *Parser) parseOption(s *parseState, name string, option *Flag, canarg bool, argument *string) (err error) {
	switch {
	case !option.canArgument():
		if argument != nil {
			return newErrorf(ErrNoArgumentForBool, "bool flag `%s' cannot have an argument", option)
		}

		err = option.Set(nil)
	case argument != nil || canarg && !s.noargs():
		var arg string
		if argument != nil {
			arg = *argument
		} else {
			arg = s.pop()

			if validationErr := option.isValidValue(arg); validationErr != nil {
				return newErrorf(ErrExpectedArgument, validationErr.Error())
			} else if p.Options&PassDoubleDash != 0 && arg == "--" {
				return newErrorf(ErrExpectedArgument, "expected argument for flag `%s', but got double dash `--'", option)
			}
		}

		if option.tag.Get("unquote") != "false" {
			arg, err = func(s string) (string, error) {
				if s == "" || s[0] != '"' {
					return s, nil
				}

				return strconv.Unquote(s)
			}(arg)
		}

		if err == nil {
			err = option.Set(&arg)
		}
	case option.IsOptional:
		option.empty()

		for _, v := range option.OptionalValue {
			if err = option.Set(&v); err != nil {
				break
			}
		}
	default:
		err = newErrorf(ErrExpectedArgument, "expected argument for flag `%s'", option)
	}

	if err != nil {
		if _, ok := err.(*Error); !ok {
			err = marshalError(option, err)
		}
	}

	return err
}

func (p *Parser) parseLong(s *parseState, name string, argument *string) error {
	option := s.lookup.longNames[name]
	if option == nil {
		return newErrorf(ErrUnknownFlag, "unknown flag `%s'", name)
	}

	// Only long options that are required can consume an argument from the argument list
	return p.parseOption(s, name, option, !option.IsOptional, argument)
}

func (p *Parser) splitShortConcatArg(s *parseState, optname string) (optname_ string, arg_ *string) {
	c, n := utf8.DecodeRuneInString(optname)
	if n == len(optname) {
		return optname, nil
	}

	first := string(c)
	if option := s.lookup.shortNames[first]; option != nil && option.canArgument() {
		arg := optname[n:]
		return first, &arg
	}

	return optname, nil
}

func (p *Parser) parseShort(s *parseState, optname string, argument *string) error {
	if argument == nil {
		optname, argument = p.splitShortConcatArg(s, optname)
	}

	for i, c := range optname {
		shortname := string(c)
		option := s.lookup.shortNames[shortname]
		if option == nil {
			return newErrorf(ErrUnknownFlag, "unknown flag `%s'", shortname)
		}

		// Only the last short argument can consume an argument from
		// the arguments list, and only if it's non optional
		canarg := i+utf8.RuneLen(c) == len(optname) && !option.IsOptional
		if err := p.parseOption(s, shortname, option, canarg, argument); err != nil {
			return err
		}

		// Only the first option can have a concatted argument, so just
		// clear argument here
		argument = nil
	}

	return nil
}

func (p *parseState) addArgs(args ...string) bool {
	for len(p.positional) > 0 && len(args) > 0 {
		arg := p.positional[0]
		if err := convert(args[0], arg.value, arg.tag); err != nil {
			p.err = err
			return false
		}

		if !isRemainingArg(arg.value) {
			p.positional = p.positional[1:]
		}

		args = args[1:]
	}

	p.retargs = append(p.retargs, args...)
	return true
}

func (p *Parser) parseNonOption(s *parseState) bool {
	if len(s.positional) > 0 || len(s.command.commands) == 0 || len(s.retargs) > 0 {
		return s.addArgs(s.arg)
	}

	if cmd := s.lookup.commands[s.arg]; cmd != nil {
		s.command.Active = cmd
		s.positional = cmd.args
		s.lookup = makeLookupTable(cmd)
		s.command = cmd
		return true
	}

	if !s.command.SubcommandsOptional {
		_ = s.addArgs(s.arg)
		return false
	}

	return s.addArgs(s.arg)
}

func (p *Parser) showBuiltinHelp() error {
	return &Error{
		Kind:    ErrHelp,
		Message: p.WriteHelp(),
	}
}
