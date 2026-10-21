package cli_test

import (
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	. "github.com/rprtr258/cli"
)

// constructTestPort implements encoding.TextUnmarshaler.
type constructTestPort uint16

func (p *constructTestPort) UnmarshalText(b []byte) error {
	n, err := strconv.ParseUint(string(b), 10, 16)
	if err != nil {
		return err
	}
	*p = constructTestPort(n)
	return nil
}

func TestConstructBasic(t *testing.T) {
	type Out struct {
		Debug  bool    `cli:"short=d,long=debug,help='Activate debug mode'"`
		Speed  float64 `cli:"long=speed,metavar=SPEED,default=42"`
		Output string  `cli:"short=o,long=output,metavar=FILE"`
	}
	parser := Construct[Out]().Parser().ToOptions()

	AssertSuccess(t, Out{Debug: true, Speed: 3.14, Output: "out.txt"}, parser,
		"-d", "--speed", "3.14", "-o", "out.txt")
	AssertSuccess(t, Out{Debug: true, Speed: 42, Output: "out"}, parser, "--debug", "--output=out")
	// Output has no default - must fail with no arguments
	assertFailure(t, parser, "expected `-o=FILE`, pass `--help` for usage information")
}

func TestConstructPositionalBeforeNamed(t *testing.T) {
	type Out struct {
		Src   string // positional
		Dst   string // positional
		Force bool   `cli:"short=f,long=force"`
	}
	parser := Construct[Out]().Parser().ToOptions()

	AssertSuccess(t, Out{Src: "a", Dst: "b", Force: true}, parser, "-f", "a", "b")
}

func TestConstructMany(t *testing.T) {
	type Out struct {
		Files []string // positional many
		Tags  []string `cli:"long=tag,metavar=TAG"`
		Nums  []int    `cli:"short=n"`
	}
	parser := Construct[Out]().Parser().ToOptions()

	AssertSuccess(t,
		Out{Files: []string{"a", "b"}, Tags: []string{"x", "y"}, Nums: []int{1, 2}},
		parser, "a", "b", "--tag", "x", "--tag", "y", "-n", "1", "-n", "2")
	AssertSuccess(t, Out{Files: []string{}, Tags: []string{}, Nums: []int{}}, parser)
}

func TestConstructOptional(t *testing.T) {
	type Out struct {
		Speed Option[float64] `cli:"long=speed"`
		Name  Option[string]  // optional positional
	}
	parser := Construct[Out]().Parser().ToOptions()

	AssertSuccess(t, Out{}, parser)
	AssertSuccess(t,
		Out{Speed: Valid(1.5), Name: Valid("bob")},
		parser, "--speed", "1.5", "bob")

	assertFailure(t, parser,
		"couldn't parse `x`: strconv.ParseFloat: parsing \"x\": invalid syntax",
		"--speed", "x")
}

func TestConstructCount(t *testing.T) {
	type Out struct {
		Verbose Count `cli:"short=v,long=verbose"`
	}
	parser := Construct[Out]().Parser().ToOptions()

	for _, args := range [][]string{{"-v", "-v", "-v"}, {"-vvv"}, {"--verbose", "--verbose", "--verbose"}} {
		AssertSuccess(t, Out{Verbose: 3}, parser, args...)
	}
	AssertSuccess(t, Out{}, parser)
}

func TestConstructEnv(t *testing.T) {
	type Out struct {
		Home string `cli:"long=home,env=CLI_CONSTRUCT_TEST_HOME"`
	}
	parser := Construct[Out]().Parser().ToOptions()

	t.Setenv("CLI_CONSTRUCT_TEST_HOME", "/tmp/x")
	AssertSuccess(t, Out{Home: "/tmp/x"}, parser)
	AssertSuccess(t, Out{Home: "/override"}, parser, "--home", "/override")
}

func TestConstructMissingAggregated(t *testing.T) {
	type Out struct {
		Aaa string `cli:"long=aaa,metavar=A"`
		Bbb string `cli:"long=bbb,metavar=B"`
	}
	parser := Construct[Out]().Parser().ToOptions()

	_, err := parser.RunInner(NewArgs())
	require.NotNil(t, err)
	msg := UnwrapStderr(err)
	require.Contains(t, msg, "--aaa")
	require.Contains(t, msg, "--bbb")
}

func TestConstructParseError(t *testing.T) {
	type Out struct {
		N int `cli:"short=n"`
	}
	parser := Construct[Out]().Parser().ToOptions()

	_, err := parser.RunInner(NewArgs("-n", "12x"))
	require.NotNil(t, err)
	require.Contains(t, UnwrapStderr(err), "couldn't parse `12x`")
}

