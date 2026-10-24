package cli

import (
	"encoding"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"
)

// Count is a field type for [Construct]: the field holds the number of times the
// associated flag was repeated on the command line (`-v -v -v` and `-vvv` both give 3).
// A Count field requires a short or a long name.
// Count is a marker type for [Construct] fields parsed as the number of times
// the flag was repeated, e.g. Verbose cli.Count with a "short=v,long=verbose" tag.
type Count int

var (
	countType           = reflect.TypeFor[Count]()
	cliPkgPath          = reflect.TypeFor[Opt[int]]().PkgPath()
	boolType            = reflect.TypeFor[bool]()
	textUnmarshalerType = reflect.TypeFor[encoding.TextUnmarshaler]()
)

// isOptionType checks whether t is an instantiated cli.Opt[E]. Note, reflect
// reports instantiated generic type names like "Opt[float64]".
func isOptionType(t reflect.Type) bool {
	return t.Kind() == reflect.Struct &&
		t.PkgPath() == cliPkgPath &&
		strings.HasPrefix(t.Name(), "Opt[") &&
		t.NumField() == 2 &&
		t.Field(0).Name == "Value" && t.Field(1).Name == "Valid"
}

// fieldPart is a parser for a single struct field: `eval` returns a value assignable
// to the field, `meta` is its contribution to help/usage, `named` marks named
// (non-positional) parsers and `command` marks command fields - named fields are
// evaluated first, then commands, then positionals.
type fieldPart struct {
	ix      int
	named   bool
	command bool
	meta    meta
	eval    func(*state) (reflect.Value, parseError)
}

// Construct builds a parser for the struct `T` from `cli` struct tags - the Go
// counterpart of the compose-into-struct pattern. It lets you define the whole command
// line interface as one typed struct:
//
//	type Out struct {
//	    Debug  bool           `cli:"short=d,long=debug,help='Activate debug mode'"`
//	    Speed  float64        `cli:"long=speed,metavar=SPEED,default=42"`
//	    Home   string         `cli:"long=home,env=HOME"`
//	    Files  []string       // positional, collected
//	    Verbose cli.Count     `cli:"short=v,long=verbose"`
//	}
//
//	out := cli.Construct[Out]().ToOptions().Run(os.Args)
//
// # Field types
//
// Every exported field becomes a parser chosen by the field type:
//
//   - bool -> switch, requires a short or a long name
//   - string, ints, uints, floats, time.Duration -> argument, or positional if the
//     field has no short/long/env
//   - types implementing [encoding.TextUnmarshaler] -> argument or positional,
//     parsed with UnmarshalText; this is the extension point for custom types
//   - []E -> zero or more of the element parser
//   - []*E -> zero or more of the element parser, elements are pointers to values
//   - *E -> optional element parser: nil when absent, pointer to the value otherwise
//   - cli.Opt[E] -> optional element parser
//   - map[K]V -> repeated `key:value` named argument accumulating into a map;
//     keys and values are parsed like scalar fields, entries split at the first
//     colon; requires a short or a long name
//   - struct with `command=NAME` -> subcommand: the inner struct fields are
//     parsed only when the command name is given; the field keeps its zero
//     value when the command is absent
//   - struct with `group=NAME` -> option group: the inner struct fields are
//     parsed as usual options of the parent but rendered in help under their
//     own `NAME` section
//   - cli.Count -> number of times the flag was repeated
//
// Nested wrappers ([]E, Opt[E]) over wrappers ([]byte is fine, [][]int or
// Opt[[]int] are not) are rejected. Struct-typed fields require either
// `command=` or `group=` - a nested struct without either panics.
//
// # Tags
//
// Tags are comma separated `key=value` pairs; double-quoted values may contain
// commas and standard Go escape sequences. Repeated keys accumulate: multiple
// `short`/`long`/`env` entries become hidden aliases, `default`/`choice`/`alias`
// collect values. Supported keys:
//
//   - short=a - short name, one character
//   - long=name - long name
//   - metavar=X - placeholder shown in help, defaults to the uppercased field name
//   - help=... - help message
//   - env=VAR - environment variable fallback
//   - default=v - value used when the field is absent; parsed with the same rules
//     as the field type and shown in help; not valid for bool, []E, *E, Opt[E] and
//     Count; on map fields repeatable, each entry is `key:value`
//   - positional - forces positional parsing; fields without short/long/env are
//     positional automatically
//   - command=NAME - struct field is a subcommand with the given name
//   - alias=A - hidden long alias for a command field, repeatable
//   - group=NAME - struct field options are rendered in help under their own section
//   - hidden - parse the field but omit it from help and completion
//   - choice=V - allowed argument values, repeatable; a value not listed fails
//     parsing with `couldn't parse `x`: expected one of ...`
//
// `cli:"-"` skips a field. Untagged exported fields participate as if the tag was
// empty; unexported fields are skipped.
//
// # Ordering
//
// Named fields are parsed first, then command fields, then positionals,
// regardless of declaration order, so the "positional and command items go last"
// invariant always holds.
//
// Construct panics on misconfiguration: invalid tag syntax, unknown tag key,
// unsupported field type, missing names where they are required.
func Construct[T any]() *ParseCon[T] {
	t := reflect.TypeFor[T]()
	if t.Kind() != reflect.Struct {
		panic(fmt.Sprintf("cli.Construct[%s]: type parameter must be a struct, got %s", t, t.Kind()))
	}

	meta, eval := constructStruct(t)
	return &ParseCon[T]{
		inner: func(_ bool, state *state) (T, parseError) {
			v, err := eval(state)
			if err != nil {
				return *new(T), err
			}
			return v.Interface().(T), nil
		},
		meta:     meta,
		failfast: false,
	}
}

