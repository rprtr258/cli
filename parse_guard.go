package cli

// Parser fails with a message if check returns false, created with [guard](Parserguard).
type ParseGuard[T any] struct {
	inner   Parser[T]
	check   func(T) bool
	message string
}

func (self ParseGuard[T]) Parser() Parser[T] {
	return Parser[T]{self.eval, self.inner.meta}
}

func (self ParseGuard[T]) eval(args *State) (T, Error) {
	t, err := self.inner.eval(args)
	if err != nil {
		return *new(T), err
	}
	if self.check(t) {
		return t, nil
	} else {
		return *new(T), Error(ErrorGuardFailed{args.current, self.message})
	}
}

// {{{ Guard
// Validate or fail with a message
//
// If the value doesn't satisfy the constraint - the parser fails with the specified error message.
//
// # Derive usage
// Derive variant of the `Guard` takes a function name instead of a closure, mostly to keep things
// clean. The second argument can be either a string literal or a constant name for a static [str].
//
// }}}
// #[cfg_attr(not(doctest), doc = include_str!("docs2/Guard.md"))]
func (self Parser[T]) Guard(check func(T) bool, message string) ParseGuard[T] {
	return ParseGuard[T]{
		inner:   self,
		check:   check,
		message: message,
	}
}
