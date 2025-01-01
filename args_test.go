package cli

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

var M = Metavar("M")

func Statefrom(value []string) State {
	args := newArgs(value...)
	res, err := stateconstruct(args, nil, nil)
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
	s := unwrapunwrap(a.take_arg(Long("speed"), false, M))
	require.Equal(t, s, "12")
	require.Equal(t, 0, a.len())
}

func TestLong_flag_and_positional(t *testing.T) {
	a := Statefrom([]string{"--speed", "12"})
	flag := a.take_flag(Long("speed"))
	require.True(t, flag)
	require.NotEqual(t, 0, a.len())
	_, _, s, err := a.take_positional_word(M)
	require.Nil(t, err)
	require.Equal(t, s, "12")
	require.Equal(t, 0, a.len())
}

func TestMultiple_short_flags(t *testing.T) {
	args := newArgs("-vvv")
	a, _ := stateconstruct(args, []byte{'v'}, []byte{})
	require.True(t, a.take_flag(Short('v')))
	require.True(t, a.take_flag(Short('v')))
	require.True(t, a.take_flag(Short('v')))
	require.True(t, !a.take_flag(Short('v')))
	require.Equal(t, 0, a.len())
}

func TestLong_arg_with_equality(t *testing.T) {
	a := Statefrom([]string{"--speed=12"})
	s := unwrapunwrap(a.take_arg(Long("speed"), false, M))
	require.Equal(t, s, "12")
	require.Equal(t, 0, a.len())
}

func TestLong_arg_with_equality_and_minus(t *testing.T) {
	a := Statefrom([]string{"--speed=-12"})
	s := unwrapunwrap(a.take_arg(Long("speed"), true, M))
	require.Equal(t, s, "-12")
	require.Equal(t, 0, a.len())
}

func TestShort_arg_with_equality(t *testing.T) {
	a := Statefrom([]string{"-s=12"})
	s := unwrapunwrap(a.take_arg(Short('s'), false, M))
	require.Equal(t, s, "12")
	require.Equal(t, 0, a.len())
}

func TestShort_arg_with_equality_and_minus(t *testing.T) {
	a := Statefrom([]string{"-s=-12"})
	s := unwrapunwrap(a.take_arg(Short('s'), false, M))
	require.Equal(t, s, "-12")
	require.Equal(t, 0, a.len())
}

func TestShort_arg_with_equality_and_minus_is_adjacent(t *testing.T) {
	a := Statefrom([]string{"-s=-12"})
	s := unwrapunwrap(a.take_arg(Short('s'), true, M))
	require.Equal(t, s, "-12")
	require.Equal(t, 0, a.len())
}

func TestShort_arg_without_equality(t *testing.T) {
	a := Statefrom([]string{"-s", "12"})
	s := unwrapunwrap(a.take_arg(Short('s'), false, M))
	require.Equal(t, s, "12")
	require.Equal(t, 0, a.len())
}

func TestTwo_short_flags(t *testing.T) {
	a := Statefrom([]string{"-s", "-v"})
	require.True(t, a.take_flag(Short('s')))
	require.True(t, a.take_flag(Short('v')))
	require.Equal(t, 0, a.len())
}

func TestTwo_short_flags2(t *testing.T) {
	a := Statefrom([]string{"-s", "-v"})
	require.True(t, a.take_flag(Short('v')))
	require.True(t, !a.take_flag(Short('v')))
	require.True(t, a.take_flag(Short('s')))
	require.True(t, !a.take_flag(Short('s')))
	require.Equal(t, 0, a.len())
}

func TestCommand_with_flags(t *testing.T) {
	a := Statefrom([]string{"cmd", "-s", "v"})
	require.True(t, a.take_cmd("cmd"))
	s := unwrapunwrap(a.take_arg(Short('s'), false, M))
	require.Equal(t, s, "v")
	require.Equal(t, 0, a.len())
}

func TestCommand_and_positional(t *testing.T) {
	a := Statefrom([]string{"cmd", "pos"})
	require.True(t, a.take_cmd("cmd"))
	_, _, w, err := a.take_positional_word(M)
	require.Nil(t, err)
	require.Equal(t, w, "pos")
	require.Equal(t, 0, a.len())
}

func TestPositionals_after_double_dash1(t *testing.T) {
	a := Statefrom([]string{"-v", "--", "-x"})
	require.True(t, a.take_flag(Short('v')))
	_, _, w, err := a.take_positional_word(M)
	require.Nil(t, err)
	require.Equal(t, w, "-x")
	require.Equal(t, 0, a.len())
}

func TestPositionals_after_double_dash2(t *testing.T) {
	a := Statefrom([]string{"-v", "--", "-x"})
	require.True(t, a.take_flag(Short('v')))
	_, _, w, err := a.take_positional_word(M)
	require.Nil(t, err)
	require.Equal(t, w, "-x")
	require.Equal(t, 0, a.len())
}

func TestPositionals_after_double_dash3(t *testing.T) {
	a := Statefrom([]string{"-v", "12", "--", "-x"})
	w := unwrapunwrap(a.take_arg(Short('v'), false, M))
	require.Equal(t, w, "12")
	_, _, w, err := a.take_positional_word(M)
	require.Nil(t, err)
	require.Equal(t, w, "-x")
	require.Equal(t, 0, a.len())
}

func TestAmbiguity_towards_flag(t *testing.T) {
	args := newArgs("-abc")
	a, _ := stateconstruct(args, []byte{'a', 'b', 'c'}, []byte{})

	require.True(t, a.take_flag(Short('a')))
	require.True(t, a.take_flag(Short('b')))
	require.True(t, a.take_flag(Short('c')))
}

func TestAmbiguity_towards_argument(t *testing.T) {
	args := newArgs("-abc")
	a, _ := stateconstruct(args, []byte{}, []byte{'a'})

	r := unwrapunwrap(a.take_arg(Short('a'), false, M))
	require.Equal(t, r, "bc")
}

func TestAmbiguity_towards_error(t *testing.T) {
	args := newArgs("-abc")
	_, err := stateconstruct(args, []byte{'a', 'b', 'c'}, []byte{'a'})
	require.NotNil(t, err)
}

func TestAmbiguity_towards_default(t *testing.T) {
	// AKA unresolved
	a := Statefrom([]string{"-abc"})
	aa, ok := a.peek()
	require.True(t, ok)
	_, is_ambig := aa.value.(argWord)
	require.True(t, is_ambig)
}
