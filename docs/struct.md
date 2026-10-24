# Struct-based declaration

Struct-based declaration lets you describe the command line as a plain Go struct.
Field types choose the consumers:

| Field type | Meaning |
|-|-|
| `bool` | a switch |
| primitives | an argument (a positional if the field has no name) |
| `[]E` | zero or more values |
| `[]*E` | zero or more values, elements are pointers |
| `*E` | optional value, `nil` when absent |
| `cli.Opt[E]` | optional value |
| `cli.Count` | number of repeats |
| `map[K]V` | repeated `KEY:VALUE` named argument accumulating into a map; entries split at the first colon |
| struct with `command=NAME` | subcommand: inner fields are parsed only when the command name is given |
| struct with `group=NAME` | option group: inner fields parse as usual options but are rendered in help under their own section |

Non-trivial options — short names, help text, env fallbacks, defaults — go into the
`cli` struct tag:

```go
type Out struct {
	Debug bool   `cli:"short=d,long=debug,help='Activate debug mode'"`
	Home  string `cli:"long=home,env=HOME"`
	Tags  map[string]string `cli:"long=tag,help='Tags as KEY:VALUE'"`
	Pet   string `cli:"long=pet,choice=cat,choice=dog"`
	HideMe string `cli:"long=hide-me,hidden"`
}
```

Supported tag keys:

| Tag | Meaning |
|-|-|
| `short=a` | short name, one character; repeatable for hidden aliases |
| `long=name` | long name; repeatable for hidden aliases |
| `metavar=X` | placeholder shown in help, defaults to the uppercased field name |
| `help=...` | help message |
| `env=VAR` | environment variable fallback |
| `default=v` | value used when the field is absent; not valid for bool, slice, pointer, `cli.Opt[E]` and `cli.Count`; on map fields repeatable, each entry is `key:value` |
| `positional` | forces positional parsing; fields without short/long/env are positional automatically |
| `command=NAME` | struct field is a subcommand with the given name |
| `alias=A` | hidden long alias for a command field, repeatable |
| `group=NAME` | struct field options are rendered in help under their own section |
| `hidden` | parse the field but omit it from help and completion |
| `choice=V` | allowed argument values, repeatable; a value not listed fails parsing |

Named fields are parsed first, then command fields, then positionals, regardless of
declaration order. A nested struct field requires either `command=` or `group=` —
a plain nested struct panics.

Turn a struct into a parser with `Construct`:

```go
conf := cli.Construct[Config]().ToOptions().Run(os.Args)
```

This is the same runtime as the [combinator declaration](combinator.md), just with
the layout written as types and tags instead of builder calls.

## Where to go next

- [Ways to consume data](consumers.md) — what each field type maps to in the combinator API.
- [Help and UX](help.md) — documenting what you declared.
