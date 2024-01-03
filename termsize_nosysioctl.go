//go:build plan9 || appengine || wasm
// +build plan9 appengine wasm

package cli

func getTerminalColumns() int {
	return 80
}
