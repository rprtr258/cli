package cli

import (
	"io"
	stditer "iter"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"testing"

	"github.com/rprtr258/fun"
	funiter "github.com/rprtr258/fun/iter"
	"github.com/stretchr/testify/require"
)

type (
	PSOS = fun.Pair[string, Opt[string]]
	SSOS = stditer.Seq2[string, Opt[string]]
)

func test_completer(input string) SSOS {
	vec := test_completer_descr(input)
	return func(yield func(string, Opt[string]) bool) {
		for v := range vec {
			if !yield(v, Invalid[string]()) {
				return
			}
		}
	}
}

func test_completer2(input string) SSOS {
	items := []string{"auto", "mala"}

	return seq2(
		funiter.Map(
			funiter.
				FromMany(items...).
				Filter(func(item string) bool { return strings.HasPrefix(item, input) }),
			func(item string) PSOS { return PSOS{item, Invalid[string]()} }))
}

func test_completer_descr(input string) SSOS {
	items := []string{"alpha", "beta", "banana", "cat", "durian"}
	return seq2(
		funiter.Map(
			funiter.
				FromMany(items...).
				Filter(func(item string) bool { return strings.HasPrefix(item, input) }),
			func(item string) PSOS { return PSOS{item, Valid(item)} }))
}

func complete_alpha(input string) SSOS {
	//nolint:gocritic // candidate-first order is intentional: match when the
	// typed input is a prefix of the completion candidate.
	if strings.HasPrefix("alpha", input) {
		return seq2(funiter.FromMany(PSOS{"alpha", Valid("alpha description")}))
	} else {
		return seq2(funiter.FromNothing[PSOS]())
	}
}

func complete_beta(input string) SSOS {
	//nolint:gocritic // candidate-first order is intentional: match when the
	// typed input is a prefix of the completion candidate.
	if strings.HasPrefix("beta", input) {
		return seq2(funiter.FromMany(PSOS{"beta", Valid("beta description")}))
	} else {
		return seq2(funiter.FromNothing[PSOS]())
	}
}

func TestMixing_shell_and_positional_1_flag_or_pos(t *testing.T) {
	arg := ReqFlag(Short('b').Help("Option b"), uint(10))
	pos := Parse(Positional("DISTANCE", ParseString).
		Parser().CompleteShell(ShellCompFile(Invalid[string]())).
		Parser().Guard(func(s string) bool { return s != "" }, "unreachable"),
		ParseUint)

	assertCompletion(t, Choice(arg, pos).ToOptions(), "-b\t-b\t\tOption b\n\nFile { mask: None }\n", "")
}

func TestMixing_shell_and_positional_2_arg_or_pos(t *testing.T) {
	arg := Short('b').Help("Option b").Argument("HELLO", ParseUint)
	pos := Positional("DISTANCE", ParseUint).CompleteShell(ShellCompFile(Invalid[string]()))

	assertCompletion(t, Choice(arg, pos).ToOptions(), "-b\t-b=HELLO\t\tOption b\n\nFile { mask: None }\n", "")
}

func TestMixing_shell_and_positional_3_flag_and_pos(t *testing.T) {
	arg := ReqFlag(Short('b').Help("Option b"), uint(10))
	pos := Positional("DISTANCE", ParseUint).CompleteShell(ShellCompFile(Invalid[string]()))

	assertCompletion(t, Pair(arg, pos).ToOptions(), "-b\t-b\t\tOption b\n\nFile { mask: None }\n", "")
}

func TestMixing_shell_and_positional_4_arg_and_pos(t *testing.T) {
	arg := Short('b').Help("Option b").Argument("HELLO", ParseUint)
	pos := Positional("DISTANCE", ParseUint).CompleteShell(ShellCompFile(Invalid[string]()))

	assertCompletion(t, Pair(arg, pos).ToOptions(), "-b\t-b=HELLO\t\tOption b\n\nFile { mask: None }\n", "")
}

