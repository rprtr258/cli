package cli

import (
	"os"

	"github.com/rprtr258/fun"
)

// Information about the parser
// /
// No longer public, users are only interacting with it via [OptionParser]
type Info struct {
	version         Option[Doc] // version field, see [version][Infoversion]
	descr           Option[Doc] // Custom description field, see [descr][Infodescr]
	header          Option[Doc] // Custom header field, see [header][Infoheader]
	footer          Option[Doc] // Custom footer field, see [footer][Infofooter]
	usage           Option[Doc] // Custom usage field, see [usage][Infousage]
	help_arg        NamedArg
	version_arg     NamedArg
	help_if_no_args bool
	max_width       uint
}

func NewInfo() Info {
	return Info{
		version:         Invalid[Doc](),
		descr:           Invalid[Doc](),
		header:          Invalid[Doc](),
		footer:          Invalid[Doc](),
		usage:           Invalid[Doc](),
		help_arg:        Short('h').Long("help").Help("Prints help information"),
		version_arg:     Short('V').Long("version").Help("Prints version information"),
		help_if_no_args: false,
		max_width:       100,
	}
}

// Ready to run [Parser] with additional information attached
//
// Created with [to_options](Parserto_options)
//
// In addition to the inner parser `OptionParser` contains documentation about a program or a
// subcommand as a whole, version, custom usage, if specified, and handles custom parsers for
// `--version` and `--help` flags.
type OptionParser[T any] struct {
	inner Parser[T]
	info  Info
}

// Execute the [OptionParser], extract a parsed value or print some diagnostic and exit
//
// # Usage
// ```rust
// # use bpaf*;
// // Parses number of repetitions of `-v` on a command line
//
//	fn verbosity()OptionParser[uint] {
//	    parser := short('v')
//	        .req_flag(struct{}{})
//	        .many()
//	        .map(|xs|xs.len());
//
//	    parser
//	        .to_options()
//	        .descr("Takes verbosity flag and does nothing else")
//	}
//
//	fn main() {
//	    let verbosity: uint = verbosity().Run();
//	}
//
// ```
func (self OptionParser[T]) Run() T {
	t, err := self.run_inner(current_args())
	if err != nil {
		print_message(err, self.info.max_width)
		os.Exit(exit_code(err))
	}
	return t
}

// Execute the [OptionParser] and produce a values for unit tests or manual processing
//
// ```rust
// # use bpaf*;
// # /*
// #[test]
// func positional_argument() {
// # */
//
//	parser :=
//	    positional<String>("FILE")
//	        .help("File to process")
//	        .to_options();
//
//	help := parser
//	    .run_inner(&["--help"])
//	    .unwrap_err()
//	    .unwrap_stdout();
//	expected_help := "\
//
// Usage: FILE
//
// Available positional items:
//
//	FILE        File to process
//
// Available fun.Options:
//
//	-h, --help  Prints help information
//
// ";
//
//	require.Equal(t,expected_help, help);
//
// # /*
// }
// # */
// ```
//
// See also [Args] and it's `From` impls to produce input and
// [ParseFailureunwrap_stderr] / [ParseFailureunwrap_stdout] for processing results.
//
// # Errors
//
// If parser can't produce desired result `run_inner` returns [ParseFailure]
// which represents runtime behavior: one branch to print something to stdout and exit with
// success and the other branch to print something to stderr and exit with failure.
//
// `bpaf` generates contents of this `ParseFailure` using expected textual output from
// [parse](Parserparse), stdout/stderr isn't actually captured.
//
// Exact string reperentations may change between versions including minor releases.
func (self OptionParser[T]) run_inner(args Args) (T, ParseFailure) {
	// prepare available short flags and arguments for disambiguation
	short_flags := []byte(nil)
	short_args := []byte(nil)
	collect_shorts(self.inner.meta, &short_flags, &short_args)
	short_flags = append(short_flags, self.info.help_arg.short...)
	short_flags = append(short_flags, self.info.version_arg.short...)
	err := Invalid[Message]()
	state := stateconstruct(args, short_flags, short_args, &err)

	// this only handles disambiguation failure in construct
	if msg, ok := err.Unpack(); ok && !func() bool { _, ok := state.comp_ref(); return ok }() {
		return *new(T), fun.Ptr(Message(msg)).render(state, self.inner.meta)
	}

	return self.run_subparser(&state)
}

