package flags

import (
	"bytes"
	"path"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"testing"

	"github.com/rprtr258/fun"
	"github.com/stretchr/testify/assert"
)

type TestComplete struct{}

func (t *TestComplete) Complete(match string) []Completion {
	options := []string{
		"hello world",
		"hello universe",
		"hello multiverse",
	}

	ret := make([]Completion, 0, len(options))

	for _, o := range options {
		if strings.HasPrefix(o, match) {
			ret = append(ret, Completion{
				Item: o,
			})
		}
	}

	return ret
}

var completionTestOptions struct {
	Verbose  bool `short:"v" long:"verbose" description:"Verbose messages"`
	Debug    bool `short:"d" long:"debug" description:"Enable debug"`
	Info     bool `short:"i" description:"Display info"`
	Version  bool `long:"version" description:"Show version"`
	Required bool `long:"required" required:"true" description:"This is required"`
	Hidden   bool `long:"hidden" hidden:"true" description:"This is hidden"`

	AddCommand struct {
		Positional struct {
			Filename Filename
		} `positional-args:"yes"`
	} `command:"add" description:"add an item"`

	AddMultiCommand struct {
		Positional struct {
			Filename []Filename
		} `positional-args:"yes"`
		Extra []Filename `short:"f"`
	} `command:"add-multi" description:"add multiple items"`

	AddMultiCommandFlag struct {
		Files []Filename `short:"f"`
	} `command:"add-multi-flag" description:"add multiple items via flags"`

	RemoveCommand struct {
		Other bool     `short:"o"`
		File  Filename `short:"f" long:"filename"`
	} `command:"rm" description:"remove an item"`

	RenameCommand struct {
		Completed TestComplete `short:"c" long:"completed"`
	} `command:"rename" description:"rename an item"`

	HiddenCommand struct{} `command:"hidden" description:"hidden command" hidden:"true"`
}

type completionTest struct {
	Args      []string
	Completed []string
}

