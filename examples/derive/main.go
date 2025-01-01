// pretty basic derive example with external function
package derive

// import (
// 	"fmt"

// 	"github.com/rprtr258/cli"
// )

// type PathBuf string

// // #[derive(Bpaf)]
// // #[bpaf(options, version)]
// type Opts struct {
// 	// Activate debug mode
// 	// #[bpaf(short, long)]
// 	debug bool
// 	// this comment is ignored
// 	// #[bpaf(external(verbose))]
// 	verbose uint
// 	// Set speed
// 	// #[bpaf(argument("SPEED"), fallback(42.0))]
// 	speed float64
// 	// Output file
// 	output PathBuf

// 	// #[bpaf(guard(positive, "must be positive"), fallback(1))]
// 	nb_cars          uint32
// 	files_to_process []PathBuf
// }

// func verbose() cli.Parser[uint] {
// 	// number of occurrences of the v/verbose flag capped at 3
// 	return cli.Map(cli.Many(cli.ReqFlag(cli.Short('v').
// 		Long("verbose").
// 		Help(`Increase the verbosity
// You can specify it up to 3 times
// either as -v -v -v or as -vvv`),
// 		struct{}{})),
// 		func(xs []struct{}) uint { return uint(len(xs)) }).
// 		Guard(func(x uint) bool { return x <= 3 }, "It doesn't get any more verbose than this").
// 		Parser()
// }

// func positive(input uint32) bool {
// 	return input > 1
// }

// func Main() {
// 	fmt.Println(opts().run())
// }
