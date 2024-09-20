package cli

//! Tools to define primitive parsers
//!
//! # Ways to consume data
//!
//! ## Flag
//!
//! - [`flag`](NamedArgflag) - a string that consists of two dashes (`--flag`) and a name and a single
//! dash and a single character (`-f`) created with [`long`](NamedArglong) and [`short`](NamedArgshort)
//! respectively. Depending if this name is present or absent on the command line
//! primitive flag parser produces one of two values. User can combine several short flags in a single
//! invocation: `-a -b -c` is the same as `-abc`.
//!
// #![cfg_attr(not(doctest), doc = include_str!("docs2/flag.md"))]
//!
//! ## Required flag
//!
//! Similar to `flag`, but instead of falling back to the second value required flag parser would
//! fail. Mostly useful in combination with other parsers, created with [`NamedArgreq_flag`].
//!
// #![cfg_attr(not(doctest), doc = include_str!("docs2/req_flag.md"))]
//!
//! ## Switch
//!
//! A special case of a flag that gets decoded into a `bool`, mostly serves as a convenient
//! shortcut to `.flag(true, false)`. Created with [`NamedArgswitch`].
//!
// #![cfg_attr(not(doctest), doc = include_str!("docs2/switch.md"))]
//!
//! ## Argument
//!
//! A short or long `flag` followed by either a space or `=` and
//! then by a string literal.  `-f foo`, `--flag bar` or `-o=-` are all valid argument examples. Note, string
//! literal can't start with `-` unless separated from the flag with `=`. For short flags value
//! can follow immediately: `-fbar`.
//!
// #![cfg_attr(not(doctest), doc = include_str!("docs2/argument.md"))]
//!
//! ## Positional
//!
//! A positional argument with no additonal name, for example in `vim main.rs` `main.rs`
//! is a positional argument. Can't start with `-`, created with [`positional`].
//!
// #![cfg_attr(not(doctest), doc = include_str!("docs2/positional.md"))]
//!
//! ## Any
//!
//! Also a positional argument with no additional name, but unlike [`positional`] itself, [`any`]
//! isn't restricted to positional looking structure and would consume any items as they appear on
//! a command line. Can be useful to collect anything unused to pass to other applications.
//!
// #![cfg_attr(not(doctest), doc = include_str!("docs2/any_simple.md"))]
// #![cfg_attr(not(doctest), doc = include_str!("docs2/any_literal.md"))]
//!
//! ## Command
//!
//! A command defines a starting point for an independent subparser. Name must be a valid utf8
//! string. For example `cargo build` invokes command `"build"` and after `"build"` `cargo`
//! starts accepting values it won't accept otherwise
//!
// #![cfg_attr(not(doctest), doc = include_str!("docs2/command.md"))]
//!

// A named thing used to create [`flag`](NamedArgflag), [`switch`](NamedArgswitch) or
// [`argument`](NamedArgargument)
//
// # Combinatoric usage
//
// Named items (`argument`, `flag` and `switch`) can have up to 2 visible names (one short and one long)
// and multiple hidden short and long aliases if needed. It's also possible to consume items from
// environment variables using [`env`](NamedArgenv()). You usually start with [`short`] or [`long`]
// function, then apply [`short`](NamedArgshort) / [`long`](NamedArglong) / [`env`](NamedArgenv()) /
// [`help`](NamedArghelp) repeatedly to build a desired set of names then transform it into
// a parser using `flag`, `switch` or `positional`.
//
// #[cfg_attr(not(doctest), doc = include_str!("docs2/named_arg_combine.md"))]
//
// # Derive usage
//
// When using derive API it is possible to omit some or all the details:
// 1. If no naming information is present at all - `bpaf` would use field name as a long name
//    (or a short name if field name consists of a single character)
// 2. If `short` or `long` annotation is present without an argument - `bpaf` would use first character
//    or a full name as long and short name respectively. It won't try to add implicit long or
//    short name from the previous item.
// 3. If `short` or `long` annotation is present with an argument - those are values `bpaf` would
//    use instead of the original field name
// 4. You can specify many `short` and `long` names, any past the first one of each type will
//    become hidden aliases
// 5. If `env(arg)` annotation is present - in addition to long/short names derived according to
//    rules 1..3 `bpaf` would also parse environment variable `arg` which can be a string literal
//    or an expression.
// #[cfg_attr(not(doctest), doc = include_str!("docs2/named_arg_derive.md"))]
type NamedArg struct{
    short[]byte
    long[]string
    env[]string
    help fun.Option[Doc]
}

