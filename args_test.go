package cli

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

var M = Metavar("M")

func Statefrom(value []string) State {
	args := NewArgs(value...)
	res, err := newState(args, nil, nil)
	if err != nil {
		panic(fmt.Sprintf("Couldn't construct state: %#v/%#v", err, res))
	}
	return res
}

func unwrapunwrap(s Option[string], err Error) string {
	if err != nil {
		panic(err)
	}
	return s.unwrap()
}

func TestLong_arg(t *testing.T) {
	a := Statefrom([]string{"--speed", "12"})
	s := unwrapunwrap(a.takeFlag(Long("speed"), false, M))
	require.Equal(t, s, "12")
	require.Zero(t, a.len())
}

func TestLong_flag_and_positional(t *testing.T) {
	a := Statefrom([]string{"--speed", "12"})
	flag := a.takeBoolFlag(Long("speed"))
	require.True(t, flag)
	require.NotZero(t, a.len())
	_, _, s, err := a.takePositional(M)
	require.Nil(t, err)
	require.Equal(t, s, "12")
	require.Zero(t, a.len())
}

func TestMultiple_short_flags(t *testing.T) {
	args := NewArgs("-vvv")
	a, _ := newState(args, []byte{'v'}, []byte{})
	require.True(t, a.takeBoolFlag(Short('v')))
	require.True(t, a.takeBoolFlag(Short('v')))
	require.True(t, a.takeBoolFlag(Short('v')))
	require.True(t, !a.takeBoolFlag(Short('v')))
	require.Zero(t, a.len())
}

func TestLong_arg_with_equality(t *testing.T) {
	a := Statefrom([]string{"--speed=12"})
	s := unwrapunwrap(a.takeFlag(Long("speed"), false, M))
	require.Equal(t, s, "12")
	require.Zero(t, a.len())
}

func TestLong_arg_with_equality_and_minus(t *testing.T) {
	a := Statefrom([]string{"--speed=-12"})
	s := unwrapunwrap(a.takeFlag(Long("speed"), true, M))
	require.Equal(t, s, "-12")
	require.Zero(t, a.len())
}

func TestShort_arg_with_equality(t *testing.T) {
	a := Statefrom([]string{"-s=12"})
	s := unwrapunwrap(a.takeFlag(Short('s'), false, M))
	require.Equal(t, s, "12")
	require.Zero(t, a.len())
}

func TestShort_arg_with_equality_and_minus(t *testing.T) {
	a := Statefrom([]string{"-s=-12"})
	s := unwrapunwrap(a.takeFlag(Short('s'), false, M))
	require.Equal(t, s, "-12")
	require.Zero(t, a.len())
}

func TestShort_arg_with_equality_and_minus_is_adjacent(t *testing.T) {
	a := Statefrom([]string{"-s=-12"})
	s := unwrapunwrap(a.takeFlag(Short('s'), true, M))
	require.Equal(t, s, "-12")
	require.Zero(t, a.len())
}

func TestShort_arg_without_equality(t *testing.T) {
	a := Statefrom([]string{"-s", "12"})
	s := unwrapunwrap(a.takeFlag(Short('s'), false, M))
	require.Equal(t, s, "12")
	require.Zero(t, a.len())
}

func TestTwo_short_flags(t *testing.T) {
	a := Statefrom([]string{"-s", "-v"})
	require.True(t, a.takeBoolFlag(Short('s')))
	require.True(t, a.takeBoolFlag(Short('v')))
	require.Zero(t, a.len())
}

func TestTwo_short_flags2(t *testing.T) {
	a := Statefrom([]string{"-s", "-v"})
	require.True(t, a.takeBoolFlag(Short('v')))
	require.True(t, !a.takeBoolFlag(Short('v')))
	require.True(t, a.takeBoolFlag(Short('s')))
	require.True(t, !a.takeBoolFlag(Short('s')))
	require.Zero(t, a.len())
}

func TestCommand_with_flags(t *testing.T) {
	a := Statefrom([]string{"cmd", "-s", "v"})
	require.True(t, a.take_cmd("cmd"))
	s := unwrapunwrap(a.takeFlag(Short('s'), false, M))
	require.Equal(t, s, "v")
	require.Zero(t, a.len())
}

func TestCommand_and_positional(t *testing.T) {
	a := Statefrom([]string{"cmd", "pos"})
	require.True(t, a.take_cmd("cmd"))
	_, _, w, err := a.takePositional(M)
	require.Nil(t, err)
	require.Equal(t, w, "pos")
	require.Zero(t, a.len())
}

func TestPositionals_after_double_dash1(t *testing.T) {
	a := Statefrom([]string{"-v", "--", "-x"})
	require.True(t, a.takeBoolFlag(Short('v')))
	_, _, w, err := a.takePositional(M)
	require.Nil(t, err)
	require.Equal(t, w, "-x")
	require.Zero(t, a.len())
}

func TestPositionals_after_double_dash2(t *testing.T) {
	a := Statefrom([]string{"-v", "--", "-x"})
	require.True(t, a.takeBoolFlag(Short('v')))
	_, _, w, err := a.takePositional(M)
	require.Nil(t, err)
	require.Equal(t, w, "-x")
	require.Zero(t, a.len())
}

func TestPositionals_after_double_dash3(t *testing.T) {
	a := Statefrom([]string{"-v", "12", "--", "-x"})
	w := unwrapunwrap(a.takeFlag(Short('v'), false, M))
	require.Equal(t, w, "12")
	_, _, w, err := a.takePositional(M)
	require.Nil(t, err)
	require.Equal(t, w, "-x")
	require.Zero(t, a.len())
}

func TestAmbiguity_towards_flag(t *testing.T) {
	args := NewArgs("-abc")
	a, _ := newState(args, []byte{'a', 'b', 'c'}, []byte{})

	require.True(t, a.takeBoolFlag(Short('a')))
	require.True(t, a.takeBoolFlag(Short('b')))
	require.True(t, a.takeBoolFlag(Short('c')))
}

func TestAmbiguity_towards_argument(t *testing.T) {
	args := NewArgs("-abc")
	a, _ := newState(args, []byte{}, []byte{'a'})

	r := unwrapunwrap(a.takeFlag(Short('a'), false, M))
	require.Equal(t, r, "bc")
}

func TestAmbiguity_towards_error(t *testing.T) {
	args := NewArgs("-abc")
	_, err := newState(args, []byte{'a', 'b', 'c'}, []byte{'a'})
	require.NotNil(t, err)
}

func TestAmbiguity_towards_default(t *testing.T) {
	// AKA unresolved
	a := Statefrom([]string{"-abc"})
	aa, ok := a.peek()
	require.True(t, ok)
	is_ambig := aa.kind == ArgKindWord
	require.True(t, is_ambig)
}
