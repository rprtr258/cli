package cli_test

import (
	"iter"
	stditer "iter"
	"strconv"
	"strings"
	"testing"

	"github.com/rprtr258/fun"
	funiter "github.com/rprtr258/fun/iter"
	"github.com/stretchr/testify/assert"

	. "github.com/rprtr258/cli"
)

func zalupa12[K, V any](seq funiter.Seq[fun.Pair[K, V]]) iter.Seq2[K, V] {
	return func(yield func(K, V) bool) {
		for v := range seq {
			if !yield(v.K, v.V) {
				return
			}
		}
	}
}

type PSOS = fun.Pair[string, Option[string]]
type SSOS = stditer.Seq2[string, Option[string]]

func test_completer(input string) SSOS {
	vec := test_completer_descr(input)
	return func(yield func(string, Option[string]) bool) {
		for v := range vec {
			if !yield(v, Invalid[string]()) {
				return
			}
		}
	}
}

func test_completer2(input string) SSOS {
	items := []string{"auto", "mala"}

	return zalupa12(
		funiter.Map(
			funiter.
				FromMany(items...).
				Filter(func(item string) bool { return strings.HasPrefix(item, input) }),
			func(item string) PSOS { return PSOS{item, Invalid[string]()} }))
}

func test_completer_descr(input string) SSOS {
	items := []string{"alpha", "beta", "banana", "cat", "durian"}
	return zalupa12(
		funiter.Map(
			funiter.
				FromMany(items...).
				Filter(func(item string) bool { return strings.HasPrefix(item, input) }),
			func(item string) PSOS { return PSOS{item, Valid(item)} }))
}

func complete_alpha(input string) SSOS {
	if strings.HasPrefix("alpha", input) {
		return zalupa12(funiter.FromMany(PSOS{"alpha", Valid("alpha description")}))
	} else {
		return zalupa12(funiter.FromNothing[PSOS]())
	}
}

func complete_beta(input string) SSOS {
	if strings.HasPrefix("beta", input) {
		return zalupa12(funiter.FromMany(PSOS{"beta", Valid("beta description")}))
	} else {
		return zalupa12(funiter.FromNothing[PSOS]())
	}
}

func TestMixing_shell_and_positional_1_flag_or_pos(t *testing.T) {
	arg := ReqFlag(Short('b').Help("Option b"), uint(10))
	pos := Parse(Positional("DISTANCE", parseString).
		Parser().CompleteShell(ShellCompFile(Invalid[string]())).
		Parser().Guard(func(s string) bool { return s != "" }, "unreachable").Parser(),
		parseUint).Parser()

	_, err := Choice(arg, pos).
		Parser().ToOptions().
		RunInner(NewArgs("").SetComp(0))
	r := UnwrapStdout(err)
	assert.Equal(t, "-b\t-b\t\tOption b\n\nFile { mask: None }\n", r)
}

func TestMixing_shell_and_positional_2_arg_or_pos(t *testing.T) {
	arg := Argument(Short('b').Help("Option b"), "HELLO", parseUint).Parser()
	pos := Positional("DISTANCE", parseUint).Parser().CompleteShell(ShellCompFile(Invalid[string]())).Parser()

	_, err := Choice(arg, pos).
		ToOptions().
		RunInner(NewArgs("").SetComp(0))
	r := UnwrapStdout(err)
	assert.Equal(t, r, "-b\t-b=HELLO\t\tOption b\n\nFile { mask: None }\n")
}

func TestMixing_shell_and_positional_3_flag_and_pos(t *testing.T) {
	arg := ReqFlag(Short('b').Help("Option b"), uint(10)).Parser()
	pos := Positional("DISTANCE", parseUint).Parser().CompleteShell(ShellCompFile(Invalid[string]())).Parser()

	_, err := Pair(arg, pos).
		Parser().ToOptions().
		RunInner(NewArgs("").SetComp(0))
	r1 := UnwrapStdout(err)
	assert.Equal(t, r1, "-b\t-b\t\tOption b\n\nFile { mask: None }\n")
}

func TestMixing_shell_and_positional_4_arg_and_pos(t *testing.T) {
	arg := Argument(Short('b').Help("Option b"), "HELLO", parseUint).Parser()
	pos := Positional("DISTANCE", parseUint).Parser().CompleteShell(ShellCompFile(Invalid[string]())).Parser()

	_, err := Pair(arg, pos).
		Parser().ToOptions().
		RunInner(NewArgs("").SetComp(0))
	r := UnwrapStdout(err)
	assert.Equal(t, r, "-b\t-b=HELLO\t\tOption b\n\nFile { mask: None }\n")
}

func TestStatic_complete_test_1(t *testing.T) {
	a := Short('a').Long("avocado").Help("Use avocado").Switch()
	b := Short('b').Long("banana").Help("Use banana").Switch()
	bb := Long("bananananana").Help("I'm Batman").Switch()
	c := Argument(Long("calculator").
		Help("calculator expression"),
		"EXPR", parseString)

	parser := Pair(Pair(a, b), Pair(bb, c)).Parser().ToOptions()

	t.Run("wtf", func(t *testing.T) {
		_, err := parser.RunInner(NewArgs("--").SetComp(0))
		r := UnwrapStdout(err)
		assert.Equal(t, r, "--avocado\t--avocado\t\tUse avocado\n--banana\t--banana\t\tUse banana\n--bananananana\t--bananananana\t\tI'm Batman\n--calculator\t--calculator=EXPR\t\tcalculator expression\n\n")
	})

	{
		_, err := parser.RunInner(NewArgs("-b").SetComp(0))
		r := UnwrapStdout(err)
		assert.Equal(t, r, "--banana")
	}

	// this used to be disambiguation, not anymore

	t.Run("wtf", func(t *testing.T) {
		_, err := parser.RunInner(NewArgs("-vvvv").SetComp(0))
		r := UnwrapStdout(err)
		assert.Equal(t, r, "-vvvv\n")
	})

	t.Run("wtf", func(t *testing.T) {
		_, err := parser.RunInner(NewArgs("-v").SetComp(0))
		r := UnwrapStdout(err)
		assert.Equal(t, r, "-v\n")
	})

	t.Run("wtf", func(t *testing.T) {
		_, err := parser.RunInner(NewArgs("--b").SetComp(0))
		r := UnwrapStdout(err)
		assert.Equal(t, "--banana\t--banana\t\tUse banana\n--bananananana\t--bananananana\t\tI'm Batman\n\n", r)
	})

	{
		_, err := parser.RunInner(NewArgs("--a").SetComp(0))
		r := UnwrapStdout(err)
		assert.Equal(t, r, "--avocado")
	}

	t.Run("wtf", func(t *testing.T) {
		_, err := parser.RunInner(NewArgs("--banana").SetComp(0))
		r := UnwrapStdout(err)
		assert.Equal(t, r, "--banana\t--banana\t\tUse banana\n--bananananana\t--bananananana\t\tI'm Batman\n\n")
	})

	{
		_, err := parser.RunInner(NewArgs("--bananan").SetComp(0))
		r := UnwrapStdout(err)
		assert.Equal(t, r, "--bananananana")
	}
}

