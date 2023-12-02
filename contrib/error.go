package flags

import (
	"fmt"

	"github.com/samber/lo"
)

// ErrorKind represents the type of error.
type ErrorKind int

const (
	ErrUnknown           ErrorKind = iota // ErrUnknown indicates a generic error.
	ErrExpectedArgument                   // ErrExpectedArgument indicates that an argument was expected.
	ErrUnknownFlag                        // ErrUnknownFlag indicates an unknown flag.
	ErrUnknownGroup                       // ErrUnknownGroup indicates an unknown group.
	ErrMarshal                            // ErrMarshal indicates a marshalling error while converting values.
	ErrHelp                               // ErrHelp indicates that the built-in help was shown (the error contains the help message).
	ErrNoArgumentForBool                  // ErrNoArgumentForBool indicates that an argument was given for a boolean flag (which don't not take any arguments).
	ErrRequired                           // ErrRequired indicates that a required flag was not provided.
	ErrShortNameTooLong                   // ErrShortNameTooLong indicates that a short flag name was specified, longer than one character.
	ErrDuplicatedFlag                     // ErrDuplicatedFlag indicates that a short or long flag has been defined more than once
	ErrTag                                // ErrTag indicates an error while parsing flag tags.
	ErrCommandRequired                    // ErrCommandRequired indicates that a command was required but not specified
	ErrUnknownCommand                     // ErrUnknownCommand indicates that an unknown command was specified.
	ErrInvalidChoice                      // ErrInvalidChoice indicates an invalid option value which only allows a certain number of choices.
	ErrInvalidTag                         // ErrInvalidTag indicates an invalid tag or invalid use of an existing tag
)

func (e ErrorKind) String() string {
	return lo.Switch[ErrorKind, string](e).
		Case(ErrUnknown, "unknown").
		Case(ErrExpectedArgument, "expected argument").
		Case(ErrUnknownFlag, "unknown flag").
		Case(ErrUnknownGroup, "unknown group").
		Case(ErrMarshal, "marshal").
		Case(ErrHelp, "help").
		Case(ErrNoArgumentForBool, "no argument for bool").
		Case(ErrRequired, "required").
		Case(ErrShortNameTooLong, "short name too long").
		Case(ErrDuplicatedFlag, "duplicated flag").
		Case(ErrTag, "tag").
		Case(ErrCommandRequired, "command required").
		Case(ErrUnknownCommand, "unknown command").
		Case(ErrInvalidChoice, "invalid choice").
		Case(ErrInvalidTag, "invalid tag").
		Default("unrecognized error type")
}

func (e ErrorKind) Error() string {
	return e.String()
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

func newError(tp ErrorKind, message string) *Error {
	return &Error{
		Kind:    tp,
		Message: message,
	}
}

func newErrorf(tp ErrorKind, format string, args ...interface{}) *Error {
	return newError(tp, fmt.Sprintf(format, args...))
}

func wrapError(err error) *Error {
	if ret, ok := err.(*Error); ok {
		return ret
	}

	return newError(ErrUnknown, err.Error())
}
