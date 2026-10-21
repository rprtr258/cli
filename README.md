# cli

Lightweight command line argument parser for Go with a combinator-style API.

A parser is assembled from small building blocks, composed into larger parsers, and
finally run against the command line. The result is a regular Go value - the library
either produces it or prints a message describing the problem (missing argument,
validation failure, `--help` request, etc).

## Install

```bash
go get github.com/rprtr258/cli
```

## Quick start

```go
package main

import (
	"fmt"
	"os"

	"github.com/rprtr258/cli"
	"github.com/rprtr258/fun"
)

type Rectangle struct {
	Width, Height int
}

func rectangle() cli.Parser[Rectangle] {
	invalidSize := "Sides of a rectangle must be 1..20 units long"
	invalidArea := "Area of a rectangle must not exceed 200 units square"
	width := cli.Argument[int](cli.Long("width").Help("Width of the rectangle"), "PX", parseInt).
		Guard(func(x int) bool { return 1 <= x && x <= 20 }, invalidSize)
	height := cli.Argument[int](cli.Long("height").Help("Height of the rectangle"), "PX", parseInt).
		Guard(func(x int) bool { return 1 <= x && x <= 20 }, invalidSize)
	return cli.Pair(width, height).
		Map(func(wh fun.Pair[int, int]) Rectangle {
			return Rectangle{Width: wh.K, Height: wh.V}
		}).
		Guard(func(r Rectangle) bool { return r.Width*r.Height <= 200 }, invalidArea)
}

func main() {
	rect := rectangle().ToOptions().Run(os.Args)
	fmt.Println(rect)
}
```

```console
$ app --width 10 --height 5
{10 5}
$ app --width 0
Error:  --width=P: Sides of a rectangle must be 1..20 units long
```

## Struct-based declaration

For plain option structs `Construct` builds the whole parser from `cli` struct
tags - no manual wiring required:

```go
type Out struct {
	Debug   bool    `cli:"short=d,long=debug,help='Activate debug mode'"`
	Speed   float64 `cli:"long=speed,metavar=SPEED,default=42"`
	Home    string  `cli:"long=home,env=HOME"`
	Files   []string
	Verbose cli.Count `cli:"short=v,long=verbose"`
}

out := cli.Construct[Out]().ToOptions().Run(os.Args)
```

Field types choose the consumers: `bool` - a switch, primitives - an argument
(or a positional if the field has no name), `[]E` - zero or more, `cli.Opt[E]` -
optional, `cli.Count` - number of repeats.

## Ways to consume data

| Builder | Consumes | Produces |
|-|-|-|
| `Short('f')` / `Long("flag")` / `Env("VAR")` | - | `NamedArg` to build on |
| `NamedArg.Switch` | `-f` / `--flag` | `Parser[bool]` |
| `Flag` | `-f` / `--flag` | `Parser[T]`, one of two values |
| `ReqFlag` | `-f` (required) | `Parser[T]` |
| `Argument` | `-f foo`, `--flag=bar` | `Parser[T]` |
| `Positional` | a free-standing value | `Parser[T]` |
| `Any` | any unconsumed item | `Parser[T]` |
| `Command("cmd")` | a subcommand | independent subparser |

## Transforming parsers

- `Parser.Map` / `Parse` - convert the produced value
- `Guard` - validate the value with a custom check and error message
- `Optional` / `OptionalCatch` - make a parser optional
- `Many` / `Some` / `ManyCatch` / `SomeCatch` - collect repeated items into a slice
- `Collect` - collect repeated items into a slice
- `Count` - count the number of occurrences of a flag
- `Fallback` / `FallbackWith` - provide default values

## Composing parsers

- `Pair` / `Triple` - sequential composition of two or three parsers
- `Choice` - parallel composition: the first parser that consumes the leftmost value wins
- `Construct` - struct composition from `cli` tags
- `Parser.Adjacent` - restrict a composed group to adjacent command line items

## Improving user experience

- `Help` - document flags, arguments and positional items
- `Hide`, `HideUsage`, `GroupHelp` - control what is shown in `--help`
- `Description`, `Header`, `Footer`, `Usage`, `WithUsage`, `MaxWidth` - shape the help output
- `Version` - add a `--version` switch
- `HelpParser`, `VersionParser` - customize the flags that trigger help and version

## Shell completion

Flag and value names are completed out of the box through the dynamic completion
protocol of the shell. Register the stub once and both bash, zsh and fish will
query your binary:

```bash
source <(your-binary --bpaf-complete-style-bash)
```

Custom dynamic completion is available with `Parser.CompleteShell` for
files/directories and raw values, and with `Parser.Complete` for fully custom
completion logic.

## Running the parser

`ToOptions` finalizes a parser into an `OptionParser`. `OptionParser.Run` parses
`os.Args` (or any argument list) and returns the value, exiting the process on
`--help`, `--version` or parse failures. `OptionParser.RunE` returns an error
instead of exiting, which is useful for tests and libraries:

```go
value, err := parser.RunE([]string{"prog", "-vvv"})
```

## Testing

`CheckInvariants` panics with a descriptive message when a parser is built with
violated invariants, such as positional items or commands placed before named
options. Call it in tests to catch misconfiguration early:

```go
rectangle().ToOptions().CheckInvariants(false)
```

## Examples

The [examples](examples/) directory contains runnable programs covering most of
the API: subcommands, environment variables, adjacent groups, dynamic completion,
custom help output and more.