// constructStruct is the reflection core of [Construct]: it builds the meta and
// the eval for a struct type, producing values of the exact type `t`.
func constructStruct(t reflect.Type) (m meta, evalFn func(*state) (reflect.Value, parseError)) {
	if t.Kind() != reflect.Struct {
		panic(fmt.Sprintf("cli.Construct: %s is not a struct", t))
	}

	var named, commands, positional []fieldPart
	for i := range t.NumField() {
		f := t.Field(i)
		tag, tagged := f.Tag.Lookup("cli")
		if tag == "-" {
			// explicitly skipped field, exported or not
			continue
		}
		if !f.IsExported() {
			if tagged && tag != "" {
				panic(fmt.Sprintf("cli.Construct: field %q: unexported field with a cli tag cannot be parsed, export it or drop the tag", f.Name))
			}
			continue
		}
		part := buildFieldPart(i, f.Name, f.Type, parseCliTag(tag))
		switch {
		case part.command:
			commands = append(commands, part)
		case part.named:
			named = append(named, part)
		default:
			positional = append(positional, part)
		}
	}

	parts := make([]fieldPart, 0, len(named)+len(commands)+len(positional))
	parts = append(parts, named...)
	parts = append(parts, commands...)
	parts = append(parts, positional...)
	metas := make([]meta, 0, len(parts))
	for _, p := range parts {
		metas = append(metas, p.meta)
	}
	var mt meta = metaSkip{}
	switch len(metas) {
	case 1:
		mt = metas[0]
	default:
		mt = metaAnd(metas)
	}

	eval := func(state *state) (reflect.Value, parseError) {
		out := reflect.New(t).Elem()
		var missing []missingItem
		for _, p := range parts {
			v, err := p.eval(state)
			if err != nil {
				if items, ok := err.(errorMissing); ok {
					missing = append(missing, items...)
					continue
				}
				return reflect.Value{}, err
			}
			out.Field(p.ix).Set(v)
		}
		if len(missing) > 0 {
			return reflect.Value{}, errorMissing(missing)
		}
		return out, nil
	}

	return mt, eval
}

// parseCliTag splits a tag value like `short=d,long=x,help='a, b'` into key/value
// pairs. Bare keys map to empty values, repeated keys accumulate into lists,
// double-quoted values may contain commas and Go escape sequences.
func parseCliTag(s string) map[string][]string {
	res := map[string][]string{}
	setKeyValue := func(key, val string) {
		if key == "" {
			panic("cli.Construct: empty tag key")
		}
		res[key] = append(res[key], val)
	}
	i := 0
	for i < len(s) {
		for i < len(s) && (s[i] == ' ' || s[i] == '\t') {
			i++
		}
		if i >= len(s) {
			break
		}
		j := i
		for j < len(s) && s[j] != '=' && s[j] != ',' {
			j++
		}
		key := strings.TrimSpace(s[i:j])
		if j >= len(s) || s[j] == ',' {
			setKeyValue(key, "")
			i = j + 1
			continue
		}
		i = j + 1 // skip '='
		if i < len(s) && (s[i] == '"' || s[i] == '\'') {
			end, err := scanQuoted(s, i)
			if err != nil {
				panic(fmt.Sprintf("cli.Construct: tag %q: %v", key, err))
			}
			val, err := unquoteTag(s[i:end])
			if err != nil {
				panic(fmt.Sprintf("cli.Construct: tag %q: %v", key, err))
			}
			setKeyValue(key, val)
			i = end
			if i < len(s) && s[i] != ',' {
				panic(fmt.Sprintf("cli.Construct: tag %q: expected `,` after quoted value", key))
			}
			i++
		} else {
			j = i
			for j < len(s) && s[j] != ',' {
				j++
			}
			setKeyValue(key, strings.TrimSpace(s[i:j]))
			i = j + 1
		}
	}
	return res
}

