package examples_test

// Every example is executed as a subprocess of the test binary and its output is
// checked. The subprocess isolation keeps `os.Exit` from failed parses or `--help`
// handling from killing the test suite and captures the real stdout/stderr.
//
// The child process is dispatched from [TestMain] via the RUN_EXAMPLE environment
// variable; example command line arguments are passed through EXAMPLE_ARGS joined
// with 0x1f (arguments can contain spaces, and 0x1f cannot appear inside an
// environment variable).

import (
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"os/exec"
	"strings"
	"testing"
	"unicode"

	"github.com/stretchr/testify/require"

	"github.com/rprtr258/cli/examples"
)

func TestMain(m *testing.M) {
	if name := os.Getenv("RUN_EXAMPLE"); name != "" {
		fn := lookup(name)
		os.Args = append([]string{name}, parseExampleArgs()...)
		fn()
		os.Exit(0)
	}
	os.Exit(m.Run())
}

// parseExampleArgs decodes the EXAMPLE_ARGS environment variable.
func parseExampleArgs() []string {
	raw := os.Getenv("EXAMPLE_ARGS")
	if raw == "" {
		return nil
	}
	return strings.Split(raw, "\x1f")
}

// examplesByName maps an example name to its function. It must cover every
// exported no-argument example function in the package - see
// [TestAllExamplesCovered].
var examplesByName = map[string]func(){
	"app_name":                         examples.AppName,
	"at_least_two":                     examples.AtLeastTwo,
	"basic":                            examples.Basic,
	"cargo_cmd":                        examples.CargoCmd,
	"cat":                              examples.Cat,
	"compression":                      examples.Compression,
	"confusing":                        examples.Confusing,
	"construct_struct":                 examples.ConstructStruct,
	"coreutils":                        examples.Coreutils,
	"csample":                          examples.Csample,
	"customize_help":                   examples.CustomizeHelp,
	"dd":                               examples.Dd,
	"derive":                           examples.Main,
	"derive_commands":                  examples.DeriveCommands,
	"derive_rudo":                      examples.DeriveRudo,
	"derive_show_asm":                  examples.DeriveShowAsm,
	"derive_smart_pointer":             examples.DeriveSmartPointer,
	"derive_this_or_that":              examples.DeriveThisOrThat,
	"dynamic":                          examples.Dynamic,
	"dynamic_tree":                     examples.DynamicTree,
	"env_logger":                       examples.EnvLogger,
	"fallback_command":                 examples.FallbackCommand,
	"env_variable":                     examples.EnvVariable,
	"enum_in_args":                     examples.EnumInArgs,
	"enum_tuple":                       examples.EnumTuple,
	"ex_positional":                    examples.ExPositional,
	"filenames":                        examples.Filenames,
	"find":                             examples.Find,
	"flatten":                          examples.Flatten,
	"git":                              examples.Git,
	"many_comma_separated_args":        examples.ManyCommaSeparatedArgs,
	"many_comma_separated_args_derive": examples.ManyCommaSeparatedArgsDerive,
	"multiple_fallback":                examples.MultipleFallback,
	"negative":                         examples.Negative,
	"no_import":                        examples.NoImport,
	"numeric_prefix":                   examples.NumericPrefix,
	"positional_derive":                examples.PositionalDerive,
	"rectangle":                        examples.Rectangle,
	"sensors":                          examples.Sensors,
	"shared_args":                      examples.SharedArgs,
	"simple_dynamic":                   examples.SimpleDynamic,
	"top_to_bottom":                    examples.TopToBottom,
	"travel":                           examples.Travel,
	"verbose":                          examples.Verbose,
	"very_custom_usage":                examples.VeryCustomUsage,
	"with_examples":                    examples.WithExamples,
	"wtf":                              examples.Wtf,
	"xorg":                             examples.Xorg,
}

// lookup returns the example function by name, panicking on unknown names.
func lookup(name string) func() {
	fn, ok := examplesByName[name]
	if !ok {
		panic("unknown example: " + name)
	}
	return fn
}

// runExample runs one example in a subprocess of the test binary, requiring a
// zero exit code and the expected substrings in the combined output.
func runExample(t *testing.T, name string, wantSubstrings []string, args ...string) {
	t.Helper()

	cmd := exec.Command(os.Args[0])
	cmd.Env = append(os.Environ(),
		"RUN_EXAMPLE="+name,
		"EXAMPLE_ARGS="+strings.Join(args, "\x1f"),
	)
	out, err := cmd.CombinedOutput()
	require.NoError(t, err, "%s: args %q, output:\n%s", name, args, out)
	for _, want := range wantSubstrings {
		require.Contains(t, string(out), want, "%s: args %q", name, args)
	}
}

func TestExamples(t *testing.T) {
	for _, test := range exampleTests {
		t.Run(test.name, func(t *testing.T) {
			runExample(t, test.name, test.want, test.args...)
		})
	}
}

