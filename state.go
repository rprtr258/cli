package cli

import "errors"

type state struct {
	args              []string
	isParsed          []bool
	posArgsStart      int
	posArgsStartFixed bool
}

func newState(args ...string) *state {
	return &state{
		args:              args,
		isParsed:          make([]bool, len(args)),
		posArgsStart:      0,
		posArgsStartFixed: false,
	}
}

func (st *state) popFlagBool(s string) bool {
	end := len(st.args)
	if st.posArgsStartFixed {
		end = st.posArgsStart
	}

	for i := 0; i < end; i++ {
		if st.isParsed[i] || st.args[i] != s {
			continue
		}

		st.isParsed[i] = true
		st.posArgsStart = max(st.posArgsStart, i+1)
		return true
	}

	return false
}

func (st *state) popPositional() (string, bool) {
	if st.posArgsStart >= len(st.args) {
		return "", false
	}

	res := st.args[st.posArgsStart]
	st.isParsed[st.posArgsStart] = true
	st.posArgsStart++
	st.posArgsStartFixed = true
	return res, true
}

func parse[T any](p parser[T], s *state) (T, error) {
	t, err := p.parse(s)
	if err != nil {
		return *new(T), err
	}

	for i := 0; i < len(s.args); i++ {
		if !s.isParsed[i] {
			return *new(T), errors.New("unexpected argument `" + s.args[i] + "`")
		}
	}

	return t, nil
}
