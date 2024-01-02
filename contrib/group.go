package flags

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/rprtr258/cli/internal"
	"github.com/rprtr258/fun"
	"github.com/rprtr258/fun/iter"
)

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
	if sbase := base; sbase != "" {
		ivbase, err := strconv.ParseInt(sbase, 10, 32)
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
		if val.Bool() {
			return "true", nil
		}

		return "false", nil
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

		ret := "["

		for i := 0; i < val.Len(); i++ {
			if i != 0 {
				ret += ", "
			}

			item, err := convertToString(val.Index(i), options)
			if err != nil {
				return "", err
			}

			ret += item
		}

		return ret + "]", nil
	case reflect.Map:
		ret := "{"

		for i, key := range val.MapKeys() {
			if i != 0 {
				ret += ", "
			}

			keyitem, err := convertToString(key, options)
			if err != nil {
				return "", err
			}

			item, err := convertToString(val.MapIndex(key), options)
			if err != nil {
				return "", err
			}

			ret += keyitem + ":" + item
		}

		return ret + "}", nil
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

				// Re-assign from the new value
				unmarshaler = retval.Interface().(Unmarshaler)
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

func isPrint(s string) bool {
	for _, c := range s {
		if !strconv.IsPrint(c) {
			return false
		}
	}

	return true
}

func quoteIfNeeded(s string) string {
	if !isPrint(s) {
		return strconv.Quote(s)
	}

	return s
}

func unquoteIfPossible(s string) (string, error) {
	if s == "" || s[0] != '"' {
		return s, nil
	}

	return strconv.Unquote(s)
}

func ToDict[K comparable, V, T any](seq iter.Seq[T], f func(T) (K, V)) map[K]V {
	ret := make(map[K]V)
	seq(func(t T) bool {
		k, v := f(t)
		ret[k] = v
		return true
	})
	return ret
}

// Option flag information. Contains a description of the option, short and
// long name as well as a default value and whether an argument for this
// flag is optional.
type Option struct {
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
	OptionalArgument bool

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
func (option *Option) LongNameWithNamespace() string {
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
func (option *Option) String() string {
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
func (option *Option) Value() any {
	return option.value.Interface()
}

// Set the value of an option to the specified value. An error will be returned
// if the specified value could not be converted to the corresponding option
// value type.
func (option *Option) Set(value *string) error {
	kind := option.value.Type().Kind()

	if internal.OneOf(kind, reflect.Map, reflect.Slice) && option.clearReferenceBeforeSet {
		option.empty()
	}

	option.isSet = true
	option.preventDefault = true
	option.clearReferenceBeforeSet = false

	if len(option.Choices) != 0 {
		if fun.All(option.Choices, func(choice string) bool {
			return choice != *value
		}) {
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

func (option *Option) setDefault(value *string) error {
	if option.preventDefault {
		return nil
	}

	if err := option.Set(value); err != nil {
		return err
	}

	option.isSetDefault = true
	option.preventDefault = false

	return nil
}

func (option *Option) showInHelp() bool {
	return !option.Hidden && (option.ShortName != 0 || option.LongName != "")
}

func (option *Option) canArgument() bool {
	return isInterface[Unmarshaler](option.value) != nil || !option.isBool()
}

func (option *Option) empty() {
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

func (option *Option) clearDefault() error {
	if option.preventDefault {
		return nil
	}

	option.isSetDefault = true

	if len(option.Default) > 0 {
		option.empty()

		for _, d := range option.Default {
			err := option.setDefault(&d)
			if err != nil {
				return err
			}
		}
	} else if internal.OneOf(option.value.Type().Kind(), reflect.Map, reflect.Slice) && option.value.IsNil() {
		option.empty()
	}

	return nil
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

func (option *Option) isBool() bool {
	for tp := option.value.Type(); ; tp = tp.Elem() {
		if !internal.OneOf(tp.Kind(), reflect.Slice, reflect.Ptr) {
			return tp.Kind() == reflect.Bool ||
				tp.Kind() == reflect.Func && tp.NumIn() == 0
		}
	}
}

func (option *Option) isSignedNumber() bool {
	for tp := option.value.Type(); ; tp = tp.Elem() {
		if !internal.OneOf(tp.Kind(), reflect.Slice, reflect.Ptr) {
			return internal.OneOf(tp.Kind(),
				reflect.Int,
				reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
				reflect.Float32, reflect.Float64)
		}
	}
}

func (option *Option) isFunc() bool {
	return option.value.Type().Kind() == reflect.Func
}

func typeof[T any]() reflect.Type {
	return reflect.TypeOf((*T)(nil)).Elem()
}

func (option *Option) call(value *string) error {
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

func (option *Option) updateDefaultLiteral() string {
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
		return internal.Join(quoteIfNeeded, ", ", defs...)
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

func (option *Option) isValidValue(arg string) error {
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

func (g *Group) scang() error {
	return g.scanType(g.scanSubGroupHandler)
}

// A Parser provides command line option parsing. It can contain several
// option groups each with their own set of options.
type Parser struct {
	// Embedded, see Command for more information
	*Command

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
	UnknownOptionHandler func(option string, arg SplitArgument, args []string) ([]string, error)

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
		c.eachGroup(func(grp *Group) bool {
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

func (p *Parser) writeHelpOption(writer *bufio.Writer, option *Option, info alignmentInfo) {
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

				name := arg.Name + internal.Iverson(isRemainingArg(arg.value), "...")
				if !allcmd.ArgsRequired && arg.Required <= 0 {
					name = "[" + name + "]"
				}

				fmt.Fprint(bw, name)
			}

			if allcmd.Active == nil && len(allcmd.commands) > 0 {
				var co, cc string
				if allcmd.SubcommandsOptional {
					co, cc = "[", "]"
				} else {
					co, cc = "<", ">"
				}

				if visibleCommandss := visibleCommands(allcmd.commands); len(visibleCommandss) > 3 {
					fmt.Fprint(bw, " "+co+"command"+cc)
				} else {
					fmt.Fprint(bw, " "+co+strings.Join(internal.Map(func(subc *Command) string {
						return subc.Name
					}, visibleCommands(allcmd.commands)...), " | ")+cc)
				}
			}
		}

		fmt.Fprintln(bw)

		if cmd.LongDescription != "" {
			fmt.Fprintln(bw)
			fmt.Fprintln(bw, wrapText(cmd.LongDescription, aligninfo.terminalColumns, ""))
		}
	}

	for c := p.Command; c != nil; c = c.Active {
		printcmd := c != p.Command

		c.eachGroup(func(grp *Group) bool {
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
			return true
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

		maxnamelen := internal.MaxBy(func(v *Command) int {
			aliasesTotalLen := 0
			for _, alias := range v.Aliases {
				aliasesTotalLen += len(alias)
			}
			return len(v.Name) + len(v.Aliases) + aliasesTotalLen
		}, scommands...)

		for _, c := range scommands {
			fullNameRepresentation := c.Name + internal.Iverson(len(c.Aliases) > 0, ","+strings.Join(c.Aliases, ","))
			fmt.Fprint(bw, "  "+fullNameRepresentation)

			if c.ShortDescription != "" {
				pad := strings.Repeat(" ", maxnamelen-len(fullNameRepresentation))
				fmt.Fprintf(bw, "%s  %s", pad, c.ShortDescription)
			}

			fmt.Fprintln(bw)
		}
	}
}

// SplitArgument represents the argument value of an option that was passed using
// an argument separator.
type SplitArgument interface {
	// String returns the option's value as a string, and a boolean indicating
	// if the option was present.
	Value() (string, bool)
}

type strArgument struct {
	value *string
}

func (s strArgument) Value() (string, bool) {
	if s.value == nil {
		return "", false
	}

	return *s.value, true
}

// Options provides parser options that change the behavior of the option
// parser.
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
	// remaining command line arguments (i.e. they will not be parsed for
	// flags).
	PassDoubleDash

	// IgnoreUnknown ignores any unknown options and passes them as
	// remaining command line arguments instead of generating an error.
	IgnoreUnknown

	// PrintErrors prints any errors which occurred during parsing to
	// os.Stderr. In the special case of ErrHelp, the message will be printed
	// to os.Stdout.
	PrintErrors

	// PassAfterNonOption passes all arguments after the first non option
	// as remaining command line arguments. This is equivalent to strict
	// POSIX processing.
	PassAfterNonOption

	// Default is a convenient default set of options which should cover
	// most of the uses of the flags package.
	Default = HelpFlag | PrintErrors | PassDoubleDash
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

// Parse is a convenience function to parse command line options with default
// settings. The provided data is a pointer to a struct representing the
// default option group (named "Application Options"). For more control, use
// flags.NewParser.
func Parse(data any) ([]string, error) {
	return NewParser(data, Default).Parse()
}

// ParseArgs is a convenience function to parse command line options with default
// settings. The provided data is a pointer to a struct representing the
// default option group (named "Application Options"). The args argument is
// the list of command line arguments to parse. If you just want to parse the
// default program command line arguments (i.e. os.Args), then use flags.Parse
// instead. For more control, use flags.NewParser.
func ParseArgs(data any, args ...string) ([]string, error) {
	return NewParser(data, Default).ParseArgs(args...)
}

// NewParser creates a new parser. It uses os.Args[0] as the application
// name and then calls Parser.NewNamedParser (see Parser.NewNamedParser for
// more details). The provided data is a pointer to a struct representing the
// default option group (named "Application Options"), or nil if the default
// group should not be added. The options parameter specifies a set of options
// for the parser.
func NewParser(data any, options Options) *Parser {
	p := NewNamedParser(path.Base(os.Args[0]), options)

	if data != nil {
		g, err := p.AddGroup("Application Options", "", data)

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
		Command:               newCommand(appname, "", "", nil),
		Options:               options,
		NamespaceDelimiter:    ".",
		EnvNamespaceDelimiter: "_",
	}

	p.Command.parent = p

	return p
}

// Parse parses the command line arguments from os.Args using Parser.ParseArgs.
// For more detailed information see ParseArgs.
func (p *Parser) Parse() ([]string, error) {
	return p.ParseArgs(os.Args[1:]...)
}

// ParseArgs parses the command line arguments according to the option groups that
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
func (p *Parser) ParseArgs(args ...string) ([]string, error) {
	if p.internalError != nil {
		return nil, p.internalError
	}

	p.eachOption(func(c *Command, g *Group, option *Option) {
		option.clearReferenceBeforeSet = true
		option.defaultLiteral = option.updateDefaultLiteral()
	})

	// Add built-in help group to all commands if necessary
	if p.Options&HelpFlag != 0 {
		p.addHelpGroups(p.showBuiltinHelp)
	}

	if _, ok := os.LookupEnv("GO_FLAGS_COMPLETION"); ok {
		comp := &completion{parser: p}
		items := comp.complete(args)

		if p.CompletionHandler == nil {
			p.CompletionHandler = func(items []Completion) {
				comp.print(os.Stdout, items)
			}
		}

		p.CompletionHandler(items)
		return nil, nil
	}

	s := &parseState{
		args:       args,
		retargs:    make([]string, 0, len(args)),
		positional: p.Command.args,
		lookup:     fillParseState(p.Command),
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
				// If PassAfterNonOption is set then all remaining arguments
				// are considered positional
				if s.addArgs(s.arg) != nil {
					break
				}

				if s.addArgs(s.args...) != nil {
					break
				}

				break
			}

			// Note: this also sets s.err, so we can just check for
			// nil here and use s.err later
			if p.parseNonOption(s) != nil {
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
				modifiedArgs, err := p.UnknownOptionHandler(optname, strArgument{argument}, s.args)
				if err != nil {
					s.err = err
					break
				}

				s.args = modifiedArgs
			}
		}
	}

	if s.err == nil {
		p.eachOption(func(c *Command, g *Group, option *Option) {
			err := option.clearDefault()
			if err != nil {
				if _, ok := err.(*Error); !ok {
					err = p.marshalError(option, err)
				}
				s.err = err
			}
		})

		s.checkRequired(p)
	}

	reterr := func() error {
		switch {
		case s.err != nil:
			return s.err
		case len(s.command.commands) != 0 && !s.command.SubcommandsOptional:
			return s.estimateCommand()
		default:
			if cmd, ok := s.command.data.(Commander); ok {
				if p.CommandHandler != nil {
					return p.CommandHandler(cmd, s.retargs)
				}

				return cmd.Execute(s.retargs)
			}

			if p.CommandHandler != nil {
				return p.CommandHandler(nil, s.retargs)
			}

			return nil
		}
	}()
	if reterr != nil {
		return func() []string {
			if ourErr, ok := reterr.(*Error); !ok || ourErr.Kind != ErrHelp {
				return append([]string{s.arg}, s.args...)
			} else {
				return s.args
			}
		}(), p.printError(reterr)
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

func (p *parseState) checkRequired(parser *Parser) {
	var required []*Option
	for c := parser.Command; c != nil; c = c.Active {
		c.eachGroup(func(g *Group) bool {
			for _, option := range g.options {
				if !option.isSet && option.Required {
					required = append(required, option)
				}
			}
			return true
		})
	}

	if len(required) == 0 {
		if len(p.positional) == 0 {
			return
		}

		var reqnames []string
		for _, arg := range p.positional {
			if argRequired := !isRemainingArg(arg.value) && p.command.ArgsRequired || arg.Required != -1 || arg.RequiredMaximum != -1; !argRequired {
				continue
			}

			if isRemainingArg(arg.value) {
				if arg.value.Len() < arg.Required {
					reqnames = append(reqnames, "`"+arg.Name+" (at least "+fmt.Sprint(arg.Required)+" argument"+internal.Iverson(arg.Required > 1, "s, but got only "+fmt.Sprint(arg.value.Len()))+")`")
				} else if arg.RequiredMaximum != -1 && arg.value.Len() > arg.RequiredMaximum {
					reqnames = append(reqnames, func() string {
						if arg.RequiredMaximum == 0 {
							return "`" + arg.Name + " (zero arguments)`"
						} else {
							return "`" + arg.Name + " (at most " + fmt.Sprint(arg.RequiredMaximum) + " argument" + internal.Iverson(arg.RequiredMaximum > 1, "s, but got "+fmt.Sprint(arg.value.Len())) + ")`"
						}
					}())
				}
			} else {
				reqnames = append(reqnames, "`"+arg.Name+"`")
			}
		}
		if len(reqnames) == 0 {
			return
		}

		p.err = &Error{
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

		return
	}

	names := internal.Map(func(k *Option) string {
		return "`" + k.String() + "'"
	}, required...)
	sort.Strings(names)

	p.err = &Error{
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

func (p *parseState) estimateCommand() error {
	cmdnames := internal.Map(func(v *Command) string {
		return v.Name
	}, visibleCommands(p.command.commands)...)

	if len(p.retargs) != 0 {
		c, l := getClosest(p.retargs[0], cmdnames)
		return &Error{
			Kind: ErrUnknownCommand,
			Message: fmt.Sprintf("Unknown command `%s'", p.retargs[0]) + func() string {
				switch {
				case l*2 < len(c):
					return fmt.Sprintf(", did you mean `%s'?", c)
				case len(cmdnames) == 1:
					return fmt.Sprintf(". You should use the %s command", cmdnames[0])
				case len(cmdnames) > 1:
					return fmt.Sprintf(". Please specify one command of: %s or %s",
						strings.Join(cmdnames[:len(cmdnames)-1], ", "),
						cmdnames[len(cmdnames)-1])
				default:
					return ""
				}
			}(),
		}
	}

	return &Error{
		Kind: ErrCommandRequired,
		Message: func() string {
			switch {
			case len(cmdnames) == 1:
				return fmt.Sprintf("Please specify the %s command", cmdnames[0])
			case len(cmdnames) > 1:
				return fmt.Sprintf("Please specify one command of: %s or %s",
					strings.Join(cmdnames[:len(cmdnames)-1], ", "),
					cmdnames[len(cmdnames)-1])
			default:
				return ""
			}
		}(),
	}
}

func (p *Parser) parseOption(s *parseState, name string, option *Option, canarg bool, argument *string) (err error) {
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
			arg, err = unquoteIfPossible(arg)
		}

		if err == nil {
			err = option.Set(&arg)
		}
	case option.OptionalArgument:
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
			err = p.marshalError(option, err)
		}
	}

	return err
}

func (p *Parser) marshalError(option *Option, err error) *Error {
	valueType := option.value.Type()
	return &Error{
		Kind: ErrMarshal,
		Message: "invalid argument for flag `" + option.String() + "'" +
			internal.Iverson(valueType.Kind() != reflect.Func, " (expected "+valueType.String()+")") +
			": " + err.Error(),
	}
}

func (p *Parser) parseLong(s *parseState, name string, argument *string) error {
	option := s.lookup.longNames[name]
	if option == nil {
		return newErrorf(ErrUnknownFlag, "unknown flag `%s'", name)
	}

	// Only long options that are required can consume an argument from the argument list
	return p.parseOption(s, name, option, !option.OptionalArgument, argument)
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
		canarg := i+utf8.RuneLen(c) == len(optname) && !option.OptionalArgument
		if err := p.parseOption(s, shortname, option, canarg, argument); err != nil {
			return err
		}

		// Only the first option can have a concatted argument, so just
		// clear argument here
		argument = nil
	}

	return nil
}

func (p *parseState) addArgs(args ...string) error {
	for len(p.positional) > 0 && len(args) > 0 {
		arg := p.positional[0]
		if err := convert(args[0], arg.value, arg.tag); err != nil {
			p.err = err
			return err
		}

		if !isRemainingArg(arg.value) {
			p.positional = p.positional[1:]
		}

		args = args[1:]
	}

	p.retargs = append(p.retargs, args...)
	return nil
}

func (p *Parser) parseNonOption(s *parseState) error {
	if len(s.positional) > 0 {
		return s.addArgs(s.arg)
	}

	if len(s.command.commands) > 0 && len(s.retargs) == 0 {
		if cmd := s.lookup.commands[s.arg]; cmd != nil {
			s.command.Active = cmd
			s.positional = cmd.args
			s.lookup = fillParseState(cmd)
			s.command = cmd
			return nil
		} else if !s.command.SubcommandsOptional {
			_ = s.addArgs(s.arg)
			return newErrorf(ErrUnknownCommand, "Unknown command `%s'", s.arg)
		}
	}

	return s.addArgs(s.arg)
}

func (p *Parser) showBuiltinHelp() error {
	var b bytes.Buffer
	p.WriteHelp(&b)
	return &Error{
		Kind:    ErrHelp,
		Message: b.String(),
	}
}

func (p *Parser) printError(err error) error {
	if err != nil && p.Options&PrintErrors != 0 {
		fmt.Fprintln(os.Stderr, err)
	}
	return err
}
