package cli

import (
	"context"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/rprtr258/cli/internal"
	"github.com/rprtr258/fun"
	"github.com/rprtr258/fun/set"
)

func scanMultiTag(value string) (map[string][]string, error) {
	ret := map[string][]string{}
	for v := value; v != ""; {
		i := 0

		// Skip whitespace
		for i < len(v) && v[i] == ' ' {
			i++
		}
		v = v[i:]
		if v == "" {
			break
		}

		// Scan to colon to find key
		i = 0
		for i < len(v) && v[i] != ' ' && v[i] != ':' && v[i] != '"' {
			i++
		}
		if i >= len(v) {
			return nil, newErrorf(ErrTag, "expected `:' after key name, but got end of tag (in `%v`)", value)
		}
		if v[i] != ':' {
			return nil, newErrorf(ErrTag, "expected `:' after key name, but got `%v' (in `%v`)", v[i], value)
		}
		if i+1 >= len(v) {
			return nil, newErrorf(ErrTag, "expected `\"' to start tag value at end of tag (in `%v`)", value)
		}
		if v[i+1] != '"' {
			return nil, newErrorf(ErrTag, "expected `\"' to start tag value, but got `%v' (in `%v`)", v[i+1], value)
		}
		name := v[:i]
		v = v[i+1:]

		// Scan quoted string to find value
		i = 1
		for i < len(v) && v[i] != '"' {
			if v[i] == '\n' {
				return nil, newErrorf(ErrTag, "unexpected newline in tag value `%v' (in `%v`)", name, value)
			}

			if v[i] == '\\' {
				i++
			}
			i++
		}
		if i >= len(v) {
			return nil, newErrorf(ErrTag, "expected end of tag value `\"' at end of tag (in `%v`)", value)
		}

		val, err := strconv.Unquote(v[:i+1])
		if err != nil {
			return nil, newErrorf(ErrTag, "Malformed value of tag `%v:%v` => %v (in `%v`)", name, v[:i+1], err, value)
		}

		v = v[i+1:]

		ret[name] = append(ret[name], val)
	}
	return ret, nil
}

func multitagGet(x map[string][]string, key string) string {
	v := x[key]
	if len(v) == 0 {
		return ""
	}

	return v[len(v)-1]
}

// // Marshaler is the interface implemented by types that can marshal themselves
// // to a string representation of the flag.
// type Marshaler interface {
// 	// MarshalFlag marshals a flag value to its string representation.
// 	MarshalFlag() (string, error)
// }

