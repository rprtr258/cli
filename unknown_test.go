package cli

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type unknownOpts struct {
	Verbose []bool `short:"v" long:"verbose" description:"Verbose output"`
}

func TestUnknownFlags(t *testing.T) {
	var opts unknownOpts
	p := NewParser(&opts, 0)
	_, err := p.ParseArgs([]string{"-f"})
	assert.Error(t, err)
}

func TestIgnoreUnknownFlags(t *testing.T) {
	var opts unknownOpts
	p := NewParser(&opts, IgnoreUnknown)
	got, err := p.ParseArgs([]string{"hello", "world", "-v", "--foo=bar", "--verbose", "-f"})
	assert.NoError(t, err)

	exargs := []string{"hello", "world", "--foo=bar", "-f"}
	assert.Equal(t, exargs, got)
}