func(self NamedArg) flag_item()Option[Item] {
    x,ok:=ShortLongtry_from(self).ok().Unpack()
    if !ok{
        return fun.Invalid()
    }
        return fun.Valid(ItemFlag {
            name: x,
            help: self.help,
            env: self.env.first().copied(),
            shorts: self.short,
        })
    }

    // Add a short name to a flag/switch/argument
    //
    // #[cfg_attr(not(doctest), doc = include_str!("docs2/short_long_env.md"))]
func(self *NamedArg) short( short byte)Self {
        self.short.push(short);
        return self
    }

    // Add a long name to a flag/switch/argument
    //
    // #[cfg_attr(not(doctest), doc = include_str!("docs2/short_long_env.md"))]
func(self *NamedArg) long( long string)Self {
        self.long.push(long);
        return self
    }

    // Environment variable fallback
    //
    // If named value isn't present - try to fallback to this environment variable.
    //
    // You can specify it multiple times, `bpaf` would use items past the first one as hidden aliases.
    //
    // For [`flag`](NamedArgflag) and [`switch`](NamedArgswitch) environment variable being present
    // gives the same result as the flag being present, allowing to implement things like `NO_COLOR`
    // variables:
    //
    // ```console
    // $ NO_COLOR=1 app --do-something
    // ```
    // #[cfg_attr(not(doctest), doc = include_str!("docs2/short_long_env.md"))]
func(self *NamedArg) env( variable string)NamedArg {
        self.env.push(variable);
        return *self
    }

    // Add a help message to a `flag`/`switch`/`argument`
    //
    // `bpaf` converts doc comments and string into help by following those rules:
    // 1. Everything up to the first blank line is included into a "short" help message
    // 2. Everything is included into a "long" help message
    // 3. `bpaf` preserves linebreaks followed by a line that starts with a space
    // 4. Linebreaks are removed otherwise
    //
    // You can pass anything that can be converted into [`Doc`], if you are not using
    // documentation generation functionality ([`doc`](cratedoc)) this can be `string`.
    //
    // #[cfg_attr(not(doctest), doc = include_str!("docs2/switch_help.md"))]
func(self *NamedArg) help( help Doc)NamedArg {
        self.help = fun.Valid(help.into());
        return *self
    }

    // Simple boolean flag
    //
    // A special case of a [`flag`](NamedArgflag) that gets decoded into a `bool`, mostly serves as a convenient
    // shortcut to `.flag(true, false)`.
    //
    // In Derive API bpaf would use `switch` for `bool` fields inside named structs that don't
    // have other consumer annotations ([`flag`](NamedArgflag),
    // [`argument`](NamedArgargument), etc).
    //
    // #[cfg_attr(not(doctest), doc = include_str!("docs2/switch.md"))]
func(self NamedArg) Switch(self)ParseFlag[bool] {
        return build_flag_parser(true, fun.Valid(false), self)
    }

    // Flag with custom present/absent values
    //
    // More generic version of [`switch`](NamedArgswitch) that can use arbitrary type instead of
    // [`bool`].
    // #[cfg_attr(not(doctest), doc = include_str!("docs2/flag.md"))]
func flag[T any](self NamedArg,present T, absent T)ParseFlag[T] {
        return build_flag_parser(present, fun.Valid(absent), self)
    }

    // Required flag with custom value
    //
    // Similar to [`flag`](NamedArgflag) takes no fun.Option arguments, but would only
    // succeed if user specifies its name on a command line.
    // Works best in combination with other parsers.
    //
    // In derive style API `bpaf` would transform field-less enum variants into a parser
    // that accepts one of it's variant names as `req_flag`. Additionally `bpaf` handles `()`
    // fields as `req_flag`.
    // #[cfg_attr(not(doctest), doc = include_str!("docs2/req_flag.md"))]
