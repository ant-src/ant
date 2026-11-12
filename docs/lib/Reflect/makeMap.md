<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Reflect.makeMap

```ant signature
func Reflect.makeMap(t: Reflect.ErasedType, entries: Array[Entry[Reflect.Value, Reflect.Value]]): Reflect.Value
```

Reflect.makeMap returns a new map of the map type t holding entries, a later entry with an == key
replacing an earlier one. It panics if t is not a map type, or a key's or value's type is not
assignable to its slot.

<!-- end generated -->

## Details

- `Reflect.makeMap` returns a new map and leaves `entries` unchanged.
- A later entry with an `==` key replaces only the value: the map keeps the first key and its
  position.
- The result is a `Reflect.Value`. Unbox it with `Reflect.to` at the static map type.

## Basic examples

Build a `Map[string, int]` from boxed keys and values:

```ant
let t = Reflect.erase(type(Map[string, int]))
let ann = Entry{ key: Reflect.valueOf(type(string), "ann"), value: Reflect.valueOf(type(int), 31) }
Reflect.makeMap(t, [ann]) // Reflect.Value(Map{ "ann": 31 }: Map[string, int])
```

Unbox the result with `Reflect.to`:

```ant
let t = Reflect.erase(type(Map[string, int]))
let ann = Entry{ key: Reflect.valueOf(type(string), "ann"), value: Reflect.valueOf(type(int), 31) }
Reflect.to(Reflect.makeMap(t, [ann]), type(Map[string, int])) // Map{ "ann": 31 }
```

## Scope

Small helpers keep the boxing of keys and values short:

```ant
func str(s: string): Reflect.Value { Reflect.valueOf(type(string), s) }
func num(n: int): Reflect.Value { Reflect.valueOf(type(int), n) }
let t = Reflect.erase(type(Map[string, int]))
Reflect.makeMap(t, [Entry{ key: str("a"), value: num(1) }, Entry{ key: str("b"), value: num(2) }])
// Reflect.Value(Map{ "a": 1, "b": 2 }: Map[string, int])
```

An empty array gives an empty map of the type:

```ant
Reflect.makeMap(Reflect.erase(type(Map[string, bool])), []) // Reflect.Value(Map{}: Map[string, bool])
```

A repeated key keeps its first position and takes the value of its last entry:

```ant
func str(s: string): Reflect.Value { Reflect.valueOf(type(string), s) }
func num(n: int): Reflect.Value { Reflect.valueOf(type(int), n) }
let t = Reflect.erase(type(Map[string, int]))
let votes = [
  Entry{ key: str("red"), value: num(1) },
  Entry{ key: str("blue"), value: num(2) },
  Entry{ key: str("red"), value: num(3) },
]
Reflect.makeMap(t, votes) // Reflect.Value(Map{ "red": 3, "blue": 2 }: Map[string, int])
```

A value fits a union slot when its type is one of the members, such as `null` in `int | null`:

```ant
let t = Reflect.erase(type(Map[string, int | null]))
let limit = Entry{ key: Reflect.valueOf(type(string), "limit"), value: Reflect.valueOf(type(null), null) }
Reflect.makeMap(t, [limit]) // Reflect.Value(Map{ "limit": null }: Map[string, int | null])
```

Keys may be records or arrays. They compare with `==`, so two equal arrays are one key:

```ant
type Cell = { row: int, col: int }
let t = Reflect.erase(type(Map[Cell, string]))
let cell = Reflect.valueOf(type(Cell), Cell{ row: 0, col: 1 })
Reflect.makeMap(t, [Entry{ key: cell, value: Reflect.valueOf(type(string), "x") }])
// Reflect.Value(Map{ Cell{ row: 0, col: 1 }: "x" }: Map[Cell, string])
let byList = Reflect.erase(type(Map[Array[int], string]))
let a = Entry{ key: Reflect.valueOf(type(Array[int]), [1, 2]), value: Reflect.valueOf(type(string), "a") }
let b = Entry{ key: Reflect.valueOf(type(Array[int]), [1, 2]), value: Reflect.valueOf(type(string), "b") }
Reflect.makeMap(byList, [a, b])
// Reflect.Value(Map{ [1, 2]: "b" }: Map[Array[int], string])
```

Make the map type from type values held at run time with `Reflect.mapOf`:

```ant
let key = type(string)
let value = type(bool)
let t = Reflect.erase(Reflect.mapOf(key, value))
let on = Entry{ key: Reflect.valueOf(key, "debug"), value: Reflect.valueOf(value, true) }
Reflect.makeMap(t, [on]) // Reflect.Value(Map{ "debug": true }: Map[string, bool])
```

Rebuild a map from the entries `Reflect.entries` reads:

```ant
let d = Reflect.valueOf(type(Map[string, int]), Map{ x: 1, y: 2 })
Reflect.makeMap(Reflect.type(d), Reflect.entries(d)) // Reflect.Value(Map{ "x": 1, "y": 2 }: Map[string, int])
```

Name the arguments:

```ant
Reflect.makeMap(t: Reflect.erase(type(Map[int, int])), entries: []) // Reflect.Value(Map{}: Map[int, int])
```

