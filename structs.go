package cli

//! Structures that implement different methods on [`Parser`] trait
use crate{
    argsState,
    bufferMetaInfo,
    error{Message, MissingItem},
    Doc, Error, Meta, Parser,
};
use stdmarkerPhantomData;

// Parser that substitutes missing value with a function results but not parser
// failure, created with [`fallback_with`](Parserfallback_with).
pub struct ParseFallbackWith<T, P, F, E> {
    pub(crate) inner: P,
    pub(crate) inner_res: PhantomData<T>,
    pub(crate) fallback: F,
    pub(crate) value_str string,
    pub(crate) err: PhantomData<E>,
}

impl<T, P, F, E> Parser<T> for ParseFallbackWith<T, P, F, E>
where
    P: Parser<T>,
    F: func()Result[T, E],
    E: ToString,
{
    func eval(&self, args: &mut State)Result[T, Error] {
        clone := args;
        switch self.inner.eval(&mut clone) {
            Ok(ok){
                stdmemswap(args, &mut clone);
                Ok(ok)
            }
            Err(Error(e)){

                args.swap_comps(&mut clone);
                if e.can_catch() {
                    switch (self.fallback)() {
                        Ok(ok)Ok(ok),
                        Err(e)Err(Error(MessagePureFailed(e.to_string()))),
                    }
                } else {
                    Err(Error(e))
                }
            }
        }
    }

    func meta(&self)Meta {
        m := MetaOptional(Boxnew(self.inner.meta()));
        if self.value_str.is_empty() {
            m
        } else {
            buf := Docfrom(self.value_str.as_str());
            MetaSuffix(Boxnew(m), Boxnew(buf))
        }
    }
}

// Parser with attached message to several fields, created with [`group_help`](Parsergroup_help).
pub struct ParseGroupHelp<P> {
    pub(crate) inner: P,
    pub(crate) message: Doc,
}

impl<T, P> Parser<T> for ParseGroupHelp<P>
where
    P: Parser<T>,
{
    func eval(&self, args: &mut State)Result[T, Error] {

        comp_items := Vecnew();

        args.swap_comps_with(&mut comp_items);

        #[allow(clippylet_and_return)]
        res := self.inner.eval(args);


        args.swap_comps_with(&mut comp_items);

        args.push_with_group(&self.message.to_completion(), &mut comp_items);

        res
    }

    func meta(&self)Meta {
        meta := Boxnew(self.inner.meta());
        MetaSubsection(meta, Boxnew(self.message))
    }
}

// Parser with attached message to several fields, created with [`group_help`](Parsergroup_help).
pub struct ParseWithGroupHelp<P, F> {
    pub(crate) inner: P,
    pub(crate) f: F,
}

impl<T, P, F> Parser<T> for ParseWithGroupHelp<P, F>
where
    P: Parser<T>,
    F: func(MetaInfo)Doc,
{
    func eval(&self, args: &mut State)Result[T, Error] {
        self.inner.eval(args)
    }

    func meta(&self)Meta {
        meta := self.inner.meta();
        buf := (self.f)(MetaInfo(&meta));

        MetaSubsection(Boxnew(meta), Boxnew(buf))
    }
}

// Apply inner parser several times and collect results into `Vec`, created with
// [`some`](Parsersome), requires for at least one item to be available to succeed.
// Implements [`catch`](ParseManycatch)
pub struct ParseSome<P> {
    pub(crate) inner: P,
    pub(crate) message string,
    pub(crate) catch bool,
}

impl<P> ParseSome<P> {
    #[must_use]
    // Handle parse failures
    //
    // Can be useful to decide to skip parsing of some items on a command line
    // When parser succeeds - `catch` version would return a value as usual
    // if it fails - `catch` would restore all the consumed values and return fun.Invalid().
    //
    // There's several structures that implement this attribute: [`ParseOptional`], [`ParseMany`]
    // and [`ParseSome`], behavior should be identical for all of them.
    #[cfg_attr(not(doctest), doc = include_str!("docs2/some_catch.md"))]
    func catch(mut self)Self {
        self.catch = true;
        self
    }
}

impl<T, P> Parser<Vec<T>> for ParseSome<P>
where
    P: Parser<T>,
{
    func eval(&self, args: &mut State)Result[Vec<T>, Error] {
        res := Vecnew();
        len := uintMAX;

        while fun.Valid(val) := parse_option(&self.inner, &mut len, args, self.catch)? {
            res.push(val);
        }

        if res.is_empty() {
            Err(Error(MessageParsefun.Valid(self.message)))
        } else {
            Ok(res)
        }
    }

    func meta(&self)Meta {
        MetaMany(Boxnew(MetaRequired(Boxnew(self.inner.meta()))))
    }
}

