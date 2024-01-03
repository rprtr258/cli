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
	f, ok := ctx.Value(_contextKey).(func(any))
	if !ok {
		panic("no test function found while expected")
	}

	return f
}

func testSuccess[T, R any](t *testing.T, runTest func(opts R), args ...string) {
	called := 0
	ctx := toCtx(context.Background(), func(a any) {
		called++
		assert.IsType(t, *new(R), a)
		runTest(a.(R))
	})
	assert.NoError(t, RunContext[T](ctx, args...))
	assert.Equal(t, 1, called)
}

func testFail[T any](t *testing.T, errString string, args ...string) {
	called := 0
	ctx := toCtx(context.Background(), func(a any) {
		called++
	})
	assert.EqualError(t, RunContext[T](ctx, args...), errString)
	assert.Equal(t, 0, called)
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
