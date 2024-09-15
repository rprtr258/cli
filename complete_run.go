package cli

import (
	"os"
	"strconv"
	"strings"

	"github.com/rprtr258/fun"
)

func dump_bash_completer(name string) {
	println(`_bpaf_dynamic_completion()
{{
    line="$1 --bpaf-complete-rev=8 ${{COMP_WORDS[@]:1}}"
    if [[ ${{COMP_WORDS[-1]}} == "" ]]; then
        line="${{line}} \"\""
    fi
    source <( eval ${{line}})
}}
complete -o nosort -F _bpaf_dynamic_completion ` + name)
}

func dump_zsh_completer(name string) {
	println(`#compdef ` + name + `
local line
line="${{words[1]}} --bpaf-complete-rev=7 ${{words[@]:1}}"
if [[ ${{words[-1]}} == "" ]]; then
    line="${{line}} \"\""
fi
source <(eval ${{line}})
`)
}

func dump_fish_completer(_ string) {
	println(`set -l current (commandline --tokenize --current-process)
set -l tmpline $current[1] --bpaf-complete-rev=9 $current[2..]
if test (commandline --current-process) != (string trim (commandline --current-process))
    set tmpline $tmpline ""
end
source ( $tmpline | psub )`)
}

// I would love to support elvish better but debugger is not a thing
// and on any error in code it simply replies "no candidates" with no
// obvious way even to print "you are here"...
// https://github.com/elves/elvish/issues/803
func dump_elvish_completer(name string) {
	rev := "1"
	println(`set edit:completion:arg-completer[` + name + `] = {{ |@args| var args = $args[1..];
     var @lines = ( ` + name + ` --bpaf-complete-rev=` + rev + ` $@args );
     use str;
     for line $lines {{
         var @arg = (str:split \"\\t\" $line)
         try {{
             edit:complex-candidate $arg[0] &display=( printf \"%-19s %s\" $arg[0] $arg[1] )
         }} catch {{
             edit:complex-candidate $line
         }}
     }}
}}`)
}

type ArgScanner struct {
	revision fun.Option[uint]
	name     fun.Option[string]
}

func (self *ArgScanner) check_next(arg string) bool {
	// this only works when there's a name
	if name, ok := self.name.Unpack(); ok {
		matched := true
		switch arg {
		case "--bpaf-complete-style-zsh":
			dump_zsh_completer(name)
		case "--bpaf-complete-style-bash":
			dump_bash_completer(name)
		case "--bpaf-complete-style-fish":
			dump_fish_completer(name)
		case "--bpaf-complete-style-elvish":
			dump_elvish_completer(name)
		default:
			matched = false
		}
		if matched {
			os.Exit(0)
		}
	}
	if ver, ok := strings.CutPrefix(arg, "--bpaf-complete-rev="); ok {
		if ver, err := strconv.Atoi(ver); err == nil {
			self.revision = fun.Valid(uint(ver))
		}
		return true
	}
	return false
}

func (self ArgScanner) done() fun.Option[Complete] {
	revision, ok := self.revision.Unpack()
	if !ok {
		return fun.Invalid[Complete]()
	}
	return fun.Valid(Completenew(revision))
}
