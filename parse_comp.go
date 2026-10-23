package cli

import "iter"

// ParseComp is a parser that replaces metavar placeholders with actual info in
// shell completion.
type ParseComp[T any] struct {
	inner Parser[T]
	op    func(T) iter.Seq2[string, Opt[string]]
	group Opt[string]
}

func (self ParseComp[T]) Parser() Parser[T] {
	return Parser[T]{self.eval, self.inner.meta}
}

// ToOptions prepares the parser for running, see [Parser.ToOptions].
func (self ParseComp[T]) ToOptions() *OptionParser[T] { return self.Parser().ToOptions() }

// Map applies a pure transformation to the parsed value, see [Parser.Map].
func (self ParseComp[T]) Map[R any](fn func(T) R) Parser[R] { return self.Parser().Map(fn) }

// Hide hides this parser from help generation, see [Parser.Hide].
func (self ParseComp[T]) Hide() Parser[T] { return self.Parser().Hide() }

// HideUsage hides this parser from the usage line, see [Parser.HideUsage].
func (self ParseComp[T]) HideUsage() Parser[T] { return self.Parser().HideUsage() }

// CustomUsage overrides the usage line, see [Parser.CustomUsage].
func (self ParseComp[T]) CustomUsage(usage Doc) Parser[T] { return self.Parser().CustomUsage(usage) }

// GroupHelp attaches a help message to this parser, see [Parser.GroupHelp].
func (self ParseComp[T]) GroupHelp(message string) Parser[T] { return self.Parser().GroupHelp(message) }

// Guard validates the parsed value, see [Parser.Guard].
func (self ParseComp[T]) Guard(check func(T) bool, message string) Parser[T] {
	return self.Parser().Guard(check, message)
}

// Fallback provides a default value, see [Parser.Fallback].
func (self ParseComp[T]) Fallback(value T) Parser[T] { return self.Parser().Fallback(value) }

// FallbackDisplay provides a default value shown in help, see [Parser.FallbackDisplay].
func (self ParseComp[T]) FallbackDisplay(value T) Parser[T] {
	return self.Parser().FallbackDisplay(value)
}

// FallbackWith provides a lazily computed default value, see [Parser.FallbackWith].
func (self ParseComp[T]) FallbackWith(fallback func() (T, error)) Parser[T] {
	return self.Parser().FallbackWith(fallback)
}

// FallbackWithDisplay provides a lazily computed default value shown in help, see [Parser.FallbackWithDisplay].
func (self ParseComp[T]) FallbackWithDisplay(fallback func() (T, error)) Parser[T] {
	return self.Parser().FallbackWithDisplay(fallback)
}

// Last keeps only the last matched value, see [Parser.Last].
func (self ParseComp[T]) Last() Parser[T] { return self.Parser().Last() }

// Count the number of matches, see [Parser.Count].
func (self ParseComp[T]) Count() Parser[int] { return self.Parser().Count() }

// Complete adds custom shell completion, see [Parser.Complete].
func (self ParseComp[T]) Complete(op func(T) iter.Seq2[string, Opt[string]]) ParseComp[T] {
	return self.Parser().Complete(op)
}

// CompleteShell adds file or directory shell completion, see [Parser.CompleteShell].
func (self ParseComp[T]) CompleteShell(op ShellComp) Parser[T] {
	return self.Parser().CompleteShell(op)
}

// Group attaches a group name to parsed values.
func (self *ParseComp[T]) Group(group string) *ParseComp[T] {
	self.group = Valid(group)
	return self
}

func (self ParseComp[T]) eval(state *state) (T, parseError) {
	// stash old
	comp_items := []comp(nil)
	state.swap_comps_with(&comp_items)

	res, err := self.inner.eval(state)

	// restore old, now metavars added by inner parser, if any, are in comp_items
	state.swap_comps_with(&comp_items)

	if comp, ok := state.comp_mut(); ok && err != nil {
		comp.extend_comps(comp_items)
		return *new(T), err
	}

	if err != nil {
		return *new(T), err
	}

	// completion function generates suggestions based on the parsed inner value, for
	// that `res` must contain a parsed value
	depth := state.depth()
	if comp, ok := state.comp_mut(); ok {
		for i := range comp_items {
			ci := &comp_items[i]
			if is_arg, ok := ci.is_metavar(); ok {
				type suggestion struct {
					replacement string
					description Opt[string]
				}
				suggestions := []suggestion(nil)
				for replacement, description := range self.op(res) {
					suggestions = append(suggestions, suggestion{replacement, description})
				}
				// strip metavar when completion makes a single good suggestion
				if len(suggestions) != 1 {
					comp.push_comp(ci)
				}
				for _, s := range suggestions {
					group := self.group
					comp.push_value(
						s.replacement,
						s.description,
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

// Complete adds dynamic shell completion.
//
// Allows to generate autocompletion information for the shell. The completer places generated
// input in place of metavar placeholders, so running it on something that doesn't have a
// positional or an argument doesn't make much sense.
//
// Takes a function as a parameter that tries to complete partial input to a full one with an
// Optional description. The library would substitute the current positional item or an argument with an empty
// string if a value isn't available yet so it's best to run `complete` where parsing can't fail:
// right after an argument or a positional, but this isn't enforced.
//
// # Example
// ```console
// $ app --name L<TAB>
// $ app --name Lupusregina _
// ```
//
// # A simple example
//
// Complete the value of one parser with values of another, for example completing
// `--name` with a list of known names.
//
// ## More detailed example
//
// The callback receives the partial value parsed so far and yields candidate
// completions; the library feeds them to the shell.
func (self Parser[T]) Complete(op func(T) iter.Seq2[string, Opt[string]]) ParseComp[T] {
	return ParseComp[T]{
		inner: self,
		op:    op,
		group: Invalid[string](),
	}
}
