<!-- status: reviewed -->

# Indexing with []

`a[i]` reads an element of an array, a byte of `Bytes`, or the value of a map key. `a[i] = v` writes
an element of an array or the value of a map key.

## Details

- An array or `Bytes` index is an `int` from `0` up to but excluding the length. Any other `int`,
  negative ones included, panics. `Array.get` returns `null` instead.
- A map read gives `V | null`: the value whose key is `==` to the index, or `null` when there is
  none. A map write inserts the key or replaces its value.
- `let` stops a name from being rebound, not its elements: `xs[0] = v` works on a `let` array.
- Strings, sets, and records have no indexing. `Bytes` is read-only.

## Basic examples

Read an array element by its position, counting from `0`:

```ant
let langs = ["ant", "go", "c"]
langs[0] // "ant"
langs[2] // "c"
```

Read a map value by its key. A missing key gives `null`:

```ant
let ports = Map{ http: 80, https: 443 }
ports["https"] // 443
ports["ftp"]   // null
```

Write an element or an entry with `=`:

```ant
let scores = [10, 20, 30]
scores[1] = 25
scores // [10, 25, 30]
let ages = Map{ ada: 36 }
ages["linus"] = 54
ages // Map{ "ada": 36, "linus": 54 }
```

## Scope

### Computed indices

The index is any `int` expression:

```ant
let xs = [3, 1, 4, 1, 5]
xs[xs.length() - 1] // 5
xs[xs[1]]           // 1
```

### Nested arrays

Each `[]` indexes the result of the one before it:

```ant
let grid = [[1, 2], [3, 4]]
grid[1][0] // 3
grid[0][1] = 9
grid // [[1, 9], [3, 4]]
```

### Elements that are records

A field access follows an index directly:

```ant
type User = { name: string, admin: bool }
let users = [User{ name: "ann", admin: true }, User{ name: "bob", admin: false }]
users[1].name // "bob"
users[1].admin = true
users[1] // User{ name: "bob", admin: true }
```

### Map keys of any type

Keys compare with `==`, so a number, a record, or an array works as an index:

```ant
type Point = { x: int, y: int }
let squares = Map{ 2: 4, 3: 9 }
squares[3] // 9
let labels = Map{ Point{ x: 0, y: 0 }: "origin" }
labels[Point{ x: 0, y: 0 }] // "origin"
let cells = Map{ [1, 2]: "a" }
cells[[1, 2]] // "a"
```

### Bytes

Reading `Bytes` gives a `byte`. Compare it with a `byte`, and convert it before arithmetic with an
`int`:

```ant
let header = Bytes.fromArray([0x89, 0x50, 0x4e, 0x47])
header[1]               // 80
header[0] == byte(0x89) // true
int(header[1]) + 1      // 81
```

### Compound assignment

`+=` and the other compound operators read the element, apply the operator, and store the result:

```ant
let counts = [0, 0, 0]
counts[2] += 5
counts // [0, 0, 5]
```

### Destructuring assignment

Elements can be targets of an array destructuring, which swaps two of them in one statement:

```ant
let xs = [1, 2, 3]
[xs[0], xs[2]] = [xs[2], xs[0]]
xs // [3, 2, 1]
```

### Optional indexing

`?.[i]` gives `null` when the indexed value is `null`, and indexes it otherwise. It reads a map
nested in another map:

```ant
let config = Map{ db: Map{ host: "localhost" } }
config["db"]?.["host"]    // "localhost"
config["cache"]?.["host"] // null
```

### Evaluation order

In an assignment, the index evaluates before the right side:

```ant
func at(i: int): int {
  println("index ${i}")
  i
}

func value(): int {
  println("value")
  7
}

let xs = [0, 0]
xs[at(1)] = value()
println(xs)
```

```text
index 1
value
[0, 7]
```

## Applications

### Counting words

Read the current count, default it with `??`, and write it back:

```ant
let counts: Map[string, int] = Map{}
for word in "the ant and the bee".split(" ") {
  counts[word] = (counts[word] ?? 0) + 1
}
println(counts)
```

```text
Map{ "the": 2, "ant": 1, "and": 1, "bee": 1 }
```

### Neighbors in a grid

Guard the bounds before indexing, so a cell on the edge has fewer neighbors instead of a panic:

```ant
let grid = [
  [1, 2, 3],
  [4, 5, 6],
]

func cell(row: int, col: int): int | null {
  if row < 0 || row >= grid.length() { return null }
  grid[row].get(col)
}

func neighborSum(row: int, col: int): int {
  let up = cell(row - 1, col) ?? 0
  let down = cell(row + 1, col) ?? 0
  let left = cell(row, col - 1) ?? 0
  let right = cell(row, col + 1) ?? 0
  up + down + left + right
}

println(neighborSum(0, 0))
println(neighborSum(1, 1))
```

