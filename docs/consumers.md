# Ways to consume data

| Builder | Consumes | Produces |
|-|-|-|
| `Short('f')` / `Long("flag")` / `Env("VAR")` | - | `NamedArg` to build on |
| `NamedArg.Switch` | `-f` / `--flag` | `Parser[bool]` |
| `NamedArg.Flag` | `-f` / `--flag` | `Parser[T]`, one of two values |
| `NamedArg.ReqFlag` | `-f` (required) | `Parser[T]` |
| `NamedArg.Argument` | `-f foo`, `--flag=bar` | `Parser[T]` |
| `NamedArg.Map` | `--flag=KEY:VALUE`, repeatable | `Parser[map[K]V]` |
| `Positional` | a free-standing value | `Parser[T]` |
| `Any` | any unconsumed item | `Parser[T]` |
| `OptionParser.Command("cmd")` | a subcommand | `ParseCommand[T]`, an independent subparser |

`NamedArg.Argument` and `Positional` take a `func(string) (T, error)` to convert the raw
text. Ready-made adapters for common types are provided: `ParseString`,
`ParseInt`, `ParseInt32`, `ParseUint`, `ParseFloat`, `ParseBool`, `ParseDuration`.

## Where to go next

- [Transforming parsers](transformers.md) — validate, default and repeat what a consumer produced.
- [Composing parsers](composers.md) — combine consumers into one value.