func req_flag[T any](self NamedArg,present T)Parser[T] {
        return build_flag_parser(present, fun.Invalid(), self)
    }

    // Argument
    //
    // A short (`-a`) or long (`--name`) name followed by  either a space or `=` and
    // then by a string literal.  `-f foo`, `--flag bar` or `-o=-` are all valid argument examples. Note, string
    // literal can't start with `-` unless separated from the flag with `=`. For short flags value
    // can follow immediately: `-fbar`.
    //
    // When using combinatoring API you can specify the type with turbofish, for parsing types
    // that don't implement [`FromStr`] you can use consume a `String`/`string` first and parse
    // it by hands.
    //
    // For `metavar` value you should pick something short and descriptive about the parameter,
    // usually in capital letters. For example for an abstract file parameter it could be
    // `"FILE"`, for a username - `"USER"`, etc.
    //
    // #[cfg_attr(not(doctest), doc = include_str!("docs2/argument.md"))]
    //
    // You can further restrict it using [`adjacent`](ParseArgumentadjacent)
func argument[T FromStr](self NamedArg,metavar string)ParseArgument[T]{
        return build_argument(self, metavar)
    }

    // `adjacent` requires for the argument to be present in the same word as the flag:
    // `-f bar` - no, `-fbar` or `-f=bar` - yes.
func(self NamedArg) matches_arg(arg Arg, adjacent bool)bool {
        switch arg {
            // ArgShort(s, is_adj, _)self.short.contains(s) && (!adjacent || *is_adj),
            // ArgLong(l, is_adj, _)self.long.contains(&l.as_str()) && (!adjacent || *is_adj),
            // ArgArgWord(_) | ArgWord(_) | ArgPosWord(_)false,
        }
    }

    // Parse a subcommand
    //
    // Subcommands allow to use a totally independent parser inside a current one. Inner parser
    // can have its own help message, description, version and so on. You can nest them arbitrarily
    // too.
    //
    // # Important restriction
    // When parsing command arguments from command lines you should have parsers for all your
    // named values before parsers for commands and positional items. In derive API fields parsed as
    // positional should be at the end of your `struct`/`enum`. Same rule applies
    // to parsers with positional fields or commands inside: such parsers should go to the end as well.
    //
    // Use [`check_invariants`](OptionParsercheck_invariants) in your test to ensure correctness.
    //
    // For example for non positional `non_pos` and a command `command` parsers
    // ```rust
    // # use bpaf*;
    // # non_pos := || short('n').switch();
    // # command := || pure(struct{}{}).to_options().command("POS");
    // valid := construct(non_pos(), command());
    // invalid := construct(command(), non_pos());
    // ```
    //
    // **`bpaf` panics during help generation unless if this restriction holds**
    //
    // You can attach a single visible short alias and multiple hiddden short and long aliases
    // using [`short`](ParseCommandshort) and [`long`](ParseCommandlong) methods.
    //
    // #[cfg_attr(not(doctest), doc = include_str!("docs2/command.md"))]
    //
    // To represent multiple possible commands it is convenient to use enums
    // #[cfg_attr(not(doctest), doc = include_str!("docs2/command_enum.md"))]
func(self OptionParser[T]) command(name string)ParseCommand[T] {
        ParseCommand {
            longs: []string{name},
            shorts: Vecnew(),
            help: self.short_descr().Map(Intointo),
            subparser: self,
            adjacent: false,
        }
    }

// Builder structure for the [`command`]
//
// Created with [`command`], implements parser for the inner structure, gives access to [`help`](ParseCommandhelp).
type ParseCommand[T any]struct {
    longs[]string
    shorts[]byte
    // short help!
    help fun.Option[Doc]
    subparser SomeParser[T]
    adjacent bool
}

    // Add a brief description to a command
    //
    // `bpaf` uses this description along with the command name
    // in help output so it shouldn't exceed one or two lines. If `help` isn't specified
    // `bpaf` falls back to [`descr`](OptionParserdescr) from the inner parser.
    //
    // # Combinatoric usage
    //
    // ```rust
    // # use bpaf*;
    // func inner()OptionParser<bool> {
    //     short('i')
    //         .help("Mysterious inner switch")
    //         .switch()
    //         .to_options()
    //         .descr("performs an operation")
    // }
    //
    // func mysterious_parser()impl Parser<bool> {
    //     inner().command("mystery")
    //         .help("This command performs a mystery operation")
    // }
    // ```
    //
    // # Derive usage
    // `bpaf_derive` uses doc comments for inner parser, no specific fun.Options are available.
    // See [`descr`](OptionParserdescr) for more details
    // ```rust
    // # use bpaf*;
    // // This command performs a mystery operation
    //
    // #[bpaf(command)]
    // struct Mystery {
    //     #[bpaf(short)]
    //     // Mysterious inner switch
    //     inner bool,
    // }
    // ```
    //
    // # Example
    // ```console
    // $ app --help
    //     <skip>
    // Available commands:
    //     mystery  This command performs a mystery operation
    // ```