// scanQuoted returns the index right after the closing quote of the string
// starting at s[open] == '"' or '\”.
func scanQuoted(s string, open int) (int, error) {
	for i := open + 1; i < len(s); i++ {
		switch s[i] {
		case '\\':
			i++
		case '"', '\'':
			if s[i] == s[open] {
				return i + 1, nil
			}
		}
	}
	return 0, fmt.Errorf("unterminated quoted value")
}

// unquoteTag strips the surrounding quotes (single or double) and resolves
// backslash escapes: the backslash escapes the next character.
func unquoteTag(quoted string) (string, error) {
	if quoted[0] == '"' {
		return strconv.Unquote(quoted)
	}
	inner := quoted[1 : len(quoted)-1]
	var b strings.Builder
	for i := 0; i < len(inner); i++ {
		if inner[i] == '\\' && i+1 < len(inner) {
			i++
		}
		b.WriteByte(inner[i])
	}
	return b.String(), nil
}

// fieldCfg is the parsed `cli` tag of one field.
type fieldCfg struct {
	shorts     []byte
	longs      []string
	envs       []string
	help       Opt[string]
	metavar    Opt[string]
	defs       []string
	positional bool
	command    Opt[string]
	aliases    []string
	group      Opt[string]
	hidden     bool
	choices    []string
}

func buildFieldPart(ix int, name string, ft reflect.Type, tags map[string][]string) fieldPart {
	cfg, na, mv := parseFieldCfg(name, ft, tags)
	part := fieldPartFor(ix, name, ft, &cfg, na, mv)
	if cfg.hidden {
		part.meta = metaSkip{}
		part.eval = hideFieldEval(part.eval)
	}
	return part
}

// hideFieldEval wraps a field eval to hide it from help and completion,
// mirroring [Parser.Hide].
func hideFieldEval(inner func(*state) (reflect.Value, parseError)) func(*state) (reflect.Value, parseError) {
	return func(state *state) (reflect.Value, parseError) {
		var comps []comp
		state.swap_comps_with(&comps)
		v, err := inner(state)
		state.swap_comps_with(&comps)
		if err != nil {
			if _, ok := err.(errorMissing); ok {
				return reflect.Value{}, errorMissing(nil)
			}
		}
		return v, err
	}
}

