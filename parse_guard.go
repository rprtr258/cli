package cli

// Parser fails with a message if check returns false, created with [Parser.Guard].
func guard_eval[T any](inner Parser[T], check func(T) bool, message string, state *State) (T, Error) {
	t, err := inner.eval(state)
	if err != nil {
		return *new(T), err
	}
	if check(t) {
		return t, nil
	} else {
		return *new(T), ErrorGuardFailed{state.currentIdx, message}
	}
}

// Validate or fail with a message
//
// If the value doesn't satisfy the constraint - the parser fails with the specified error message.
//
// # Derive usage
// Derive variant of the `Guard` takes a function name instead of a closure, mostly to keep things
// clean. The second argument can be either a string literal or a constant name for a static [str].
//
// #[cfg_attr(not(doctest), doc = include_str!("docs2/Guard.md"))]
func (self Parser[T]) Guard(check func(T) bool, message string) Parser[T] {
	inner := self
	return Parser[T]{
		eval: func(state *State) (T, Error) { return guard_eval(inner, check, message, state) },
		meta: inner.meta,
	}
}