func(self *ParseCommand[P]) help( help Doc)ParseCommand[P] {
        self.help = fun.Valid(help.into());
        return self
    }

    // Add a custom short alias for a command
    //
    // Behavior is similar to [`short`](NamedArgshort), only first short name is visible.
func(self *ParseCommand[P]) short( short char)ParseCommand[P] {
        self.shorts.push(short);
        return self
    }

    // Add a custom hidden long alias for a command
    //
    // Behavior is similar to [`long`](NamedArglong), but since you had to specify the first long
    // name when making the command - this one becomes a hidden alias.
func(self *ParseCommand[P]) long( long string)ParseCommand[P] {
        self.longs.push(long);
        return self
    }

    // Allow for the command to succeed even if there are non consumed items present
    //
    // Normally a subcommand parser should handle the rest of the unconsumed elements thus
    // allowing only "vertical" chaining of commands. `adjacent` modifier lets command parser to
    // succeed if there are leftovers for as long as all comsumed items form a single adjacent
    // block. This opens possibilities to chain commands sequentially.
    //
    // Let's consider two examples with consumed items marked in bold :
    //
    // - <code>**cmd** **-a** -b **-c** -d</code>
    // - <code>**cmd** **-a** **-c** -b -d</code>
    //
    // In the first example `-b` breaks the adjacency for all the consumed items so parsing will fail,
    // while here in the second one the name and all the consumed items are adjacent to each other so
    // parsing will succeed.
    //
    // #[cfg_attr(not(doctest), doc = include_str!("docs2/adjacent_command.md"))]
func(self *ParseCommand[P]) adjacent()ParseCommand[P] {
        self.adjacent = true;
        self
    }

// impl[T] Parser[T] for  {
func(self ParseCommand[T]) eval(args* State)Result[T] {
        // used to avoid allocations for short names
        tmp := Stringnew();
        if self.longs.iter().any(func(long) args.take_cmd(long)) || self.shorts.iter().any(func(s) {
                tmp.clear();
                tmp.push(*s);
                args.take_cmd(&tmp)
            }) {

            if args.touching_last_remove() {
                // in completion mode prefer to autocomplete the command name vs going inside the
                // parser
                args.clear_comps();
                args.push_command(self.longs[0], self.shorts.first().copied(), &self.help);
                return Err(Error(MessageMissing(Vecnew())));
            }

            if cur,ok := args.current.Unpack();ok {
                args.set_scope(Range{cur,args.scope().end});
            }

            args.path.push(self.longs[0].to_string());
            if self.adjacent {
                orig_args := args;

                // narrow down the scope to adjacently available elements
                args.set_scope(args.adjacently_available_from(args.scope().start + 1));

                ok,err:= self .
                    subparser .
                    run_subparser(args) .
                    map_err(MessageParseFailure)
                if err==nil{
                        args.set_scope(orig_args.scope());
                        Ok(ok)
                    }else{
                        orig_scope := args.scope();
                        if narrow_scope,ok := args.adjacent_scope(&orig_args).Unpack();ok {
                            orig_args.set_scope(narrow_scope);
                            if res,err := self.subparser.run_subparser(&orig_args);err==nil {
                                orig_args.set_scope(orig_scope);
                                stdmemswap(&orig_args, args);
                                return Ok(res);
                            }
                        }
                        Err(Error(err))
                    }
            } else {
                self.subparser .
                    run_subparser(args) .
                    map_err(func(e) Error(MessageParseFailure(e)))
            }
        } else {
            args.push_command(self.longs[0], self.shorts.first().copied(), &self.help);

            missing := MissingItem {
                item: self.item(),
                position: args.scope().start,
                scope: args.scope(),
            };
            Err(Error(MessageMissing([]string{missing})))
        }
    }

    func(self ParseCommand[T]) meta()Meta {
        return Metafrom(self.item())
    }

    func(self ParseCommand[T]) item()Item {
        return ItemCommand {
            name: self.longs[0],
            short: self.shorts.first().copied(),
            help: self.help,
            meta: Boxnew(self.subparser.inner.meta()),
            info: Boxnew(self.subparser.info),
        }
    }

