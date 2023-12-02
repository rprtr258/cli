package cli

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func useParser(t *testing.T, data any, args ...string) (*Parser, []string) {
	parser := NewParser(data, Default&^PrintErrors)
	ret, err := parser.ParseArgs(args)
	assert.NoError(t, err)
	return parser, ret
}

func useParsedArgs(t *testing.T, data any, args ...string) []string {
	_, ret := useParser(t, data, args...)
	return ret
}

func assertParseFailed(t *testing.T, typ ErrorType, msg string, data any, args ...string) []string {
	t.Helper()

	ret, err := NewParser(data, Default&^PrintErrors).ParseArgs(args)
	assert.Equal(t, err, &Error{
		Type:    typ,
		Message: msg,
	})
	return ret
}
