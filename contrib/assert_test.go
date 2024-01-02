package flags

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func assertParserSuccess(t *testing.T, data any, args ...string) (parser_ *Parser, ret_ []string) {
	t.Helper()

	parser := NewParser(data, Default)
	ret, err := parser.ParseArgs(args...)
	assert.NoError(t, err)
	return parser, ret
}

func assertParseSuccess(t *testing.T, data any, args ...string) []string {
	t.Helper()

	_, ret := assertParserSuccess(t, data, args...)
	return ret
}

func assertParseFail(t *testing.T, errWant *Error, data any, args ...string) []string {
	t.Helper()

	parser := NewParser(data, Default)
	ret, err := parser.ParseArgs(args...)
	assert.Equal(t, errWant, err)
	return ret
}
