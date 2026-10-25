package cli

import (
	"strconv"
	"testing"

	"github.com/rprtr258/fun"
	"github.com/stretchr/testify/require"
)

func TestAdjacent(t *testing.T) {
	type cmd struct {
		a    struct{}
		b, c bool
	}

	a := Short('a').ReqFlag(struct{}{})
	b := Short('b').Switch()
	c := Short('c').Switch()

	parser := Pair(a, Pair(b, c)).Adjacent().Map(func(abc fun.Pair[struct{}, fun.Pair[bool, bool]]) cmd { return cmd{abc.K, abc.V.K, abc.V.V} }).Many().ToOptions()

	for name, tc := range map[string]struct {
		args     []string
		expected []cmd
	}{
		"ac,ab,abc": {
			// adjacent groups here argument
			// -a [-b] -c  | -a -b [-c] | -a -b -c
			args:     []string{"-a", "-c", "-a", "-b", "-a", "-b", "-c"},
			expected: []cmd{{struct{}{}, false, true}, {struct{}{}, true, false}, {struct{}{}, true, true}},
		},
		"a": {
			args:     []string{"-a"},
			expected: []cmd{{struct{}{}, false, false}},
		},
		"ac": {
			args:     []string{"-a", "-c"},
			expected: []cmd{{struct{}{}, false, true}},
		},
		"empty": {
			args:     []string{},
			expected: []cmd{},
		},
	} {
		t.Run(name, func(t *testing.T) {
			r, err := parser.run(newArgList(tc.args...))
			requireNoParseFailure(t, err)
			require.Equal(t, tc.expected, r)
		})
	}
}

func TestAdjacentPrefix(t *testing.T) {
	type cmd = fun.Pair[Opt[fun.Pair[struct{}, uint]], bool]
	a := Short('a').ReqFlag(struct{}{})
	b := Positional("X", func(s string) (uint, error) {
		u, err := strconv.ParseUint(s, 10, 32)
		return uint(u), err
	})
	ab := Pair(a, b).Adjacent().Optional()
	c := Short('c').Switch()
	parser := Pair(ab, c).ToOptions()

	for name, tc := range map[string]struct {
		args     []string
		expected cmd
	}{
		"-c": {
			args: []string{"-c"},
			expected: cmd{
				Invalid[fun.Pair[struct{}, uint]](),
				true,
			},
		},
		"-a 1": {
			args: []string{"-a", "1"},
			expected: cmd{
				Valid(ppair[struct{}, uint](struct{}{}, 1)),
				false,
			},
		},
		"-c -a 1": {
			args: []string{"-c", "-a", "1"},
			expected: cmd{
				Valid(ppair[struct{}, uint](struct{}{}, 1)),
				true,
			},
		},
		"-a 1 -c": {
			args: []string{"-a", "1", "-c"},
			expected: cmd{
				Valid(ppair[struct{}, uint](struct{}{}, 1)),
				true,
			},
		},
	} {
		t.Run(name, func(t *testing.T) {
			r, err := parser.run(newArgList(tc.args...))
			requireNoParseFailure(t, err)
			require.Equal(t, tc.expected, r)
		})
	}
}
