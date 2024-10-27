// I have group of options which could be used multiple times. All such groups should be kept without overwriting previous one.
//
//	$ prometheus_sensors_exporter \
//	    \
//	    `# 2 physical sensors located on physycial different i2c bus or address` \
//	    --sensor \
//	        --sensor-device=tmp102 \
//	        --sensor-name="temperature_tmp102_outdoor" \
//	        --sensor-i2c-bus=0 \
//	        --sensor-i2c-address=0x48 \
//	    --sensor \
//	        --sensor-device=tmp102 \
//	        --sensor-name="temperature_tmp102_indoor" \
//	        --sensor-i2c-bus=1 \
//	        --sensor-i2c-address=0x49 \
package main

import (
	"fmt"
	"iter"
	"strconv"
	"strings"

	"github.com/rprtr258/cli"
	"github.com/rprtr258/fun"
)

type Sensor struct {
	sensor  struct{}
	device  string
	name    string
	bus_id  uint
	address uint
}

type Opts struct {
	sensors []Sensor
}

// generates completion suggestions
func sensor_device_comp(input string) iter.Seq2[string, cli.Option[string]] {
	return func(yield func(string, cli.Option[string]) bool) {
		for name, descr := range func(yield func(string, string) bool) {
			_ = yield("temp100", "Main temperature sensor") &&
				yield("temp101", "Output temperature sensor") &&
				yield("tank01", "Temperature in a storage tank 1") &&
				yield("tank02", "Temperature in a storage tank 2") &&
				yield("tank03", "Temperature in a storage tank 3") &&
				yield("outdoor", "Outdoor temperature sensor")
		} {
			if strings.HasPrefix(name, input) && yield(name, cli.Valid(descr)) {
				return
			}
		}
	}
}

func parseUint(s string) (uint, error) {
	i, err := strconv.ParseUint(s, 0, 64)
	return uint(i), err
}

func parseString(s string) (string, error) {
	return s, nil
}

var opts cli.Parser[Opts] = func() cli.Parser[Opts] {
	sensor := cli.ReqFlag(cli.Long("sensor"), struct{}{})
	device := cli.Argument[string](cli.Long("sensor-device"),
		"DEVICE", parseString).
		Parser().
		Complete(sensor_device_comp)
	name := cli.Argument[string](cli.Long("sensor-name"), "NAME", parseString)

	// from_str needs to be replaced with `parse` that can deal with hex digits
	bus_id := cli.Argument(cli.Long("sensor-i2c-bus"), "BUS", parseUint)
	address := cli.Argument(cli.Long("sensor-i2c-address"), "ADDRESS", parseUint)
	sensors := cli.Many(cli.Map(
		cli.Pair(
			cli.Pair(
				cli.Pair(
					sensor.Parser(),
					device.Parser(),
				).Parser(),
				cli.Pair(
					name.Parser(),
					bus_id.Parser(),
				).Parser(),
			).Parser(),
			address.Parser(),
		).Adjacent().Parser(), func(x fun.Pair[fun.Pair[fun.Pair[struct{}, string], fun.Pair[string, uint]], uint]) Sensor {
			return Sensor{x.K.K.K, x.K.K.V, x.K.V.K, x.K.V.V, x.V}
		})) // TODO: adjacent must be after Map somehow
	return cli.Map(sensors, func(x []Sensor) Opts { return Opts{x} })
}()

func main() {
	fmt.Println(opts.Run())
}
