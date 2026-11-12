<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Reflect.setOf

```ant signature
func Reflect.setOf[T](t: Reflect.Type[T]): Reflect.Type[Set[T]]
```

Reflect.setOf returns type(Set\[T\]) for a T held as a type value.

<!-- end generated -->

## Details

- The result is `==` to the type expression it stands for and hashes alike, so a table keyed by
  `type(Set[int])` finds it.

## Basic examples

```ant
Reflect.setOf(type(int))                   // Set[int]
Reflect.setOf(type(int)) == type(Set[int]) // true
```

Inside a generic function, `Reflect.setOf` builds the type that `type(Set[T])` cannot name:

```ant
func uniqueOf[T](t: Reflect.Type[T]): Reflect.Type[Set[T]] {
  Reflect.setOf(t)
}
uniqueOf(type(string))                      // Set[string]
uniqueOf(type(string)) == type(Set[string]) // true
```

## Scope

The member type may be any type a type expression can hold, such as a record, a union, an array, a
prelude alias, or a function type:

```ant
type Point = { x: int, y: int }
Reflect.setOf(type(Point))          // Set[Point]
Reflect.setOf(type(int | null))     // Set[int | null]
Reflect.setOf(type(Array[string]))  // Set[Array[string]]
Reflect.setOf(type(Json.Value))     // Set[Json.Value]
Reflect.setOf(type(func(int): int)) // Set[func(int): int]
```

Nest it for sets of sets:

```ant
Reflect.setOf(Reflect.setOf(type(string))) // Set[Set[string]]
```

It composes with the other builders. The order decides which part may be `null`:

```ant
Reflect.setOf(Reflect.nullable(type(int)))            // Set[int | null]
Reflect.nullable(Reflect.setOf(type(int)))            // Set[int] | null
Reflect.mapOf(type(string), Reflect.setOf(type(int))) // Map[string, Set[int]]
```

## Applications

Build a `Set[T]` from boxed values for a `T` known only to the caller. `Reflect.makeSet` collapses
the duplicates and checks each member, and `Reflect.to` takes the set back out at its static type:

```ant
func distinct[T](t: Reflect.Type[T], items: Array[Reflect.Value]): Set[T] {
  let setType = t.setOf()
  Reflect.makeSet(Reflect.erase(setType), items) | Reflect.to(setType)
}

let boxed = ["go", "ant", "go"].map({ Reflect.valueOf(type(string), x) })
println(distinct(type(string), boxed))
println(distinct(type(string | null), boxed))
```

```text
Set.of(["go", "ant"])
Set.of(["go", "ant"])
```

Find the fields of a record that hold a set of a given member type, by comparing each field's type
with the one `Reflect.setOf` builds:

```ant
func setFields[T](record: Reflect.ErasedType, member: Reflect.Type[T]): Array[string] {
  let s = Reflect.shape(record)
  if !(s is Reflect.Record) {
    return []
  }
  let wanted = Reflect.erase(member.setOf())
  s.fields.filter({ x.type == wanted }).map({ x.name })
}

type Post = { title: string, tags: Set[string], authors: Set[string], likes: Set[int] }
println(setFields(Reflect.erase(type(Post)), type(string)))
println(setFields(Reflect.erase(type(Post)), type(int)))
```

```text
["tags", "authors"]
["likes"]
```

## Properties and relations

The shape of the result is a set whose member is the argument:

```ant
type Point = { x: int, y: int }
let t = type(Point)
let s = Reflect.shape(Reflect.erase(t.setOf()))
assert(s is Reflect.Set && s.member == Reflect.erase(t))
```

The builders agree with the type expressions they stand for:

```ant
assert(Reflect.setOf(type(string)) == type(Set[string]))
assert(Reflect.setOf(type(int | null)) == type(Set[int | null]))
assert(Reflect.setOf(type(int)).nullable() == type(Set[int] | null))
```

A set type is never the array type of the same member. Erased, the two compare as type values:

```ant
let member = type(int)
assert(Reflect.erase(member.setOf()) != Reflect.erase(member.arrayOf()))
```

## Possible issues

A type expression in a generic function may not mention the function's type parameter. The compile
error points to the builders:

```ant
func uniqueOf[T](t: Reflect.Type[T]): Reflect.Type[Set[T]] {
  type(Set[T])
  // error: type(...) cannot mention T, a type parameter of the enclosing function or type, since nothing is bound to it at run time; take a Reflect.Type[T] parameter and build the type from it with Reflect.arrayOf, Reflect.setOf, Reflect.mapOf, or Reflect.nullable
}
```

An erased type has no static type to build from, so passing one is a compile error. Keep the typed
`Reflect.Type[T]` and erase the result instead:

```ant
let e = Reflect.erase(type(int))
let ints = Reflect.setOf(e) // error: cannot use Reflect.ErasedType where Reflect.Type[T] is expected
```

Two typed type values compare only when one static type is assignable to the other, so comparing a
set type with an array type is a compile error. Erase both to compare them at run time, as above:

```ant
let same = type(int).setOf() == type(int).arrayOf()
// error: Reflect.Type[Set[int]] and Reflect.Type[Array[int]] cannot be compared with ==: neither is assignable to the other
```

The type only describes the set. The members given to `Reflect.makeSet` must fit its member type, or
the call panics:

```ant
let words = type(string).setOf()
let boxed = [Reflect.valueOf(type(int), 1)]
Reflect.makeSet(Reflect.erase(words), boxed)
// panic: Reflect.makeSet: a member of type int is not assignable to the member type string of Set[string]
```

## See also

[Reflect.arrayOf](arrayOf.md) · [Reflect.mapOf](mapOf.md) · [Reflect.nullable](nullable.md) ·
[Reflect.erase](erase.md) · [Reflect.shape](shape.md) · [Reflect.makeSet](makeSet.md) ·
[Reflect.to](to.md) · [Set.of](../Set/of.md) · [Reflect](index.md)
