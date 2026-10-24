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
- **Fast.** ~1 µs and a handful of allocations to parse a 3-flag command —
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
| `NamedArg.Flag` | `-f` / `--flag` | `Parser[T]`, one of two values |
| `NamedArg.ReqFlag` | `-f` (required) | `Parser[T]` |
| `NamedArg.Argument` | `-f foo`, `--flag=bar` | `Parser[T]` |
| `Positional` | a free-standing value | `Parser[T]` |
| `Any` | any unconsumed item | `Parser[T]` |
| `OptionParser.Command("cmd")` | a subcommand | `ParseCommand[T]`, an independent subparser |

`NamedArg.Argument` and `Positional` take a `func(string) (T, error)` to convert the raw
text. Ready-made adapters for common types are provided: `ParseString`,
`ParseInt`, `ParseInt32`, `ParseUint`, `ParseFloat`, `ParseBool`, `ParseDuration`.

## Transforming parsers

- `Parser.Map` - convert the produced value
- `Parser.Guard` - validate the value with a custom check and error message
- `Parser.Optional` / `Parser.OptionalCatch` - make a parser optional
- `Parser.Many` / `Parser.Some` / `Parser.ManyCatch` / `Parser.SomeCatch` - collect repeated items into a slice
- `Parser.Collect` - collect repeated items into a slice
- `Parser.Count` - count the number of occurrences of a flag
- `Parser.Fallback` / `Parser.FallbackWith` - provide default values

## Composing parsers

- `Map2`-`Map6` - sequential composition: combine 2 to 6 parsers into the value their function returns
- `Construct` - struct composition from `cli` tags
- `Pair` / `Triple` - like `Map2`/`Map3`, but producing `fun.Pair`/`Tuple` values
- `Choice` - parallel composition: the first parser that consumes the leftmost value wins
- `Adjacent` - restrict a composed group (`Pair`, `Triple`), an argument or a command to adjacent command line items

## Improving user experience

- `Help` - document flags, arguments and positional items
- `Parser.Hide`, `Parser.HideUsage`, `Parser.GroupHelp` - control what is shown in `--help`
- `OptionParser.Description`, `OptionParser.Header`, `OptionParser.Footer`, `OptionParser.Usage`, `OptionParser.WithUsage`, `OptionParser.MaxWidth` - shape the help output
- `OptionParser.Example(descr, command)` - document a usage example; examples are shown in the `Examples:` section of `--help` and in the `EXAMPLES` section of the manpage
- `OptionParser.Version` - add a `--version` switch
- `OptionParser.HelpParser`, `OptionParser.VersionParser` - customize the flags that trigger help and version

Documented examples are validated whenever the parser runs: each example
command line is parsed with the parser it is declared on, so an example that
drifts from the parser panics with a `cli usage BUG` message instead of
shipping stale documentation:

```go
parser.ToOptions().
    Description("Copy files and directories").
    Example("Copy a file to a directory", "file.txt backup/").
    Example("Copy a directory recursively", "--recursive src/ dst/")
```

```console
$ app --help
Copy files and directories

Usage: app [-r] SRC DST

<skip>

Examples:
    Copy a file to a directory  $ app file.txt backup/
    Copy a directory recursively  $ app --recursive src/ dst/
```

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
options, and when a declared usage example no longer parses. Call it in tests
to catch misconfiguration early:

```go
rectangle.ToOptions().CheckInvariants()
```

## Examples

The [examples](examples/) directory contains runnable programs covering most of
the API: subcommands, environment variables, adjacent groups, dynamic completion,
custom help output and more.
