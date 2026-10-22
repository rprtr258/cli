// A somewhat fancy example of a typical usage of this library.
//
// Most important thing is probably a notion of "A pair of values, where each value can be one of
// two different values", here it's distance which can be miles and kilometers and speed - miles
// and kilometers per hours. Generated usage line can give a better hint to what's going on
package examples

import (
	"fmt"
	"os"

	"github.com/rprtr258/cli"
)

func Travel() {
	type Segment func(
		Time func(float64),
		SpeedDistance func(speed, dist float64),
	)

	SegmentTime := func(x float64) Segment {
		return func(f func(float64), _ func(_, _ float64)) {
			if f != nil {
				f(x)
			}
		}
	}
	SegmentSpeedDistance := func(speed, dist float64) Segment {
		return func(_ func(float64), f func(speed, dist float64)) {
			if f != nil {
				f(speed, dist)
			}
		}
	}

	// Suppose we want to calcualte a total time of a travel, where parts of
	// a travel can be given either as pairs of speed and distance or just by time.
	// Speed can be given by KPH or MPH. Distance - either miles or km.

	// parsers for speeds. Both speeds are converted to the same units
	mph := cli.Argument(cli.Long("mph").
		Help("speed in MPH"),
		"SPEED", cli.ParseFloat).
		Map(func(x float64) float64 { return x * 1.6 })
	kph := cli.Argument(cli.Long("kph").Help("Speed in KPH"), "SPEED", cli.ParseFloat)

	// speed is either kph or mph, conversion to mph is handled by the parser
	speed := cli.Choice(mph, kph)

	// parsers for distances, both are converted to the same units
	km := cli.Argument(cli.Long("km").Help("Distance in KM"), "KMs", cli.ParseFloat)
	mi := cli.Argument(cli.Long("mi").
		Help("distance in miles"),
		"MILES", cli.ParseFloat).
		Map(func(x float64) float64 { return x * 1.6 })
	dist := cli.Choice(mi, km)

	// time, presumably in seconds
	time := cli.Argument(cli.Long("time").
		Help("Travel time in hours"),
		"TIME", cli.ParseFloat)

	// parsed time is trivially converted to time segment
	segment_time := time.Map(SegmentTime)

	// parsed speed/distance is packed into SpeedDistance segment
	segment_speed := cli.Map2(speed, dist, SegmentSpeedDistance)

	// segment can be either of two defined
	segment := cli.Choice(segment_speed, segment_time)

	// and we have several of them.
	parser := segment.Many().
		Guard(func(x []Segment) bool { return len(x) > 0 }, "need at least one segment").
		Guard(func(x []Segment) bool { return len(x) < 10 }, "for more than 9 segments you need to purchase a premium subscription")

	descr := "Accepts one or more travel segments"
	header := `You need to specify one or more travel segments, segment is defined by
a pair of speed and distance or by time.

This example defines two separate travel segments, one given by speed/distance combo and one by time
    travel --km 180 --kph 35 --time`
	decorated := parser.ToOptions().Description(descr).Header(header)

	// help message tries to explain what's needed:
	// either --time OR one speed and one distance, both can be given in miles or km.
	// number of speed flags must correspond to number of distance flags, more or
	// less results in parser error messages
	opt := decorated.Run(os.Args)

	fmt.Println(opt)
}
