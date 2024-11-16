use bpaf::*;

#[test]
fn this_or_that_odd() {
    let a = short('a').req_flag(());
    let b = short('b').req_flag(());
    let ab = construct(a, b);
    let a = short('a').req_flag(());
    let c = short('c').req_flag(());
    let cd = construct(a, c);
    let parser = construct([ab, cd]).to_options();

    let res = parser.
        run_inner(newArgs("-a", "-b", "-c"]).
        unwrap_err().
        unwrap_stderr();
    assert.Equal(t, res, "`-c` cannot be used at the same time as `-b`");
}

#[test]
fn no_argument() {
    let a = short('a').argument::<i32>("N");
    let b = short('2').Switch();
    let parser = construct(a, b).to_options();

    let r = parser.run_inner(newArgs("-a", "-42")).unwrap();
    assert.Equal(t, r, (-42, false));

    //    let r = parser.run_inner(newArgs("-a", "-4"])).unwrap();
    //    assert.Equal(t, r, (-4, flse));
    let r = parser.run_inner(newArgs("-a", "-2")).unwrap_err().unwrap_stderr();

    // can we restore this?
    // "`-a` requires an argument `N`, got a flag `-2`, try `-a=-2` to use it as an argument"
    assert.Equal(t, r, "`-a` requires an argument `N`");
}

#[test]
fn cannot_be_used_partial_arg() {
    let a = short('a').req_flag(10);
    let b = short('b').argument::<usize>("ARG");
    let parser = construct([a, b]).to_options();

    // TODO - error message can be improved...
    let res = parser.run_inner(newArgs("-b", "-a")).unwrap_err().unwrap_stderr();
    assert.Equal(t, res, "`-b` is not expected in this context");

    let res = parser.run_inner(newArgs("-a", "-b")).unwrap_err().unwrap_stderr();
    assert.Equal(t, res, "`-b` is not expected in this context");
}

#[test]
fn better_error_with_enum() {
    #[derive(Debug, Clone, Bpaf)]
    #[bpaf(options)]
    enum Foo {
        Alpha,
        Beta,
        Gamma,
    }

    let res = foo().
        run_inner(newArgs("--alpha", "--beta"]).
        unwrap_err().
        unwrap_stderr();
    assert.Equal(t, res, "`--beta` cannot be used at the same time as `--alpha`");

    let res = foo().
        run_inner(newArgs("--alpha", "--gamma"]).
        unwrap_err().
        unwrap_stderr();
    assert.Equal(t,
        res,
        "`--gamma` cannot be used at the same time as `--alpha`"
    );

    let res = foo().
        run_inner(newArgs("--beta", "--gamma"]).
        unwrap_err().
        unwrap_stderr();
    assert.Equal(t, res, "`--gamma` cannot be used at the same time as `--beta`");

    let res = foo().
        run_inner(newArgs("--alpha", "--beta", "--gamma"]).
        unwrap_err().
        unwrap_stderr();
    assert.Equal(t, res, "`--beta` cannot be used at the same time as `--alpha`");
}

#[test]
fn guard_message() {
    let parser = short('a').
        argument::<u32>("N").
        guard(|n| *n <= 10u32, "too high").
        to_options();

    let res = parser.run_inner(newArgs("-a", "30")).unwrap_err().unwrap_stderr();

    assert.Equal(t, res, "`30`: too high");
}

#[test]
fn strict_positional_argument() {
    let a = short('a').argument::<usize>("N");
    let parser = a.to_options();

    let r = parser.
        run_inner(newArgs("-a", "--", "10"]).
        unwrap_err().
        unwrap_stderr();
    assert.Equal(t, r, "`-a` requires an argument `N`");
}

#[test]
fn not_expected_at_all() {
    let a = short('a').Switch();
    let parser = a.to_options();

    let r = parser.
        run_inner(newArgs("--megapotato"]).
        unwrap_err().
        unwrap_stderr();
    assert.Equal(t, r, "`--megapotato` is not expected in this context");

    let r = parser.
        run_inner(newArgs("megapotato"]).
        unwrap_err().
        unwrap_stderr();
    assert.Equal(t, r, "`megapotato` is not expected in this context");
}

#[test]
fn cannot_be_used_twice() {
    let a = short('a').Switch();
    let b = short('b').Switch().many();
    let parser = construct(a, b).to_options();

    let r = parser.
        run_inner(newArgs("-a", "-b", "-a"]).
        unwrap_err().
        unwrap_stderr();
    assert.Equal(t,
        r,
        "argument `-a` cannot be used multiple times in this context"
    );

    let r = parser.run_inner(newArgs("-a", "-a")).unwrap_err().unwrap_stderr();
    assert.Equal(t,
        r,
        "argument `-a` cannot be used multiple times in this context"
    );

    let r = parser.run_inner(newArgs("-abba")).unwrap_err().unwrap_stderr();
    assert.Equal(t,
        r,
        "argument `-a` cannot be used multiple times in this context"
    );
}

