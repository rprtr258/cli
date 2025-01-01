package cli

import "iter"

// Parser that replaces metavar placeholders with actual info in shell completion
type ParseComp[T any] struct {
	inner Parser[T]
	op    func(T) iter.Seq2[string, Option[string]]
	group Option[string]
}

func (self ParseComp[T]) Parser() Parser[T] {
	return Parser[T]{self.eval, self.inner.meta}
}

// Attach group name to parsed values
func (self *ParseComp[T]) Group(group string) *ParseComp[T] {
	self.group = Valid(group)
	return self
}

func (self ParseComp[T]) eval(args *State) (T, Error) {
	// stash old
	comp_items := []Comp(nil)
	args.swap_comps_with(&comp_items)

	res, err := self.inner.eval(args)

	// restore old, now metavars added by inner parser, if any, are in comp_items
	args.swap_comps_with(&comp_items)

	if comp, ok := args.comp_mut(); ok && err != nil {
		comp.extend_comps(comp_items)
		return *new(T), err
	}

	if err != nil {
		return *new(T), err
	}

	// completion function generates suggestions based on the parsed inner value, for
	// that `res` must contain a parsed value
	depth := args.depth()
	if comp, ok := args.comp_mut(); ok {
		for _, ci := range comp_items {
			if is_arg, ok := ci.is_metavar(); ok {
				suggestions := self.op(res)
				// strip metavar when completion makes a single good suggestion
				// TODO: get back if
				// if len(suggestions) != 1 {
				// 	comp.push_comp(ci)
				// }
				for replacement, description := range suggestions {
					group := self.group
					comp.push_value(
						replacement,
						description,
						group,
						depth,
						is_arg,
					)
				}
			} else {
				comp.push_comp(ci)
			}
		}
	}
	return res, nil
}

// Dynamic shell completion
//
// Allows to generate autocompletion information for the shell. Completer places generated input
// in place of metavar placeholders, so running `completer` on something that doesn't have a
// [positional] or an [argument](NamedArgargument) doesn't make much sense.
//
// Takes a function as a parameter that tries to complete partial input to a full one with an
// fun.Optional description. `bpaf` would substitute a current positional item or an argument with an empty
// string if a value isn't available yet so it's best to run `complete` where parsing can't fail:
// right after [argument](NamedArgargument) or [positional], but this isn't enforced.
//
// # Example
// ```console
// $ app --name L<TAB>
// $ app --name Lupusregina _
// ```
//
// #[cfg_attr(not(doctest), doc = include_str!("docs2/complete.md"))]
//
// ## A simple example
//
// #[cfg_attr(not(doctest), doc = include_str!("docs2/simple_dynamic.md"))]
//
// ## More detailed example
//
// #[cfg_attr(not(doctest), doc = include_str!("docs2/derive_show_asm.md"))]
func (self Parser[T]) Complete(op func(T) iter.Seq2[string, Option[string]]) ParseComp[T] {
	return ParseComp[T]{
		inner: self,
		op:    op,
		group: Invalid[string](),
	}
}
