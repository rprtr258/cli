package cli

import (
	"os"
	"strings"
)

// parseMap is a parser for a repeated named argument accumulating `key:value`
// entries into a map, created with [NamedArg.Map].
type parseMap[K comparable, V any] struct {
	named    NamedArg
	metavar  Metavar
	adjacent bool
	key      func(string) (K, error)
	value    func(string) (V, error)
}

func (self parseMap[K, V]) Parser() Parser[map[K]V] {
	meta := meta(metaSkip{})
	if item, ok := self.item(); ok {
		meta = metaItem(item)
	}

	return Parser[map[K]V]{self.eval, meta}
}

// ToOptions prepares the parser for running, see [Parser.ToOptions].
func (self parseMap[K, V]) ToOptions() *OptionParser[map[K]V] { return self.Parser().ToOptions() }

// Adjacent restricts parsed entries to have both the flag and a value in the same word.
//
// In other words, an adjacent-restricted [NamedArg.Map] would accept `--flag=key:value`
// or `-fbar:baz` but not `--flag key:value`. Note, this is different from [ParseCon.Adjacent],
// it just plays a similar role.
func (self *parseMap[K, V]) Adjacent() *parseMap[K, V] {
	self.adjacent = true
	return self
}

func (self parseMap[K, V]) item() (item, bool) {
	name, ok := shortLongTryFrom(self.named)
	if !ok {
		return item{}, false
	}
	return item{self.named.help, itemArgument{
		name,
		self.named.short,
		Metavar(self.metavar),
		first(self.named.env),
		nil,
	}}, true
}

func (self parseMap[K, V]) eval(state *state) (map[K]V, parseError) {
	res := make(map[K]V)
	flagSeen := false
	for {
		valueStrOpt, err := state.takeFlag(self.named, self.adjacent, self.metavar)
		if err != nil {
			state.push_argument(self.named, self.metavar)
			return nil, err
		}

		valueStr, ok := valueStrOpt.Unpack()
		if !ok {
			state.push_argument(self.named, self.metavar)
			// the map flag is optional: absence gives an empty map;
			// the env variable is a fallback used only when no flags are present
			state.currentIdx = Invalid[int]()
			if !flagSeen {
				if entries, ok := find_map(self.named.env, os_lookupenv_split_entries); ok {
					for _, entry := range entries {
						if err := self.put(state, res, entry); err != nil {
							return nil, err
						}
					}
				}
			}
			return res, nil
		}
		flagSeen = true

		if state.touching_last_remove() {
			state.push_metavar(self.metavar, self.named.help, true)
		}

		if err := self.put(state, res, valueStr); err != nil {
			return nil, err
		}
	}
}

// put splits one `key:value` entry and puts it into the accumulating map.
func (self parseMap[K, V]) put(state *state, res map[K]V, entry string) parseError {
	rawKey, rawValue, ok := strings.Cut(entry, ":")
	if !ok {
		return errorParseFailed{state.currentIdx, "expected `key:value`"}
	}

	k, err := self.key(rawKey)
	if err != nil {
		return errorParseFailed{state.currentIdx, err.Error()}
	}
	v, err := self.value(rawValue)
	if err != nil {
		return errorParseFailed{state.currentIdx, err.Error()}
	}
	res[k] = v
	return nil
}

func os_lookupenv_split_entries(variable string) ([]string, bool) {
	value, ok := os.LookupEnv(variable)
	if !ok {
		return nil, false
	}
	if value == "" {
		return nil, true
	}
	return strings.Split(value, ","), true
}

// Map parses a repeated named argument accumulating `key:value` entries into a map:
// `--point x:1 --point y:2` gives `map[x:1 y:2]`. Entries are split at the first
// colon, then the key and the value are parsed with the given converters - ready-made
// ones are available: [ParseString], [ParseInt], [ParseUint], [ParseFloat].
//
// The metavar describes the entry format shown in help; leave it empty for the
// default `KEY:VALUE`.
//
// The flag is optional: if it isn't present on the command line, the result is an
// empty map. If an environment variable fallback is set with [NamedArg.Env], the
// variable holds comma-separated entries: `POINT=x:1,y:2`.
//
// You can further restrict it with [parseMap.Adjacent].
func (self NamedArg) Map[K comparable, V any](
	metavar Metavar,
	key func(string) (K, error),
	value func(string) (V, error),
) *parseMap[K, V] {
	if metavar == "" {
		metavar = "KEY:VALUE"
	}
	return &parseMap[K, V]{
		named:    self,
		metavar:  metavar,
		adjacent: false,
		key:      key,
		value:    value,
	}
}
