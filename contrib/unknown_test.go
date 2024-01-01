package flags

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUnknownFlags(t *testing.T) {
	opts := struct {
		Verbose []bool `short:"v" long:"verbose" description:"Verbose output"`
	}{}

	p := NewParser(&opts, 0)
	_, err := p.ParseArgs(
		"-f",
	)
	if err == nil {
		t.Fatal("Expected error for unknown argument")
	}
}

func TestIgnoreUnknownFlags(t *testing.T) {
	opts := struct {
		Verbose []bool `short:"v" long:"verbose" description:"Verbose output"`
	}{}

	p := NewParser(&opts, IgnoreUnknown)
	args, err := p.ParseArgs(
		"hello",
		"world",
		"-v",
		"--foo=bar",
		"--verbose",
		"-f",
	)
	assert.NoError(t, err)

	exargs := []string{
		"hello",
		"world",
		"--foo=bar",
		"-f",
	}

	assert.Equal(t, exargs, args)
}
