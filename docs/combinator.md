# Combinator declaration

A parser is assembled from small building blocks, composed into larger parsers, and
finally run against the command line. The result is a regular Go value — the library
either produces it or prints a message describing the problem (missing argument,
validation failure, `--help` request, etc).

```go
package main

import (
	"fmt"
	"os"

	"github.com/rprtr258/cli"
)

type Rectangle struct {
	Width, Height int
}

const invalidSize = "Sides of a rectangle must be 1..20 units long"

var rectangle cli.Parser[Rectangle] = cli.Map2(
	cli.Long("width").
		Help("Width of the rectangle").
		Argument("PX", cli.ParseInt).
		Guard(func(x int) bool { return 1 <= x && x <= 20 }, invalidSize),
	cli.Long("height").
		Help("Height of the rectangle").
		Argument("PX", cli.ParseInt).
		Guard(func(x int) bool { return 1 <= x && x <= 20 }, invalidSize),
	func(w, h int) Rectangle {
		return Rectangle{Width: w, Height: h}
	}).
	Guard(func(r Rectangle) bool { return r.Width*r.Height <= 200 }, "Area of a rectangle must not exceed 200 units square")

func main() {
	rect := rectangle.ToOptions().Run(os.Args)
	fmt.Println(rect)
}
```

```console
$ app --width 10 --height 5
{10 5}
$ app --width 0
parseError: `0`: Sides of a rectangle must be 1..20 units long
```

The declaration is fully checked by the compiler: `Map2` only accepts parsers of the
types the combining function expects, and `Guard` reuses the same value type, so
mistakes surface at build time.

## Where to go next

- [Struct-based declaration](struct.md) — describe the same thing with a struct and `cli` tags.
- [Ways to consume data](consumers.md) — the full catalog of consumers to build on.
