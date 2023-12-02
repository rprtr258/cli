package flags

import "reflect"

// Arg represents a positional argument on the command line.
type Arg struct {
	Name            string // Name of the positional argument (used in the help)
	Description     string // Description of the positional argument (used in the help)
	Required        int    // minimal number of required positional arguments
	RequiredMaximum int    // maximum number of required positional arguments

	value reflect.Value
	tag   multiTag
}

func (a *Arg) isRemaining() bool {
	return a.value.Type().Kind() == reflect.Slice
}
