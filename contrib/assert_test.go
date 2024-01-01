package flags

import (
	"fmt"
	"os"
	"os/exec"
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
		f := runtime.FuncForPC(ptr[i])
		file, line := f.FileLine(ptr[i])

		if file != mefile {
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

func diff(a, b string) (string, error) {
	atmp, err := os.CreateTemp("", "help-diff")
	if err != nil {
		return "", err
	}

	btmp, err := os.CreateTemp("", "help-diff")
	if err != nil {
		return "", err
	}

	if _, err := atmp.WriteString(a); err != nil {
		return "", err
	}

	if _, err := btmp.WriteString(b); err != nil {
		return "", err
	}

	ret, err := exec.Command("diff", "-u", "-d", "--label", "got", atmp.Name(), "--label", "expected", btmp.Name()).Output()

	os.Remove(atmp.Name())
	os.Remove(btmp.Name())

	if err.Error() == "exit status 1" {
		return string(ret), nil
	}

	return string(ret), err
}

func assertDiff(t *testing.T, actual, expected, msg string) {
	t.Helper()

	if actual == expected {
		return
	}

	ret, err := diff(actual, expected)

	if err != nil {
		assertErrorf(t, "Unexpected diff error: %s", err)
		assertErrorf(t, "Unexpected %s, expected:\n\n%s\n\nbut got\n\n%s", msg, expected, actual)
	} else {
		assertErrorf(t, "Unexpected %s:\n\n%s", msg, ret)
	}
}
