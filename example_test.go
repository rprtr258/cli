package cli

import "fmt"

// Choice picks the first parser that succeeds.
func ExampleChoice() {
	a := Short('a').Argument[uint32]("NUM", parseUint32)
	b := Short('b').Argument[uint32]("NUM", parseUint32)
	num := Choice(a, b).ToOptions()

	value, err := num.RunE([]string{"app", "-a", "12"})
	fmt.Println(value, err)

	value, err = num.RunE([]string{"app", "-b", "3"})
	fmt.Println(value, err)

	// Output:
	// 12 <nil>
	// 3 <nil>
}