func (self OptionParser[T]) run_subparser(args *State) (T, ParseFailure) {
	// process should work like this:
	// - inner parser is evaluated, it returns Error
	// - if error is finalized (ParseFailure) - it is simply propagated outwards,
	//   otherwise we are making a few attempts at improving it after dealing with
	//   autocomplete/help
	// - generate autocomplete, if enabled
	// - produce --help, --version
	// - Try to improve error message and finalize it otherwise
	//
	// outer parser gets value in ParseFailure format

	if self.info.help_if_no_args && args.is_empty() {
		buffer := render_help(
			args.path,
			self.info,
			self.inner.meta,
			self.info.meta(),
			true,
		)
		return *new(T), ParseFailureStdout{buffer, false}
	}

	res, errr := self.inner.eval(args)
	if errr != nil {
		var parseFailure ParseFailure
		errr(nil, nil, nil, nil, nil, func(failure ParseFailure) {
			parseFailure = failure
		}, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil)
		if parseFailure != nil {
			return *new(T), parseFailure
		}
	}

	if comp, ok := args.check_complete(); ok {
		return *new(T), ParseFailureCompletion(comp)
	}

	var err Message
	if errr != nil {
		err = Message(errr)
	} else {
		if ix, _, ok := args.items_iter()(); ok {
			err = MessageUnconsumed(ix)
		} else {
			return res, nil
		}
	}

	// handle --help and --version messages
	if extra, ok := self.info.eval(args); ok == nil {
		detailed := false
		var buffer Doc
		extra(
			func(d bool) {
				detailed = d
				buffer = render_help(
					args.path,
					self.info,
					self.inner.meta,
					self.info.meta(),
					true,
				)
			},
			func(v Doc) {
				tmp := Doc{}
				tmp.token(TokenBlockStart(BlockBlock))
				tmp.Text("Version: ")
				tmp.Doc(v)
				tmp.token(TokenBlockEnd(BlockBlock))
				buffer = tmp
			},
		)
		return *new(T), ParseFailureStdout{buffer, detailed}
	}
	return *new(T), err.render(*args, self.inner.meta)
}

// Get first line of description if Available
//
// Used internally to avoid duplicating description for [command].
func (self OptionParser[T]) short_descr() Option[Doc] {
	return self.info.descr.and_then(func(d Doc) Option[Doc] {
		v, ok := Doc.first_line(d)
		return Option[Doc]{v, ok}
	})
}

// Set the version field.
//
// By default `bpaf` won't include any version info and won't accept `--version` switch.
//
// # Combinatoric usage
//
// ```rust
// use bpaf*;
//
//	fn fun.Options()OptionParser<bool>  {
//	   short('s')
//	       .switch()
//	       .to_options()
//	       .version(env!("CARGO_PKG_VERSION"))
//	}
//
// ```
//
// # Derive usage
//
// `version` annotation is available after `options` and `command` annotations, takes
// an fun.Optional argument - version value to use, otherwise `bpaf_derive` would use value from cargo.
//
// ```rust
// # use bpaf*;
//
// #[bpaf(options, version)]
//
//	struct fun.Options {
//	    #[bpaf(short)]
//	    switch bool
//	}
//
// ```
//
// # Example
// ```console
// $ app --version
// Version: 0.5.0
// ```
func (self *OptionParser[T]) version(version Doc) *OptionParser[T] {
	self.info.version = Valid(version)
	return self
}

