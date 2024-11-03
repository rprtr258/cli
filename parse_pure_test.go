package cli

import (
	"strconv"
	"testing"

	"github.com/stretchr/testify/require"
)

// TODO: move similar funcs to batteries and use
func parseUint32(s string) (uint32, error) {
	u, err := strconv.ParseUint(s, 10, 32)
	return uint32(u), err
}

func TestDefaultValueUsingPureWithOk(t *testing.T) {
	// ~ this is a bit artifactial (it's better to use fallback_with instead actually),
	// but this example demonstrates the concept
	a := Argument(Short('t'), "N", parseUint32)
	b := PureWith(func() (uint32, Error) { return 42, nil })

	opts := Choice(a.Parser(), b.Parser()).ToOptions()

	for name, tc := range map[string]struct {
		args     []string
		expected uint32
	}{
		"parse": {
			args:     []string{"-t", "87"},
			expected: 87,
		},
		"default": {
			args:     nil,
			expected: 42,
		},
	} {
		t.Run(name, func(t *testing.T) {
			r, err := opts.run_inner(newArgs(tc.args...))
			requireNoParseFailure(t, err)
			require.Equal(t, tc.expected, r)
		})
	}
}

func TestDefaultValueUsingPureWithErr(t *testing.T) {
	a := Argument(Short('t'), "N", parseUint32)
	b := PureWith(func() (uint32, Error) { return 0, Error(MessageParseFail("some-err")) })

	opts := Choice(a.Parser(), b.Parser()).ToOptions()
	_, err := opts.run_inner(newArgs())
	require.Equal(t, "some-err", unwrap_stderr(err))
}

// func default_value_using_pure_with_ok_for_some() {
//     let user_seeds = positional[uint32]("SEED").some("at least one required");
//     let last_seeds = pure_with::<_, _, string>(|| {
//         // ~ trying to lookup the last used seeds
//         Ok(vec[3, 5, 7, 11])
//     });
//     let seeds = construct([user_seeds, last_seeds]).to_options();

//     let r = seeds.run_inner(&["23", "59"]);
//     require.Equal(t, vec[23, 59], r.unwrap());

//     let r = seeds.run_inner(&[]);
//     require.Equal(t, vec[3, 5, 7, 11], r.unwrap());
// }

// func default_value_using_pure_with_err_for_some() {
//     let user_seeds = positional::<u32>("SEED").some("at least one required");
//     let last_seeds = pure_with(|| {
//         // ~ trying to lookup the last used seeds but failing - parser fails, fallback works
//         Err("oh, no")
//     });
//     let default_seeds = pure(vec[1, 2]);
//     let seeds = construct([user_seeds, last_seeds, default_seeds]).to_options();

//     let r = seeds.run_inner(&["23", "59"]);
//     require.Equal(t, vec[23, 59], r.unwrap());

//     let r = seeds.run_inner(&[]);
//     require.Equal(t, r.unwrap(), vec[1, 2]);
// }
