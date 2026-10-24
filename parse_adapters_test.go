package cli

import (
	"testing"
	"time"
)

func TestParseAdapters_int(t *testing.T) {
	parser := Long("number").Argument("NUM", ParseInt).ToOptions()
	assertSuccess(t, 12, parser, "--number=12")
	assertFailure(t, parser, "couldn't parse `x`: invalid syntax", "--number=x")
	assertFailure(t, parser, "couldn't parse `99999999999999999999`: value out of range", "--number=99999999999999999999")
}

func TestParseAdapters_uint(t *testing.T) {
	parser := Long("number").Argument("NUM", ParseUint).ToOptions()
	assertSuccess(t, uint(12), parser, "--number=12")
	assertFailure(t, parser, "couldn't parse `-1`: invalid syntax", "--number=-1")
}

func TestParseAdapters_float(t *testing.T) {
	parser := Long("number").Argument("NUM", ParseFloat).ToOptions()
	assertSuccess(t, 1.5, parser, "--number=1.5")
	assertFailure(t, parser, "couldn't parse `x`: invalid syntax", "--number=x")
}

func TestParseAdapters_bool(t *testing.T) {
	parser := Long("flag").Argument("FLAG", ParseBool).ToOptions()
	assertSuccess(t, true, parser, "--flag=true")
	assertFailure(t, parser, "couldn't parse `x`: invalid syntax", "--flag=x")
}

func TestParseAdapters_duration(t *testing.T) {
	parser := Long("timeout").Argument("DURATION", ParseDuration).ToOptions()
	assertSuccess(t, 1500*time.Millisecond, parser, "--timeout=1s500ms")
	assertFailure(t, parser, `couldn't parse `+"`5`"+`: time: missing unit in duration "5"`, "--timeout=5")
}

func TestParseAdapters_string(t *testing.T) {
	parser := Long("name").Argument("NAME", ParseString).ToOptions()
	assertSuccess(t, "hello", parser, "--name=hello")
}

func TestCollect_no_typeparam(t *testing.T) {
	parser := Long("n").Argument("N", ParseInt).Parser().Many().ToOptions()
	assertSuccess(t, []int{1, 2}, parser, "--n", "1", "--n", "2")

	parser = Long("n").Argument("N", ParseInt).Parser().Some("no numbers given").ToOptions()
	assertSuccess(t, []int{3, 4}, parser, "--n=3", "--n=4")

	parser = Long("n").Argument("N", ParseInt).Parser().ManyCatch().ToOptions()
	assertSuccess(t, []int{5}, parser, "--n=5")

	parser = Long("n").Argument("N", ParseInt).Parser().SomeCatch("no numbers given").ToOptions()
	assertSuccess(t, []int{6, 7}, parser, "--n=6", "--n=7")

	parser = Long("n").Argument("N", ParseInt).Parser().Collect().ToOptions()
	assertSuccess(t, []int{8}, parser, "--n=8")

	parser = Long("n").Argument("N", ParseInt).Parser().CollectCatch().ToOptions()
	assertSuccess(t, []int{9}, parser, "--n=9")
}
