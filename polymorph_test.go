package cli

import (
	"testing"

	"github.com/rprtr258/fun"
)

type Poly[T any] struct {
	field T
}

func poly[T any](name string, parse func(string) (T, error)) Parser[Poly[T]] {
	field := Argument(Long(name), "ARG", parse)
	return field.Map(func(a T) Poly[T] { return Poly[T]{a} })
}

func TestGeneric_argument_field(t *testing.T) {
	a := poly("uint", ParseUint).Optional()
	b := poly("uint32", parseUint32).Optional()
	parser := Pair(a, b).ToOptions()

	type R = fun.Pair[Opt[Poly[uint]], Opt[Poly[uint32]]]
	assertSuccess(t, R{Valid(Poly[uint]{field: 12}), Invalid[Poly[uint32]]()}, parser, "--uint", "12")
	assertSuccess(t, R{Invalid[Poly[uint]](), Valid(Poly[uint32]{field: 12})}, parser, "--uint32", "12")
	assertSuccess(t, R{Valid(Poly[uint]{field: 24}), Valid(Poly[uint32]{field: 12})}, parser, "--uint32", "12", "--uint", "24")
}
