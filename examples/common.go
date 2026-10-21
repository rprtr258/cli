package examples

import "strconv"

func parseString(s string) (string, error) {
	return s, nil
}

func parseUint(s string) (uint, error) {
	u, err := strconv.ParseUint(s, 0, 64)
	return uint(u), err
}
