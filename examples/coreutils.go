// `coreutils.go` implements parsers for a few coreutils tools - they tend to have
// complicated CLI rules due to historical reasons.
package examples

import (
	"fmt"
	"os"

	"github.com/rprtr258/cli"
)

type coreutilsVerbosity uint8

const (
	coreutilsVerbosityWarn coreutilsVerbosity = iota
	coreutilsVerbosityQuiet
	coreutilsVerbosityStatus
)

// Assuming a parser can consume one of several values - try to consume them all and
// return the last one.
var coreutilsParseVerbosity = func() cli.Parser[coreutilsVerbosity] {
	warn := cli.Short('w').Long("warn").Help("Display warnings").ReqFlag(coreutilsVerbosityWarn)
	quiet := cli.Short('q').Long("quiet").Help("Display only diagnostics").ReqFlag(coreutilsVerbosityQuiet)
	status := cli.Short('s').Long("status").Help("Display status only").ReqFlag(coreutilsVerbosityStatus)
	return cli.Choice(warn, quiet, status).Last().Fallback(coreutilsVerbosityStatus)
}()

var coreutilsParseBinary = func() cli.Parser[bool] {
	binary := cli.Short('b').Long("binary").Help("Use binary mode").ReqFlag(true)
	text := cli.Short('t').Long("text").Help("Use text mode").ReqFlag(false)
	return cli.Choice(binary, text).Last().Fallback(false)
}()

type coreutilsB2Sum struct {
	binary    bool
	check     bool
	tag       bool
	verbosity coreutilsVerbosity
	strict    bool
	files     []string
}

type coreutilsBase32 struct {
	decode        bool
	ignoreGarbage bool
	wrap          cli.Opt[uint]
	file          cli.Opt[string]
}

type coreutilsBasename struct {
	multiple bool
	suffix   cli.Opt[string]
	zero     bool
	names    []string
}

type coreutilsCat struct {
	showTabs        bool
	showEnds        bool
	showNonprinting bool
	number          string // "none", "nonempty", "all"
	squeezeBlank    bool
	files           []string
}

type coreutilsCommand struct {
	kind     string // "arch", "b2sum", "base32", "basename", "cat"
	b2sum    coreutilsB2Sum
	base32   coreutilsBase32
	basename coreutilsBasename
	cat      coreutilsCat
}

var coreutilsArch = cli.Pure(coreutilsCommand{kind: "arch"}).
	ToOptions().
	Command("arch").
	Help("Print machine architecture.")

var coreutilsParseB2Sum = func() cli.Parser[coreutilsCommand] {
	check := cli.Short('c').Long("check").
		Help("read BLAKE2 sums from the FILEs and check them").Switch()
	tag := cli.Long("tag").
		Help("create a BSD-style checksum").Switch()
	strict := cli.Long("strict").
		Help("exit non-zero for improperly formatted checksum lines").Switch()
	files := cli.Positional("FILE", cli.ParseString).Many()

	return cli.Map6(
		coreutilsParseBinary, check, tag, coreutilsParseVerbosity, strict, files,
		func(binary, check, tag bool, verbosity coreutilsVerbosity, strict bool, fs []string) coreutilsCommand {
			return coreutilsCommand{kind: "b2sum", b2sum: coreutilsB2Sum{
				binary:    binary,
				check:     check,
				tag:       tag,
				verbosity: verbosity,
				strict:    strict,
				files:     fs,
			}}
		},
	).
		ToOptions().
		Command("b2sum").
		Help("Print or check BLAKE2 (512-bit) checksums.").
		Parser()
}()

var coreutilsParseBase32 = func() cli.Parser[coreutilsCommand] {
	decode := cli.Short('d').Long("decode").
		Help("decode data").Switch()
	ignoreGarbage := cli.Short('i').Long("ignore-garbage").
		Help("when decoding, ignore non-alphabet characters").Switch()
	wrap := cli.Short('w').Long("wrap").
		Help("wrap encoded lines after COLS character. Use 0 to disable line wrapping").
		Argument("COLS", cli.ParseUint).
		Optional().
		Map(func(wrap cli.Opt[uint]) cli.Opt[uint] {
			cols, ok := wrap.Unpack()
			if !ok {
				return cli.Valid(uint(76)) // default line width
			}
			if cols == 0 {
				return cli.Invalid[uint]() // 0 disables line wrapping
			}
			return wrap
		})
	file := cli.Positional("FILE", cli.ParseString).
		Help("With no FILE, or when FILE is -, read standard input.").
		Optional()

	return cli.Map4(
		decode, ignoreGarbage, wrap, file,
		func(decode, ignoreGarbage bool, wrap cli.Opt[uint], file cli.Opt[string]) coreutilsCommand {
			return coreutilsCommand{kind: "base32", base32: coreutilsBase32{
				decode:        decode,
				ignoreGarbage: ignoreGarbage,
				wrap:          wrap,
				file:          file,
			}}
		},
	).
		ToOptions().
		Command("base32").
		Help("Base32 encode or decode FILE, or standard input, to standard output.").
		Parser()
}()

