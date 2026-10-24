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

## Markdown and man pages

A parser can be rendered to Markdown and to roff (man page source) — the [README
of this repo](https://github.com/rprtr258/cli) uses the built-in Markdown
renderer, and the `roff` package produces manpages. Both are generated from the
same declaration the program runs on, so they can never disagree with the actual
behavior.

## Help formatting

`--help` prints the usage line, options with their types, metavariables,
defaults and environment fallbacks. The exact layout is [covered by golden
tests](running.md) in the repo.

## Where to go next

- [Shell completion](completion.md) — autocomplete for bash, zsh, fish and elvish.
