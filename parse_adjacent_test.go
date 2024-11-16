package cli

import (
	"strconv"
	"testing"

	"github.com/kr/pretty"
	"github.com/rprtr258/fun"
	"github.com/stretchr/testify/require"
)

func requireNoParseFailure(t *testing.T, err ParseFailure) {
	if err != nil {
		switch self := err.(type) {
		case ParseFailureStdout:
			t.Log("Unexpected ParseFailureStdout:", pretty.Sprint(self.msg), self.full)
		case ParseFailureCompletion:
			t.Log("Unexpected ParseFailureCompletion:", string(self))
		case ParseFailureStderr:
			t.Log("Unexpected ParseFailureStderr:", pretty.Sprint(Doc(self)))
		}
		t.FailNow()
	}
}

func TestAdjacent(t *testing.T) {
	type cmd struct {
		a    struct{}
		b, c bool
	}

	a := ReqFlag(Short('a'), struct{}{})
	b := Short('b').Switch()
	c := Short('c').Switch()

	parser := Many(Map(Pair(a.Parser(), Pair(b, c).Parser()).Adjacent().Parser(), func(abc fun.Pair[struct{}, fun.Pair[bool, bool]]) cmd { return cmd{abc.K, abc.V.K, abc.V.V} })).ToOptions()

	for name, tc := range map[string]struct {
		args     []string
		expected []cmd
	}{
		"ac,ab,abc": {
			// adjacent groups here argument
			// -a [-b] -c  | -a -b [-c] | -a -b -c
			args:     []string{"-a", "-c", "-a", "-b", "-a", "-b", "-c"},
			expected: []cmd{{struct{}{}, false, true}, {struct{}{}, true, false}, {struct{}{}, true, true}},
		},
		"a": {
			args:     []string{"-a"},
			expected: []cmd{{struct{}{}, false, false}},
		},
		"ac": {
			args:     []string{"-a", "-c"},
			expected: []cmd{{struct{}{}, false, true}},
		},
		"empty": {
			args:     []string{},
			expected: []cmd{},
		},
	} {
		t.Run(name, func(t *testing.T) {
			r, err := parser.run_inner(newArgs(tc.args...))
			requireNoParseFailure(t, err)
			require.Equal(t, tc.expected, r)
		})
	}
}

func TestAdjacentPrefix(t *testing.T) {
	type cmd = fun.Pair[Option[fun.Pair[struct{}, uint]], bool]
	a := ReqFlag(Short('a'), struct{}{})
	b := Positional("X", func(s string) (uint, error) {
		u, err := strconv.ParseUint(s, 10, 32)
		return uint(u), err
	})
	ab := Optional(Pair(a.Parser(), b.Parser()).Adjacent().Parser())
	c := Short('c').Switch()
	parser := Pair(ab.Parser(), c).Parser().ToOptions()

	for name, tc := range map[string]struct {
		args     []string
		expected cmd
	}{
		"-c": {
			args: []string{"-c"},
			expected: cmd{
				Invalid[fun.Pair[struct{}, uint]](),
				true,
			},
		},
		"-a 1": {
			args: []string{"-a", "1"},
			expected: cmd{
				Valid(fun.Pair[struct{}, uint]{struct{}{}, 1}),
				false,
			},
		},
		"-c -a 1": {
			args: []string{"-c", "-a", "1"},
			expected: cmd{
				Valid(fun.Pair[struct{}, uint]{struct{}{}, 1}),
				true,
			},
		},
		"-a 1 -c": {
			args: []string{"-a", "1", "-c"},
			expected: cmd{
				Valid(fun.Pair[struct{}, uint]{struct{}{}, 1}),
				true,
			},
		},
	} {
		t.Run(name, func(t *testing.T) {
			r, err := parser.run_inner(newArgs(tc.args...))
			requireNoParseFailure(t, err)
			require.Equal(t, tc.expected, r)
		})
	}
}

// func adjacent_error_message_pos_single() {
//     let a = short('a').req_flag(());
//     let b = positional::<usize>("B");
//     let c = positional::<usize>("C");
//     let d = short('d').switch();
//     let adj = construct(a, b, c).adjacent();
//     let parser = construct(adj, d).to_options();

//     let r = parser.run_inner(newArgs("-a", "10")).unwrap_err().unwrap_stderr();
//     require.Equal(r, "expected `C`, pass `--help` for usage information");
// }

// func adjacent_error_message_arg_single() {
//     let a = short('a').req_flag(());
//     let b = short('b').argument::<usize>("B");
//     let c = short('c').argument::<usize>("C");
//     let d = short('d').switch();
//     let adj = construct(a, b, c).adjacent();
//     let parser = construct(adj, d).to_options();

//     let r = parser.run_inner(newArgs("-a", "10")).unwrap_err().unwrap_stderr();
//     require.Equal(
//         r,
//         "expected `-b=B`, got `10`. Pass `--help` for usage information"
//     );
// }

