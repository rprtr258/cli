package cli

// Fail with a fixed error message
//
// This parser produces `T` of any type but instead of producing it when asked - it fails
// with a custom error message. Can be useful for creating custom logic
//
// # Combinatoric usage
//
//	func must_agree() Parser[struct{}] {
//	    a := Long("agree").ReqFlag(struct{}{})
//	    no_a := Fail("You must accept the license agreement with --agree before proceeding")
//	    return Pair([a, no_a])
//	}
//
// # Example
//
//	$ app
//	// exits with "You must accept the license agreement with --agree before proceeding"
//	$ app --agree
//	// succeeds
func Fail[T any](msg string) Parser[T] {
	eval := func(state *state) (T, parseError) {
		state.currentIdx = Invalid[int]()
		return *new(T), errorParseFail(msg)
	}
	meta := metaSkip{}
	return Parser[T]{eval, meta}
}