## Applications

Decode a JSON object into a map whose type is known only at run time. `Reflect.shape` gives the key
and value types, and `Reflect.makeScalar` boxes each leaf at its slot, converting numbers:

```ant
// decodeObject builds a map of the map type t from a JSON object whose values are scalars.
// It returns null when t is not a map type or a key or value does not fit its slot.
func decodeObject(t: Reflect.ErasedType, obj: Map[string, Json.Value]): Reflect.Value | null {
  let shape = Reflect.shape(t)
  if !(shape is Reflect.Map) { return null }
  let entries: Array[Entry[Reflect.Value, Reflect.Value]] = []
  for e in obj {
    let v = e.value
    if v is Array || v is Map { return null }
    let key = Reflect.makeScalar(shape.key, e.key)
    let value = Reflect.makeScalar(shape.value, v)
    if key == null || value == null { return null }
    entries.push(Entry{ key, value })
  }
  Reflect.makeMap(t, entries)
}

let obj = Json.parse(`{"ann": 31, "bob": 27}`) | Result.unwrap()
if obj is Map {
  let ages = decodeObject(Reflect.erase(type(Map[string, int])), obj)
  if ages != null { println(Reflect.to(ages, type(Map[string, int]))) }
  println(decodeObject(Reflect.erase(type(Map[string, bool])), obj))
}
```

```text
Map{ "ann": 31, "bob": 27 }
null
```

Drop the `null` values from any boxed map, keeping its type:

```ant
func dropNulls(d: Reflect.Value): Reflect.Value {
  let kept = Reflect.entries(d).filter({ !(Reflect.shape(Reflect.type(x.value)) is Reflect.Null) })
  Reflect.makeMap(Reflect.type(d), kept)
}

let settings: Map[string, int | null] = Map{ limit: 10, retries: null, timeout: 30 }
println(dropNulls(Reflect.valueOf(type(Map[string, int | null]), settings)))
```

```text
Reflect.Value(Map{ "limit": 10, "timeout": 30 }: Map[string, int | null])
```

## Properties and relations

Rebuilding from `Reflect.entries` gives a map `==` to the original, but a new one:

```ant
let d = Reflect.valueOf(type(Map[string, int]), Map{ b: 1, a: 2 })
let copy = Reflect.makeMap(Reflect.type(d), Reflect.entries(d))
assert(copy == d)
assert(!Reflect.same(copy, d))
assert(Reflect.type(copy) == Reflect.type(d))
```

`Reflect.makeMap` builds what a `Map` literal builds:

```ant
let t = Reflect.erase(type(Map[string, int]))
let k = Reflect.valueOf(type(string), "a")
let built = Reflect.makeMap(t, [Entry{ key: k, value: Reflect.valueOf(type(int), 1) }])
assert(Reflect.to(built, type(Map[string, int])) == Map{ a: 1 })
```

## Possible issues

`t` must be a map type:

```ant
Reflect.makeMap(Reflect.erase(type(Array[int])), []) // panic: Reflect.makeMap needs a map type, got Array[int]
```

A nullable slot such as `Map[string, int] | null` is a union, not a map type:

```ant
Reflect.makeMap(Reflect.erase(type(Map[string, int] | null)), []) // panic: Reflect.makeMap needs a map type, got Map[string, int] | null
```

Take its map member from `Reflect.shape(t).members` before the call.

A key or value whose type is not assignable to its slot panics, with the slot in the message:

```ant
let t = Reflect.erase(type(Map[string, int]))
let one = Reflect.valueOf(type(int), 1)
Reflect.makeMap(t, [Entry{ key: one, value: one }])
// panic: Reflect.makeMap: a key of type int is not assignable to the key type string of Map[string, int]
```

A box keeps its own number type, so an `int32` value does not fit an `int` slot. Box the number with
`Reflect.makeScalar` at the slot's type, which converts it:

```ant
let t = Reflect.erase(type(Map[string, int]))
let k = Reflect.valueOf(type(string), "a")
let small = Reflect.valueOf(type(int32), int32(1))
Reflect.makeMap(t, [Entry{ key: k, value: small }])
// panic: Reflect.makeMap: a value of type int32 is not assignable to the value type int of Map[string, int]
Reflect.makeScalar(Reflect.erase(type(int)), int32(1)) // Reflect.Value(1: int)
```

Test a type with `Reflect.shape` and a part with `Reflect.assignable` before the call to report bad
input instead of panicking:

```ant
let t = Reflect.erase(type(Map[string, int]))
let shape = Reflect.shape(t)
let v = Reflect.valueOf(type(bool), true)
shape is Reflect.Map && Reflect.assignable(Reflect.type(v), shape.value) // false
```

## See also

[Reflect.entries](entries.md), [Reflect.makeArray](makeArray.md), [Reflect.makeSet](makeSet.md),
[Reflect.makeScalar](makeScalar.md), [Reflect.mapOf](mapOf.md), [Reflect.to](to.md),
[Map.fromEntries](../Map/fromEntries.md)