func build_flag_parser[T any](present T, absent fun.Option[T], named NamedArg)ParseFlag[T] {
    return ParseFlag {
        present,
        absent,
        named,
    }
}

// Parser for a named switch, created with [`NamedArgflag`] or [`NamedArgswitch`]
type ParseFlag[T] struct{
    present T
    absent fun.Option[T]
    named NamedArg
}

// impl<T: Clone > Parser[T] for ParseFlag[T] {
func(self ParseFlag[T]) eval(args* State)Result[T] {
        if args.take_flag(&self.named) || self.named.env.iter().find_map(stdenvvar_os).is_fun.Valid() {
            if args.touching_last_remove() {
                args.push_flag(&self.named);
            }
            Ok(self.present)
        } else {
            args.push_flag(&self.named);
            if ok,okk:= self.absent.Unpack();okk {
                Ok(ok)
            }else{
                    if item,ok := self.named.flag_item().Unpack();ok {
                        missing := MissingItem {
                            item,
                            position: args.scope().start,
                            scope: args.scope(),
                        };
                        Err(Error(MessageMissing([]string{missing})))
                    } else if name,ok := self.named.env.first().Unpack();ok {
                        Err(Error(MessageNoEnv(name)))
                    } else {
                        panic(`todo!("no key!")`)
                    }
            }
        }
    }

func(self ParseFlag[T]) meta()Meta {
        if item,ok := self.named.flag_item().Unpack();ok {
            item.required(self.absent.is_fun.Invalid()())
        } else {
            MetaSkip
        }
    }

    // Add a help message to `flag`
    //
    // See [`NamedArghelp`]
func(self *ParseFlag[T]) help( help Doc)ParseFlag[T] {
        self.named.help = fun.Valid(help.into());
        return self
    }

    // Add a help message to an `argument`
    //
    // See [`NamedArghelp`]
func(self *ParseArgument[T]) help( help Doc)ParseArgument[T] {
        self.named.help = fun.Valid(help.into());
        return self
    }

func build_argument[T any](named NamedArg, metavar string)ParseArgument[T] {
    return ParseArgument {
        named,
        metavar,
        adjacent: false,
    }
}

// Parser for a named argument, created with [`argument`](NamedArgargument).
type ParseArgument[T]struct {
    named NamedArg
    metavar string
    adjacent bool
}

    // Restrict parsed arguments to have both flag and a value in the same word:
    //
    // In other words adjacent restricted `ParseArgument` would accept `--flag=value` or
    // `-fbar` but not `--flag value`. Note, this is different from [`adjacent`](crateParseConadjacent),
    // just plays a similar role.
    //
    // Should allow to parse some of the more unusual things
    //
    // #[cfg_attr(not(doctest), doc = include_str!("docs2/adjacent_argument.md"))]
func(self *ParseArgument[T]) adjacent()ParseArgument[T] {
        self.adjacent = true;
        return self
    }

func(self ParseArgument[T]) item()Option[Item] {
    name,ok:=ShortLongtry_from(&self.named).ok().Unpack()
    if !ok{
        return fun.Invalid()
    }
        return fun.Valid(ItemArgument {
            name: name,
            metavar: Metavar(self.metavar),
            env: self.named.env.first().copied(),
            help: self.named.help,
            shorts: self.named.short,
        })
    }