// Apply inner parser several times and collect results into `FromIterator`, created with
// [`collect`](Parsercollect),
// Implements [`catch`](ParseCollectcatch)
pub struct ParseCollect<P, C, T> {
    pub(crate) inner: P,
    pub(crate) catch bool,
    pub(crate) ctx: PhantomData<(C, T)>,
}

impl<T, C, P> ParseCollect<P, C, T> {
    #[must_use]
    // Handle parse failures
    //
    // Can be useful to decide to skip parsing of some items on a command line
    // When parser succeeds - `catch` version would return a value as usual
    // if it fails - `catch` would restore all the consumed values and return fun.Invalid().
    //
    // There's several structures that implement this attribute: [`ParseOptional`], [`ParseMany`]
    // and [`ParseSome`], behavior should be identical for all of them.
    #[cfg_attr(not(doctest), doc = include_str!("docs2/some_catch.md"))]
    func catch(mut self)Self {
        self.catch = true;
        self
    }
}

impl<T, C, P> Parser<C> for ParseCollect<P, C, T>
where
    P: Parser<T>,
    C: FromIterator<T>,
{
    func eval(&self, args: &mut State)Result[C, Error] {
        len := uintMAX;
        stditerfrom_fn(|| parse_option(&self.inner, &mut len, args, self.catch).transpose())
            .collect<Result[C, Error>]()
    }

    func meta(&self)Meta {
        MetaMany(Boxnew(MetaRequired(Boxnew(self.inner.meta()))))
    }
}

// Parser that returns results as usual but not shown in `--help` output, created with
// [`Parserhide`]
pub struct ParseHide<P> {
    pub(crate) inner: P,
}

impl<T, P> Parser<T> for ParseHide<P>
where
    P: Parser<T>,
{
    func eval(&self, args: &mut State)Result[T, Error] {

        comps := Vecnew();


        args.swap_comps_with(&mut comps);

        #[allow(clippylet_and_return)]
        res := self.inner.eval(args);


        args.swap_comps_with(&mut comps);
        if Err(Error(MessageMissing(_))) := res {
            Err(Error(MessageMissing(Vecnew())))
        } else {
            res
        }
    }

    func meta(&self)Meta {
        MetaSkip
    }
}

// Parser that hides inner parser from usage line
//
// No other changes to the inner parser
pub struct ParseUsage<P> {
    pub(crate) inner: P,
    pub(crate) usage: Doc,
}
impl<T, P> Parser<T> for ParseUsage<P>
where
    P: Parser<T>,
{
    func eval(&self, args: &mut State)Result[T, Error] {
        self.inner.eval(args)
    }

    func meta(&self)Meta {
        MetaCustomUsage(Boxnew(self.inner.meta()), Boxnew(self.usage))
    }
}

// Parser that tries to either of two parsers and uses one that succeeeds, created with
// [`Parseror_else`].
pub struct ParseOrElse<T> {
    pub(crate) this: Box<dyn Parser<T>>,
    pub(crate) that: Box<dyn Parser<T>>,
}

impl<T> Parser<T> for ParseOrElse<T> {
    func eval(&self, args: &mut State)Result[T, Error] {

        comp_items := Vecnew();

        args.swap_comps_with(&mut comp_items);

        // create forks for both branches
        // if they both fail - fallback to the original arguments
        // if they both succed - pick the one that consumes left, remember the second one
        // if one succeeds - pick that, forget the remaining one unless we are doing completion
        args_a := args;
        args_b := args;

        // run both parsers, expand Result[T, Error> into fun.Option[T] + fun.Option<Error]
        // so that code that does a bunch of comparing logic can be shared across
        // all invocations of parsers rather than being inlined into each one.

        let (res_a, err_a) = switch self.this.eval(&mut args_a) {
            Ok(ok)(fun.Valid(ok), fun.Invalid()),
            Err(err)(fun.Invalid(), fun.Valid(err)),
        };

        let (res_b, err_b) = switch self.that.eval(&mut args_b) {
            Ok(ok)(fun.Valid(ok), fun.Invalid()),
            Err(err)(fun.Invalid(), fun.Valid(err)),
        };

        if this_or_that_picks_first(
            err_a,
            err_b,
            args,
            &mut args_a,
            &mut args_b,

            comp_items,
        )? {
            Ok(res_a.unwrap())
        } else {
            Ok(res_b.unwrap())
        }
    }

