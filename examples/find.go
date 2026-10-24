// `find.go` - this is not a typical usage, but you should be able to replicate the
// command line interface used by find.
package examples

import (
	"fmt"
	"os"
	"strings"

	"github.com/rprtr258/cli"
	"github.com/rprtr258/fun"
)

type findPerms struct {
	read  bool
	write bool
	exec  bool
}

type findPermKind uint8

const (
	findPermAll findPermKind = iota
	findPermAny
	findPermExact
)

type findPerm struct {
	kind  findPermKind
	perms findPerms
}

type findOptions struct {
	paths []string
	exec  cli.Opt[[]string]
	user  cli.Opt[string]
	perm  cli.Opt[findPerm]
}

// Matches only the literal value, but unlike [cli.Literal] can appear anywhere
// on the command line - the counterpart of bpaf's `literal(...).anywhere()`.
func findLiteralAnywhere(val, help string) *cli.ParseAny[struct{}] {
	res := cli.Any("", func(s string) (struct{}, bool) {
		return struct{}{}, s == val
	}, true).
		Metavar(cli.NewDocFromPairs([]fun.Pair[string, cli.Style]{
			{K: val, V: cli.StyleLiteral},
		}))
	if help != "" {
		res = res.Help(help)
	}
	return res
}

// Parses -user xxx
var findUser = func() cli.Parser[cli.Opt[string]] {
	tag := findLiteralAnywhere("-user", "")
	value := cli.Positional("USER", cli.ParseString).Help("User name")
	return cli.Pair(tag, value).
		Adjacent().
		Map(func(p fun.Pair[struct{}, string]) string { return p.V }).
		Optional()
}()

// Parses -exec xxx yyy zzz ;
var findExec = func() cli.Parser[cli.Opt[[]string]] {
	tag := findLiteralAnywhere("-exec", "for every file find finds execute a separate shell command")

	item := cli.Any("ITEM", func(s string) (string, bool) {
		return s, s != ";"
	}, false).
		Help("command with its arguments, find will replace {} with a file name").
		Many()

	endtag := cli.Any(";", func(s string) (struct{}, bool) {
		return struct{}{}, s == ";"
	}, false).
		Help("anything after literal \";\" will be considered a regular option again")

	return cli.Triple(tag, item, endtag).
		Adjacent().
		Map(func(t cli.Tuple[struct{}, []string, struct{}]) []string { return t.B }).
		Optional()
}()

// Parses symbolic permissions `-perm -mode`, `-perm /mode` and `-perm mode`
var findPermParser = func() cli.Parser[cli.Opt[findPerm]] {
	parseMode := func(input string) (findPerms, error) {
		var perms findPerms
		for _, c := range input {
			switch c {
			case 'r':
				perms.read = true
			case 'w':
				perms.write = true
			case 'x':
				perms.exec = true
			default:
				return findPerms{}, fmt.Errorf("%s is not a valid permission string", input)
			}
		}
		return perms, nil
	}

	tag := findLiteralAnywhere("-mode", "")

	// `any` here is used to parse an arbitrary string that can also start with dash (-)
	// regular positional parser won't work here
	mode := cli.Parse(
		cli.Any("MODE", func(s string) (string, bool) { return s, true }, false).
			Help("(perm | -perm | /perm), where perm is any subset of rwx characters, ex +rw"),
		func(s string) (findPerm, error) {
			perms, err := parseMode(strings.TrimLeft(s, "-/"))
			if err != nil {
				return findPerm{}, err
			}
			switch {
			case strings.HasPrefix(s, "-"):
				return findPerm{kind: findPermAll, perms: perms}, nil
			case strings.HasPrefix(s, "/"):
				return findPerm{kind: findPermAny, perms: perms}, nil
			default:
				return findPerm{kind: findPermExact, perms: perms}, nil
			}
		},
	)

	return cli.Pair(tag, mode).
		Adjacent().
		Map(func(p fun.Pair[struct{}, findPerm]) findPerm { return p.V }).
		Optional()
}()

func Find() {
	paths := cli.Positional("PATH", cli.ParseString).Many()

	opts := cli.Map4(
		findExec, findUser, findPermParser, paths,
		func(exec cli.Opt[[]string], user cli.Opt[string], perm cli.Opt[findPerm], paths []string) findOptions {
			return findOptions{paths: paths, exec: exec, user: user, perm: perm}
		},
	).ToOptions()

	fmt.Println(opts.Run(os.Args))
}
