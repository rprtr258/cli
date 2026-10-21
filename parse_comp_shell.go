package cli

// Parser that inserts static shell completion into the dynamic shell completion
func compShell_eval[T any](inner Parser[T], op ShellComp, state *state) (T, parseError) {
	// same as with ParseComp the goal is to replace metavars added by inner parser
	// with a completion that would call a bash script.
	// unlike ParseComp we don't care if inner parser succeeds

	// stash old completions
	comp_items := []comp{}
	state.swap_comps_with(&comp_items)

	res, err := inner.eval(state)

	// at this point comp_items contains values added by the inner parser
	state.swap_comps_with(&comp_items)

	depth := state.depth()
	if comp, ok := state.comp_mut(); ok {
		for _, ci := range comp_items {
			if is_argument, ok := ci.is_metavar(); ok {
				comp.push_shell(op, is_argument, depth)
			} else {
				comp.push_comp(ci)
			}
		}
	}

	return res, err
}

// Static shell completion
//
// Allows to ask existing shell completion to provide some information such as a file or
// directory names or pass through existing shell completion scripts, see
// [ShellComp](complete_shellShellComp) for accessible functionality
//
// Places function calls in place of the metavar placeholder, so running it on
// something that doesn't have a positional or an argument doesn't
// make much sense.
//
// # Example
// ```console
// $ app --output C<TAB>
// $ app --output Cargo.toml _
// ```
//
// # Usage
//
//	out := cli.Argument[string](cli.Long("output").
//	    Help("Cargo.toml file to use as output"), "OUTPUT", nil).
//	    CompleteShell(cli.ShellCompFile(Valid("*.toml")))
//
// For multiple file types correct mask syntax is `"*.(toml|md)"`.
func (self Parser[T]) CompleteShell(op ShellComp) Parser[T] {
	inner := self
	return Parser[T]{
		eval: func(state *state) (T, parseError) { return compShell_eval(inner, op, state) },
		meta: inner.meta,
	}
}
