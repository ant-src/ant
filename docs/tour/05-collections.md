<!-- status: reviewed -->

# Arrays, maps, and sets

Ant has three collections: an array holds values in order, a map looks values up by key, and a set
holds each value once. This program builds one of each, reads it, changes it, and loops over it:

```ant
let scores = [72, 95, 88]
Array.push(scores, 61)
scores[0] = 75
println(scores[1])
println(scores)
println("${Array.length(scores)} scores")

let ages = Map{ ada: 36, linus: 54 }
ages["grace"] = 85
println(ages["ada"])
println(ages["alan"] ?? 0)
for { key, value } in ages {
  println("${key} is ${value}")
}

let tags = Set.of(["red", "blue", "red"])
Set.add(tags, "green")
println(Set.has(tags, "red"))
for tag in tags {
  println(tag)
}
```

```text
95
[75, 95, 88, 61]
4 scores
36
0
ada is 36
linus is 54
grace is 85
true
red
blue
green
```

Square brackets build an array, and `scores[1]` reads the element at index 1, counting from 0.
`scores[0] = 75` replaces an element, and `Array.push` appends one. The array changes in place:
`let` stops you from pointing `scores` at another array, not from changing the one it holds.

`Map{ ... }` builds a map. A single word before `:` is a string key, so `ada` is the key `"ada"`.
`ages["grace"] = 85` adds an entry, or replaces the value when the key is already there. Reading a
key gives the value or `null`, since the key may be missing, so `?? 0` supplies a default. A `for`
loop over a map visits each entry in the order its key was first inserted, and `{ key, value }`
takes the entry apart.

A set has no literal: `Set.of` builds one from an array and drops the second `"red"`. `Set.add` adds
a member and `Set.has` asks for one. A set keeps insertion order too.

An empty collection needs its type written out, because nothing in it says what it holds:

```ant
let names: Array[string] = []
let counts: Map[string, int] = Map{}
let seen: Set[int] = Set.of([])
Array.push(names, "ada")
counts["ada"] = 1
Set.add(seen, 7)
println(names)
println(counts)
println(seen)
```

```text
["ada"]
Map{ "ada": 1 }
Set.of([7])
```

An index out of range panics, whether you read or write it. `Array.get` returns `null` instead:

```ant
let xs = [1, 2, 3]
Array.get(xs, 5) // null
xs[5]            // panic: index 5 out of range for an array of length 3
```

Try it: push a few more scores and print `Array.sort(scores)`, or remove `"linus"` with `Map.remove`
and loop again.

The [collections guide](../guide/collections.md) explains equality, identity, and lazy sequences.
The reference has every function on [Array](../lib/Array/index.md), [Map](../lib/Map/index.md), and
[Set](../lib/Set/index.md), and the literals for [arrays](../language/literals/arrays.md) and
[maps](../language/literals/maps.md).
