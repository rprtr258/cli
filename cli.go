package cli

import (
	"fmt"
	"reflect"
	"regexp"
)

type tags struct {
	isPositional bool

	short       byte
	long        string
	description string
}

type base struct {
	args      []string
	flags     map[string][]string
	boolFlags map[string]struct{}
}

func Parse[T any](args ...string) (T, error) {
	var dest T
	typ := reflect.TypeOf(dest)
	if typ.Kind() != reflect.Struct {
		return dest, fmt.Errorf("%s is not a struct", typ.String())
	}

	isBoolFlag := map[string]bool{}
	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		if !field.IsExported() {
			continue
		}

		rawTags := string(field.Tag)
		tagsMatches := regexp.MustCompile(`(short|long|description):"([^"]*)"|positional`).FindAllStringSubmatch(rawTags, -1)

		var tags tags
		for _, tagMatch := range tagsMatches {
			switch {
			case tagMatch[0] == "positional":
				tags.isPositional = true
			case tagMatch[1] == "short":
				tags.short = tagMatch[2][0]
			case tagMatch[1] == "long":
				tags.long = tagMatch[2]
			case tagMatch[1] == "description":
				tags.description = tagMatch[2]
			}
		}

		isBoolFlag[field.Name] = field.Type.Kind() == reflect.Bool
	}

	argss := []string{}
	flags := map[string][]string{}
	boolFlags := map[string]struct{}{}
LOOP:
	for i := 1; i < len(args); i++ {
		switch arg := args[i]; {
		case arg == "--": // rest of arguments uninterpreted
			if i+1 < len(args) {
				argss = append(argss, args[i+1:]...)
			}
			break LOOP
		case arg[0] == '-': // flag
			switch {
			case arg == "-": // actually not a flag
				argss = append(argss, arg)
			case arg[1] != '-': // short flag
				return dest, fmt.Errorf("short flags are not supported for now: %s", arg)
			case arg[2] == '-': // ---aboba, not a flag
				argss = append(argss, arg)
			default: // long flag
				flagName := arg[2:]
				if isBoolFlag[flagName] {
					boolFlags[flagName] = struct{}{}
				} else {
					i++
					if i == len(args) {
						return dest, fmt.Errorf("flag %s requires value", arg)
					}

					flags[flagName] = append(flags[flagName], args[i])
				}
			}
		default: // argument
			argss = append(argss, arg)
		}
	}
	base := base{
		args:      argss,
		flags:     flags,
		boolFlags: boolFlags,
	}

	fmt.Println(base)
	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		if !field.IsExported() {
			continue
		}

		// TODO: instead of field.Name we should use field.Tag
		isBoolFlag[field.Name] = field.Type.Kind() == reflect.Bool
	}

	return dest, nil
}
