package cli

// Parser that inserts static shell completion into bpaf's dynamic shell completion
type ParseCompShell[T any] struct {
	inner Parser[T]
	op    ShellComp
}

func (self ParseCompShell[T]) Parser() Parser[T] {
	return Parser[T]{self.eval, self.inner.meta}
}

func (self ParseCompShell[T]) eval(args *State) (T, Error) {
	// same as with ParseComp the goal is to replace metavars added by inner parser
	// with a completion that would call a bash script.
	// unlike ParseComp we don't care if inner parser succeeds

	// stash old completions
	comp_items := []Comp{}
	args.swap_comps_with(&comp_items)

	res := newResult(self.inner.eval(args))

	// at this point comp_items contains values added by the inner parser
	args.swap_comps_with(&comp_items)

	depth := args.depth()
	if comp, ok := args.comp_mut().Unpack(); ok {
		for _, ci := range comp_items {
			if is_argument, ok := ci.is_metavar().Unpack(); ok {
				comp.push_shell(self.op, is_argument, depth)
			} else {
				comp.push_comp(ci)
			}
		}
	}

	return res.unpack()
}

// {{{
// Static shell completion
//
// Allows to ask existing shell completion to provide some information such as a file or
// directory names or pass through existing shell completion scripts, see
// [ShellComp](complete_shellShellComp) for accessible functionality
//
// Places function calls in place of metavar placeholder, so running `complete_shell` on
// something that doesn't have a [positional] or [argument](NamedArgargument) doesn't
// make much sense.
//
// # Example
// ```console
// $ app --output C<TAB>
// $ app --output Cargo.toml _
// ```
//
// # Combinatoric usage
// ```rust
// # use bpaf*;
//
//	func output()impl Parser<String> {
//	    long("output")
//	        .help("Cargo.toml file to use as output")
//	        .argument("OUTPUT")
//	        .complete_shell(ShellCompFile { mask: Valid("*.toml") })
//	}
//
// ```
//
// # Derive usage
// ```rust
// # use bpaf*;
//
//	struct fun.Options {
//	    // Cargo.toml file to use as output
//	    #[bpaf(argument("OUTPUT"), complete_shell(ShellCompFile { mask: Valid("*.toml") }))]
//	    output string,
//	}
//
// ```
//
// For multiple file types correct mask syntax is `"*.(toml|md)"`.
// }}}
func (self Parser[T]) CompleteShell(op ShellComp) ParseCompShell[T] {
	return ParseCompShell[T]{inner: self, op: op}
}
