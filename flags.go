package cli

import (
	"context"
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/rprtr258/cli/internal"
	"github.com/rprtr258/fun"
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

// Commander is an interface which can be implemented by any command added in
// the options. When implemented, the Execute method will be called for the last
// specified (sub)command providing the remaining command line arguments.
type Commander interface {
	// Execute will be called for the last active (sub)command.
	// The args argument contains the remaining command line arguments.
	// The error that Execute returns will be eventually passed out of the
	// Parse method of the Parser.
	Execute(context.Context) error
}

// Flag information. Contains a description of the flag, short and
// long name as well as a default value and whether an argument for this
// flag is optional.
type Flag struct {
	// either shortname or longname must be set
	shortname rune   // optional
	longname  string // optional

	// if isbool, minCount=0, maxCount=1, valuesetter is called w/ value to ignore,
	// otherwise valuesetter called with all flag values
	isbool             bool
	valuesetter        func([]string) error
	minCount, maxCount int

	// set during parsing
	values []string
}

func (f Flag) show() string {
	switch {
	case f.shortname != 0 && f.longname != "":
		return fmt.Sprintf("-%c/--%s", f.shortname, f.longname)
	case f.shortname != 0:
		return fmt.Sprintf("-%c", f.shortname)
	case f.longname != "":
		return fmt.Sprintf("--%s", f.longname)
	default:
		return ""
	}
}

type PositionalArg struct {
	name        string // Name of the positional argument (used in the help)
	description string // Description of the positional argument (used in the help)
	isrequired  bool
	valuesetter func(string) error

	// set during parsing
	value string
}

type PositionalArgsRest struct {
	name        string // Name of the positional argument (used in the help)
	description string // Description of the positional argument (used in the help)
	mincount    int    // TODO: move to setter
	defined     bool
	valuesetter func([]string) error

	// set during parsing
	values []string
}

type Command struct {
	subcommands map[string]*Command // name,aliases... -> subcommand
	flags       []Flag
	posArgs     []PositionalArg
	posArgsRest PositionalArgsRest
	execute     func(context.Context) error
}

func parseValue(s string, typ reflect.Type) (reflect.Value, error) {
	switch {
	case typ.Kind() == reflect.Bool:
		value := fun.Contains(s, "yes", "true", "1")
		return reflect.ValueOf(value), nil
	case typ.Kind() == reflect.String:
		return reflect.ValueOf(s), nil
	case typ.Kind() == reflect.Pointer && typ.Elem().Kind() == reflect.String:
		value := reflect.New(typ.Elem())
		value.Elem().SetString(s)
		return value, nil
	case fun.Contains(typ.Kind(), reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64):
		x, err := strconv.ParseInt(s, 10, 64)
		if err != nil {
			return reflect.Value{}, err
		}

		return reflect.ValueOf(x), nil
	case typ.Kind() == reflect.Ptr && fun.Contains(typ.Elem().Kind(), reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64):
		x, err := strconv.ParseInt(s, 10, 64)
		if err != nil {
			return reflect.Value{}, err
		}

		value := reflect.New(typ.Elem())
		value.Elem().SetInt(x)
		return value, nil
	default:
		return reflect.Value{}, fmt.Errorf("can't parse type: %s", typ.String())
	}
}

func newCommandImpl(v reflect.Value) (*Command, error) {
	typ := v.Type()
	if typ.Kind() != reflect.Struct {
		return nil, fmt.Errorf("%s is %s, expected struct", typ.String(), typ.Kind().String())
	}

	subcommands := map[string]*Command{}
	flags := []Flag{}
	posArgs := []PositionalArg{}
	var posArgsRest PositionalArgsRest
	for i := 0; i < typ.NumField(); i++ {
		fieldT := typ.Field(i)
		fieldV := v.Field(i)

		if !fieldT.IsExported() {
			continue
		}

		switch {
		// flag
		case fieldT.Tag.Get("short") != "" || fieldT.Tag.Get("long") != "":
			shortname, err := func() (rune, error) {
				raw, ok := fieldT.Tag.Lookup("short")
				if !ok {
					return 0, nil
				}

				runes := []rune(raw)
				if len(runes) != 1 {
					return 0, fmt.Errorf("short flag must be single character, but was %s", raw)
				}

				return runes[0], nil
			}()
			if err != nil {
				return nil, err
			}

			longname, ok := fieldT.Tag.Lookup("long")
			if ok && len(longname) <= 1 {
				return nil, fmt.Errorf("long flag must be at least two characters, but was %s", fieldT.Tag.Get("long"))
			}

			var flag Flag
			switch {
			case fieldT.Type.Kind() == reflect.Bool: // TODO: or *bool or []bool
				flag = Flag{
					shortname: shortname,
					longname:  longname,
					isbool:    true,
					minCount:  0,
					maxCount:  1,
					valuesetter: func(s []string) error {
						fieldV.SetBool(true)
						return nil
					},
				}
			case fieldT.Type.Kind() == reflect.String: // TODO: or *string
				flag = Flag{
					shortname: shortname,
					longname:  longname,
					isbool:    false,
					minCount:  1,
					maxCount:  1,
					valuesetter: func(s []string) error {
						fieldV.SetString(s[0])
						return nil
					},
				}
			case fun.Contains(fieldT.Type.Kind(), reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64): // TODO: or pointer to
				flag = Flag{
					shortname: shortname,
					longname:  longname,
					isbool:    false,
					minCount:  1,
					maxCount:  1,
					valuesetter: func(s []string) error {
						x, err := strconv.ParseUint(s[0], 10, 64)
						if err != nil {
							return fmt.Errorf("expected uint, found %q", s[0])
						}

						fieldV.SetUint(x)
						return nil
					},
				}
			case fun.Contains(fieldT.Type.Kind(), reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64) ||
				fieldT.Type.Kind() == reflect.Ptr && fun.Contains(fieldT.Type.Elem().Kind(), reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64):
				flag = Flag{
					shortname: shortname,
					longname:  longname,
					isbool:    false,
					minCount:  1,
					maxCount:  1,
					valuesetter: func(s []string) error {
						value, err := parseValue(s[0], fieldT.Type)
						if err != nil {
							return fmt.Errorf("expected int, found %q", s[0])
						}

						fieldV.Set(value)
						return nil
					},
				}
			case fieldT.Type.Kind() == reflect.Slice &&
				(fieldT.Type.Elem().Kind() == reflect.String ||
					fieldT.Type.Elem().Kind() == reflect.Pointer && fieldT.Type.Elem().Elem().Kind() == reflect.String):
				flag = Flag{
					shortname: shortname,
					longname:  longname,
					isbool:    false,
					minCount:  0,
					maxCount:  -1,
					valuesetter: func(s []string) error {
						if fieldT.Type.Elem().Kind() == reflect.Pointer {
							fieldV.Set(reflect.ValueOf(fun.Map[*string](fun.Ptr[string], s...)))
						} else {
							fieldV.Set(reflect.ValueOf(s))
						}
						return nil
					},
				}
			case fieldT.Type.Kind() == reflect.Map:
				flag = Flag{
					shortname: shortname,
					longname:  longname,
					isbool:    false,
					minCount:  0,
					maxCount:  -1,
					valuesetter: func(s []string) error {
						fieldV.Set(reflect.MakeMap(fieldT.Type))
						for _, kv := range s {
							kv := strings.SplitN(kv, ":", 2)
							if len(kv) != 2 {
								return fmt.Errorf("invalid key-value pair %q, must be key:value", kv)
							}

							key, err := parseValue(kv[0], fieldT.Type.Key())
							if err != nil {
								return fmt.Errorf("invalid key in key-value pair %q: %w", kv, err)
							}

							value, err := parseValue(kv[1], fieldT.Type.Elem())
							if err != nil {
								return fmt.Errorf("invalid value in key-value pair %q: %w", kv, err)
							}

							fieldV.SetMapIndex(key, value)
						}
						return nil
					},
				}
			default:
				return nil, fmt.Errorf("unknown field %s type: %s", fieldT.Name, fieldT.Type.String())
			}
			flags = append(flags, flag)
		// command
		case fieldT.Tag.Get("command") != "":
			name := fieldT.Tag.Get("command")
			aliases := []string{}
			if tagAliases, ok := fieldT.Tag.Lookup("aliases"); ok {
				aliases = strings.Split(tagAliases, ",")
			}

			subcmd, err := newCommandImpl(fieldV)
			if err != nil {
				return nil, fmt.Errorf("parse subcommand %s: %w", name, err)
			}

			subcommands[name] = subcmd
			for _, alias := range aliases {
				subcommands[alias] = subcmd
			}
		// positional args
		case fieldT.Tag.Get("positional-args") == "yes":
			switch {
			case fieldT.Type.Kind() == reflect.Slice && fieldT.Type.Elem().Kind() == reflect.String:
				posArgsRest = PositionalArgsRest{
					name:        "",
					description: "",
					mincount:    0,
					defined:     true,
					valuesetter: func(s []string) error {
						fieldV.Set(reflect.ValueOf(s))
						return nil
					},
				}
			case fieldT.Type.Kind() == reflect.Struct:
				posArgsRest = PositionalArgsRest{
					name:        "",
					description: "",
					mincount:    0,
					defined:     true,
					valuesetter: func(s []string) error {
						for i := 0; i < fieldT.Type.NumField(); i++ {
							fieldV.Set(reflect.ValueOf(s))
						}
						return nil
					},
				}
			default:
				return nil, fmt.Errorf("unsupported positional-args field %s with type %s, must be slice or struct", fieldT.Name, fieldT.Type.String())
			}
		default:
			return nil, fmt.Errorf("unsupported field %s with tags %q", fieldT.Name, string(fieldT.Tag))
		}
	}

	var executeFn func(context.Context) error
	if _, ok := v.Interface().(Commander); ok {
		// NOTE: v.Interface intentionally repeated 2 time in order not to copy
		// the value before any setters are called
		executeFn = func(ctx context.Context) error {
			return v.Interface().(Commander).Execute(ctx)
		}
	}

	if executeFn == nil {
		if len(flags) > 0 || len(posArgs) > 0 {
			return nil, fmt.Errorf("command %s has flags but not Execute function defined", typ.String())
		}

		if len(subcommands) == 0 {
			return nil, fmt.Errorf("command %s is leaf command, but no Execute function defined", typ.String())
		}
	}

	return &Command{
		subcommands: subcommands,
		flags:       flags,
		posArgs:     posArgs,
		posArgsRest: posArgsRest,
		execute:     executeFn,
	}, nil
}

func newCommand[T any]() (*Command, error) {
	return newCommandImpl(reflect.ValueOf(new(T)).Elem())
}

// Completion is a type containing information of a completion.
type Completion struct {
	Item        string // The completed item
	Description string // Description of the completed item (optional)
}

func (c *Command) complete(args []string) []Completion {
	return nil
}

func (p *Command) printCompletions(args ...string) {
	items := p.complete(args)
	maxl := internal.MaxBy(func(v Completion) int {
		return len(v.Item)
	}, items...)

	for _, v := range items {
		fmt.Printf(v.Item)
		// TODO: problems with colored description
		// TODO: description get into completion wtf
		if v.Description != "" {
			fmt.Printf("%s  # %s", strings.Repeat(" ", maxl-len(v.Item)), v.Description)
		}
		fmt.Println()
	}
}