func TestStatic_complete_test_1(t *testing.T) {
	a := Short('a').Long("avocado").Help("Use avocado").Switch()
	b := Short('b').Long("banana").Help("Use banana").Switch()
	bb := Long("bananananana").Help("I'm Batman").Switch()
	c := Long("calculator").
		Help("calculator expression").Argument(
		"EXPR", ParseString)

	parser := Pair(Pair(a, b), Pair(bb, c)).ToOptions()

	t.Run("wtf", func(t *testing.T) {
		assertCompletion(t, parser, "--avocado\t--avocado\t\tUse avocado\n--banana\t--banana\t\tUse banana\n--bananananana\t--bananananana\t\tI'm Batman\n--calculator\t--calculator=EXPR\t\tcalculator expression\n\n", "--")
	})

	assertCompletion(t, parser, "--banana", "-b")

	// this used to be disambiguation, not anymore

	t.Run("wtf", func(t *testing.T) {
		assertCompletion(t, parser, "-vvvv\n", "-vvvv")
	})

	t.Run("wtf", func(t *testing.T) {
		assertCompletion(t, parser, "-v\n", "-v")
	})

	t.Run("wtf", func(t *testing.T) {
		assertCompletion(t, parser, "--banana\t--banana\t\tUse banana\n--bananananana\t--bananananana\t\tI'm Batman\n\n", "--b")
	})

	assertCompletion(t, parser, "--avocado", "--a")

	t.Run("wtf", func(t *testing.T) {
		assertCompletion(t, parser, "--banana\t--banana\t\tUse banana\n--bananananana\t--bananananana\t\tI'm Batman\n\n", "--banana")
	})

	assertCompletion(t, parser, "--bananananana", "--bananan")
}

func TestLong_and_short_arguments(t *testing.T) {
	parser := Short('p').
		Long("potato").Argument(
		"POTATO", ParseString).
		Parser().ToOptions()

	assertCompletion(t, parser, "--potato", "-p")

	assertCompletion(t, parser, "\tPOTATO\t\t\n\n", "-p", "")

	assertCompletion(t, parser, "\tPOTATO\t\t\n\n", "-p", "x")
}

func TestShort_command_alias(t *testing.T) {
	a := Long("potato").Argument(
		"A", ParseString).
		Parser().ToOptions().
		Command("cmd_a").
		Short('a')

	b := Long("potato").Argument(
		"A", ParseString).
		Parser().ToOptions().
		Command("cmd_b").
		Short('b')
	parser := Choice(a, b).ToOptions()

	assertCompletion(t, parser, "cmd_a", "a")

	assertCompletion(t, parser, "cmd_a", "cmd_a")

	t.Run("wtf", func(t *testing.T) {
		assertCompletion(t, parser, "--potato", "b", "")
	})
}

func TestSingle_command_completes_to_full(t *testing.T) {
	parser := Short('a').Switch().ToOptions().Command("cmd").ToOptions()

	assertCompletion(t, parser, "cmd", "c")

	assertCompletion(t, parser, "cmd", "cmd")
}

func TestStatic_complete_test_2(t *testing.T) {
	a := Long("potato").Argument(
		"SHAPE", ParseString).
		Parser().ToOptions().
		Command("check").
		Short('C').
		Help("check packages")

	b := Long("megapotato").Argument(
		"MEGA", ParseString).
		Parser().ToOptions().
		Command("clean").
		Help("clean target dir")

	c := Long("makan").Argument(
		"BKT", ParseString).
		Parser().ToOptions().
		Command("build").
		Short('b').
		Help("build project")

	parser := Choice(a, b, c).ToOptions()

	t.Run("wtf", func(t *testing.T) {
		assertCompletion(t, parser, "check\tcheck\t\tcheck packages\nclean\tclean\t\tclean target dir\n\n", "c")
	})

	t.Run("wtf", func(t *testing.T) {
		assertCompletion(t, parser, "--potato", "check", "")
	})

	assertCompletion(t, parser, "check", "check")

	assertCompletion(t, parser, "--potato", "C", "--p")

	t.Run("wtf", func(t *testing.T) {
		assertCompletion(t, parser, "x\n", "x")
	})

	t.Run("wtf", func(t *testing.T) {
		assertCompletion(t, parser, "check\tcheck\t\tcheck packages\nclean\tclean\t\tclean target dir\nbuild\tbuild\t\tbuild project\n\n", "")
	})

	assertCompletion(t, parser, "check", "ch")
}

