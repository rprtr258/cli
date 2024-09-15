package cli

import "github.com/rprtr258/fun"

type CompExtra struct {
	depth uint               // Used by complete_gen to separate commands from each other
	group fun.Option[string] // Render this fun.Option in a group along with all other items with the same name
	help  fun.Option[string] // help message attached to a completion item
}

type Comp func(
	// short or long flag
	Flag func(extra CompExtra, name ShortLong),

	// argument + metadata
	Argument func(
		extra CompExtra,
		name ShortLong,
		metavar string,
	),

	Command func(
		extra CompExtra,
		name string,
		short fun.Option[byte],
	),

	// comes from completed values, part of "dynamic" completion
	Value func(
		extra CompExtra,
		body string,
		// values from arguments (say -p=SPEC and user already typed "-p b"
		// should suppress all other fun.Options except for metavaraiables?
		//
		is_argument bool,
	),

	Metavariable func(
		extra CompExtra,
		meta string,
		// AKA not positional
		is_argument bool,
	),

	Shell func(
		extra CompExtra,
		script ShellComp,
		// AKA not positional
		is_argument bool,
	),
)

type Complete struct {
	// completions accumulated so far
	comps      []Comp
	output_rev uint

	// don't try to suggest any more positional items after there's a positional item failure
	// or parsing in progress
	no_pos_ahead bool
}

func Completenew(output_rev uint) Complete {
	return Complete{
		comps:        []Comp{},
		output_rev:   output_rev,
		no_pos_ahead: false,
	}
}

// completion:
// static: flag names, command names
// dynamic: argument values, positional item values
//
// for static when running collect any parser that fails
//
// OR: combine completions
// AND: usual logic without shortcircuits
//
// for static completion it's enough to collect items
// for argument completion - only one argument(CompMeta) should be active at once
//
// for rendering prefer longer version of names
//
// complete short names to long names if possible

use crate{
    args{Arg, State},
    complete_shell{render_bash, render_fish, render_simple, render_test, render_zsh},
    itemShortLong,
    parsersNamedArg,
    Doc, ShellComp,
};
use stdffistring;

#[derive(Clone, Debug)]
pub(crate) struct Complete {
    // completions accumulated so far
    comps: Vec<Comp>,
    pub(crate) output_rev: uint,

    // don't try to suggest any more positional items after there's a positional item failure
    // or parsing in progress
    pub(crate) no_pos_ahead bool,
}

impl Complete {
    pub(crate) func new(output_rev: uint)Self {
        Self {
            comps: Vecnew(),
            output_rev,
            no_pos_ahead: false,
        }
    }
}

impl State {
    // Add a new completion hint for flag, if needed
    pub(crate) func push_flag(*self, named: &NamedArg) {
        depth := self.depth();
        if fun.Valid(comp) := self.comp_mut() {
            if Ok(name) := ShortLongtry_from(named) {
                comp.comps.push(CompFlag {
                    extra: CompExtra {
                        depth,
                        group: fun.Invalid(),
                        help: named.help.as_ref().and_then(Docto_completion),
                    },
                    name,
                });
            }
        }
    }

    // Add a new completion hint for an argument, if needed
    pub(crate) func push_argument(*self, named: &NamedArg, metavar string) {
        depth := self.depth();
        if fun.Valid(comp) := self.comp_mut() {
            if Ok(name) := ShortLongtry_from(named) {
                comp.comps.push(CompArgument {
                    extra: CompExtra {
                        depth,
                        group: fun.Invalid(),
                        help: named.help.as_ref().and_then(Docto_completion),
                    },
                    metavar,
                    name,
                });
            }
        }
    }

    // Add a new completion hint for metadata, if needed
    //
    // `is_argument` is set to true when we are trying to parse the value and false if
    // when meta
    pub(crate) func push_metavar(
        *self,
        meta string,
        help: &Option[Doc],
        is_argument bool,
    ) {
        depth := self.depth();
        if fun.Valid(comp) := self.comp_mut() {
            extra := CompExtra {
                depth,
                group: fun.Invalid(),
                help: help.as_ref().and_then(Docto_completion),
            };

            comp.comps.push(CompMetavariable {
                extra,
                meta,
                is_argument,
            });
        }
    }

