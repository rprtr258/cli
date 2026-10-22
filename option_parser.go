package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/rprtr258/fun"
)

// Ready to run [Parser] with additional information attached
//
// Created with [Parser.ToOptions]
//
// In addition to the inner parser `OptionParser` contains documentation about a program or a
// subcommand as a whole, version, custom usage, if specified, and handles custom parsers for
// `--version` and `--help` flags.
type OptionParser[T any] struct {
	inner Parser[T]
	info  info
}

// Execute the [OptionParser], extract a parsed value or print some diagnostic and exit
//
// # Usage
//
//	// Parses the number of repetitions of -v on a command line
//	verbosity := cli.Short('v').Switch().Many().
//	    Map(func(vs []struct{}) int { return len(vs) }).
//	    ToOptions().
//	    Description("Takes verbosity flag and does nothing else")
//	var verbosityCount int = verbosity.Run(os.Args)
func (self OptionParser[T]) Run(args []string) T {
	t, err := self.RunE(args)
	if err == nil {
		return t
	}

	failure := err.(FailureError) // RunE only ever returns FailureError on failure

	// Prints a message to `stdout` or `stderr` appropriate to the failure.
	color := colorDefault
	switch f := failure.failure.(type) {
	case parseFailureStdout:
		fmt.Println(f.msg.render_console(f.full, color, self.info.max_width))
	case parseFailureCompletion:
		fmt.Print(string(f))
	case parseFailureStderr:
		fmt.Fprintln(os.Stderr, color.push_str(StyleInvalid, "parseError: ")+Doc(f).render_console(true, color, self.info.max_width))
	}

	// matches Rust's behavior: exit with the code (info.rs:94)
	os.Exit(failure.ExitCode())
	return t
}

// Execute the [OptionParser], extract a parsed value or return a [FailureError] without
// printing anything and without exiting the process
//
// This is the embeddable counterpart of [Run]: use it when the parser is a part of a bigger
// application or when the parser needs to be tested from outside of this package.
//
// `args` is the same as for [Run]: the first element is the program name, usually `os.Args`.
//
// # Usage
//
//	t, err := parser.RunE(os.Args)
//	if err != nil {
//		var failure cli.FailureError
//		if errors.As(err, &failure) {
//			out := os.Stdout
//			if failure.IsError() {
//				out = os.Stderr
//			}
//			fmt.Fprintln(out, failure.Error())
//			os.Exit(failure.ExitCode())
//		}
//	}
func (self OptionParser[T]) RunE(args []string) (T, error) {
	current_args := argList{
		items: args[1:],
		name:  Valid(filepath.Base(args[0])),
		comp:  Invalid[uint](),
	}

	t, failure := self.run(current_args)
	if failure != nil {
		return *new(T), FailureError{failure: failure, maxWidth: self.info.max_width}
	}

	return t, nil
}

// Diagnostic produced by a failed [OptionParser.RunE] call, this is the same thing
// [OptionParser.Run] prints before exiting
//
// The diagnostic is one of:
//   - help or version message: belongs on `stdout`, [FailureError.ExitCode] is 0
//   - shell completion script: belongs on `stdout`, [FailureError.ExitCode] is 0
//   - error message: belongs on `stderr`, [FailureError.ExitCode] is 1
type FailureError struct {
	failure  parseFailure
	maxWidth uint
}

// Rendered diagnostic message, without the `parseError: ` prefix and without a trailing newline
func (self FailureError) Error() string {
	color := colorDefault
	switch f := self.failure.(type) {
	case parseFailureStdout:
		return f.msg.render_console(f.full, color, self.maxWidth)
	case parseFailureCompletion:
		return string(f)
	case parseFailureStderr:
		return Doc(f).render_console(true, color, self.maxWidth)
	default:
		return ""
	}
}

// IsError reports whether the diagnostic is an error message: it belongs on `stderr` and
// [FailureError.ExitCode] is 1. Otherwise it is help, version or completion output which
// belongs on `stdout` and [FailureError.ExitCode] is 0.
func (self FailureError) IsError() bool {
	_, ok := self.failure.(parseFailureStderr)
	return ok
}

// Exit code [OptionParser.Run] uses for this diagnostic, see [FailureError.IsError]
func (self FailureError) ExitCode() int {
	return fun.IF(self.IsError(), 1, 0)
}

