package cli

import (
	"errors"
	"strconv"
	"time"
)

// Ready-made conversion functions from raw command line strings to common
// types, for use with [Argument] and [Positional].
//
// Both [Argument] and [Positional] take a `func(string) (T, error)` to convert
// the raw text; these adapters cover the common types so no hand-written
// `strconv` glue is needed. Custom conversions keep working as before.
//
// # Usage
//
//	width := cli.Argument(cli.Long("width").Help("Width of the rectangle"), "PX", cli.ParseInt)

// unwrapNumError unwraps the strconv number error into the underlying cause.
func unwrapNumError(err error) error {
	var numErr *strconv.NumError
	if errors.As(err, &numErr) {
		return numErr.Err
	}
	return err
}

// ParseString is the identity conversion for [Argument] and [Positional]:
// the parsed value is the string itself.
func ParseString(s string) (string, error) {
	return s, nil
}

// ParseInt parses a signed base-10 integer, for use with [Argument] and
// [Positional].
func ParseInt(s string) (int, error) {
	x, err := strconv.Atoi(s)
	if err != nil {
		return 0, unwrapNumError(err)
	}
	return x, nil
}

// ParseUint parses an unsigned base-10 integer, for use with [Argument] and
// [Positional].
func ParseUint(s string) (uint, error) {
	x, err := strconv.ParseUint(s, 10, 0)
	if err != nil {
		return 0, unwrapNumError(err)
	}
	return uint(x), nil
}

// ParseFloat parses a base-10 floating point number, for use with [Argument]
// and [Positional].
func ParseFloat(s string) (float64, error) {
	x, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0, unwrapNumError(err)
	}
	return x, nil
}

// ParseBool parses a boolean accepting `1, t, T, TRUE, true, True, 0, f, F,
// FALSE, false, False`, for use with [Argument] and [Positional].
func ParseBool(s string) (bool, error) {
	x, err := strconv.ParseBool(s)
	if err != nil {
		return false, unwrapNumError(err)
	}
	return x, nil
}

// ParseDuration parses a duration like `1h2m4s` with [time.ParseDuration],
// for use with [Argument] and [Positional].
func ParseDuration(s string) (time.Duration, error) {
	return time.ParseDuration(s)
}
