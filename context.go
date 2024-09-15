package cli

type Context struct {
	Bools map[string]struct{}
	Flags map[string]string
	Args  []string
}

func isFlag(arg string) bool {
	isShort := len(arg) > 1 && arg[0] == '-' && arg[1] != '-'
	isLong := len(arg) > 2 && arg[0] == '-' && arg[1] == '-' && arg[2] != '-'
	return isShort || isLong
}

func newContext(args ...string) *Context {
	ctx := &Context{
		Bools: map[string]struct{}{},
		Flags: map[string]string{},
	}
	for i := 0; i < len(args); i++ {
		arg := args[i]
		if arg == "--" {
			if i+1 < len(args) {
				ctx.Args = append(ctx.Args, args[i+1:]...)
			}
			break
		}

		if !isFlag(arg) {
			ctx.Args = append(ctx.Args, arg)
			continue
		}

		name := arg[1:]
		if arg[1] == '-' {
			name = arg[2:]
		}

		if i+1 < len(args) && !isFlag(args[i+1]) {
			ctx.Flags[name] = args[i+1]
			i++
			continue
		}

		ctx.Bools[name] = struct{}{}
	}
	return ctx
}
