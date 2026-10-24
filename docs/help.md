# Help and UX

## Error messages

Missing arguments and failed `Guard` checks produce helpful errors that list all
valid ways to use the parser:

```console
$ app --user
parseError: missing required argument: --user=USER
$ app --width 0
parseError: `0`: Sides of a rectangle must be 1..20 units long
```

## Customizing help output

Use these methods on `OptionParser` before `.Run`:

| Method | Effect |
|-|-|
| `Description(...)` | Program description at the top of help |
| `Appname(...)` | Program name instead of `os.Args[0]` |
| `Version(...)` | Print `--version` output and exit |
| `Copyright(...)` | Copyright header at the top of help |
| `Example(descr, command)` | Documented usage example in help and manpage |

## Examples

`Example(descr, command)` documents a usage example. Examples are shown in the
`Examples:` section of the `--help` output and in the `EXAMPLES` section of
rendered manpages:

```go
parser.ToOptions().
    Description("Copy files and directories").
    Example("Copy a file to a directory", "file.txt backup/").
    Example("Copy a directory recursively", "--recursive src/ dst/")
```

```console
$ app --help
Copy files and directories

Usage: app [-r] SRC DST

<skip>

Examples:
    Copy a file to a directory  $ app file.txt backup/
    Copy a directory recursively  $ app --recursive src/ dst/
```

The command line is written relative to the parser the example is declared on:
without the program name and, for subcommands, without the command name — the
program name and command path are prepended automatically when rendering.

Examples are validated whenever the parser runs: the command line is tokenized
(shell-style, quotes are supported) and parsed with the parser the example is
declared on. If an example stops parsing after a parser change, `Run`, `RunE`
and `CheckInvariants` panic with a `cli usage BUG` message naming the stale
example, so the documentation cannot silently disagree with the parser. This
also executes the example converters, so example command lines must not have
side effects.

## Man pages

The `roff` package renders manpage source from the same declaration the program
runs on, so the manpage can never disagree with the actual behavior. Declared
[examples](#examples) are rendered into the `EXAMPLES` section. Export to
Markdown is planned (see the roadmap comment in `doc.go`) but not implemented
yet.

## Help formatting

`--help` prints the usage line, options with their types, metavariables,
defaults and environment fallbacks. The exact layout is [covered by golden
tests](running.md) in the repo.

## Where to go next

- [Shell completion](completion.md) — autocomplete for bash, zsh, fish and elvish.