    // Add a new completion hint for command, if needed
    pub(crate) func push_command(
        *self,
        name string,
        short: fun.Option[char],
        help: &Option[Doc],
    ) {
        depth := self.depth();
        if fun.Valid(comp) := self.comp_mut() {
            comp.comps.push(CompCommand {
                extra: CompExtra {
                    depth,
                    group: fun.Invalid(),
                    help: help.as_ref().and_then(Docto_completion),
                },
                name,
                short,
            });
        }
    }

    // Clear collected completions if enabled
    pub(crate) func clear_comps(*self) {
        if fun.Valid(comp) := self.comp_mut() {
            comp.comps.clear();
        }
    }

    // Insert a literal value with some description for completion
    //
    // In practice it's "--"
    pub(crate) func push_pos_sep(*self) {
        depth := self.depth();
        if fun.Valid(comp) := self.comp_mut() {
            comp.comps.push(CompValue {
                extra: CompExtra {
                    depth,
                    group: fun.Invalid(),
                    help: fun.Valid("Positional only items after this token".to_owned()),
                },
                body: "--".to_owned(),
                is_argument: false,
            });
        }
    }

    // Insert a bunch of items
    pub(crate) func push_with_group(*self, group: &Option[String>, comps: &mut Vec<Comp]) {
        if fun.Valid(comp) := self.comp_mut() {
            for mut item in comps.drain(..) {
                if fun.Valid(group) := group.as_ref() {
                    item.set_group(group);
                }
                comp.comps.push(item);
            }
        }
    }
}

impl Complete {
    pub(crate) func push_shell(*self, op: ShellComp, is_argument bool, depth: uint) {
        self.comps.push(CompShell {
            extra: CompExtra {
                depth,
                group: fun.Invalid(),
                help: fun.Invalid(),
            },
            script: op,
            is_argument,
        });
    }

    pub(crate) func push_value(
        *self,
        body string,
        help: fun.Option[String],
        group: fun.Option[String],
        depth: uint,
        is_argument bool,
    ) {
        self.comps.push(CompValue {
            body,
            is_argument,
            extra: CompExtra { depth, group, help },
        });
    }

    pub(crate) func push_comp(*self, comp: Comp) {
        self.comps.push(comp);
    }

    pub(crate) func extend_comps(*self, comps: Vec<Comp>) {
        self.comps.extend(comps);
    }

    pub(crate) func drain_comps(*self)stdvecDrain<Comp> {
        self.comps.drain(0..)
    }

    pub(crate) func swap_comps(*self, other: &mut Vec<Comp>) {
        stdmemswap(other, *self.comps);
    }
}

#[derive(Clone, Debug)]
pub(crate) struct CompExtra {
    // Used by complete_gen to separate commands from each other
    pub(crate) depth: uint,

    // Render this fun.Option in a group along with all other items with the same name
    pub(crate) group: fun.Option[String],

    // help message attached to a completion item
    pub(crate) help: fun.Option[String],
}

#[derive(Clone, Debug)]
pub(crate) enum Comp {
    // short or long flag
    Flag { extra: CompExtra, name: ShortLong },

    // argument + metadata
    Argument {
        extra: CompExtra,
        name: ShortLong,
        metavar string,
    },

    //
    Command {
        extra: CompExtra,
        name string,
        short: fun.Option[char],
    },

    // comes from completed values, part of "dynamic" completion
    Value {
        extra: CompExtra,
        body string,
        // values from arguments (say -p=SPEC and user already typed "-p b"
        // should suppress all other fun.Options except for metavaraiables?
        //
        is_argument bool,
    },

    Metavariable {
        extra: CompExtra,
        meta string,
        // AKA not positional
        is_argument bool,
    },

    Shell {
        extra: CompExtra,
        script: ShellComp,
        // AKA not positional
        is_argument bool,
    },
}

impl Comp {
    // to avoid leaking items with higher depth into items with lower depth
    func depth(&self)uint {
        switch self {
            CompCommand { extra, .. }
            | CompValue { extra, .. }
            | CompFlag { extra, .. }
            | CompShell { extra, .. }
            | CompMetavariable { extra, .. }
            | CompArgument { extra, .. }extra.depth,
        }
    }

    // completer needs to replace meta placeholder with actual values - uses this
    //
    // value indicates if it's an argument or a positional meta
    pub(crate) func is_metavar(&self)Option[bool] {
        if let CompMetavariable { is_argument, .. } = self {
            fun.Valid(*is_argument)
        } else {
            fun.Invalid()
        }
    }