func parseFieldCfg(name string, ft reflect.Type, tags map[string][]string) (fieldCfg, NamedArg, Metavar) {
	cfg := fieldCfg{}
	for key, vals := range tags {
		val := vals[0]
		nonEmpty := val != ""
		switch key {
		case "short":
			for _, s := range vals {
				if s == "" || len(s) != 1 {
					panic(fmt.Sprintf("cli.Construct: field %q: short name must be one character, got %q", name, s))
				}
				cfg.shorts = append(cfg.shorts, s[0])
			}
		case "long":
			for _, l := range vals {
				if l == "" {
					panic(fmt.Sprintf("cli.Construct: field %q: long name must not be empty", name))
				}
				cfg.longs = append(cfg.longs, l)
			}
		case "env":
			for _, e := range vals {
				if e == "" {
					panic(fmt.Sprintf("cli.Construct: field %q: env name must not be empty", name))
				}
				cfg.envs = append(cfg.envs, e)
			}
		case "help":
			if nonEmpty {
				cfg.help = Valid(val)
			}
		case "metavar":
			if !nonEmpty {
				panic(fmt.Sprintf("cli.Construct: field %q: metavar must not be empty", name))
			}
			cfg.metavar = Valid(val)
		case "default":
			for _, d := range vals {
				if d == "" {
					panic(fmt.Sprintf("cli.Construct: field %q: default value must not be empty", name))
				}
				cfg.defs = append(cfg.defs, d)
			}
		case "positional":
			if nonEmpty {
				panic(fmt.Sprintf("cli.Construct: field %q: `positional` does not take a value", name))
			}
			cfg.positional = true
		case "command":
			if !nonEmpty {
				panic(fmt.Sprintf("cli.Construct: field %q: command name must not be empty", name))
			}
			cfg.command = Valid(val)
		case "alias":
			for _, a := range vals {
				if a == "" {
					panic(fmt.Sprintf("cli.Construct: field %q: alias must not be empty", name))
				}
				cfg.aliases = append(cfg.aliases, a)
			}
		case "group":
			if !nonEmpty {
				panic(fmt.Sprintf("cli.Construct: field %q: group name must not be empty", name))
			}
			cfg.group = Valid(val)
		case "hidden":
			if nonEmpty {
				panic(fmt.Sprintf("cli.Construct: field %q: `hidden` does not take a value", name))
			}
			cfg.hidden = true
		case "choice":
			for _, c := range vals {
				if c == "" {
					panic(fmt.Sprintf("cli.Construct: field %q: choice must not be empty", name))
				}
				cfg.choices = append(cfg.choices, c)
			}
		default:
			panic(fmt.Sprintf("cli.Construct: field %q: unknown tag key %q", name, key))
		}
	}

	named := len(cfg.shorts) > 0 || len(cfg.longs) > 0 || len(cfg.envs) > 0
	if cfg.positional && named {
		panic(fmt.Sprintf("cli.Construct: field %q: field cannot be both positional and named", name))
	}
	if cfg.group.Valid && cfg.command.Valid {
		panic(fmt.Sprintf("cli.Construct: field %q: field cannot be both a command and a group", name))
	}
	if len(cfg.aliases) > 0 && !cfg.command.Valid {
		panic(fmt.Sprintf("cli.Construct: field %q: alias is only valid for command fields", name))
	}
	if cfg.command.Valid && cfg.positional {
		panic(fmt.Sprintf("cli.Construct: field %q: command field cannot be positional", name))
	}
	if (cfg.command.Valid || cfg.group.Valid) && (ft.Kind() != reflect.Struct || ft == countType || isOptionType(ft)) {
		panic(fmt.Sprintf("cli.Construct: field %q: command and group tags are only valid for struct fields, got %s", name, ft))
	}
	if cfg.metavar.Valid && (ft == countType || ft == boolType) {
		panic(fmt.Sprintf("cli.Construct: field %q: metavar is only valid for value-taking fields", name))
	}
	if len(cfg.defs) > 1 && ft.Kind() != reflect.Map {
		panic(fmt.Sprintf("cli.Construct: field %q: multiple default values are only valid for map fields", name))
	}

	mv := Metavar(strings.ToUpper(name))
	if cfg.metavar.Valid {
		mv = Metavar(cfg.metavar.Value)
	}
	na := NamedArg{}
	for _, s := range cfg.shorts {
		na = na.Short(s)
	}
	for _, l := range cfg.longs {
		na = na.Long(l)
	}
	for _, e := range cfg.envs {
		na = na.Env(e)
	}
	if cfg.help.Valid {
		na = na.Help(cfg.help.Value)
	}
	return cfg, na, mv
}