func TestLong_and_short_arguments(t *testing.T) {
	parser := Argument(Short('p').
		Long("potato"),
		"POTATO", parseString).
		Parser().ToOptions()

	{
		_, err := parser.RunInner(NewArgs("-p").SetComp(0))
		r := UnwrapStdout(err)
		assert.Equal(t, r, "--potato")
	}

	{
		_, err := parser.RunInner(NewArgs("-p", "").SetComp(0))
		r := UnwrapStdout(err)
		assert.Equal(t, r, "\tPOTATO\t\t\n\n")
	}

	{
		_, err := parser.RunInner(NewArgs("-p", "x").SetComp(0))
		r := UnwrapStdout(err)
		assert.Equal(t, r, "\tPOTATO\t\t\n\n")
	}
}

func TestShort_command_alias(t *testing.T) {
	a := Argument(Long("potato"),
		"A", parseString).
		Parser().ToOptions().
		Command("cmd_a").
		Short('a')

	b := Argument(Long("potato"),
		"A", parseString).
		Parser().ToOptions().
		Command("cmd_b").
		Short('b')
	parser := Choice(a, b).ToOptions()

	{
		_, err := parser.RunInner(NewArgs("a").SetComp(0))
		r := UnwrapStdout(err)
		assert.Equal(t, r, "cmd_a")
	}

	{
		_, err := parser.RunInner(NewArgs("cmd_a").SetComp(0))
		r := UnwrapStdout(err)
		assert.Equal(t, r, "cmd_a")
	}

	t.Run("wtf", func(t *testing.T) {
		_, err := parser.RunInner(NewArgs("b", "").SetComp(0))
		r := UnwrapStdout(err)
		assert.Equal(t, r, "--potato")
	})
}

func TestSingle_command_completes_to_full(t *testing.T) {
	parser := Short('a').Switch().Parser().ToOptions().Command("cmd").Parser().ToOptions()

	{
		_, err := parser.RunInner(NewArgs("c").SetComp(0))
		r := UnwrapStdout(err)
		assert.Equal(t, r, "cmd")
	}

	{
		_, err := parser.RunInner(NewArgs("cmd").SetComp(0))
		r := UnwrapStdout(err)
		assert.Equal(t, r, "cmd")
	}
}

func TestStatic_complete_test_2(t *testing.T) {
	a := Argument(Long("potato"),
		"SHAPE", parseString).
		Parser().ToOptions().
		Command("check").
		Short('C').
		Help("check packages")

	b := Argument(Long("megapotato"),
		"MEGA", parseString).
		Parser().ToOptions().
		Command("clean").
		Help("clean target dir")

	c := Argument(Long("makan"),
		"BKT", parseString).
		Parser().ToOptions().
		Command("build").
		Short('b').
		Help("build project")

	parser := Choice(a, b, c).ToOptions()

	t.Run("wtf", func(t *testing.T) {
		_, err := parser.RunInner(NewArgs("c").SetComp(0))
		r := UnwrapStdout(err)
		assert.Equal(t, "check\tcheck\t\tcheck packages\nclean\tclean\t\tclean target dir\n\n", r)
	})

	t.Run("wtf", func(t *testing.T) {
		_, err := parser.RunInner(NewArgs("check", "").SetComp(0))
		r := UnwrapStdout(err)
		assert.Equal(t, r, "--potato")
	})

	{
		_, err := parser.RunInner(NewArgs("check").SetComp(0))
		r := UnwrapStdout(err)
		assert.Equal(t, r, "check")
	}

	{
		_, err := parser.RunInner(NewArgs("C", "--p").SetComp(0))
		r := UnwrapStdout(err)
		assert.Equal(t, r, "--potato")
	}

	t.Run("wtf", func(t *testing.T) {
		_, err := parser.RunInner(NewArgs("x").SetComp(0))
		r := UnwrapStdout(err)
		assert.Equal(t, r, "x\n")
	})

	t.Run("wtf", func(t *testing.T) {
		_, err := parser.RunInner(NewArgs("").SetComp(0))
		r := UnwrapStdout(err)
		assert.Equal(t, r, "check\tcheck\t\tcheck packages\nclean\tclean\t\tclean target dir\nbuild\tbuild\t\tbuild project\n\n")
	})

	{
		_, err := parser.RunInner(NewArgs("ch").SetComp(0))
		r := UnwrapStdout(err)
		assert.Equal(t, r, "check")
	}
}