func TestStatic_complete_test_3(t *testing.T) {
	a := Long("potato").Help("po").Argument("P", ParseString)
	b := Long("banana").Help("ba").Argument("B", ParseString)
	ab := Pair(a, b)
	c := Long("durian").Argument("D", ParseString)
	parser := Pair(ab, c).ToOptions()

	assertCompletion(t, parser, "--potato\t--potato=P\t\tpo\n--banana\t--banana=B\t\tba\n--durian\t--durian=D\t\t\n\n", "")

	assertCompletion(t, parser, "--potato\t--potato=P\t\tpo\n--banana\t--banana=B\t\tba\n--durian\t--durian=D\t\t\n\n", "-")

	assertCompletion(t, parser, "--potato\t--potato=P\t\tpo\n--banana\t--banana=B\t\tba\n--durian\t--durian=D\t\t\n\n", "--")

	assertCompletion(t, parser, "--durian", "--d")
}

func TestStatic_complete_test_4(t *testing.T) {
	a := Short('a').Argument("A", ParseString)
	b := Short('b').Argument("B", ParseString)
	parser := Pair(a, b).ToOptions()

	assertCompletion(t, parser, "\tA\t\t\n\n", "-a", "")

	t.Run("wtf", func(t *testing.T) {
		assertCompletion(t, parser, "-a", "-a")
	})

	t.Run("wtf", func(t *testing.T) {
		assertCompletion(t, parser, "-a\t-a=A\t\t\n-b\t-b=B\t\t\n\n", "")
	})

	t.Run("wtf", func(t *testing.T) {
		assertCompletion(t, parser, "-a\t-a=A\t\t\n-b\t-b=B\t\t\n\n", "-")
	})

	t.Run("wtf", func(t *testing.T) {
		assertCompletion(t, parser, "--\n", "--")
	})
}

func TestStatic_complete_test_5(t *testing.T) {
	a := Short('a').Argument("A", ParseString)
	b := Short('b').Argument("B", ParseString)
	c := Short('c').Argument("C", ParseString)
	d := Short('d').Argument("D", ParseString)
	ab := Pair(a, b)
	cd := Pair(c, d)
	parser := Pair(ab, cd).ToOptions()

	t.Run("wtf", func(t *testing.T) {
		assertCompletion(t, parser, "-b\t-b=B\t\t\n-c\t-c=C\t\t\n-d\t-d=D\t\t\n\n", "-a", "x", "")
	})

	assertCompletion(t, parser, "\tA\t\t\n\n", "-a", "")

	assertCompletion(t, parser, "-a", "-a")

	t.Run("wtf", func(t *testing.T) {
		assertCompletion(t, parser, "-a\t-a=A\t\t\n-b\t-b=B\t\t\n-c\t-c=C\t\t\n-d\t-d=D\t\t\n\n", "")
	})

	assertCompletion(t, parser, `-a	-a=A`+"\t\t"+`
-b	-b=B`+"\t\t"+`
-c	-c=C`+"\t\t"+`
-d	-d=D`+"\t\t"+`

`, "-")

	t.Run("wtf", func(t *testing.T) {
		assertCompletion(t, parser, "--\n", "--")
	})
}

func TestStatic_complete_test_6(t *testing.T) {
	a := Short('a').Argument("A", ParseString).Parser().Optional()
	b := Short('b').Argument("B", ParseString).Parser().Many()
	parser := Pair(a, b).ToOptions()

	assertCompletion(t, parser, "-a\t-a=A\t\t\n-b\t-b=B\t\t\n\n", "-b", "x", "")

	assertCompletion(t, parser, "\tA\t\t\n\n", "-a", "")

	assertCompletion(t, parser, "-a\t-a=A\t\t\n-b\t-b=B\t\t\n\n", "")

	assertCompletion(t, parser, "-b", "-a", "x", "")

	assertCompletion(t, parser, "-a", "-a")

	assertCompletion(t, parser, "-a\t-a=A\t\t\n-b\t-b=B\t\t\n\n", "")

	assertCompletion(t, parser, "\tB\t\t\n\n", "-b", "")
}

func TestStatic_complete_test_7(t *testing.T) {
	a := Short('a').Help("switch").Switch()
	b := Positional("FILE", ParseString).Help("File to use")
	parser := Pair(a, b).ToOptions()

	assertCompletion(t, parser, "-a\t-a\t\tswitch\n\tFILE\t\tFile to use\n\n", "")

	assertCompletion(t, parser, "\tFILE\t\tFile to use\n\n", "-a", "")

	assertCompletion(t, parser, "\tFILE\t\tFile to use\n\n", "-a", "x")
}

