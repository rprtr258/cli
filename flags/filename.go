package flags

import (
	"fmt"
	"os"
)

type Filename string

func (f *Filename) Set(s string) error {
	// TODO: if s == "-" use stdin

	if stat, err := os.Stat(s); os.IsNotExist(err) {
		return fmt.Errorf("no such file: %s", s)
	} else if !stat.Mode().IsRegular() {
		return fmt.Errorf("not a file: %s", s)
	}

	*f = Filename(s)
	return nil
}
