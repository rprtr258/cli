// `confusing.go` - you can do potentially confusing things if you really-really have to :)
//
// ./binary --token "token" complex1 42 - ./binary complex1 42
// The first form requires the token because complex commands make api calls, the
// second one does not.
package examples

import (
	"errors"
	"fmt"
	"os"
	"strconv"

	"github.com/rprtr258/cli"
	"github.com/rprtr258/fun"
)

func Confusing() {
	// this datatype is intended for program consumption, int field in complex commands
	// is a shared top level argument
	type Command struct {
		kind  string // "simple", "complex1", "complex2"
		token string
		arg   int
	}

	// this datatype is just an intermediate representation,
	// it exists only to ensure that the final type (Command) can be used without unwraps
	type PreCommand struct {
		kind string
		arg  int
	}

	token := cli.Long("token").
		Help("Token used for complex commands").
		Argument("TOKEN", cli.ParseString).
		Optional()

	// start with defining 3 commands: simple, complex1 and complex2
	simple := cli.Pure(PreCommand{kind: "simple"}).
		ToOptions().
		Command("simple")

	complex1 := cli.Positional("ARG", cli.ParseInt).
		Map(func(arg int) PreCommand { return PreCommand{kind: "complex1", arg: arg} }).
		ToOptions().
		Description("This is complex command 1").
		Command("complex1")

	parseI16 := func(s string) (int, error) {
		v, err := strconv.ParseInt(s, 10, 16)
		return int(v), err
	}
	complex2 := cli.Positional("ARG", parseI16).
		Map(func(arg int) PreCommand { return PreCommand{kind: "complex2", arg: arg} }).
		ToOptions().
		Description("This is complex command 2").
		Command("complex2")

	// compose them to accept any of those
	preparser := cli.Choice(simple, complex1, complex2)

	// make a parser that accepts optional token and one of incomplete commands
	// then create complete command or fail
	parser := cli.Parse(
		cli.Pair(token, preparser),
		func(p fun.Pair[cli.Opt[string], PreCommand]) (Command, error) {
			switch p.V.kind {
			case "simple":
				return Command{kind: "simple"}, nil
			case "complex1", "complex2":
				tok, ok := p.K.Unpack()
				if !ok {
					return Command{}, errors.New("You must specify token to use with --token")
				}
				return Command{kind: p.V.kind, token: tok, arg: p.V.arg}, nil
			default:
				return Command{}, errors.New("unreachable")
			}
		},
	)

	cmd := parser.ToOptions().Run(os.Args)
	fmt.Println(cmd)
}