// Set the description field
//
// Description field should be 1-2 lines long briefly explaining program purpose. If
// description field is present `bpaf` would print it right before the usage line.
//
// # Combinatoric usage
// ```rust
// # use bpaf*;
//
//	fn fun.Options()OptionParser<bool>  {
//	   short('s')
//	       .switch()
//	       .to_options()
//	       .Descr("This is a description")
//	       .header("This is a header")
//	       .footer("This is a footer")
//	}
//
// ```
//
// # Derive usage
//
// `bpaf_derive` uses doc comments on the `struct` / `enum` to derive description, it skips single empty
// lines and uses double empty lines break it into blocks. `bpaf_derive` would use first block as the
// description, second block - header, third block - footer.
//
// ```rust
// # use bpaf*;
//
// #[bpaf(options, version)]
// // This is a description
// //
// //
// // This is a header
// //
// //
// // This is a footer
// //
// //
// // This is just a comment
//
//	struct fun.Options {
//	    #[bpaf(short)]
//	    switch bool
//	}
//
// ```
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
// Available fun.Options:
//
//	-s
//	-h, --help     Prints help information
//	-V, --version  Prints version information
//
// This is a footer
// ```
func (self *OptionParser[T]) Descr(descr string) *OptionParser[T] {
	self.info.descr = Valid(newDocFromString(descr))
	return self
}

// Set the Header field
//
// `bpaf` displays the Header between the usage line and a list of the available fun.Options in `--help` output
//
// # Combinatoric usage
// ```rust
// # use bpaf*;
//
//	fn fun.Options()OptionParser<bool>  {
//	   short('s')
//	       .switch()
//	       .to_options()
//	       .descr("This is a description")
//	       .Header("This is a Header")
//	       .footer("This is a footer")
//	}
//
// ```
//
// # Derive usage
//
// `bpaf_derive` uses doc comments on the `struct` / `enum` to derive description, it skips single empty
// lines and uses double empty lines break it into blocks. `bpaf_derive` would use first block as the
// description, second block - Header, third block - footer.
//
// ```rust
// # use bpaf*;
//
// #[bpaf(options, version)]
// // This is a description
// //
// //
// // This is a Header
// //
// //
// // This is a footer
// //
// //
// // This is just a comment
//
//	struct fun.Options {
//	    #[bpaf(short)]
//	    switch bool
//	}
//
// ```
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
// Available fun.Options:
//
//	-s
//	-h, --help     Prints help information
//	-V, --version  Prints version information
//
// This is a footer
// ```
func (self *OptionParser[T]) Header(header string) *OptionParser[T] {
	self.info.header = Valid(newDocFromString(header))
	return self
}

// Set the Footer field
//
// `bpaf` displays the Footer after list of the available fun.Options in `--help` output
//
// # Combinatoric usage
// ```rust
// # use bpaf*;
//
//	fn fun.Options()OptionParser<bool>  {
//	   short('s')
//	       .switch()
//	       .to_options()
//	       .descr("This is a description")
//	       .header("This is a header")
//	       .Footer("This is a Footer")
//	}
//
// ```
//
// # Derive usage
//
// `bpaf_derive` uses doc comments on the `struct` / `enum` to derive description, it skips single empty
// lines and uses double empty lines break it into blocks. `bpaf_derive` would use first block as the
// description, second block - header, third block - Footer.
//
// ```rust
// # use bpaf*;
//
// #[bpaf(options, version)]
// // This is a description
// //
// //
// // This is a header
// //
// //
// // This is a Footer
// //
// //
// // This is just a comment
//
//	struct fun.Options {
//	    #[bpaf(short)]
//	    switch bool
//	}
//
// ```
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
// Available fun.Options:
//
//	-s
//	-h, --help     Prints help information
//	-V, --version  Prints version information
//
// This is a Footer
// ```
func (self *OptionParser[T]) Footer(footer string) *OptionParser[T] {
	self.info.footer = Valid(newDocFromString(footer))
	return self
}

// Set custom Usage field
//
// Custom Usage field to use instead of one derived by `bpaf`.
func (self *OptionParser[T]) Usage(usage Doc) *OptionParser[T] {
	self.info.usage = Valid(usage)
	return self
}

// Generate new usage line using automatically derived usage
//
// You can customize the surroundings of the usage line while still
// having part that frequently changes generated by bpaf
//
// At the moment this method is not directly supported by derive API,
// but since it gives you an object of [`OptionParser[T]`](OptionParser)
// type you can alter it using Combinatoric API:
func (self *OptionParser[T]) WithUsage(f func(Doc) Doc) *OptionParser[T] {
	buf := Doc{}
	buf.write_meta(self.inner.meta, true)
	self.info.usage = Valid(f(buf))
	return self
}