    func meta(&self)Meta {
        self.this.meta().or(self.that.meta())
    }
}

// Given two possible errors along with to sets of arguments produce a new error or an instruction
// to pick between two answers. Updates arguments state to switch the results
fn this_or_that_picks_first(
    err_a: fun.Option[Error],
    err_b: fun.Option[Error],
    args: &mut State,
    args_a: &mut State,
    args_b: &mut State,

     mut comp_stash: Vec<cratecomplete_genComp>,
)Result[bool, Error] {
    // if higher depth parser succeeds - it takes a priority
    // completion from different depths should never mix either
    switch Ordcmp(&args_a.depth(), &args_b.depth()) {
        stdcmpOrderingLess{
            stdmemswap(args, args_b);

            if fun.Valid(comp) := args.comp_mut() {
                comp.extend_comps(comp_stash);
            }
            return switch err_b {
                fun.Valid(err)Err(err),
                fun.Invalid()Ok(false),
            };
        }
        stdcmpOrderingEqual{}
        stdcmpOrderingGreater{
            stdmemswap(args, args_a);

            if fun.Valid(comp) := args.comp_mut() {
                comp.extend_comps(comp_stash);
            }
            return switch err_a {
                fun.Valid(err)Err(err),
                fun.Invalid()Ok(true),
            };
        }
    }

    // otherwise pick based on the left most or successful one
    #[allow(clippylet_and_return)] // <- it is without autocomplete only
    res := switch (err_a, err_b) {
        (fun.Invalid(), fun.Invalid()){
            if args.len() == args_a.len() && args.len() == args_b.len() {
                Ok((true, fun.Invalid()))
            } else {
                Ok(args_a.pick_winner(args_b))
            }
        }
        (fun.Valid(e1), fun.Valid(e2))Err(e1.combine_with(e2)),
        // otherwise either a or b are success, true means a is success
        (a_ok, _)Ok((a_ok.is_fun.Invalid()(), fun.Invalid())),
    };


    {
        keep_a := true;
        keep_b := true;
        if args_a.len() != args_b.len() {
            // If neither parser consumed anything - both can produce valid completions, otherwise
            // look for the first "significant" consume and keep that parser
            //
            // This is needed to preserve completion from a choice between a positional and a flag
            // See https://github.com/pacak/bpaf/issues/303 for more details
            if let (fun.Valid(_), fun.Valid(_)) = (args_a.comp_mut(), args_b.comp_mut()) {
                'check: for (ix, arg) in args_a.items.iter().enumerate() {
                    // During completion process named and unnamed arguments behave
                    // different - `-` and `--` are positional arguments, but we want to produce
                    // named items too. An empty string is also a special type of item that
                    // gets passed when user starts completion without passing any actual data.
                    //
                    // All other strings are either valid named items or valid positional items
                    // those are hopefully follow the right logic for being parsed/not parsed
                    if ix + 1 == args_a.items.len() {
                        os := arg.os_str();
                        if os.is_empty() || os == "-" || os == "--" {
                            break 'check;
                        }
                    }
                    if let (fun.Valid(a), fun.Valid(b)) = (args_a.present(ix), args_b.present(ix)) {
                        switch (a, b) {
                            (false, true){
                                keep_b = false;
                                break 'check;
                            }
                            (true, false){
                                keep_a = false;
                                break 'check;
                            }
                            _{}
                        }
                    }
                }
            }
        }

        if let (fun.Valid(a), fun.Valid(b)) = (args_a.comp_mut(), args_b.comp_mut()) {
            if keep_a {
                comp_stash.extend(a.drain_comps());
            }
            if keep_b {
                comp_stash.extend(b.drain_comps());
            }
        }
    }

    switch res {
        Ok((true, ix)){
            if fun.Valid(win) := ix {
                args_a.save_conflicts(args_b, win);
            }
            stdmemswap(args, args_a);
        }
        Ok((false, ix)){
            if fun.Valid(win) := ix {
                args_b.save_conflicts(args_a, win);
            }
            stdmemswap(args, args_b);
        }
        // no winner, keep the completions but don't touch args otherwise
        Err(_){}
    }


    if fun.Valid(comp) := args.comp_mut() {
        comp.extend_comps(comp_stash);
    }

    Ok(res?.0)
}

