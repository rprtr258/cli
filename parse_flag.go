package cli

import "os"

// Parser for a named switch, created with [NamedArg.Switch], [Flag] or [ReqFlag].
//
// This is an unexported constructor shared by all flag parsers; the fluent
// combinators continue on the returned [Parser].
func flagParser[T any](present T, absent Option[T], named NamedArg) Parser[T] {
	eval := func(state *State) (T, Error) {
		if state.takeBoolFlag(named) || func() bool {
			_, ok := find_map(named.env, os.LookupEnv)
			return ok
		}() {
			if state.touching_last_remove() {
				state.push_flag(named)
			}
			return present, nil
		}
		state.push_flag(named)
		if ok, okk := absent.Unpack(); okk {
			return ok, nil
		}

		if item, ok := named.flag_item(); ok {
			missing := MissingItem{
				item:     item,
				position: state.scope.start(),
				scope:    state.scope,
			}
			return *new(T), ErrorMissing([]MissingItem{missing})
		} else if len(named.env) > 0 {
			return *new(T), ErrorNoEnv(named.env[0])
		} else {
			panic(`todo!("no key!")`)
		}
	}

	meta := Meta(MetaSkip{})
	if item, ok := named.flag_item(); ok {
		meta = item.required(!absent.Valid)
	}
	return Parser[T]{eval, meta}
}

// Flag with custom present/absent values
//
// More generic version of [NamedArg.Switch] that can use arbitrary type instead of
// [bool].
func Flag[T any](self NamedArg, present, absent T) Parser[T] {
	return flagParser(present, Valid(absent), self)
}

// Required flag with custom value
//
// Similar to [Flag] takes no Option arguments, but would only
// succeed if user specifies its name on a command line.
// Works best in combination with other parsers.
func ReqFlag[T any](self NamedArg, present T) Parser[T] {
	return flagParser(present, Invalid[T](), self)
}
