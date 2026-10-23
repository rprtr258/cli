package cli

import "fmt"

func fallbackMeta(innerMeta meta, value_str string) meta {
	m := meta(metaOptional{innerMeta})
	if value_str != "" {
		m = metaSuffix{m, NewDoc(value_str)}
	}
	return m
}

// fallback_eval substitutes a missing value but not a parse failure; created
// with [Parser.Fallback] and [Parser.FallbackDisplay].
func fallback_eval[T any](inner Parser[T], value T, state *state) (T, parseError) {
	ok, e := inner.eval(state)
	if e != nil {
		if can_catch(parseError(e)) {
			return value, nil
		} else {
			return *new(T), parseError(e)
		}
	}
	return ok, nil
}

// fallback_with_eval substitutes a missing value with the function result but
// not a parser failure; created with [Parser.FallbackWith] and
// [Parser.FallbackWithDisplay].
func fallback_with_eval[T any](inner Parser[T], fallback func() (T, error), state *state) (T, parseError) {
	clone := state.clone()
	ok, e := inner.eval(&clone)
	if e == nil {
		*state, clone = clone, *state
		return ok, nil
	} else {
		*state, clone = clone, *state
		if can_catch(parseError(e)) {
			ok, e := fallback()
			if e != nil {
				return *new(T), errorPureFailed(e.Error())
			}
			return ok, nil
		} else {
			return *new(T), parseError(e)
		}
	}
}

func fallbackParser[T any](inner Parser[T], eval func(*state) (T, parseError), value_str string) Parser[T] {
	return Parser[T]{eval, fallbackMeta(inner.meta, value_str)}
}

// Fallback uses this value as default if the value isn't present on a command line.
//
// # The parser would still fail if the value is present but failure comes from some transformation
//
// # See also
// [Parser.FallbackWith] would allow to Fallback to a value that
// comes from a failing computation such as reading a file. By default Fallback value will
// not be shown in the `--help` output, use [Parser.FallbackDisplay] to change that.
func (self Parser[T]) Fallback(value T) Parser[T] {
	inner := self
	return fallbackParser(inner, func(state *state) (T, parseError) {
		return fallback_eval(inner, value, state)
	}, "")
}

// FallbackDisplay uses this value as default if the value isn't present on a
// command line and shows it in --help using the representation of the value.
func (self Parser[T]) FallbackDisplay(value T) Parser[T] {
	inner := self
	return fallbackParser(inner, func(state *state) (T, parseError) {
		return fallback_eval(inner, value, state)
	}, fmt.Sprintf("[default: %v]", value))
}

// FallbackWith uses the value produced by this function as default if the value isn't present.
//
// # The parser would still fail if the value is present but failure comes from some earlier transformation
//
// # See also
// [Parser.Fallback] implements similar logic expect that failures aren't expected.
// By default fallback value will not be shown in the `--help` output, use
// [Parser.FallbackWithDisplay] to change that.
func (self Parser[T]) FallbackWith(fallback func() (T, error)) Parser[T] {
	inner := self
	return fallbackParser(inner, func(state *state) (T, parseError) {
		return fallback_with_eval(inner, fallback, state)
	}, "")
}

// FallbackWithDisplay uses the value produced by this function as default if
// the value isn't present and shows the computed default in --help output.
//
// # If fallback function fails - no value will show up
func (self Parser[T]) FallbackWithDisplay(fallback func() (T, error)) Parser[T] {
	inner := self
	value_str := ""
	if val, err := fallback(); err == nil {
		value_str = fmt.Sprintf("[default: %v]", val)
	}
	return fallbackParser(inner, func(state *state) (T, parseError) {
		return fallback_with_eval(inner, fallback, state)
	}, value_str)
}
