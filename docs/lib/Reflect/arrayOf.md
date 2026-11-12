<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Reflect.arrayOf

```ant signature
func Reflect.arrayOf[T](t: Reflect.Type[T]): Reflect.Type[Array[T]]
```

Reflect.arrayOf returns type(Array\[T\]) for a T held as a type value, which a generic function
cannot write itself, since type(...) may not mention a type parameter.

<!-- end generated -->

## Details

- The result is `==` to the type expression it stands for and hashes alike, so a table keyed by
  `type(Array[int])` finds it.

## Basic examples

```ant
Reflect.arrayOf(type(int))                     // Array[int]
Reflect.arrayOf(type(int)) == type(Array[int]) // true
```

Inside a generic function, `Reflect.arrayOf` builds the type that `type(Array[T])` cannot name:

```ant
func listOf[T](t: Reflect.Type[T]): Reflect.Type[Array[T]] {
  Reflect.arrayOf(t)
}
listOf(type(string))                        // Array[string]
listOf(type(string)) == type(Array[string]) // true
```

## Scope

The element type may be any type a type expression can hold, such as a record, a union, a prelude
type, or a function type:

```ant
type Point = { x: int, y: int }
Reflect.arrayOf(type(Point))             // Array[Point]
Reflect.arrayOf(type(int | null))        // Array[int | null]
Reflect.arrayOf(type(Bytes))             // Array[Bytes]
Reflect.arrayOf(type(Json.Value))        // Array[Json.Value]
Reflect.arrayOf(type(func(int): string)) // Array[func(int): string]
```

Nest it for arrays of arrays:

```ant
Reflect.arrayOf(Reflect.arrayOf(type(string))) // Array[Array[string]]
```

It composes with the other builders. The order decides which part may be `null`:

```ant
Reflect.arrayOf(Reflect.nullable(type(int)))            // Array[int | null]
Reflect.nullable(Reflect.arrayOf(type(int)))            // Array[int] | null
Reflect.mapOf(type(string), Reflect.arrayOf(type(int))) // Map[string, Array[int]]
```

Call it through `.` on a type value, or as a pipe stage:

```ant
type(bool).arrayOf()           // Array[bool]
type(bool) | Reflect.arrayOf() // Array[bool]
```

The element type of a user's generic type comes from the caller, who writes its type arguments:

```ant
type Box[T] = { value: T }
func boxes[T](t: Reflect.Type[Box[T]]): Reflect.Type[Array[Box[T]]] {
  t.arrayOf()
}
boxes(type(Box[int])) // Array[Box[int]]
```

## Applications

Build an `Array[T]` from boxed values for a `T` known only to the caller, as a typed decoder does
for a list. `Reflect.makeArray` checks each item against the element type, and `Reflect.to` takes
the array back out at its static type:

```ant
func collect[T](t: Reflect.Type[T], items: Array[Reflect.Value]): Array[T] {
  let arrayType = t.arrayOf()
  Reflect.makeArray(Reflect.erase(arrayType), items) | Reflect.to(arrayType)
}

let boxed = [1, 2, 3].map({ Reflect.valueOf(type(int), x) })
println(collect(type(int), boxed))
println(collect(type(int | null), boxed))
```

```text
[1, 2, 3]
[1, 2, 3]
```

Look up a codec by type value. A table keyed by `type(Array[int])` finds the entry for a type built
with `Reflect.arrayOf`, since the two are `==` and hash alike:

```ant
let names: Map[Reflect.ErasedType, string] = Map{}
names[Reflect.erase(type(Array[int]))] = "int list"

func nameOf[T](t: Reflect.Type[T]): string {
  names.get(Reflect.erase(t.arrayOf())) ?? "unknown"
}

nameOf(type(int))    // "int list"
nameOf(type(string)) // "unknown"
```

## Properties and relations

The shape of the result is an array whose element is the argument:

```ant
type Point = { x: int, y: int }
let t = type(Point)
let s = Reflect.shape(Reflect.erase(t.arrayOf()))
assert(s is Reflect.Array && s.element == Reflect.erase(t))
```

The builders agree with the type expressions they stand for:

```ant
assert(Reflect.arrayOf(type(string)) == type(Array[string]))
assert(Reflect.arrayOf(type(int | null)) == type(Array[int | null]))
assert(Reflect.arrayOf(type(int)).nullable() == type(Array[int] | null))
```

## Possible issues

A type expression in a generic function may not mention the function's type parameter. The compile
error points to the builders:

```ant
func listOf[T](t: Reflect.Type[T]): Reflect.Type[Array[T]] {
  type(Array[T])
  // error: type(...) cannot mention T, a type parameter of the enclosing function or type, since nothing is bound to it at run time; take a Reflect.Type[T] parameter and build the type from it with Reflect.arrayOf, Reflect.setOf, Reflect.mapOf, or Reflect.nullable
}
```

An erased type has no static type to build from, so passing one is a compile error. Keep the typed
`Reflect.Type[T]` and erase the result instead:

```ant
let e = Reflect.erase(type(int))
let ints = Reflect.arrayOf(e) // error: cannot use Reflect.ErasedType where Reflect.Type[T] is expected
```

The type only describes the array. The items given to `Reflect.makeArray` must fit its element type,
or the call panics:

```ant
let words = type(string).arrayOf()
let boxed = [Reflect.valueOf(type(int), 1)]
Reflect.makeArray(Reflect.erase(words), boxed)
// panic: Reflect.makeArray: an item of type int is not assignable to the element type string of Array[string]
```

## See also

[Reflect.setOf](setOf.md) · [Reflect.mapOf](mapOf.md) · [Reflect.nullable](nullable.md) ·
[Reflect.erase](erase.md) · [Reflect.shape](shape.md) · [Reflect.makeArray](makeArray.md) ·
[Reflect.to](to.md) · [Reflect](index.md)