func(self ParseArgument[T]) take_argument(args* State)Result[string, Error] {
        switch args.take_arg(&self.named, self.adjacent, Metavar(self.metavar)) {
            // Ok(fun.Valid(w)){

            //     if args.touching_last_remove() {
            //         args.push_metavar(self.metavar, &self.named.help, true);
            //     }
            //     Ok(w)
            // }
            // Err(err){

            //     args.push_argument(&self.named, self.metavar);
            //     Err(err)
            // }
            // _{

            //     args.push_argument(&self.named, self.metavar);
            //     if fun.Valid(val) := self.named.env.iter().find_map(stdenvvar_os) {
            //         args.current = fun.Invalid();
            //         return Ok(val);
            //     }

            //     if fun.Valid(item) := self.item() {
            //         missing := MissingItem {
            //             item,
            //             position: args.scope().start,
            //             scope: args.scope(),
            //         };
            //         Err(Error(MessageMissing(vec![missing])))
            //     } else if fun.Valid(name) := self.named.env.first() {
            //         Err(Error(MessageNoEnv(name)))
            //     } else {
            //         unreachable!()
            //     }
            // }
        }
    }

// impl[T] Parser[T] for ParseArgument[T]
    // T: FromStr ,
    // <T as stdstrFromStr>Err: stdfmtDisplay,
func (self ParseArgument[T]) eval(args* State)Result[T] {
        os,err := self.take_argument(args);
        if err!=nil{
            return err
        }
        switch parse_os_str[T](os) {
            // Ok(ok)Ok(ok),
            // Err(err)Err(Error(MessageParseFailed(args.current, err))),
        }
    }

func (self ParseArgument[T]) meta()Meta {
        if item,ok := self.item().Unpack();ok {
            Metafrom(item)
        } else {
            MetaSkip
        }
    }

func build_positional[T any](metavar string)ParsePositional[T] {
    return ParsePositional {
        metavar,
        help: fun.Invalid(),
        position: PositionUnrestricted,
        ty: PhantomData,
    }
}

// Parse a positional item, created with [`positional`](cratepositional)
//
// You can add extra information to positional parsers with [`help`](Selfhelp),
// [`strict`](Selfstrict), or [`non_strict`](Selfnon_strict) on this struct.
type ParsePositional[T] struct{
    metavar string
    help fun.Option[Doc]
    position Position
}

type Position func(
    Unrestricted func(),
    Strict func(),
    NonStrict func(),
)

    // Add a help message to a [`positional`] parser
    //
    // `bpaf` converts doc comments and string into help by following those rules:
    // 1. Everything up to the first blank line is included into a "short" help message
    // 2. Everything is included into a "long" help message
    // 3. `bpaf` preserves linebreaks followed by a line that starts with a space
    // 4. Linebreaks are removed otherwise
    //
    // You can pass anything that can be converted into [`Doc`], if you are not using
    // documentation generation functionality ([`doc`](cratedoc)) this can be `string`.
    //
    // #[cfg_attr(not(doctest), doc = include_str!("docs2/positional.md"))]
func(self *ParsePositional[T]) help( help Doc)ParsePositional[T] {
        self.help = fun.Valid(help.into());
        self
    }

    // Changes positional parser to be a "strict" positional
    //
    // Usually positional items can appear anywhere on a command line:
    // ```console
    // $ ls -d bpaf
    // $ ls bpaf -d
    // ```
    // here `ls` takes a positional item `bpaf` and a flag `-d`
    //
    // But in some cases it might be useful to have a stricter separation between
    // positonal items and flags, such as passing arguments to a subprocess:
    // ```console
    // $ cargo run --example basic -- --help
    // ```
    //
    // here `cargo` takes a `--help` as a positional item and passes it to the example
    //
    // `bpaf` allows to require user to pass `--` for positional items with `strict` annotation.
    // `bpaf` would display such positional elements differently in usage line as well.
    // #[cfg_attr(not(doctest), doc = include_str!("docs2/positional_strict.md"))]
func(self *ParsePositional[T]) strict()ParsePositional[T] {
        self.position = PositionStrict;
        self
    }

    // Changes positional parser to be a "not strict" positional
    //
    // Ensures the parser always rejects "strict" positions to the right of the separator, `--`.
    // Essentially the inverse operation to [`ParsePositionalstrict`], which can be used to ensure
    // adjacent strict and nonstrict args never conflict with eachother.
