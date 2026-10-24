# Examples

The [examples directory](https://github.com/rprtr258/cli/tree/master/examples)
in the repo contains runnable programs covering most of the API — each file is
adapted from a clap.rs example with the same name. A few entry points:

- [`basic.go`](https://github.com/rprtr258/cli/blob/master/examples/basic.go) — minimal flag with help.
- [`rectangle.go`](https://github.com/rprtr258/cli/blob/master/examples/rectangle.go) — two validated arguments composed with `Map2` and `Guard`.
- [`construct_struct.go`](https://github.com/rprtr258/cli/blob/master/examples/construct_struct.go) — struct-based declaration.
- [`git.go`](https://github.com/rprtr258/cli/blob/master/examples/git.go) — subcommands.
- [`env_variable.go`](https://github.com/rprtr258/cli/blob/master/examples/env_variable.go) — fallback to environment variables.
- [`customize_help.go`](https://github.com/rprtr258/cli/blob/master/examples/customize_help.go) — appname, description, copyright, version.

Run any of them from the repo root:

```bash
go run ./examples/rectangle --width 10 --height 5
```
