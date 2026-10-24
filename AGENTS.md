## Verification commands

Run all checks and formatting through the make targets:

- `make test` — full test suite; exits 0 = done.
- `make bench` — run benchmarks (`go test -bench=. -benchmem -run='^$'`).
- `make lint` — golangci-lint, includes the gofumpt formatting check.
- `make fmt` — apply gofumpt formatting via golangci-lint.

Invoke formatters and vetters exclusively through these targets.

## Documentation stays in sync with the API

When changing user-facing behavior — new, removed or changed `OptionParser`/
`Parser` methods, help or manpage output, validation or error behavior — update
the docs in the same change, not as a follow-up:

- `README.md` (feature bullets and code/console snippets)
- `docs/*.md` (mkdocs site; e.g. methods table and sections in `docs/help.md`,
  testing notes in `docs/running.md`, runnable examples list in
  `docs/examples.md`)
- godoc comments on the touched declarations (keep the `# Usage` / `# Example`
  / `# Panics` conventions used across the codebase)

Console snippets in docs must match real output; generate them by running the
parser instead of writing them from memory. Then run `make test` and
`make lint` as usual.

## Generic methods: self-instantiating return types

A method on `Parser[T]` must not return `Parser[X[T]]` directly (e.g.
`Parser[Opt[T]]` or `Parser[[]T]`): the go1.27 compiler rejects the definition
itself with `instantiation cycle: T instantiated as Opt[T]`, even when the
method is never called.

Fix with a fresh type parameter whose constraint is the exact type, and convert
the inner result — the singleton constraint drives inference, so call sites stay
argument-free (`x.Optional()`):

```go
func (self Parser[T]) Optional[R Opt[T]]() Parser[R] {
	inner := self.Parser()
	return Parser[R]{
		eval: func(state *state) (R, parseError) {
			x, err := optional_eval(inner, false, state)
			return R(x), err
		},
		meta: metaOptional{inner.meta},
	}
}
```

Precedent in this codebase: `Parser.Many[R []T]`, `Parser.Some[R []T]`,
`Parser.Optional[R Opt[T]]`, `Parser.OptionalCatch[R Opt[T]]`.
