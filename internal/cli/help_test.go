package cli_test

// import (
// 	"context"
// 	"strings"
// 	"testing"

// 	"github.com/stretchr/testify/require"
// )

// type cmdBommand struct {
// 	Hidden bool `long:"hidden" description:"A hidden fun.Option"`
// }

// type cmdCommand struct {
// 	ExtraVerbose []bool `long:"extra-verbose" description:"Use for extra verbosity"`
// }

// type cmdHidden struct {
// 	ExtraVerbose []bool `long:"extra-verbose" description:"Use for extra verbosity"`
// }

// type cmdParentSub struct {
// 	Opt string `long:"opt" description:"This is a sub command fun.Option"`
// }

// type cmdParent struct {
// 	Opt        string       `long:"opt" description:"This is a parent command fun.Option"`
// 	SubCommand cmdParentSub `command:"sub" description:"A sub command"`
// }

// type cmdHelp struct {
// 	Verbose          []bool    `short:"v" long:"verbose" description:"Show verbose debug information"`
// 	PtrSlice         []*string `long:"ptrslice" description:"A slice of pointers to string"`
// 	EmptyDescription bool      `long:"empty-description"`

// 	Default           string            `long:"default" default:"Some\nvalue" description:"Test default value"`
// 	DefaultArray      []string          `long:"default-array" default:"Some value" default:"Other\tvalue" description:"Test default array value"`
// 	DefaultMap        map[string]string `long:"default-map" default:"some:value" default:"another:value" description:"Testdefault map value"`
// 	EnvDefault1       string            `long:"env-default1" default:"Some value" description:"Test env-default1 value"`
// 	EnvDefault2       string            `long:"env-default2" description:"Test env-default2 value"`
// 	OptionWithArgName string            `long:"opt-with-arg-name" value-name:"something" description:"Option with named argument"`
// 	OptionWithChoices string            `long:"opt-with-choices" value-name:"choice" choice:"dog" choice:"cat" description:"Option with choices"`
// 	Hidden            string            `long:"hidden" description:"Hidden fun.Option"`

// 	HiddenOptionWithVeryLongName bool `long:"this-hidden-option-has-a-ridiculously-long-name"`

// 	StringSlice []string       `short:"s" default:"some" default:"value" description:"A slice of strings"`
// 	IntMap      map[string]int `long:"intmap" default:"a:1" description:"A map from string to int"`

// 	// TODO: get back when groups are implemented
// 	// HiddenGroup struct {
// 	// 	InsideHiddenGroup string `long:"inside-hidden-group" description:"Inside hidden group"`
// 	// 	Padder            bool   `long:"this-option-in-a-hidden-group-has-a-ridiculously-long-name"`
// 	// } `group:"Hidden group"`

// 	// GroupWithOnlyHiddenOptions struct {
// 	// 	SecretFlag bool `long:"secret" description:"Hidden flag in a non-hidden group"`
// 	// }

// 	// Sip struct {
// 	// 	Opt                  string `long:"opt" description:"This is a subgroup fun.Option"`
// 	// 	HiddenInsideGroup    string `long:"hidden-inside-group" description:"Hidden inside group"`
// 	// 	NotHiddenInsideGroup string `long:"not-hidden-inside-group" description:"Not hidden inside group"`

// 	// 	Sap struct {
// 	// 		Opt string `long:"opt" description:"This is a subsubgroup fun.Option"`
// 	// 	}
// 	// }

// 	Bommand       cmdBommand `command:"bommand" description:"A command with only hidden fun.Options"`
// 	Command       cmdCommand `command:"command" alias:"cm" alias:"cmd" description:"A command"`
// 	HiddenCommand cmdHidden  `command:"hidden-command" description:"A hidden command"`
// 	ParentCommand cmdParent  `command:"parent" description:"A parent command"`

// 	Args struct {
// 		Filename     string  `positional-arg-name:"filename" description:"A filename with a long description to trigger line wrapping"`
// 		Number       int     `positional-arg-name:"num" description:"A number"`
// 		HiddenInHelp float32 `positional-arg-name:"hidden-in-help"`
// 	} `positional-args:"yes"`
// }

