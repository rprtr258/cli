package flags

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUnknownFlags(t *testing.T) {
	opts := struct {
		Verbose []bool `short:"v" long:"verbose" description:"Verbose output"`
	}{}

	_, err := NewParser(&opts, 0).ParseArgs("-f")
	assert.Error(t, err)
}

func TestIgnoreUnknownFlags(t *testing.T) {
	opts := struct {
		Verbose []bool `short:"v" long:"verbose" description:"Verbose output"`
	}{}

	args, err := NewParser(&opts, IgnoreUnknown).ParseArgs(
		"hello",
		"world",
		"-v",
		"--foo=bar",
		"--verbose",
		"-f",
	)
	assert.NoError(t, err)
	assert.Equal(t, []string{
		"hello",
		"world",
		"--foo=bar",
		"-f",
	}, args)
}
