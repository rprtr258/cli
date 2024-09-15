package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"

	"github.com/rprtr258/fun"

	"github.com/rprtr258/cli/internal"
)

// Run parses the command line arguments according to the type.
// Error indicates either parsing or execution error.
// TODO: add root command description
func Run[T any](args ...string) error {
	origcmd, err := newCommand[T]()
	if err != nil {
		// TODO: return some ParseError type
		return err
	}

	cmdname := args[0]
	args = args[1:] // throw off executable name

	// find appropriate subcommand
	subcommands := []string{filepath.Base(cmdname)}
	cmd := origcmd
	for ; len(args) > 0; args = args[1:] {
		for k, subc := range cmd.subcommands {
			if k == args[0] || fun.Contains(args[0], subc.aliases...) {
				cmd = subc
				subcommands = append(subcommands, args[0])
				goto CONTINUE
			}
		}
		break
	CONTINUE:
	}

	// add help flags to each subcommand
	cmd.flags = append(cmd.flags, flag{
		shortname:    'h',
		longname:     "help",
		description:  fun.Optional("Print this help message", true),
		defaultValue: nil,
		isbool:       true,
		minCount:     0,
		maxCount:     1,
		valuesetter: func([]string) error {
			return nil
		},
	})

	// check if we should just print completions, then exit
	if _, ok := os.LookupEnv("GO_FLAGS_COMPLETION"); ok {
		cmd.printCompletions(args...)
		return nil
	}

	// parse arguments, including flags and positional arguments, which means filling all value(s) fields

	// parse flags
	// do not return flag error immediately since we can meet help flag which must be used instead
	var flagErr error
	for ; len(args) > 0; args = args[1:] {
		arg := args[0]

		if arg == "-h" || arg == "--help" {
			return newHelpError(cmd, subcommands)
		}

		// explicit start of positionals
		if arg == "--" {
			args = args[1:]
			break
		}

		// not a flag, start parsing positionals
		isFlag := len(arg) > 1 && arg[0] == '-' && arg[1] != '-' ||
			len(arg) > 2 && arg[0] == '-' && arg[1] == '-' && arg[2] != '-'
		if !isFlag {
			break
		}

		curFlag, i, ok := fun.Index(func(f flag) bool {
			return arg == "-"+string(f.shortname) ||
				arg == "--"+f.longname
		}, cmd.flags...)
		if !ok { // unknown flag, sounds more like first positional arg
			break
		}

		switch {
		case curFlag.isbool:
			cmd.flags[i].values = append(cmd.flags[i].values, "")
		case len(args) > 1:
			args = args[1:]
			cmd.flags[i].values = append(cmd.flags[i].values, args[0])
		default:
			flagErr = &Error{
				Kind:    ErrExpectedArgument,
				Message: fmt.Sprintf("expected argument for flag %s", curFlag.show()),
			}
		}
	}
	if flagErr != nil {
		// help flag not met, return flag error
		return flagErr
	}

	if cmd.execute == nil {
		keys := make([]string, 0, len(cmd.subcommands))
		for k, subc := range cmd.subcommands {
			keys = append(keys, k)
			keys = append(keys, subc.aliases...)
		}
		sort.Strings(keys)
		if len(args) > 0 {
			return newUnknownCommandError(cmd, subcommands, args[0], keys)
		}

		return newCommandRequiredError(cmd, subcommands, keys)
	}

	// parse positional args
	if len(args) < len(cmd.posArgs) {
		return &Error{ // TODO: separate error types
			Kind: ErrExpectedArgument,
			Message: fmt.Sprintf(
				"positional arguments are missing: %s",
				internal.Join(func(s positionalArg) string {
					return s.name
				}, ", ", cmd.posArgs[len(args):]...)),
		}
	}

	if cmd.posArgsRest.Valid {
		cmd.posArgsRest.Value.values = args[len(cmd.posArgs):]
	} else if len(args) > len(cmd.posArgs) {
		return fmt.Errorf(
			"extraneous positional arguments found: [%s]",
			internal.Join(strconv.Quote, ", ", args[len(cmd.posArgs):]...))
	}

	for i, arg := range args[:len(cmd.posArgs)] {
		cmd.posArgs[i].value = arg
	}

	// filled arguments values, now time to call all setters and validate all of that

	for _, f := range cmd.flags {
		if err := func() error {
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
					if f.minCount == 1 {
						return &Error{
							Kind:    ErrExpectedArgument,
							Message: "missing value",
						}
					}

					return &Error{
						Kind:    ErrExpectedArgument,
						Message: fmt.Sprintf("too few arguments, minimum %d required", f.minCount),
					}
				}

				if f.maxCount != -1 && len(f.values) > f.maxCount {
					return &Error{
						Kind:    ErrExpectedArgument,
						Message: fmt.Sprintf("too many arguments, maximum %d expected", f.maxCount),
					}
				}

				if len(f.values) > 0 {
					if err := f.valuesetter(f.values); err != nil {
						return err
					}
				} else if len(f.defaultValue) > 0 {
					if err := f.valuesetter(f.defaultValue); err != nil {
						return err
					}
				}
			}

			return nil
		}(); err != nil {
			return fmt.Errorf("invalid %s flag value: %w", f.show(), err)
		}
	}

	for _, arg := range cmd.posArgs {
		if err := arg.valuesetter(arg.value); err != nil {
			return fmt.Errorf("invalid %s positional argument value: %w", arg.name, err)
		}
	}

	if posArgsRest, ok := cmd.posArgsRest.Unpack(); ok {
		if err := posArgsRest.valuesetter(posArgsRest.values); err != nil {
			return fmt.Errorf("invalid rest of positional arguments: %w", err)
		}
	}

	return cmd.execute(ctx)
}
