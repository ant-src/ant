<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Reflect.nullable

```ant signature
func Reflect.nullable[T](t: Reflect.Type[T]): Reflect.Type[T | null]
```

Reflect.nullable returns type(T \| null) for a T held as a type value. Unions flatten, so a T that
already includes null gives the same type.

<!-- end generated -->

## Details

- Adding `null` to a type that already admits it changes nothing, so `t.nullable()` is a safe way to
  make any `T` optional.

## Basic examples

```ant
Reflect.nullable(type(int))                     // int | null
Reflect.nullable(type(int)) == type(int | null) // true
```

Inside a generic function, `Reflect.nullable` builds the type that `type(T | null)` cannot name:

```ant
func optional[T](t: Reflect.Type[T]): Reflect.Type[T | null] {
  Reflect.nullable(t)
}
optional(type(string))                        // string | null
optional(type(string)) == type(string | null) // true
```

## Scope

The argument may be any type a type expression can hold, such as a record, a container, a prelude
type, or a function type:

```ant
type Point = { x: int, y: int }
Reflect.nullable(type(Point))              // Point | null
Reflect.nullable(type(Array[int]))         // Array[int] | null
Reflect.nullable(type(Result[int, Error])) // Result[int, Error] | null
Reflect.nullable(type(func(int): string))  // (func(int): string) | null
```

A union gains `null` as one more member:

```ant
Reflect.nullable(type(int | string)) // int | string | null
```

A type that already includes `null` stays the same type, whether it is written out, named by an
alias, or `null` itself:

```ant
type MaybeInt = int | null
Reflect.nullable(type(int | null))                     // int | null
Reflect.nullable(type(null))                           // null
Reflect.nullable(type(MaybeInt)) == type(MaybeInt)     // true
Reflect.nullable(type(Json.Value)) == type(Json.Value) // true
```

Call it through `.` on a type value, or as a pipe stage:

```ant
type(bool).nullable()           // bool | null
type(bool) | Reflect.nullable() // bool | null
```

It composes with the other builders. The order decides which part may be `null`:

```ant
Reflect.arrayOf(Reflect.nullable(type(int)))                 // Array[int | null]
Reflect.nullable(Reflect.arrayOf(type(int)))                 // Array[int] | null
Reflect.mapOf(type(string), Reflect.nullable(type(float64))) // Map[string, float64 | null]
```

## Applications

Take a boxed value out at an optional type, for a `T` known only to the caller, as a decoder does
for a field that may be missing. `Reflect.to` accepts both a `T` and a `null`:

```ant
func unbox[T](t: Reflect.Type[T], d: Reflect.Value): T | null {
  Reflect.to(d, t.nullable())
}

println(unbox(type(int), Reflect.valueOf(type(int), 42)))
println(unbox(type(int), Reflect.valueOf(type(null), null)))
```

```text
42
null
```

Ask whether a type already admits `null`: adding `null` leaves such a type unchanged. The static
types `Reflect.Type[T | null]` and `Reflect.Type[T]` differ, so compare the two erased:

```ant
func acceptsNull[T](t: Reflect.Type[T]): bool {
  Reflect.erase(t.nullable()) == Reflect.erase(t)
}

acceptsNull(type(int))           // false
acceptsNull(type(string | null)) // true
acceptsNull(type(Json.Value))    // true
```

Look up a codec by type value. A table keyed by `type(int | null)` finds the entry for a type built
with `Reflect.nullable`, since the two are `==` and hash alike:

```ant
let names: Map[Reflect.ErasedType, string] = Map{}
names[Reflect.erase(type(int | null))] = "optional int"

func nameOf[T](t: Reflect.Type[T]): string {
  names.get(Reflect.erase(t.nullable())) ?? "unknown"
}

nameOf(type(int))        // "optional int"
nameOf(type(int | null)) // "optional int"
nameOf(type(string))     // "unknown"
```

## Properties and relations

The shape of the result is a union of the argument and `null`:

```ant
let t = type(int).nullable()
let s = Reflect.shape(Reflect.erase(t))
assert(s is Reflect.Union && s.members == [Reflect.erase(type(int)), Reflect.erase(type(null))])
```

The argument is assignable to the result, and so is `null`:

```ant
type Point = { x: int, y: int }
let t = Reflect.erase(type(Point))
let n = Reflect.erase(type(Point).nullable())
assert(Reflect.assignable(t, n))
assert(Reflect.assignable(Reflect.erase(type(null)), n))
```

The call is idempotent:

```ant
let once = type(string).nullable()
assert(once.nullable() == once)
```

## Possible issues

A type expression in a generic function may not mention the function's type parameter. The compile
error points to the builders:

```ant
func optional[T](t: Reflect.Type[T]): Reflect.Type[T | null] {
  type(T | null)
  // error: type(...) cannot mention T, a type parameter of the enclosing function or type, since nothing is bound to it at run time; take a Reflect.Type[T] parameter and build the type from it with Reflect.arrayOf, Reflect.setOf, Reflect.mapOf, or Reflect.nullable
}
```

An erased type has no static type to build from, so passing one is a compile error. Keep the typed
`Reflect.Type[T]` and erase the result instead:

```ant
let e = Reflect.erase(type(int))
let maybe = Reflect.nullable(e) // error: cannot use Reflect.ErasedType where Reflect.Type[T] is expected
```

## See also

[Reflect.arrayOf](arrayOf.md) · [Reflect.setOf](setOf.md) · [Reflect.mapOf](mapOf.md) ·
[Reflect.erase](erase.md) · [Reflect.shape](shape.md) · [Reflect.assignable](assignable.md) ·
[Reflect.to](to.md) · [Reflect](index.md)
