//go:build !windows && !plan9 && !appengine && !wasm
// +build !windows,!plan9,!appengine,!wasm

package cli

import (
	"golang.org/x/sys/unix"
)

func getTerminalColumns() int {
	ws, err := unix.IoctlGetWinsize(0, unix.TIOCGWINSZ)
	if err != nil {
		return 80
	}

	if ws.Col <= 0 {
		return 80
	}

	return int(ws.Col)
}