func fieldPartFor(ix int, name string, ft reflect.Type, cfg *fieldCfg, na NamedArg, mv Metavar) fieldPart {
	named := len(cfg.shorts) > 0 || len(cfg.longs) > 0 || len(cfg.envs) > 0

	// Count field: number of repetitions of the flag. ReqFlag fails when the flag
	// is absent so the count starts at 0, a plain Switch would always count 1.
	if ft == countType {
		if !named {
			panic(fmt.Sprintf("cli.Construct: field %q: cli.Count field requires a short or a long name", name))
		}
		if len(cfg.defs) > 0 {
			panic(fmt.Sprintf("cli.Construct: field %q: cli.Count field cannot have a default", name))
		}
		if len(cfg.choices) > 0 {
			panic(fmt.Sprintf("cli.Construct: field %q: choice is not valid for Count fields", name))
		}
		p := na.ReqFlag[struct{}](struct{}{}).Count()
		return fieldPart{ix: ix, named: true, meta: p.meta, eval: func(state *state) (reflect.Value, parseError) {
			n, err := p.eval(state)
			if err != nil {
				return reflect.Value{}, err
			}
			return reflect.ValueOf(Count(n)), nil
		}}
	}

	// Nested struct: a command (its own parser reachable by name) or a group
	// (options parsed as usual but rendered in their own help section).
	if ft.Kind() == reflect.Struct && !isOptionType(ft) {
		return structFieldPart(ix, name, ft, cfg, na, mv)
	}

	// Map field: a repeated `key:value` named argument accumulating into a map.
	if ft.Kind() == reflect.Map {
		return mapFieldPart(ix, name, ft, cfg, na, mv, named)
	}

	// Unwrap one optional container: []E, *E or cli.Opt[E].
	isSlice := ft.Kind() == reflect.Slice
	isOption := isOptionType(ft)
	isPtr := ft.Kind() == reflect.Pointer
	if len(cfg.defs) > 0 && (isSlice || isOption) {
		panic(fmt.Sprintf("cli.Construct: field %q: []E and cli.Opt[E] fields cannot have a default", name))
	}
	leaf := ft
	switch {
	case isSlice:
		leaf = ft.Elem()
	case isOption:
		leaf = ft.Field(0).Type
	case isPtr:
		leaf = ft.Elem()
	}
	// []*E: slice elements are pointers to values
	ptrElem := isSlice && leaf.Kind() == reflect.Pointer
	if ptrElem {
		leaf = leaf.Elem()
	}
	if isOption && (leaf.Kind() == reflect.Pointer || isOptionType(leaf)) {
		panic(fmt.Sprintf("cli.Construct: field %q: nested %s is not supported, use a single cli.Opt[E] wrapper", name, leaf))
	}
	if leaf.Kind() == reflect.Slice || leaf.Kind() == reflect.Pointer || isOptionType(leaf) {
		panic(fmt.Sprintf("cli.Construct: field %q: nested %s is not supported, use a single []E, []*E, *E or cli.Opt[E] wrapper", name, leaf))
	}
	if isPtr && (leaf == boolType || leaf == countType) {
		panic(fmt.Sprintf("cli.Construct: field %q: *bool and *cli.Count fields are not supported, use bool, cli.Count or cli.Opt[...]", name))
	}

	// bool leaf: a switch.
	if leaf == boolType {
		if !named {
			panic(fmt.Sprintf("cli.Construct: field %q: bool field requires a short or a long name", name))
		}
		if len(cfg.defs) > 0 {
			panic(fmt.Sprintf("cli.Construct: field %q: bool field cannot have a default", name))
		}
		if len(cfg.choices) > 0 {
			panic(fmt.Sprintf("cli.Construct: field %q: choice is not valid for bool fields", name))
		}
		p := na.Switch()
		return wrapTyped(ix, ft, p, named, isSlice, isOption)
	}

	// Scalar leaf: builtin converter or encoding.TextUnmarshaler.
	conv, ok := scalarConv(leaf)
	if !ok {
		if reflect.PointerTo(leaf).Implements(textUnmarshalerType) {
			conv = textConv(leaf)
		} else {
			panic(fmt.Sprintf("cli.Construct: field %q: unsupported field type %s, implement encoding.TextUnmarshaler to parse custom types", name, leaf))
		}
	}
	var p Parser[any]
	if named {
		pa := na.Argument[any](mv, conv)
		if len(cfg.choices) > 0 {
			pa = pa.Choice(cfg.choices...)
		}
		p = pa.Parser()
	} else {
		if len(cfg.choices) > 0 {
			panic(fmt.Sprintf("cli.Construct: field %q: choice is only valid for named fields", name))
		}
		pp := Positional[any](mv, conv)
		if cfg.help.Valid {
			pp = pp.Help(cfg.help.Value)
		}
		p = pp.Parser()
	}
	if len(cfg.defs) > 0 {
		dv, err := conv(cfg.defs[0])
		if err != nil {
			panic(fmt.Sprintf("cli.Construct: field %q: bad default value %q: %v", name, cfg.defs[0], err))
		}
		p = p.FallbackDisplay(dv)
	}
	box := reflect.ValueOf
	if ptrElem {
		box = func(v any) reflect.Value {
			ptr := reflect.New(leaf)
			ptr.Elem().Set(reflect.ValueOf(v))
			return ptr
		}
	}
	if isPtr {
		return fieldPart{ix: ix, named: named, meta: p.meta, eval: func(state *state) (reflect.Value, parseError) {
			v, err := p.eval(state)
			if err != nil {
				if _, ok := err.(errorMissing); ok {
					// absent optional pointer: the field stays nil
					return reflect.Zero(ft), nil
				}
				return reflect.Value{}, err
			}
			ptr := reflect.New(leaf)
			ptr.Elem().Set(reflect.ValueOf(v))
			return ptr, nil
		}}
	}
	return wrapAny(ix, ft, p, named, isSlice, isOption, box)
}