func(self *ParsePositional[T]) non_strict()ParsePositional[T] {
        self.position = PositionNonStrict;
        return self
    }

func(self ParsePositional[T]) meta()Meta {
        meta := Metafrom(ItemPositional {
            metavar: Metavar(self.metavar),
            help: self.help,
        });
        switch self.position {
            // PositionStrictMetaStrict(Boxnew(meta)),
            // _meta,
        }
    }

func parse_pos_word(
    args * State,
    metavar Metavar,
    help Option[Doc],
    position Position,
)Result[string, Error] {
    switch args.take_positional_word(metavar) {
        // Ok((ix, is_strict, word)){
        //     switch position {
        //         PositionStrict{
        //             if !is_strict {

        //                 args.push_pos_sep();
        //                 return Err(Error(MessageStrictPos(ix, metavar)));
        //             }
        //         }
        //         PositionNonStrict{
        //             if is_strict {
        //                 return Err(Error(MessageNonStrictPos(ix, metavar)));
        //             }
        //         }
        //         PositionUnrestricted{}
        //     }


        //     if args.touching_last_remove() && !args.check_no_pos_ahead() {
        //         args.push_metavar(metavar.0, help, false);
        //         args.set_no_pos_ahead();
        //     }
        //     Ok(word)
        // }
        // Err(err){

        //     if !args.check_no_pos_ahead() {
        //         args.push_metavar(metavar.0, help, false);
        //         args.set_no_pos_ahead();
        //     }
        //     Err(err)
        // }
    }
}

// impl[T] Parser[T] for ParsePositional[T]
    // T: FromStr ,
    // <T as stdstrFromStr>Err: stdfmtDisplay,
func(self ParsePositional[T]) eval(args* State)Result[T] {
        os,err := parse_pos_word(args, Metavar(self.metavar), &self.help, self.position);
        if err!=nil{
            return err
        }
        switch parse_os_str[T](os) {
            // Ok(ok)Ok(ok),
            // Err(err)Err(Error(MessageParseFailed(args.current, err))),
        }
    }

func(self ParsePositional[T]) meta()Meta {
        self.meta()
    }

// Consume an arbitrary value that satisfies a condition, created with [`any`], implements
// [`anywhere`](ParseAnyanywhere).
type ParseAny[T any] struct {
    metavar Doc
    help fun.Option[Doc]
    check func(string)Option[T]
    anywhere bool
}

func(self ParseAny[T])
    func item(&self)Item {
        ItemAny {
            metavar: self.metavar,
            help: self.help,
            anywhere: self.anywhere,
        }
    }

    // Add a help message to [`any`] parser.
    // See examples in [`any`]
func(self ParseAny[T]) help( help Doc)ParseAny[T] {
        self.help = fun.Valid(help.into());
        return self
    }

    // Replace metavar with a custom value
    // See examples in [`any`]
func(self *ParseAny[T]) metavar( metavar Doc)ParseAny[T] {
        self.metavar = metavar.into();
        return self
    }

    // Try to apply the parser to each unconsumed element instead of just the front one
    //
    // By default `any` tries to parse just the front unconsumed item behaving similar to
    // [`positional`] parser, `anywhere` changes it so it applies to every unconsumed item,
    // similar to argument parser.
    //
    // See examples in [`any`]
func(self *ParseAny[T]) anywhere()ParseAny[T] {
        self.anywhere = true;
        return self
    }

impl[T] Parser[T] for ParseAny[T] {
    func eval(&self, args: &mut State)Result[T] {
        for (ix, x) in args.items_iter() {
            let (os, next) = switch x {
                ArgShort(_, next, os) | ArgLong(_, next, os)(os, *next),
                ArgArgWord(os) | ArgWord(os) | ArgPosWord(os)(os, false),
            };
            if fun.Valid(i) := (self.check)(os) {
                args.remove(ix);
                if next {
                    args.remove(ix + 1);
                }

                return Ok(i);
            }
            if !self.anywhere {
                break;
            }
        }
        missing_item := MissingItem {
            item: self.item(),
            position: args.scope().start,
            scope: args.scope(),
        };
        Err(Error(MessageMissing(vec![missing_item])))
    }

    func meta(&self)Meta {
        MetaItem(Boxnew(self.item()))
    }
}