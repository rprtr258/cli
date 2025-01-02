package cli

import "fmt"

// Parser that substitutes missing value but not parse failure, created with [fallback](Parserfallback).
type ParseFallback[T any] struct {
	inner     Parser[T]
	value     T
	value_str string
}

func (self ParseFallback[T]) Parser() Parser[T] {
	return Parser[T]{self.eval, self.meta()}
}

func (self ParseFallback[T]) eval(state *State) (T, Error) {
	clone := state.clone()
	ok, e := self.inner.eval(&clone)
	if e != nil {
		*state, clone = clone, *state
		if can_catch(Error(e)) {
			return self.value, nil
		} else {
			return *new(T), Error(e)
		}
	}
	*state = clone
	return ok, nil
}

func (self ParseFallback[T]) meta() Meta {
	m := MetaOptional{self.inner.meta}
	if self.value_str == "" {
		return m
	} else {
		buf := NewDoc(self.value_str)
		return MetaSuffix{m, buf}
	}
}

// Show [fallback](Parserfallback) value in `--help` using [Display](stdfmtDisplay)
// representation
//
// #[cfg_attr(not(doctest), doc = include_str!("docs2/dis_fallback.md"))]
func (self *ParseFallback[T]) DisplayFallback() ParseFallback[T] {
	self.value_str = fmt.Sprintf("[default: %v]", self.value)
	return *self
}

// Show [fallback](Parserfallback) value in `--help` using [Debug](stdfmtDebug)
// representation
//
// #[cfg_attr(not(doctest), doc = include_str!("docs2/deb_fallback_with.md"))]
func (self *ParseFallback[T]) DebugFallback() ParseFallback[T] {
	self.value_str = fmt.Sprintf("[default: %#v]", self.value)
	return *self
}

// Use this value as default if the value isn't present on a command line
//
// # Parser would still fail if the value is present but failure comes from some transformation
//
// #[cfg_attr(not(doctest), doc = include_str!("docs2/dis_fallback.md"))]
//
// # See also
// [fallback_with](Parserfallback_with) would allow to try to Fallback to a value that
// comes from a failing computation such as reading a file. By default Fallback value will
// not be shown in the `--help` output, you can change that by using
// [display_fallback](ParseFallbackdisplay_fallback) and
// [debug_fallback](ParseFallbackdebug_fallback).
func (self Parser[T]) Fallback(value T) *ParseFallback[T] {
	return &ParseFallback[T]{
		inner:     self,
		value:     value,
		value_str: "",
	}
}

// Parser that substitutes missing value with a function results but not parser
// failure, created with [fallback_with](Parserfallback_with).
type ParseFallbackWith[T any] struct {
	inner     Parser[T]
	fallback  func() (T, Error)
	value_str string
}

func (self ParseFallbackWith[T]) Parser() Parser[T] {
	return Parser[T]{self.eval, self.meta()}
}

func (self ParseFallbackWith[T]) eval(state *State) (T, Error) {
	clone := state.clone()
	ok, e := self.inner.eval(&clone)
	if e == nil {
		*state, clone = clone, *state
		return ok, nil
	} else {
		*state, clone = clone, *state
		if can_catch(Error(e)) {
			ok, e := self.fallback()
			if e != nil {
				return *new(T), ErrorPureFailed(errorToString(e))
			}
			return ok, nil
		} else {
			return *new(T), Error(e)
		}
	}
}

func (self ParseFallbackWith[T]) meta() Meta {
	m := MetaOptional{self.inner.meta}
	if self.value_str == "" {
		return m
	} else {
		buf := NewDoc(self.value_str)
		return MetaSuffix{m, buf}
	}
}

// Show [fallback_with](Parserfallback_with) value in `--help` using [Display](stdfmtDisplay)
// representation
//
// # If fallback function fails - no value will show up
//
// #[cfg_attr(not(doctest), doc = include_str!("docs2/dis_fallback_with.md"))]
func (self *ParseFallbackWith[T]) DisplayFallback() ParseFallbackWith[T] {
	if val, err := self.fallback(); err == nil {
		self.value_str = fmt.Sprintf("[default: %v]", val)
	}
	return *self
}

// Show [fallback_with](Parserfallback_with) value in `--help` using [Debug](stdfmtDebug)
// representation
//
// # If fallback function fails - no value will show up
//
// #[cfg_attr(not(doctest), doc = include_str!("docs2/deb_fallback.md"))]
func (self *ParseFallbackWith[T]) debug_fallback() *ParseFallbackWith[T] {
	val, err := self.fallback()
	if err == nil {
		self.value_str = fmt.Sprintf("[default: %#v]", val)
	}
	return self
}

// Use value produced by this function as default if the value isn't present
//
// # Would still fail if the value is present but failure comes from some earlier transformation
//
// #[cfg_attr(not(doctest), doc = include_str!("docs2/dis_fallback_with.md"))]
//
// # See also
// [fallback](Parserfallback) implements similar logic expect that failures aren't expected.
// By default fallback value will not be shown in the `--help` output, you can change that by using
// [display_fallback](ParseFallbackWithdisplay_fallback) and
// [debug_fallback](ParseFallbackWithdebug_fallback).
func (self Parser[T]) FallbackWith(fallback func() (T, Error)) ParseFallbackWith[T] {
	return ParseFallbackWith[T]{
		inner:     self,
		fallback:  fallback,
		value_str: "",
	}
}
