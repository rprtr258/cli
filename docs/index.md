# rprtr258/cli

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

## Install

Requires Go 1.27 or newer.

```bash
go get github.com/rprtr258/cli
```

## Where to go next

- [Combinator declaration](combinator.md) and [struct-based declaration](struct.md) — the two ways to describe your command line.
- [Ways to consume data](consumers.md) — switches, flags, arguments, positionals, subcommands.
- [Transforming parsers](transformers.md) and [composing parsers](composers.md) — build bigger parsers out of smaller ones.
- [Help and UX](help.md), [shell completion](completion.md), [running and testing](running.md).
- The [examples](examples.md) directory contains runnable programs covering most of the API.