func TestStatic_complete_test_8(t *testing.T) {
	parser := Short('a').
		Long("durian").
		Switch().
		Parser().ToOptions().
		Command("nom").
		Parser().ToOptions()

	assertCompletion(t, parser, "nom", "")

	assertCompletion(t, parser, "--durian", "nom", "")

	assertCompletion(t, parser, "--durian", "nom", "-a")

	assertCompletion(t, parser, "\n", "nom", "-a", "")
}

func TestJust_positional(t *testing.T) {
	parser := Positional("FILE", ParseString).
		Help("File to use").
		Parser().ToOptions()

	assertCompletion(t, parser, "\tFILE\t\tFile to use\n\n", "")

	assertCompletion(t, parser, "\tFILE\t\tFile to use\n\n", "xxx")
}

func TestDynamic_complete_test_1(t *testing.T) {
	parser := Short('a').Argument(
		"ARG", ParseString).
		Parser().Complete(test_completer).
		Parser().ToOptions()

	assertCompletion(t, parser, "\tARG\t\t\nbeta\tbeta\t\t\nbanana\tbanana\t\t\n\n", "-a", "b")

	assertCompletion(t, parser, "beta", "-a", "be")

	assertCompletion(t, parser, "beta", "-a", "beta")

	assertCompletion(t, parser, "-a", "")

	assertCompletion(t, parser, "\tARG\t\t\nalpha\talpha\t\t\nbeta\tbeta\t\t\nbanana\tbanana\t\t\ncat\tcat\t\t\ndurian\tdurian\t\t\n\n", "-a", "")
}

func TestDynamic_complete_test_2(t *testing.T) {
	parser := Short('a').Argument("ARG", ParseString).ToOptions()

	assertCompletion(t, parser, "\tARG\t\t\n\n", "-a", "b")
}

func TestDynamic_complete_test_3(t *testing.T) {
	a := Short('a').Long("avocado").Help("Use avocado").Switch()
	b := Short('b').Long("banana").Help("Use banana").Switch()
	bb := Long("bananananana").Help("I'm Batman").Switch()
	c := Long("calculator").
		Help("calculator expression").Argument(
		"EXPR", ParseString).
		Parser().Complete(test_completer)
	parser := Pair(Pair(a, b), Pair(bb, c)).ToOptions()

	assertCompletion(t, parser, "\tEXPR\t\tcalculator expression\nalpha\talpha\t\t\nbeta\tbeta\t\t\nbanana\tbanana\t\t\ncat\tcat\t\t\ndurian\tdurian\t\t\n\n", "--calculator", "")
}

func TestDynamic_complete_test_4(t *testing.T) {
	parser := Long("name").Argument(
		"NAME", ParseString).
		Parser().Complete(test_completer_descr).
		Parser().ToOptions()

	assertCompletion(t, parser, "\tNAME\t\t\nalpha\talpha\t\talpha\nbeta\tbeta\t\tbeta\nbanana\tbanana\t\tbanana\ncat\tcat\t\tcat\ndurian\tdurian\t\tdurian\n\n", "--name", "")

	assertCompletion(t, parser, "alpha", "--name", "a")
}

func TestStatic_with_hide(t *testing.T) {
	a := Short('a').Switch()
	b := Short('b').Switch().Hide()
	parser := Pair(a, b).ToOptions()

	assertCompletion(t, parser, "-a", "")
}

func TestStatic_with_fallback_and_hide(t *testing.T) {
	a := Short('a').Switch()
	b := Short('b').Switch().Hide()
	parser := Pair(a, b).Fallback(fun.Pair[bool, bool]{false, false}).ToOptions()

	assertCompletion(t, parser, "-a", "")
}

func TestCsample_mystery(t *testing.T) {
	a := Short('a').Long("avocado").Help("Use avocado").Switch()
	b := Short('b').Long("banana").Help("Use banana").Switch()
	bb := Long("bananananana").Help("I'm Batman").Switch()
	c := Long("calculator").
		Help("calculator expression").Argument(
		"EXPR", ParseString).
		Parser().Complete(test_completer)
	parser := Pair(Pair(a, b), Pair(bb, c)).
		Parser().ToOptions().
		Description("Dynamic autocomplete example").
		Footer("footer")

	assertCompletion(t, parser, "--avocado\t--avocado\t\tUse avocado\n--banana\t--banana\t\tUse banana\n--bananananana\t--bananananana\t\tI'm Batman\n--calculator\t--calculator=EXPR\t\tcalculator expression\n\n", "")
}

