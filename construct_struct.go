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
// (non-positional) parsers - they must be evaluated before positionals.
type fieldPart struct {
	ix    int
	named bool
	meta  meta
	eval  func(*state) (reflect.Value, parseError)
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
//   - cli.Opt[E] -> optional element parser
//   - cli.Count -> number of times the flag was repeated
//
// Nested wrappers ([]E, Opt[E]) over wrappers ([]byte is fine, [][]int or
// Opt[[]int] are not) are rejected.
//
// # Tags
//
// Tags are comma separated `key=value` pairs; double-quoted values may contain
// commas and standard Go escape sequences. Supported keys:
//
//   - short=a - short name, one character
//   - long=name - long name
//   - metavar=X - placeholder shown in help, defaults to the uppercased field name
//   - help=... - help message
//   - env=VAR - environment variable fallback
//   - default=v - value used when the field is absent; parsed with the same rules
//     as the field type and shown in help; not valid for bool, []E, Opt[E] and Count
//   - positional - forces positional parsing; fields without short/long/env are
//     positional automatically
//
// `cli:"-"` skips a field. Untagged exported fields participate as if the tag was
// empty; unexported fields are skipped.
//
// # Ordering
//
// Named fields are parsed before positionals regardless of declaration order, so
// the "positional items go last" invariant always holds.
//
// Construct panics on misconfiguration: invalid tag syntax, unknown tag key,
// unsupported field type, missing names where they are required.
func Construct[T any]() *ParseCon[T] {
	t := reflect.TypeFor[T]()
	if t.Kind() != reflect.Struct {
		panic(fmt.Sprintf("cli.Construct[%s]: type parameter must be a struct, got %s", t, t.Kind()))
	}

	var named, positional []fieldPart
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
		if part.named {
			named = append(named, part)
		} else {
			positional = append(positional, part)
		}
	}

	parts := make([]fieldPart, 0, len(named)+len(positional))
	parts = append(parts, named...)
	parts = append(parts, positional...)
	metas := make([]meta, 0, len(parts))
	for _, p := range parts {
		metas = append(metas, p.meta)
	}
	var meta meta = metaSkip{}
	switch len(metas) {
	case 1:
		meta = metas[0]
	default:
		meta = metaAnd(metas)
	}

	eval := func(_ bool, state *state) (T, parseError) {
		out := reflect.New(t).Elem()
		var missing []missingItem
		for _, p := range parts {
			v, err := p.eval(state)
			if err != nil {
				if items, ok := err.(errorMissing); ok {
					missing = append(missing, items...)
					continue
				}
				return *new(T), err
			}
			out.Field(p.ix).Set(v)
		}
		if len(missing) > 0 {
			return *new(T), errorMissing(missing)
		}
		return out.Interface().(T), nil
	}

	return &ParseCon[T]{inner: eval, meta: meta, failfast: false}
}

