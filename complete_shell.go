package cli

import (
	"fmt"
	"strings"

	"github.com/rprtr258/fun"
)

type Shell string

func (self Shell) fmt(f *stdfmtFormatter) error {
	if _, err := f.write_char('\''); err != nil {
		return err
	}
	for _, c := range self {
		if c == '\'' {
			if _, err := f.write_str("'\\''"); err != nil {
				return err
			}
		} else {
			if _, err := f.write_char(c); err != nil {
				return err
			}
		}
	}
	if _, err := f.write_char('\''); err != nil {
		return err
	}
	return nil
}

type ShellComp func(
	// A file or directory name with an fun.Optional file mask.
	//
	// For bash filemask should start with `*.` or contain only the
	// extension
	File func(
		// fun.Optional filemask to use, no spaces, no tabs
		mask fun.Option[string],
	),

	// Similar to `File` but limited to directories only
	// For bash filemask should start with `*.` or contain only the
	// extension
	Dir func(
		// fun.Optional filemask to use, no spaces, no tabs
		mask fun.Option[string],
	),

	// You can also specify a raw value to use for each supported shell
	//
	// It is possible to fill in values for shells you don't want to support
	// with empty strings but the code is not going to work for those shells
	Raw func(
		// This raw string will be used for `bash` shell
		// <https://www.gnu.org/software/bash/manual/html_node/Command-Line-Editing.html>
		bash string,

		// This raw string will be used for `zsh` shell
		// <https://zsh.sourceforge.io/Doc/Release/Completion-System.html>
		zsh string,

		// This raw string will be used for `fish` shell
		// <https://fishshell.com/docs/current/completions.html>
		fish string,

		// This raw string will be used for `elvish` shell
		// <https://elv.sh/ref/edit.html#completion-api>
		elvish string,
	),

	// Don't produce anything at all from this parser - can be useful if you want to compose
	// bpaf completion with shell completion
	Nothing func(),
)

// Parser that inserts static shell completion into bpaf's dynamic shell completion

type ParseCompShell [P]struct {
	inner P
	op    ShellComp
}

func eval[P Parser[T]](self ParseCompShell[P], args *State) Result[T] {
	// same as with ParseComp the goal is to replace metavars added by inner parser
	// with a completion that would call a bash script.
	// unlike ParseComp we don't care if inner parser succeeds

	// stash old completions
	comp_items := Vecnew()
	args.swap_comps_with(&comp_items)

	res := self.inner.eval(args)

	// at this point comp_items contains values added by the inner parser
	args.swap_comps_with(&comp_items)

	depth := args.depth()
	if comp, ok := args.comp_mut().Unpack(); ok {
		for _, ci := range comp_items {
			if is_argument, ok := ci.is_metavar().Unpack(); ok {
				comp.push_shell(self.op, is_argument, depth)
			} else {
				comp.push_comp(ci)
			}
		}
	}

	return res
}

func meta[P Parser[T]](self ParseCompShell[P]) Meta {
	return self.inner.meta()
}

func render_zsh(
	items []ShowComp,
	ops []ShellComp,
	full_lit string,
) Result[string, error] {
	res := Stringnew()

	if items.is_empty() && ops.is_empty() {
		return Ok(fmt.Sprintf("compadd -- {}\n", full_lit))
	}

	for _, op := range ops {
		var err error
		op(
			func(mask fun.Option[string]) {
				if mask, ok := mask.Unpack(); ok {
					_, err = fmt.Fprintln(res, "_files -g {}", Shell(mask))
				} else {
					_, err = fmt.Fprintln(res, "_files")
				}
			},
			func(mask fun.Option[string]) {
				if mask, ok := mask.Unpack(); ok {
					_, err = fmt.Fprintln(res, "_files -/ -g {}", Shell(mask))
				} else {
					_, err = fmt.Fprintln(res, "_files -/")
				}
			},
			func(
				_ string,
				zsh string,
				_ string,
				_ string,
			) {
				_, err = fmt.Fprintln(res, "{}", zsh)
			},
			nil,
		)
		if err != nil {
			return err
		}
	}

	if items.len() == 1 {
		if items[0].subst.is_empty() {
			if _, err := fmt.Fprintln(res, "compadd -- {}", Shell(items[0].pretty.as_str())); err != nil {
				return err
			}
			if _, err := fmt.Fprintln(res, "compadd ''"); err != nil {
				return err
			}
			return res, nil
		} else {
			return fmt.Sprintf("compadd -- {}\n", Shell(items[0].subst.as_str())), nil
		}
	}
	if _, err := fmt.Fprintln(res, "local -a descr"); err != nil {
		return err
	}

	for _, item := range items {
		if _, err := fmt.Fprintln(res, "descr=({})", Shell(&item.to_string())); err != nil {
			return err
		}
		//        fmt.Fprintln(res, "args=(\"{}\")", item.subst)?;
		if group, ok := item.extra.group.Unpack(); ok {
			if _, err := fmt.Fprintln(
				res,
				"compadd -l -d descr -V {} -X {} -- {}",
				Shell(group),
				Shell(group),
				Shell(&item.subst),
			); err != nil {
				return err
			}
		} else {
			// it seems sorting as well as not sorting is done in a group,
			// by default group contains just one element so and `-o nosort`
			// does nothing, while `-V whatever` stops sorting...
			if _, err := fmt.Fprintln(
				res,
				"compadd -l -V nosort -d descr -- {}",
				Shell(&item.subst),
			); err != nil {
				return err
			}
		}
	}
	Ok(res)
}