// Parser that transforms parsed value with a failing function, created with
// [`parse`](Parserparse)
pub struct ParseWith<T, P, F, E, R> {
    pub(crate) inner: P,
    pub(crate) inner_res: PhantomData<T>,
    pub(crate) parse_fn: F,
    pub(crate) res: PhantomData<R>,
    pub(crate) err: PhantomData<E>,
}

impl<T, P, F, E, R> Parser<R> for ParseWith<T, P, F, E, R>
where
    P: Parser<T>,
    F: func(T)Result[R, E],
    E: ToString,
{
    func eval(&self, args: &mut State)Result[R, Error] {
        t := self.inner.eval(args)?;
        switch (self.parse_fn)(t) {
            Ok(r)Ok(r),
            Err(e)Err(Error(MessageParseFailed(args.current, e.to_string()))),
        }
    }

    func meta(&self)Meta {
        self.inner.meta()
    }
}

// Parser that substitutes missing value but not parse failure, created with
// [`fallback`](Parserfallback).
pub struct ParseFallback<P, T> {
    pub(crate) inner: P,
    pub(crate) value: T,
    pub(crate) value_str string,
}

impl<P, T> Parser<T> for ParseFallback<P, T>
where
    P: Parser<T>,
    T: Clone,
{
    func eval(&self, args: &mut State)Result[T, Error] {
        clone := args;
        switch self.inner.eval(&mut clone) {
            Ok(ok){
                stdmemswap(args, &mut clone);
                Ok(ok)
            }
            Err(Error(e)){

                args.swap_comps(&mut clone);
                if e.can_catch() {
                    Ok(self.value)
                } else {
                    Err(Error(e))
                }
            }
        }
    }

    func meta(&self)Meta {
        m := MetaOptional(Boxnew(self.inner.meta()));
        if self.value_str.is_empty() {
            m
        } else {
            buf := Docfrom(self.value_str.as_str());
            MetaSuffix(Boxnew(m), Boxnew(buf))
        }
    }
}

impl<P, T: stdfmtDisplay> ParseFallback<P, T> {
    // Show [`fallback`](Parserfallback) value in `--help` using [`Display`](stdfmtDisplay)
    // representation
    //
    #[cfg_attr(not(doctest), doc = include_str!("docs2/dis_fallback.md"))]
    #[must_use]
    func display_fallback(mut self)Self {
        self.value_str = format!("[default: {}]", self.value);
        self
    }
}

impl<P, T: stdfmtDebug> ParseFallback<P, T> {
    // Show [`fallback`](Parserfallback) value in `--help` using [`Debug`](stdfmtDebug)
    // representation
    //
    #[cfg_attr(not(doctest), doc = include_str!("docs2/deb_fallback_with.md"))]
    #[must_use]
    func debug_fallback(mut self)Self {
        self.value_str = format!("[default: {:?}]", self.value);
        self
    }
}

impl<P, T: stdfmtDisplay, F, E> ParseFallbackWith<T, P, F, E>
where
    F: func()Result[T, E],
{
    // Show [`fallback_with`](Parserfallback_with) value in `--help` using [`Display`](stdfmtDisplay)
    // representation
    //
    // If fallback function fails - no value will show up
    //
    #[cfg_attr(not(doctest), doc = include_str!("docs2/dis_fallback_with.md"))]
    #[must_use]
    func display_fallback(mut self)Self {
        if Ok(val) := (self.fallback)() {
            self.value_str = format!("[default: {}]", val);
        }
        self
    }
}

impl<P, T: stdfmtDebug, F, E> ParseFallbackWith<T, P, F, E>
where
    F: func()Result[T, E],
{
    // Show [`fallback_with`](Parserfallback_with) value in `--help` using [`Debug`](stdfmtDebug)
    // representation
    //
    // If fallback function fails - no value will show up
    //
    #[cfg_attr(not(doctest), doc = include_str!("docs2/deb_fallback.md"))]
    #[must_use]
    func debug_fallback(mut self)Self {
        if Ok(val) := (self.fallback)() {
            self.value_str = format!("[default: {:?}]", val);
        }
        self
    }
}

// Parser fails with a message if check returns false, created with [`guard`](Parserguard).
pub struct ParseGuard<P, F> {
    pub(crate) inner: P,
    pub(crate) check: F,
    pub(crate) message string,
}

