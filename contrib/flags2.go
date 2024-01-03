package flags

import (
	"context"
	"fmt"
	"os"
	"reflect"
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

func newCommandImpl(v reflect.Value) (*Command, error) {
	typ := v.Type()
	if typ.Kind() != reflect.Struct {
		return nil, fmt.Errorf("%s is %s, expected struct", typ.String(), typ.Kind().String())
	}

	subcommands := map[string]*Command{}
	flags := []Flag{}
	posArgs := []PositionalArg{}

	for i := 0; i < typ.NumField(); i++ {
		fieldT := typ.Field(i)
		fieldV := v.Field(i)

		if !fieldT.IsExported() {
			continue
		}

		switch {
		// flag
		case fieldT.Tag.Get("short") != "" || fieldT.Tag.Get("long") != "":
			shortname, ok := fieldT.Tag.Lookup("short")
			shortnameRunes := []rune(shortname)
			if ok && len(shortnameRunes) != 1 {
				return nil, fmt.Errorf("short flag must be single character, but was %s", fieldT.Tag.Get("short"))
			}

			longname, ok := fieldT.Tag.Lookup("long")
			if ok && len(longname) <= 1 {
				return nil, fmt.Errorf("long flag must be at least two characters, but was %s", fieldT.Tag.Get("long"))
			}

			var flag Flag
			switch fieldT.Type.Kind() {
			case reflect.Bool: // TODO: or *bool or []bool
				flag = Flag{
					shortname: shortnameRunes[0],
					longname:  longname,
					isbool:    true,
					minCount:  0,
					maxCount:  1,
					valuesetter: func(s []string) error {
						fieldV.SetBool(true)
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
		default:
			return nil, fmt.Errorf("unsupported field %s", fieldT.Name)
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
			return nil, fmt.Errorf("command %s has no subcommands, but no Execute function defined", typ.String())
		}
	}

	return &Command{
		subcommands: subcommands,
		flags:       flags,
		posArgs:     posArgs,
		posArgsRest: PositionalArgsRest{
			name:        "",
			description: "",
			mincount:    0,
			valuesetter: func([]string) error {
				return nil
			},
		},
		execute: executeFn,
	}, nil
}

func newCommand[T any]() (*Command, error) {
	return newCommandImpl(reflect.ValueOf(new(T)).Elem())
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

func Run[T any](args ...string) error {
	return RunContext[T](context.Background(), args...)
}

// RunContext parses the command line arguments according to the data.
// Error indicates parsing error.
//
// When the common help group has been added (AddHelp) and either -h or --help
// was specified in the command line arguments, the special error type ErrHelp
// is returned with help text to print.
func RunContext[T any](ctx context.Context, args ...string) error {
	cmd, err := newCommand[T]()
	if err != nil {
		return err
	}

	// add help flags to each subcommand
	var showHelp bool
	cmd.flags = append(cmd.flags, Flag{
		shortname: 'h',
		longname:  "help",
		isbool:    true,
		minCount:  0,
		maxCount:  1,
		valuesetter: func([]string) error {
			showHelp = true
			return nil
		},
	})

	args = args[1:] // throw off executable name

	// check if we should just print completions, then exit
	if _, ok := os.LookupEnv("GO_FLAGS_COMPLETION"); ok {
		cmd.printCompletions(args...)
		return nil
	}

	// find appropriate subcommand
	for ; len(args) > 0; args = args[1:] {
		if subcommand, ok := cmd.subcommands[args[0]]; ok {
			cmd = subcommand
		} else {
			break
		}
	}

	if cmd.execute == nil {
		return &Error{
			Kind: ErrCommandRequired,
			// TODO: squash subcommands aliases, print suggestion based on distance
			Message: "no command specified, expected one of: " + strings.Join(fun.Keys(cmd.subcommands), ", "),
		}
	}

	// parse arguments, including flags and positional arguments, which means
	// filling all value(s) fields

	// parse flags
	for ; len(args) > 0; args = args[1:] {
		arg := args[0]

		// not a flag, start parsing positionals
		if !isFlag(arg) {
			break
		}

		// explicit start of positionals
		if arg == "--" {
			args = args[1:]
			break
		}

		curFlag, i, ok := fun.Index(func(f Flag) bool {
			return arg == "-"+string(f.shortname) ||
				arg == "--"+f.longname
		}, cmd.flags...)
		if !ok { // unknown flag, sounds more like first positional arg
			break
		}

		if curFlag.isbool {
			cmd.flags[i].values = append(cmd.flags[i].values, "")
		} else if len(args) > 1 {
			args = args[1:]
			cmd.flags[i].values = append(cmd.flags[i].values, args[0])
		} else {
			return &Error{
				Kind:    ErrExpectedArgument,
				Message: fmt.Sprintf("expected argument for flag %s", curFlag.longname),
			}
		}
	}

	// parse positional args
	if len(args) < len(cmd.posArgs) && cmd.posArgs[len(args)].isrequired {
		return &Error{ // TODO: separate error types
			Kind:    ErrExpectedArgument,
			Message: fmt.Sprintf("positional argument %s is required", cmd.posArgs[len(args)].name),
		}
	}

	for i, arg := range fun.Subslice(0, len(cmd.posArgs), args...) {
		cmd.posArgs[i].value = arg
	}

	if len(args) >= len(cmd.posArgs) {
		cmd.posArgsRest.values = args[len(cmd.posArgs):]
	}

	// filled arguments values, now time to call all setters and validate all of that

	for _, f := range cmd.flags {
		if f.isbool {
			if len(f.values) > 1 {
				return &Error{
					Kind:    ErrDuplicatedFlag,
					Message: fmt.Sprintf("bool flag %#v specified too many times, expected at most 1", f),
				}
			}

			if len(f.values) == 1 {
				if err := f.valuesetter(f.values); err != nil {
					return err
				}
			}
		} else {
			if len(f.values) < f.minCount {
				return &Error{
					Kind:    ErrExpectedArgument,
					Message: fmt.Sprintf("too few arguments for flag %#v, minimum %d required", f, f.minCount),
				}
			}

			if len(f.values) > f.maxCount {
				return &Error{
					Kind:    ErrExpectedArgument,
					Message: fmt.Sprintf("too many arguments for flag %#v, maximum %d expected", f, f.maxCount),
				}
			}

			if f.values != nil {
				if err := f.valuesetter(f.values); err != nil {
					return err
				}
			}
		}
	}

	// TODO: use even in presence of errors in other flags
	if showHelp {
		return &Error{
			Kind:    ErrHelp,
			Message: cmd.writeHelp(),
		}
	}

	return cmd.execute(ctx)
}
