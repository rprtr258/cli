# Composing parsers

Composition combines parsers of different types into one:

| Composer | Signature |
|-|-|
| `Map2(f)` | two parsers, `(T, U) -> V` |
| `Map3(f)` | three parsers, `(T, U, S) -> V` |
| `Map4(f)` | four parsers, `(T, U, S, R) -> V` |
| `Sequence(parsers...)` | any parsers, `[]any -> V` |

Use `Map2`–`Map4` when the shape is known — they stay fully type-checked. Use
`Sequence` for arbitrary many pieces when the result can be assembled from `any`
values.

Both styles coexist with [struct-based declaration](struct.md): build pieces with
combinators, then feed them into a struct or combine into one `Parser[T]`.

## Where to go next

- [Help and UX](help.md) — document the composed command line.