func TestStatic_complete_test_3(t *testing.T) {
	a := Argument(Long("potato").Help("po"), "P", parseString)
	b := Argument(Long("banana").Help("ba"), "B", parseString)
	ab := Pair(a, b)
	c := Argument(Long("durian"), "D", parseString)
	parser := Pair(ab, c).Parser().ToOptions()

	{
		_, err := parser.RunInner(NewArgs("").SetComp(0))
		r := UnwrapStdout(err)

		assert.Equal(t, r, "--potato\t--potato=P\t\tpo\n--banana\t--banana=B\t\tba\n--durian\t--durian=D\t\t\n\n")
	}

	{
		_, err := parser.RunInner(NewArgs("-").SetComp(0))
		r := UnwrapStdout(err)

		assert.Equal(t, r, "--potato\t--potato=P\t\tpo\n--banana\t--banana=B\t\tba\n--durian\t--durian=D\t\t\n\n")
	}

	{
		_, err := parser.RunInner(NewArgs("--").SetComp(0))
		r := UnwrapStdout(err)
		assert.Equal(t, r, "--potato\t--potato=P\t\tpo\n--banana\t--banana=B\t\tba\n--durian\t--durian=D\t\t\n\n")
	}

	{
		_, err := parser.RunInner(NewArgs("--d").SetComp(0))
		r := UnwrapStdout(err)
		assert.Equal(t, r, "--durian")
	}
}

func TestStatic_complete_test_4(t *testing.T) {
	a := Argument(Short('a'), "A", parseString)
	b := Argument(Short('b'), "B", parseString)
	parser := Pair(a, b).Parser().ToOptions()

	{
		_, err := parser.RunInner(NewArgs("-a", "").SetComp(0))
		r := UnwrapStdout(err)
		assert.Equal(t, r, "\tA\t\t\n\n")
	}

	t.Run("wtf", func(t *testing.T) {
		_, err := parser.RunInner(NewArgs("-a").SetComp(0))
		r := UnwrapStdout(err)
		assert.Equal(t, r, "-a")
	})

	t.Run("wtf", func(t *testing.T) {
		_, err := parser.RunInner(NewArgs("").SetComp(0))
		r := UnwrapStdout(err)
		assert.Equal(t, r, "-a\t-a=A\t\t\n-b\t-b=B\t\t\n\n")
	})

	t.Run("wtf", func(t *testing.T) {
		_, err := parser.RunInner(NewArgs("-").SetComp(0))
		r := UnwrapStdout(err)
		assert.Equal(t, r, "-a\t-a=A\t\t\n-b\t-b=B\t\t\n\n")
	})

	t.Run("wtf", func(t *testing.T) {
		_, err := parser.RunInner(NewArgs("--").SetComp(0))
		r := UnwrapStdout(err)
		assert.Equal(t, r, "--\n")
	})
}

func TestStatic_complete_test_5(t *testing.T) {
	a := Argument(Short('a'), "A", parseString)
	b := Argument(Short('b'), "B", parseString)
	c := Argument(Short('c'), "C", parseString)
	d := Argument(Short('d'), "D", parseString)
	ab := Pair(a, b)
	cd := Pair(c, d)
	parser := Pair(ab, cd).Parser().ToOptions()

	t.Run("wtf", func(t *testing.T) {
		_, err := parser.RunInner(NewArgs("-a", "x", "").SetComp(0))
		r := UnwrapStdout(err)
		assert.Equal(t, r, "-b\t-b=B\t\t\n-c\t-c=C\t\t\n-d\t-d=D\t\t\n\n")
	})

	{
		_, err := parser.RunInner(NewArgs("-a", "").SetComp(0))
		r := UnwrapStdout(err)
		assert.Equal(t, r, "\tA\t\t\n\n")
	}

	{
		_, err := parser.RunInner(NewArgs("-a").SetComp(0))
		r := UnwrapStdout(err)
		assert.Equal(t, r, "-a")
	}

	t.Run("wtf", func(t *testing.T) {
		_, err := parser.RunInner(NewArgs("").SetComp(0))
		r := UnwrapStdout(err)
		assert.Equal(t, r, "-a\t-a=A\t\t\n-b\t-b=B\t\t\n-c\t-c=C\t\t\n-d\t-d=D\t\t\n\n")
	})

	{
		_, err := parser.RunInner(NewArgs("-").SetComp(0))
		r := UnwrapStdout(err)

		assert.Equal(t, `-a	-a=A`+"\t\t"+`
-b	-b=B`+"\t\t"+`
-c	-c=C`+"\t\t"+`
-d	-d=D`+"\t\t"+`

`, r)
	}

	t.Run("wtf", func(t *testing.T) {
		_, err := parser.RunInner(NewArgs("--").SetComp(0))
		r := UnwrapStdout(err)
		assert.Equal(t, r, "--\n")
	})
}

func TestStatic_complete_test_6(t *testing.T) {
	a := Optional(Argument(Short('a'), "A", parseString).Parser(), false)
	b := Many(Argument(Short('b'), "B", parseString), false)
	parser := Pair(a, b).Parser().ToOptions()

	{
		_, err := parser.RunInner(NewArgs("-b", "x", "").SetComp(0))
		r := UnwrapStdout(err)
		assert.Equal(t, r, "-a\t-a=A\t\t\n-b\t-b=B\t\t\n\n")
	}

	{
		_, err := parser.RunInner(NewArgs("-a", "").SetComp(0))
		r := UnwrapStdout(err)
		assert.Equal(t, r, "\tA\t\t\n\n")
	}

	{
		_, err := parser.RunInner(NewArgs("").SetComp(0))
		r := UnwrapStdout(err)
		assert.Equal(t, r, "-a\t-a=A\t\t\n-b\t-b=B\t\t\n\n")
	}

	{
		_, err := parser.RunInner(NewArgs("-a", "x", "").SetComp(0))
		r := UnwrapStdout(err)
		assert.Equal(t, r, "-b")
	}

	{
		_, err := parser.RunInner(NewArgs("-a").SetComp(0))
		r := UnwrapStdout(err)
		assert.Equal(t, r, "-a")
	}

	{
		_, err := parser.RunInner(NewArgs("").SetComp(0))
		r := UnwrapStdout(err)
		assert.Equal(t, r, "-a\t-a=A\t\t\n-b\t-b=B\t\t\n\n")
	}

	{
		_, err := parser.RunInner(NewArgs("-b", "").SetComp(0))
		r := UnwrapStdout(err)
		assert.Equal(t, r, "\tB\t\t\n\n")
	}
}