// Check the invariants the parser relies on for normal operations
//
// Takes a parameter whether to check for cosmetic invariants or not
// (max help width exceeding 120 symbols, etc), currently not in use
//
// Best used as part of your test suite:
//
//	parser.ToOptions().CheckInvariants(false)
//
// # Panics
//
// CheckInvariants indicates problems with panic.
func (self OptionParser[T]) CheckInvariants(bool) {
	positional_invariant_check(self.inner.meta, true)
}

// Execute the [OptionParser] and produce a value for unit tests or manual processing
//
//	parser := cli.Positional[string]("FILE", nil).
//	    Help("File to process").
//	    Parser().
//	    ToOptions()
//
//	_, err := parser.RunE([]string{"prog", "--help"})
//	var failure cli.FailureError
//	require.ErrorAs(t, err, &failure)
//	expectedHelp := `
//
// Usage: prog FILE
//
// Available positional items:
//
//	FILE        File to process
//
// Available options:
//
//	-h, --help  Prints help information
//
// `
//
//	require.Equal(t, expectedHelp, failure.Error())
//
// # Errors
//
// If the parser can't produce the desired result, run returns a [parseFailure]
// which represents runtime behavior: one branch to print something to stdout and exit with
// success, the other to print something to stderr and exit with failure.
//
// The library generates the contents of the [parseFailure] using the expected textual output
// of the parser; stdout/stderr isn't actually captured.
//
// Exact string representations may change between versions, including minor releases.
func (self OptionParser[T]) run(args argList) (T, parseFailure) {
	// prepare available short flags and arguments for disambiguation
	var short_flags []byte
	var short_args []byte
	collect_shorts(self.inner.meta, &short_flags, &short_args)
	short_flags = append(short_flags, self.info.help_arg.short...)
	short_flags = append(short_flags, self.info.version_arg.short...)

	// this only handles disambiguation failure in construct
	state, err := newState(args, short_flags, short_args)
	if err != nil && !state.comp.Valid {
		return *new(T), render(&err, state, self.inner.meta)
	}

	return self.run_subparser(&state)
}

func (self OptionParser[T]) run_subparser(state *state) (T, parseFailure) {
	// process should work like this:
	// - inner parser is evaluated, it returns parseError
	// - if error is finalized (parseFailure) - it is simply propagated outwards,
	//   otherwise we are making a few attempts at improving it after dealing with autocomplete/help
	// - generate autocomplete, if enabled
	// - produce --help, --version
	// - try to improve error message and finalize it otherwise

	if self.info.help_if_no_args && state.len() == 0 {
		buffer := render_help(
			state.path,
			self.info,
			self.inner.meta,
			self.info.meta(),
		)
		return *new(T), parseFailureStdout{buffer, false}
	}

	res, err := self.inner.eval(state)
	if err != nil {
		if parseFailure, ok := err.(errorParseFailure); ok {
			return *new(T), parseFailure.e
		}
	}
	if comp, ok := state.check_complete(); ok {
		return *new(T), parseFailureCompletion(comp)
	} else if err == nil {
		if ix, _, ok := state.items_iter()(); ok {
			err = errorUnconsumed(ix)
		} else {
			return res, nil
		}
	}

	// handle --help and --version messages
	extra, errInfo := self.info.eval(state)
	if errInfo != nil {
		return *new(T), render(&err, *state, self.inner.meta)
	}

	msg := func() Doc {
		if extra.Help {
			return render_help(
				state.path,
				self.info,
				self.inner.meta,
				self.info.meta(),
			)
		} else {
			doc := Doc{}
			doc.token(tokenBlockStart(blockBlock))
			doc.Text("Version: ")
			doc.Doc(extra.Version)
			doc.token(tokenBlockEnd(blockBlock))
			return doc
		}
	}()
	return *new(T), parseFailureStdout{msg, extra.Help}
}

// Get first line of description if Available
//
// Used internally to avoid duplicating description for [command].
func (self OptionParser[T]) short_descr() Opt[Doc] {
	return self.info.descr.and_then(func(d Doc) Opt[Doc] {
		v, ok := Doc.first_line(d)
		return Opt[Doc]{v, ok}
	})
}

// Set the version field.
//
// By default the library won't include any version info and won't accept `--version` switch.
//
// # Combinatoric usage
//
//	parser := cli.Short('s').Switch().
//	    Parser().
//	    ToOptions().
//	    Version("0.5.0")
//
// # Example
// ```console
// $ app --version
// Version: 0.5.0
// ```
func (self *OptionParser[T]) Version(version string) *OptionParser[T] {
	self.info.version = Valid(NewDoc(version))
	return self
}

