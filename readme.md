## Usage

```go
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

// what next ???
```

## Similar workings
https://github.com/mkideal/cli