func TestStatic_complete_test_7(t *testing.T) {
	a := Short('a').Help("switch").Switch()
	b := Positional("FILE", parseString).Help("File to use")
	parser := Pair(a, b).Parser().ToOptions()

	{
		_, err := parser.RunInner(NewArgs("").SetComp(0))
		r := UnwrapStdout(err)
		assert.Equal(t, r, "-a\t-a\t\tswitch\n\tFILE\t\tFile to use\n\n")
	}

	{
		_, err := parser.RunInner(NewArgs("-a", "").SetComp(0))
		r := UnwrapStdout(err)
		assert.Equal(t, r, "\tFILE\t\tFile to use\n\n")
	}

	{
		_, err := parser.RunInner(NewArgs("-a", "x").SetComp(0))
		r := UnwrapStdout(err)
		assert.Equal(t, r, "\tFILE\t\tFile to use\n\n")
	}
}

func TestStatic_complete_test_8(t *testing.T) {
	parser := Short('a').
		Long("durian").
		Switch().
		Parser().ToOptions().
		Command("nom").
		Parser().ToOptions()

	{
		_, err := parser.RunInner(NewArgs("").SetComp(0))
		r := UnwrapStdout(err)
		assert.Equal(t, r, "nom")
	}

	{
		_, err := parser.RunInner(NewArgs("nom", "").SetComp(0))
		r := UnwrapStdout(err)
		assert.Equal(t, r, "--durian")
	}

	{
		_, err := parser.RunInner(NewArgs("nom", "-a").SetComp(0))
		r := UnwrapStdout(err)
		assert.Equal(t, r, "--durian")
	}

	{
		_, err := parser.RunInner(NewArgs("nom", "-a", "").SetComp(0))
		r := UnwrapStdout(err)
		assert.Equal(t, r, "\n")
	}
}

func TestJust_positional(t *testing.T) {
	parser := Positional("FILE", parseString).
		Help("File to use").
		Parser().ToOptions()

	{
		_, err := parser.RunInner(NewArgs("").SetComp(0))
		r := UnwrapStdout(err)
		assert.Equal(t, r, "\tFILE\t\tFile to use\n\n")
	}

	{
		_, err := parser.RunInner(NewArgs("xxx").SetComp(0))
		r := UnwrapStdout(err)
		assert.Equal(t, r, "\tFILE\t\tFile to use\n\n")
	}
}

func TestDynamic_complete_test_1(t *testing.T) {
	parser := Argument(Short('a'),
		"ARG", parseString).
		Parser().Complete(test_completer).
		Parser().ToOptions()

	{
		_, err := parser.RunInner(NewArgs("-a", "b").SetComp(0))
		r := UnwrapStdout(err)

		assert.Equal(t, r, "\tARG\t\t\nbeta\tbeta\t\t\nbanana\tbanana\t\t\n\n")
	}

	{
		_, err := parser.RunInner(NewArgs("-a", "be").SetComp(0))
		r := UnwrapStdout(err)
		assert.Equal(t, r, "beta")
	}

	{
		_, err := parser.RunInner(NewArgs("-a", "beta").SetComp(0))
		r := UnwrapStdout(err)
		assert.Equal(t, r, "beta")
	}

	{
		_, err := parser.RunInner(NewArgs("").SetComp(0))
		r := UnwrapStdout(err)
		assert.Equal(t, r, "-a")
	}

	{
		_, err := parser.RunInner(NewArgs("-a", "").SetComp(0))
		r := UnwrapStdout(err)
		assert.Equal(t, r, "\tARG\t\t\nalpha\talpha\t\t\nbeta\tbeta\t\t\nbanana\tbanana\t\t\ncat\tcat\t\t\ndurian\tdurian\t\t\n\n")
	}
}

func TestDynamic_complete_test_2(t *testing.T) {
	parser := Argument(Short('a'), "ARG", parseString).Parser().ToOptions()

	_, err := parser.RunInner(NewArgs("-a", "b").SetComp(0))
	r := UnwrapStdout(err)
	assert.Equal(t, r, "\tARG\t\t\n\n")
}

func TestDynamic_complete_test_3(t *testing.T) {
	a := Short('a').Long("avocado").Help("Use avocado").Switch()
	b := Short('b').Long("banana").Help("Use banana").Switch()
	bb := Long("bananananana").Help("I'm Batman").Switch()
	c := Argument(Long("calculator").
		Help("calculator expression"),
		"EXPR", parseString).
		Parser().Complete(test_completer)
	parser := Pair(Pair(a, b), Pair(bb, c)).Parser().ToOptions()

	_, err := parser.RunInner(NewArgs("--calculator", "").SetComp(0))
	r := UnwrapStdout(err)
	assert.Equal(t, r, "\tEXPR\t\tcalculator expression\nalpha\talpha\t\t\nbeta\tbeta\t\t\nbanana\tbanana\t\t\ncat\tcat\t\t\ndurian\tdurian\t\t\n\n")
}

func TestDynamic_complete_test_4(t *testing.T) {
	parser := Argument(Long("name"),
		"NAME", parseString).
		Parser().Complete(test_completer_descr).
		Parser().ToOptions()

	{
		_, err := parser.RunInner(NewArgs("--name", "").SetComp(0))
		r := UnwrapStdout(err)
		assert.Equal(t, r, "\tNAME\t\t\nalpha\talpha\t\talpha\nbeta\tbeta\t\tbeta\nbanana\tbanana\t\tbanana\ncat\tcat\t\tcat\ndurian\tdurian\t\tdurian\n\n")
	}

	{
		_, err := parser.RunInner(NewArgs("--name", "a").SetComp(0))
		r := UnwrapStdout(err)
		assert.Equal(t, r, "alpha")
	}
}

