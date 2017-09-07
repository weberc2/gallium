GALLIUM
-------

A proposal for a pragmatic, functional, syntactically-Rust-like language that
compiles to Go. Check out the [example][0].

## Disclaimer

I have no experience building a programming language. This will probably never
come to fruition, but I figure the worst I can do is learn a lot about why this
idea is hopelessly naive, which is progress.

## Rationale

Let's start with some values. I believe a small learning curve is important.
Simplicity and familiarity are key to a small learning curve. Tooling is
really important. A great language with poor tooling is mediocre at best. In
particular, good support for the simple cases is underappreciated.

I think functional programming is fundamentally useful, but there isn't a
functional language that strikes me as particularly pragmatic. Haskell is too
pure and abstract for many programmers to learn quickly. OCaml seems to make
pragmatic decisions around purity, but its standard library and project tooling
leave a lot to be desired. Rust nails purity and abstraction and project
tooling (plus a clean, familiar syntax!), but affine typing and manual memory
management impose an unacceptable productivity cost for a large swath of
applications[^a]. Go has a world-class runtime, blazing fast compiler[^b],
great project tooling, no runtime dependencies, but limited immutability, no
sum types, and no type-safe generics[^c].

[^a]: Yes, Rust precludes a lot of concurrency bugs, but it's my opinion that
      the time spent pacifying the compiler is greater than the time spent
      debugging concurrency bugs in a language like Go. YMMV.
[^b]: Partly due to its lack of generics...
[^c]: It's possible much of this will be fixed in Go2, but no one knows when
      Go2 will be released or the degree to which it will even address these
      problems.

[0]: ./example.ga