func TestOnly_positionals_after_double_dash(t *testing.T) {
	a := Short('a').Switch()
	b := Short('b').Switch()
	c := Short('c').Switch()
	d := Positional("D", ParseString)
	parser := Pair(Pair(a, b), Pair(c, d)).ToOptions()

	assertCompletion(t, parser, "--\n", "-a", "--")

	assertCompletion(t, parser, "-a", "-a")

	assertCompletion(t, parser, "-b\t-b\t\t\n-c\t-c\t\t\n\tD\t\t\n\n", "-a", "")

	assertCompletion(t, parser, "\tD\t\t\n\n", "--", "")
}

func TestMany_does_not_duplicate_metadata(t *testing.T) {
	parser := Positional("D", ParseString).Parser().Many().ToOptions()
	assertCompletion(t, parser, "\tD\t\t\n\n", "xxx")
}

func TestSome_does_not_duplicate_metadata(t *testing.T) {
	parser := Positional("D", ParseString).Parser().Some("").ToOptions()
	assertCompletion(t, parser, "\tD\t\t\n\n", "xxx")
}

func TestOnly_positionals_after_positionals(t *testing.T) {
	a := Short('a').Switch()
	d := Positional("D", ParseString).Parser().Many()
	parser := Pair(a, d).ToOptions()

	assertCompletion(t, parser, "-a\t-a\t\t\n\tD\t\t\n\n", "")

	assertCompletion(t, parser, "\tD\t\t\n\n", "xxx")

	assertCompletion(t, parser, "\tD\t\t\n\n", "xxx", "yyy")

	assertCompletion(t, parser, "-a\t-a\t\t\n\tD\t\t\n\n", "xxx", "")
}

func TestPositionals_complete_in_order(t *testing.T) {
	a := Positional("A", ParseString).Complete(complete_alpha)
	b := Positional("B", ParseString).Complete(complete_beta)
	parser := Pair(a, b).ToOptions()

	assertCompletion(t, parser, "alpha", "")

	assertCompletion(t, parser, "alpha", "a")

	assertCompletion(t, parser, "\tA\t\t\n\n", "x")

	assertCompletion(t, parser, "beta", "xxx", "")

	assertCompletion(t, parser, "beta", "xxx", "b")

	assertCompletion(t, parser, "\tB\t\t\n\n", "xxx", "yyy")
}

func TestShould_be_able_to_suggest_positional_along_with_non_positionals_flags(t *testing.T) {
	a := Short('a').Argument("A", ParseString).Complete(complete_alpha)
	b := Positional("B", ParseString).Complete(complete_beta)
	parser := Pair(a, b).ToOptions()

	assertCompletion(t, parser, "-a\t-a=A\t\t\nbeta\tbeta\t\tbeta description\n\n", "")
}

func TestShould_be_able_to_suggest_double_dash(t *testing.T) {
	c_b := func(string) SSOS {
		return seq2(funiter.FromMany(PSOS{"--", Invalid[string]()}))
	}
	a := Long("arg").Argument(
		"ARG", ParseString).
		Parser().Complete(c_b).
		Parser().
		Optional()

	parser := a.ToOptions()

	assertCompletion(t, parser, "--arg", "")

	assertCompletion(t, parser, "--", "--arg", "")
}

func TestNon_strict_and_double_dash(t *testing.T) {
	a := Short('a').Switch()
	b := Positional("B", ParseString).NonStrict()
	parser := Pair(a, b).ToOptions()

	assertCompletion(t, parser, "-a\t-a\t\t\n\tB\t\t\n\n", "")

	assertCompletion(t, parser, "\n", "--", "")
}

func TestSuggest_double_dash_automatically_for_strictly_positional(t *testing.T) {
	a := Short('a').Switch()
	b := Positional("B", ParseString).Strict()
	parser := Pair(a, b).ToOptions()

	assertCompletion(t, parser, "-a\t-a\t\t\n--\t--\t\tPositional only items after this token\n\n", "")

	assertCompletion(t, parser, "-a\t-a\t\t\n--\t--\t\tPositional only items after this token\n\n", "-")

	assertCompletion(t, parser, "--\n", "--")

	assertCompletion(t, parser, "\tB\t\t\n\n", "--", "")
}

