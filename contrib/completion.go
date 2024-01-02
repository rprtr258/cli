package flags

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"unicode/utf8"

	"github.com/rprtr258/cli/internal"
	"github.com/rprtr258/fun"
)

// Completion is a type containing information of a completion.
type Completion struct {
	Item        string // The completed item
	Description string // Description of the completed item (optional)
}

// Completer is an interface which can be implemented by types
// to provide custom command line argument completion.
type Completer interface {
	// Complete receives a prefix representing a (partial) value
	// for its type and should provide a list of possible valid completions.
	Complete(prefix string) []Completion
}

type completion struct {
	cmd *Command
}

func (c *completion) skipPositional(s *parseState, n int) {
	if n >= len(s.positional) {
		s.positional = nil
	} else {
		s.positional = s.positional[n:]
	}
}

func (c *completion) completeOptionNames(s *parseState, prefix, match string, short bool) []Completion {
	if short && match != "" {
		return []Completion{
			{
				Item: prefix + match,
			},
		}
	}

	var results []Completion
	repeats := map[string]struct{}{}
	for name, opt := range s.lookup.longNames {
		if strings.HasPrefix(name, match) && !opt.Hidden {
			results = append(results, Completion{
				Item:        _defaultOptDelimiterLong + name,
				Description: opt.Description,
			})

			if short {
				repeats[string(opt.ShortName)] = struct{}{}
			}
		}
	}
	if short {
		for name, opt := range s.lookup.shortNames {
			if _, ok := repeats[name]; !ok && strings.HasPrefix(name, match) && !opt.Hidden {
				results = append(results, Completion{
					Item:        string(_defaultOptDelimiterShort) + name,
					Description: opt.Description,
				})
			}
		}
	}
	return results
}

func (c *completion) completeCommands(s *parseState, match string) []Completion {
	return internal.FilterMap(func(cmd *Command) (Completion, bool) {
		return Completion{
			Item:        cmd.Name,
			Description: cmd.Flags.ShortDescription,
		}, cmd.Flags.data != c && !cmd.Flags.Hidden && strings.HasPrefix(cmd.Name, match)
	}, s.command.commands...)
}

func (c *completion) completeValue(value reflect.Value, prefix, match string) []Completion {
	if value.Kind() == reflect.Slice {
		value = reflect.New(value.Type().Elem())
	}

	var ret []Completion
	if cmp, ok := value.Interface().(Completer); ok {
		ret = cmp.Complete(match)
	} else if value.CanAddr() {
		if cmp, ok = value.Addr().Interface().(Completer); ok {
			ret = cmp.Complete(match)
		}
	}

	return fun.Map[Completion](func(v Completion) Completion {
		return Completion{
			Item:        prefix + v.Item,
			Description: v.Description,
		}
	}, ret...)
}

func (c *completion) complete(args []string) []Completion {
	if len(args) == 0 {
		args = []string{""}
	}

	s := &parseState{
		args:       args,
		command:    c.cmd,
		positional: c.cmd.args,
		lookup:     makeLookupTable(c.cmd),
	}
	var opt *Flag
	for len(s.args) > 1 {
		arg := s.pop()
		if arg == "--" {
			opt = nil
			c.skipPositional(s, len(s.args)-1)
			break
		}

		if isFlag(arg) {
			prefix, optname, islong := stripFlagPrefix(arg)
			optname, argument := splitFlag(prefix, optname, islong)

			if argument == nil {
				var o *Flag
				canarg := true
				if islong {
					o = s.lookup.longNames[optname]
				} else {
					for i, r := range optname {
						sname := string(r)
						o = s.lookup.shortNames[sname]
						if o == nil {
							break
						}

						if i == 0 && o.canArgument() && len(optname) != len(sname) {
							canarg = false
							break
						}
					}
				}

				if o == nil {
					opt = nil
					c.skipPositional(s, len(s.args)-1)
					break
				} else if o != nil && o.canArgument() && !o.IsOptional && canarg {
					if len(s.args) > 1 {
						s.pop()
					} else {
						opt = o
					}
				}
			}
		} else {
			if len(s.positional) > 0 {
				if !isRemainingArg(s.positional[0].value) {
					// Don't advance beyond a remaining positional arg
					// (because it consumes all subsequent args)
					s.positional = s.positional[1:]
				}
			} else if cmd, ok := s.lookup.commands[arg]; ok {
				s.positional = cmd.args
				s.lookup = makeLookupTable(cmd)
				s.command = cmd
			}

			opt = nil
		}
	}

	lastarg := s.args[len(s.args)-1]
	switch {
	case opt != nil: // Completion for the argument of 'opt'
		return c.completeValue(opt.value, "", lastarg)
	case lastarg != "" && lastarg[0] == '-': // Complete the option
		prefix, optname, islong := stripFlagPrefix(lastarg)
		optname, argument := splitFlag(prefix, optname, islong)

		switch {
		case argument == nil && !islong:
			rname, n := utf8.DecodeRuneInString(optname)
			sname := string(rname)

			if opt := s.lookup.shortNames[sname]; opt != nil && opt.canArgument() {
				return c.completeValue(opt.value, prefix+sname, optname[n:])
			}

			return c.completeOptionNames(s, prefix, optname, true)
		case argument != nil:
			opt = fun.IF(islong, s.lookup.longNames, s.lookup.shortNames)[optname]
			if opt == nil {
				return nil
			}

			return c.completeValue(opt.value, prefix+optname+"=", *argument)
		case islong:
			return c.completeOptionNames(s, prefix, optname, false)
		default:
			return c.completeOptionNames(s, prefix, optname, true)
		}
	case len(s.positional) > 0: // Complete for positional argument
		return c.completeValue(s.positional[0].value, "", lastarg)
	case len(s.command.commands) > 0: // Complete for command
		return c.completeCommands(s, lastarg)
	default:
		return nil
	}
}

// Filename is a string alias which provides filename completion.
type Filename string

// Complete returns a list of existing files with the given prefix.
func (f *Filename) Complete(prefix string) []Completion {
	ret, _ := filepath.Glob(prefix + "*")
	if len(ret) == 1 {
		if info, err := os.Stat(ret[0]); err == nil && info.IsDir() {
			ret[0] += "/"
		}
	}

	return fun.Map[Completion](func(v string) Completion {
		return Completion{
			Item:        v,
			Description: "",
		}
	}, ret...)
}
