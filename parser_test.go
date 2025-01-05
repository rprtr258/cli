package cli

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPure(t *testing.T) {
	parser := Pure(123)
	x, err := parser.parse(newState())
	require.Nil(t, err)
	require.Equal(t, 123, x)
}

func TestFail(t *testing.T) {
	parser := Fail[int]("nope")
	_, err := parser.parse(newState())
	require.Equal(t, "nope", err.Error())
}