func TestConstructTextUnmarshaler(t *testing.T) {
	type Out struct {
		Port     constructTestPort `cli:"short=p,long=port"`
		Timeout  time.Duration     `cli:"long=timeout,default=10s"`
		Deadline time.Duration     `cli:"long=deadline"`
	}
	parser := Construct[Out]().Parser().ToOptions()

	AssertSuccess(t,
		Out{Port: 8080, Timeout: 10 * time.Second, Deadline: 90 * time.Second},
		parser, "-p", "8080", "--deadline", "1m30s")

	_, err := parser.RunInner(NewArgs("-p", "not-a-port"))
	require.NotNil(t, err)
}

func TestConstructAdjacentStructGroup(t *testing.T) {
	type Rect struct {
		Rect   bool    `cli:"long=rect"`
		Width  float64 `cli:"long=width,metavar=W"`
		Height float64 `cli:"long=height,metavar=H"`
	}
	parser := Construct[Rect]().Adjacent().Parser().ToOptions()

	AssertSuccess(t, Rect{Rect: true, Width: 10, Height: 20}, parser,
		"--rect", "--width", "10", "--height", "20")

	// group broken by an unconsumable item
	_, err := parser.RunInner(NewArgs("--rect", "--width", "10", "stop", "--height", "20"))
	require.NotNil(t, err)
}

func TestConstructSkipAndUntagged(t *testing.T) {
	type Out struct {
		Skip  int    `cli:"-"`
		Extra string // untagged: required positional
	}
	parser := Construct[Out]().Parser().ToOptions()
	AssertSuccess(t, Out{Extra: "hello"}, parser, "hello")
}

func TestConstructHelp(t *testing.T) {
	type Out struct {
		Debug bool    `cli:"short=d,long=debug,help='Activate debug mode'"`
		Speed float64 `cli:"long=speed,metavar=SPEED,default=42"`
		Src   string
	}
	parser := Construct[Out]().Parser().ToOptions()

	_, err := parser.RunInner(NewArgs("--help"))
	require.NotNil(t, err)
	out := UnwrapStdout(err)
	require.Contains(t, out, "-d, --debug")
	require.Contains(t, out, "Activate debug mode")
	require.Contains(t, out, "--speed=SPEED")
	require.Contains(t, out, "[default: 42]")
	require.Contains(t, out, "SRC")
}

func assertPanicsContains(t *testing.T, f func(), substr string) {
	t.Helper()
	defer func() {
		r := recover()
		require.NotNil(t, r, "expected panic containing %q", substr)
		msg, ok := r.(string)
		require.True(t, ok, "panic value must be a string, got %T", r)
		require.Contains(t, msg, substr)
	}()
	f()
}

func TestConstructPanics(t *testing.T) {
	type BadTagKey struct {
		X int `cli:"short=n,bogus=1"`
	}
	assertPanicsContains(t, func() { Construct[BadTagKey]() }, "unknown tag key")

	type BadShort struct {
		X int `cli:"short=ab"`
	}
	assertPanicsContains(t, func() { Construct[BadShort]() }, "short name must be one character")

	type BoolNoName struct {
		D bool
	}
	assertPanicsContains(t, func() { Construct[BoolNoName]() }, "bool field requires a short or a long name")

	type Nested struct {
		X [][]int
	}
	assertPanicsContains(t, func() { Construct[Nested]() }, "nested")

	type OptionWithDefault struct {
		X Option[int] `cli:"long=x,default=1"`
	}
	assertPanicsContains(t, func() { Construct[OptionWithDefault]() }, "cannot have a default")

	type SliceWithDefault struct {
		X []int `cli:"long=x,default=1"`
	}
	assertPanicsContains(t, func() { Construct[SliceWithDefault]() }, "cannot have a default")

	type Unsupported struct {
		X map[string]string `cli:"long=x"`
	}
	assertPanicsContains(t, func() { Construct[Unsupported]() }, "unsupported field type")

	type PosNamed struct {
		X string `cli:"positional,short=x"`
	}
	assertPanicsContains(t, func() { Construct[PosNamed]() }, "cannot be both positional and named")

	type CountNoName struct {
		V Count
	}
	assertPanicsContains(t, func() { Construct[CountNoName]() }, "cli.Count field requires a short or a long name")

	type Unexported struct {
		x int `cli:"long=x"`
	}
	assertPanicsContains(t, func() { Construct[Unexported]() }, "unexported")

	type BadDefault struct {
		X int `cli:"long=x,default=abc"`
	}
	assertPanicsContains(t, func() { Construct[BadDefault]() }, "bad default value")
}

func TestConstructQuotedHelp(t *testing.T) {
	type Out struct {
		Speed float64 `cli:"long=speed,help='Set speed, m/s'"`
	}
	parser := Construct[Out]().Parser().ToOptions()
	AssertSuccess(t, Out{Speed: 5}, parser, "--speed", "5")
}