func TestMultiple_adjacent_single_letter_vals(t *testing.T) {
	// but not really - test is incomplete
	a := Short('a').Switch()
	b := Short('b').Switch()
	c := Short('c').Switch()
	parser := Pair(Pair(a, b), c).ToOptions()

	// with no input the right behavior is to suggest all the switches
	assertCompletion(t, parser, "-a\t-a\t\t\n-b\t-b\t\t\n-c\t-c\t\t\n\n", "")

	// with a single item present separately we should suggest the remaining two
	assertCompletion(t, parser, "-b\t-b\t\t\n-c\t-c\t\t\n\n", "-a", "")

	// trying to complete a pair of flags should dump current state
	assertCompletion(t, parser, "-ab\n", "-ab")

	// with a single valid item we should suggest it
	assertCompletion(t, parser, "-c", "-ab", "")

	// should I change the representation for the completions? Knowing that somethnig is a
	// flag with a short or long name is no longer necessary
	// now I need to be able to represent adding a -v to -f to get -fv, so pretty printed value is
	// -v, substitution is `-fv`...
	//
	// FWIW bash doesn't complete short flags in combination... should we behave the same way?
	//
	// This whole test is incomplete. We are not actually testing or even supporting
}

func TestAmbiguity_no_resolve(t *testing.T) {
	a0 := Short('a').Switch().Count()
	a1 := Short('a').Argument("AAAAAA", ParseInt)
	parser := Choice(a0, a1).ToOptions()

	assertCompletion(t, parser, "-aaa\n", "-aaa")
}

func TestAmbiguity_to_flags(t *testing.T) {
	parser := Short('a').Switch().Many().ToOptions()

	assertCompletion(t, parser, "-aaa\n", "-aaa")
}

func TestShort_argument_variants(t *testing.T) {
	parser := Short('a').Argument("META", ParseString).ToOptions()
	assertCompletion(t, parser, "\tMETA\t\t\n\n", "-a=aa")

	assertCompletion(t, parser, "\tMETA\t\t\n\n", "-a", "aa")

	assertCompletion(t, parser, "\tMETA\t\t\n\n", "-aaa")
}

func TestLong_argument_variants(t *testing.T) {
	parser := Long("alpha").Argument(
		"META", ParseString).
		Parser().Complete(complete_beta).
		Parser().ToOptions()

	assertCompletion(t, parser, "--alpha=beta", "--alpha=beta")

	assertCompletion(t, parser, "\tMETA\t\t\n\n", "--alpha", "Regina")
}

func TestZsh_style_completion_visible(t *testing.T) {
	a := Short('a').
		Long("argument").
		Help("this is an argument").Argument(
		"ARG", ParseString)
	b := Short('b').Argument("BANANA", ParseString)
	parser := Pair(a, b).GroupHelp("items").ToOptions()

	assertCompletion(t, parser, "--argument\t--argument=ARG\titems\tthis is an argument\n-b\t-b=BANANA\titems\t\n\n", "")
}

func TestZsh_many_positionals(t *testing.T) {
	parser := Positional("POS", ParseString).Parser().Many().ToOptions()

	assertCompletion(t, parser, "\tPOS\t\t\n\n", "")

	assertCompletion(t, parser, "\tPOS\t\t\n\n", "p")
}

func TestZsh_help_single_line_only(t *testing.T) {
	a := Short('a').Help("hello world").Argument("X", ParseString)
	b := Short('b').Help("hello from switch").Switch()
	parser := Pair(a, b).ToOptions()

	assertCompletion(t, parser, "-a\t-a=X\t\thello world\n-b\t-b\t\thello from switch\n\n", "")
}

func TestShell_help_single_line_only(t *testing.T) {
	a := Short('a').Help("hello 1\n\nworld").Argument("X", ParseString)
	b := Short('b').Help("hello 2\n\nworld").Argument("Y", ParseString)
	parser := Pair(a, b).ToOptions()

	assertCompletion(t, parser, "-a\t-a=X\t\thello 1\n-b\t-b=Y\t\thello 2\n\n", "")
}

