package flags

import (
	"context"
	"fmt"
	"os"
	"reflect"
	"sort"
	"strings"

	"github.com/rprtr258/cli/internal"
	"github.com/rprtr258/fun"
	"github.com/rprtr258/fun/iter"
	"github.com/samber/lo"
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

func Run(data any, args ...string) error {
	return RunContext(context.Background(), data, args...)
}

// RunContext parses the command line arguments according to the data.
// Error indicates parsing error.
//
// When the common help group has been added (AddHelp) and either -h or --help
// was specified in the command line arguments, the special error type ErrHelp
// is returned with help text to print.
func RunContext(ctx context.Context, data any, args ...string) error {
	p := newParser(args[0], data)
	args = args[1:] // throw off executable name

	p.eachFlag(func(c *Command, g *flagSet, option *Flag) {
		option.clearReferenceBeforeSet = true
		option.defaultLiteral = option.updateDefaultLiteral()
	})
	p.addHelpGroups(p.writeHelp)

	if _, ok := os.LookupEnv("GO_FLAGS_COMPLETION"); ok {
		p.printCompletions(args...)
		return nil
	}

	s := &parseState{
		args:       args,
		retargs:    make([]string, 0, len(args)),
		positional: p.args,
		lookup:     makeLookupTable(p),
		command:    p,
	}
	for !s.noargs() {
		arg := s.pop()

		// When PassDoubleDash is set and we encounter a --, then
		// simply append all the rest as arguments and break out
		if arg == "--" {
			s.addArgs(s.args...)
			break
		}

		if !isFlag(arg) {
			if s.lookup.commands[arg] == nil {
				// If PassAfterNonOption is set then all remaining arguments are considered positional
				s.addArgs(s.arg)
				if s.err == nil {
					s.addArgs(s.args...)
				}
				break
			}

			if cmd := s.lookup.commands[s.arg]; len(s.positional) == 0 && len(s.command.commands) > 0 && len(s.retargs) == 0 && cmd != nil {
				s.command.Active = cmd
				s.positional = cmd.args
				s.lookup = makeLookupTable(cmd)
				s.command = cmd
			} else if s.addArgs(s.arg); s.err != nil {
				break
			}

			continue
		}

		prefix, optname, islong := stripFlagPrefix(arg)
		optname, argument := splitFlag(prefix, optname, islong)

		if err := fun.IF(islong, p.parseLong, p.parseShort)(s, optname, argument); err != nil {
			if parseErr := wrapError(err); parseErr.Kind != ErrUnknownFlag {
				s.err = parseErr
				break
			}
		}
	}

	if s.err == nil {
		p.eachFlag(func(c *Command, g *flagSet, flag *Flag) {
			if flag.preventDefault {
				return
			}

			flag.isSetDefault = true

			if len(flag.Default) > 0 {
				flag.empty()

				for _, d := range flag.Default {
					if flag.preventDefault {
						break
					}

					if err := flag.set(&d); err != nil {
						if err != nil {
							s.err = marshalError(flag, err)
						}
						break
					}

					flag.isSetDefault = true
					flag.preventDefault = false
				}
			} else if fun.Contains(flag.value.Type().Kind(), reflect.Map, reflect.Slice) && flag.value.IsNil() {
				flag.empty()
			}
		})

		var required []*Flag
		for c := p; c != nil; c = c.Active {
			for _, option := range c.Flags.flags {
				if !option.isSet && option.Required {
					required = append(required, option)
				}
			}
		}

		if len(required) == 0 {
			reqnames := iter.MapFilter(iter.FromMany(s.positional...),
				func(arg PositionalArg) (string, bool) {
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
				Slice()
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
		} else {
			names := fun.Map[string](func(k *Flag) string {
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
	} else {
		return s.err
	}

	if len(s.command.commands) != 0 && !s.command.SubcommandsOptional {
		cmdnames := fun.Map[string](func(v *Command) string {
			return v.Name
		}, visibleCommands(s.command.commands)...)

		if len(s.retargs) != 0 {
			return newUnknownCommandError(s.retargs[0], cmdnames)
		}

		return newCommandRequiredError(cmdnames)
	}

	if cmd, ok := s.command.Flags.data.(Commander); ok {
		return cmd.Execute(s.retargs)
	}

	return nil
}