// Set the description field
//
// Description field should be 1-2 lines long briefly explaining program purpose. If
// description field is present the library prints it right before the usage line.
//
// # Combinatoric usage
//
//	parser := cli.Short('s').Switch().
//	    Parser().
//	    ToOptions().
//	    Description("This is a description").
//	    Header("This is a header").
//	    Footer("This is a footer")
//
// # Example
//
// ```console
// This is a description
//
// Usage: [-s]
//
// # This is a header
//
// Available options:
//
//	-s
//	-h, --help     Prints help information
//	-V, --version  Prints version information
//
// This is a footer
// ```
func (self *OptionParser[T]) Description(descr string) *OptionParser[T] {
	self.info.descr = Valid(NewDoc(descr))
	return self
}

// Set the Header field
//
// The library displays the header between the usage line and a list of the available options in `--help` output.
//
// # Combinatoric usage
//
//	parser := cli.Short('s').Switch().
//	    Parser().
//	    ToOptions().
//	    Description("This is a description").
//	    Header("This is a Header").
//	    Footer("This is a footer")
//
// # Example
//
// ```console
// This is a description
//
// Usage: [-s]
//
// # This is a Header
//
// Available options:
//
//	-s
//	-h, --help     Prints help information
//	-V, --version  Prints version information
//
// This is a footer
// ```
func (self *OptionParser[T]) Header(header string) *OptionParser[T] {
	self.info.header = Valid(NewDoc(header))
	return self
}

// Set the Footer field
//
// The library displays the footer after the list of the available options in `--help` output.
//
// # Combinatoric usage
//
//	parser := cli.Short('s').Switch().
//	    Parser().
//	    ToOptions().
//	    Description("This is a description").
//	    Header("This is a header").
//	    Footer("This is a Footer")
//
// # Example
//
// ```console
// This is a description
//
// Usage: [-s]
//
// # This is a header
//
// Available options:
//
//	-s
//	-h, --help     Prints help information
//	-V, --version  Prints version information
//
// This is a Footer
// ```
func (self *OptionParser[T]) Footer(footer string) *OptionParser[T] {
	self.info.footer = Valid(NewDoc(footer))
	return self
}

// Set custom usage line
//
// Custom usage line to use instead of one derived by the library.
func (self *OptionParser[T]) Usage(usage Doc) *OptionParser[T] {
	self.info.usage = Valid(usage)
	return self
}

// Generate new usage line using automatically derived usage
//
// You can customize the surroundings of the usage line while still
// having the frequently changing part generated by the library.
func (self *OptionParser[T]) WithUsage(f func(Doc) Doc) *OptionParser[T] {
	doc := Doc{}
	doc.write_meta(self.inner.meta, true)
	self.info.usage = Valid(f(doc))
	return self
}

// Customize parser for `--help`
//
// By default the library displays help when the program is called with either `--help` or `-h`; you
// can customize those names and description in the help message.
//
// Note, `--help` is something the user expects to work
func (self *OptionParser[T]) HelpParser(parser NamedArg) *OptionParser[T] {
	self.info.help_arg = parser
	return self
}

// Customize parser for `--version`
//
// By default the library displays version information when the program is called with either `--version`
// or `-V` (and version is available); you can customize those names and description in the help message.
//
// Note, `--version` is something the user expects to work
func (self *OptionParser[T]) VersionParser(parser NamedArg) *OptionParser[T] {
	self.info.version_arg = parser
	return self
}

// Print help if the app was called with no parameters
//
// By default the library tries to parse the command line and displays the best possible
// error it can come up with. If the application requires a subcommand or some argument
// and the user specified none - it might be a better experience for the user to print
// the help message.
//
//	parser := cli.Short('a').Switch().
//	    Parser().
//	    ToOptions().
//	    FallbackToUsage()
func (self *OptionParser[T]) FallbackToUsage() *OptionParser[T] {
	self.info.help_if_no_args = true
	return self
}

// Set the width of the help message printed to the terminal upon failure
//
// By default, the help message is printed with a width of 100 characters.
// This method allows to change where the help message is wrapped.
//
// Setting the max width too low may negatively affect the readability of the help message.
// Also, the alignment padding of broken lines is always applied.
func (self *OptionParser[T]) MaxWidth(width uint) *OptionParser[T] {
	self.info.max_width = width
	return self
}
