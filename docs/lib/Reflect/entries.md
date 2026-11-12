<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Reflect.entries

```ant signature
func Reflect.entries(d: Reflect.Value): Array[Entry[Reflect.Value, Reflect.Value]]
```

Reflect.entries returns a map's entries, keys and values boxed at the key and value types, in a new
array read when it is called. It panics if d's type is not a map type.

<!-- end generated -->

## Details

- The entries come in the map's insertion order.
- When the key or value type is a union, each box takes the member its key or value belongs to.
- The array is read when `Reflect.entries` is called: entries added to the map afterwards do not
  appear in it. The boxes refer to the map's own keys and values, not copies.

## Basic examples

Box a map with [Reflect.valueOf](valueOf.md), then read its entries:

```ant
let ages = Reflect.valueOf(type(Map[string, int]), Map{ "ann": 31, "bob": 27 })
Reflect.entries(ages)
// [Entry{ key: Reflect.Value("ann": string), value: Reflect.Value(31: int) }, Entry{ key: Reflect.Value("bob": string), value: Reflect.Value(27: int) }]
```

Unbox a key or a value with [Reflect.to](to.md) or, for a leaf, [Reflect.scalar](scalar.md):

```ant
let ages = Reflect.valueOf(type(Map[string, int]), Map{ "ann": 31, "bob": 27 })
let first = Reflect.entries(ages)[0]
Reflect.to(first.key, type(string)) // "ann"
Reflect.scalar(first.value)         // 31
```

## Scope

An empty map gives an empty array:

```ant
Reflect.entries(Reflect.valueOf(type(Map[int, string]), Map{})) // []
```

The entries keep the map's insertion order, not the order of the keys:

```ant
let d = Reflect.valueOf(type(Map[string, int]), Map{ "b": 2, "a": 1 })
Reflect.entries(d).map({ x.key.scalar() }) // ["b", "a"]
```

When the key or value type is a union, each box takes the member its value belongs to, so
[Reflect.type](type.md) is never the union:

```ant
let mixed: Map[string | int, bool | string] = Map{ "a": true, 2: "two" }
let d = Reflect.valueOf(type(Map[string | int, bool | string]), mixed)
Reflect.entries(d).map({ "${x.key.type()} -> ${x.value.type()}" })
// ["string -> bool", "int -> string"]
```

A nullable value type boxes a stored `null` at `null`:

```ant
let scores: Map[string, int | null] = Map{ "ann": 90, "bob": null }
let d = Reflect.valueOf(type(Map[string, int | null]), scores)
Reflect.entries(d).map({ x.value.type() }) // [int, null]
```

The box may hold a map under a wider type, such as a nullable map; `Reflect.valueOf` resolves the
union, so the box's type is the map type:

```ant
let d = Reflect.valueOf(type(Map[string, int] | null), Map{ "k": 1 })
Reflect.entries(d)
// [Entry{ key: Reflect.Value("k": string), value: Reflect.Value(1: int) }]
```

Keys of a named primitive type keep their tag:

```ant
type UserId = int
let names = Map{ UserId(7): "ann" }
Reflect.entries(Reflect.valueOf(type(Map[UserId, string]), names))
// [Entry{ key: Reflect.Value(UserId(7): UserId), value: Reflect.Value("ann": string) }]
```

A value that is itself a map is boxed at its map type, so `Reflect.entries` reads it in turn:

```ant
let nested = Map{ "colors": Map{ "red": true } }
let d = Reflect.valueOf(type(Map[string, Map[string, bool]]), nested)
let inner = Reflect.entries(d)[0].value
Reflect.entries(inner)
// [Entry{ key: Reflect.Value("red": string), value: Reflect.Value(true: bool) }]
```

## Applications

Render any map whose keys and values are leaves, without knowing their types:

```ant
func render(d: Reflect.Value): string {
  Reflect.entries(d)
    | Array.map({ "${x.key.scalar()}=${x.value.scalar()}" })
    | String.join(", ")
}

println(render(Reflect.valueOf(type(Map[string, int]), Map{ "a": 1, "b": 2 })))
println(render(Reflect.valueOf(type(Map[int, bool]), Map{ 404: false, 200: true })))
```

```text
a=1, b=2
404=false, 200=true
```

Walk a value built from maps, arrays, and leaves, and use the shape to choose between them:

```ant
func show(d: Reflect.Value): string {
  let s = Reflect.shape(d.type())
  if s is Reflect.Map {
    let body = Reflect.entries(d)
      | Array.map({ "${show(x.key)}: ${show(x.value)}" })
      | String.join(", ")
    return "{${body}}"
  }
  if s is Reflect.Array {
    let body = Reflect.items(d) | Array.map(show) | String.join(", ")
    return "[${body}]"
  }
  String.from(d.scalar())
}

let inventory = Map{ "fruit": ["apple", "pear"], "nuts": [] }
println(show(Reflect.valueOf(type(Map[string, Array[string]]), inventory)))
```

```text
{fruit: [apple, pear], nuts: []}
```

Loop over the entries and destructure each one into its key and value:

```ant
let d = Reflect.valueOf(type(Map[string, int]), Map{ "a": 1, "b": 2 })
for { key, value } in Reflect.entries(d) {
  println("${key.scalar()}=${value.scalar()}")
}
```

```text
a=1
b=2
```

## Properties and relations

The entries have as many elements as the map, and the keys come in the order of
[Map.keys](../Map/keys.md):

```ant
let m = Map{ "x": 1, "y": 2, "z": 3 }
let es = Reflect.entries(Reflect.valueOf(type(Map[string, int]), m))
assert(es.length() == m.size())
assert(es.map({ x.key.to(type(string)) }) == m.keys())
```

[Reflect.makeMap](makeMap.md) builds an equal map back from the entries and the map's type:

```ant
let d = Reflect.valueOf(type(Map[string, int]), Map{ "a": 1, "b": 2 })
let copy = Reflect.makeMap(d.type(), Reflect.entries(d))
assert(copy == d)
assert(!Reflect.same(copy, d))
```

## Possible issues

`Reflect.entries` reads only maps. An array or a set panics; read those with
[Reflect.items](items.md):

```ant
let xs = Reflect.valueOf(type(Array[int]), [1, 2])
Reflect.entries(xs)
// panic: Reflect.entries needs a value of a map type, got one of type Array[int]
```

A nullable map type that holds `null` boxes it at `null`, so `Reflect.entries` panics. Check
`Reflect.shape(d.type()) is Reflect.Map` before reading the entries:

```ant
let m: Map[string, int] | null = null
let d = Reflect.valueOf(type(Map[string, int] | null), m)
Reflect.entries(d)
// panic: Reflect.entries needs a value of a map type, got one of type null
```

The array is a snapshot of the entries at the call. An entry added later needs a new call:

```ant
let m = Map{ "a": 1 }
let d = Reflect.valueOf(type(Map[string, int]), m)
let before = Reflect.entries(d)
m["b"] = 2
before.length()             // 1
Reflect.entries(d).length() // 2
```

The boxes are not copies. A record taken out of an entry is the record in the map, so changing it
changes the map:

```ant
type Point = { x: int, y: int }
let points = Map{ "a": Point{ x: 1, y: 2 } }
let d = Reflect.valueOf(type(Map[string, Point]), points)
let p = Reflect.entries(d)[0].value.to(type(Point))
p.x = 10
points // Map{ "a": Point{ x: 10, y: 2 } }
```

## See also

[Reflect.items](items.md) · [Reflect.field](field.md) · [Reflect.makeMap](makeMap.md) ·
[Reflect.valueOf](valueOf.md) · [Reflect.to](to.md) · [Reflect.scalar](scalar.md) ·
[Reflect.shape](shape.md) · [Map.entries](../Map/entries.md)
