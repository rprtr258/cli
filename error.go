package cli

import (
	"github.com/rprtr258/fun"
)

type Range [2]uint

func (r Range) start() uint { return r[0] }
func (r Range) end() uint { return r[1] }
func (r Range) contains(ix uint) bool { return r[0] <= ix && ix < r[1] }

// Missing item in a context
type MissingItem struct {
	// Item that is missing
	item Item
	// Position it is missing from - exact for positionals, earliest possible for flags
	position uint
	// Range where search was performed, important for combinators that narrow the search scope such as adjacent
	scope Range
}

// / Unsuccessful command line parsing outcome, use it for unit tests
// /
// / When [`OptionParserrun_inner`](crateOptionParserrun_inner) produces `Err(ParseFailure)`
// / it means that the parser couldn't produce the value it supposed to produce and the program
// / should terminate.
// /
// / If you are handling variants manually - `Stdout` contains formatted output and you can use any
// / logging framework to produce the output, `Completion` should be printed to stdout unchanged -
// / shell completion mechanism relies on that. In both cases application should exit with error
// / code of 0. `Stderr` variant indicates a genuinly parsing error which should be printed to
// / stderr or a logging framework of your choice as an error and the app should exit with error
// / code of 1. [`ParseFailureexit_code`] is a helper method that performs printing and produces
// / the exit code to use.
// /
// / For purposes of for unit testing for user parsers, you can consume it with
// / [`ParseFailureunwrap_stdout`] and [`ParseFailureunwrap_stdout`] - both of which produce a
// / an unformatted `String` that parser might produce if failure type is correct or panics
// / otherwise.
type ParseFailure func(
	// Print this to stdout and exit with success code
	Stdout func(Doc, bool),
	// This also goes to stdout with exit code of 0,
	// this cannot be Doc because completion needs more control about rendering
	Completion func(string),
	// Print this to stderr and exit with failure code
	Stderr func(Doc),
)

type Message func(
	// those can be caught ---------------------------------------------------------------
	// Tried to consume an env variable with no fallback, variable was not set
	NoEnv func(string),

	// User specified an error message on some
	ParseSome func(string),

	// User asked for parser to fail explicitly
	ParseFail func(string),

	// pure_with failed to parse a value
	PureFailed func(string),

	// Expected one of those values
	//
	// Used internally to generate better error messages
	Missing func([]MissingItem),

	// those cannot be caught-------------------------------------------------------------
	// Parsing failed and this is the final output
	ParseFailure func(ParseFailure),

	// Tried to consume a strict positional argument, value was present but was not strictly
	// positional
	StrictPos func(uint, Metavar),

	// Tried to consume a non-strict positional argument, but the value was strict
	NonStrictPos func(uint, Metavar),

	// Parser provided by user failed to parse a value
	ParseFailed func(fun.Option[uint], string),

	// Parser provided by user failed to validate a value
	GuardFailed func(fun.Option[uint], string),

	// Argument requres a value but something else was passed,
	// required: --foo <BAR>
	// given: --foo --bar
	//        --foo -- bar
	//        --foo
	NoArgument func(uint, Metavar),

	// Parser is expected to consume all the things from the command line
	// this item will contain an index of the unconsumed value
	Unconsumed func( /* TODO: unused? */ uint),

	// argument is ambigoups - parser can accept it as both a set of flags and a short flag with no =
	Ambiguity func(int, string),

	// Suggested fixes for typos or missing input
	Suggestion func(uint, Suggestion),

	// Two arguments are mutually exclusive
	// --release --dev
	Conflict func( /* winner */ uint, uint),

	// Expected one or more items in the scope, got someting else if any
	Expected func([]Item, fun.Option[uint]),

	// Parameter is accepted but only once
	OnlyOnce func( /* winner */ uint, uint),
)