// func adjacent_error_message_pos_many() {
//     let a = short('a').req_flag(());
//     let b = positional::<usize>("B");
//     let c = positional::<usize>("C");
//     let d = short('d').switch();
//     let adj = construct(a, b, c).adjacent().many();
//     let parser = construct(adj, d).to_options();

//     let r = parser.run_inner(newArgs("-a", "10")).unwrap_err().unwrap_stderr();
//     require.Equal(r, "expected `C`, pass `--help` for usage information");
// }

// func adjacent_error_message_arg_many() {
//     let a = short('a').req_flag(());
//     let b = short('b').argument::<usize>("B");
//     let c = short('c').argument::<usize>("C");
//     let d = short('d').switch();
//     let adj = construct(a, b, c).adjacent().many();
//     let parser = construct(adj, d).to_options();

//     let r = parser.run_inner(newArgs("-a", "10")).unwrap_err().unwrap_stderr();
//     // this should ask for -b or -c and complain on 10...
//     require.Equal(
//         r,
//         "expected `-b=B`, got `10`. Pass `--help` for usage information"
//     );
// }

// func adjacent_is_adjacent() {
//     let a = short('a').req_flag(());
//     let b = positional::<usize>("B");
//     let parser = construct(a, b).adjacent().map(|t| t.1).many().to_options();

//     let r = parser
//         .run_inner(newArgs("-a", "-a", "10", "20"])
//         .unwrap_err()
//         .unwrap_stderr();
//     require.Equal(r, "expected `B`, pass `--help` for usage information");

//     let r = parser.run_inner(newArgs("-a", "10", "-a", "20")).unwrap();
//     require.Equal(r, [10, 20]);
// }

// func adjacent_with_switch() {
//     let a = short('a').req_flag(());
//     let b = positional::<usize>("B");
//     let ab = construct(a, b).adjacent().map(|t| t.1).many();
//     let c = short('c').switch();
//     let parser = construct(ab, c).to_options();

//     let r = parser.run_inner(newArgs("-a", "10", "-c")).unwrap();
//     require.Equal(r, (vec[10], true));

//     let r = parser.run_inner(newArgs("-a", "10", "-c", "-a", "20")).unwrap();
//     require.Equal(r, (vec[10, 20], true));

//     let r = parser.run_inner(newArgs("-c", "-a", "10", "-a", "20")).unwrap();
//     require.Equal(r, (vec[10, 20], true));
// }

// func adjacent_limits_commands() {
//     let x = pure(()).to_options().command("a").adjacent();
//     let s = short('s').switch();
//     let parser = construct(s, x).to_options();

//     let r = parser.run_inner(newArgs("a", "-s")).unwrap();
//     require.Equal(r, (true, ()));
// }

// func commands_and_adjacent() {
//     let eat = positional::<String>("FOOD")
//         .to_options()
//         .command("eat")
//         .help("eat something")
//         .adjacent();

//     let sleep = long("time")
//         .argument::<String>("HOURS")
//         .to_options()
//         .command("sleep")
//         .help("sleep for a bit")
//         .adjacent();

//     let cmds = construct([eat, sleep]);
//     let switch = short('s').switch();

//     let parser = construct(switch, cmds).to_options();

//     let r = parser.run_inner(newArgs("sleep", "--time", "12", "-s")).unwrap();
//     require.Equal(r, (true, "12".to_owned()));

//     let r = parser.run_inner(newArgs("--help")).unwrap_err().unwrap_stdout();

//     // TODO - this is ugly
//     let expected = "\
// Usage: [-s] COMMAND ...

// Available options:
//     -s
//     -h, --help  Prints help information

// Available commands:
//     eat         eat something
//     sleep       sleep for a bit
// ";

//     require.Equal(r, expected);
// }

// func two_adjacent_args() {
//     let x = short('x').argument::<usize>("X");
//     let y = short('y').argument::<usize>("Y");
//     let c = short('c').switch();
//     let point = construct(x, y).adjacent();
//     let parser = construct(point, c).to_options();

//     let r = parser.run_inner(newArgs("-x", "3", "-y", "4", "-c")).unwrap();
//     require.Equal(r, ((3, 4), true));

//     // they are adjacent to each other, but the way it is coded currently - they must be adjacent
//     // to the first element.
//     // Proper fix is to split "adjacent" into "adjacent to" and "adjacent block"

//     // let r = parser.run_inner(newArgs("-y", "3", "-x", "4", "-c")).unwrap();
//     // require.Equal(r, ((4, 3), true));

//     let r = parser
//         .run_inner(newArgs("-y", "3", "-c", "-x", "4"])
//         .unwrap_err()
//         .unwrap_stderr();
//     require.Equal(r, "expected `-y=Y`, pass `--help` for usage information");
// }