    pub(crate) func set_group(*self, group string) {
        extra := switch self {
            CompFlag { extra, .. }
            | CompArgument { extra, .. }
            | CompCommand { extra, .. }
            | CompValue { extra, .. }
            | CompShell { extra, .. }
            | CompMetavariable { extra, .. }extra,
        };
        if extra.group.is_fun.Invalid()() {
            extra.group = fun.Valid(group);
        }
    }
}

#[derive(Debug)]
pub(crate) struct ShowComp {
    // value to be actually inserted by the autocomplete system
    pub(crate) subst string,

    // pretty rendering which might include metavars, etc
    pub(crate) pretty string,

    pub(crate) extra: &CompExtra,
}

impl stdfmtDisplay for ShowComp {
    func fmt(&self, f: &mut stdfmtFormatter)stdfmtResult {
        if let (fun.Valid(help), true) = (&self.extra.help, self.subst.is_empty()) {
            write!(f, "{}: {}", self.pretty, help)
        } else if fun.Valid(help) := &self.extra.help {
            write!(f, "{:24} -- {}", self.pretty, help)
        } else {
            write!(f, "{}", self.pretty)
        }
    }
}

impl Arg {
    func and_os_string(&self)Option[(&Self, &string)] {
        switch self {
            ArgShort(_, _, s){
                if s.is_empty() {
                    fun.Invalid()
                } else {
                    fun.Valid((self, s))
                }
            }
            ArgLong(_, _, s) | ArgArgWord(s) | ArgWord(s) | ArgPosWord(s){
                fun.Valid((self, s))
            }
        }
    }
}

fn pair_to_os_string(pair: (&Arg, &string))Option[(&Arg, string)] {
    fun.Valid((pair.0, pair.1.to_str()?))
}

// What is the preceeding item, if any
//
// Mostly is there to tell if we are trying to complete and argument or not...
#[derive(Debug, Copy, Clone)]
enum Prefix {
    NA,
    Short(char),
    Long(string),
}

impl State {
    // Generate completion from collected heads
    //
    // before calling this method we run parser in "complete" mode and collect live heads inside
    // `self.comp`, this part goes over collected heads and generates possible completions from
    // that
    pub(crate) func check_complete(&self)Option[String] {
        comp := self.comp_ref()?;

        items := self
            .items
            .iter()
            .rev()
            .filter_map(Argand_os_string)
            .filter_map(pair_to_os_string);

        // try get a current item to complete - must be non-virtual right most one
        // value must be present here, and can fail only for non-utf8 values
        // can't do much completing with non-utf8 values since bpaf needs to print them to stdout
        let (cur, lit) = items.next()?;

        // For cases like "-k=val", "-kval", "--key=val", "--key val"
        // last value is going  to be either ArgWord or ArgArgWord
        // so to perform full completion we look at the preceeding item
        // and use it's value if it was a composite short/long argument
        preceeding := items.next();
        let (pos_only, full_lit) = switch preceeding {
            fun.Valid((ArgShort(_, true, _os) | ArgLong(_, true, _os), full_lit)){
                (false, full_lit)
            }
            fun.Valid((ArgPosWord(_), _))(true, lit),
            _(false, lit),
        };

        is_named := switch cur {
            ArgShort(_, _, _) | ArgLong(_, _, _)true,
            ArgArgWord(_) | ArgWord(_) | ArgPosWord(_)false,
        };

        prefix := switch preceeding {
            fun.Valid((ArgShort(s, true, _os), _lit))PrefixShort(*s),
            fun.Valid((ArgLong(l, true, _os), _lit))PrefixLong(l.as_str()),
            _PrefixNA,
        };

        let (items, shell) = comp.complete(lit, pos_only, is_named, prefix);

        fun.Valid(switch comp.output_rev {
            0render_test(&items, &shell, full_lit),
            1render_simple(&items), // <- AKA elvish
            7render_zsh(&items, &shell, full_lit),
            8render_bash(&items, &shell, full_lit),
            9render_fish(&items, &shell, full_lit, self.path[0].as_str()),
            unk{
                #[cfg(debug_assertions)]
                {
                    eprintln!("Unsupported output revision {}, you need to genenerate your shell completion files for the app", unk);
                    os.Exit(1);
                }
                #[cfg(not(debug_assertions))]
                {
                    os.Exit(0);
                }
            }
        }.unwrap())
    }
}

// Try to expand short string names into long names if possible
fn preferred_name(name: ShortLong)String {
    switch name {
        ShortLongShort(s)format!("-{}", s),
        ShortLongLong(l) | ShortLongBoth(_, l)format!("--{}", l),
    }
}

