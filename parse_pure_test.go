package cli

import (
	"errors"
	"strconv"
	"testing"

	"github.com/stretchr/testify/require"
)

// TODO: move similar funcs to batteries and use
func parseUint32(s string) (uint32, error) {
	u, err := strconv.ParseUint(s, 10, 32)
	if err != nil {
		return 0, errors.New("invalid digit found in string")
	}
	return uint32(u), nil
}

// TODO: move similar funcs to batteries and use
func parseInt32(s string) (int32, error) {
	u, err := strconv.ParseInt(s, 10, 32)
	if err != nil {
		return 0, errors.New("invalid digit found in string")
	}
	return int32(u), nil
}

// TODO: move similar funcs to batteries and use
func parseString(s string) (string, error) {
	return s, nil
}

// TODO: move similar funcs to batteries and use
func parseUint(s string) (uint, error) {
	u, err := strconv.ParseUint(s, 10, 32)
	if err != nil {
		return 0, errors.New("invalid digit found in string")
	}
	return uint(u), nil
}

func TestDefaultValueUsingPureWithOk(t *testing.T) {
	// ~ this is a bit artifactial (it's better to use fallback_with instead actually),
	// but this example demonstrates the concept
	a := Argument(Short('t'), "N", parseUint32)
	b := PureWith(func() (uint32, Error) { return 42, nil })

	opts := Choice(a.Parser(), b.Parser()).ToOptions()

	for name, tc := range map[string]struct {
		args     []string
		expected uint32
	}{
		"parse": {
			args:     []string{"-t", "87"},
			expected: 87,
		},
		"default": {
			args:     nil,
			expected: 42,
		},
	} {
		t.Run(name, func(t *testing.T) {
			r, err := opts.run_inner(newArgs(tc.args...))
			requireNoParseFailure(t, err)
			require.Equal(t, tc.expected, r)
		})
	}
}

func TestDefaultValueUsingPureWithErr(t *testing.T) {
	a := Argument(Short('t'), "N", parseUint32)
	b := PureWith(func() (uint32, Error) { return 0, ErrorParseFail("some-err") })

	opts := Choice(a.Parser(), b.Parser()).ToOptions()
	_, err := opts.run_inner(newArgs())
	require.Equal(t, "some-err", unwrap_stderr(err))
}

func TestDefaultValueUsingPureWithOkForSome(t *testing.T) {
	user_seeds := Positional("SEED", parseUint32).Parser().Some("at least one required")
	last_seeds := PureWith(func() ([]uint32, Error) {
		// ~ trying to lookup the last used seeds
		return []uint32{3, 5, 7, 11}, nil
	})
	seeds := Choice(ParseSomeParser(user_seeds), last_seeds.Parser()).ToOptions()

	for name, tc := range map[string]struct {
		args     []string
		expected []uint32
	}{
		"23,59": {
			args:     []string{"23", "59"},
			expected: []uint32{23, 59},
		},
		"empty": {
			args:     nil,
			expected: []uint32{3, 5, 7, 11},
		},
	} {
		t.Run(name, func(t *testing.T) {
			r, err := seeds.run_inner(newArgs(tc.args...))
			requireNoParseFailure(t, err)
			require.Equal(t, tc.expected, r)
		})
	}
}

func TestDefaultValueUsingPureWithErrForSome(t *testing.T) {
	user_seeds := Positional("SEED", parseUint32).Parser().Some("at least one required")
	last_seeds := PureWith(func() ([]uint32, Error) {
		// ~ trying to lookup the last used seeds but failing - parser fails, fallback works
		return nil, ErrorParseFail("oh, no")
	})
	default_seeds := Pure([]uint32{1, 2})
	seeds := Choice(ParseSomeParser(user_seeds), last_seeds.Parser(), default_seeds.Parser()).ToOptions()

	for name, tc := range map[string]struct {
		args     []string
		expected []uint32
	}{
		"23,59": {
			args:     []string{"23", "59"},
			expected: []uint32{23, 59},
		},
		"empty": {
			args:     nil,
			expected: []uint32{1, 2},
		},
	} {
		t.Run(name, func(t *testing.T) {
			r, err := seeds.run_inner(newArgs(tc.args...))
			requireNoParseFailure(t, err)
			require.Equal(t, tc.expected, r)
		})
	}
}
