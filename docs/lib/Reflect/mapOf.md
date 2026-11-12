<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Reflect.mapOf

```ant signature
func Reflect.mapOf[K, V](k: Reflect.Type[K], v: Reflect.Type[V]): Reflect.Type[Map[K, V]]
```

Reflect.mapOf returns type(Map\[K, V\]) for a K and a V held as type values.

<!-- end generated -->

## Details

- The result is `==` to the type expression it stands for and hashes alike, so a table keyed by
  `type(Map[string, int])` finds it.

## Basic examples

```ant
Reflect.mapOf(type(string), type(int))                           // Map[string, int]
Reflect.mapOf(type(string), type(int)) == type(Map[string, int]) // true
```

Inside a generic function, `Reflect.mapOf` builds the type that `type(Map[K, V])` cannot name:

```ant
func tableOf[K, V](k: Reflect.Type[K], v: Reflect.Type[V]): Reflect.Type[Map[K, V]] {
  Reflect.mapOf(k, v)
}
tableOf(type(int), type(bool))                         // Map[int, bool]
tableOf(type(int), type(bool)) == type(Map[int, bool]) // true
```

## Scope

The key and the value may be any type a type expression can hold, such as a record, a union, a
container, or a prelude type:

```ant
type Point = { x: int, y: int }
Reflect.mapOf(type(Point), type(int | null))         // Map[Point, int | null]
Reflect.mapOf(type(Array[int]), type(Set[string]))   // Map[Array[int], Set[string]]
Reflect.mapOf(type(byte), type(Json.Value))          // Map[byte, Json.Value]
Reflect.mapOf(type(string), type(func(int): string)) // Map[string, func(int): string]
```

Nest it for maps of maps:

```ant
Reflect.mapOf(type(string), Reflect.mapOf(type(string), type(int))) // Map[string, Map[string, int]]
```

It composes with the other builders:

```ant
Reflect.mapOf(type(string), Reflect.arrayOf(type(int)))  // Map[string, Array[int]]
Reflect.mapOf(Reflect.nullable(type(int)), type(string)) // Map[int | null, string]
Reflect.nullable(Reflect.mapOf(type(string), type(int))) // Map[string, int] | null
```

Call it through `.` on the key type, or as a pipe stage:

```ant
type(string).mapOf(type(int))            // Map[string, int]
type(string) | Reflect.mapOf(type(bool)) // Map[string, bool]
```

## Applications

Decode a JSON object into a `Map[string, V]` for a `V` known only to the caller.
[`Reflect.makeScalar`](makeScalar.md) checks each value against `V`, [`Reflect.makeMap`](makeMap.md)
builds the map at the type `Reflect.mapOf` gives, and [`Reflect.to`](to.md) takes it out at its
static type:

```ant
func decodeTable[V](text: string, v: Reflect.Type[V]): Result[Map[string, V], Error] {
  let json = try Json.parse(text)
  if !(json is Map) { return Result.fail("not an object") }
  let slot = Reflect.erase(v)
  let entries: Array[Entry[Reflect.Value, Reflect.Value]] = []
  for { key, value } in json.entries() {
    if value is Array || value is Map { return Result.fail("${key} does not fit ${v}") }
    let boxed = Reflect.makeScalar(slot, value) ?? return Result.fail("${key} does not fit ${v}")
    entries.push(Entry{ key: Reflect.valueOf(type(string), key), value: boxed })
  }
  let mapType = Reflect.mapOf(type(string), v)
  Result.ok(Reflect.makeMap(Reflect.erase(mapType), entries) | Reflect.to(mapType))
}

println(decodeTable(`{"apples": 3, "pears": 5}`, type(int)))
println(decodeTable(`{"apples": 3, "pears": 5}`, type(float32)))
println(decodeTable(`{"apples": 3.5}`, type(int)))
println(decodeTable(`{"on": true}`, type(bool)))
```

```text
Ok{ value: Map{ "apples": 3, "pears": 5 } }
Ok{ value: Map{ "apples": 3.0, "pears": 5.0 } }
Err{ error: Error{ message: "apples does not fit int" } }
Ok{ value: Map{ "on": true } }
```

Look up a codec by type value. A table keyed by `type(Map[string, int])` finds the entry for a type
built with `Reflect.mapOf`, since the two are `==` and hash alike:

```ant
let codecs: Map[Reflect.ErasedType, string] = Map{}
codecs[Reflect.erase(type(Map[string, int]))] = "counts"

func codecFor[V](v: Reflect.Type[V]): string {
  codecs.get(Reflect.erase(type(string).mapOf(v))) ?? "generic"
}

codecFor(type(int))   // "counts"
codecFor(type(int64)) // "counts"
codecFor(type(bool))  // "generic"
```

`type(int64)` finds the same entry, since `int` is an alias for `int64`.

## Properties and relations

The shape of the result is a map whose key and value are the arguments:

```ant
type Point = { x: int, y: int }
let k = type(string)
let v = type(Point)
let s = Reflect.shape(Reflect.erase(Reflect.mapOf(k, v)))
assert(s is Reflect.Map && s.key == Reflect.erase(k) && s.value == Reflect.erase(v))
```

The builder agrees with the type expression it stands for, aliases included. Swapping the arguments
gives another type; compare the two erased, since their static types differ:

```ant
type Counts = Map[string, int]
assert(Reflect.mapOf(type(string), type(int)) == type(Counts))
let byName = Reflect.erase(Reflect.mapOf(type(string), type(int)))
let byNumber = Reflect.erase(Reflect.mapOf(type(int), type(string)))
assert(byName != byNumber)
```

## Possible issues

A type expression in a generic function may not mention the function's type parameters. The compile
error points to the builders:

```ant
func tableOf[K, V](k: Reflect.Type[K], v: Reflect.Type[V]): Reflect.Type[Map[K, V]] {
  type(Map[K, V])
  // error: type(...) cannot mention K, a type parameter of the enclosing function or type, since nothing is bound to it at run time; take a Reflect.Type[K] parameter and build the type from it with Reflect.arrayOf, Reflect.setOf, Reflect.mapOf, or Reflect.nullable
}
```

An erased type has no static type to build from, so passing one is a compile error. Keep the typed
`Reflect.Type[K]` and erase the result instead:

```ant
let k = Reflect.erase(type(string))
let m = Reflect.mapOf(k, type(int))
// error: cannot use Reflect.ErasedType where Reflect.Type[K] is expected
```

The type only describes the map. The entries given to `Reflect.makeMap` must fit its key and value
types, or the call panics:

```ant
let counts = type(string).mapOf(type(int))
let one = Reflect.valueOf(type(int), 1)
Reflect.makeMap(Reflect.erase(counts), [Entry{ key: one, value: one }])
// panic: Reflect.makeMap: a key of type int is not assignable to the key type string of Map[string, int]
```

## See also

[Reflect.arrayOf](arrayOf.md) · [Reflect.setOf](setOf.md) · [Reflect.nullable](nullable.md) ·
[Reflect.erase](erase.md) · [Reflect.shape](shape.md) · [Reflect.makeMap](makeMap.md) ·
[Reflect.entries](entries.md) · [Reflect.to](to.md) · [Reflect](index.md)
