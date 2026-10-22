# AGENTS.md

## Verification commands

Run all checks and formatting through the make targets:

- `make test` — full test suite; exits 0 = done.
- `make lint` — golangci-lint, includes the gofumpt formatting check.
- `make fmt` — apply gofumpt formatting via golangci-lint.

Invoke formatters and vetters exclusively through these targets. The reason: the
`gofmt`/`go` binaries on the default PATH are an older toolchain than `go.mod`
requires and reject the generic methods this library uses, while `golangci-lint`
is built with the correct toolchain.
