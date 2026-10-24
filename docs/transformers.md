# Transforming parsers

Once a consumer produces a `Parser[T]`, the chain below changes how it is parsed. The
return type in brackets matters: some transformers only exist on parsers of some
types, e.g. `Some` only makes sense on repeatable things and `Def` is not
applicable to switches, which are always present.

| Transformer | Signature | What it does |
|-|-|-|
| `Guard` | `Parser[T] -> Parser[T]` | Check the parsed value is valid: `.Guard(func(t T) bool, "message")` |
| `Def` | `Parser[T] -> Parser[T]` | Fall back when missing: `.Def(x)` |
| `Map` | `Parser[T] -> Parser[U]` | Pure transform of the value: `.Map(func(t T) U)` |
| `Map2` | `(Parser[T], Parser[U]) -> Parser[V]` | Pure transform of two parsers |
| `Some` | `Parser[T] -> Parser[[]T]` | One or more values |
| `Many` | `Parser[T] -> Parser[[]T]` | Zero or more values |
| `Optional` | `Parser[T] -> Parser[Opt[T]]` | Zero or one value |
| `Count` | `Parser[T] -> Parser[int]` | Number of repeats, ignores the values |
| `DefaultsTo` | `Parser[T] -> Parser[Ext[T]]` | Tells which values were explicitly passed and which are default |
| `SetEnv` | `Parser[T] -> Parser[T]` | Take the value from the given environment variable if the command line does not provide it |

Missing required arguments and failed `Guard`s produce user-readable errors
automatically, listing all valid ways to use the parser.

## Where to go next

- [Composing parsers](composers.md) — combine several `Parser[T]` into one.