func render_bash(
	items []ShowComp,
	ops []ShellComp,
	full_lit string,
) Result[string, error] {
	// Bash is strange when it comes to completion - rather than taking
	// a glob - _filedir takes an extension which it later to include uppercase
	// version as well and to include "*." in front. For compatibility with
	// zsh and other shells - this code strips "*." from the beginning....
	//
	// Second difference between bash and zsh is that if you are trying to
	// allow for multiple extensions zsh takes a sane "*.(foo|bar|baz)" approach,
	// while bash wants it to be "@(foo|bar|baz)"
	//
	// This doesn't cover all the possible masks, I suspect that the right way of
	// handling this would be ignoring the shell machinery and handling masks on the
	// Rust side... But for now try this
	//
	bashmask := func(i string) string {
		i := i.strip_prefix("*.").unwrap_or(i)

		if i.starts_with('(') {
			return fmt.Sprintf("@{}", i)
		} else {
			return i
		}
	}

	res := Stringnew()

	if items.is_empty() && ops.is_empty() {
		return Ok(fmt.Sprintf("COMPREPLY+=({})\n", Shell(full_lit)))
	}

	init := "local cur prev words cword ; _init_completion || return ;"
	for _, op := range ops {
		op(
			func(mask fun.Option[string]) {
				if mask, ok := mask.Unpack(); ok {
					fmt.Fprintln(res, "{} _filedir {}", init, Shell(&bashmask(mask)))
				} else {
					fmt.Fprintf(res, "{} _filedir", init)
				}
			},
			func(mask fun.Option[string]) {
				if mask, ok := mask.Unpack(); ok {
					fmt.Fprintln(res, "{} _filedir -d {}", init, Shell(&bashmask(mask)))
				} else {
					fmt.Fprintf(res, "{} _filedir -d", init)
				}
			},
			func(bash string, _ string, _ string, _ string) { fmt.Fprintln(res, "{}", bash) },
			nil,
		)
	}

	if items.len() == 1 {
		if items[0].subst.is_empty() {
			fmt.Fprintln(res, "COMPREPLY+=( {} '')", Shell(&items[0].pretty))
		} else {
			fmt.Fprintln(res, "COMPREPLY+=( {} )\n", Shell(&items[0].subst))
		}

		return Ok(res)
	}
	prev := ""
	for _, item := range items.iter() {
		if group, ok := item.extra.group.Unpack(); ok {
			if prev != group {
				prev = group
				fmt.Fprintln(res, "COMPREPLY+=({})", Shell(group))
			}
		}
		fmt.Fprintln(res, "COMPREPLY+=({})", Shell(&item.to_string()))
	}

	Ok(res)
}

func render_test(
	items []ShowComp,
	ops []ShellComp,
	lit string,
) Result[string, stdfmtError] {
	if items.is_empty() && ops.is_empty() {
		return Ok(fmt.Sprintf("{}\n", lit))
	}

	if items.len() == 1 && ops.is_empty() && !items[0].subst.is_empty() {
		return Ok(items[0].subst)
	}

	res := Stringnew()
	for _, op := range items {
		fmt.Fprintln(
			res,
			"{}\t{}\t{}\t{}",
			op.subst,
			op.pretty,
			op.extra.group.as_deref().unwrap_or(""),
			op.extra.help.as_deref().unwrap_or(""),
		)
	}
	fmt.Fprintln(res)
	for _, op := range ops {
		fmt.Fprintln(res, "%#v", op)
	}

	Ok(res)
}

func render_fish(
	items []ShowComp,
	ops []ShellComp,
	full_lit string,
	app string,
) Result[string, stdfmtError] {
	res := Stringnew()
	if items.is_empty() && ops.is_empty() {
		return Ok(fmt.Sprintf("complete -c {} --arguments={}", app, full_lit))
	}
	shared := fun.IF(ops.is_empty(), "-f ", "")
	for _, item := range items.iter().rev().filter(func(i) bool { return !i.subst.is_empty() }) {
		fmt.Fprintf(res, "complete -c {} {}", app, shared)
		if long, ok := item.subst.strip_prefix("--").Unpack(); ok {
			fmt.Fprintf(res, "--long-option {} ", long)
		} else if short, ok := item.subst.strip_prefix('-').Unpack(); ok {
			fmt.Fprintf(res, "--short-option {} ", short)
		} else {
			fmt.Fprintf(res, "-a {} ", item.subst)
		}
		if help, ok := item.extra.help.Unpack(); ok {
			fmt.Fprintf(res, "-d %#v", help)
		}
		fmt.Fprintln(res)
	}

	Ok(res)
}

func render_simple(items []ShowComp) string {
	var res strings.Builder
	if len(items) == 1 {
		fmt.Fprintln(&res, items[0].subst)
	} else {
		for _, item := range items {
			if descr, ok := item.extra.help.Unpack(); ok {
				fmt.Fprintln(&res,
					"{}\t{}",
					item.subst,
					strings.Split(descr, "\n").next().unwrap_or(""),
				)
			} else {
				fmt.Fprintln(&res, item.subst)
			}
		}
	}
	return res.String()
}
