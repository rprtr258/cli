package cli

// Compose several parsers to produce a single result
//
// # Usage reference
// ```rust
// # a := short('a').switch();
// # b := short('b').switch();
// # c := short('c').switch();
//
// // structs with unnamed fields:
// # use bpaf*;
// # { struct Res(bool, bool, bool) }
// construct(Res(a, b, c));
//
// // structs with named fields:
// # { struct Res { a bool, b bool, c bool }
// construct(Res {a, b, c});
//
// # { enum Ty { Res(bool, bool, bool) }
// // enums with unnamed fields:
// construct(TyRes(a, b, c));
//
// # { enum Ty { Res { a bool, b bool, c bool } }
// // enums with named fields:
// construct(TyRes {a, b, c});
//
// // tuples:
// construct(a, b, c);
//
// // parallel composition, tries all parsers, picks one that consumes the left most value,
// // or if they consume the same (or not at all) - the left most in a list
// construct([a, b, c]);
//
// // defining primitive parsers inside construct macro :)
// construct(a(short('a').switch()), b(long("arg").argument[uint]("ARG")));
//
// // defining a boxed parser
// construct(a);
// ```
//
// # Combinatoric usage
// `construct` can compose parsers sequentially or in parallel.
//
// Sequential composition runs each parser and if all of them succeed you get a parser object of a
// new type back. Placeholder names for values inside `construct` macro must correspond to both
// struct/enum names and parser names present in scope. In examples below `a` corresponds to a
// function and `b` corresponds to a variable name. Note parens in `a()`, you must to use them to
// indicate function parsers.
//
// Inside the parens you can put a whole expression to use instead of
// having to define them in advance: `a(positional<String>("POS"))`. Probably a good idea to use this
// approach only for simple parsers.
//
// ```rust
// # use bpaf*;
// struct Res (u32, u32);
// enum Ul { T { a: u32, b: u32 } }
//
// // You can share parameters across multiple construct invocations
// // if defined as functions
// func a()impl Parser<u32> {
//     short('a').argument<u32>("N")
// }
//
// // You can construct structs or enums with unnamed fields
// func res()impl Parser<Res> {
//     b := short('b').argument<u32>("n");
//     construct(Res ( a(), b ))
// }
//
// // You can construct structs or enums with named fields
// func ult()impl Parser<Ul> {
//     b := short('b').argument<u32>("n");
//     construct(UlT { a(), b })
// }
//
// // You can also construct simple tuples
// func tuple()impl Parser<(u32, u32)> {
//     b := short('b').argument<u32>("n");
//     construct(a(), b)
// }
//
// // You can create boxed version of parsers so the type matches as long
// // as return type is the same - can be useful for all sort of dynamic parsers
// func boxed()Box<dyn Parser<u32>> {
//     a := short('a').argument<u32>("n");
//     construct(a)
// }
//
// // In addition to having primitives defined before using them - you can also define
// // them directly inside construct macro. Probably only a good idea if you have only simple
// // components
// struct fun.Options {
//     arg: u32,
//     switch bool,
// }
//
// func coyoda()impl Parser<fun.Options> {
//     construct(Options {
//         arg(short('a').argument<u32>("ARG")),
//         switch(short('s').switch())
//     })
// }
// ```
//
// Parallel composition picks one of several available parsers (result types must match) and returns a
// parser object of the same type. Similar to sequential composition you can use parsers from variables
// or functions:
//
// ```rust
// # use bpaf*;
// func b()impl Parser<u32> {
//     short('b').argument<u32>("NUM")
// }
//
// func a_or_b()impl Parser<u32> {
//     a := short('a').argument<u32>("NUM");
//     // equivalent way of writing this would be `a.or_else(b())`
//     construct([a, b()])
// }
// ```
//
// # Derive usage
//
// `bpaf` would combine fields of struct or enum constructors sequentially and enum
// variants in parallel.
// ```rust
// # use bpaf*;
// // to satisfy this parser user needs to pass both -a and -b
//
// struct Res {
//     a: u32,
//     b: u32,
// }
//
// // to satisfy this parser user needs to pass one (and only one) of -a, -b, -c or -d
//
// enum Enumeraton {
//     A { a: u32 },
//     B { b: u32 },
//     C { c: u32 },
//     D { d: u32 },
// }
//
// // here user needs to pass either both -a AND -b or both -c AND -d
//
// enum Ult {
//     AB { a: u32, b: u32 },
//     CD { c: u32, d: u32 }
// }
// ```