```text
6
12
```

### Reading a header field

Check the length and the magic prefix first, then read the version byte at its fixed offset:

```ant
let magic = "AN".toBytes()

func version(data: Bytes): int | null {
  if data.length() < 3 || !data.startsWith(magic) { return null }
  int(data[2])
}

println(version(Bytes.fromArray([0x41, 0x4e, 0x02, 0xff])))
println(version(Bytes.fromArray([0x41, 0x4e])))
println(version("GIF".toBytes()))
```

```text
2
null
null
```

## Properties and relations

A map read is `Map.get`, and a map write is `Map.set`:

```ant
let m = Map{ a: 1 }
assert(m["a"] == m.get("a"))
assert(m["z"] == m.get("z"))
m["b"] = 2
m.set("c", 3)
assert(m == Map{ a: 1, b: 2, c: 3 })
```

An in-range array read is `Array.get`, which also answers `null` out of range:

```ant
let xs = [10, 20]
assert(xs[1] == xs.get(1))
assert(xs.get(2) == null)
```

## Possible issues

### Out of range

An index past the end panics, and so does a negative index. There is no counting from the end:

```ant
let xs = [10, 20, 30]
xs[3]  // panic: index 3 out of range for an array of length 3
xs[-1] // panic: index -1 out of range for an array of length 3
```

An empty array has no valid index:

```ant
let empty: Array[int] = []
empty[0] // panic: index 0 out of range for an array of length 0
```

Use `xs.get(i)` or `xs.last()` when the index may be out of range.

### An array element is not nullable

An array read gives `T`, not `T | null`, so `??` after it is a compile error. Use `get` to get a
`null` you can default:

```ant
let xs = [1, 2]
let first = xs[0] ?? 0 // error: ?? needs a nullable left side, got int
```

```ant
let xs = [1, 2]
xs.get(5) ?? 0 // 0
```

### A map read is nullable

A map read gives `V | null`, so the result cannot be indexed again or used as a `V` before it is
narrowed:

```ant
let config = Map{ db: Map{ host: "localhost" } }
println(config["db"]["host"])
// error: Map[string, string] | null cannot be indexed
```

Use `?.[` as in Scope, `??`, or a guard.

### Missing key and stored null

When the value type includes `null`, a read cannot tell a missing key from a stored `null`. Ask
`has`:

```ant
let limits: Map[string, int | null] = Map{ cpu: null }
limits["cpu"]        // null
limits["memory"]     // null
limits.has("cpu")    // true
limits.has("memory") // false
```

### Comparing a byte with an int

A `Bytes` read is a `byte`, and an integer literal next to it is an `int`, so `==` between them is a
compile error. Convert the literal:

```ant
let header = Bytes.fromArray([0x89, 0x50])
println(header[0] == 0x89) // error: byte and int cannot be compared with ==: neither is assignable to the other
```

```ant
let header = Bytes.fromArray([0x89, 0x50])
header[0] == byte(0x89) // true
```

### Strings have no indexing

Split a string into what you need first, such as `chars` for code points:

```ant
let s = "ant"
println(s[0]) // error: string cannot be indexed
```

```ant
"ant".chars()[0] // "a"
```

### Records have no dynamic access

A record's fields are fixed names. Use a map when the keys are data:

```ant
type User = { name: string }
let user = User{ name: "ann" }
println(user["name"]) // error: a record has no dynamic access; use a Map when the keys are data
```

### A byte of Bytes is a `byte`

Assigning into `Bytes` takes a `byte`, not an `int`:

```ant
let b = "hi".toBytes()
b[0] = 72
// error: cannot use int where byte is expected
```

### Index type

An array index is an `int`. A float, even a whole one, is a compile error; convert it first:

```ant
let xs = [1, 2]
println(xs[1.0]) // error: cannot use float64 where int is expected
```

## See also

- [Array.get](../../lib/Array/get.md), [Array.first](../../lib/Array/first.md),
  [Array.last](../../lib/Array/last.md)
- [Map.get](../../lib/Map/get.md), [Map.set](../../lib/Map/set.md), [Map.has](../../lib/Map/has.md)
- [Bytes](../../lib/Bytes/index.md), [String.chars](../../lib/String/chars.md)
- [Optional chaining](optional-chaining.md), [Null coalescing](null-coalescing.md),
  [Assignment](assignment.md)
- [Arrays](../types/Array.md), [Maps](../types/Map.md)