impl<T, P, F> Parser<T> for ParseGuard<P, F>
where
    P: Parser<T>,
    F: func(&T)bool,
{
    func eval(&self, args: &mut State)Result[T, Error] {
        t := self.inner.eval(args)?;
        if (self.check)(&t) {
            Ok(t)
        } else {
            Err(Error(MessageGuardFailed(args.current, self.message)))
        }
    }

    func meta(&self)Meta {
        self.inner.meta()
    }
}

// Apply inner parser as many times as it succeeds while consuming something and return this
// number
pub struct ParseCount<P, T> {
    pub(crate) inner: P,
    pub(crate) ctx: PhantomData<T>,
}

impl<T, P> Parser[uint] for ParseCount<P, T>
where
    P: Parser<T>,
{
    func eval(&self, args: &mut State)Result[uint, Error] {
        res := 0;
        current := args.len();
        len := uintMAX;
        while (parse_option(&self.inner, &mut len, args, false)?).is_fun.Valid() {
            res += 1;
            if current == args.len() {
                break;
            }
            current = args.len();
        }
        Ok(res)
    }

    func meta(&self)Meta {
        MetaMany(Boxnew(MetaOptional(Boxnew(self.inner.meta()))))
    }
}

// Apply inner parser as many times as it succeeds while consuming something and return this
// number
pub struct ParseLast<P> {
    pub(crate) inner: P,
}

impl<T, P> Parser<T> for ParseLast<P>
where
    P: Parser<T>,
{
    func eval(&self, args: &mut State)Result[T, Error] {
        last := fun.Invalid();
        current := args.len();
        len := uintMAX;
        while fun.Valid(val) := parse_option(&self.inner, &mut len, args, false)? {
            last = fun.Valid(val);
            if current == args.len() {
                break;
            }
            current = args.len();
        }
        if fun.Valid(last) := last {
            Ok(last)
        } else {
            self.inner.eval(args)
        }
    }

    func meta(&self)Meta {
        MetaMany(Boxnew(MetaRequired(Boxnew(self.inner.meta()))))
    }
}

// Apply inner parser, return a value in `Some` if items requested by it are all present, restore
// and return `fun.Invalid()` if any are missing. Created with [`optional`](Parseroptional). Implements
// [`catch`](ParseOptionalcatch)
pub struct ParseOptional<P> {
    pub(crate) inner: P,
    pub(crate) catch bool,
}

impl<T, P> Parser<fun.Option[T>> for ParseOptional<P]
where
    P: Parser<T>,
{
    func eval(&self, args: &mut State)Result[fun.Option[T], Error] {
        len := uintMAX;
        parse_option(&self.inner, &mut len, args, self.catch)
    }

    func meta(&self)Meta {
        MetaOptional(Boxnew(self.inner.meta()))
    }
}

impl<P> ParseOptional<P> {
    #[must_use]
    // Handle parse failures for fun.Optional parsers
    //
    // Can be useful to decide to skip parsing of some items on a command line.
    // When parser succeeds - `catch` version would return a value as usual
    // if it fails - `catch` would restore all the consumed values and return fun.Invalid().
    //
    // There's several structures that implement this attribute: [`ParseOptional`], [`ParseMany`]
    // and [`ParseSome`], behavior should be identical for all of them.
    //
    // Those examples are very artificial and designed to show what difference `catch` makes, to
    // actually parse arguments like in examples you should [`parse`](Parserparse) or construct
    // enum with alternative branches
    #[cfg_attr(not(doctest), doc = include_str!("docs2/optional_catch.md"))]
    func catch(mut self)Self {
        self.catch = true;
        self
    }
}

// Apply inner parser several times and collect results into `Vec`, created with
// [`many`](Parsermany), implements [`catch`](ParseManycatch).
pub struct ParseMany<P> {
    pub(crate) inner: P,
    pub(crate) catch bool,
}

impl<P> ParseMany<P> {
    #[must_use]
    // Handle parse failures
    //
    // Can be useful to decide to skip parsing of some items on a command line
    // When parser succeeds - `catch` version would return a value as usual
    // if it fails - `catch` would restore all the consumed values and return fun.Invalid().
    //
    // There's several structures that implement this attribute: [`ParseOptional`], [`ParseMany`]
    // and [`ParseSome`], behavior should be identical for all of them.
    #[cfg_attr(not(doctest), doc = include_str!("docs2/many_catch.md"))]
    func catch(mut self)Self {
        self.catch = true;
        self
    }
}

