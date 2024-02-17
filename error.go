package cli

import (
	"fmt"
	"strings"

	"github.com/samber/lo"
)

// ErrorKind represents the type of error.
type ErrorKind int

const (
	ErrUnknown          ErrorKind = iota // ErrUnknown indicates a generic error.
	ErrExpectedArgument                  // ErrExpectedArgument indicates that an argument was expected.
	ErrUnknownFlag                       // ErrUnknownFlag indicates an unknown flag.
	ErrUnknownGroup                      // ErrUnknownGroup indicates an unknown group.
	ErrHelp                              // ErrHelp indicates that the built-in help was shown (the error contains the help message).
	ErrShortNameTooLong                  // ErrShortNameTooLong indicates that a short flag name was specified, longer than one character.
	ErrDuplicatedFlag                    // ErrDuplicatedFlag indicates that a short or long flag has been defined more than once
	ErrTag                               // ErrTag indicates an error while parsing flag tags.
	ErrCommandRequired                   // ErrCommandRequired indicates that a command was required but not specified
	ErrUnknownCommand                    // ErrUnknownCommand indicates that an unknown command was specified.
	ErrInvalidTag                        // ErrInvalidTag indicates an invalid tag or invalid use of an existing tag
)

func (e ErrorKind) String() string {
	return lo.Switch[ErrorKind, string](e).
		Case(ErrUnknown, "unknown").
		Case(ErrExpectedArgument, "expected argument").
		Case(ErrUnknownFlag, "unknown flag").
		Case(ErrUnknownGroup, "unknown group").
		Case(ErrHelp, "help").
		Case(ErrShortNameTooLong, "short name too long").
		Case(ErrDuplicatedFlag, "duplicated flag").
		Case(ErrTag, "tag").
		Case(ErrCommandRequired, "command required").
		Case(ErrUnknownCommand, "unknown command").
		Case(ErrInvalidTag, "invalid tag").
		Default("unrecognized error type")
}

// Error represents a parser error. The error returned from Parse is of this
// type. The error contains both a Type and Message.
type Error struct {
	Kind    ErrorKind // The kind of error
	Message string    // The error message
}

// Error returns the error's message
func (e *Error) Error() string {
	return e.Message
}

func newErrorf(tp ErrorKind, format string, args ...any) *Error {
	return &Error{
		Kind:    tp,
		Message: fmt.Sprintf(format, args...),
	}
}

func wrapError(err error) *Error {
	if ret, ok := err.(*Error); ok {
		return ret
	}

	return &Error{
		Kind:    ErrUnknown,
		Message: err.Error(),
	}
}

// TODO: OPTIMIZE: less memory could be used
func dist(s, t string) int {
	if s == "" {
		return len(t)
	}

	if t == "" {
		return len(s)
	}

	w := len(t) + 1
	buf := make([]int, (len(s)+1)*w)

	dists := make([][]int, len(s)+1)
	for i := range dists {
		dists[i] = buf[w*i : w*(i+1)]
		dists[i][0] = i
	}

	for j := range t {
		dists[0][j] = j
	}

	for i, sc := range s {
		for j, tc := range t {
			if sc == tc {
				dists[i+1][j+1] = dists[i][j]
			} else {
				dists[i+1][j+1] = min(
					dists[i][j],
					dists[i+1][j],
					dists[i][j+1],
				) + 1
			}
		}
	}

	return dists[len(s)][len(t)]
}

func getClosest(s string, ss []string) (closest_ string, distMin_ int) {
	if len(ss) == 0 {
		return "", 0
	}

	closest, distMin := "", -1
	for _, t := range ss {
		if distCur := dist(s, t); distMin == -1 || distCur < distMin {
			closest, distMin = t, distCur
		}
	}
	return closest, distMin
}

func newUnknownCommandError(name string, cmdnames []string) error {
	c, l := getClosest(name, cmdnames)
	return &Error{
		Kind: ErrUnknownCommand,
		Message: fmt.Sprintf("Unknown command %q", name) + func() string {
			switch {
			case l*2 < len(c):
				return fmt.Sprintf(", did you mean %s?", c)
			case len(cmdnames) == 0:
				return ""
			case len(cmdnames) == 1:
				return fmt.Sprintf(". You should use the %s command", cmdnames[0])
			default:
				return fmt.Sprintf(". Please specify one command of: %s or %s",
					strings.Join(cmdnames[:len(cmdnames)-1], ", "),
					cmdnames[len(cmdnames)-1])
			}
		}(),
	}
}

func newCommandRequiredError(cmdnames []string) error {
	return &Error{
		Kind: ErrCommandRequired,
		Message: lo.Switch[int, string](len(cmdnames)).
			Case(0, "???").
			CaseF(1, func() string {
				return fmt.Sprintf("Please specify the %s command", cmdnames[0])
			}).
			DefaultF(func() string {
				return fmt.Sprintf("Please specify one command of: %s or %s",
					strings.Join(cmdnames[:len(cmdnames)-1], ", "),
					cmdnames[len(cmdnames)-1])
			}),
	}
}