var completionTests = func() map[string]completionTest {
	_, sourcefile, _, _ := runtime.Caller(0)
	completionTestSourcedir := filepath.Join(filepath.SplitList(path.Dir(sourcefile))...)

	completionTestFilename := []string{filepath.Join(completionTestSourcedir, "completion.go"), filepath.Join(completionTestSourcedir, "completion_test.go")}

	completionTestSubdir := []string{
		filepath.Join(completionTestSourcedir, "examples", "add.go"),
		filepath.Join(completionTestSourcedir, "examples", "bash-completion"),
		filepath.Join(completionTestSourcedir, "examples", "main.go"),
		filepath.Join(completionTestSourcedir, "examples", "rm.go"),
	}

	return map[string]completionTest{
		"Short names": {
			[]string{"-"},
			[]string{
				"--debug     # Enable debug",
				"--required  # This is required",
				"--verbose   # Verbose messages",
				"--version   # Show version",
				"-i          # Display info",
			},
		},
		"Short names full": {
			[]string{"-i"},
			[]string{"-i"},
		},
		"Short names concatenated": {
			[]string{"-dv"},
			[]string{"-dv"},
		},
		"Long names with descriptions": {
			[]string{"--"},
			[]string{
				"--debug     # Enable debug",
				"--required  # This is required",
				"--verbose   # Verbose messages",
				"--version   # Show version",
			},
		},
		"Long names partial": {
			[]string{"--ver"},
			[]string{
				"--verbose  # Verbose messages",
				"--version  # Show version",
			},
		},
		"Commands with descriptions": {
			[]string{""},
			[]string{
				"add             # add an item",
				"add-multi       # add multiple items",
				"add-multi-flag  # add multiple items via flags",
				"rename          # rename an item",
				"rm              # remove an item",
			},
		},
		"Commands partial": {
			[]string{"r"},
			[]string{
				"rename  # rename an item",
				"rm      # remove an item",
			},
		},
		"Positional filename": {
			[]string{"add", filepath.Join(completionTestSourcedir, "completion")},
			completionTestFilename,
		},
		"Multiple positional filename (1 arg)": {
			[]string{"add-multi", filepath.Join(completionTestSourcedir, "completion")},
			completionTestFilename,
		},
		"Multiple positional filename (2 args)": {
			[]string{"add-multi", filepath.Join(completionTestSourcedir, "completion.go"), filepath.Join(completionTestSourcedir, "completion")},
			completionTestFilename,
		},
		"Multiple positional filename (3 args)": {
			[]string{"add-multi", filepath.Join(completionTestSourcedir, "completion.go"), filepath.Join(completionTestSourcedir, "completion.go"), filepath.Join(completionTestSourcedir, "completion")},
			completionTestFilename,
		},
		"Flag filename": {
			[]string{"rm", "-f", path.Join(completionTestSourcedir, "completion")},
			completionTestFilename,
		},
		"Flag short concat last filename": {
			[]string{"rm", "-of", path.Join(completionTestSourcedir, "completion")},
			completionTestFilename,
		},
		"Flag concat filename": {
			[]string{"rm", "-f" + path.Join(completionTestSourcedir, "completion")},
			[]string{"-f" + completionTestFilename[0], "-f" + completionTestFilename[1]},
		},
		"Flag equal concat filename": {
			[]string{"rm", "-f=" + path.Join(completionTestSourcedir, "completion")},
			[]string{"-f=" + completionTestFilename[0], "-f=" + completionTestFilename[1]},
		},
		"Flag concat long filename": {
			[]string{"rm", "--filename=" + path.Join(completionTestSourcedir, "completion")},
			[]string{"--filename=" + completionTestFilename[0], "--filename=" + completionTestFilename[1]},
		},
		"Flag long filename": {
			[]string{"rm", "--filename", path.Join(completionTestSourcedir, "completion")},
			completionTestFilename,
		},
		"To subdir": {
			[]string{"rm", "--filename", path.Join(completionTestSourcedir, "examples/bash-")},
			[]string{path.Join(completionTestSourcedir, "examples/bash-completion/")},
		},
		"Subdirectory": {
			[]string{"rm", "--filename", path.Join(completionTestSourcedir, "examples") + "/"},
			completionTestSubdir,
		},
		"Custom completed": {
			[]string{"rename", "-c", "hello un"},
			[]string{"hello universe"},
		},
		"Multiple flag filename": {
			[]string{"add-multi-flag", "-f", filepath.Join(completionTestSourcedir, "completion")},
			completionTestFilename,
		},
	}
}()

func TestCompletion(t *testing.T) {
	p := NewParser(&completionTestOptions, Default)
	c := &completion{parser: p}

	for name, test := range completionTests {
		t.Run(name, func(t *testing.T) {
			items := fun.Map[string](func(v Completion) string {
				return v.Item
			}, c.complete(test.Args)...)

			assert.Equal(t, len(test.Completed), len(items))
			sort.Strings(items)
			sort.Strings(test.Completed) // TODO: remove cringe
			for i, v := range items {
				assert.True(t, strings.HasPrefix(test.Completed[i], v), "%q %q", test.Completed[i], v)
			}
		})
	}
}

func TestParserCompletion(t *testing.T) {
	for name, test := range completionTests {
		t.Run(name, func(t *testing.T) {
			t.Setenv("GO_FLAGS_COMPLETION", "1")

			var buf bytes.Buffer
			p := NewParser(&completionTestOptions, None)
			p.CompletionHandler = func(items []Completion) {
				comp := &completion{parser: p}
				comp.print(&buf, items)
			}

			_, err := p.Run(test.Args...)
			assert.NoError(t, err)

			got := strings.Split(buf.String(), "\n")
			assert.ElementsMatch(t, test.Completed, got[:len(got)-1])
		})
	}
}
