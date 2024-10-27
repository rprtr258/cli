package cli

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

var M Metavar = Metavar("M")

func Statefrom(value []string) State {
	args := newArgs(value...)
	msg := Invalid[Message]()
	res := stateconstruct(args, nil, nil, &msg)
	if err, ok := msg.Unpack(); ok {
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

func long_arg(t *testing.T) {
	a := Statefrom([]string{"--speed", "12"})
	s := unwrapunwrap(a.take_arg(Long("speed"), false, M))
	require.Equal(t, s, "12")
	require.True(t, a.is_empty())
}

func long_flag_and_positional(t *testing.T) {
	a := Statefrom([]string{"--speed", "12"})
	flag := a.take_flag(Long("speed"))
	require.True(t, flag)
	require.True(t, !a.is_empty())
	_, _, s, err := a.take_positional_word(M)
	require.Nil(t, err)
	require.Equal(t, s, "12")
	require.True(t, a.is_empty())
}

func multiple_short_flags(t *testing.T) {
	args := newArgs("-vvv")
	err := Invalid[Message]()
	a := stateconstruct(args, []byte{'v'}, []byte{}, &err)
	require.True(t, a.take_flag(Short('v')))
	require.True(t, a.take_flag(Short('v')))
	require.True(t, a.take_flag(Short('v')))
	require.True(t, !a.take_flag(Short('v')))
	require.True(t, a.is_empty())
}

func long_arg_with_equality(t *testing.T) {
	a := Statefrom([]string{"--speed=12"})
	s := unwrapunwrap(a.take_arg(Long("speed"), false, M))
	require.Equal(t, s, "12")
	require.True(t, a.is_empty())
}

func long_arg_with_equality_and_minus(t *testing.T) {
	a := Statefrom([]string{"--speed=-12"})
	s := unwrapunwrap(a.take_arg(Long("speed"), true, M))
	require.Equal(t, s, "-12")
	require.True(t, a.is_empty())
}

func short_arg_with_equality(t *testing.T) {
	a := Statefrom([]string{"-s=12"})
	s := unwrapunwrap(a.take_arg(Short('s'), false, M))
	require.Equal(t, s, "12")
	require.True(t, a.is_empty())
}

func short_arg_with_equality_and_minus(t *testing.T) {
	a := Statefrom([]string{"-s=-12"})
	s := unwrapunwrap(a.take_arg(Short('s'), false, M))
	require.Equal(t, s, "-12")
	require.True(t, a.is_empty())
}

func short_arg_with_equality_and_minus_is_adjacent(t *testing.T) {
	a := Statefrom([]string{"-s=-12"})
	s := unwrapunwrap(a.take_arg(Short('s'), true, M))
	require.Equal(t, s, "-12")
	require.True(t, a.is_empty())
}

func short_arg_without_equality(t *testing.T) {
	a := Statefrom([]string{"-s", "12"})
	s := unwrapunwrap(a.take_arg(Short('s'), false, M))
	require.Equal(t, s, "12")
	require.True(t, a.is_empty())
}

func two_short_flags(t *testing.T) {
	a := Statefrom([]string{"-s", "-v"})
	require.True(t, a.take_flag(Short('s')))
	require.True(t, a.take_flag(Short('v')))
	require.True(t, a.is_empty())
}

func two_short_flags2(t *testing.T) {
	a := Statefrom([]string{"-s", "-v"})
	require.True(t, a.take_flag(Short('v')))
	require.True(t, !a.take_flag(Short('v')))
	require.True(t, a.take_flag(Short('s')))
	require.True(t, !a.take_flag(Short('s')))
	require.True(t, a.is_empty())
}

func command_with_flags(t *testing.T) {
	a := Statefrom([]string{"cmd", "-s", "v"})
	require.True(t, a.take_cmd("cmd"))
	s := unwrapunwrap(a.take_arg(Short('s'), false, M))
	require.Equal(t, s, "v")
	require.True(t, a.is_empty())
}

func command_and_positional(t *testing.T) {
	a := Statefrom([]string{"cmd", "pos"})
	require.True(t, a.take_cmd("cmd"))
	_, _, w, err := a.take_positional_word(M)
	require.Nil(t, err)
	require.Equal(t, w, "pos")
	require.True(t, a.is_empty())
}

func positionals_after_double_dash1(t *testing.T) {
	a := Statefrom([]string{"-v", "--", "-x"})
	require.True(t, a.take_flag(Short('v')))
	_, _, w, err := a.take_positional_word(M)
	require.Nil(t, err)
	require.Equal(t, w, "-x")
	require.True(t, a.is_empty())
}

func positionals_after_double_dash2(t *testing.T) {
	a := Statefrom([]string{"-v", "--", "-x"})
	require.True(t, a.take_flag(Short('v')))
	_, _, w, err := a.take_positional_word(M)
	require.Nil(t, err)
	require.Equal(t, w, "-x")
	require.True(t, a.is_empty())
}

func positionals_after_double_dash3(t *testing.T) {
	a := Statefrom([]string{"-v", "12", "--", "-x"})
	w := unwrapunwrap(a.take_arg(Short('v'), false, M))
	require.Equal(t, w, "12")
	_, _, w, err := a.take_positional_word(M)
	require.Nil(t, err)
	require.Equal(t, w, "-x")
	require.True(t, a.is_empty())
}

func ambiguity_towards_flag(t *testing.T) {
	args := newArgs("-abc")
	err := Invalid[Message]()
	a := stateconstruct(args, []byte{'a', 'b', 'c'}, []byte{}, &err)

	require.True(t, a.take_flag(Short('a')))
	require.True(t, a.take_flag(Short('b')))
	require.True(t, a.take_flag(Short('c')))
}

func ambiguity_towards_argument(t *testing.T) {
	args := newArgs("-abc")
	err := Invalid[Message]()
	a := stateconstruct(args, []byte{}, []byte{'a'}, &err)

	r := unwrapunwrap(a.take_arg(Short('a'), false, M))
	require.Equal(t, r, "bc")
}

func ambiguity_towards_error(t *testing.T) {
	args := newArgs("-abc")
	err := Invalid[Message]()
	_ = stateconstruct(args, []byte{'a', 'b', 'c'}, []byte{'a'}, &err)
	require.True(t, err.Valid)
}

func ambiguity_towards_default(t *testing.T) {
	// AKA unresolved
	a := Statefrom([]string{"-abc"})
	var is_ambig bool
	a.peek().unwrap().value(nil, nil, nil, func() { is_ambig = true }, nil)
	require.True(t, is_ambig)
}
