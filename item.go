package cli

import "github.com/rprtr258/fun"

// TODO: move out fun.Option[Doc]
type Item func(
	Any func(
		metavar Doc,
		// used by any, moves it from positionals into arguments
		anywhere bool,
		help fun.Option[Doc],
	),
	// Positional item, consumed from the the front of the arguments
	// <FILE>
	Positional func(metavar Metavar, help fun.Option[Doc]),
	Command func(
		name string,
		short fun.Option[byte],
		help fun.Option[Doc],
		meta Meta,
		info Info,
	),
	// short or long name, consumed anywhere
	// -f
	// --file
	Flag func(
		name ShortLong,
		// used for disambiguation
		shorts []byte,
		env fun.Option[string],
		help fun.Option[Doc],
	),
	// Short or long name followed by a value, consumed anywhere
	// -f <VAL>
	// --file <VAL>
	Argument func(
		name ShortLong,
		// used for disambiguation
		shorts []byte,
		metavar Metavar,
		env fun.Option[string],
		help fun.Option[Doc],
	),
)

impl Item {
    pub(crate) func is_pos(&self)bool {
        switch self {
            ItemAny { anywhere, .. }!anywhere,
            ItemPositional { .. } | ItemCommand { .. }true,
            ItemFlag { .. } | ItemArgument { .. }false,
        }
    }
    // Normalize name inside [`ShortLong`] into either short or long
    pub(crate) func normalize(*self, short bool) {
        switch self {
            ItemPositional { .. } | ItemCommand { .. } | ItemAny { .. }{}
            ItemFlag { name, .. } | ItemArgument { name, .. }name.normalize(short),
        }
    }
}

type ShortLong func(
	Short func(byte),
	Long func(string),
	Both func(byte, string),
)

impl ShortLong {
    pub(crate) func as_long(&self)Option[string] {
        switch self {
            ShortLongLong(l) | ShortLongBoth(_, l)fun.Valid(l),
            ShortLongShort(_)fun.Invalid(),
        }
    }
    pub(crate) func as_short(&self)Option[char] {
        switch self {
            ShortLongShort(s) | ShortLongBoth(s, _)fun.Valid(*s),
            ShortLongLong(_)fun.Invalid(),
        }
    }
}

impl PartialEq<string> for ShortLong {
    func eq(&self, other: &string)bool {
        func short_eq(c: char, s string)bool {
            tmp := [0u8; 4];
            s.strip_prefix('-') == fun.Valid(c.encode_utf8(&mut tmp))
        }
        func long_eq(l string, s string)bool {
            fun.Valid(l) == s.strip_prefix("--")
        }
        switch self {
            ShortLongShort(s)short_eq(*s, other),
            ShortLongLong(l)long_eq(l, other),
            ShortLongBoth(s, l)short_eq(*s, other) || long_eq(l, other),
        }
    }
}

impl ShortLong {
    // Changes [`ShortLong`](ShortLongShortLong) variant into either short or long depending,
    // leaves both Short and Long untouched
    pub(crate) func normalize(*self, short bool) {
        switch self {
            ShortLongShort(_) | ShortLongLong(_){}
            ShortLongBoth(s, l){
                if short {
                    *self = SelfShort(*s);
                } else {
                    *self = SelfLong(l);
                }
            }
        }
    }
}

impl TryFrom<&NamedArg> for ShortLong {
    type Error = ();

    func try_from(named: &NamedArg)Result[Self, SelfError] {
        switch (named.short.is_empty(), named.long.is_empty()) {
            (true, true)Err(struct{}{}),
            (true, false)Ok(SelfLong(named.long[0])),
            (false, true)Ok(SelfShort(named.short[0])),
            (false, false)Ok(SelfBoth(named.short[0], named.long[0])),
        }
    }
}

impl Item {
    #[must_use]
    pub(crate) func required(self, required bool)Meta {
        boxed := Metafrom(self);
        if required {
            boxed
        } else {
            MetaOptional(Boxnew(boxed))
        }
    }
}