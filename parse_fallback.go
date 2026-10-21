package cli

import "fmt"

func fallbackMeta(innerMeta Meta, value_str string) Meta {
	m := Meta(MetaOptional{innerMeta})
	if value_str != "" {
		m = MetaSuffix{m, NewDoc(value_str)}
	}
	return m
}

// Parser that substitutes missing value but not parse failure, created with
// [Parser.Fallback] and [Parser.FallbackDisplay].
func fallback_eval[T any](inner Parser[T], value T, state *State) (T, Error) {
	ok, e := inner.eval(state)
	if e != nil {
		if can_catch(Error(e)) {
			return value, nil
		} else {
			return *new(T), Error(e)
		}
	}
	return ok, nil
}

// Parser that substitutes missing value with a function results but not parser
// failure, created with [Parser.FallbackWith] and [Parser.FallbackWithDisplay].
func fallback_with_eval[T any](inner Parser[T], fallback func() (T, Error), state *State) (T, Error) {
	clone := state.clone()
	ok, e := inner.eval(&clone)
	if e == nil {
		*state, clone = clone, *state
		return ok, nil
	} else {
		*state, clone = clone, *state
		if can_catch(Error(e)) {
			ok, e := fallback()
			if e != nil {
				return *new(T), ErrorPureFailed(errorToString(e))
			}
			return ok, nil
		} else {
			return *new(T), Error(e)
		}
	}
}

func fallbackParser[T any](inner Parser[T], eval func(*State) (T, Error), value_str string) Parser[T] {
	return Parser[T]{eval, fallbackMeta(inner.meta, value_str)}
}

// Use this value as default if the value isn't present on a command line
//
// # Parser would still fail if the value is present but failure comes from some transformation
//
// # See also
// [Parser.FallbackWith] would allow to Fallback to a value that
// comes from a failing computation such as reading a file. By default Fallback value will
// not be shown in the `--help` output, use [Parser.FallbackDisplay] to change that.
func (self Parser[T]) Fallback(value T) Parser[T] {
	inner := self
	return fallbackParser(inner, func(state *State) (T, Error) {
		return fallback_eval(inner, value, state)
	}, "")
}

// Use this value as default if the value isn't present on a command line
// and show it in `--help` using the representation of the value
func (self Parser[T]) FallbackDisplay(value T) Parser[T] {
	inner := self
	return fallbackParser(inner, func(state *State) (T, Error) {
		return fallback_eval(inner, value, state)
	}, fmt.Sprintf("[default: %v]", value))
}

// Use value produced by this function as default if the value isn't present
//
// # Would still fail if the value is present but failure comes from some earlier transformation
//
// # See also
// [Parser.Fallback] implements similar logic expect that failures aren't expected.
// By default fallback value will not be shown in the `--help` output, use
// [Parser.FallbackWithDisplay] to change that.
func (self Parser[T]) FallbackWith(fallback func() (T, Error)) Parser[T] {
	inner := self
	return fallbackParser(inner, func(state *State) (T, Error) {
		return fallback_with_eval(inner, fallback, state)
	}, "")
}

// Use value produced by this function as default if the value isn't present
// and show the computed default in `--help` output
//
// # If fallback function fails - no value will show up
func (self Parser[T]) FallbackWithDisplay(fallback func() (T, Error)) Parser[T] {
	inner := self
	value_str := ""
	if val, err := fallback(); err == nil {
		value_str = fmt.Sprintf("[default: %v]", val)
	}
	return fallbackParser(inner, func(state *State) (T, Error) {
		return fallback_with_eval(inner, fallback, state)
	}, value_str)
}
