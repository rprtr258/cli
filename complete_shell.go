package cli

import (
	"fmt"
	"slices"
	"strings"

	"github.com/rprtr258/fun"
	"github.com/rprtr258/fun/iter"
)

// Shell identifies a supported interactive shell, such as "bash", "zsh" or "fish".
type Shell string

func (self Shell) fmt(f *strings.Builder) {
	f.WriteByte('\'')
	for _, c := range self {
		if c == '\'' {
			f.WriteString("'\\''")
		} else {
			f.WriteRune(c)
		}
	}
	f.WriteByte('\'')
}

// ShellComp is a completion callback for shell completion of the current word.
//
// Implement it to provide custom completion logic and pass it to
// [Parser.CompleteShell].
type ShellComp func(
	// A file or directory name with an optional file mask.
	//
	// For bash filemask should start with `*.` or contain only the
	// extension
	File func(
		// Optional filemask to use, no spaces, no tabs
		mask Opt[string],
	),

	// Similar to `File` but limited to directories only
	// For bash filemask should start with `*.` or contain only the
	// extension
	Dir func(
		// Optional filemask to use, no spaces, no tabs
		mask Opt[string],
	),

	// You can also specify a raw value to use for each supported shell
	//
	// It is possible to fill in values for shells you don't want to support
	// with empty strings but the code is not going to work for those shells
	Raw func(
		// This raw string will be used for `bash` shell
		// <https://www.gnu.org/software/bash/manual/html_node/Command-Line-Editing.html>
		bash,

		// This raw string will be used for `zsh` shell
		// <https://zsh.sourceforge.io/Doc/Release/Completion-System.html>
		zsh,

		// This raw string will be used for `fish` shell
		// <https://fishshell.com/docs/current/completions.html>
		fish,

		// This raw string will be used for `elvish` shell
		// <https://elv.sh/ref/edit.html#completion-api>
		elvish string,
	),

	// Don't produce anything from this parser - can be useful if you want to compose
	// this library's completion with shell completion
	Nothing func(),
)

// ShellCompFile completes with a file or directory name matching an optional mask
func ShellCompFile(mask Opt[string]) ShellComp {
	return func(
		f func(mask Opt[string]),
		_ func(mask Opt[string]),
		_ func(_, _, _, _ string),
		_ func(),
	) {
		if f != nil {
			f(mask)
		}
	}
}

// ShellCompDir completes with a directory name matching an optional mask
func ShellCompDir(mask Opt[string]) ShellComp {
	return func(
		_ func(mask Opt[string]),
		f func(mask Opt[string]),
		_ func(_, _, _, _ string),
		_ func(),
	) {
		if f != nil {
			f(mask)
		}
	}
}

func render_zsh(
	items []ShowComp,
	ops []ShellComp,
	full_lit string,
) string {
	if len(items) == 0 && len(ops) == 0 {
		return fmt.Sprintf("compadd -- %v\n", full_lit)
	}

	res := &strings.Builder{}
	for _, op := range ops {
		op(
			func(mask Opt[string]) {
				if mask, ok := mask.Unpack(); ok {
					fmt.Fprintf(res, "_files -g %v\n", Shell(mask))
				} else {
					fmt.Fprintln(res, "_files")
				}
			},
			func(mask Opt[string]) {
				if mask, ok := mask.Unpack(); ok {
					fmt.Fprintf(res, "_files -/ -g %v\n", Shell(mask))
				} else {
					fmt.Fprintln(res, "_files -/")
				}
			},
			func(
				_ string,
				zsh string,
				_ string,
				_ string,
			) {
				fmt.Fprintln(res, zsh)
			},
			nil,
		)
	}

	if len(items) == 1 {
		if items[0].subst == "" {
			fmt.Fprintf(res, "compadd -- %v\n", Shell(items[0].pretty))
			fmt.Fprintln(res, "compadd ''")
			return res.String()
		} else {
			return fmt.Sprintf("compadd -- %v\n", Shell(items[0].subst))
		}
	}
	fmt.Fprintln(res, "local -a descr")

	for _, item := range items {
		fmt.Fprintf(res, "descr=(%v)\n", Shell(item.to_string()))
		fmt.Fprintf(res, "args=(%q)\n", item.subst)
		if group, ok := item.extra.group.Unpack(); ok {
			fmt.Fprintf(
				res,
				"compadd -l -d descr -V %v -X %v -- %v\n",
				Shell(group),
				Shell(group),
				Shell(item.subst),
			)
		} else {
			// it seems sorting as well as not sorting is done in a group,
			// by default group contains just one element so and `-o nosort`
			// does nothing, while `-V whatever` stops sorting...
			fmt.Fprintf(
				res,
				"compadd -l -V nosort -d descr -- %v\n",
				Shell(item.subst),
			)
		}
	}
	return res.String()
}

