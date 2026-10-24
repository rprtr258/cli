# Shell completion

Completion for bash, zsh, fish and elvish shells is generated from the parser
declaration. Pass a shell name to get a completion script:

```go
import "github.com/rprtr258/cli"

bashCompletion := cli.CompleteBash(myParser)
zshCompletion := cli.CompleteZsh(myParser)
fishCompletion := cli.CompleteFish(myParser)
elvishCompletion := cli.CompleteElvish(myParser)
```

Completion is generated from the same source of truth as `--help`, so new flags
appear in the shell as soon as they exist in the parser — nothing to keep in
sync by hand.

## Where to go next

- [Help and UX](help.md) — help output and manpages.
