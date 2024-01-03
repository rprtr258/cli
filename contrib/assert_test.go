package flags

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func assertNoError[T any](t *testing.T, args ...string) {
	t.Helper()

	ctx := context.WithValue(context.Background(), _contextKey, t)
	assert.NoError(t, RunContext[T](ctx, args...))
}

func assertEqual[T any](t *testing.T, errWant *Error, args ...string) {
	t.Helper()

	ctx := context.WithValue(context.Background(), _contextKey, t)
	assert.Equal(t, errWant, RunContext[T](ctx, args...))
}