func render_bash(
	items []ShowComp,
	ops []ShellComp,
	full_lit string,
) string {
	// Bash is strange when it comes to completion - rather than taking
	// a glob - _filedir takes an extension which it later to include uppercase
	// version as well and to include "*." in front. For compatibility with
	// zsh and other shells - this code strips "*." from the beginning....
	//
	// Second difference between bash and zsh is that if you are trying to
	// allow for multiple extensions zsh takes a sane "*.(foo|bar|baz)" approach,
	// while bash wants it to be "@(foo|bar|baz)"
	//
	// This doesn't cover all the possible masks; the right way of handling
	// them might be ignoring the shell machinery altogether, but for now
	// this translation is what we have.
	//
	bashmask := func(i string) string {
		ii := strip_prefix(i, "*.").OrDefault(i)

		if starts_withb(ii, '(') {
			return fmt.Sprintf("@%v", ii)
		} else {
			return ii
		}
	}

	if len(items) == 0 && len(ops) == 0 {
		return fmt.Sprintf("COMPREPLY+=(%v)\n", Shell(full_lit))
	}

	const init = "local cur prev words cword ; _init_completion || return ;"

	res := &strings.Builder{}
	for _, op := range ops {
		op(
			func(mask Opt[string]) {
				if mask, ok := mask.Unpack(); ok {
					fmt.Fprintf(res, "%v _filedir %v\n", init, Shell(bashmask(mask)))
				} else {
					fmt.Fprintf(res, "%v _filedir", init)
				}
			},
			func(mask Opt[string]) {
				if mask, ok := mask.Unpack(); ok {
					fmt.Fprintf(res, "%v _filedir -d %v\n", init, Shell(bashmask(mask)))
				} else {
					fmt.Fprintf(res, "%v _filedir -d", init)
				}
			},
			func(bash string, _ string, _ string, _ string) { fmt.Fprintln(res, bash) },
			nil,
		)
	}

	if len(items) == 1 {
		if items[0].subst == "" {
			fmt.Fprintf(res, "COMPREPLY+=( %v '')\n", Shell(items[0].pretty))
		} else {
			fmt.Fprintf(res, "COMPREPLY+=( %v )\n", Shell(items[0].subst))
		}

		return res.String()
	}
	prev := ""
	for _, item := range items {
		if group, ok := item.extra.group.Unpack(); ok {
			if prev != group {
				prev = group
				fmt.Fprintf(res, "COMPREPLY+=(%v)\n", Shell(group))
			}
		}
		fmt.Fprintf(res, "COMPREPLY+=(%v)\n", Shell(item.to_string()))
	}

	return res.String()
}

// debug_string renders the shell completion callback the way it is shown in
// test output.
func (self ShellComp) debug_string() string {
	option_str := func(o Opt[string]) string {
		if s, ok := o.Unpack(); ok {
			return fmt.Sprintf("Some(%#v)", s)
		}
		return "None"
	}

	res := ""
	self(
		func(mask Opt[string]) { res = fmt.Sprintf("File { mask: %v }", option_str(mask)) },
		func(mask Opt[string]) { res = fmt.Sprintf("Dir { mask: %v }", option_str(mask)) },
		func(bash, zsh, fish, elvish string) {
			res = fmt.Sprintf("Raw { bash: %#v, zsh: %#v, fish: %#v, elvish: %#v }", bash, zsh, fish, elvish)
		},
		func() { res = "Nothing" },
	)
	return res
}

func render_test(
	items []ShowComp,
	ops []ShellComp,
	lit string,
) string {
	if len(items) == 0 && len(ops) == 0 {
		return fmt.Sprintf("%v\n", lit)
	}

	if len(items) == 1 && len(ops) == 0 && items[0].subst != "" {
		return items[0].subst
	}

	res := &strings.Builder{}
	for _, op := range items {
		fmt.Fprintf(
			res,
			"%v\t%v\t%v\t%v\n",
			op.subst,
			op.pretty,
			op.extra.group.OrDefault(""),
			op.extra.help.OrDefault(""),
		)
	}
	fmt.Fprintln(res)
	for _, op := range ops {
		fmt.Fprintf(res, "%v\n", op.debug_string())
	}

	return res.String()
}

func render_fish(
	items []ShowComp,
	ops []ShellComp,
	full_lit string,
	app string,
) string {
	if len(items) == 0 && len(ops) == 0 {
		return fmt.Sprintf("complete -c %v --arguments=%v", app, full_lit)
	}

	res := &strings.Builder{}
	shared := fun.IF(len(ops) == 0, "-f ", "")
	for item := range iter.Seq2[int, ShowComp](slices.Backward(items)).Values().Filter(func(i ShowComp) bool { return i.subst != "" }) {
		fmt.Fprintf(res, "complete -c %v %v", app, shared)
		if long, ok := strip_prefix(item.subst, "--").Unpack(); ok {
			fmt.Fprintf(res, "--long-option %v ", long)
		} else if short, ok := strip_prefixb(item.subst, '-').Unpack(); ok {
			fmt.Fprintf(res, "--short-option %v ", short)
		} else {
			fmt.Fprintf(res, "-a %v ", item.subst)
		}
		if help, ok := item.extra.help.Unpack(); ok {
			fmt.Fprintf(res, "-d %#v", help)
		}
		fmt.Fprintln(res)
	}
	return res.String()
}

func render_simple(items []ShowComp) string {
	res := &strings.Builder{}
	if len(items) == 1 {
		fmt.Fprintln(res, items[0].subst)
	} else {
		for _, item := range items {
			if descr, ok := item.extra.help.Unpack(); ok {
				fmt.Fprintf(res,
					"%v\t%v\n",
					item.subst,
					first(strings.Split(descr, "\n")).OrDefault(""),
				)
			} else {
				fmt.Fprintln(res, item.subst)
			}
		}
	}
	return res.String()
}
