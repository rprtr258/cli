package flags

import "strconv"

func scanMultiTag(value string) (map[string][]string, error) {
	ret := map[string][]string{}
	for v := value; v != ""; {
		i := 0

		// Skip whitespace
		for i < len(v) && v[i] == ' ' {
			i++
		}
		v = v[i:]
		if v == "" {
			break
		}

		// Scan to colon to find key
		i = 0
		for i < len(v) && v[i] != ' ' && v[i] != ':' && v[i] != '"' {
			i++
		}
		if i >= len(v) {
			return nil, newErrorf(ErrTag, "expected `:' after key name, but got end of tag (in `%v`)", value)
		}
		if v[i] != ':' {
			return nil, newErrorf(ErrTag, "expected `:' after key name, but got `%v' (in `%v`)", v[i], value)
		}
		if i+1 >= len(v) {
			return nil, newErrorf(ErrTag, "expected `\"' to start tag value at end of tag (in `%v`)", value)
		}
		if v[i+1] != '"' {
			return nil, newErrorf(ErrTag, "expected `\"' to start tag value, but got `%v' (in `%v`)", v[i+1], value)
		}
		name := v[:i]
		v = v[i+1:]

		// Scan quoted string to find value
		i = 1
		for i < len(v) && v[i] != '"' {
			if v[i] == '\n' {
				return nil, newErrorf(ErrTag, "unexpected newline in tag value `%v' (in `%v`)", name, value)
			}

			if v[i] == '\\' {
				i++
			}
			i++
		}
		if i >= len(v) {
			return nil, newErrorf(ErrTag, "expected end of tag value `\"' at end of tag (in `%v`)", value)
		}

		val, err := strconv.Unquote(v[:i+1])
		if err != nil {
			return nil, newErrorf(ErrTag, "Malformed value of tag `%v:%v` => %v (in `%v`)", name, v[:i+1], err, value)
		}

		v = v[i+1:]

		ret[name] = append(ret[name], val)
	}
	return ret, nil
}

func ParseMultiTag(value string) (map[string][]string, error) {
	vals, err := scanMultiTag(value)
	return vals, err
}

func multitagGet(x map[string][]string, key string) string {
	if v, ok := x[key]; ok {
		return v[len(v)-1]
	}

	return ""
}

func GetMany(x map[string][]string, key string) []string {
	return x[key]
}
