# cli

A command line argument parser for Go where a parser is an ordinary `Parser[T]`
value, composed from small building blocks with generics — so option types and
validation are checked by the Go compiler, not by reflection at runtime. The
parser you build drives everything: argument parsing, `--help` output, shell
completion and even manpage rendering — you declare the command line once.

- **Type-safe.** Parsers are generic `Parser[T]` values; composition is
  checked at compile time, no struct-tag reflection required.
- **One source of truth.** `--help` text, bash/zsh/fish/elvish completion and
  roff manpages are all generated from the same parser definition.
- **One dependency.** Just [`github.com/rprtr258/fun`](https://github.com/rprtr258/fun); everything else is stdlib.
- **Fast.** ~0.8 µs and 9 allocations to parse a 3-flag command (i7-1165G7) —
  negligible for a once-per-process task.
- **Two declaration styles.** Combinator builders or `cli` struct tags via
  `Construct` — same runtime underneath.

```go
type Config struct {
	Verbose cli.Count `cli:"short=v,long=verbose"`
	Speed   float64   `cli:"long=speed,metavar=SPEED,default=42"`
	Files   []string
}

conf := cli.Construct[Config]().ToOptions().Run(os.Args)
```

```console
$ app -vv --speed 3.14 a.txt b.txt
{2 3.14 [a.txt b.txt]}
```

## Contents

- [Install](#install)
- [Combinator declaration](#combinator-declaration)
- [Struct-based declaration](#struct-based-declaration)
- [Ways to consume data](#ways-to-consume-data)
- [Transforming parsers](#transforming-parsers)
- [Composing parsers](#composing-parsers)
- [Improving user experience](#improving-user-experience)
- [Shell completion](#shell-completion)
- [Running the parser](#running-the-parser)
- [Testing](#testing)
- [Examples](#examples)

## Install

Requires Go 1.27 or newer.

```bash
go get github.com/rprtr258/cli
```

## Combinator declaration

A parser is assembled from small building blocks, composed into larger parsers, and
finally run against the command line. The result is a regular Go value - the library
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

func rectangle() cli.Parser[Rectangle] {
	invalidSize := "Sides of a rectangle must be 1..20 units long"
	invalidArea := "Area of a rectangle must not exceed 200 units square"
	width := cli.Argument(cli.Long("width").Help("Width of the rectangle"), "PX", cli.ParseInt).
		Guard(func(x int) bool { return 1 <= x && x <= 20 }, invalidSize)
	height := cli.Argument(cli.Long("height").Help("Height of the rectangle"), "PX", cli.ParseInt).
		Guard(func(x int) bool { return 1 <= x && x <= 20 }, invalidSize)
	return cli.Map2(width, height, func(w, h int) Rectangle {
		return Rectangle{Width: w, Height: h}
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
parseError: `0`: Sides of a rectangle must be 1..20 units long
```

## Struct-based declaration

Field types choose the consumers: `bool` - a switch, primitives - an argument
(or a positional if the field has no name), `[]E` - zero or more, `cli.Opt[E]` -
optional, `cli.Count` - number of repeats. Non-trivial options (short names,
help text, env fallbacks, defaults) go into the `cli` struct tag:

```go
type Out struct {
	Debug bool   `cli:"short=d,long=debug,help='Activate debug mode'"`
	Home  string `cli:"long=home,env=HOME"`
}
```

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

`Argument` and `Positional` take a `func(string) (T, error)` to convert the raw
text. Ready-made adapters for common types are provided: `ParseString`,
`ParseInt`, `ParseUint`, `ParseFloat`, `ParseBool`, `ParseDuration`.

## Transforming parsers

- `Parser.Map` - convert the produced value
- `Guard` - validate the value with a custom check and error message
- `Parser.Optional` / `Parser.OptionalCatch` - make a parser optional
- `Parser.Many` / `Parser.Some` / `Parser.ManyCatch` / `Parser.SomeCatch` - collect repeated items into a slice
- `Parser.Collect` - collect repeated items into a slice
- `Parser.Count` - count the number of occurrences of a flag
- `Fallback` / `FallbackWith` - provide default values

## Composing parsers

- `Map2`-`Map6` - sequential composition: combine 2 to 6 parsers into the value their function returns
- `Construct` - struct composition from `cli` tags
- `Pair` / `Triple` - like `Map2`/`Map3`, but producing `fun.Pair`/`Tuple` values
- `Choice` - parallel composition: the first parser that consumes the leftmost value wins
- `Parser.Adjacent` - restrict a composed group to adjacent command line items

## Improving user experience

- `Help` - document flags, arguments and positional items
- `Hide`, `HideUsage`, `GroupHelp` - control what is shown in `--help`
- `Description`, `Header`, `Footer`, `Usage`, `WithUsage`, `MaxWidth` - shape the help output
- `Version` - add a `--version` switch
- `HelpParser`, `VersionParser` - customize the flags that trigger help and version

## Shell completion

Flag and value names are completed out of the box through the dynamic completion
protocol of the shell. Register the stub once and bash, zsh, fish and elvish will
query your binary:

```bash
source <(your-binary --cli-complete-style-bash)
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
