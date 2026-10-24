# Running and testing

## Running

A parser runs against `os.Args`:

```go
conf := parser.ToOptions().Run(os.Args)
```

`Run` either returns the parsed value or terminates the process: it prints the
error (or `--help` / `--version` output) and exits with an appropriate code. For
programs that want to handle errors themselves, `Parse` returns the value or an
error instead of exiting.

## Testing

Parsers are ordinary values, so tests feed `os.Args`-shaped slices into the
parser and assert on the result — no subprocesses or golden files are required:

```go
func TestRectangle(t *testing.T) {
	rect, err := rectangle.Parse([]string{"--width", "10", "--height", "5"})
	require.NoError(t, err)
	require.Equal(t, Rectangle{Width: 10, Height: 5}, rect)
}
```

Help output is verified with golden files in the repo, and the CI runs the
[lint](https://github.com/rprtr258/cli/blob/master/.github/workflows/ci.yml)
and test suites on every push to `master` and on pull requests.

## Where to go next

- [Examples](examples.md) — runnable programs covering most of the API.
