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
			arg, ok := s.popPositional()
			if !ok {
				return *new(T), errors.New("expected `" + name + "`, pass `--help` for usage information")
			}

			return parse(arg)
		},
	}
}

func FlagBoolShort(c byte) parser[bool] {
	return parser[bool]{
		parse: func(s *state) (bool, error) {
			return s.popFlagBool("-" + string(c)), nil
		},
	}
}

func FlagBoolLong(long string) parser[bool] {
	return parser[bool]{
		parse: func(s *state) (bool, error) {
			return s.popFlagBool("--" + long), nil
		},
	}
}

func OneOf[T any](parsers ...parser[T]) parser[T] {
	if len(parsers) == 0 {
		panic("at least one parser is required")
	}

	return parser[T]{
		parse: func(s *state) (T, error) {
			errs := []error{}
			for _, p := range parsers {
				scratch := s.clone()
				if x, err := p.parse(&scratch); err == nil {
					*s = scratch
					return x, nil
				} else {
					errs = append(errs, err)
				}
			}
			return *new(T), errors.Join(errs...)
		},
	}
}

func regFlagShort(c byte) parser[bool] {
	return parser[bool]{
		parse: func(s *state) (bool, error) {
			if ok := s.popFlagBool("-" + string(c)); !ok {
				return false, errors.New("expected `-" + string(c) + "=true`, got `-" + string(c) + "`. Pass `--help` for usage information")
			}
			return true, nil
		},
	}
}

func regFlagLong(long string) parser[bool] {
	return parser[bool]{
		parse: func(s *state) (bool, error) {
			if ok := s.popFlagBool("--" + long); !ok {
				return false, errors.New("expected `--" + long + "=true`, got `--" + long + "`. Pass `--help` for usage information")
			}
			return true, nil
		},
	}
}

func FlagBoolShortLong(c byte, long string) parser[bool] {
	return OneOf(
		regFlagShort(c),
		regFlagLong(long),
		Pure(false),
	)
}

func FlagShort[T any](c byte, name string, parse func(string) (T, error)) parser[T] {
	return parser[T]{
		parse: func(s *state) (T, error) {
			val, ok := s.popFlag("-" + string(c))
			if !ok {
				return *new(T), errors.New("expected `-" + string(c) + "=" + name + "`, got `-" + string(c) + "`. Pass `--help` for usage information")
			}

			return parse(val)
		},
	}
}

func FlagLong[T any](long string, name string, parse func(string) (T, error)) parser[T] {
	return parser[T]{
		parse: func(s *state) (T, error) {
			val, ok := s.popFlag("--" + long)
			if !ok {
				return *new(T), errors.New("expected `-" + long + "=" + name + "`, got `--" + long + "`. Pass `--help` for usage information")
			}

			return parse(val)
		},
	}
}

func FlagShortLong[T any](c byte, long string, name string, parse func(string) (T, error)) parser[T] {
	return OneOf(
		FlagShort(c, name, parse),
		FlagLong(long, name, parse),
	)
}