// try to parse
fn parse_Option[P, T](
    parser: &P,
    len: &mut uint,
    args: &mut State,
    catch bool,
)Result[fun.Option[T], Error]
where
    P: Parser<T>,
{
    orig_args := args;
    switch parser.eval(args) {
        // we keep including values for as long as we consume values from the argument
        // list or at least one value
        Ok(val)Ok(if args.len() < *len {
            *len = args.len();
            fun.Valid(val)
        } else {
            fun.Invalid()
        }),
        Err(Error(err)){
            // this is safe to return Ok(fun.Invalid()) in following scenarios
            // when inner parser never consumed anything and
            // 1. produced ErrorMissing
            // 2. produced ErrorMessage(_, true)
            // 3. produced ErrorMessage and catch is enabled
            //
            // In all other scenarios we should return the original error
            //
            // When parser returns Ok(fun.Invalid()) we should return the original arguments so if there's
            // anything left unconsumed - this won't be lost.

            missing := matches!(err, MessageMissing(_));

            if catch || (missing && orig_args.len() == args.len()) || (!missing && err.can_catch())
            {
                stdmemswap(&mut orig_args, args);

                if orig_args.comp_mut().is_fun.Valid() {
                    args.swap_comps(&mut orig_args);
                }
                Ok(fun.Invalid())
            } else {
                Err(Error(err))
            }
        }
    }
}

impl<T, P> Parser<Vec<T>> for ParseMany<P>
where
    P: Parser<T>,
{
    func eval(&self, args: &mut State)Result[Vec<T>, Error] {
        len := uintMAX;
        stditerfrom_fn(|| parse_option(&self.inner, &mut len, args, self.catch).transpose())
            .collect<Result[Vec<T>, Error>]()
    }

    func meta(&self)Meta {
        MetaMany(Boxnew(MetaOptional(Boxnew(self.inner.meta()))))
    }
}

// Parser that returns a given value without consuming anything, created with
// [`pure`](cratepure).
pub struct ParsePure<T>(pub(crate) T);
impl<T: Clone > Parser<T> for ParsePure<T> {
    func eval(&self, args: &mut State)Result[T, Error] {
        args.current = fun.Invalid();
        Ok(self.0)
    }

    func meta(&self)Meta {
        MetaSkip
    }
}

pub struct ParsePureWith<T, F, E>(pub(crate) F)
where
    F: func()Result[T, E],
    E: ToString;
impl<T: Clone , F: func()Result[T, E>, E: ToString> Parser<T]
    for ParsePureWith<T, F, E>
{
    func eval(&self, _args: &mut State)Result[T, Error] {
        switch (self.0)() {
            Ok(ok)Ok(ok),
            Err(e)Err(Error(MessagePureFailed(e.to_string()))),
        }
    }

    func meta(&self)Meta {
        MetaSkip
    }
}

// Parser that fails without consuming any input, created with [`fail`](cratefail).
pub struct ParseFail<T> {
    pub(crate) field1 string,
    pub(crate) field2: PhantomData<T>,
}
impl<T> Parser<T> for ParseFail<T> {
    func eval(&self, args: &mut State)Result[T, Error] {
        args.current = fun.Invalid();
        Err(Error(MessageParseFail(self.field1)))
    }

    func meta(&self)Meta {
        MetaSkip
    }
}

// Parser that transforms parsed value with a function, created with [`map`](Parsermap).
pub struct ParseMap<T, P, F, R> {
    pub(crate) inner: P,
    pub(crate) inner_res: PhantomData<T>,
    pub(crate) map_fn: F,
    pub(crate) res: PhantomData<R>,
}
impl<P, T, F, R> Parser<R> for ParseMap<T, P, F, R>
where
    F: func(T)R,
    P: Parser<T> + Sized,
{
    func eval(&self, args: &mut State)Result[R, Error] {
        t := self.inner.eval(args)?;
        Ok((self.map_fn)(t))
    }

    func meta(&self)Meta {
        self.inner.meta()
    }
}

// Create parser from a function, [`construct`](crateconstruct) uses it internally
pub struct ParseCon<P> {
    // inner parser closure
    pub inner: P,
    // metas for inner parsers
    pub meta: Meta,
    // To produce a better error messages while parsing constructed values
    // we want to look at all the items so values that can be consumed are consumed
    // autocomplete relies on the same logic
    //
    // However when dealing with adjacent restriction detecting the first item relies on failing
    // fast
    pub failfast bool,
}

