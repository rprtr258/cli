package cli

import (
	"os"
	"strconv"
	"strings"
)

func dump_bash_completer(name string) {
	println(`_cli_dynamic_completion()
{{
    line="$1 --cli-complete-rev=8 ${{COMP_WORDS[@]:1}}"
    if [[ ${{COMP_WORDS[-1]}} == "" ]]; then
        line="${{line}} \"\""
    fi
    source <( eval ${{line}})
}}
complete -o nosort -F _cli_dynamic_completion ` + name)
}

func dump_zsh_completer(name string) {
	println(`#compdef ` + name + `
local line
line="${{words[1]}} --cli-complete-rev=7 ${{words[@]:1}}"
if [[ ${{words[-1]}} == "" ]]; then
    line="${{line}} \"\""
fi
source <(eval ${{line}})
`)
}

func dump_fish_completer(_ string) {
	println(`set -l current (commandline --tokenize --current-process)
set -l tmpline $current[1] --cli-complete-rev=9 $current[2..]
if test (commandline --current-process) != (string trim (commandline --current-process))
    set tmpline $tmpline ""
end
source ( $tmpline | psub )`)
}

// dump_elvish_completer dumps the elvish completion script.
//
// Elvish completion is limited: the runtime offers no way to debug the
// completion script, and any error in it silently results in no candidates
// with no diagnostics. See https://github.com/elves/elvish/issues/803.
func dump_elvish_completer(name string) {
	rev := "1"
	println(`set edit:completion:arg-completer[` + name + `] = ` + "{{" + ` |@args| var args = $args[1..];
     var @lines = ( ` + name + ` --cli-complete-rev=` + rev + ` $@args );
     use str;
     for line $lines ` + "{{" + `
         var @arg = (str:split \"\\t\" $line)
         try ` + "{{" + `
             edit:complex-candidate $arg[0] &display=( printf \"%-19s %s\" $arg[0] $arg[1] )
         }} catch {{
             edit:complex-candidate $line
         }}
     }}
}}`)
}

type argScanner struct {
	revision Opt[uint]
	name     Opt[string]
}

func (self *argScanner) check_next(arg string) bool {
	// Style requests only work when the parser name is known.
	if name, ok := self.name.Unpack(); ok {
		matched := true
		switch arg {
		case "--cli-complete-style-zsh":
			dump_zsh_completer(name)
		case "--cli-complete-style-bash":
			dump_bash_completer(name)
		case "--cli-complete-style-fish":
			dump_fish_completer(name)
		case "--cli-complete-style-elvish":
			dump_elvish_completer(name)
		default:
			matched = false
		}
		if matched {
			// The dumped script is all the shell needs, so stop here.
			os.Exit(0)
		}
	}
	if ver, ok := strings.CutPrefix(arg, "--cli-complete-rev="); ok {
		if ver, err := strconv.Atoi(ver); err == nil {
			self.revision = Valid(uint(ver))
		}
		return true
	}
	return false
}

func (self argScanner) done() (Complete, bool) {
	revision, ok := self.revision.Unpack()
	if !ok {
		return Complete{}, false
	}
	return newComplete(revision), true
}