func TestStatic_with_hide(t *testing.T) {
	a := Short('a').Switch()
	b := Short('b').Switch().Parser().Hide()
	parser := Pair(a, b).Parser().ToOptions()

	_, err := parser.RunInner(NewArgs("").SetComp(0))
	r := UnwrapStdout(err)
	assert.Equal(t, r, "-a")
}

func TestStatic_with_fallback_and_hide(t *testing.T) {
	a := Short('a').Switch()
	b := Short('b').Switch().Parser().Hide()
	parser := Pair(a, b).Parser().Fallback(fun.Pair[bool, bool]{false, false}).Parser().ToOptions()

	_, err := parser.RunInner(NewArgs("").SetComp(0))
	r := UnwrapStdout(err)
	assert.Equal(t, r, "-a")
}

func TestCsample_mystery(t *testing.T) {
	a := Short('a').Long("avocado").Help("Use avocado").Switch()
	b := Short('b').Long("banana").Help("Use banana").Switch()
	bb := Long("bananananana").Help("I'm Batman").Switch()
	c := Argument(Long("calculator").
		Help("calculator expression"),
		"EXPR", parseString).
		Parser().Complete(test_completer)
	parser := Pair(Pair(a, b), Pair(bb, c)).
		Parser().ToOptions().
		Descr("Dynamic autocomplete example").
		Footer("footer")

	_, err := parser.RunInner(NewArgs("").SetComp(0))
	r := UnwrapStdout(err)
	assert.Equal(t, r, "--avocado\t--avocado\t\tUse avocado\n--banana\t--banana\t\tUse banana\n--bananananana\t--bananananana\t\tI'm Batman\n--calculator\t--calculator=EXPR\t\tcalculator expression\n\n")
}

func TestOnly_positionals_after_double_dash(t *testing.T) {
	a := Short('a').Switch()
	b := Short('b').Switch()
	c := Short('c').Switch()
	d := Positional("D", parseString)
	parser := Pair(Pair(a, b), Pair(c, d)).Parser().ToOptions()

	{
		_, err := parser.RunInner(NewArgs("-a", "--").SetComp(0))
		r := UnwrapStdout(err)
		assert.Equal(t, r, "--\n")
	}

	{
		_, err := parser.RunInner(NewArgs("-a").SetComp(0))
		r := UnwrapStdout(err)
		assert.Equal(t, r, "-a")
	}

	{
		_, err := parser.RunInner(NewArgs("-a", "").SetComp(0))
		r := UnwrapStdout(err)
		assert.Equal(t, r, "-b\t-b\t\t\n-c\t-c\t\t\n\tD\t\t\n\n")
	}

	{
		_, err := parser.RunInner(NewArgs("--", "").SetComp(0))
		r := UnwrapStdout(err)
		assert.Equal(t, r, "\tD\t\t\n\n")
	}
}

func TestMany_does_not_duplicate_metadata(t *testing.T) {
	parser := Many(Positional("D", parseString), false).Parser().ToOptions()
	_, err := parser.RunInner(NewArgs("xxx").SetComp(0))
	r := UnwrapStdout(err)
	assert.Equal(t, r, "\tD\t\t\n\n")
}

func TestSome_does_not_duplicate_metadata(t *testing.T) {
	parser := ParseSomeParser(Positional("D", parseString).Parser().Some("", false)).ToOptions()
	_, err := parser.RunInner(NewArgs("xxx").SetComp(0))
	r := UnwrapStdout(err)
	assert.Equal(t, r, "\tD\t\t\n\n")
}

func TestOnly_positionals_after_positionals(t *testing.T) {
	a := Short('a').Switch()
	d := Many(Positional("D", parseString), false)
	parser := Pair(a, d).Parser().ToOptions()

	{
		_, err := parser.RunInner(NewArgs("").SetComp(0))
		r := UnwrapStdout(err)
		assert.Equal(t, r, "-a\t-a\t\t\n\tD\t\t\n\n")
	}

	{
		_, err := parser.RunInner(NewArgs("xxx").SetComp(0))
		r := UnwrapStdout(err)
		assert.Equal(t, r, "\tD\t\t\n\n")
	}

	{
		_, err := parser.RunInner(NewArgs("xxx", "yyy").SetComp(0))
		r := UnwrapStdout(err)
		assert.Equal(t, r, "\tD\t\t\n\n")
	}

	{
		_, err := parser.RunInner(NewArgs("xxx", "").SetComp(0))
		r := UnwrapStdout(err)
		assert.Equal(t, r, "-a\t-a\t\t\n\tD\t\t\n\n")
	}
}

func TestPositionals_complete_in_order(t *testing.T) {
	a := Positional("A", parseString).Parser().Complete(complete_alpha)
	b := Positional("B", parseString).Parser().Complete(complete_beta)
	parser := Pair(a, b).Parser().ToOptions()

	{
		_, err := parser.RunInner(NewArgs("").SetComp(0))
		r := UnwrapStdout(err)
		assert.Equal(t, r, "alpha")
	}

	{
		_, err := parser.RunInner(NewArgs("a").SetComp(0))
		r := UnwrapStdout(err)
		assert.Equal(t, r, "alpha")
	}

	{
		_, err := parser.RunInner(NewArgs("x").SetComp(0))
		r := UnwrapStdout(err)
		assert.Equal(t, r, "\tA\t\t\n\n")
	}

	{
		_, err := parser.RunInner(NewArgs("xxx", "").SetComp(0))
		r := UnwrapStdout(err)
		assert.Equal(t, r, "beta")
	}

	{
		_, err := parser.RunInner(NewArgs("xxx", "b").SetComp(0))
		r := UnwrapStdout(err)
		assert.Equal(t, r, "beta")
	}

	{
		_, err := parser.RunInner(NewArgs("xxx", "yyy").SetComp(0))
		r := UnwrapStdout(err)
		assert.Equal(t, r, "\tB\t\t\n\n")
	}
}