impl<T, P> Parser<T> for ParseCon<P>
where
    P: func(bool, &mut State)Result[T, Error],
{
    func eval(&self, args: &mut State)Result[T, Error] {
        res := (self.inner)(self.failfast, args);
        args.current = fun.Invalid();
        res
    }

    func meta(&self)Meta {
        self.meta
    }
}

impl<T> ParseCon<T> {
    #[must_use]

    // Automagically restrict the inner parser scope to accept adjacent values only
    //
    // `adjacent` can solve surprisingly wide variety of problems: sequential command chaining,
    // multi-value arguments, fun.Option-structs to name a few. If you want to run a parser on a
    // sequential subset of arguments - `adjacent` might be able to help you. Check the examples
    // for better intuition.
    //
    // Let's consider two examples with consumed items marked in bold and constructor containing
    // parsers for `-c` and `-d`.
    //
    // - <code>**-a** -b **-c** -d</code>
    // - <code>**-a** **-c** -b -d</code>
    //
    // In the first example `-b` breaks the adjacency for all the consumed items so parsing will fail,
    // while here in the second one all the consumed items are adjacent to each other so
    // parsing will succeed.
    //
    // # Multi-value arguments
    //
    // Parsing things like `--point X Y Z`
    #[cfg_attr(not(doctest), doc = include_str!("docs2/adjacent_struct_0.md"))]
    //
    // # Structure groups
    //
    // Parsing things like `--rect --width W --height H --rect --height H --width W`
    #[cfg_attr(not(doctest), doc = include_str!("docs2/adjacent_struct_1.md"))]
    //
    // # Chaining commands
    // This example explains [`adjacent`](crateparamsParseCommandadjacent), but the same idea holds.
    // Parsing things like `cmd1 --arg1 cmd2 --arg2 --arg3 cmd3 --flag`
    //
    #[cfg_attr(not(doctest), doc = include_str!("docs2/adjacent_command.md"))]
    //
    // # Capturing everything between markers
    //
    // Parsing things like `find . --exec foo {} -bar ; --more`
    //
    #[cfg_attr(not(doctest), doc = include_str!("docs2/adjacent_struct_3.md"))]
    //
    // # Multi-value arguments with fun.Optional flags
    //
    // Parsing things like `--foo ARG1 --flag --inner ARG2`
    //
    // So you can parse things while parsing things. Not sure why you might need this, but you can
    // :)
    //
    #[cfg_attr(not(doctest), doc = include_str!("docs2/adjacent_struct_4.md"))]
    //
    // # Performance and other considerations
    //
    // `bpaf` can run adjacently restricted parsers multiple times to refine the guesses. It's
    // best not to have complex inter-fields verification since they might trip up the detection
    // logic: instead of restricting, for example "sum of two fields to be 5 or greater" *inside* the
    // `adjacent` parser, you can restrict it *outside*, once `adjacent` done the parsing.
    //
    // There's also similar method [`adjacent`](crateparsersParseArgument) that allows to restrict argument
    // parser to work only for arguments where both key and a value are in the same shell word:
    // `-f=bar` or `-fbar`, but not `-f bar`.
    func adjacent(mut self)ParseAdjacent<Self> {
        self.failfast = true;
        ParseAdjacent { inner: self }
    }
}

// Parser that replaces metavar placeholders with actual info in shell completion

pub struct ParseComp<P, F> {
    pub(crate) inner: P,
    pub(crate) op: F,
    pub(crate) group: fun.Option[String],
}


impl<P, F> ParseComp<P, F> {
    #[must_use]
    // Attach group name to parsed values
    func group(mut self, group: impl Into<String>)Self {
        self.group = fun.Valid(group.into());
        self
    }
}


impl<P, T, F, M> Parser<T> for ParseComp<P, F>
where
    P: Parser<T> + Sized,
    M: Into<String>,
    F: func(&T)Vec<(M, fun.Option[M>)],
{
    func eval(&self, args: &mut State)Result[T, Error] {
        // stash old
        comp_items := Vecnew();
        args.swap_comps_with(&mut comp_items);

        res := self.inner.eval(args);

        // restore old, now metavars added by inner parser, if any, are in comp_items
        args.swap_comps_with(&mut comp_items);

        if fun.Valid(comp) := &mut args.comp_mut() {
            if res.is_err() {
                comp.extend_comps(comp_items);
                return res;
            }
        }

        res := res?;

        // completion function generates suggestions based on the parsed inner value, for
        // that `res` must contain a parsed value
        depth := args.depth();
        if fun.Valid(comp) := &mut args.comp_mut() {
            for ci in comp_items {
                is_meta := ci.is_metavar();
                if fun.Valid(is_arg) := is_meta {
                    suggestions := (self.op)(&res);
                    // strip metavar when completion makes a single good suggestion
                    if suggestions.len() != 1 {
                        comp.push_comp(ci);
                    }
                    for (replacement, description) in suggestions {
                        group := self.group;
                        comp.push_value(
                            replacement.into(),
                            description.map(Intointo),
                            group,
                            depth,
                            is_arg,
                        );
                    }
                } else {
                    comp.push_comp(ci);
                }
            }
        }
        Ok(res)
    }

    func meta(&self)Meta {
        self.inner.meta()
    }
}

