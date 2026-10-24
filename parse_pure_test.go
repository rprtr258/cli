package cli

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDefaultValueUsingPureWithOk(t *testing.T) {
	// ~ this is a bit artifactial (it's better to use fallback_with instead actually),
	// but this example demonstrates the concept
	a := Short('t').Argument("N", ParseUint32)
	b := PureWith(func() (uint32, error) { return 42, nil })

	opts := Choice(a, b).ToOptions()

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
			r, err := opts.RunInner(newArgList(tc.args...))
			RequireNoParseFailure(t, err)
			require.Equal(t, tc.expected, r)
		})
	}
}

func TestDefaultValueUsingPureWithErr(t *testing.T) {
	a := Short('t').Argument("N", ParseUint32)
	b := PureWith(func() (uint32, error) { return 0, errors.New("some-err") })

	opts := Choice(a, b).ToOptions()
	assertFailure(t, opts, "some-err")
}

func TestDefaultValueUsingPureWithOkForSome(t *testing.T) {
	user_seeds := Positional("SEED", ParseUint32).Parser().Some("at least one required")
	last_seeds := PureWith(func() ([]uint32, error) {
		// ~ trying to lookup the last used seeds
		return []uint32{3, 5, 7, 11}, nil
	})
	seeds := Choice(user_seeds, last_seeds).ToOptions()

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
			r, err := seeds.RunInner(newArgList(tc.args...))
			RequireNoParseFailure(t, err)
			require.Equal(t, tc.expected, r)
		})
	}
}

func TestDefaultValueUsingPureWithErrForSome(t *testing.T) {
	user_seeds := Positional("SEED", ParseUint32).Parser().Some("at least one required")
	last_seeds := PureWith(func() ([]uint32, error) {
		// ~ trying to lookup the last used seeds but failing - parser fails, fallback works
		return nil, errors.New("oh, no")
	})
	default_seeds := Pure([]uint32{1, 2})
	seeds := Choice(user_seeds, last_seeds, default_seeds).ToOptions()

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
			r, err := seeds.RunInner(newArgList(tc.args...))
			RequireNoParseFailure(t, err)
			require.Equal(t, tc.expected, r)
		})
	}
}
