package flags

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

type contextKey struct{}

var _contextKey = contextKey{}

func toCtx(ctx context.Context, runTest func(any)) context.Context {
	return context.WithValue(ctx, _contextKey, runTest)
}

func fromCtx(ctx context.Context) func(any) {
	return ctx.Value(_contextKey).(func(any))
}

func assertNoError[T any](t *testing.T, args ...string) {
	t.Helper()

	args = append([]string{"x"}, args...)

	ctx := context.WithValue(context.Background(), _contextKey, t)
	assert.NoError(t, RunContext[T](ctx, args...))
}

func assertEqual[T any](t *testing.T, errWant *Error, args ...string) {
	t.Helper()

	args = append([]string{"x"}, args...)

	ctx := context.WithValue(context.Background(), _contextKey, t)
	assert.Equal(t, errWant, RunContext[T](ctx, args...))
}
