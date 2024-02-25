package flags

import (
	"fmt"
	"strings"
)

// Duration flag type accepts strings like "1y 2mo 3w", "4d 5h 6m 7s 8ms 9us 10ns"
type Duration struct {
	Years, Months, Weeks, Days              int
	Hours, Minutes, Seconds                 int
	Milliseconds, Microseconds, Nanoseconds int
}

func (d *Duration) UnmarshalFlag(s string) error {
	parts := strings.Split(s, " ")
	for _, part := range parts {
		if part == "" {
			continue
		}

		var num int
		sig := 1
		if part[0] == '-' {
			sig = -1
			part = part[1:]
		}
		for len(part) > 0 && part[0] >= '0' && part[0] <= '9' {
			num = num*10 + int(part[0]-'0')
			part = part[1:]
		}
		num *= sig

		switch part {
		case "y":
			d.Years += num
		case "mo":
			d.Months += num
		case "w":
			d.Weeks += num
		case "d":
			d.Days += num
		case "h":
			d.Hours += num
		case "m":
			d.Minutes += num
		case "s":
			d.Seconds += num
		case "ms":
			d.Milliseconds += num
		case "us":
			d.Microseconds += num
		case "ns":
			d.Nanoseconds += num
		default:
			return fmt.Errorf("cant parse part %q", part)
		}
	}

	return nil
}