// Check the invariants `bpaf` relies on for normal operations
//
// Takes a parameter whether to check for cosmetic invariants or not
// (max help width exceeding 120 symbols, etc), currently not in use
//
// Best used as part of your test suite:
// ```rust
// # use bpaf*;
// #[test]
// func check_options() {
// # fun.Options := || short('p').switch().to_options();
//
//	    fun.Options().check_invariants(false)
//	}
//
// ```
//
// # Panics
//
// `check_invariants` indicates problems with panic
func (self OptionParser[T]) check_invariants(bool) {
	positional_invariant_check(self.inner.meta, true)
}

// Customize parser for `--help`
//
// By default `bpaf` displays help when program is called with either `--help` or `-h`, you
// can customize those names and description in the help message
//
// Note, `--help` is something user expects to work
func (self *OptionParser[T]) help_parser(parser NamedArg) *OptionParser[T] {
	self.info.help_arg = parser
	return self
}

// Customize parser for `--version`
//
// By default `bpaf` displays version information when program is called with either `--version`
// or `-V` (and version is available), you can customize those names and description in the help message
//
// Note, `--version` is something user expects to work
func (self *OptionParser[T]) version_parser(parser NamedArg) *OptionParser[T] {
	self.info.version_arg = parser
	return self
}

// Print help if app was called with no parameters
//
// By default `bpaf` tries to parse command line fun.Options and displays the best possible
// error it can come up with. If application requires a subcommand or some argument
// and user specified Invalid() - it might be a better experience for user to print
// the help message.
//
// ```rust
// # use bpaf*;
// # func fun.Options()OptionParser<bool> { short('a').switch().to_options() }
// // create fun.Option parser in a usual way, derive or combinatoric API
// opts := fun.Options().FallbackToUsage().run();
// ```
//
// For derive macro you can specify `FallbackToUsage` in top level annotations
// for fun.Options and for individual commands if fallback to useage is the desired behavior:
//
// ```ignore
//
//	enum Commands {
//	    #[bpaf(command, FallbackToUsage)]
//	    Action {
//	        ...
//	    }
//	}
//
// ```
//
// # Or
//
// ```ignore
//
// #[bpaf(options, FallbackToUsage)]
//
//	struct fun.Options {
//	    ...
//	}
//
//	fn main() {
//	    fun.Options := fun.Options().run(); // falls back to usage
//	}
//
// ```
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
func (self *OptionParser[T]) max_width(width uint) *OptionParser[T] {
	self.info.max_width = width
	return self
}

func (self Info) mk_help_parser() ParseFlag[struct{}] {
	return ReqFlag(self.help_arg, struct{}{})
}

func (self Info) mk_version_parser() ParseFlag[struct{}] {
	return ReqFlag(self.version_arg, struct{}{})
}

func (self Info) eval(args *State) (ExtraParams, Error) {
	help := self.mk_help_parser()
	if _, err := help.eval(args); err == nil {
		return ExtraParamsHelp(true), nil
	}

	if version, ok := self.version.Unpack(); ok {
		if _, err := self.mk_version_parser().eval(args); err == nil {
			return ExtraParamsVersion(version), nil
		}
	}

	// error message is not actually used anywhere
	return *new(ExtraParams), Error(MessageParseFail("not a version or help"))
}

func (self Info) meta() Meta {
	help := self.mk_help_parser().meta()
	if self.version.Valid {
		return MetaAnd{help, self.mk_version_parser().meta()}
	} else {
		return help
	}
}

type ExtraParams = func(
	Help func(bool),
	Version func(Doc),
)

func ExtraParamsHelp(b bool) ExtraParams {
	return func(Help func(bool), _ func(Doc)) {
		if Help != nil {
			Help(b)
		}
	}
}

func ExtraParamsVersion(v Doc) ExtraParams {
	return func(_ func(bool), Version func(Doc)) {
		if Version != nil {
			Version(v)
		}
	}
}
