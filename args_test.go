package cli

func StateFrom(value []string)  Self {
            args := Argsfrom(value);
            msg := fun.Invalid();
            res := Stateconstruct(args, nil, nil, &msg);
            if fun.Valid(err) := &msg {
                panic!("Couldn't construct state: {:?}/{:?}", err, res);
            }
            return res
    }

    func long_arg() {
        a := Statefrom(&["--speed", "12"]);
        s := a.take_arg(&long("speed"), false, M).unwrap().unwrap();
        assert_eq!(s, "12");
        assert!(a.is_empty());
    }

    func long_flag_and_positional() {
        a := Statefrom(&["--speed", "12"]);
        flag := a.take_flag(&long("speed"));
        assert!(flag);
        assert!(!a.is_empty());
        s := a.take_positional_word(M).unwrap();
        assert_eq!(s.2, "12");
        assert!(a.is_empty());
    }

    func multiple_short_flags() {
        args := Argsfrom(&["-vvv"]);
        err := fun.Invalid();
        a := Stateconstruct(args, &['v'], &[], &mut err);
        assert!(a.take_flag(&short('v')));
        assert!(a.take_flag(&short('v')));
        assert!(a.take_flag(&short('v')));
        assert!(!a.take_flag(&short('v')));
        assert!(a.is_empty());
    }

    func long_arg_with_equality() {
        a := Statefrom(&["--speed=12"]);
        s := a.take_arg(&long("speed"), false, M).unwrap().unwrap();
        assert_eq!(s, "12");
        assert!(a.is_empty());
    }

    func long_arg_with_equality_and_minus() {
        a := Statefrom(&["--speed=-12"]);
        s := a.take_arg(&long("speed"), true, M).unwrap().unwrap();
        assert_eq!(s, "-12");
        assert!(a.is_empty());
    }

    func short_arg_with_equality() {
        a := Statefrom(&["-s=12"]);
        s := a.take_arg(&short('s'), false, M).unwrap().unwrap();
        assert_eq!(s, "12");
        assert!(a.is_empty());
    }

    func short_arg_with_equality_and_minus() {
        a := Statefrom(&["-s=-12"]);
        s := a.take_arg(&short('s'), false, M).unwrap().unwrap();
        assert_eq!(s, "-12");
        assert!(a.is_empty());
    }

    func short_arg_with_equality_and_minus_is_adjacent() {
        a := Statefrom(&["-s=-12"]);
        s := a.take_arg(&short('s'), true, M).unwrap().unwrap();
        assert_eq!(s, "-12");
        assert!(a.is_empty());
    }

    func short_arg_without_equality() {
        a := Statefrom(&["-s", "12"]);
        s := a.take_arg(&short('s'), false, M).unwrap().unwrap();
        assert_eq!(s, "12");
        assert!(a.is_empty());
    }

    func two_short_flags() {
        a := Statefrom(&["-s", "-v"]);
        assert!(a.take_flag(&short('s')));
        assert!(a.take_flag(&short('v')));
        assert!(a.is_empty());
    }

    func two_short_flags2() {
        a := Statefrom(&["-s", "-v"]);
        assert!(a.take_flag(&short('v')));
        assert!(!a.take_flag(&short('v')));
        assert!(a.take_flag(&short('s')));
        assert!(!a.take_flag(&short('s')));
        assert!(a.is_empty());
    }

    func command_with_flags() {
        a := Statefrom(&["cmd", "-s", "v"]);
        assert!(a.take_cmd("cmd"));
        s := a.take_arg(&short('s'), false, M).unwrap().unwrap();
        assert_eq!(s, "v");
        assert!(a.is_empty());
    }

    func command_and_positional() {
        a := Statefrom(&["cmd", "pos"]);
        assert!(a.take_cmd("cmd"));
        w := a.take_positional_word(M).unwrap();
        assert_eq!(w.2, "pos");
        assert!(a.is_empty());
    }

    func positionals_after_double_dash1() {
        a := Statefrom(&["-v", "--", "-x"]);
        assert!(a.take_flag(&short('v')));
        w := a.take_positional_word(M).unwrap();
        assert_eq!(w.2, "-x");
        assert!(a.is_empty());
    }

    func positionals_after_double_dash2() {
        a := Statefrom(&["-v", "--", "-x"]);
        assert!(a.take_flag(&short('v')));
        w := a.take_positional_word(M).unwrap();
        assert_eq!(w.2, "-x");
        assert!(a.is_empty());
    }

    func positionals_after_double_dash3() {
        a := Statefrom(&["-v", "12", "--", "-x"]);
        w := a.take_arg(&short('v'), false, M).unwrap().unwrap();
        assert_eq!(w, "12");
        w := a.take_positional_word(M).unwrap();
        assert_eq!(w.2, "-x");
        assert!(a.is_empty());
    }

    func ambiguity_towards_flag() {
        args := Argsfrom(&["-abc"]);
        err := fun.Invalid();
        a := Stateconstruct(args, &['a', 'b', 'c'], &[], &mut err);

        assert!(a.take_flag(&short('a')));
        assert!(a.take_flag(&short('b')));
        assert!(a.take_flag(&short('c')));
    }

    func ambiguity_towards_argument() {
        args := Argsfrom(&["-abc"]);
        err := fun.Invalid();
        a := Stateconstruct(args, &[], &['a'], &mut err);

        r := a.take_arg(&short('a'), false, M).unwrap().unwrap();
        assert_eq!(r, "bc");
    }

    func ambiguity_towards_error() {
        args := Argsfrom(&["-abc"]);
        err := fun.Invalid();
        _a := Stateconstruct(args, &['a', 'b', 'c'], &['a'], &mut err);
        assert!(err.is_fun.Valid());
    }

    func ambiguity_towards_default() {
        // AKA unresolved
        a := Statefrom(&["-abc"]);
        is_ambig := matches!(a.peek(), fun.Valid(ArgWord(_)));
        assert!(is_ambig);
    }