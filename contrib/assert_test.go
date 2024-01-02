package flags

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func assertParserSuccess(t *testing.T, data any, args ...string) *Parser {
	t.Helper()

	parser := NewParser(data, Default)
	err := parser.Run(args...)
	assert.NoError(t, err)
	return parser
}

func assertParseFail(t *testing.T, errWant *Error, data any, args ...string) {
	t.Helper()

	parser := NewParser(data, Default)
	err := parser.Run(args...)
	assert.Equal(t, errWant, err)
}