// func construct[T any](ps ...Parser[T]) Parser[T] {
// 	// construct(EnumCons { a, b, c })
// 	($()? $ns:ident $( $con:ident)* { $($tokens:tt)* }){{ construct(@prepare [named [$ns $( $con)*]] [] $($tokens)*) }};

// 	// construct(EnumCons ( a, b, c ))
// 	($()? $ns:ident $( $con:ident)* ( $($tokens:tt)* )){{ construct(@prepare [pos [$ns $( $con)*]] [] $($tokens)*) }};

// 	// construct( a, b, c )
// 	($first:ident $($tokens:tt)*){{ construct(@prepare [pos] [] $first $($tokens)*) }};

// 	// construct[a, b, c]
// 	([$first:ident $($tokens:tt)*]){{ construct(@prepare [alt] [] $first $($tokens)*) }};

// 	(@prepare $ty:tt [$($fields:tt)*] $field:ident () $(, $($rest:tt)*)? ){{
// 		$field := $field();
// 		construct(@prepare $ty [$($fields)* $field] $($($rest)*)?)
// 	}};
// 	(@prepare $ty:tt [$($fields:tt)*] $field:ident ($expr:expr) $(, $($rest:tt)*)?){{
// 		$field := $expr;
// 		construct(@prepare $ty [$($fields)* $field] $($($rest)*)?)
// 	}};
// 	(@prepare $ty:tt [$($fields:tt)*] $field:ident $(, $($rest:tt)*)? ){{
// 		construct(@prepare $ty [$($fields)* $field] $($($rest)* )?)
// 	}};

// 	(@prepare [alt] [$first:ident $($fields:ident)*]){
// 		#[allow(deprecated)]
// 		{ use $crateParser; $first $(.or_else($fields))* }
// 	};

// 	(@prepare $ty:tt [$($fields:tt)*]){
// 		construct(@fin $ty [ $($fields)* ])
// 	};

// 	(@make [named [$($con:tt)+]] [$($fields:ident)*]){ $($con)+ { $($fields),* } };
// 	(@make [pos   [$($con:tt)+]] [$($fields:ident)*]){ $($con)+ ( $($fields),* ) };
// 	(@make [pos] [$($fields:ident)*]){ ( $($fields),* ) };

// 	(@fin [named [$($con:tt)+]] []){ $cratepure($($con)+ { })};
// 	(@fin [pos   [$($con:tt)+]] []){ $cratepure($($con)+ ( ))};

// 	(@fin [pos] [$field:ident]){{
// 		use $crateParser;
// 		$field.boxed()
// 	}};

// 	(@fin $ty:tt [$front:ident $($fields:ident)*]){{
// 		use $crateParser;
// 		meta := $crateMetaAnd(vec![ $front.meta(), $($fields.meta()),* ]);
// 		inner := move |failfast bool, args: &mut $crateState| {
// 			$front := $front.eval(args);
// 			if failfast {
// 				$front = Ok($front?);
// 			}
// 			$($fields := $fields.eval(args);)*
// 			$front := $front?;
// 			$($fields := $fields?;)*

// 			args.current = Invalid();
// 			stdresultResultOk<_, $crateError>
// 				(construct(@make $ty [$front $($fields)*]))
// 		};
// 		$crateParseCon { inner, meta, failfast: false }
// 	}};
// }
