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

	AssertSuccess(t, Invalid[Flag](), parser)
	AssertSuccess(t, Valid(FlagN), parser, "-y", "-y", "-n")
	AssertSuccess(t, Valid(FlagY), parser, "-y", "-y", "-n", "-y")
}
