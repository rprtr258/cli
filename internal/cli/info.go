package cli

// Information about the parser
//
// No longer public, users are only interacting with it via [OptionParser]
type Info struct {
	version               Option[Doc] // version field, see [version][Infoversion]
	descr                 Option[Doc] // Custom description field, see [descr][Infodescr]
	header                Option[Doc] // Custom header field, see [header][Infoheader]
	footer                Option[Doc] // Custom footer field, see [footer][Infofooter]
	usage                 Option[Doc] // Custom usage field, see [usage][Infousage]
	help_arg, version_arg NamedArg
	help_if_no_args       bool
	max_width             uint
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

func (self Info) clone() Info {
	return Info{
		self.version.clone(),
		self.descr.clone(),
		self.header.clone(),
		self.footer.clone(),
		self.usage.clone(),
		self.help_arg.clone(),
		self.version_arg,
		self.help_if_no_args,
		self.max_width,
	}
}

func (self Info) mk_help_parser() ParseFlag[struct{}] {
	return ReqFlag(self.help_arg, struct{}{})
}

func (self Info) mk_version_parser() ParseFlag[struct{}] {
	return ReqFlag(self.version_arg, struct{}{})
}

type ExtraParams struct {
	Help    bool
	Version Doc
}

// TODO: unused
func (self Info) Parser() Parser[ExtraParams] {
	return Parser[ExtraParams]{self.eval, self.meta()}
}

func (self Info) eval(state *State) (ExtraParams, Error) {
	help := self.mk_help_parser()
	if _, err := help.eval(state); err == nil {
		return ExtraParams{true, Doc{}}, nil
	}

	if version, ok := self.version.Unpack(); ok {
		if _, err := self.mk_version_parser().eval(state); err == nil {
			return ExtraParams{false, version}, nil
		}
	}

	// error message is not actually used anywhere
	return ExtraParams{}, ErrorParseFail("not a version or help")
}

func (self Info) meta() Meta {
	help := self.mk_help_parser().meta()
	if self.version.Valid {
		return MetaAnd{help, self.mk_version_parser().meta()}
	} else {
		return help
	}
}