// func TestHelp(t *testing.T) {
// 	err := Run[cmdHelp](context.Background(), "x", "--help")
// 	require.Equal(t, Error{
// 		Kind: ErrHelp,
// 		Message: strings.Join([]string{
// 			"Usage: \x1b[94mx\x1b[0m \x1b[91m[COMMAND]\x1b[0m \x1b[93m[FLAGS]\x1b[0m \x1b[95m[ARGUMENTS]\x1b[0m",
// 			"",
// 			"",
// 			"",
// 			"Commands:",
// 			"\t\x1b[91mbommand\x1b[0m          A command with only hidden fun.Options",
// 			"\t\x1b[91mcommand\x1b[0m,\x1b[31mcm\x1b[0m,\x1b[31mcmd\x1b[0m   A command",
// 			"\t\x1b[91mhidden-command\x1b[0m   A hidden command",
// 			"\t\x1b[91mparent\x1b[0m           A parent command",
// 			"",
// 			"Arguments:",
// 			"\t\x1b[95m[Filename]\x1b[0m \x1b[92mstring\x1b[0m        A filename with a long description to trigger line wrapping",
// 			"\t\x1b[95m[Number]\x1b[0m \x1b[92mstring\x1b[0m          A number",
// 			"\t\x1b[95m[HiddenInHelp]\x1b[0m \x1b[92mstring\x1b[0m    ",
// 			"",
// 			"Flags:",
// 			"\t\x1b[93m-v\x1b[0m, \x1b[93m--verbose\x1b[0m                                        Show verbose debug information",
// 			"\t\x1b[93m--ptrslice\x1b[0m                                           A slice of pointers to string",
// 			"\t\x1b[93m--empty-description\x1b[0m                                  ",
// 			"\t\x1b[93m--default\x1b[0m                                            Test default value (default: Some\nvalue)",
// 			"\t\x1b[93m--default-array\x1b[0m                                      Test default array value (default: Some value, Other\tvalue)",
// 			"\t\x1b[93m--default-map\x1b[0m                                        Testdefault map value (default: some:value, another:value)",
// 			"\t\x1b[93m--env-default1\x1b[0m                                       Test env-default1 value (default: Some value)",
// 			"\t\x1b[93m--env-default2\x1b[0m                                       Test env-default2 value",
// 			"\t\x1b[93m--opt-with-arg-name\x1b[0m                                  fun.Option with named argument",
// 			"\t\x1b[93m--opt-with-choices\x1b[0m                                   fun.Option with choices",
// 			"\t\x1b[93m--hidden\x1b[0m                                             Hidden fun.Option",
// 			"\t\x1b[93m--this-hidden-option-has-a-ridiculously-long-name\x1b[0m    ",
// 			"\t\x1b[93m-s\x1b[0m                                                   A slice of strings (default: some, value)",
// 			"\t\x1b[93m--intmap\x1b[0m                                             A map from string to int (default: a:1)",
// 			"\t\x1b[93m-h\x1b[0m, \x1b[93m--help\x1b[0m                                           Print this help message",
// 			"",
// 		}, "\n"),
// 	}, err)
// }

// type helpCommandNoOptions struct {
// 	Command cmdParentSub `command:"command" description:"A command"`
// }

// func TestHelpCommand(t *testing.T) {
// 	err := Run[helpCommandNoOptions](context.Background(), "x", "command", "--help")
// 	require.Equal(t, &Error{
// 		Kind: ErrHelp,
// 		Message: strings.Join([]string{
// 			"Usage: \x1b[94mx command\x1b[0m \x1b[93m[FLAGS]\x1b[0m",
// 			"",
// 			"A command",
// 			"",
// 			"Flags:",
// 			"\t\x1b[93m--opt\x1b[0m         This is a sub command fun.Option",
// 			"\t\x1b[93m-h\x1b[0m, \x1b[93m--help\x1b[0m    Print this help message",
// 			"",
// 		}, "\n"),
// 	}, err)
// }

// type cmdHelpDefaults struct {
// 	WithDefault             string `long:"with-default" default:"default-value" description:"With default"`
// 	WithoutDefault          string `long:"without-default" description:"Without default"`
// 	WithProgrammaticDefault string `long:"with-programmatic-default" description:"With programmatic default" default:"default-value"`
// }

// func TestHelpDefaults(t *testing.T) {
// 	for name, args := range map[string][]string{
// 		"just help":  {"x", "-h"},
// 		"some flags": {"x", "--with-default", "other-value", "--with-programmatic-default", "other-value", "-h"},
// 	} {
// 		t.Run(name, func(t *testing.T) {
// 			err := Run[cmdHelpDefaults](context.Background(), args...)
// 			require.Equal(t, &Error{
// 				Kind: ErrHelp,
// 				Message: strings.Join([]string{
// 					"Usage: \x1b[94mx\x1b[0m \x1b[93m[FLAGS]\x1b[0m",
// 					"",
// 					"",
// 					"",
// 					"Flags:",
// 					"\t\x1b[93m--with-default\x1b[0m                 With default (default: default-value)",
// 					"\t\x1b[93m--without-default\x1b[0m              Without default",
// 					"\t\x1b[93m--with-programmatic-default\x1b[0m    With programmatic default (default: default-value)",
// 					"\t\x1b[93m-h\x1b[0m, \x1b[93m--help\x1b[0m                     Print this help message",
// 					"",
// 				}, "\n"),
// 			}, err)
// 		})
// 	}
// }