func TestShould_be_able_to_suggest_positional_along_with_non_positionals_flags(t *testing.T) {
	a := Argument(Short('a'), "A", parseString).Parser().Complete(complete_alpha)
	b := Positional("B", parseString).Parser().Complete(complete_beta)
	parser := Pair(a, b).Parser().ToOptions()

	_, err := parser.RunInner(NewArgs("").SetComp(0))
	r := UnwrapStdout(err)
	assert.Equal(t, r, "-a\t-a=A\t\t\nbeta\tbeta\t\tbeta description\n\n")
}

func TestShould_be_able_to_suggest_double_dash(t *testing.T) {
	c_b := func(string) SSOS {
		return zalupa12(funiter.FromMany(PSOS{"--", Invalid[string]()}))
	}
	a := Optional(Argument(Long("arg"),
		"ARG", parseString).
		Parser().Complete(c_b).Parser(), false)

	// parser := Pair(a).Parser().ToOptions()
	parser := a.Parser().ToOptions()

	{
		_, err := parser.RunInner(NewArgs("").SetComp(0))
		r := UnwrapStdout(err)
		assert.Equal(t, r, "--arg")
	}

	{
		_, err := parser.RunInner(NewArgs("--arg", "").SetComp(0))
		r := UnwrapStdout(err)
		assert.Equal(t, r, "--")
	}
}

func TestNon_strict_and_double_dash(t *testing.T) {
	a := Short('a').Switch()
	b := Positional("B", parseString).NonStrict()
	parser := Pair(a, b).Parser().ToOptions()

	{
		_, err := parser.RunInner(NewArgs("").SetComp(0))
		r := UnwrapStdout(err)
		assert.Equal(t, r, "-a\t-a\t\t\n\tB\t\t\n\n")
	}

	{
		_, err := parser.RunInner(NewArgs("--", "").SetComp(0))
		r := UnwrapStdout(err)
		assert.Equal(t, r, "\n")
	}
}

func TestSuggest_double_dash_automatically_for_strictly_positional(t *testing.T) {
	a := Short('a').Switch()
	b := Positional("B", parseString).Strict()
	parser := Pair(a, b).Parser().ToOptions()

	{
		_, err := parser.RunInner(NewArgs("").SetComp(0))
		r := UnwrapStdout(err)

		assert.Equal(t, r, "-a\t-a\t\t\n--\t--\t\tPositional only items after this token\n\n")
	}

	{
		_, err := parser.RunInner(NewArgs("-").SetComp(0))
		r := UnwrapStdout(err)

		assert.Equal(t, r, "-a\t-a\t\t\n--\t--\t\tPositional only items after this token\n\n")
	}

	{
		_, err := parser.RunInner(NewArgs("--").SetComp(0))
		r := UnwrapStdout(err)
		assert.Equal(t, r, "--\n")
	}

	{
		_, err := parser.RunInner(NewArgs("--", "").SetComp(0))
		r := UnwrapStdout(err)
		assert.Equal(t, r, "\tB\t\t\n\n")
	}
}

func TestMultiple_adjacent_single_letter_vals(t *testing.T) {
	// but not really - test is incomplete
	a := Short('a').Switch()
	b := Short('b').Switch()
	c := Short('c').Switch()
	parser := Pair(Pair(a, b), c).Parser().ToOptions()

	{
		// with no input the right behavior is to suggest all the switches
		_, err := parser.RunInner(NewArgs("").SetComp(0))
		r := UnwrapStdout(err)
		assert.Equal(t, r, "-a\t-a\t\t\n-b\t-b\t\t\n-c\t-c\t\t\n\n")
	}

	{
		// with a single item present separately we should suggest the remaining two
		_, err := parser.RunInner(NewArgs("-a", "").SetComp(0))
		r := UnwrapStdout(err)
		assert.Equal(t, r, "-b\t-b\t\t\n-c\t-c\t\t\n\n")
	}

	{
		// trying to complete a pair of flags should dump current state
		_, err := parser.RunInner(NewArgs("-ab").SetComp(0))
		r := UnwrapStdout(err)
		assert.Equal(t, r, "-ab\n")
	}

	{
		// with a single valid item we should suggest it
		_, err := parser.RunInner(NewArgs("-ab", "").SetComp(0))
		r := UnwrapStdout(err)
		assert.Equal(t, r, "-c")
	}

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
	a0 := Short('a').Switch().Parser().Count().Parser()
	a1 := Argument(Short('a'), "AAAAAA", parseInt)
	parser := Choice(a0, a1).Parser().ToOptions()

	_, err := parser.RunInner(NewArgs("-aaa").SetComp(0))
	r := UnwrapStdout(err)

	assert.Equal(t, r, "-aaa\n")
}

func TestAmbiguity_to_flags(t *testing.T) {
	parser := Many(Short('a').Switch(), false).Parser().ToOptions()

	_, err := parser.RunInner(NewArgs("-aaa").SetComp(0))
	r := UnwrapStdout(err)
	assert.Equal(t, r, "-aaa\n")
}

func TestShort_argument_variants(t *testing.T) {
	parser := Argument(Short('a'), "META", parseString).Parser().ToOptions()
	{
		_, err := parser.RunInner(NewArgs("-a=aa").SetComp(0))
		r := UnwrapStdout(err)
		assert.Equal(t, r, "\tMETA\t\t\n\n")
	}

	{
		_, err := parser.RunInner(NewArgs("-a", "aa").SetComp(0))
		r := UnwrapStdout(err)
		assert.Equal(t, r, "\tMETA\t\t\n\n")
	}

	{
		_, err := parser.RunInner(NewArgs("-aaa").SetComp(0))
		r := UnwrapStdout(err)
		assert.Equal(t, r, "\tMETA\t\t\n\n")
	}
}