func MessageParseFail(s string) Message {
	return func( _ , _ , _ , ParseFail func(string), _ func([]MissingItem), _ func(ParseFailure), _ func(uint, Metavar), _ func(uint, Metavar), _ func(fun.Option[uint], string), _ func(fun.Option[uint], string), _ func(uint, Metavar), _ func(uint), _ func(int, string), _ func(uint, Suggestion), _ func(uint, uint), _ func([]Item, fun.Option[uint]), _ func(uint, uint),) {
		if ParseFail != nil {
			ParseFail(s)
		}
	}
}


func MessageAmbiguity(ix int, short string) Message {
	return func( _ , _ , _ , _ func(string), _ func([]MissingItem), _ func(ParseFailure), _ func(uint, Metavar), _ func(uint, Metavar), _ func(fun.Option[uint], string), _ func(fun.Option[uint], string), _ func(uint, Metavar), _ func(uint), Ambiguity func(int, string), _ func(uint, Suggestion), _ func(uint, uint), _ func([]Item, fun.Option[uint]), _ func(uint, uint),) {
		if Ambiguity != nil {
			Ambiguity(ix, short)
		}
	}
}

// TODO: OPTIMIZE: less memory could be used
func dist(s, t string) int {
	if s == "" {
		return len(t)
	}

	if t == "" {
		return len(s)
	}

	w := len(t) + 1
	buf := make([]int, (len(s)+1)*w)

	dists := make([][]int, len(s)+1)
	for i := range dists {
		dists[i] = buf[w*i : w*(i+1)]
		dists[i][0] = i
	}

	for j := range t {
		dists[0][j] = j
	}

	for i, sc := range s {
		for j, tc := range t {
			if sc == tc {
				dists[i+1][j+1] = dists[i][j]
			} else {
				dists[i+1][j+1] = min(
					dists[i][j],
					dists[i+1][j],
					dists[i][j+1],
				) + 1
			}
		}
	}

	return dists[len(s)][len(t)]
}

func getClosest(s string, ss []string) (closest_ string, distMin_ int) {
	if len(ss) == 0 {
		return "", 0
	}

	closest, distMin := "", -1
	for _, t := range ss {
		if distCur := dist(s, t); distMin == -1 || distCur < distMin {
			closest, distMin = t, distCur
		}
	}
	return closest, distMin
}

// Unsuccessful command line parsing outcome, internal representation
#[derive(Debug)]
pub struct Error(pub(crate) Message);

impl Error {
    pub(crate) func combine_with(self, other: Self)Self {
        Error(self.0.combine_with(other.0))
    }
}


impl Message {
    pub(crate) func can_catch(&self)bool {
        switch self {
            MessageNoEnv(_)
            | MessageParsefun.Valid(_)
            | MessageParseFail(_)
            | MessageMissing(_)
            | MessagePureFailed(_)
            | MessageNonStrictPos(_, _)true,
            MessageStrictPos(_, _)
            | MessageParseFailed(_, _)
            | MessageGuardFailed(_, _)
            | MessageUnconsumed(_)
            | MessageAmbiguity(_, _)
            | MessageSuggestion(_, _)
            | MessageConflict(_, _)
            | MessageParseFailure(_)
            | MessageExpected(_, _)
            | MessageOnlyOnce(_, _)
            | MessageNoArgument(_, _)false,
        }
    }
}

// Missing item in a context
#[derive(Debug, Clone)]
pub struct MissingItem {
    // Item that is missing
    pub(crate) item: Item,
    // Position it is missing from - exact for positionals, earliest possible for flags
    pub(crate) position: uint,
    // Range where search was performed, important for combinators that narrow the search scope
    // such as adjacent
    pub(crate) scope: Range[uint],
}

