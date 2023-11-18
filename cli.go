package main

import (
	"log"
	"os"
	"reflect"
	"regexp"
	"strings"

	"github.com/kr/pretty"
)

type flags struct {
	Help    bool     `cli:"h,help" description:"Show help"`
	IP      string   `cli:"ip" description:"IP address of the target"`
	Port    int      `cli:"p,port" description:"Port to use"`
	Targets []string `cli:"t,target" description:"Targets to scan"`
	Args    []string `description:"Extra arguments"`
}

func (f flags) Run() error {
	log.Println("RUN", f)
	return nil
}

type Filename string

func (f Filename) Validate() error {
	_, err := os.Stat(string(f))
	return err
}

type ConfigFlag struct {
	Config Filename `cli:"f,config" description:"config file to use"`
}

// manage running processes
type app struct {
	Daemon *struct {
		Start  *struct{} `cmd:"start,restart" description:"launch daemon process"`
		Stop   *struct{} `cmd:"stop,kill" description:"stop daemon process"`
		Run    *struct{} `cmd:"run" description:"run daemon server without daemonizing, DON'T USE BY HAND IF YOU DON'T KNOW WHAT YOU ARE DOING"`
		Status *struct{} `cmd:"status,ps" description:"check daemon status"`
		Logs   *struct {
			Follow bool `cli:"f, follow" description:"follow logs"`
		} `cmd:"logs" description:"check daemon logs"`
	} `cmd:"daemon,d" description:"manage daemon"`
	Run *struct {
		ConfigFlag
		Name  string         `cli:"n,name" description:"set a name for the process"`
		Tag   string         `cli:"t,tag" description:"add specified tag"`
		Cwd   string         `cli:"cwd" description:"set working directory"`
		Watch *regexp.Regexp `cli:"watch" description:"restart on changes to files matching specified regex"`
		Args  []string       `description:"process and args"`
	} `cmd:"run" description:"create and run new process" category:"management"`
	Start *struct {
		ConfigFlag
		Names []string `cli:"n,name" description:"name(s) of process(es) to start"`
		Tags  []string `cli:"t,tag" description:"tag(s) of process(es) to start"`
		ID    []uint64 `cli:"id" description:"id(s) of process(es) to start"`
	} `cmd:"start" description:"start already added process(es)" category:"management"`
	Server string `cli:"s,server" description:"server address"`
}

func tags(tag string) map[string]string {
	res := map[string]string{}
	i := 0
	for ; i < len(tag); i++ {
		j := i + 1
		d := 0
		for ; j < len(tag) && d < 2; j++ {
			if tag[j] == '\\' {
				j++
			} else if tag[j] == '"' {
				d++
			}
		}
		kv := tag[i:j]
		o := strings.IndexRune(kv, ':')
		k, v := kv[:o], kv[o+2:len(kv)-1]
		res[k] = v
		i = j
	}
	return res
}

type rawArgs struct {
	bools map[string]struct{}
	flags map[string][]string
	args  []string
}

type m struct {
	subcommands map[string]m
}

func parseCommands(t reflect.Type) m {
	res := m{subcommands: map[string]m{}}
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		tg := tags(string(f.Tag))
		if _, ok := tg["cmd"]; ok {
			subcommand := parseCommands(f.Type.Elem())
			for _, alias := range strings.Split(tg["cmd"], ",") {
				res.subcommands[alias] = subcommand
			}
		}
	}
	return res
}

func Parse[T any](args []string) {
	cmd := reflect.TypeOf((*T)(nil)).Elem()
	cmds := parseCommands(cmd)
	pretty.Println(cmds)
	for i, arg := range args {
		if cmdcmd, ok := cmds.subcommands[arg]; ok {
			cmds = cmdcmd
		} else {
			args = args[i+1:]
			break
		}
	}

	pretty.Println(cmds, args)
	// bools := map[string]struct{}{}
	// for i := 0; i < t.NumField(); i++ {
	// 	f := t.Field(i)
	// 	tg := tags(string(f.Tag))
	// 	if _, ok := tg["cli"]; ok { // flag
	// 		flagNames := strings.Split(tg["cli"], ",")
	// 		fmt.Println("FLAG", f.Name, "\t", f.Type.String())
	// 		if f.Type.Kind() == reflect.Bool {
	// 			for _, name := range flagNames {
	// 				bools[name] = struct{}{}
	// 			}
	// 		}
	// 	} else if _, ok := tg["cmd"]; ok { // command
	// 		aliases := strings.Split(tg["cli"], ",")
	// 		fmt.Println("COMMAND", aliases, f.Name, "\t", f.Type.String())
	// 	}
	// }

	// a := rawArgs{
	// 	bools: map[string]struct{}{},
	// 	flags: map[string][]string{},
	// 	args:  nil,
	// }
	// log.Println(a)
}

func main() {
	log.SetFlags(log.Lshortfile)
	Parse[app]([]string{
		"pm", "run",
		"--name", "azhopich",
	}[1:])
}
