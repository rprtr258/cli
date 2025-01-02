package cli

import (
	"testing"

	"github.com/rprtr258/fun"
	"github.com/stretchr/testify/require"
)

type Poly[T any] struct {
	field T
}

func poly[T any](name string, parse func(string) (T, error)) Parser[Poly[T]] {
	field := Argument(Long(name), "ARG", parse)
	return Map(field.Parser(), func(a T) Poly[T] { return Poly[T]{a} })
}

func TestGeneric_argument_field(t *testing.T) {
	a := Optional(poly("uint", ParseUint), false)
	b := Optional(poly("uint32", ParseUint32), false)
	parser := Pair(a, b).Parser().ToOptions()

	type R = fun.Pair[Option[Poly[uint]], Option[Poly[uint32]]]
	{
		r, err := parser.run_inner(newArgs("--uint", "12"))
		require.Nil(t, err)
		require.Equal(t, r, R{Valid(Poly[uint]{field: 12}), Invalid[Poly[uint32]]()})
	}
	{
		r, err := parser.run_inner(newArgs("--uint32", "12"))
		require.Nil(t, err)
		require.Equal(t, r, R{Invalid[Poly[uint]](), Valid(Poly[uint32]{field: 12})})
	}
	{
		r, err := parser.run_inner(newArgs("--uint32", "12", "--uint", "24"))
		require.Nil(t, err)
		require.Equal(t, r, R{Valid(Poly[uint]{field: 24}), Valid(Poly[uint32]{field: 12})})
	}
}
