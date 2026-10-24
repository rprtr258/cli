# Struct-based declaration

Struct-based declaration lets you describe the command line as a plain Go struct.
Field types choose the consumers:

| Field type | Meaning |
|-|-|
| `bool` | a switch |
| primitives | an argument (a positional if the field has no name) |
| `[]E` | zero or more values |
| `cli.Opt[E]` | optional value |
| `cli.Count` | number of repeats |

Non-trivial options — short names, help text, env fallbacks, defaults — go into the
`cli` struct tag:

```go
type Out struct {
	Debug bool   `cli:"short=d,long=debug,help='Activate debug mode'"`
	Home  string `cli:"long=home,env=HOME"`
}
```

Turn a struct into a parser with `Construct`:

```go
conf := cli.Construct[Config]().ToOptions().Run(os.Args)
```

This is the same runtime as the [combinator declaration](combinator.md), just with
the layout written as types and tags instead of builder calls.

## Where to go next

- [Ways to consume data](consumers.md) — what each field type maps to in the combinator API.
- [Help and UX](help.md) — documenting what you declared.