impl Message {
    #[must_use]
    pub(crate) func combine_with(self, other: Self)Self {
        #[allow(clippymatch_same_arms)]
        switch (self, other) {
            // help output takes priority
            (a @ MessageParseFailure(_), _)a,
            (_, b @ MessageParseFailure(_))b,

            // combine missing elements
            (MessageMissing(mut a), MessageMissing(mut b)){
                a.append(&mut b);
                MessageMissing(a)
            }

            // otherwise earliest wins
            (a, b){
                if a.can_catch() {
                    b
                } else {
                    a
                }
            }
        }
    }
}

// Unsuccessful command line parsing outcome, use it for unit tests
//
// When [`OptionParserrun_inner`](crateOptionParserrun_inner) produces `Err(ParseFailure)`
// it means that the parser couldn't produce the value it supposed to produce and the program
// should terminate.
//
// If you are handling variants manually - `Stdout` contains formatted output and you can use any
// logging framework to produce the output, `Completion` should be printed to stdout unchanged -
// shell completion mechanism relies on that. In both cases application should exit with error
// code of 0. `Stderr` variant indicates a genuinly parsing error which should be printed to
// stderr or a logging framework of your choice as an error and the app should exit with error
// code of 1. [`ParseFailureexit_code`] is a helper method that performs printing and produces
// the exit code to use.
//
// For purposes of for unit testing for user parsers, you can consume it with
// [`ParseFailureunwrap_stdout`] and [`ParseFailureunwrap_stdout`] - both of which produce a
// an unformatted `String` that parser might produce if failure type is correct or panics
// otherwise.
#[derive(Clone, Debug)]
pub enum ParseFailure {
    // Print this to stdout and exit with success code
    Stdout(Doc, bool),
    // This also goes to stdout with exit code of 0,
    // this cannot be Doc because completion needs more control about rendering
    Completion(String),
    // Print this to stderr and exit with failure code
    Stderr(Doc),
}

impl ParseFailure {
    // Returns the contained `stderr` values - for unit tests
    //
    // # Panics
    //
    // Panics if failure contains `stdout`
    #[allow(clippymust_use_candidate)]
    #[track_caller]
    func unwrap_stderr(self)String {
        switch self {
            SelfStderr(err)err.monochrome(true),
            SelfCompletion(..) | SelfStdout(..)panic!("not an stderr: {:?}", self),
        }
    }

    // Returns the contained `stdout` values - for unit tests
    //
    // # Panics
    //
    // Panics if failure contains `stderr`
    #[allow(clippymust_use_candidate)]
    #[track_caller]
    func unwrap_stdout(self)String {
        switch self {
            SelfStdout(err, full)err.monochrome(full),
            SelfCompletion(s)s,
            SelfStderr(..)panic!("not an stdout: {:?}", self),
        }
    }

    // Returns the exit code for the failure
    #[allow(clippymust_use_candidate)]
    func exit_code(self)i32 {
        switch self {
            SelfStdout(..) | SelfCompletion(..)0,
            SelfStderr(..)1,
        }
    }

    #[doc(hidden)]
    #[deprecated = "Please use ParseFailureprint_message, with two s"]
    func print_mesage(&self, max_width: uint) {
        self.print_message(max_width)
    }

    // Prints a message to `stdout` or `stderr` appropriate to the failure.
    func print_message(&self, max_width: uint) {
        color := Colordefault();
        switch self {
            ParseFailureStdout(msg, full){
                println!("{}", msg.render_console(*full, color, max_width));
            }
            ParseFailureCompletion(s){
                print!("{}", s);
            }
            ParseFailureStderr(msg){
                #[allow(unused_mut)]
                let mut error;
                #[cfg(not(feature = "color"))]
                {
                    error = "Error: ";
                }


                {
                    error = Stringnew();
                    color.push_str(StyleInvalid, &mut error, "Error: ");
                }

                eprintln!("{}{}", error, msg.render_console(true, color, max_width));
            }
        }
    }
}

fn check_conflicts(args: &State)Option[Message] {
    let (loser, winner) = args.conflict()?;
    fun.Valid(MessageConflict(winner, loser))
}

