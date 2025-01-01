package cli

import "testing"

func TestToggle_flag(t *testing.T) {
	type Flag bool
	const (
		FlagN Flag = false
		FlagY Flag = true
	)

	parser := toggle_flag(
		Short('y'), FlagY,
		Short('n'), FlagN,
	).ToOptions()

	assertSuccess(t, Invalid[Flag](), parser)
	// assertSuccess(t, Valid(FlagN), parser, "-y", "-y", "-n")
	// assertSuccess(t, Valid(FlagY), parser, "-y", "-y", "-n", "-y")
}