func TestLong_argument_variants(t *testing.T) {
	parser := Argument(Long("alpha"),
		"META", parseString).
		Parser().Complete(complete_beta).
		Parser().ToOptions()

	{
		_, err := parser.RunInner(NewArgs("--alpha=beta").SetComp(0))
		r := UnwrapStdout(err)
		assert.Equal(t, r, "--alpha=beta")
	}

	{
		_, err := parser.RunInner(NewArgs("--alpha", "Regina").SetComp(0))
		r := UnwrapStdout(err)
		assert.Equal(t, r, "\tMETA\t\t\n\n")
	}
}

func TestZsh_style_completion_visible(t *testing.T) {
	a := Argument(Short('a').
		Long("argument").
		Help("this is an argument"),
		"ARG", parseString)
	b := Argument(Short('b'), "BANANA", parseString)
	parser := Pair(a, b).Parser().GroupHelp("items").Parser().ToOptions()

	_, err := parser.RunInner(NewArgs("").SetComp(0))
	r := UnwrapStdout(err)
	assert.Equal(t, r, "--argument\t--argument=ARG\titems\tthis is an argument\n-b\t-b=BANANA\titems\t\n\n")
}

func TestZsh_many_positionals(t *testing.T) {
	parser := Many(Positional("POS", parseString), false).Parser().ToOptions()

	{
		_, err := parser.RunInner(NewArgs("").SetComp(0))
		r := UnwrapStdout(err)
		assert.Equal(t, r, "\tPOS\t\t\n\n")
	}

	{
		_, err := parser.RunInner(NewArgs("p").SetComp(0))
		r := UnwrapStdout(err)
		assert.Equal(t, r, "\tPOS\t\t\n\n")
	}
}

func TestZsh_help_single_line_only(t *testing.T) {
	a := Argument(Short('a').Help("hello world"), "X", parseString)
	b := Short('b').Help("hello from switch").Switch()
	parser := Pair(a, b).Parser().ToOptions()

	_, err := parser.RunInner(NewArgs("").SetComp(0))
	r := UnwrapStdout(err)
	assert.Equal(t, r, "-a\t-a=X\t\thello world\n-b\t-b\t\thello from switch\n\n")
}

func TestShell_help_single_line_only(t *testing.T) {
	a := Argument(Short('a').Help("hello 1\n\nworld"), "X", parseString)
	b := Argument(Short('b').Help("hello 2\n\nworld"), "Y", parseString)
	parser := Pair(a, b).Parser().ToOptions()

	_, err := parser.RunInner(NewArgs("").SetComp(0))
	r := UnwrapStdout(err)

	assert.Equal(t, r, "-a\t-a=X\t\thello 1\n-b\t-b=Y\t\thello 2\n\n")
}

// TODO: get back
// func TestDerive_decorations(t*testing.T) {
//     #[derive(Debug, Clone, Bpaf)]
//     #[allow(dead_code)]
//     /// == Cargo options
//     struct CargoOpts {
//         /// optimize
//         release: bool,
//         /// pick target
//         target: String,
//     }

//     #[derive(Debug, Clone, Bpaf)]
//     #[allow(dead_code)]
//     /// == Application options
//     struct AppOpts {
//         /// pick focus
//         focus: String,
//         /// inline rust
//         inline: bool,
//     }

//     #[derive(Debug, Clone, Bpaf)]
//     #[allow(dead_code)]
//     #[bpaf(options)]
//     struct Opts {
//         #[bpaf(external)]
//         cargo_opts: CargoOpts,
//         #[bpaf(external)]
//         app_opts: AppOpts,
//     }

//     parser := opts();

//     r := parser.
//         RunInner(NewArgs("").SetComp(0)).
//         unwrap_err().
//         UnwrapStdout();
//     assert.Equal(t, r, `
// --release\t--release\t== Cargo options\toptimize
// --target\t--target=ARG\t== Cargo options\tpick target
// --focus\t--focus=ARG\t== Application options\tpick focus
// --inline\t--inline\t== Application options\tinline rust\n\n`);
// }

func TestZsh_complete_info(t *testing.T) {
	foo := func(string) SSOS {
		return zalupa12(funiter.FromMany([]PSOS{{"hello", Valid("word")}, {"sample", Invalid[string]()}}...))
	}
	parser := Argument(Short('a'),
		"X", parseString).
		Parser().Complete(foo).
		Parser().ToOptions()

	{
		_, err := parser.RunInner(NewArgs("").SetComp(0))
		r := UnwrapStdout(err)
		assert.Equal(t, r, "-a")
	}

	{
		_, err := parser.RunInner(NewArgs("-").SetComp(0))
		r := UnwrapStdout(err)
		assert.Equal(t, r, "-a")
	}

	{
		_, err := parser.RunInner(NewArgs("-a", "").SetComp(0))
		r := UnwrapStdout(err)
		assert.Equal(t, r, "\tX\t\t\nhello\thello\t\tword\nsample\tsample\t\t\n\n")
	}
}

func TestPair_of_positionals(t *testing.T) {
	// with Positional items only current item should make suggestions, not both...
	a := Positional("A", parseString).Parser().Complete(test_completer)
	b := Positional("B", parseString).Parser().Complete(test_completer2)
	parser := Pair(a, b).Parser().ToOptions()

	{
		_, err := parser.RunInner(NewArgs("a").SetComp(0))
		r := UnwrapStdout(err)
		assert.Equal(t, r, "alpha")
	}

	{
		_, err := parser.RunInner(NewArgs("alpha", "a").SetComp(0))
		r := UnwrapStdout(err)
		assert.Equal(t, r, "auto")
	}
}

