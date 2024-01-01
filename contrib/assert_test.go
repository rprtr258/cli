package flags

import (
	"fmt"
	"path"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
)

func assertCallerInfo() (file string, line int) {
	ptr := make([]uintptr, 15)
	n := runtime.Callers(1, ptr)
	if n == 0 {
		return "", 0
	}

	mefile, meline := runtime.FuncForPC(ptr[0]).FileLine(ptr[0])
	for i := 2; i < n; i++ {
		if file, line := runtime.FuncForPC(ptr[i]).FileLine(ptr[i]); file != mefile {
			return file, line
		}
	}
	return mefile, meline
}

func assertErrorf(t *testing.T, format string, args ...any) {
	t.Helper()

	msg := fmt.Sprintf(format, args...)
	file, line := assertCallerInfo()
	t.Errorf("%s:%d: %s", path.Base(file), line, msg)
}

func assertParserSuccess(t *testing.T, data any, args ...string) (parser_ *Parser, ret_ []string) {
	parser := NewParser(data, Default&^PrintErrors)
	ret, err := parser.ParseArgs(args...)
	assert.NoError(t, err)
	return parser, ret
}

func assertParseSuccess(t *testing.T, data any, args ...string) []string {
	_, ret := assertParserSuccess(t, data, args...)
	return ret
}

func assertParseFail(t *testing.T, errWant *Error, data any, args ...string) []string {
	t.Helper()

	parser := NewParser(data, Default&^PrintErrors)
	ret, err := parser.ParseArgs(args...)
	assert.Equal(t, errWant, err)
	return ret
}
