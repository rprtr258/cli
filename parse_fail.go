package cli

// Parser that fails without consuming any input, created with [Fail].
type ParseFail[T any] struct {
	field1 string
}

func (self ParseFail[T]) eval(args *State) (T, Error) {
	args.current = Invalid[int]()
	return *new(T), ErrorParseFail(self.field1)
}

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
// ```console
// $ app
// // exits with "You must accept the license agreement with --agree before proceeding"
// $ app --agree
// // succeeds
// ```
func Fail[T any](msg string) ParseFail[T] {
	return ParseFail[T]{
		field1: msg,
	}
}