// structFieldPart builds a field part for a nested struct field: a subcommand
// (`command=NAME`) or an option group (`group=NAME`).
func structFieldPart(ix int, name string, ft reflect.Type, cfg *fieldCfg, na NamedArg, mv Metavar) fieldPart {
	if !cfg.command.Valid && !cfg.group.Valid {
		panic(fmt.Sprintf("cli.Construct: field %q: nested struct requires either a command=NAME or a group=NAME tag", name))
	}
	if cfg.metavar.Valid {
		panic(fmt.Sprintf("cli.Construct: field %q: metavar is not valid for struct fields", name))
	}
	if len(cfg.defs) > 0 {
		panic(fmt.Sprintf("cli.Construct: field %q: default is not valid for struct fields", name))
	}
	if cfg.positional {
		panic(fmt.Sprintf("cli.Construct: field %q: struct field cannot be positional, use command= or group= instead", name))
	}
	if len(cfg.choices) > 0 {
		panic(fmt.Sprintf("cli.Construct: field %q: choice is not valid for struct fields", name))
	}
	if len(cfg.envs) > 0 {
		panic(fmt.Sprintf("cli.Construct: field %q: env is not valid for struct fields", name))
	}

	meta, eval := constructStruct(ft)

	if cfg.command.Valid {
		cmd := Parser[reflect.Value]{eval, meta}.ToOptions().Command(cfg.command.Value)
		for _, s := range cfg.shorts {
			cmd = cmd.Short(s)
		}
		for _, alias := range cfg.aliases {
			cmd = cmd.Long(alias)
		}
		if cfg.help.Valid {
			cmd = cmd.Help(cfg.help.Value)
		}
		p := cmd.Parser()
		if cfg.hidden {
			p = p.Hide()
		}
		return fieldPart{
			ix: ix, named: true, command: true, meta: p.meta,
			eval: func(state *state) (reflect.Value, parseError) {
				v, err := p.eval(state)
				if err != nil {
					if _, ok := err.(errorMissing); ok {
						// command not taken: the field keeps its zero value
						return reflect.Zero(ft), nil
					}
					return reflect.Value{}, err
				}
				return v, nil
			},
		}
	}

	// group: the options are parsed as usual parent options but rendered in
	// help under their own section, same as [Parser.GroupHelp]
	p := Parser[reflect.Value]{eval, meta}.GroupHelp(cfg.group.Value)
	if cfg.hidden {
		p = p.Hide()
	}
	return fieldPart{ix: ix, named: true, meta: p.meta, eval: p.eval}
}

