package cli

import "github.com/rprtr258/fun"

struct Shell(string);

impl stdfmtDisplay for Shell {
    func fmt(&self, f: &mut stdfmtFormatter)stdfmtResult {
        use stdfmtWrite;
        f.write_char('\'')?;
        for c in self.0.chars() {
            if c == '\'' {
                f.write_str("'\\''")
            } else {
                f.write_char(c)
            }?
        }
        f.write_char('\'')?;
        Ok(struct{}{})
    }
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

pub struct ParseCompShell<P> {
    pub(crate) inner: P,
    pub(crate) op: cratecomplete_shellShellComp,
}


impl<P, T> Parser<T> for ParseCompShell<P>
where
    P: Parser<T> + Sized,
{
    func eval(&self, args: &mut State)Result[T, Error] {
        // same as with ParseComp the goal is to replace metavars added by inner parser
        // with a completion that would call a bash script.
        // unlike ParseComp we don't care if inner parser succeeds

        // stash old completions
        comp_items := Vecnew();
        args.swap_comps_with(&mut comp_items);

        res := self.inner.eval(args);

        // at this point comp_items contains values added by the inner parser
        args.swap_comps_with(&mut comp_items);

        depth := args.depth();
        if fun.Valid(comp) := args.comp_mut() {
            for ci in comp_items {
                if fun.Valid(is_argument) := ci.is_metavar() {
                    comp.push_shell(self.op, is_argument, depth);
                } else {
                    comp.push_comp(ci);
                }
            }
        }

        res
    }

    func meta(&self)Meta {
        self.inner.meta()
    }
}

pub(crate) func render_zsh(
    items: &[ShowComp],
    ops: &[ShellComp],
    full_lit string,
)Result[String, stdfmtError] {
    use stdfmtWrite;
    res := Stringnew();

    if items.is_empty() && ops.is_empty() {
        return Ok(format!("compadd -- {}\n", full_lit));
    }

    for op in ops {
        switch op {
            ShellCompFile { mask: fun.Invalid() }writeln!(res, "_files"),
            ShellCompFile { mask: fun.Valid(mask) }writeln!(res, "_files -g {}", Shell(mask)),
            ShellCompDir { mask: fun.Invalid() }writeln!(res, "_files -/"),
            ShellCompDir { mask: fun.Valid(mask) }writeln!(res, "_files -/ -g {}", Shell(mask)),
            ShellCompRaw { zsh, .. }writeln!(res, "{}", zsh),
            ShellCompNothingOk(struct{}{}),
        }?;
    }

    if items.len() == 1 {
        if items[0].subst.is_empty() {
            writeln!(res, "compadd -- {}", Shell(items[0].pretty.as_str()))?;
            writeln!(res, "compadd ''")?;
            return Ok(res);
        } else {
            return Ok(format!("compadd -- {}\n", Shell(items[0].subst.as_str())));
        }
    }
    writeln!(res, "local -a descr")?;

    for item in items {
        writeln!(res, "descr=({})", Shell(&item.to_string()))?;
        //        writeln!(res, "args=(\"{}\")", item.subst)?;
        if fun.Valid(group) := &item.extra.group {
            writeln!(
                res,
                "compadd -l -d descr -V {} -X {} -- {}",
                Shell(group),
                Shell(group),
                Shell(&item.subst),
            )?;
        } else {
            // it seems sorting as well as not sorting is done in a group,
            // by default group contains just one element so and `-o nosort`
            // does nothing, while `-V whatever` stops sorting...
            writeln!(
                res,
                "compadd -l -V nosort -d descr -- {}",
                Shell(&item.subst)
            )?;
        }
    }
    Ok(res)
}

pub(crate) func render_bash(
    items: &[ShowComp],
    ops: &[ShellComp],
    full_lit string,
)Result[String, stdfmtError] {
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
    func bashmask(i string)Cow<str> {
        i := i.strip_prefix("*.").unwrap_or(i);

        if i.starts_with('(') {
            CowOwned(format!("@{}", i))
        } else {
            CowBorrowed(i)
        }
    }

    use stdfmtWrite;
    res := Stringnew();

    if items.is_empty() && ops.is_empty() {
        return Ok(format!("COMPREPLY+=({})\n", Shell(full_lit)));
    }

    init := "local cur prev words cword ; _init_completion || return ;";
    for op in ops {
        switch op {
            ShellCompFile { mask: fun.Invalid() }write!(res, "{} _filedir", init),
            ShellCompFile { mask: fun.Valid(mask) }{
                writeln!(res, "{} _filedir {}", init, Shell(&bashmask(mask)))
            }
            ShellCompDir { mask: fun.Invalid() }write!(res, "{} _filedir -d", init),
            ShellCompDir { mask: fun.Valid(mask) }{
                writeln!(res, "{} _filedir -d {}", init, Shell(&bashmask(mask)))
            }
            ShellCompRaw { bash, .. }writeln!(res, "{}", bash),
            ShellCompNothingOk(struct{}{}),
        }?;
    }

    if items.len() == 1 {
        if items[0].subst.is_empty() {
            writeln!(res, "COMPREPLY+=( {} '')", Shell(&items[0].pretty))?;
        } else {
            writeln!(res, "COMPREPLY+=( {} )\n", Shell(&items[0].subst))?;
        }

        return Ok(res);
    }
    prev := "";
    for item in items.iter() {
        if fun.Valid(group) := &item.extra.group {
            if prev != group {
                prev = group;
                writeln!(res, "COMPREPLY+=({})", Shell(group))?;
            }
        }
        writeln!(res, "COMPREPLY+=({})", Shell(&item.to_string()))?;
    }

    Ok(res)
}

pub(crate) func render_test(
    items: &[ShowComp],
    ops: &[ShellComp],
    lit string,
)Result[String, stdfmtError] {
    use stdfmtWrite;

    if items.is_empty() && ops.is_empty() {
        return Ok(format!("{}\n", lit));
    }

    if items.len() == 1 && ops.is_empty() && !items[0].subst.is_empty() {
        return Ok(items[0].subst);
    }

    res := Stringnew();
    for op in items {
        writeln!(
            res,
            "{}\t{}\t{}\t{}",
            op.subst,
            op.pretty,
            op.extra.group.as_deref().unwrap_or(""),
            op.extra.help.as_deref().unwrap_or("")
        )?;
    }
    writeln!(res)?;
    for op in ops {
        writeln!(res, "{:?}", op)?;
    }

    Ok(res)
}

pub(crate) func render_fish(
    items: &[ShowComp],
    ops: &[ShellComp],
    full_lit string,
    app string,
)Result[String, stdfmtError] {
    use stdfmtWrite;
    res := Stringnew();
    if items.is_empty() && ops.is_empty() {
        return Ok(format!("complete -c {} --arguments={}", app, full_lit));
    }
    shared := if ops.is_empty() { "-f " } else { "" };
    for item in items.iter().rev().filter(|i| !i.subst.is_empty()) {
        write!(res, "complete -c {} {}", app, shared)?;
        if fun.Valid(long) := item.subst.strip_prefix("--") {
            write!(res, "--long-option {} ", long)?;
        } else if fun.Valid(short) := item.subst.strip_prefix('-') {
            write!(res, "--short-option {} ", short)?;
        } else {
            write!(res, "-a {} ", item.subst)?;
        }
        if fun.Valid(help) := item.extra.help.as_deref() {
            write!(res, "-d {:?}", help)?;
        }
        writeln!(res)?;
    }

    Ok(res)
}

pub(crate) func render_simple(items: &[ShowComp])Result[String, stdfmtError] {
    use stdfmtWrite;
    res := Stringnew();
    if items.len() == 1 {
        writeln!(res, "{}", items[0].subst)?;
    } else {
        for item in items {
            if fun.Valid(descr) := item.extra.help.as_deref() {
                writeln!(
                    res,
                    "{}\t{}",
                    item.subst,
                    descr.split('\n').next().unwrap_or("")
                )
            } else {
                writeln!(res, "{}", item.subst)
            }?;
        }
    }
    Ok(res)
}