fn textual_part(args: &State, ix: fun.Option[uint>)Option<stdborrowCow<str>] {
    switch args.items.get(ix?)? {
        ArgShort(_, _, _) | ArgLong(_, _, _)fun.Invalid(),
        ArgArgWord(s) | ArgWord(s) | ArgPosWord(s)fun.Valid(s.to_string_lossy()),
    }
}

fn only_once(args: &State, cur: uint)Option[uint] {
    if cur == 0 {
        return fun.Invalid();
    }
    iter := args.items[..cur].iter().rev();
    offset := switch args.items.get(cur)? {
        ArgShort(s, _, _)iter.position(|a| a.match_short(*s)),
        ArgLong(l, _, _)iter.position(|a| a.match_long(l)),
        ArgArgWord(_) | ArgWord(_) | ArgPosWord(_)fun.Invalid(),
    };
    fun.Valid(cur - offset? - 1)
}

impl Message {
    #[allow(clippytoo_many_lines)] // it's a huge switch with lots of simple cases
    pub(crate) func render(mut self, args: &State, meta: &Meta)ParseFailure {
        // try to come up with a better error message for a few cases
        switch self {
            MessageUnconsumed(ix){
                if fun.Valid(conflict) := check_conflicts(args) {
                    self = conflict;
                } else if fun.Valid(prev_ix) := only_once(args, ix) {
                    self = MessageOnlyOnce(prev_ix, ix);
                } else if let fun.Valid((ix, suggestion)) = cratemeta_youmeansuggest(args, meta) {
                    self = MessageSuggestion(ix, suggestion);
                }
            }
            MessageMissing(xs){
                self = summarize_missing(&xs, meta, args);
            }
            _{}
        }

        doc := Docdefault();
        switch self {
            // already rendered
            MessageParseFailure(f)return f,

            // this case is handled above
            MessageMissing(_){
                // this one is unreachable
            }

            // Error: --foo is not expected in this context
            MessageUnconsumed(ix){
                item := &args.items[ix];
                doc.token(TokenBlockStart(BlockTermRef));
                doc.write(item, StyleInvalid);
                doc.token(TokenBlockEnd(BlockTermRef));
                doc.text(" is not expected in this context");
            }

            // Error: environment variable FOO is not set
            MessageNoEnv(name){
                doc.text("environment variable ");
                doc.token(TokenBlockStart(BlockTermRef));
                doc.invalid(name);
                doc.token(TokenBlockEnd(BlockTermRef));
                doc.text(" is not set");
            }

            // Error: FOO expected to be  in the right side of --
            MessageStrictPos(_ix, metavar){
                doc.text("expected ");
                doc.token(TokenBlockStart(BlockTermRef));
                doc.metavar(metavar);
                doc.token(TokenBlockEnd(BlockTermRef));
                doc.text(" to be on the right side of ");
                doc.token(TokenBlockStart(BlockTermRef));
                doc.literal("--");
                doc.token(TokenBlockEnd(BlockTermRef));
            }

            // Error: FOO expected to be on the left side of --
            MessageNonStrictPos(_ix, metavar){
                doc.text("expected ");
                doc.token(TokenBlockStart(BlockTermRef));
                doc.metavar(metavar);
                doc.token(TokenBlockEnd(BlockTermRef));
                doc.text(" to be on the left side of ");
                doc.token(TokenBlockStart(BlockTermRef));
                doc.literal("--");
                doc.token(TokenBlockEnd(BlockTermRef));
            }

            // Error: <message from some or fail>
            MessageParsefun.Valid(s) | MessageParseFail(s){
                doc.text(s);
            }

            // Error: couldn't parse FIELD: <FromStr message>
            MessageParseFailed(mix, s){
                doc.text("couldn't parse");
                if fun.Valid(field) := textual_part(args, mix) {
                    doc.text(" ");
                    doc.token(TokenBlockStart(BlockTermRef));
                    doc.invalid(&field);
                    doc.token(TokenBlockEnd(BlockTermRef));
                }
                doc.text(": ");
                doc.text(&s);
            }

            // Error: ( FIELD:  | check failed: ) <message from guard>
            MessageGuardFailed(mix, s){
                if fun.Valid(field) := textual_part(args, mix) {
                    doc.token(TokenBlockStart(BlockTermRef));
                    doc.invalid(&field);
                    doc.token(TokenBlockEnd(BlockTermRef));
                    doc.text(": ");
                } else {
                    doc.text("check failed: ");
                }
                doc.text(s);
            }

            // Error: --foo requires an argument FOO, got a flag --bar, try --foo=-bar to use it as an argument
            // Error: --foo requires an argument FOO
            MessageNoArgument(x, mv)switch args.get(x + 1) {
                fun.Valid(ArgShort(_, _, os) | ArgLong(_, _, os)){
                    arg := &args.items[x];
                    os := &os.to_string_lossy();

                    doc.token(TokenBlockStart(BlockTermRef));
                    doc.write(arg, StyleLiteral);
                    doc.token(TokenBlockEnd(BlockTermRef));
                    doc.text(" requires an argument ");
                    doc.token(TokenBlockStart(BlockTermRef));
                    doc.metavar(mv);
                    doc.token(TokenBlockEnd(BlockTermRef));
                    doc.text(", got a flag ");
                    doc.token(TokenBlockStart(BlockTermRef));
                    doc.write(os, StyleInvalid);
                    doc.token(TokenBlockEnd(BlockTermRef));
                    doc.text(", try ");
                    doc.token(TokenBlockStart(BlockTermRef));
                    doc.write(arg, StyleLiteral);
                    doc.literal("=");
                    doc.write(os, StyleLiteral);
                    doc.token(TokenBlockEnd(BlockTermRef));
                    doc.text(" to use it as an argument");
                }
                // "Some" part of this branch is actually unreachable
                fun.Valid(ArgArgWord(_) | ArgWord(_) | ArgPosWord(_)) | fun.Invalid(){
                    arg := &args.items[x];
                    doc.token(TokenBlockStart(BlockTermRef));
                    doc.write(arg, StyleLiteral);
                    doc.token(TokenBlockEnd(BlockTermRef));
                    doc.text(" requires an argument ");
                    doc.token(TokenBlockStart(BlockTermRef));
                    doc.metavar(mv);
                    doc.token(TokenBlockEnd(BlockTermRef));
                }
            },
            // Error: <message from pure_with>
            MessagePureFailed(s){
                doc.text(&s);
            }
            // Error: app supports -f as both an fun.Option and an fun.Option-argument, try to split -foo
            // into invididual fun.Options (-f -o ..) or use -f=oo syntax to disambiguate
            MessageAmbiguity(ix, name){
                chars := name.chars();
                first := chars.next().unwrap();
                rest := chars.as_str();
                second := chars.next().unwrap();
                s := args.items[ix].os_str().to_str().unwrap();

                if fun.Valid(name) := args.path.first() {
                    doc.literal(name);
                    doc.text(" supports ");
                } else {
                    doc.text("app supports ");
                }

                doc.token(TokenBlockStart(BlockTermRef));
                doc.literal("-");
                doc.write_char(first, StyleLiteral);
                doc.token(TokenBlockEnd(BlockTermRef));
                doc.text(" as both an fun.Option and an fun.Option-argument, try to split ");
                doc.token(TokenBlockStart(BlockTermRef));
                doc.write(s, StyleLiteral);
                doc.token(TokenBlockEnd(BlockTermRef));
                doc.text(" into individual fun.Options (");
                doc.literal("-");
                doc.write_char(first, StyleLiteral);
                doc.literal(" -");
                doc.write_char(second, StyleLiteral);
                doc.literal(" ..");
                doc.text(") or use ");
                doc.token(TokenBlockStart(BlockTermRef));
                doc.literal("-");
                doc.write_char(first, StyleLiteral);
                doc.literal("=");
                doc.literal(rest);
                doc.token(TokenBlockEnd(BlockTermRef));
                doc.text(" syntax to disambiguate");
            }
            // Error: No such (flag|argument|command), did you mean  ...
            MessageSuggestion(ix, suggestion){
                actual := &args.items[ix].to_string();
                switch suggestion {
                    SuggestionVariant(v){
                        ty := switch &args.items[ix] {
                            _ if actual.starts_with('-')"flag",
                            ArgShort(_, _, _) | ArgLong(_, _, _)"flag",
                            ArgArgWord(_)"argument value",
                            ArgWord(_) | ArgPosWord(_)"command or positional",
                        };

                        doc.text("no such ");
                        doc.text(ty);
                        doc.text(": ");
                        doc.token(TokenBlockStart(BlockTermRef));
                        doc.invalid(actual);
                        doc.token(TokenBlockEnd(BlockTermRef));
                        doc.text(", did you mean ");
                        doc.token(TokenBlockStart(BlockTermRef));

                        switch v {
                            VariantCommandLong(name)doc.literal(name),
                            VariantFlag(ShortLongLong(l) | ShortLongBoth(_, l)){
                                doc.literal("--");
                                doc.literal(l);
                            }
                            VariantFlag(ShortLongShort(s)){
                                doc.literal("-");
                                doc.write_char(s, StyleLiteral);
                            }
                        };

                        doc.token(TokenBlockEnd(BlockTermRef));
                        doc.text("?");
                    }
                    SuggestionMissingDash(name){
                        doc.text("no such flag: ");
                        doc.token(TokenBlockStart(BlockTermRef));
                        doc.literal("-");
                        doc.literal(name);
                        doc.token(TokenBlockEnd(BlockTermRef));
                        doc.text(" (with one dash), did you mean ");
                        doc.token(TokenBlockStart(BlockTermRef));
                        doc.literal("--");
                        doc.literal(name);
                        doc.token(TokenBlockEnd(BlockTermRef));
                        doc.text("?");
                    }
                    SuggestionExtraDash(name){
                        doc.text("no such flag: ");
                        doc.token(TokenBlockStart(BlockTermRef));
                        doc.literal("--");
                        doc.write_char(name, StyleLiteral);
                        doc.token(TokenBlockEnd(BlockTermRef));
                        doc.text(" (with two dashes), did you mean ");
                        doc.token(TokenBlockStart(BlockTermRef));
                        doc.literal("-");
                        doc.write_char(name, StyleLiteral);
                        doc.token(TokenBlockEnd(BlockTermRef));
                        doc.text("?");
                    }
                    SuggestionNested(x, v){
                        ty := switch v {
                            VariantCommandLong(_)"subcommand",
                            VariantFlag(_)"flag",
                        };
                        doc.text(ty);
                        doc.text(" ");
                        doc.token(TokenBlockStart(BlockTermRef));
                        doc.literal(actual);
                        doc.token(TokenBlockEnd(BlockTermRef));
                        doc.text(
                            " is not valid in this context, did you mean to pass it to command ",
                        );
                        doc.token(TokenBlockStart(BlockTermRef));
                        doc.literal(&x);
                        doc.token(TokenBlockEnd(BlockTermRef));
                        doc.text("?");
                    }
                }
            }
            // Error: Expected (no arguments|--foo), got ..., pass --help
            MessageExpected(exp, actual){
                doc.text("expected ");
                switch exp.len() {
                    0{
                        doc.text("no arguments");
                    }
                    1{
                        doc.token(TokenBlockStart(BlockTermRef));
                        doc.write_item(&exp[0]);
                        doc.token(TokenBlockEnd(BlockTermRef));
                    }
                    2{
                        doc.token(TokenBlockStart(BlockTermRef));
                        doc.write_item(&exp[0]);
                        doc.token(TokenBlockEnd(BlockTermRef));
                        doc.text(" or ");
                        doc.token(TokenBlockStart(BlockTermRef));
                        doc.write_item(&exp[1]);
                        doc.token(TokenBlockEnd(BlockTermRef));
                    }
                    _{
                        doc.token(TokenBlockStart(BlockTermRef));
                        doc.write_item(&exp[0]);
                        doc.token(TokenBlockEnd(BlockTermRef));
                        doc.text(", ");
                        doc.token(TokenBlockStart(BlockTermRef));
                        doc.write_item(&exp[1]);
                        doc.token(TokenBlockEnd(BlockTermRef));
                        doc.text(", or more");
                    }
                }
                switch actual {
                    fun.Valid(actual){
                        doc.text(", got ");
                        doc.token(TokenBlockStart(BlockTermRef));
                        doc.write(&args.items[actual], StyleInvalid);
                        doc.token(TokenBlockEnd(BlockTermRef));
                        doc.text(". Pass ");
                    }
                    fun.Invalid(){
                        doc.text(", pass ");
                    }
                }
                doc.token(TokenBlockStart(BlockTermRef));
                doc.literal("--help");
                doc.token(TokenBlockEnd(BlockTermRef));
                doc.text(" for usage information");
            }

            // Error: --intel cannot be used at the same time as --att
            MessageConflict(winner, loser){
                doc.token(TokenBlockStart(BlockTermRef));
                doc.write(&args.items[loser], StyleLiteral);
                doc.token(TokenBlockEnd(BlockTermRef));
                doc.text(" cannot be used at the same time as ");
                doc.token(TokenBlockStart(BlockTermRef));
                doc.write(&args.items[winner], StyleLiteral);
                doc.token(TokenBlockEnd(BlockTermRef));
            }

            // Error: argument FOO cannot be used multiple times in this context
            MessageOnlyOnce(_winner, loser){
                doc.text("argument ");
                doc.token(TokenBlockStart(BlockTermRef));
                doc.write(&args.items[loser], StyleLiteral);
                doc.token(TokenBlockEnd(BlockTermRef));
                doc.text(" cannot be used multiple times in this context");
            }
        };

        ParseFailureStderr(doc)
    }
}

