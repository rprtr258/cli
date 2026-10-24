package cli

import (
	"flag"
	"testing"

	"github.com/stretchr/testify/require"
)

type benchOut struct {
	n     uint
	speed float64
	out   string
}

var benchArgs = []string{"prog", "-n", "42", "-s", "3.14", "-o", "out.txt"}

func BenchmarkCli(b *testing.B) {
	options := Map3(
		Short('n').Argument("NUM", ParseUint),
		Short('s').Argument("SPEED", ParseFloat),
		Short('o').Argument("OUT", ParseString),
		func(n uint, speed float64, out string) benchOut {
			return benchOut{n, speed, out}
		},
	).ToOptions()

	b.ReportAllocs()
	for b.Loop() {
		out, err := options.RunE(benchArgs)
		require.NoError(b, err)
		require.Equal(b, benchOut{42, 3.14, "out.txt"}, out)
	}
}

func BenchmarkFlag(b *testing.B) {
	fs := flag.NewFlagSet("prog", flag.ContinueOnError)
	n := fs.Uint("n", 0, "n")
	speed := fs.Float64("s", 0, "s")
	out := fs.String("o", "", "o")

	b.ReportAllocs()
	for b.Loop() {
		require.NoError(b, fs.Parse(benchArgs[1:]))
		require.Equal(b, *n, uint(42))
		require.Equal(b, *speed, 3.14)
		require.Equal(b, *out, "out.txt")
	}
}