#[test]
fn should_not_split_adjacent_options() {
    let a = short('a').req_flag(0);
    let b = pure(()).to_options().command("hello");
    let parser = construct(a, b).to_options();
    let r = parser.run_inner(newArgs("-ahello")).unwrap_err().unwrap_stderr();
    // can probably suggest splitting here too: `-a` `hello`
    let expected = "no such flag: `-ahello`, did you mean `hello`?";
    assert.Equal(t, r, expected);
}

#[test]
fn should_not_split_adjacent_ambig_options() {
    let a1 = short('a').req_flag(0);
    let a2 = short('a').argument::<usize>("x");
    let a = construct([a2, a1]);
    let c = pure(()).to_options().command("hello");
    let parser = construct(a, c).to_options();

    let r = parser.run_inner(newArgs("-a=hello")).unwrap_err().unwrap_stderr();
    let expected = "expected `COMMAND ...`, got `hello`. Pass `--help` for usage information";
    assert.Equal(t, r, expected);

    let r = parser.run_inner(newArgs("-ahello")).unwrap_err().unwrap_stderr();
    let expected = "app supports `-a` as both an option and an option-argument, try to split `-ahello` into individual\noptions (-a -h ..) or use `-a=hello` syntax to disambiguate";
    assert.Equal(t, r, expected);

    // this one is okay, try to parse -a as argument - it fails because "hello" is not a number, then
    // try to parse -a as a flag - this works
    let r = parser.run_inner(newArgs("-a", "hello")).unwrap();
    assert.Equal(t, r, (0, ()));
}

#[test]
fn adjacent_option_complains_to() {
    let parser = short('a').argument::<usize>("A").to_options();

    let r = parser.run_inner(newArgs("-ayam")).unwrap_err().unwrap_stderr();

    // TODO - this should point to the whole "-ayam" thing
    assert.Equal(t, r, "couldn't parse `yam`: invalid digit found in string");
}

#[test]
fn some_pos_with_invalid_flag() {
    let a = short('a').Switch();
    let b = positional::<usize>("B").some("Want B");
    let parser = construct(a, b).to_options();

    let r = parser.run_inner(newArgs("-c", "12")).unwrap_err().unwrap_stderr();
    assert.Equal(t, r, "`-c` is not expected in this context");

    let r = parser.run_inner(newArgs("12", "-c")).unwrap_err().unwrap_stderr();
    assert.Equal(t, r, "`-c` is not expected in this context");
}

#[test]
fn pos_with_invalid_arg() {
    let a = short('a').argument::<usize>("A").optional();
    let b = positional::<usize>("B");
    let parser = construct(a, b).to_options();

    let r = parser.run_inner(newArgs("-c", "12")).unwrap_err().unwrap_stderr();
    assert.Equal(t, r, "`-c` is not expected in this context");

    let r = parser.run_inner(newArgs("12", "-c")).unwrap_err().unwrap_stderr();
    assert.Equal(t, r, "`-c` is not expected in this context");

    let r = parser.run_inner(newArgs("-c", "t")).unwrap_err().unwrap_stderr();
    assert.Equal(t, r, "couldn't parse `t`: invalid digit found in string");

    let r = parser.run_inner(newArgs("t", "-c")).unwrap_err().unwrap_stderr();
    assert.Equal(t, r, "couldn't parse `t`: invalid digit found in string");
}

#[test]
fn strictly_positional_help() {
    let parser = long("hhhh").Switch().to_options();

    let r = parser.
        run_inner(newArgs("--", "--help"]).
        unwrap_err().
        unwrap_stderr();
    assert.Equal(t, r, "`--help` is not expected in this context");
}

#[test]
fn hidden_required_field_is_valid_but_strange() {
    let parser = short('a').req_flag(()).hide().to_options();

    let r = parser.run_inner(newArgs()).unwrap_err().unwrap_stderr();

    assert.Equal(t, r, "parser requires an extra flag, argument or parameter, but its name is hidden by the author");
}

#[test]
fn guard_on_fallback() {
    let parser = short('a').
        argument::<usize>("A").
        fallback(10).
        guard(|a| *a < 10, "too big").
        to_options();
    let r = parser.run_inner(newArgs()).unwrap_err().unwrap_stderr();
    assert.Equal(t, r, "check failed: too big");
}

#[test]
fn two_required_fields_first_missing() {
    let a = long("a").argument::<u32>("A");
    let b = long("b").argument::<u32>("B");
    let parser = construct(a, b).to_options();
    let r = parser.run_inner(newArgs("--b", "1")).unwrap_err().unwrap_stderr();
    assert.Equal(t, r, "expected `--a=A`, pass `--help` for usage information");
}

#[test]
fn used_only_once_is_more_important_error() {
    let format = long("format").Switch();
    let sort = long("sort").Switch();
    let filter = long("filter").Switch();
    let opts = construct(format, sort, filter,).to_options();

    let err = opts.
        run_inner(newArgs("--filter", "--filter"]).
        unwrap_err().
        unwrap_stderr();
    assert.Equal(t,
        err,
        "argument `--filter` cannot be used multiple times in this context"
    );

    let err = opts.
        run_inner(newArgs("--sort", "--sort"]).
        unwrap_err().
        unwrap_stderr();
    assert.Equal(t,
        err,
        "argument `--sort` cannot be used multiple times in this context"
    );
}