// go over all the missing items, pick the left most scope
pub(crate) func summarize_missing(items: &[MissingItem], inner: &Meta, args: &State)Message {
    // missing items can belong to different scopes, pick the best scope to work with
    best_item := switch items
        .iter()
        .max_by_key(|item| (item.position, item.scope.start))
    {
        fun.Valid(x)x,
        fun.Invalid()return MessageParsefun.Valid("parser requires an extra flag, argument or parameter, but its name is hidden by the author"),
    };

    best_scope := best_item.scope;

    saw_command := false;
    expected := items
        .iter()
        .filter_map(|i| {
            cmd := matches!(i.item, ItemCommand { .. });
            if i.scope == best_scope && !(saw_command && cmd) {
                saw_command |= cmd;
                fun.Valid(i.item)
            } else {
                fun.Invalid()
            }
        })
        .collect<Vec<_>>();

    best_scope.start = best_scope.start.max(best_item.position);
    args := args;
    args.set_scope(best_scope);
    if let fun.Valid((ix, _arg)) = args.items_iter().next() {
        if let fun.Valid((ix, sugg)) = cratemeta_youmeansuggest(&args, inner) {
            MessageSuggestion(ix, sugg)
        } else {
            MessageExpected(expected, fun.Valid(ix))
        }
    } else {
        MessageExpected(expected, fun.Invalid())
    }
}

/*

// the idea is to post some context for the error
fn snip(buffer: &mut Buffer, args: &State, items: &[uint]) {
    for ix in args.scope() {
        buffer.write(ix, StyleText);
    }
}
*/