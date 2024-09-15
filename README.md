go-flags: a go library for parsing command line arguments
=========================================================

[![GoDoc](https://godoc.org/github.com/jessevdk/go-flags?status.png)](https://godoc.org/github.com/jessevdk/go-flags)
[![Build Status](https://travis-ci.org/jessevdk/go-flags.svg?branch=master)](https://travis-ci.org/jessevdk/go-flags)
[![Coverage Status](https://img.shields.io/coveralls/jessevdk/go-flags.svg)](https://coveralls.io/r/jessevdk/go-flags?branch=master)

This library provides similar functionality to the builtin flag library of go, but provides much more functionality and nicer formatting. From the documentation:

package cli provides an extensive command line fun.Option parser. The flags package is similar in functionality to the go builtin flag package but provides more fun.Options and uses reflection to provide a convenient and succinct way of specifying command line fun.Options.

## Supported features
The following features are supported:
- fun.Options with short names: `-v`
- fun.Options with long names: `--verbose`
- fun.Options with and without arguments (`bool` vs any other type)
- fun.Options with fun.Optional arguments and default values
- Multiple fun.Option groups each containing a set of fun.Options
- Generate and print well-formatted help message
- Passing remaining command line arguments after `--`
- Supports multiple short fun.Options: `-aux`
- Supports all primitive and standard library go types: `string`, `int{8..64}`, `uint{8..64}`, `float`, `time.Duration`, etc.
- Supports maps


## Basic usage
The flags package uses structs, reflection and struct field tags to allow users to specify command line fun.Options. This results in very simple and concise specification of your application fun.Options. For example:

```go
type fun.Options struct {
	Verbose []bool `short:"v" long:"verbose" description:"Show verbose debug information"`
}
```

This specifies one fun.Option with a short name `-v` and a long name `--verbose`. When either `-v` or `--verbose` is found on the command line, a `true` value will be appended to the `Verbose` field. e.g. when specifying `-vvv`, the resulting value of `Verbose` will be `{true, true, true}`.

## Example
```go
type cmdRoot struct {
	// Slice of bool will append 'true' each time the fun.Option
	// is encountered (can be set multiple times, like -vvv)
	Verbose []bool `short:"v" long:"verbose" description:"Show verbose debug information"`

	// Example of automatic marshalling to desired type (uint)
	Offset uint `long:"offset" description:"Offset"`

	// Example of a callback, called each time the fun.Option is found.
	Call func(string) `short:"c" description:"Call phone number"`

	// Example of a required flag
	Name string `short:"n" long:"name" description:"A name" required:"true"`

	// Example of a flag restricted to a pre-defined set of strings
	Animal string `long:"animal" choice:"cat" choice:"dog"`

	// Example of a value name
	File string `short:"f" long:"file" description:"A file" value-name:"FILE"`

	// Example of a pointer
	Ptr *int `short:"p" description:"A pointer to an integer"`

	// Example of a slice of strings
	StringSlice []string `short:"s" description:"A slice of strings"`

	// Example of a slice of pointers
	PtrSlice []*string `long:"ptrslice" description:"A slice of pointers to string"`

	// Example of a map
	IntMap map[string]int `long:"intmap" description:"A map from string to int"`

	Args []string `positional-args:"yes"`
}

func (x cmdRoot) Execute(context.Context) error {
	fmt.Println("Verbosity:", x.Verbose)
	fmt.Println("Offset:", x.Offset)
	fmt.Println("Name:", x.Name)
	fmt.Println("Animal:", x.Animal)
	fmt.Println("Ptr:", *x.Ptr)
	fmt.Println("StringSlice:", x.StringSlice)
	fmt.Println("PtrSlice:", *x.PtrSlice[0], *x.PtrSlice[1])
	fmt.Println("IntMap:", x.IntMap)
	fmt.Println("Remaining args:", x.Args)

	// Output: Verbosity: [true true]
	// Offset: 5
	// Name: Me
	// Ptr: 3
	// StringSlice: [hello world]
	// PtrSlice: [hello world]
	// IntMap: [a:1 b:5]
	// Remaining args: arg1 arg2 arg3
}

// Parse flags. Note that here we use flags.RunContext for
// the sake of making a working example. Normally, you would simply use
// flags.Parse(&opts) which uses os.Args
err := flags.RunContext[cmdRoot](
	"-vv",
	"--offset=5",
	"-n", "Me",
	"--animal", "dog", // anything other than "cat" or "dog" will raise an error
	"-p", "3",
	"-s", "hello",
	"-s", "world",
	"--ptrslice", "hello",
	"--ptrslice", "world",
	"--intmap", "a:1",
	"--intmap", "b:5",
	"arg1",
	"arg2",
	"arg3",
)
if err != nil {
	panic(err)
}
```

Slice fun.Options work exactly the same as primitive type fun.Options, except that whenever the fun.Option is encountered, a value is appended to the slice.

Map fun.Options from string to primitive type are also supported. On the command line, you specify the value for such an fun.Option as `key:value`. For example

```go
type fun.Options struct {
	AuthorInfo string[string] `short:"a"`
}
```

Then, the `AuthorInfo` map can be filled with something like `-a name:Jesse -a "surname:van den Kieboom"`.

Finally, for full control over the conversion between command line argument values and fun.Options, user defined types can choose to implement the Marshaler and Unmarshaler interfaces.


## Available field tags
|Tag|Description|
|-|-|
|`short`|the short name of the flag, single character|
|`long`|`the long name of the flag, at least two characters|
|`description`|the description of the flag (optional)|
|`long-description`|the long description of the flag. Currently only displayed in generated man pages (optional)|
|`default`|the default value of an flag. This tag can be specified multiple times in the case of slices or maps (optional)|
|`value-name`|the name of the argument value (to be shown in the help) (optional)|

Either the `short:` tag or the `long:` must be specified to make the field eligible as an
flag.


## Available positional args tags
|Tag|Description|
|-|-|
|`positional-args`|when specified on a field with a struct type, uses the fields of that struct to parse remaining positional command line arguments into (in order of the fields). If a field has a slice type, then all remaining arguments will be added to it. (optional)|
|`positional-arg-name`|used on a field in a positional argument struct; name of the positional argument placeholder to be shown in the help (optional)|


## Commands
Commands are often used in monolithic applications that support various commands or actions. Take git for example, all of the add, commit, checkout, etc. are called commands. Using commands you can easily separate multiple functions of your application.

To specify a command add a struct field to your fun.Options struct annotated with the `command:"command-name"` tag.

Use `alias:"alias-name"` tag to add aliases for command.

When parsing ends and there is an active command and that command implements the Commander interface, then its Execute method will be run with the remaining command line arguments.

Command structs can have fun.Options which become valid to parse after the command has been specified on the command line, in addition to the fun.Options of all the parent commands. I.e. considering a `-v` flag on the parser and an add command, the following are equivalent:

```bash
./app -v add
./app add -v
```

However, if the `-v` flag is defined on the add command, then the first of the two examples above would fail since the `-v` flag is not defined before the add command.

## Completion
There is builtin support to provide bash completion of flags, commands and argument values. To use completion, the binary which uses go-flags can be invoked in a special environment to list completion of the current command line argument. It should be noted that this `executes` your application, and it is up to the user to make sure there are no negative side effects (for example from init functions).

Setting the environment variable `GO_FLAGS_COMPLETION=1` enables completion by replacing the argument parsing routine with the completion routine which outputs completions for the passed arguments. The basic invocation to complete a set of arguments is therefore:

```bash
GO_FLAGS_COMPLETION=1 ./completion-example arg1 arg2 arg3
```

where `completion-example` is the binary, `arg1` and `arg2` are the current arguments, and `arg3` (the last argument) is the argument to be completed. If the `GO_FLAGS_COMPLETION` is set to "verbose", then descriptions of possible completion items will also be shown, if there are more than 1 completion items.

To use this with bash completion, a simple file can be written which calls the binary which supports go-flags completion:

```bash
_completion_example() {
	# All arguments except the first one
	args=("${COMP_WORDS[@]:1:$COMP_CWORD}")

	# Only split on newlines
	local IFS=$'\n'

	# Call completion (note that the first element of COMP_WORDS is
	# the executable itself)
	COMPREPLY=($(GO_FLAGS_COMPLETION=1 ${COMP_WORDS[0]} "${args[@]}"))
	return 0
}

complete -F _completion_example completion-example
```

Completion requires the parser fun.Option PassDoubleDash and is therefore enforced if the environment variable `GO_FLAGS_COMPLETION` is set.

Customized completion for argument values is supported by implementing the `flags.Completer` interface for the argument value type. An example of a type which does so is the `flags.Filename` type, an alias of string allowing simple filename completion. A slice or array argument value whose element type implements `flags.Completer` will also be completed.