package cli

// Compose several parsers to produce a single result
//
// Go provides several composition tools:
//
//   - [Pair] and [Triple] combine two or three parsers into one that produces a
//     [fun.Pair] or [Tuple] value.
//   - [Choice] picks the first parser that consumes the leftmost value.
//   - [Construct] builds a parser for a struct type: every exported field with a
//     `cli` tag contributes a parser for that field.
//   - [Parser.Adjacent] restricts a composed group to adjacent command line
//     items.
//
// # Sequential composition
//
// Sequential composition runs each parser and if all of them succeed you get a
// parser producing the combined value:
//
//	width := cli.Argument[uint32](cli.Short('w'), "W", parseUint32).Parser()
//	height := cli.Argument[uint32](cli.Short('h'), "H", parseUint32).Parser()
//	rect := cli.Pair(width, height).Parser()
//
// # Parallel composition
//
// Parallel composition picks one of several parsers with the same result type and
// returns a parser of the same type. Similar to sequential composition you can use
// parsers defined as variables or functions:
//
//	func aOrB() cli.Parser[uint32] {
//	    a := cli.Argument[uint32](cli.Short('a'), "NUM", parseUint32).Parser()
//	    b := cli.Argument[uint32](cli.Short('b'), "NUM", parseUint32).Parser()
//	    // an equivalent way of writing this would be a.OrElse(b)
//	    return cli.Choice(a, b)
//	}
//
// # Structs
//
// [Construct] combines the fields of a struct sequentially: to satisfy the parser
// the user needs to pass values for all the fields, in the order they are declared.
// Field names are chosen with `cli` struct tags, for example
// `Height int `cli:"long=height"` `. Optional fields have types such as Opt[T] or
// []T parsed with [Optional] and [Many] respectively.
//
// To represent multiple alternatives with [Construct] define a field per
// alternative and wrap each one in [Optional].
