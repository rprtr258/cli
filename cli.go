package cli

import (
	"fmt"
	"reflect"
	"regexp"
)

type fieldID int

type flagMeta struct {
	isRest bool
	isBool bool

	idx fieldID

	short       byte
	long        string
	description string
}

type base struct {
	args      []string
	flags     map[fieldID][]string
	boolFlags map[fieldID]struct{}
}

func Parse[T any](args ...string) (T, error) {
	var dest T
	typ := reflect.TypeOf(dest)
	if typ.Kind() != reflect.Struct {
		return dest, fmt.Errorf("%s is not a struct", typ.String())
	}

	flagsMeta := map[fieldID]flagMeta{}
	shorts := map[byte]fieldID{}
	longs := map[string]fieldID{}
	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		if !field.IsExported() {
			continue
		}

		id := fieldID(i)

		rawTags := string(field.Tag)
		tagsMatches := regexp.MustCompile(`(short|long|description):"([^"]*)"|rest`).FindAllStringSubmatch(rawTags, -1)

		tags := flagMeta{
			isBool: field.Type.Kind() == reflect.Bool,
			idx:    id,
		}
		for _, tagMatch := range tagsMatches {
			switch {
			case tagMatch[0] == "rest":
				tags.isRest = true
			case tagMatch[1] == "short":
				tags.short = tagMatch[2][0]
			case tagMatch[1] == "long":
				tags.long = tagMatch[2]
			case tagMatch[1] == "description":
				tags.description = tagMatch[2]
			}
		}

		flagsMeta[fieldID(i)] = tags

		if tags.short != 0 {
			shorts[tags.short] = id
		}
		if tags.long != "" {
			longs[tags.long] = id
		}
	}

	argss := []string{}
	flags := map[fieldID][]string{}
	boolFlags := map[fieldID]struct{}{}
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
				j := longs[arg[2:]]
				if flagsMeta[j].isBool {
					boolFlags[j] = struct{}{}
				} else {
					i++
					if i == len(args) {
						return dest, fmt.Errorf("flag %s requires value", arg)
					}

					flags[j] = append(flags[j], args[i])
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
	val := reflect.ValueOf(&dest)
	for i := 0; i < typ.NumField(); i++ {
		field := val.Elem().Field(i)
		fieldStruct := typ.Field(i)
		if !fieldStruct.IsExported() {
			continue
		}

		id := fieldID(i)

		var meta flagMeta
		for _, m := range flagsMeta {
			if m.idx == id {
				meta = m
			}
		}

		switch {
		case meta.isRest:
			field.Set(reflect.ValueOf(base.args))
		case fieldStruct.Type.Kind() == reflect.Bool:
			_, ok := base.boolFlags[id]
			field.Set(reflect.ValueOf(ok))
		case fieldStruct.Type.Kind() == reflect.String:
			switch vals := base.flags[id]; len(vals) {
			case 0:
				return dest, fmt.Errorf("flag --%s requires value", meta.long)
			case 1:
				field.Set(reflect.ValueOf(vals[0]))
			default:
				return dest, fmt.Errorf("flag --%s requires exactly one value", meta.long)
			}
		}
	}

	return dest, nil
}