// exampleTests maps every example to the command line it is run with and the
// substrings expected in its output. It must cover every entry of
// [examplesByName] - see [TestAllExamplesCovered].
var exampleTests = []struct {
	name string
	args []string
	want []string
}{
	{"app_name", []string{"-u", "alice"}, []string{"alice"}},
	{"at_least_two", []string{"-f", "-f"}, []string{"{}"}},
	{"basic", []string{"-d", "-vv", "-s", "100.5", "-o", "out.txt", "-n", "7", "-f", "a.rs", "-f", "b.rs"}, []string{"100.5", "out.txt", "7"}},
	{"cargo_cmd", []string{"cmd", "-w", "3", "-h", "4"}, []string{"3 4"}},
	{"compression", []string{"-5"}, []string{"5"}},
	{"confusing", []string{"--token", "secret", "complex1", "42"}, []string{"secret"}},
	{"construct_struct", []string{"-d", "-v", "--speed", "12.5", "-o", "out", "-n", "7", "file1"}, []string{"12.5"}},
	{"coreutils", []string{"arch"}, []string{"arch"}},
	{"csample", []string{"-a", "-b", "--bananananana", "--calculator", "1+1"}, []string{"1+1"}},
	{"customize_help", []string{"-d"}, []string{"true"}},
	{"dd", []string{"if=/tmp/a", "of=/tmp/b", "bs=1024", "-m"}, []string{"/tmp/a", "/tmp/b"}},
	{"derive", []string{"--debug", "-vv", "--speed", "55", "out"}, []string{"55"}},
	{"derive_commands", []string{"explain", "serde", "derive"}, []string{"serde"}},
	{"derive_rudo", []string{"add", "buy milk"}, []string{"buy milk"}},
	{"derive_show_asm", []string{"--dry", "--offline", "func1"}, []string{"func1"}},
	{"derive_smart_pointer", []string{"--name", "Bilbo", "1", "2", "3"}, nil},
	{"derive_this_or_that", []string{"--on"}, nil},
	{"dynamic", []string{"--banana", "--width", "42", "--name", "hello"}, []string{"42", "hello"}},
	{"dynamic_tree", []string{"echoCommand"}, []string{"echo 'some text'"}},
	{"env_logger", []string{"-vvv", "--potato", "7"}, []string{"7"}},
	{"fallback_command", []string{"build", "2"}, []string{"2"}},
	{"env_variable", []string{"--key", "k1"}, []string{"k1"}},
	{"enum_in_args", []string{"--baz", "foobar"}, []string{"2"}},
	{"enum_tuple", []string{"foo", "-b", "hi"}, nil},
	{"ex_positional", []string{"--value", "9", "f1", "f2"}, []string{"9"}},
	{"filenames", []string{"a.md", "b.toml"}, []string{"a.md"}},
	{"find", []string{"-user", "bob", "/etc", "-exec", "ls", "-la", ";", "-mode", "rwx"}, []string{"bob", "/etc"}},
	{"flatten", []string{"-v", "-u", "bob", "-g", "dev"}, []string{"bob"}},
	{"git", []string{"fetch", "--dry_run"}, nil},
	{"many_comma_separated_args", []string{"--ports", "1,2,3", "--ports", "4,5"}, []string{"1 2 3 4 5"}},
	{"many_comma_separated_args_derive", []string{"--ports", "1,2,3", "--ports", "4,5"}, []string{"1 2 3 4 5"}},
	{"multiple_fallback", []string{"--field1", "1", "--field2", "2"}, []string{"1 2"}},
	{"negative", []string{"--age=-1"}, []string{"age: -1"}},
	{"no_import", []string{"-d", "-s", "11"}, []string{"11"}},
	{"numeric_prefix", []string{"10", "list"}, []string{"10"}},
	{"positional_derive", []string{"--value", "7", "f1", "f2"}, []string{"7"}},
	{"rectangle", []string{"-w", "3", "-h", "4", "-v"}, []string{"3 4"}},
	{"sensors", []string{"--sensor", "--sensor-device", "temp100", "--sensor-name", "n1", "--sensor-i2c-bus", "1", "--sensor-i2c-address", "2"}, []string{"temp100"}},
	{"shared_args", []string{"action", "--number", "3", "x", "y"}, []string{"x y"}},
	{"simple_dynamic", []string{"--crate", "cargo-hackerman", "--json", "dep1"}, []string{"cargo-hackerman"}},
	{"top_to_bottom", []string{"-d", "-o", "out", "-n", "3", "-f", "f1"}, []string{"out"}},
	{"travel", []string{"--time", "2", "--mph", "60", "--km", "100"}, nil},
	{"verbose", []string{"-vvv", "--trim"}, []string{"3"}},
	{"very_custom_usage", []string{"cmd", "--opt"}, []string{"true"}},
	{"with_examples", []string{"file.txt", "backup/"}, []string{"file.txt -> backup/"}},
	{"with_examples", []string{"--recursive", "src/", "dst/"}, []string{"src/ -> dst/ recursively"}},
	{"wtf", []string{"add", "-a", "f1", "f2"}, []string{"Adding (all=true)", "\"f1\", \"f2\""}},
	{"wtf", []string{"rm", "-f", "f1"}, []string{"Removing (force=true)", "\"f1\""}},
	{"wtf", []string{"example", "-n", "foo", "--offset", "7", "--point", "3,4", "-s", "a", "-s", "b", "f.txt"}, []string{"Example(verbose=false, offset=7, name=\"foo\", point=3,4, file=\"-\", num=None, strs=[]string{\"a\", \"b\"}, files=[]string{\"f.txt\"})"}},
	{"wtf", []string{"example", "-n", "foo", "-p", "5", "-f", "cfg.yml"}, []string{"Example(verbose=false, offset=0, name=\"foo\", point=1,2, file=\"cfg.yml\", num=Some(5), strs=[]string{}, files=[]string{})"}},
	{"wtf", []string{"docker", "image", "ls", "-o", "yaml", "img1"}, []string{"Listing images (verbose=false, output=yaml)"}},
	{"xorg", []string{"-t", "+ext", "GLX"}, []string{"GLX"}},
}

