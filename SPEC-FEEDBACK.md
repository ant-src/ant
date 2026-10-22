# Spec feedback

Points where the implementation found the specification silent, ambiguous or inconsistent with
itself. Each entry says what the implementation does today. An entry is deleted once the spec
settles it, and the tests follow the spec.

No open entries. The last batch, settled on 2026-10-22: the sufficient condition for visibility
and the wording of the data-race guarantees (chapter 9), `Os.exit` and deferred blocks, the
precedence of a function type's return type, narrowing by assignment, literals in `if`, `while`
and `for` headers, `1.x` as a literal error, `Json.stringify` on a cyclic value, destructuring
an array literal, and the exponent form of floats in `String.from`.
