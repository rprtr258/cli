package cli

import "slices"

// info holds information about the parser.
//
// No longer public: users only interact with it via [OptionParser].
type info struct {
	version               Opt[Doc]  // version field, see [version][Infoversion]
	descr                 Opt[Doc]  // Custom description field, see [descr][Infodescr]
	header                Opt[Doc]  // Custom header field, see [header][Infoheader]
	footer                Opt[Doc]  // Custom footer field, see [footer][Infofooter]
	usage                 Opt[Doc]  // Custom usage field, see [usage][Infousage]
	examples              []example // Usage examples, see [Parser.Examples]
	help_arg, version_arg NamedArg
	help_if_no_args       bool
	max_width             uint
}

func newInfo() info {
	return info{
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

func (self *info) clone() info {
	return info{
		self.version.clone(),
		self.descr.clone(),
		self.header.clone(),
		self.footer.clone(),
		self.usage.clone(),
		slices.Clone(self.examples),
		self.help_arg.clone(),
		self.version_arg,
		self.help_if_no_args,
		self.max_width,
	}
}

func (self *info) mk_help_parser() Parser[struct{}] {
	return self.help_arg.ReqFlag(struct{}{})
}

func (self *info) mk_version_parser() Parser[struct{}] {
	return self.version_arg.ReqFlag(struct{}{})
}

type extraParams struct {
	Help    bool
	Version Doc
}

func (self *info) eval(state *state) (extraParams, parseError) {
	help := self.mk_help_parser()
	if _, err := help.eval(state); err == nil {
		return extraParams{true, Doc{}}, nil
	}

	if version, ok := self.version.Unpack(); ok {
		if _, err := self.mk_version_parser().eval(state); err == nil {
			return extraParams{false, version}, nil
		}
	}

	// error message is not actually used anywhere
	return extraParams{}, errorParseFail("not a version or help")
}

func (self *info) meta() meta {
	help := self.mk_help_parser().meta
	if self.version.Valid {
		return metaAnd{help, self.mk_version_parser().meta}
	} else {
		return help
	}
}
