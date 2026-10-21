package cli_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	. "github.com/rprtr258/cli"
)

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
			r, err := opts.RunInner(NewArgs(tc.args...))
			RequireNoParseFailure(t, err)
			require.Equal(t, tc.expected, r)
		})
	}
}

func TestDefaultValueUsingPureWithErr(t *testing.T) {
	a := Argument(Short('t'), "N", parseUint32)
	b := PureWith(func() (uint32, Error) { return 0, ErrorParseFail("some-err") })

	opts := Choice(a.Parser(), b.Parser()).ToOptions()
	assertFailure(t, opts, "some-err")
}

func TestDefaultValueUsingPureWithOkForSome(t *testing.T) {
	user_seeds := Some(Positional("SEED", parseUint32), "at least one required", false)
	last_seeds := PureWith(func() ([]uint32, Error) {
		// ~ trying to lookup the last used seeds
		return []uint32{3, 5, 7, 11}, nil
	})
	seeds := Choice(user_seeds, last_seeds.Parser()).ToOptions()

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
			r, err := seeds.RunInner(NewArgs(tc.args...))
			RequireNoParseFailure(t, err)
			require.Equal(t, tc.expected, r)
		})
	}
}

func TestDefaultValueUsingPureWithErrForSome(t *testing.T) {
	user_seeds := Some(Positional("SEED", parseUint32), "at least one required", false)
	last_seeds := PureWith(func() ([]uint32, Error) {
		// ~ trying to lookup the last used seeds but failing - parser fails, fallback works
		return nil, ErrorParseFail("oh, no")
	})
	default_seeds := Pure([]uint32{1, 2})
	seeds := Choice(user_seeds, last_seeds.Parser(), default_seeds.Parser()).ToOptions()

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
			r, err := seeds.RunInner(NewArgs(tc.args...))
			RequireNoParseFailure(t, err)
			require.Equal(t, tc.expected, r)
		})
	}
}
