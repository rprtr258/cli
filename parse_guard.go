package cli

// guard_eval fails with a message if check returns false; created with
// [Parser.Guard].
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

// Guard validates the parsed value or fails with a message.
//
// If the value doesn't satisfy the constraint - the parser fails with the specified error message.
//
// The message can be a plain string or a preformatted [Doc].
//
// # Example
//
//	collectA := cli.Short('a').Argument("N", parseUint32).
//	    Guard(func(a uint32) bool { return a < 10 }, "`a` must be below 10")
func (self Parser[T]) Guard(check func(T) bool, message string) Parser[T] {
	inner := self
	return Parser[T]{
		eval: func(state *state) (T, parseError) { return guard_eval(inner, check, message, state) },
		meta: inner.meta,
	}
}
