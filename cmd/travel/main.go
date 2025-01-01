// A somewhat fancy example of a typical `bpaf` usage.
//
// Most important thing is probably a notion of "A pair of values, where each value can be one of
// two different values", here it's distance which can be miles and kilometers and speed - miles
// and kilometers per hours. Generated usage line can give a better hint to what's going on
package main

import (
	"fmt"
	"os"
	"strconv"

	"github.com/rprtr258/cli"
)

type Segment func(
	Time func(float64),
	SpeedDistance func(speed, dist float64),
)

func SegmentTime(x float64) Segment {
	return func(f func(float64), _ func(_, _ float64)) {
		if f != nil {
			f(x)
		}
	}
}
func SegmentSpeedDistance(speed, dist float64) Segment {
	return func(_ func(float64), f func(speed, dist float64)) {
		if f != nil {
			f(speed, dist)
		}
	}
}

func parseFloat64(s string) (float64, error) {
	return strconv.ParseFloat(s, 64)
}

func main() {
	// Suppose we want to calcualte a total time of a travel, where parts of
	// a travel can be given either as pairs of speed and distance or just by time.
	// Speed can be given by KPH or MPH. Distance - either miles or km.

	// parsers for speeds. Both speeds are converted to the same units
	mph := cli.Map(cli.Argument(cli.Long("mph").
		Help("speed in MPH"),
		"SPEED", parseFloat64).Parser(),
		func(x float64) float64 { return x * 1.6 })
	kph := cli.Argument(cli.Long("kph").Help("Speed in KPH"), "SPEED", parseFloat64)

	// speed is either kph or mph, conversion to mph is handled by the parser
	speed := cli.Choice(mph, kph.Parser())

	// parsers for distances, both are converted to the same units
	km := cli.Argument(cli.Long("km").Help("Distance in KM"), "KMs", parseFloat64)
	mi := cli.Map(cli.Argument(cli.Long("mi").
		Help("distance in miles"),
		"MILES", parseFloat64).Parser(),
		func(x float64) float64 { return x * 1.6 })
	dist := cli.Choice(mi, km.Parser())

	// time, presumably in seconds
	time := cli.Argument(cli.Long("time").
		Help("Travel time in hours"),
		"TIME", parseFloat64)

	// parsed time is trivially converted to time segment
	segment_time := cli.Map(time.Parser(), SegmentTime)

	// parsed speed/distance is packed into SpeedDistance segment
	segment_speed := cli.Map2(speed, dist, SegmentSpeedDistance)

	// segment can be either of two defined
	segment := cli.Choice(segment_speed, segment_time)

	// and we have several of them.
	parser := cli.Many(segment, false).Parser().
		Guard(func(x []Segment) bool { return len(x) > 0 }, "need at least one segment").
		Parser().
		Guard(func(x []Segment) bool { return len(x) < 10 }, "for more than 9 segments you need to purchase a premium subscription")

	descr := "Accepts one or more travel segments"
	header := `You need to specify one or more travel segments, segment is defined by
a pair of speed and distance or by time.

This example defines two separate travel segments, one given by speed/distance combo and one by time
    travel --km 180 --kph 35 --time`
	decorated := parser.Parser().ToOptions().Descr(descr).Header(header)

	// help message tries to explain what's needed:
	// either --time OR one speed and one distance, both can be given in miles or km.
	// number of speed flags must correspond to number of distance flags, more or
	// less results in parser error messages
	opt := decorated.Run(os.Args)

	fmt.Println(opt)
}