func TestZsh_complete_info(t *testing.T) {
	foo := func(string) SSOS {
		return seq2(funiter.FromMany([]PSOS{{"hello", Valid("word")}, {"sample", Invalid[string]()}}...))
	}
	parser := Short('a').Argument(
		"X", ParseString).
		Parser().Complete(foo).
		Parser().ToOptions()

	assertCompletion(t, parser, "-a", "")

	assertCompletion(t, parser, "-a", "-")

	assertCompletion(t, parser, "\tX\t\t\nhello\thello\t\tword\nsample\tsample\t\t\n\n", "-a", "")
}

func TestPair_of_positionals(t *testing.T) {
	// with Positional items only current item should make suggestions, not both...
	a := Positional("A", ParseString).Complete(test_completer)
	b := Positional("B", ParseString).Complete(test_completer2)
	parser := Pair(a, b).ToOptions()

	assertCompletion(t, parser, "alpha", "a")

	assertCompletion(t, parser, "auto", "alpha", "a")
}

func TestDouble_dash_as_positional(t *testing.T) {
	parser := Positional("P", ParseString).
		Help("Help").
		Parser().Complete(test_completer).
		Parser().ToOptions()

	assertCompletion(t, parser, "alpha", "--", "a")

	assertCompletion(t, parser, "alpha", "a")

	assertCompletion(t, parser, "-\n", "-")

	assertCompletion(t, parser, "--\n", "--")

	assertCompletion(t, parser, "--\n", "--")

	assertCompletion(t, parser, "\tP\t\tHelp\n\n", "x")
}

func TestStrict_positional_completion(t *testing.T) {
	a := Long("arg").Switch()
	p := Positional("S", ParseString).
		Strict().
		Parser().Complete(func(string) SSOS { return seq2(funiter.FromMany(PSOS{"--hello", Invalid[string]()})) })
	parser := Pair(a, p).ToOptions()

	assertCompletion(t, parser, "--arg\t--arg\t\t\n--hello\t--hello\t\t\n\n", "--")

	assertCompletion(t, parser, "--arg", "--a")

	assertCompletion(t, parser, "--hello", "--", "")

	assertCompletion(t, parser, "--hello", "--", "--h")
}

func TestAvoid_inserting_metavars(t *testing.T) {
	parser := Short('a').Argument("A", ParseString).ToOptions()

	assertCompletion(t, parser, "-a", "")

	assertCompletion(t, parser, "\tA\t\t\n\n", "-a", "")
}

func TestShell_dir_completion(t *testing.T) {
	parser := Short('a').Argument(
		"FILE", ParseString).
		Parser().CompleteShell(ShellCompDir(Invalid[string]())).
		Parser().ToOptions()

	assertCompletion(t, parser, "\nDir { mask: None }\n", "-a", "")
}

func TestGenerate_unparseable_items(t *testing.T) {
	one := Pure(struct{}{}).ToOptions().Command("cone")
	two := Pure(struct{}{}).ToOptions().Command("ctwo")
	e := Short('e').Switch()

	one_e := Map2(e, one, func(_ bool, _ struct{}) struct{} { return struct{}{} })
	parser := Choice(one_e, two).ToOptions()

	// passing -e restricts branch with cmd_two
	assertCompletion(t, parser, "cone", "-e", "")

	// passing -e restricts branch with cmd_two
	assertCompletion(t, parser, "cone", "-e", "c")

	assertCompletion(t, parser, "-e\t-e\t\t\ncone\tcone\t\t\nctwo\tctwo\t\t\n\n", "")
}

func TestComplete_with_fallback(t *testing.T) {
	parser := Parse(Long("name").Argument(
		"NAME", ParseString).
		Parser().Complete(test_completer_descr),
		func(x string) (uint16, error) {
			u, err := strconv.ParseUint(x, 10, 16)
			return uint16(u), err
		}).
		Parser().Fallback(10).
		Parser().ToOptions()

	assertCompletion(t, parser, `	NAME`+"\t\t"+`
alpha	alpha		alpha
beta	beta		beta
banana	banana		banana
cat	cat		cat
durian	durian		durian

`, "--name", "")
}

func TestMix_of_options_and_positional_completions(t *testing.T) {
	a := Short('a').
		Long("arg").
		Help("Alhpa argument").Argument(
		"ALPHA", ParseString).
		Parser().Complete(complete_alpha)
	b := Positional("BETA", ParseString).
		Help("Beta argument").
		Parser().Complete(complete_beta)
	parser := Pair(a, b).ToOptions()

	assertCompletion(t, parser, "--arg\t--arg=ALPHA\t\tAlhpa argument\nbeta\tbeta\t\tbeta description\n\n", "")
}

