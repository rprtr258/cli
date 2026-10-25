package cli

import "os"

// flagParser parses a named switch; created with [NamedArg.Switch], [Flag]
// or [ReqFlag].
//
// This is an unexported constructor shared by all flag parsers; the fluent
// combinators continue on the returned [Parser].
//
// # Panics
//
// Panics if the flag has no name: add at least one [NamedArg.Short],
// [NamedArg.Long] or [NamedArg.Env].
func flagParser[T any](present T, absent Opt[T], named NamedArg) Parser[T] {
	if _, ok := named.flag_item(); !ok && len(named.env) == 0 {
		panic("cli usage BUG: flag must have at least one name: use Short, Long or Env")
	}

	eval := func(state *state) (T, parseError) {
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
			missing := missingItem{
				item:     item,
				position: state.scope.start(),
				scope:    state.scope,
			}
			return *new(T), errorMissing([]missingItem{missing})
		} else if len(named.env) > 0 {
			return *new(T), errorNoEnv(named.env[0])
		}
		panic("unreachable") // nameless flags are rejected by flagParser
	}

	meta := meta(metaSkip{})
	if item, ok := named.flag_item(); ok {
		meta = item.required(!absent.Valid)
	}
	return Parser[T]{eval, meta}
}

// Flag with custom present/absent values
//
// More generic version of [NamedArg.Switch] that can use arbitrary type instead of
// [bool].
//
// # Panics
//
// Panics if the flag has no name: add at least one [NamedArg.Short],
// [NamedArg.Long] or [NamedArg.Env].
func (self NamedArg) Flag[T any](present, absent T) Parser[T] {
	return flagParser(present, Valid(absent), self)
}

// ReqFlag is a required flag with custom value.
//
// Similar to [Flag] takes no Opt arguments, but would only
// succeed if user specifies its name on a command line.
// Works best in combination with other parsers.
//
// # Panics
//
// Panics if the flag has no name: add at least one [NamedArg.Short],
// [NamedArg.Long] or [NamedArg.Env].
func (self NamedArg) ReqFlag[T any](present T) Parser[T] {
	return flagParser(present, Invalid[T](), self)
}
