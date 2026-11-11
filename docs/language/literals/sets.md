<!-- status: reviewed -->

# Set literals

Ant has no set literal. A set is written as a call to `Set.of` on an array literal:
`Set.of([1, 2, 3])` is a `Set[int]`. A set also renders in that form, so what you write is what
`println` and the REPL show.

## Details

- `Set.of` keeps the first of each `==` value, in the order of the array. Duplicates in the array
  are dropped, not an error.
- The empty set is `Set.of([])`. It takes its element type from the expected type, such as an
  annotation, as `[]` does.
- Each evaluation builds a new set, which you can then change with `Set.add` and `Set.remove`.

## Basic examples

```ant
Set.of([1, 2, 3])       // Set.of([1, 2, 3])
Set.of(["red", "blue"]) // Set.of(["red", "blue"])
Set.of([3, 1, 3, 2])    // Set.of([3, 1, 2])
```

The empty set takes its element type from an annotation:

```ant
let seen: Set[string] = Set.of([])
seen.add("ant") // true
seen            // Set.of(["ant"])
```

## Scope

Everything an array literal allows goes inside, spreads included:

```ant
let primary = ["red", "blue"]
Set.of([...primary, "green", "red"]) // Set.of(["red", "blue", "green"])
```

Elements of different types give a set of their union type:

```ant
let mixed = Set.of([1, "one", null])
mixed.has("one") // true
mixed            // Set.of([1, "one", null])
```

Records, arrays, and maps are members by contents, so equal values collapse into one:

```ant
type Point = { x: int, y: int }
Set.of([Point{ x: 1, y: 2 }, Point{ x: 1, y: 2 }]) // Set.of([Point{ x: 1, y: 2 }])
Set.of([[1, 2], [1, 2], [2, 1]])                   // Set.of([[1, 2], [2, 1]])
```

A member is hashed when it is added, so do not change a record or array after it goes into a set;
the set may no longer find it.

A set can be a field default, and every record built gets a set of its own:

```ant
type User = { name: string, roles: Set[string] = Set.of([]) }
let ann = User{ name: "Ann" }
let bob = User{ name: "Bob" }
ann.roles.add("admin") // true
bob.roles              // Set.of([])
```

In an array literal, `Set.of([])` takes its element type from the other elements, and in a map
literal from the other values:

```ant
let teams = [Set.of(["ann", "bob"]), Set.of([])]
teams.map({ x.size() }) // [2, 0]
let roles = Map{ ann: Set.of(["admin"]), bob: Set.of([]) }
roles["bob"]?.size() // 0
```

A long set goes one member per line, with a trailing comma in the array:

```ant
let weekdays = Set.of([
  "Mon",
  "Tue",
  "Wed",
  "Thu",
  "Fri",
])
println(weekdays.has("Sat"))
```

```text
false
```

## Applications

Keep a fixed list of allowed values, and check input against it:

```ant
let methods = Set.of(["GET", "HEAD", "POST"])
for m in ["GET", "PATCH"] {
  let verdict = if methods.has(m) { "allowed" } else { "rejected" }
  println("${m}: ${verdict}")
}
```

```text
GET: allowed
PATCH: rejected
```

Remove duplicates from a list while keeping its order:

```ant
let tags = "go ant go zig ant" | String.split(" ") | Set.of() | Set.values()
tags // ["go", "ant", "zig"]
```

## Properties and relations

Two sets with the same members are `==`, in any order:

```ant
assert(Set.of([1, 2, 3]) == Set.of([3, 2, 1]))
assert(Set.of([1, 1, 2]) == Set.of([1, 2]))
```

Their order still shows in `Set.values` and in rendering:

```ant
Set.of([3, 2, 1]).values() // [3, 2, 1]
```

Two evaluations of the same expression are equal, but separate sets:

```ant
let a = Set.of([1])
let b = Set.of([1])
assert(a == b)
assert(!Set.same(a, b))
```

## Possible issues

`Set{}` is read as a record literal with the head `Set`:

```ant
let s: Set[int] = Set{} // error: a record literal needs a record type as its head
```

`Set.of` takes one array, not the members as separate arguments:

```ant
let s = Set.of(1, 2) // error: Set.of takes 1 argument, got 2
```

The empty set needs an expected type:

```ant
let s = Set.of([]) // error: cannot infer the element type of an empty array; annotate it
```

A set has no indexing. Use `Set.values` for an array to index:

```ant
let s = Set.of([10, 20])
println(s[0]) // error: Set[int] cannot be indexed
```

## See also

[Set](../types/Set.md) · [Set.of](../../lib/Set/of.md) · [Set functions](../../lib/Set/index.md) ·
[Array literals](arrays.md) · [Map literals](maps.md) · [Record literals](records.md)