// mapFieldPart builds a field part for a map[K]V field: a repeated
// `key:value` named argument accumulating into a map.
func mapFieldPart(ix int, name string, ft reflect.Type, cfg *fieldCfg, na NamedArg, mv Metavar, named bool) fieldPart {
	if !named {
		panic(fmt.Sprintf("cli.Construct: field %q: map field requires a short or a long name", name))
	}
	if len(cfg.envs) > 0 {
		panic(fmt.Sprintf("cli.Construct: field %q: env is not valid for map fields", name))
	}
	if len(cfg.choices) > 0 {
		panic(fmt.Sprintf("cli.Construct: field %q: choice is not valid for map fields", name))
	}
	if !cfg.metavar.Valid {
		mv = "KEY:VALUE"
	}

	kt, vt := ft.Key(), ft.Elem()
	if !kt.Comparable() {
		panic(fmt.Sprintf("cli.Construct: field %q: map key type %s is not comparable", name, kt))
	}
	keyConv, ok := scalarConv(kt)
	if !ok {
		if !reflect.PointerTo(kt).Implements(textUnmarshalerType) {
			panic(fmt.Sprintf("cli.Construct: field %q: unsupported map key type %s, implement encoding.TextUnmarshaler to parse custom types", name, kt))
		}
		keyConv = textConv(kt)
	}
	valConv, ok := scalarConv(vt)
	if !ok {
		if !reflect.PointerTo(vt).Implements(textUnmarshalerType) {
			panic(fmt.Sprintf("cli.Construct: field %q: unsupported map value type %s, implement encoding.TextUnmarshaler to parse custom types", name, vt))
		}
		valConv = textConv(vt)
	}

	meta := meta(metaSkip{})
	if sl, ok := shortLongTryFrom(na); ok {
		meta = metaItem(item{na.help, itemArgument{sl, na.short, mv, first(na.env), nil}})
	}

	eval := func(state *state) (reflect.Value, parseError) {
		res := reflect.MakeMap(ft)
		for {
			valueStrOpt, err := state.takeFlag(na, false, mv)
			if err != nil {
				state.push_argument(na, mv)
				return reflect.Value{}, err
			}

			valueStr, ok := valueStrOpt.Unpack()
			if !ok {
				state.push_argument(na, mv)
				// the map flag is optional: absence gives the zero map when
				// nothing was collected, or the FallbackDisplay default map
				// when `default=` entries are given
				state.currentIdx = Invalid[int]()
				if res.Len() > 0 {
					return res, nil
				}
				if len(cfg.defs) > 0 {
					return reflect.Value{}, errorMissing(nil)
				}
				return reflect.Zero(ft), nil
			}

			if state.touching_last_remove() {
				state.push_metavar(mv, na.help, true)
			}

			if err := putMapEntry(state, res, valueStr, keyConv, valConv); err != nil {
				return reflect.Value{}, err
			}
		}
	}

	p := Parser[reflect.Value]{eval, meta}
	if len(cfg.defs) > 0 {
		def := reflect.MakeMap(ft)
		for _, entry := range cfg.defs {
			if err := putMapEntry(nil, def, entry, keyConv, valConv); err != nil {
				panic(fmt.Sprintf("cli.Construct: field %q: bad default value %q: %s", name, entry, err.(errorParseFailed).msg))
			}
		}
		p = p.FallbackDisplay(def)
	}
	return fieldPart{ix: ix, named: true, meta: p.meta, eval: p.eval}
}

// putMapEntry splits one `key:value` entry at the first colon, parses the key
// and the value with the given converters and puts it into the map.
func putMapEntry(state *state, res reflect.Value, entry string, keyConv, valConv func(string) (any, error)) parseError {
	pos := Invalid[int]()
	if state != nil {
		pos = state.currentIdx
	}
	rawKey, rawValue, ok := strings.Cut(entry, ":")
	if !ok {
		return errorParseFailed{pos, "expected `key:value`"}
	}

	k, err := keyConv(rawKey)
	if err != nil {
		return errorParseFailed{pos, err.Error()}
	}
	v, err := valConv(rawValue)
	if err != nil {
		return errorParseFailed{pos, err.Error()}
	}
	res.SetMapIndex(reflect.ValueOf(k), reflect.ValueOf(v))
	return nil
}

// wrapTyped wraps a parser whose result type matches the field type exactly
// (bool switch); for []bool and cli.Opt[bool] fields it applies the wrapper.
func wrapTyped[T any](ix int, ft reflect.Type, p Parser[T], named, isSlice, isOption bool) fieldPart {
	wrapEval := func(inner func(*state) (reflect.Value, parseError), meta meta) fieldPart {
		return fieldPart{ix: ix, named: named, meta: meta, eval: inner}
	}
	pe := func(state *state) (reflect.Value, parseError) {
		v, err := p.eval(state)
		if err != nil {
			return reflect.Value{}, err
		}
		return reflect.ValueOf(v), nil
	}
	switch {
	case isSlice:
		pm := p.Many()
		return wrapEval(func(state *state) (reflect.Value, parseError) {
			xs, err := pm.eval(state)
			if err != nil {
				return reflect.Value{}, err
			}
			res := reflect.MakeSlice(ft, len(xs), len(xs))
			for i, x := range xs {
				res.Index(i).Set(reflect.ValueOf(x))
			}
			return res, nil
		}, pm.meta)
	case isOption:
		pm := p.Optional()
		return wrapEval(func(state *state) (reflect.Value, parseError) {
			v, err := pm.eval(state)
			if err != nil {
				return reflect.Value{}, err
			}
			out := reflect.New(ft).Elem()
			if v.Valid {
				out.Field(0).Set(reflect.ValueOf(v.Value))
				out.Field(1).SetBool(true)
			}
			return out, nil
		}, pm.meta)
	default:
		return wrapEval(pe, p.meta)
	}
}

