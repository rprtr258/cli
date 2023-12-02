package cli

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type Filename string

func TestParse(t *testing.T) {
	type Opts struct {
		Verbose bool     `short:"v" long:"verbose" description:"Show verbose debug information"`
		Name    string   `short:"n" long:"name" description:"A name"`
		Args    []string `positional`
	}
	got, err := Parse[Opts]("app --name aboba --verbose xdd")
	assert.NoError(t, err)
	assert.Equal(t, Opts{
		Verbose: true,
		Name:    "aboba",
		Args:    []string{"xdd"},
	}, got)
}
