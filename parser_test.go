package cli

import (
	"strconv"
	"testing"

	"github.com/stretchr/testify/require"
)

func assertParse[T any](t *testing.T, parser parser[T], expected T, args ...string) {
	t.Helper()

	x, err := parse(parser, args...)
	require.Nil(t, err)
	require.Equal(t, expected, x)
}

func assertFail[T any](t *testing.T, parser parser[T], expected string, args ...string) {
	t.Helper()

	_, err := parse(parser, args...)
	require.NotNil(t, err)
	require.Equal(t, expected, err.Error())
}

func TestFail(t *testing.T) {
	parser := Fail[int]("nope")
	assertFail(t, parser, "nope")
}

func TestPure(t *testing.T) {
	parser := Pure(123)
	assertParse(t, parser, 123)
}

func TestPositional(t *testing.T) {
	parser := Positional("ARG", strconv.Atoi)
	assertParse(t, parser, 123, "123")
}

func TestFlagBoolShort(t *testing.T) {
	parser := FlagBoolShort('a')
	assertParse(t, parser, true, "-a")
	assertFail(t, parser, "unexpected argument `-b`", "-b")
}

func TestFlagBoolLong(t *testing.T) {
	parser := FlagBoolLong("foo")
	assertParse(t, parser, true, "--foo")
	assertFail(t, parser, "unexpected argument `--bar`", "--bar")
}

func TestFlagBoolShortLong(t *testing.T) {
	parser := FlagBoolShortLong('a', "foo")
	assertParse(t, parser, true, "-a")
	assertFail(t, parser, "unexpected argument `-b`", "-b")
	assertParse(t, parser, true, "--foo")
	assertFail(t, parser, "unexpected argument `--bar`", "--bar")
}

func TestFlagShort(t *testing.T) {
	parser := FlagShort('a', "foo", strconv.Atoi)
	assertParse(t, parser, 123, "-a", "123")
}

func TestFlagLong(t *testing.T) {
	parser := FlagLong("foo", "bar", strconv.Atoi)
	assertParse(t, parser, 123, "--foo", "123")
}

func TestFlagShortLong(t *testing.T) {
	parser := FlagShortLong('a', "foo", "bar", strconv.Atoi)
	assertParse(t, parser, 123, "-a", "123")
	assertParse(t, parser, 123, "--foo", "123")
}

func TestReqFlag(t *testing.T) {
	parser := FlagLong("foo", "bar", strconv.Atoi)
	assertParse(t, parser, 123, "--foo", "123")
}