func TestDouble_dash_as_positional(t *testing.T) {
	parser := Positional("P", parseString).
		Help("Help").
		Parser().Complete(test_completer).
		Parser().ToOptions()

	{
		_, err := parser.RunInner(NewArgs("--", "a").SetComp(0))
		r := UnwrapStdout(err)
		assert.Equal(t, r, "alpha")
	}

	{
		_, err := parser.RunInner(NewArgs("a").SetComp(0))
		r := UnwrapStdout(err)
		assert.Equal(t, r, "alpha")
	}

	{
		_, err := parser.RunInner(NewArgs("-").SetComp(0))
		r := UnwrapStdout(err)
		assert.Equal(t, r, "-\n")
	}

	{
		_, err := parser.RunInner(NewArgs("--").SetComp(0))
		r := UnwrapStdout(err)
		assert.Equal(t, r, "--\n")
	}

	{
		_, err := parser.RunInner(NewArgs("--").SetComp(0))
		r := UnwrapStdout(err)
		assert.Equal(t, r, "--\n")
	}

	{
		_, err := parser.RunInner(NewArgs("x").SetComp(0))
		r := UnwrapStdout(err)
		assert.Equal(t, r, "\tP\t\tHelp\n\n")
	}
}

func TestStrict_positional_completion(t *testing.T) {
	a := Long("arg").Switch()
	p := Positional("S", parseString).
		Strict().
		Parser().Complete(func(string) SSOS { return zalupa12(funiter.FromMany(PSOS{"--hello", Invalid[string]()})) })
	parser := Pair(a, p).Parser().ToOptions()

	{
		_, err := parser.RunInner(NewArgs("--").SetComp(0))
		r := UnwrapStdout(err)
		assert.Equal(t, r, "--arg\t--arg\t\t\n--hello\t--hello\t\t\n\n")
	}

	{
		_, err := parser.RunInner(NewArgs("--a").SetComp(0))
		r := UnwrapStdout(err)
		assert.Equal(t, r, "--arg")
	}

	{
		_, err := parser.RunInner(NewArgs("--", "").SetComp(0))
		r := UnwrapStdout(err)
		assert.Equal(t, r, "--hello")
	}

	{
		_, err := parser.RunInner(NewArgs("--", "--h").SetComp(0))
		r := UnwrapStdout(err)
		assert.Equal(t, r, "--hello")
	}
}

func TestAvoid_inserting_metavars(t *testing.T) {
	parser := Argument(Short('a'), "A", parseString).Parser().ToOptions()

	{
		_, err := parser.RunInner(NewArgs("").SetComp(0))
		r := UnwrapStdout(err)
		assert.Equal(t, r, "-a")
	}

	{
		_, err := parser.RunInner(NewArgs("-a", "").SetComp(0))
		r := UnwrapStdout(err)
		assert.Equal(t, r, "\tA\t\t\n\n")
	}
}

func TestShell_dir_completion(t *testing.T) {
	parser := Argument(Short('a'),
		"FILE", parseString).
		Parser().CompleteShell(ShellCompDir(Invalid[string]())).
		Parser().ToOptions()

	_, err := parser.RunInner(NewArgs("-a", "").SetComp(0))
	r := UnwrapStdout(err)
	assert.Equal(t, r, "\nDir { mask: None }\n")
}

func TestGenerate_unparseable_items(t *testing.T) {
	one := Pure(struct{}{}).Parser().ToOptions().Command("cone")
	two := Pure(struct{}{}).Parser().ToOptions().Command("ctwo")
	e := Short('e').Switch()

	one_e := Map(Pair(e, one).Parser(), func(x fun.Pair[bool, struct{}]) struct{} { return x.V })
	parser := Choice(one_e, two).Parser().ToOptions()

	{
		// passing -e restricts branch with cmd_two
		_, err := parser.RunInner(NewArgs("-e", "").SetComp(0))
		r := UnwrapStdout(err)
		assert.Equal(t, r, "cone")
	}

	{
		// passing -e restricts branch with cmd_two
		_, err := parser.RunInner(NewArgs("-e", "c").SetComp(0))
		r := UnwrapStdout(err)
		assert.Equal(t, r, "cone")
	}

	{
		_, err := parser.RunInner(NewArgs("").SetComp(0))
		r := UnwrapStdout(err)
		assert.Equal(t, r, "-e\t-e\t\t\ncone\tcone\t\t\nctwo\tctwo\t\t\n\n")
	}
}

func TestComplete_with_fallback(t *testing.T) {
	parser := Parse(Argument(Long("name"),
		"NAME", parseString).
		Parser().Complete(test_completer_descr).Parser(),
		func(x string) (uint16, error) {
			u, err := strconv.ParseUint(x, 10, 16)
			return uint16(u), err
		}).
		Parser().Fallback(10).
		Parser().ToOptions()

	_, err := parser.RunInner(NewArgs("--name", "").SetComp(0))
	r := UnwrapStdout(err)
	assert.Equal(t, r, `	NAME`+"\t\t"+`
alpha	alpha		alpha
beta	beta		beta
banana	banana		banana
cat	cat		cat
durian	durian		durian

`)
}

func TestMix_of_options_and_positional_completions(t *testing.T) {
	a := Argument(Short('a').
		Long("arg").
		Help("Alhpa argument"),
		"ALPHA", parseString).
		Parser().Complete(complete_alpha)
	b := Positional("BETA", parseString).
		Help("Beta argument").
		Parser().Complete(complete_beta)
	parser := Pair(a, b).Parser().ToOptions()

	_, err := parser.RunInner(NewArgs("").SetComp(0))
	r := UnwrapStdout(err)

	assert.Equal(t, r, "--arg\t--arg=ALPHA\t\tAlhpa argument\nbeta\tbeta\t\tbeta description\n\n")
}

func TestPositionals_with_no_completions_are_not_duplicated(t *testing.T) {
	a := Argument(Short('a').
		Long("arg").
		Help("Alhpa argument"),
		"ALPHA", parseString)
	b := Positional("BETA", parseString).Help("Beta argument")

	parser := Pair(a, b).Parser().ToOptions()

	_, err := parser.RunInner(NewArgs("").SetComp(0))
	r := UnwrapStdout(err)

	assert.Equal(t, r, "--arg\t--arg=ALPHA\t\tAlhpa argument\n\tBETA\t\tBeta argument\n\n")
}