var coreutilsParseBasename = func() cli.Parser[coreutilsCommand] {
	multiple := cli.Short('a').Long("multiple").
		Help("support multiple arguments and treat each as a NAME").Switch()
	suffix := cli.Short('s').Long("suffix").
		Help("remove a trailing SUFFIX; implies -a").
		Argument("SUFFIX", cli.ParseString).
		Optional()
	zero := cli.Short('z').Long("zero").
		Help("end each output line with NUL, not newline").Switch()
	names := cli.Positional("NAME", cli.ParseString).
		Help("Print NAME with any leading directory components removed.").
		Many()

	return cli.Map4(
		multiple, suffix, zero, names,
		func(multiple bool, suffix cli.Opt[string], zero bool, names []string) coreutilsCommand {
			if _, ok := suffix.Unpack(); ok {
				multiple = true
			}
			return coreutilsCommand{kind: "basename", basename: coreutilsBasename{
				multiple: multiple,
				suffix:   suffix,
				zero:     zero,
				names:    names,
			}}
		},
	).
		ToOptions().
		Command("basename").
		Parser()
}()

var coreutilsParseCat = func() cli.Parser[coreutilsCommand] {
	// "extra" flags fold into the main ones below
	showAll := cli.Short('A').Long("show-all").
		Help("equivalent to -vET").Switch()
	numberNonblankExtra := cli.Short('b').Long("number-nonblank").
		Help("number nonempty output lines, overrides -n").Switch()
	showNonPrintingEnds := cli.Short('e').
		Help("equivalent to -vE").Switch()

	showTabs := cli.Short('T').Long("show-tabs").
		Help("display TAB characters as ^I").Switch()
	showEnds := cli.Short('E').
		Help("display $ at end of each line").Switch()
	showNonprinting := cli.Short('v').
		Help("use ^ and M- notation, except for LFD and TAB").Switch()
	numberAll := cli.Short('n').Long("number").
		Help("number all output lines").Switch()
	squeezeBlank := cli.Short('s').Long("squeeze-blank").
		Help("suppress repeated empty output lines").Switch()
	files := cli.Positional("FILE", cli.ParseString).
		Help("Concatenate FILE(s) to standard output.").
		Many()

	extra := cli.Map3(
		showAll, numberNonblankExtra, showNonPrintingEnds,
		func(showAll, numberNonblankExtra, showNonPrintingEnds bool) coreutilsCat {
			return coreutilsCat{
				showTabs:        showAll,
				showEnds:        showAll || showNonPrintingEnds,
				showNonprinting: showAll || showNonPrintingEnds,
				number:          map[bool]string{true: "nonempty", false: "none"}[numberNonblankExtra],
			}
		},
	)
	main := cli.Map6(
		showTabs, showEnds, showNonprinting, numberAll, squeezeBlank, files,
		func(showTabs, showEnds, showNonprinting, numberAll, squeezeBlank bool, fs []string) coreutilsCat {
			return coreutilsCat{
				showTabs:        showTabs,
				showEnds:        showEnds,
				showNonprinting: showNonprinting,
				number:          map[bool]string{true: "all", false: "none"}[numberAll],
				squeezeBlank:    squeezeBlank,
				files:           fs,
			}
		},
	)

	return cli.Map2(extra, main, func(merged, main coreutilsCat) coreutilsCommand {
		if main.showTabs {
			merged.showTabs = true
		}
		if main.showEnds {
			merged.showEnds = true
		}
		if main.showNonprinting {
			merged.showNonprinting = true
		}
		if main.number != "none" {
			merged.number = main.number
		}
		merged.squeezeBlank = main.squeezeBlank
		merged.files = main.files
		return coreutilsCommand{kind: "cat", cat: merged}
	}).
		ToOptions().
		Command("cat").
		Parser()
}()

func Coreutils() {
	parser := cli.Choice(
		coreutilsArch,
		coreutilsParseB2Sum,
		coreutilsParseBase32,
		coreutilsParseBasename,
		coreutilsParseCat,
	).ToOptions()

	fmt.Println(parser.Run(os.Args))
}
