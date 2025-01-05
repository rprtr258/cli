package cli

import "errors"

type parser[T any] struct {
	parse func(*state) (T, error)
}

func Pure[T any](x T) parser[T] {
	return parser[T]{
		parse: func(*state) (T, error) {
			return x, nil
		},
	}
}

func Fail[T any](msg string) parser[T] {
	return parser[T]{
		parse: func(*state) (T, error) {
			return *new(T), errors.New(msg)
		},
	}
}

func Positional[T any](name string, parse func(string) (T, error)) parser[T] {
	return parser[T]{
		parse: func(s *state) (T, error) {
			if s.posArgsStart >= len(s.args) {
				return *new(T), errors.New("expected `" + name + "`, pass `--help` for usage information")
			}

			arg := s.args[s.posArgsStart]
			s.posArgsStart++
			s.posArgsStartFixed = true
			return parse(arg)
		},
	}
}