// wrapAny wraps a Parser[any] whose dynamic values have the field's element type
// and materializes the field value: plain, slice or cli.Opt.
func wrapAny(ix int, ft reflect.Type, p Parser[any], named, isSlice, isOption bool, box func(any) reflect.Value) fieldPart {
	pe := func(state *state) (reflect.Value, parseError) {
		v, err := p.eval(state)
		if err != nil {
			return reflect.Value{}, err
		}
		return box(v), nil
	}
	switch {
	case isSlice:
		pm := p.Many()
		return fieldPart{ix: ix, named: named, meta: pm.meta, eval: func(state *state) (reflect.Value, parseError) {
			xs, err := pm.eval(state)
			if err != nil {
				return reflect.Value{}, err
			}
			res := reflect.MakeSlice(ft, len(xs), len(xs))
			for i, x := range xs {
				res.Index(i).Set(box(x))
			}
			return res, nil
		}}
	case isOption:
		pm := p.Optional()
		return fieldPart{ix: ix, named: named, meta: pm.meta, eval: func(state *state) (reflect.Value, parseError) {
			v, err := pm.eval(state)
			if err != nil {
				return reflect.Value{}, err
			}
			out := reflect.New(ft).Elem()
			if v.Valid {
				out.Field(0).Set(reflect.ValueOf(v.Value))
				out.Field(1).SetBool(true)
			}
			return out, nil
		}}
	default:
		return fieldPart{ix: ix, named: named, meta: p.meta, eval: pe}
	}
}

// scalarConv returns a converter for builtin scalar types; values come back
// dynamically typed but with the exact static type of t.
func scalarConv(t reflect.Type) (func(string) (any, error), bool) {
	switch t {
	case reflect.TypeFor[string]():
		return func(s string) (any, error) { return s, nil }, true
	case reflect.TypeFor[int]():
		return func(s string) (any, error) { v, err := strconv.Atoi(s); return v, unwrapNumError(err) }, true
	case reflect.TypeFor[int8]():
		return func(s string) (any, error) { v, err := strconv.ParseInt(s, 10, 8); return int8(v), unwrapNumError(err) }, true
	case reflect.TypeFor[int16]():
		return func(s string) (any, error) {
			v, err := strconv.ParseInt(s, 10, 16)
			return int16(v), unwrapNumError(err)
		}, true
	case reflect.TypeFor[int32]():
		return func(s string) (any, error) {
			v, err := strconv.ParseInt(s, 10, 32)
			return int32(v), unwrapNumError(err)
		}, true
	case reflect.TypeFor[int64]():
		return func(s string) (any, error) { v, err := strconv.ParseInt(s, 10, 64); return v, unwrapNumError(err) }, true
	case reflect.TypeFor[uint]():
		return func(s string) (any, error) {
			v, err := strconv.ParseUint(s, 10, strconv.IntSize)
			return uint(v), unwrapNumError(err)
		}, true
	case reflect.TypeFor[uint8]():
		return func(s string) (any, error) {
			v, err := strconv.ParseUint(s, 10, 8)
			return uint8(v), unwrapNumError(err)
		}, true
	case reflect.TypeFor[uint16]():
		return func(s string) (any, error) {
			v, err := strconv.ParseUint(s, 10, 16)
			return uint16(v), unwrapNumError(err)
		}, true
	case reflect.TypeFor[uint32]():
		return func(s string) (any, error) {
			v, err := strconv.ParseUint(s, 10, 32)
			return uint32(v), unwrapNumError(err)
		}, true
	case reflect.TypeFor[uint64]():
		return func(s string) (any, error) { v, err := strconv.ParseUint(s, 10, 64); return v, unwrapNumError(err) }, true
	case reflect.TypeFor[float32]():
		return func(s string) (any, error) {
			v, err := strconv.ParseFloat(s, 32)
			return float32(v), unwrapNumError(err)
		}, true
	case reflect.TypeFor[float64]():
		return func(s string) (any, error) { v, err := strconv.ParseFloat(s, 64); return v, unwrapNumError(err) }, true
	case reflect.TypeFor[time.Duration]():
		return func(s string) (any, error) { return time.ParseDuration(s) }, true
	}
	return nil, false
}

// textConv converts via encoding.TextUnmarshaler on the leaf type.
func textConv(t reflect.Type) func(string) (any, error) {
	return func(s string) (any, error) {
		v := reflect.New(t).Elem()
		if err := v.Addr().Interface().(encoding.TextUnmarshaler).UnmarshalText([]byte(s)); err != nil {
			return nil, err
		}
		return v.Interface(), nil
	}
}