func TestPositionals_with_no_completions_are_not_duplicated(t *testing.T) {
	a := Short('a').
		Long("arg").
		Help("Alhpa argument").Argument(
		"ALPHA", ParseString)
	b := Positional("BETA", ParseString).Help("Beta argument")

	parser := Pair(a, b).ToOptions()

	assertCompletion(t, parser, "--arg\t--arg=ALPHA\t\tAlhpa argument\n\tBETA\t\tBeta argument\n\n", "")
}

func captureStdout(t *testing.T, f func()) string {
	t.Helper()

	r, w, err := os.Pipe()
	require.NoError(t, err)
	orig := os.Stdout
	os.Stdout = w
	defer func() { os.Stdout = orig }()
	f()
	require.NoError(t, w.Close())
	data, err := io.ReadAll(r)
	require.NoError(t, err)
	return string(data)
}

func TestDump_completer_scripts(t *testing.T) {
	const name = "app"

	tests := []struct {
		script   string
		expected string
		dump     func(string)
	}{
		{
			script: "bash",
			dump:   dump_bash_completer,
			expected: `_cli_dynamic_completion()
{
    line="$1 --cli-complete-rev=8 ${COMP_WORDS[@]:1}"
    if [[ ${COMP_WORDS[-1]} == "" ]]; then
        line="${line} \"\""
    fi
    source <( eval ${line})
}
complete -o nosort -F _cli_dynamic_completion app
`,
		},
		{
			script: "zsh",
			dump:   dump_zsh_completer,
			expected: `#compdef app
local line
line="${words[1]} --cli-complete-rev=7 ${words[@]:1}"
if [[ ${words[-1]} == "" ]]; then
    line="${line} \"\""
fi
source <(eval ${line})
`,
		},
		{
			script: "fish",
			dump:   dump_fish_completer,
			expected: `set -l current (commandline --tokenize --current-process)
set -l tmpline $current[1] --cli-complete-rev=9 $current[2..]
if test (commandline --current-process) != (string trim (commandline --current-process))
    set tmpline $tmpline ""
end
source ( $tmpline | psub )
`,
		},
		{
			script: "elvish",
			dump:   dump_elvish_completer,
			expected: `set edit:completion:arg-completer[app] = { |@args| var args = $args[1..];
     var @lines = ( app --cli-complete-rev=1 $@args );
     use str;
     for line $lines {
         var @arg = (str:split "\t" $line)
         try {
             edit:complex-candidate $arg[0] &display=( printf "%-19s %s" $arg[0] $arg[1] )
         } catch {
             edit:complex-candidate $line
         }
     }
}
`,
		},
	}

	for _, tt := range tests {
		t.Run(t.Name()+"_"+tt.script, func(t *testing.T) {
			script := captureStdout(t, func() { tt.dump(name) })
			require.Equal(t, tt.expected, script)
		})
	}
}

// TestDump_scripts_syntax parses the generated scripts with the shell
// interpreters available on the machine, skipping the absent ones. The dump
// functions are not exercised anywhere else, so a syntax regression would go
// unnoticed until a user runs the stub.
func TestDump_scripts_syntax(t *testing.T) {
	checks := []struct {
		script   string
		binary   string
		args     []string
		generate func() string
	}{
		{script: "bash", binary: "bash", args: []string{"-n"}, generate: func() string {
			return captureStdout(t, func() { dump_bash_completer("app") })
		}},
		{script: "zsh", binary: "zsh", args: []string{"-n"}, generate: func() string {
			return captureStdout(t, func() { dump_zsh_completer("app") })
		}},
		{script: "fish", binary: "fish", args: []string{"--no-execute"}, generate: func() string {
			return captureStdout(t, func() { dump_fish_completer("app") })
		}},
	}

	for _, check := range checks {
		t.Run(t.Name()+"_"+check.script, func(t *testing.T) {
			if _, err := exec.LookPath(check.binary); err != nil {
				t.Skipf("%s is not installed", check.binary)
			}
			script := check.generate()
			cmd := exec.Command(check.binary, check.args...)
			cmd.Stdin = strings.NewReader(script)
			out, err := cmd.CombinedOutput()
			require.NoError(t, err, "%s -n failed:\n%s", check.binary, out)
		})
	}
}
