package flags

import (
	"context"
	"fmt"
	"os"
	"sort"
	"strconv"

	"github.com/rprtr258/fun"

	"github.com/rprtr258/cli/internal"
)

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
		keys := fun.Keys(cmd.subcommands)
		sort.Strings(keys)
		if len(args) > 0 {
			return newUnknownCommandError(args[0], keys)
		}

		// TODO: squash subcommands aliases, print suggestion based on distance
		return newCommandRequiredError(keys)
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

	if len(args) > len(cmd.posArgs) && !cmd.posArgsRest.defined {
		return &Error{ // TODO: separate error types
			Kind: ErrExpectedArgument,
			Message: fmt.Sprintf(
				"extraneous positional arguments found: [%s]",
				internal.Join(strconv.Quote, ", ", args[len(cmd.posArgs):]...),
			),
		}
	}

	for i, arg := range fun.Subslice(0, len(cmd.posArgs), args...) {
		cmd.posArgs[i].value = arg
	}

	if len(args) > len(cmd.posArgs) {
		if !cmd.posArgsRest.defined {
			return fmt.Errorf("extraneous positional arguments found: [%s]", internal.Join(strconv.Quote, ", ", args[len(cmd.posArgs):]...))
		}

		cmd.posArgsRest.values = args[len(cmd.posArgs):]
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

				if f.values != nil {
					if err := f.valuesetter(f.values); err != nil {
						return err
					}
				}
			}

			return nil
		}(); err != nil {
			return fmt.Errorf("invalid %s flag value: %w", f.show(), err)
		}
	}

	if cmd.posArgsRest.defined {
		if err := cmd.posArgsRest.valuesetter(cmd.posArgsRest.values); err != nil {
			return fmt.Errorf("set posArgsRest: %w", err)
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
