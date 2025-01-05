package cli

import (
	"errors"

	"github.com/rprtr258/fun"
)

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

func opt(x fun.Option[bool]) bool { return x.Value && x.Valid }

func FlagBoolShort(c byte) parser[bool] {
	return Map(Optional(regFlagShort(c)), opt)
}

func FlagBoolLong(long string) parser[bool] {
	return Map(Optional(regFlagLong(long)), opt)
}

func FlagBoolShortLong(c byte, long string) parser[bool] {
	return OneOf(
		regFlagShort(c),
		regFlagLong(long),
		Pure(false),
	)
}

func Optional[T any](p parser[T]) parser[fun.Option[T]] {
	return parser[fun.Option[T]]{
		parse: func(s *state) (fun.Option[T], error) {
			scratch := s.clone()
			if val, err := p.parse(&scratch); err != nil {
				return fun.Invalid[T](), nil
			} else {
				*s = scratch
				return fun.Valid(val), nil
			}
		},
	}
}

func Map[T, R any](p parser[T], f func(T) R) parser[R] {
	return parser[R]{
		parse: func(s *state) (R, error) {
			x, err := p.parse(s)
			if err != nil {
				return *new(R), err
			}
			return f(x), nil
		},
	}
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