// CustomFlag is the interface implemented by types that can unmarshal a flag argument to themselves.
// The provided value is directly passed from the command line.
type CustomFlag interface {
	// UnmarshalFlag unmarshals a string value representation to the flag value,
	// which therefore needs to be a pointer receiver.
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

// flag information. Contains a description of the flag, short and
// long name as well as a default value and whether an argument for this
// flag is optional.
type flag struct {
	// either shortname or longname must be set
	shortname rune   // optional
	longname  string // optional

	description  fun.Option[string]
	defaultValue []string

	// if isbool, minCount=0, maxCount=1, valuesetter is called w/ value to ignore,
	// otherwise valuesetter called with all flag values
	isbool             bool
	valuesetter        func([]string) error
	minCount, maxCount int

	// set during parsing
	values []string
}

func (f flag) show() string {
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

type positionalArg struct {
	name        string // Name of the positional argument (used in the help)
	description string // Description of the positional argument (used in the help)
	valuesetter func(string) error

	// set during parsing
	value string
}

type positionalArgsRest struct {
	name        string // Name of the positional argument (used in the help)
	description string // Description of the positional argument (used in the help)
	mincount    int    // TODO: move to setter
	valuesetter func([]string) error

	// set during parsing
	values []string
}

type command struct {
	description string
	subcommands map[string]*command // name -> subcommand
	aliases     []string
	flags       []flag
	posArgs     []positionalArg
	posArgsRest fun.Option[positionalArgsRest]
	execute     func(context.Context) error
}

func parseValue(s string, typ reflect.Type) (reflect.Value, error) {
	// standard library types
	if typ.String() == "time.Duration" {
		x, err := time.ParseDuration(s)
		if err != nil {
			return reflect.Value{}, err
		}

		return reflect.ValueOf(x), nil
	}

	// primitive types
	switch typ.Kind() {
	// case typ.Kind() == reflect.Bool:
	// 	value := fun.Contains(s, "yes", "true", "1")
	// 	return reflect.ValueOf(value), nil
	case reflect.String:
		return reflect.ValueOf(s), nil
	// case reflect.Pointer && typ.Elem().Kind() == reflect.String:
	// 	value := reflect.New(typ.Elem())
	// 	value.Elem().SetString(s)
	// 	return value, nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		x, err := strconv.ParseInt(s, 10, 64)
		if err != nil {
			return reflect.Value{}, err
		}

		switch typ.Kind() {
		case reflect.Int:
			return reflect.ValueOf(int(x)), nil
		case reflect.Int8:
			return reflect.ValueOf(int8(x)), nil
		case reflect.Int16:
			return reflect.ValueOf(int16(x)), nil
		case reflect.Int32:
			return reflect.ValueOf(int32(x)), nil
		case reflect.Int64:
			return reflect.ValueOf(int64(x)), nil
		default:
			panic("unreachable")
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		x, err := strconv.ParseUint(s, 10, 64)
		if err != nil {
			return reflect.Value{}, err
		}

		switch typ.Kind() {
		case reflect.Uint:
			return reflect.ValueOf(uint(x)), nil
		case reflect.Uint8:
			return reflect.ValueOf(uint8(x)), nil
		case reflect.Uint16:
			return reflect.ValueOf(uint16(x)), nil
		case reflect.Uint32:
			return reflect.ValueOf(uint32(x)), nil
		case reflect.Uint64:
			return reflect.ValueOf(uint64(x)), nil
		default:
			panic("unreachable")
		}
	case reflect.Float32, reflect.Float64:
		x, err := strconv.ParseFloat(s, 64)
		if err != nil {
			return reflect.Value{}, err
		}

		switch typ.Kind() {
		case reflect.Float32:
			return reflect.ValueOf(float32(x)), nil
		case reflect.Float64:
			return reflect.ValueOf(float64(x)), nil
		default:
			panic("unreachable")
		}
	// case typ.Kind() == reflect.Pointer && fun.Contains(typ.Elem().Kind(), reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64):
	// 	x, err := strconv.ParseInt(s, 10, 64)
	// 	if err != nil {
	// 		return reflect.Value{}, err
	// 	}

	// 	value := reflect.New(typ.Elem())
	// 	value.Elem().SetInt(x)
	// 	return value, nil
	default:
		return reflect.Value{}, fmt.Errorf("can't parse type: %s", typ.String())
	}
}

// TODO: more typed tags
func parseFlag(
	flagT reflect.Type,
	flagV reflect.Value,
	tags map[string][]string,
) (flag, error) {
	nameShort, err := func() (rune, error) {
		raw, ok := tags["short"]
		if !ok {
			return 0, nil
		}

		if len(raw) != 1 {
			return 0, fmt.Errorf("must be specified once, but was %v", raw)
		}

		runes := []rune(raw[0])
		if len(runes) != 1 {
			return 0, fmt.Errorf("must be single character, but was %q", raw[0])
		}

		return runes[0], nil
	}()
	if err != nil {
		return flag{}, fmt.Errorf("scan short name: %w", err)
	}

	nameLong, err := func() (string, error) {
		raw, ok := tags["long"]
		if !ok {
			return "", nil
		}

		if len(raw) != 1 {
			return "", fmt.Errorf("must be specified once, but was %v", raw)
		}

		if ok && len(raw[0]) <= 1 {
			return "", fmt.Errorf("must be at least two characters, but was %q", raw[0])
		}

		return raw[0], nil
	}()
	if err != nil {
		return flag{}, fmt.Errorf("scan long name: %w", err)
	}

	if nameShort == 0 && nameLong == "" {
		return flag{}, fmt.Errorf("flag must have short or long name")
	}

	description, err := func() (fun.Option[string], error) {
		raw, ok := tags["description"]
		if !ok {
			return fun.Option[string]{}, nil
		}

		if len(raw) != 1 {
			return fun.Option[string]{}, fmt.Errorf("must be specified once, but was %v", raw)
		}

		return fun.Valid(raw[0]), nil
	}()
	if err != nil {
		return flag{}, fmt.Errorf("scan description: %w", err)
	}

	defaultValues := tags["default"]
	// in order not to pass default values to subflags
	// TODO: remove
	delete(tags, "default")

	// custom flag
	if customFlag, ok := flagV.Addr().Interface().(CustomFlag); ok {
		return flag{
			shortname:    nameShort,
			longname:     nameLong,
			description:  description,
			defaultValue: defaultValues,
			isbool:       false,
			minCount:     fun.IF(len(defaultValues) > 0, 0, 1),
			maxCount:     1,
			valuesetter: func(s []string) error {
				return customFlag.UnmarshalFlag(s[0])
			},
		}, nil
	}

	switch flagT.Kind() {
	// primitive types
	case reflect.Bool:
		if len(defaultValues) > 0 {
			return flag{}, fmt.Errorf("bool flag cannot have default value, since then it always will be true")
		}

		return flag{
			shortname:    nameShort,
			longname:     nameLong,
			description:  description,
			defaultValue: nil,
			isbool:       true,
			minCount:     0,
			maxCount:     1,
			valuesetter: func(s []string) error {
				flagV.SetBool(true)
				return nil
			},
		}, nil
	case reflect.String:
		if len(defaultValues) > 1 {
			return flag{}, fmt.Errorf("cannot have multiple default values")
		}

		return flag{
			shortname:    nameShort,
			longname:     nameLong,
			description:  description,
			defaultValue: defaultValues,
			isbool:       false,
			minCount:     fun.IF(len(defaultValues) > 0, 0, 1),
			maxCount:     1,
			valuesetter: func(s []string) error {
				flagV.SetString(s[0])
				return nil
			},
		}, nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		if len(defaultValues) > 1 {
			return flag{}, fmt.Errorf("cannot have multiple default values")
		}

		return flag{
			shortname:    nameShort,
			longname:     nameLong,
			description:  description,
			defaultValue: defaultValues,
			isbool:       false,
			minCount:     fun.IF(len(defaultValues) > 0, 0, 1),
			maxCount:     1,
			valuesetter: func(s []string) error {
				x, err := strconv.ParseUint(s[0], 10, 64)
				if err != nil {
					return fmt.Errorf("expected uint, found %q", s[0])
				}

				flagV.SetUint(x)
				return nil
			},
		}, nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if len(defaultValues) > 1 {
			return flag{}, fmt.Errorf("cannot have multiple default values")
		}

		return flag{
			shortname:    nameShort,
			longname:     nameLong,
			description:  description,
			defaultValue: defaultValues,
			isbool:       false,
			minCount:     fun.IF(len(defaultValues) > 0, 0, 1),
			maxCount:     1,
			valuesetter: func(s []string) error {
				value, err := parseValue(s[0], flagT)
				if err != nil {
					return fmt.Errorf("expected int, found %q", s[0])
				}

				flagV.Set(value)
				return nil
			},
		}, nil
	case reflect.Float32, reflect.Float64:
		if len(defaultValues) > 1 {
			return flag{}, fmt.Errorf("cannot have multiple default values")
		}

		return flag{
			shortname:    nameShort,
			longname:     nameLong,
			description:  description,
			defaultValue: defaultValues,
			isbool:       false,
			minCount:     fun.IF(len(defaultValues) > 0, 0, 1),
			maxCount:     1,
			valuesetter: func(s []string) error {
				value, err := parseValue(s[0], flagT)
				if err != nil {
					return fmt.Errorf("expected float, found %q", s[0])
				}

				flagV.Set(value)
				return nil
			},
		}, nil
	// composite types
	case reflect.Pointer:
		if len(defaultValues) > 0 {
			return flag{}, fmt.Errorf("pointer flag cannot have default value, use pointed type directly instead")
		}

		if flagT.Elem().Kind() == reflect.Slice {
			return flag{}, fmt.Errorf("pointer flag cannot be slice, use slice type directly instead")
		}
		if flagT.Elem().Kind() == reflect.Map {
			return flag{}, fmt.Errorf("pointer flag cannot be map, use map type directly instead")
		}

		// TODO: parse subflag once, pass reflect.Value to set into valuesetter
		subval := reflect.New(flagT.Elem())
		subflag, err := parseFlag(flagT.Elem(), subval.Elem(), tags)
		if err != nil {
			return flag{}, err
		}

		return flag{
			shortname:    nameShort,
			longname:     nameLong,
			description:  description,
			defaultValue: nil,
			isbool:       subflag.isbool,
			minCount:     0,
			maxCount:     1,
			valuesetter: func(s []string) error {
				subval := reflect.New(flagT.Elem())
				// TODO: parse subflag once, pass reflect.Value to set into valuesetter
				subflag, err := parseFlag(flagT.Elem(), subval.Elem(), tags)
				if err != nil {
					return err
				}

				if err := subflag.valuesetter(s); err != nil {
					return err
				}

				flagV.Set(subval)
				return nil
			},
		}, nil
	case reflect.Slice:
		return flag{
			shortname:    nameShort,
			longname:     nameLong,
			description:  description,
			defaultValue: defaultValues,
			isbool:       false,
			minCount:     0,
			maxCount:     -1,
			valuesetter: func(ss []string) error {
				vals := make([]reflect.Value, len(ss))
				for i, s := range ss {
					vals[i] = reflect.New(flagT.Elem()).Elem()

					// TODO: parse subflag once, pass reflect.Value to set into valuesetter
					subflag, err := parseFlag(flagT.Elem(), vals[i], tags)
					if err != nil {
						return err
					}

					if err := subflag.valuesetter([]string{s}); err != nil {
						return fmt.Errorf("slice element %d: %w", i, err)
					}
				}
				flagV.Set(reflect.Append(reflect.MakeSlice(flagT, 0, len(ss)), vals...))
				return nil
			},
		}, nil
	case reflect.Array:
		if len(defaultValues) != flagT.Len() {
			return flag{}, fmt.Errorf("expected %d default values, found %d", flagT.Len(), len(defaultValues))
		}

		return flag{
			shortname:    nameShort,
			longname:     nameLong,
			description:  description,
			defaultValue: defaultValues,
			isbool:       false,
			minCount:     flagT.Len(),
			maxCount:     flagT.Len(),
			valuesetter: func(ss []string) error {
				vals := reflect.New(flagT).Elem()
				for i, s := range ss {
					// TODO: parse subflag once, pass reflect.Value to set into valuesetter
					subflag, err := parseFlag(flagT.Elem(), vals.Index(i), tags)
					if err != nil {
						return err
					}

					if err := subflag.valuesetter([]string{s}); err != nil {
						return fmt.Errorf("slice element %d: %w", i, err)
					}
				}
				flagV.Set(vals)
				return nil
			},
		}, nil
	case reflect.Map:
		// TODO: custom delimiter
		return flag{
			shortname:    nameShort,
			longname:     nameLong,
			description:  description,
			defaultValue: defaultValues,
			isbool:       false,
			minCount:     0,
			maxCount:     -1,
			valuesetter: func(s []string) error {
				flagV.Set(reflect.MakeMap(flagT))
				for _, kv := range s {
					kv := strings.SplitN(kv, ":", 2)
					if len(kv) != 2 {
						return fmt.Errorf("invalid key-value pair %q, must be key:value", kv)
					}

					// TODO: parse subflagKey once, pass reflect.Value to set into valuesetter
					valKey := reflect.New(flagT.Key()).Elem()
					subflagKey, err := parseFlag(flagT.Key(), valKey, tags)
					if err != nil {
						return err
					}
					// TODO: parse subflagKey once, pass reflect.Value to set into valuesetter
					valValue := reflect.New(flagT.Elem()).Elem()
					subflagValue, err := parseFlag(flagT.Elem(), valValue, tags)
					if err != nil {
						return err
					}

					if err := subflagKey.valuesetter([]string{kv[0]}); err != nil {
						return fmt.Errorf("invalid key in key-value pair %q: %w", kv, err)
					}

					if err := subflagValue.valuesetter([]string{kv[1]}); err != nil {
						return fmt.Errorf("invalid value in key-value pair %q: %w", kv, err)
					}

					flagV.SetMapIndex(valKey, valValue)
				}
				return nil
			},
		}, nil
	default:
		return flag{}, fmt.Errorf("invalid flag type: %s", flagT.String())
	}
}

func newCommandImpl(tags reflect.StructTag, v reflect.Value) (*command, error) {
	typ := v.Type()
	if typ.Kind() != reflect.Struct {
		return nil, fmt.Errorf("%s is %s, expected struct", typ.String(), typ.Kind().String())
	}

	subcommands := map[string]*command{}
	shortFlagsSeen := set.New[rune](0)
	longFlagsSeen := set.New[string](0)
	flags := []flag{}
	posArgs := []positionalArg{}
	var posArgsRest fun.Option[positionalArgsRest]
	for i := 0; i < typ.NumField(); i++ {
		fieldF := typ.Field(i)
		fieldV := v.Field(i)

		if fieldF.Anonymous {
			fieldF, fieldV = fieldF.Type.Field(0), fieldV.Field(0)
		}

		if !fieldF.IsExported() {
			continue
		}

		switch {
		// command
		case fieldF.Tag.Get("command") != "":
			name := fieldF.Tag.Get("command")
			subcmd, err := newCommandImpl(fieldF.Tag, fieldV)
			if err != nil {
				return nil, fmt.Errorf("parse subcommand %s: %w", name, err)
			}

			subcommands[name] = subcmd
		// positional args
		case fieldF.Tag.Get("positional-args") == "yes":
			switch {
			case fieldF.Type.Kind() == reflect.Slice:
				posArgsRest = fun.Valid(positionalArgsRest{
					name:        "",
					description: "",
					mincount:    0,
					valuesetter: func(ss []string) error {
						vals := make([]reflect.Value, len(ss))
						for i, s := range ss {
							vals[i] = reflect.New(fieldF.Type.Elem()).Elem()

							// TODO: parse subrest once
							subrest, err := parseFlag(fieldF.Type.Elem(), vals[i], map[string][]string{
								"long": {"POSHELNAHUY"},
							})
							if err != nil {
								return err
							}

							if err := subrest.valuesetter([]string{s}); err != nil {
								return fmt.Errorf("slice element %d: %w", i, err)
							}
						}
						fieldV.Set(reflect.Append(reflect.MakeSlice(fieldF.Type, 0, len(ss)), vals...))
						return nil
					},
				})
			case fieldF.Type.Kind() == reflect.Struct:
				for i := 0; i < fieldF.Type.NumField(); i++ {
					argF := fieldF.Type.Field(i)
					if !argF.IsExported() {
						continue
					}

					argT := argF.Type
					argV := fieldV.Field(i)

					if argT.Kind() == reflect.Slice {
						if i != fieldF.Type.NumField()-1 {
							return nil, fmt.Errorf("positional args must be the last field")
						}

						posArgsRest = fun.Valid(positionalArgsRest{
							name:        "",
							description: "",
							mincount:    0,
							valuesetter: func(ss []string) error {
								vals := make([]reflect.Value, len(ss))
								for i, s := range ss {
									vals[i] = reflect.New(argT.Elem()).Elem()

									// TODO: parse subrest once
									subrest, err := parseFlag(argT.Elem(), vals[i], map[string][]string{
										"long": {"POSHELNAHUY"},
									})
									if err != nil {
										return err
									}

									if err := subrest.valuesetter([]string{s}); err != nil {
										return fmt.Errorf("slice element %d: %w", i, err)
									}
								}
								argV.Set(reflect.Append(reflect.MakeSlice(argT, 0, len(ss)), vals...))
								return nil
							},
						})

						break
					}

					name := internal.Or(argF.Tag.Get("name"), argF.Name)
					posArgs = append(posArgs, positionalArg{
						name:        name,
						description: argF.Tag.Get("description"),
						valuesetter: func(s string) error {
							val, err := parseValue(s, argT)
							if err != nil {
								return fmt.Errorf("invalid value %q for %s: %w", s, name, err)
							}

							argV.Set(val)
							return nil
						},
					})
				}
			default:
				return nil, fmt.Errorf("unsupported positional-args field %s with type %s, must be slice or struct", fieldF.Name, fieldF.Type.String())
			}
		// TODO: flag/command groups
		default:
			tags, err := scanMultiTag(string(fieldF.Tag))
			if err != nil {
				return nil, fmt.Errorf("parse flag tags %q: %w", fieldF.Tag, err)
			}

			flagg, err := parseFlag(fieldF.Type, fieldV, tags)
			if err != nil {
				return nil, fmt.Errorf("parse flag %q: %w", fieldF.Name, err)
			}

			if flagg.shortname != 0 && shortFlagsSeen.Contains(flagg.shortname) {
				return nil, newErrorf(ErrDuplicatedFlag, "flag shortname %q is duplicated", flagg.shortname)
			}
			if flagg.longname != "" && longFlagsSeen.Contains(flagg.longname) {
				return nil, newErrorf(ErrDuplicatedFlag, "flag longname %q is duplicated", flagg.longname)
			}

			shortFlagsSeen.Add(flagg.shortname)
			longFlagsSeen.Add(flagg.longname)
			flags = append(flags, flagg)
		}
	}

	var executeFn func(context.Context) error
	if _, ok := v.Interface().(Commander); ok {
		// NOTE: v.Interface intentionally repeated 2 time in order not to copy
		// the value before any setters are called
		executeFn = func(ctx context.Context) error {
			return v.Interface().(Commander).Execute(ctx)
		}
	} else {
		if len(flags) > 0 || len(posArgs) > 0 {
			return nil, fmt.Errorf("command %s has flags but not Execute method defined", typ.String())
		}

		if len(subcommands) == 0 {
			return nil, fmt.Errorf("command %s is leaf command, but no Execute method defined", typ.String())
		}
	}

	mt, err := scanMultiTag(string(tags))
	if err != nil {
		return nil, fmt.Errorf("parse command tags %q: %w", tags, err)
	}

	return &command{
		subcommands: subcommands,
		flags:       flags,
		posArgs:     posArgs,
		posArgsRest: posArgsRest,
		execute:     executeFn,
		description: tags.Get("description"),
		aliases:     mt["alias"],
	}, nil
}

func newCommand[T any]() (*command, error) {
	return newCommandImpl("", reflect.ValueOf(new(T)).Elem())
}
