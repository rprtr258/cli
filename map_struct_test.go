package cli

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMapStruct2(t *testing.T) {
	width := Long("width").Argument("PX", ParseInt)
	height := Long("height").Argument("PX", ParseInt)

	type Rect struct {
		Width  int
		Height int
	}
	parser := MapStruct2[Rect](width, height).ToOptions()

	assertSuccess(t, Rect{Width: 10, Height: 5}, parser, "--width 10 --height 5")
	assertSuccess(t, Rect{Width: -3, Height: 0}, parser, "--width=-3 --height=0")
	assertFailure(t, parser, "expected `--height=PX`, pass `--help` for usage information", "--width 10")
}

func TestMapStruct3(t *testing.T) {
	x := Long("x").Argument("X", ParseInt)
	y := Long("y").Argument("Y", ParseInt)
	z := Long("z").Argument("Z", ParseInt)

	type Point struct{ X, Y, Z int }
	parser := MapStruct3[Point](x, y, z).ToOptions()

	assertSuccess(t, Point{X: 1, Y: 2, Z: 3}, parser, "--x 1 --y 2 --z 3")
}

func TestMapStruct6(t *testing.T) {
	type Box struct{ MinX, MinY, MinZ, MaxX, MaxY, MaxZ int }

	parser := MapStruct6[Box](
		Long("min-x").Argument("X", ParseInt),
		Long("min-y").Argument("Y", ParseInt),
		Long("min-z").Argument("Z", ParseInt),
		Long("max-x").Argument("X", ParseInt),
		Long("max-y").Argument("Y", ParseInt),
		Long("max-z").Argument("Z", ParseInt),
	).ToOptions()

	assertSuccess(
		t, Box{MinX: 1, MinY: 2, MinZ: 3, MaxX: 4, MaxY: 5, MaxZ: 6}, parser,
		"--min-x 1 --min-y 2 --min-z 3 --max-x 4 --max-y 5 --max-z 6",
	)
}

func TestMapStructPanics(t *testing.T) {
	width := Long("width").Argument("PX", ParseInt)
	height := Long("height").Argument("PX", ParseInt)

	type NotAStruct int
	require.PanicsWithValue(t,
		"cli.MapStruct2[cli.NotAStruct]: type parameter must be a struct, got int",
		func() { MapStruct2[NotAStruct](width, height) },
	)

	type OneField struct{ Width int }
	require.PanicsWithValue(t,
		"cli.MapStruct2[cli.OneField]: want struct with 2 fields, got 1",
		func() { MapStruct2[OneField](width, height) },
	)

	type ThreeFields struct{ X, Y, Z int }
	require.PanicsWithValue(t,
		"cli.MapStruct2[cli.ThreeFields]: want struct with 2 fields, got 3",
		func() { MapStruct2[ThreeFields](width, height) },
	)

	type HasPrivate struct {
		Width  int
		height int
	}
	require.PanicsWithValue(t,
		"cli.MapStruct2[cli.HasPrivate]: field \"height\" is unexported, export it to fill it with a parser value",
		func() { MapStruct2[HasPrivate](width, height) },
	)

	type WrongType struct {
		Width  string
		Height int
	}
	require.PanicsWithValue(t,
		"cli.MapStruct2[cli.WrongType]: field \"Width\": want string, got parser producing int",
		func() { MapStruct2[WrongType](width, height) },
	)
}
