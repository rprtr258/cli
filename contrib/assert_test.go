package flags

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func assertParserSuccess(t *testing.T, data any, args ...string) {
	t.Helper()

	assert.NoError(t, Run(data, args...))
}

func assertParseFail(t *testing.T, errWant *Error, data any, args ...string) {
	t.Helper()

	assert.Equal(t, errWant, Run(data, args...))
}
