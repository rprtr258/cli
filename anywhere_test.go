package cli

import (
	"testing"

	"github.com/rprtr258/fun"
	"github.com/stretchr/testify/require"
)

func TestParse_anywhere_positional(t *testing.T) {
	a := Any("X", func(h string) (string, bool) { return h, h != "--help" }).
		Help("all the things").
		Anywhere()

	b := Short('b').Help("batch mode").Switch()
	var parser *OptionParser[fun.Pair[string, bool]] = Pair(a, b).Parser().ToOptions()

	_, r := parser.run_inner(newArgs("--help"))

	const expected = `Usage: X [-b]

Available options:
    X           all the things
    -b          batch mode
    -h, --help  Prints help information
`

	require.Equal(t, unwrap_stdout(r), expected)
	// this should be allowed because "anywhere" prevents anything inside from being positional
	parser.check_invariants(true)
}

func TestParse_anywhere_no_catch(t *testing.T) {
	a := ReqFlag(Short('a'), struct{}{})
	b := Positional("X", parseUint)
	ab := Pair(a, b).Adjacent()
	c := Short('c').Switch()
	parser := Pair(ab, c).Parser().ToOptions()

	// Usage: -a <x> [-c],

	{
		_, r := parser.run_inner(newArgs("3", "-a"))
		require.Equal(t, unwrap_stderr(r), "expected `X`, pass `--help` for usage information")
	}
	{
		_, r := parser.run_inner(newArgs("-a"))
		require.Equal(t, unwrap_stderr(r), "expected `X`, pass `--help` for usage information")
	}
	{
		_, r := parser.run_inner(newArgs("-a", "221b"))
		require.Equal(t, unwrap_stderr(r), "couldn't parse `221b`: invalid digit found in string")
	}
	{
		_, r := parser.run_inner(newArgs("-c", "-a"))
		require.Equal(t, unwrap_stderr(r), "expected `X`, pass `--help` for usage information")
	}
	{
		_, r := parser.run_inner(newArgs("-c", "-a", "221b"))
		require.Equal(t, unwrap_stderr(r), "couldn't parse `221b`: invalid digit found in string")
	}
	{
		_, r := parser.run_inner(newArgs("-a", "-c"))
		require.Equal(t, unwrap_stderr(r), "expected `X`, pass `--help` for usage information")
	}
	{
		_, r := parser.run_inner(newArgs("-a", "221b", "-c"))
		require.Equal(t, unwrap_stderr(r), "couldn't parse `221b`: invalid digit found in string")
	}
}

// #[test]
// fn anywhere_catch_optional() {
//     let a = short('a').req_flag(());
//     let b = positional::<usize>("x");
//     let ab = construct(a, b).adjacent().optional().catch();
//     let bc = short('a').Switch();
//     let parser = construct(ab, bc).to_options();

//     let r = parser.run_inner(newArgs("-a", "10")).unwrap();
//     require.Equal(t, r, (Some(((), 10)), false));

//     let r = parser.run_inner(newArgs("-a")).unwrap();
//     require.Equal(t, r, (None, true));

//     let r = parser.run_inner(newArgs()).unwrap();
//     require.Equal(t, r, (None, false));
// }

// #[test]
// fn anywhere_catch_many() {
//     let a = short('a').req_flag(());
//     let b = positional::<usize>("x");
//     let ab = construct(a, b).adjacent().many().catch();
//     let bc = short('a').Switch();
//     let parser = construct(ab, bc).to_options();

//     let r = parser.run_inner(newArgs("-a")).unwrap();

//     require.Equal(t, r, (vec![], true));

//     let r = parser.run_inner(newArgs("-a", "10")).unwrap();
//     require.Equal(t, r, (vec![((), 10)], false));

//     let r = parser.run_inner(newArgs()).unwrap();
//     require.Equal(t, r, (Vec::new(), false));
// }

// #[test]
// fn anywhere_catch_fallback() {
//     let a = short('a').req_flag(());
//     let b = positional::<usize>("x");
//     let ab = construct(a, b).adjacent().fallback(((), 10));
//     let bc = short('a').Switch();
//     let parser = construct(ab, bc).to_options();

//     let r = parser.run_inner(newArgs("-a", "12")).unwrap();
//     require.Equal(t, r, (((), 12), false));

//     let r = parser.run_inner(newArgs("-a")).unwrap();
//     require.Equal(t, r, (((), 10), true));

//     let r = parser.run_inner(newArgs()).unwrap();
//     require.Equal(t, r, (((), 10), false));
// }

// #[test]
// fn parse_anywhere_catch_optional() {
//     let a = short('a').req_flag(());
//     let b = positional::<usize>("x");

//     // optional + catch makes it so parser succeeds without consuming anything
//     // usually leaving `-a` untouched to be consumed by something else
//     let ab = construct(a, b).adjacent().optional().catch();
//     let c = short('c').Switch();
//     let parser = construct(ab, c).to_options();

//     let r = parser.
//         run_inner(newArgs("-a", "221b"]).
//         unwrap_err().
//         unwrap_stderr();
//     require.Equal(t, r, "`-a` is not expected in this context");

//     let r = parser.run_inner(newArgs("3", "-a")).unwrap_err().unwrap_stderr();
//     require.Equal(t, r, "`3` is not expected in this context");

//     let r = parser.run_inner(newArgs("-a")).unwrap_err().unwrap_stderr();
//     require.Equal(t, r, "`-a` is not expected in this context");

//     let r = parser.run_inner(newArgs("-c", "-a")).unwrap_err().unwrap_stderr();
//     require.Equal(t, r, "`-a` is not expected in this context");

//     let r = parser.
//         run_inner(newArgs("-c", "-a", "221b"]).
//         unwrap_err().
//         unwrap_stderr();
//     require.Equal(t, r, "`-a` is not expected in this context");

//     let r = parser.run_inner(newArgs("-a", "-c")).unwrap_err().unwrap_stderr();
//     require.Equal(t, r, "`-a` is not expected in this context");

//     let r = parser.
//         run_inner(newArgs("-a", "221b", "-c"]).
//         unwrap_err().
//         unwrap_stderr();
//     require.Equal(t, r, "`-a` is not expected in this context");
// }

// #[test]
// fn anywhere_literal() {
//     let tag = any(
//         "-mode",
//         |x: String| if x == "-mode" { Some(()) } else { None },
//     );
//     let mode = positional::<usize>("value");
//     let a = construct(tag, mode).adjacent().many().catch();
//     let b = short('b').Switch();
//     let parser: OptionParser<(Vec<((), usize)>, bool)> = construct(a, b).to_options();

//     let r = parser.run_inner(newArgs("-b", "-mode", "12")).unwrap();
//     require.Equal(t, r, (vec![((), 12)], true));

//     let r = parser.run_inner(newArgs("-mode", "12", "-b")).unwrap();
//     require.Equal(t, r, (vec![((), 12)], true));

//     let r = parser.run_inner(newArgs("-mode", "12")).unwrap();
//     require.Equal(t, r, (vec![((), 12)], false));
// }
