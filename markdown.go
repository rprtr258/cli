package cli

// import (
// 	"io"
// 	"os"
// 	"testing"

// 	"github.com/stretchr/testify/require"
// )

// func write_updated(new_val string, path string) (bool, error) {
// 	file, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE, 0o644)
// 	if err != nil {
// 		return false, err
// 	}
// 	defer file.Close()
// 	current_val, err := io.ReadAll(file)
// 	if err != nil {
// 		return false, err
// 	}
// 	if string(current_val) != new_val {
// 		if err := os.WriteFile(path, []byte(new_val), 0o644); err != nil {
// 			return false, err
// 		}
// 		return false, nil
// 	}
// 	return true, nil
// }

// func TestSimple(t *testing.T) {
// 	kraken := Short('d').
// 		Long("kraken").
// 		Help("Unleash the kraken").
// 		Switch()

// 	user := Argument(Long("user").
// 		Env("USER").
// 		Help("Log in as this user"),
// 		"USER", parseString)

// 	options := Pair(kraken, user).
// 		ToOptions().
// 		descr("I am a program and I do things").
// 		header("Sometimes they even work.").
// 		footer("Beware `-d`, dragons be here")
// 	roff := options.render_markdown("simple")

// 	ok, err := write_updated(&roff, "tests/markdown.md")
// 	require.NoError(t, err)
// 	require.True(t, ok)
// }

// // func TestNested(t*testing.T) {
// //     #[derive(Debug, Clone, Bpaf)]
// //     /// Options
// //     #[bpaf(options)]
// //     enum Options {
// //         #[bpaf(command)]
// //         /// Alpha
// //         Alpha,

// //         #[bpaf(command)]
// //         /// Beta
// //         ///
// //         /// More Beta
// //         ///
// //         /// Even More Beta
// //         Beta,
// //     }

// //     r := options().render_markdown("options");
// //     expected := "

// // # Command summary

// //   * [`options`↴](#options)
// //   * [`options alpha`↴](#options-alpha)
// //   * [`options beta`↴](#options-beta)

// // ## options

// // Options

// // **Usage**: **`options`** _`COMMAND ...`_

// // **Available options:**
// // - **`-h`**, **`--help`** &mdash; \n  Prints help information

// // **Available commands:**
// // - **`alpha`** &mdash; \n  Alpha
// // - **`beta`** &mdash; \n  Beta

// // ## options alpha

// // Alpha

// // **Usage**: **`options`** **`alpha`** \n
// // **Available options:**
// // - **`-h`**, **`--help`** &mdash; \n  Prints help information

// // ## options beta

// // Beta

// // More Beta

// // Even More Beta

// // **Usage**: **`options`** **`beta`** \n
// // **Available options:**\n- **`-h`**, **`--help`** &mdash; \n  Prints help information

// // ";
// //     require.Equal(t,r, expected);
// // }

// func TestMulti_line_help(t *testing.T) {
// 	opts := Short('a').Help("help\n\nmore help").Switch().ToOptions()
// 	r := opts.render_markdown("ml")

// 	expected := `# ml

// **Usage**: **` + "`" + `ml` + "`" + `** \\[**` + "`" + `-a` + "`" + `**\\]

// **Available options:**
// - **` + "`" + `-a` + "`" + `** &mdash; \n  help

//   more help
// - **` + "`" + `-h` + "`" + `**, **` + "`" + `--help` + "`" + `** &mdash; \n  Prints help information

// `
// 	require.Equal(t, r, expected)
// }

// func TestNo_help(t *testing.T) {
// 	a := short('a').help("help").Switch()
// 	b := short('b').Switch()
// 	opts := Pair(a, b).to_options()

// 	r := opts.render_markdown("ml")

// 	expected := `# ml

// **Usage**: **` + "`" + `ml` + "`" + `** \\[**` + "`" + `-a` + "`" + `**\\] \\[**` + "`" + `-b` + "`" + `**\\]

// **Available options:**
// - **` + "`" + `-a` + "`" + `** &mdash; \n  help
// - **` + "`" + `-b` + "`" + `**\n- **` + "`" + `-h` + "`" + `**, **` + "`" + `--help` + "`" + `** &mdash; \n  Prints help information\n\n\n`

// 	require.Equal(t, r, expected)
// }

// // func TestCodeblock_space_help(t*testing.T) {
// //     #[derive(Bpaf, Clone, Debug)]
// //     #[bpaf(options)]
// //     struct Options {
// //         /// Verbose help
// //         ///
// //         ///     block
// //         ///     of
// //         ///     code
// //         verbose: bool,
// //     }

// //     r := options().run_inner(newArgs("-hh")).unwrap_err().unwrap_stdout();
// //     expected := "\
// // Usage: [--verbose]

// // Available options:
// //         --verbose  Verbose help
// //                    block
// //                    of
// //                    code
// //     -h, --help     Prints help information
// // ";

// //     require.Equal(t,r, expected);

// //     r := options().render_markdown("ml");
// //     expected := "\
// // # ml

// // **Usage**: **`ml`** \\[**`--verbose`**\\]

// // **Available options:**
// // - **`    --verbose`** &mdash; \n  Verbose help

// //   \n
// //   ```text
// //   block
// //   of
// //   code
// //   ```

// // - **`-h`**, **`--help`** &mdash; \n  Prints help information\n\n\n";

// //     require.Equal(t,r, expected);
// // }

// // func TestCodeblock_ticks_help(t*testing.T) {
// //     #[derive(Bpaf, Clone, Debug)]
// //     #[bpaf(options)]
// //     struct Options {
// //         /// Verbose help
// //         ///
// //         /// ```text
// //         /// block
// //         ///     of
// //         ///     code
// //         /// ```
// //         verbose: bool,
// //     }

// //     r := options().run_inner(newArgs("-hh")).unwrap_err().unwrap_stdout();
// //     expected := "\
// // Usage: [--verbose]

// // Available options:
// //         --verbose  Verbose help
// //                    ```text
// //                    block
// //                        of
// //                        code
// //                    ```
// //     -h, --help     Prints help information
// // ";
// //     require.Equal(t,r, expected);

// //     r := options().render_markdown("ml");
// //     expected := "\
// // # ml

// // **Usage**: **`ml`** \\[**`--verbose`**\\]

// // **Available options:**
// // - **`    --verbose`** &mdash; \n  Verbose help

// //   \n  ```text
// //   block
// //       of
// //       code
// //   ```

// // - **`-h`**, **`--help`** &mdash; \n  Prints help information\n\n\n";

// //     require.Equal(t,r, expected);
// // }
