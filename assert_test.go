package cli

import (
	"fmt"
	"path"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
)

func assertCallerInfo() (string, int) {
	ptr := make([]uintptr, 15)
	n := runtime.Callers(1, ptr)

	if n == 0 {
		return "", 0
	}

	mef := runtime.FuncForPC(ptr[0])
	mefile, meline := mef.FileLine(ptr[0])

	for i := 2; i < n; i++ {
		f := runtime.FuncForPC(ptr[i])
		file, line := f.FileLine(ptr[i])

		if file != mefile {
			return file, line
		}
	}

	return mefile, meline
}

func assertErrorf(t *testing.T, format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	file, line := assertCallerInfo()
	t.Errorf("%s:%d: %s", path.Base(file), line, msg)
}

func assertFatalf(t *testing.T, format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	file, line := assertCallerInfo()
	t.Fatalf("%s:%d: %s", path.Base(file), line, msg)
}

func assertParserSuccess(t *testing.T, data any, args ...string) (*Parser, []string) {
	parser := NewParser(data, Default&^PrintErrors)
	ret, err := parser.ParseArgs(args)
	assert.NoError(t, err)

	return parser, ret
}

func assertParseSuccess(t *testing.T, data any, args ...string) []string {
	_, ret := assertParserSuccess(t, data, args...)
	return ret
}

func assertError(t *testing.T, err error, typ ErrorType, msg string) {
	if err == nil {
		assertFatalf(t, "Expected error: \"%s\", but no error occurred", msg)
		return
	}

	if e, ok := err.(*Error); !ok {
		assertFatalf(t, "Expected Error type, but got %#v", err)
	} else {
		assert.Equal(t, typ, e.Type)
		assert.Equal(t, msg, e.Message)
	}
}

func assertParseFail(t *testing.T, typ ErrorType, msg string, data any, args ...string) []string {
	parser := NewParser(data, Default&^PrintErrors)
	ret, err := parser.ParseArgs(args)

	assertError(t, err, typ, msg)
	return ret
}
