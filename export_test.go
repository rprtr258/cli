package cli

import (
	"testing"

	"github.com/kr/pretty"
	"github.com/rprtr258/fun"
)

func ppair[T, R any](a T, b R) fun.Pair[T, R] {
	return fun.Pair[T, R]{a, b}
}

func requireNoParseFailure(t *testing.T, err parseFailure) {
	t.Helper()

	if err != nil {
		switch self := err.(type) {
		case parseFailureStdout:
			t.Log("Unexpected parseFailureStdout:", pretty.Sprint(self.msg), self.full)
		case parseFailureCompletion:
			t.Log("Unexpected parseFailureCompletion:", string(self))
		case parseFailureStderr:
			t.Log("Unexpected parseFailureStderr:", pretty.Sprint(Doc(self)))
		}
		t.FailNow()
	}
}