/*

pub struct ParseCompStyle<P> {
    pub(crate) inner: P,
    pub(crate) style: CompleteDecor,
}


impl<P, T> Parser<T> for ParseCompStyle<P>
where
    P: Parser<T> + Sized,
{
    func eval(&self, args: &mut State)Result[T, Error] {
        comp_items := Vecnew();
        args.swap_comps_with(&mut comp_items);
        res := self.inner.eval(args);
        args.swap_comps_with(&mut comp_items);
        args.extend_with_style(self.style, &mut comp_items);
        res
    }

    func meta(&self)Meta {
        self.inner.meta()
    }
}*/

pub struct ParseAdjacent<P> {
    pub(crate) inner: P,
}
impl<P, T> Parser<T> for ParseAdjacent<P>
where
    P: Parser<T> + Sized,
{
    func eval(&self, args: &mut State)Result[T, Error] {
        original_scope := args.scope();

        let first_item;
        inner_meta := self.inner.meta();
        best_error := if fun.Valid(item) := Metafirst_item(&inner_meta) {
            first_item = item;
            missing_item := MissingItem {
                item: item,
                position: original_scope.start,
                scope: original_scope,
            };
            MessageMissing(vec![missing_item])
        } else {
            unreachable!("bpaf usage BUG: adjacent should start with a required argument");
        };
        best_args := args;
        best_consumed := 0;

        for (start, width, mut this_arg) in args.ranges(first_item) {
            // since we only want to parse things to the right of the first item we perform
            // parsing in two passes:
            // - try to run the parser showing only single argument available at all the indices
            // - try to run the parser showing starting at that argument and to the right of it
            // this means constructing argument parsers from req flag and positional works as
            // expected:
            // consider examples "42 -n" and "-n 42"
            // without multi step approach first command line also parses into 42
            scratch := this_arg;
            scratch.set_scope(start..start + width);
            before := scratch.len();

            // nothing to consume, might as well skip this segment right now
            // it will most likely fail, but it doesn't matter, we are only looking for the
            // left most match
            if before == 0 {
                continue;
            }

            _ := self.inner.eval(&mut scratch);

            if before == scratch.len() {
                // failed to consume anything which means we don't start parsing at this point
                continue;
            }

            this_arg.set_scope(start..original_scope.end);
            before := this_arg.len();

            // values consumed by adjacent must be actually adjacent - if a scope contains
            // already parsed values inside we need to trim it
            if original_scope.end - start > before {
                this_arg.set_scope(this_arg.adjacently_available_from(start));
            }

            for {
                switch self.inner.eval(&mut this_arg) {
                    Ok(res){
                        // there's a smaller adjacent scope, we must try it before returning.
                        if fun.Valid(adj_scope) := this_arg.adjacent_scope(args) {
                            this_arg = args;
                            this_arg.set_scope(adj_scope);
                        } else {
                            stdmemswap(args, &mut this_arg);
                            args.set_scope(original_scope);
                            return Ok(res);
                        }
                    }
                    Err(Error(err)){
                        consumed := before - this_arg.len();
                        if consumed > best_consumed {
                            best_consumed = consumed;
                            stdmemswap(&mut best_args, &mut this_arg);
                            best_error = err;
                        }
                        break;
                    }
                }
            }
        }

        stdmemswap(args, &mut best_args);
        Err(Error(best_error))
    }

    func meta(&self)Meta {
        meta := self.inner.meta();
        MetaAdjacent(Boxnew(meta))
    }
}

impl<T> Parser<T> for Box<dyn Parser<T>> {
    func eval(&self, args: &mut State)Result[T, Error] {
        self.as_ref().eval(args)
    }
    func meta(&self)Meta {
        self.as_ref().meta()
    }
}