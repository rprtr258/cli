package cli

// Parser fails with a message if check returns false, created with [Parser.Guard].
func guard_eval[T any](inner Parser[T], check func(T) bool, message string, state *state) (T, parseError) {
	t, err := inner.eval(state)
	if err != nil {
		return *new(T), err
	}
	if check(t) {
		return t, nil
	} else {
		return *new(T), errorGuardFailed{state.currentIdx, message}
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
// # Example
//
//	collectA := cli.Argument[uint32](cli.Short('a'), "N", parseUint32).Parser().
//	    Guard(func(a uint32) bool { return a < 10 }, "`a` must be below 10")
func (self Parser[T]) Guard(check func(T) bool, message string) Parser[T] {
	inner := self
	return Parser[T]{
		eval: func(state *state) (T, parseError) { return guard_eval(inner, check, message, state) },
		meta: inner.meta,
	}
}