// parseCliTag splits a tag value like `short=d,long=x,help='a, b'` into key/value
// pairs. Bare keys map to empty values, double-quoted values may contain commas
// and Go escape sequences.
func parseCliTag(s string) map[string]string {
	res := map[string]string{}
	setKeyValue := func(key, val string) {
		if key == "" {
			panic("cli.Construct: empty tag key")
		}
		if _, dup := res[key]; dup {
			panic(fmt.Sprintf("cli.Construct: duplicate tag key %q", key))
		}
		res[key] = val
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
	short      Opt[byte]
	long       Opt[string]
	env        Opt[string]
	help       Opt[string]
	metavar    Opt[string]
	def        Opt[string]
	positional bool
}

func buildFieldPart(ix int, name string, ft reflect.Type, tags map[string]string) fieldPart {
	cfg := fieldCfg{}
	for key, val := range tags {
		nonEmpty := val != ""
		switch key {
		case "short":
			if !nonEmpty || len(val) != 1 {
				panic(fmt.Sprintf("cli.Construct: field %q: short name must be one character, got %q", name, val))
			}
			cfg.short = Valid(val[0])
		case "long":
			if !nonEmpty {
				panic(fmt.Sprintf("cli.Construct: field %q: long name must not be empty", name))
			}
			cfg.long = Valid(val)
		case "env":
			if !nonEmpty {
				panic(fmt.Sprintf("cli.Construct: field %q: env name must not be empty", name))
			}
			cfg.env = Valid(val)
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
			if !nonEmpty {
				panic(fmt.Sprintf("cli.Construct: field %q: default value must not be empty", name))
			}
			cfg.def = Valid(val)
		case "positional":
			if nonEmpty {
				panic(fmt.Sprintf("cli.Construct: field %q: `positional` does not take a value", name))
			}
			cfg.positional = true
		default:
			panic(fmt.Sprintf("cli.Construct: field %q: unknown tag key %q", name, key))
		}
	}

	named := cfg.short.Valid || cfg.long.Valid || cfg.env.Valid
	if cfg.positional && named {
		panic(fmt.Sprintf("cli.Construct: field %q: field cannot be both positional and named", name))
	}
	if cfg.metavar.Valid && (ft == countType || ft == boolType) {
		panic(fmt.Sprintf("cli.Construct: field %q: metavar is only valid for value-taking fields", name))
	}

	mv := Metavar(strings.ToUpper(name))
	if cfg.metavar.Valid {
		mv = Metavar(cfg.metavar.Value)
	}
	na := NamedArg{}
	if cfg.short.Valid {
		na = na.Short(cfg.short.Value)
	}
	if cfg.long.Valid {
		na = na.Long(cfg.long.Value)
	}
	if cfg.env.Valid {
		na = na.Env(cfg.env.Value)
	}
	if cfg.help.Valid {
		na = na.Help(cfg.help.Value)
	}

	// Count field: number of repetitions of the flag. ReqFlag fails when the flag
	// is absent so the count starts at 0, a plain Switch would always count 1.
	if ft == countType {
		if !named {
			panic(fmt.Sprintf("cli.Construct: field %q: cli.Count field requires a short or a long name", name))
		}
		if cfg.def.Valid {
			panic(fmt.Sprintf("cli.Construct: field %q: cli.Count field cannot have a default", name))
		}
		p := ReqFlag[struct{}](na, struct{}{}).Count()
		return fieldPart{ix: ix, named: true, meta: p.meta, eval: func(state *state) (reflect.Value, parseError) {
			n, err := p.eval(state)
			if err != nil {
				return reflect.Value{}, err
			}
			return reflect.ValueOf(Count(n)), nil
		}}
	}

	// Unwrap one optional container: []E or cli.Opt[E].
	isSlice := ft.Kind() == reflect.Slice
	isOption := isOptionType(ft)
	if cfg.def.Valid && (isSlice || isOption) {
		panic(fmt.Sprintf("cli.Construct: field %q: []E and cli.Opt[E] fields cannot have a default", name))
	}
	leaf := ft
	if isSlice {
		leaf = ft.Elem()
	} else if isOption {
		leaf = ft.Field(0).Type
	}
	if leaf.Kind() == reflect.Slice || isOptionType(leaf) {
		panic(fmt.Sprintf("cli.Construct: field %q: nested %s is not supported, use a single []E or cli.Opt[E] wrapper", name, leaf))
	}

	// bool leaf: a switch.
	if leaf == boolType {
		if !named {
			panic(fmt.Sprintf("cli.Construct: field %q: bool field requires a short or a long name", name))
		}
		if cfg.def.Valid {
			panic(fmt.Sprintf("cli.Construct: field %q: bool field cannot have a default", name))
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
		p = Argument[any](na, mv, conv).Parser()
	} else {
		pp := Positional[any](mv, conv)
		if cfg.help.Valid {
			pp = pp.Help(cfg.help.Value)
		}
		p = pp.Parser()
	}
	if cfg.def.Valid {
		dv, err := conv(cfg.def.Value)
		if err != nil {
			panic(fmt.Sprintf("cli.Construct: field %q: bad default value %q: %v", name, cfg.def.Value, err))
		}
		p = p.FallbackDisplay(dv)
	}
	return wrapAny(ix, ft, p, named, isSlice, isOption)
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
func wrapAny(ix int, ft reflect.Type, p Parser[any], named, isSlice, isOption bool) fieldPart {
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
		return fieldPart{ix: ix, named: named, meta: pm.meta, eval: func(state *state) (reflect.Value, parseError) {
			xs, err := pm.eval(state)
			if err != nil {
				return reflect.Value{}, err
			}
			res := reflect.MakeSlice(ft, len(xs), len(xs))
			for i, x := range xs {
				res.Index(i).Set(reflect.ValueOf(x))
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
		return func(s string) (any, error) { return strconv.Atoi(s) }, true
	case reflect.TypeFor[int8]():
		return func(s string) (any, error) { v, err := strconv.ParseInt(s, 10, 8); return int8(v), err }, true
	case reflect.TypeFor[int16]():
		return func(s string) (any, error) { v, err := strconv.ParseInt(s, 10, 16); return int16(v), err }, true
	case reflect.TypeFor[int32]():
		return func(s string) (any, error) { v, err := strconv.ParseInt(s, 10, 32); return int32(v), err }, true
	case reflect.TypeFor[int64]():
		return func(s string) (any, error) { return strconv.ParseInt(s, 10, 64) }, true
	case reflect.TypeFor[uint]():
		return func(s string) (any, error) { v, err := strconv.ParseUint(s, 10, strconv.IntSize); return uint(v), err }, true
	case reflect.TypeFor[uint8]():
		return func(s string) (any, error) { v, err := strconv.ParseUint(s, 10, 8); return uint8(v), err }, true
	case reflect.TypeFor[uint16]():
		return func(s string) (any, error) { v, err := strconv.ParseUint(s, 10, 16); return uint16(v), err }, true
	case reflect.TypeFor[uint32]():
		return func(s string) (any, error) { v, err := strconv.ParseUint(s, 10, 32); return uint32(v), err }, true
	case reflect.TypeFor[uint64]():
		return func(s string) (any, error) { return strconv.ParseUint(s, 10, 64) }, true
	case reflect.TypeFor[float32]():
		return func(s string) (any, error) { v, err := strconv.ParseFloat(s, 32); return float32(v), err }, true
	case reflect.TypeFor[float64]():
		return func(s string) (any, error) { return strconv.ParseFloat(s, 64) }, true
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