// TestAllExamplesCovered checks that every exported no-argument function in the
// examples package (i.e. every example) is present in [examplesByName] and in
// the [TestExamples] table, so new examples cannot silently go unchecked.
func TestAllExamplesCovered(t *testing.T) {
	fset := token.NewFileSet()
	pkgs, err := parser.ParseDir(fset, ".", func(fi fs.FileInfo) bool {
		return !strings.HasSuffix(fi.Name(), "_test.go")
	}, 0)
	require.NoError(t, err)

	funcs := make(map[string]struct{})
	for _, file := range pkgs["examples"].Files {
		for _, decl := range file.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if !ok || fn.Recv != nil || !ast.IsExported(fn.Name.Name) ||
				fn.Type.Results != nil || len(fn.Type.Params.List) != 0 {
				continue
			}
			funcs[fn.Name.Name] = struct{}{}
		}
	}
	require.NotEmpty(t, funcs)

	// every example function must map to exactly one example key, and the
	// dispatch map must cover exactly those keys
	wantKeys := make(map[string]struct{}, len(funcs))
	for funcName := range funcs {
		if mapped, ok := funcNameToExample[funcName]; ok {
			wantKeys[mapped] = struct{}{}
			continue
		}
		wantKeys[camelToSnake(funcName)] = struct{}{}
	}

	gotKeys := make(map[string]struct{}, len(examplesByName))
	for name := range examplesByName {
		gotKeys[name] = struct{}{}
	}
	require.Equal(t, wantKeys, gotKeys, "examplesByName must cover every example function exactly")

	// and every dispatched example must be present in the run table; cat is
	// the exception, it needs a real temp file and runs in [TestExamples_cat]
	for name := range gotKeys {
		if name == "cat" {
			continue
		}
		require.True(t, containsName(exampleTests, name), "example %s is not in the TestExamples table", name)
	}
}

// funcNameToExample maps example functions whose example key is not derived
// from the function name mechanically.
var funcNameToExample = map[string]string{
	"Main": "derive", // examples.Main is the derive.rs example
}

// camelToSnake converts an exported function name like ManyCommaSeparatedArgs
// to an example key like many_comma_separated_args.
func camelToSnake(s string) string {
	var b strings.Builder
	for i, c := range s {
		if unicode.IsUpper(c) {
			if i > 0 {
				b.WriteByte('_')
			}
			b.WriteRune(unicode.ToLower(c))
		} else {
			b.WriteRune(c)
		}
	}
	return b.String()
}

func containsName(tests []struct {
	name string
	args []string
	want []string
}, name string,
) bool {
	for _, test := range tests {
		if test.name == name {
			return true
		}
	}
	return false
}

func TestExamples_cat(t *testing.T) {
	// cat with no FILE reads stdin, so a real file is required here
	file := t.TempDir() + "/cat.txt"
	require.NoError(t, os.WriteFile(file, []byte("meow\n"), 0o600))

	runExample(t, "cat", []string{"meow"}, file)
}

// runExampleFailure runs one example in a subprocess and requires a non-zero exit.
func runExampleFailure(t *testing.T, name string, args ...string) {
	t.Helper()

	cmd := exec.Command(os.Args[0])
	cmd.Env = append(os.Environ(),
		"RUN_EXAMPLE="+name,
		"EXAMPLE_ARGS="+strings.Join(args, "\x1f"),
	)
	out, err := cmd.CombinedOutput()
	require.Error(t, err, "%s: args %q, output:\n%s", name, args, out)
}

func TestExamples_failures(t *testing.T) {
	// guard failure: at_least_two requires at least two -f
	runExampleFailure(t, "at_least_two", "-f")
	// missing required token
	runExampleFailure(t, "confusing", "complex1", "42")
	// guard failure: verbose caps at 3
	runExampleFailure(t, "verbose", "-vvvv")
	// invalid enum value
	runExampleFailure(t, "enum_in_args", "--baz", "nope")
}