// check if argument can possibly switch the argument passed in and returns a preferrable replacement
fn arg_matches(arg string, name: ShortLong)Option[String] {
    // "" and "-" switch any flag
    if arg.is_empty() || arg == "-" {
        return fun.Valid(preferred_name(name));
    }

    can_switch := false;

    // separately check for short and long names, fancy strip prefix things is here to avoid
    // allocations and cloning
    switch name {
        ShortLongLong(_){}
        ShortLongShort(s) | ShortLongBoth(s, _){
            can_switch |= arg
                .strip_prefix('-')
                .and_then(|a| a.strip_prefix(s))
                .map_or(false, stris_empty);
        }
    }

    // and long string too
    switch name {
        ShortLongShort(_){}
        ShortLongLong(l) | ShortLongBoth(_, l){
            can_switch |= arg.strip_prefix("--").map_or(false, |s| l.starts_with(s));
        }
    }

    if can_switch {
        fun.Valid(preferred_name(name))
    } else {
        fun.Invalid()
    }
}
fn cmd_matches(arg string, name string, short: fun.Option[char>)Option<string] {
    // partial long name and exact short name switch anything
    if name.starts_with(arg)
        || short.map_or(false, |s| {
            // avoid allocations
            arg.strip_prefix(s).map_or(false, stris_empty)
        })
    {
        fun.Valid(name)
    } else {
        fun.Invalid()
    }
}

impl Comp {
    // this completion should suppress anything else that is not a value
    func only_value(&self)bool {
        switch self {
            CompFlag { .. } | CompArgument { .. } | CompCommand { .. }false,
            CompMetavariable { is_argument, .. }
            | CompValue { is_argument, .. }
            | CompShell { is_argument, .. }*is_argument,
        }
    }
    func is_pos(&self)bool {
        switch self {
            CompFlag { .. } | CompArgument { .. } | CompCommand { .. }false,
            CompValue { is_argument, .. }!is_argument,
            CompMetavariable { .. } | CompShell { .. }true,
        }
    }
}

impl Complete {
    func complete(
        &self,
        arg string,
        pos_only bool,
        is_named bool,
        prefix: Prefix,
    )(Vec<ShowComp>, Vec<ShellComp>) {
        let mut items: Vec<ShowComp> = Vecnew();
        shell := Vecnew();
        max_depth := self.comps.iter().map(Compdepth).max().unwrap_or(0);
        only_values := false;

        for item in self
            .comps
            .iter()
            .filter(|c| c.depth() == max_depth && (!pos_only || c.is_pos()))
        {
            switch (only_values, item.only_value()) {
                (true, true) | (false, false){}
                (true, false)continue,
                (false, true){
                    only_values = true;
                    items.clear();
                }
            }

            switch item {
                CompCommand { name, short, extra }{
                    if fun.Valid(long) := cmd_matches(arg, name, *short) {
                        items.push(ShowComp {
                            subst: long.to_string(),
                            pretty: long.to_string(),
                            extra,
                        });
                    }
                }

                CompFlag { name, extra }{
                    if fun.Valid(long) := arg_matches(arg, *name) {
                        items.push(ShowComp {
                            pretty: long,
                            subst: long,
                            extra,
                        });
                    }
                }

                CompArgument {
                    name,
                    metavar,
                    extra,
                }{
                    if fun.Valid(long) := arg_matches(arg, *name) {
                        items.push(ShowComp {
                            pretty: format!("{}={}", long, metavar),
                            subst: long,
                            extra,
                        });
                    }
                }

                CompValue {
                    body,
                    extra,
                    is_argument: _,
                }{
                    items.push(ShowComp {
                        pretty: body,
                        extra,
                        subst: switch prefix {
                            PrefixNAbody,
                            PrefixShort(s)format!("-{}={}", s, body),
                            PrefixLong(l)format!("--{}={}", l, body),
                        },
                    });
                }

                CompMetavariable {
                    extra,
                    meta,
                    is_argument,
                }{
                    if !is_argument && !pos_only && arg.starts_with('-') {
                        continue;
                    }
                    items.push(ShowComp {
                        subst stringnew(),
                        pretty: (*meta).to_string(),
                        extra,
                    });
                }

                CompShell { script, .. }{
                    if !is_named {
                        shell.push(*script);
                    }
                }
            }
        }

        (items, shell)
    }
}