//go:build !windows || forceposix
// +build !windows forceposix

package flags

import (
	"strings"

	"github.com/rprtr258/fun"
)

const (
	_defaultOptDelimiterShort = '-'
	_defaultOptDelimiterLong  = "--"
	_defaultArgDelimiterName  = '='
)

func isFlag(arg string) bool {
	return len(arg) > 1 && arg[0] == '-' && arg[1] != '-' ||
		len(arg) > 2 && arg[0] == '-' && arg[1] == '-' && arg[2] != '-'
}

// stripFlagPrefix returns the option without the prefix and whether or
// not the option is a long option or not.
func stripFlagPrefix(arg string) (prefix, name string, islong bool) {
	switch {
	case strings.HasPrefix(arg, "--"):
		return "--", arg[2:], true
	case strings.HasPrefix(arg, "-"):
		return "-", arg[1:], false
	default:
		return "", arg, false
	}
}

// splitFlag attempts to split the passed option into a name and an argument.
// When there is no argument specified, nil will be returned for it.
func splitFlag(prefix, option string, islong bool) (option_ string, rest_ *string) {
	if pos := strings.Index(option, "="); islong && pos >= 0 || !islong && pos == 1 {
		return option[:pos], fun.Ptr(option[pos+1:])
	}

	return option, nil
}

// addHelpGroup adds a new group that contains default help parameters.
func (c *Command) addHelpGroup(help func() string) *Group {
	ret, _ := c.AddGroup("Help Options", "", &struct {
		ShowHelp func() error `short:"h" long:"help" description:"Show this help message"`
	}{
		ShowHelp: func() error {
			return &Error{
				Kind:    ErrHelp,
				Message: help(),
			}
		},
	})
	ret.isBuiltinHelp = true
	return ret
}
