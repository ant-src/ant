<!-- status: reviewed -->

# == and !=

`a == b` is `true` when two values are equal, and `a != b` is its negation. Equality is structural:
it compares what values hold, not where they live.

## Details

- One operand's type must be assignable to the other's. Comparing an `int` with a `string`, or a
  `Point` with a `Vec`, is a compile error.
- Arrays, maps, sets, and records compare by contents, at every depth. Functions, sync objects such
  as `Mutex` and `Channel`, `Random` generators, and resources such as file handles and processes
  compare by identity.
- `==` never panics. Values of different kinds are unequal, so `x == null` is `false` for any
  non-null `x`.
- Floats follow IEEE 754: `NaN` equals nothing, itself included, and `-0.0 == 0.0`.

## Basic examples

```ant
1 == 1           // true
1 != 2           // true
"ant" == "Ant"   // false
[1, 2] == [1, 2] // true
```

Records are equal when they have the same type and equal fields:

```ant
type Point = { x: int, y: int }
Point{ x: 1, y: 2 } == Point{ x: 1, y: 2 } // true
Point{ x: 1, y: 2 } != Point{ x: 2, y: 1 } // true
```

## Scope

### Collections

Arrays compare their elements in order. Maps compare keys and values, and sets compare members, in
any order:

```ant
[1, 2] == [2, 1]                                       // false
[[1, 2], [3]] == [[1, 2], [3]]                         // true
Map{ "a": 1, "b": 2 } == Map{ "b": 2, "a": 1 }         // true
Set.of(["tea", "coffee"]) == Set.of(["coffee", "tea"]) // true
```

### Strings

Strings are equal when they hold the same code points. Case matters, and so does how an accented
letter is written:

```ant
"ant" == "ant"         // true
"ant" == "ANT"         // false
"ant".equalFold("ANT") // true
"é" == "e\u{301}"      // false
```

### null and unions

`== null` tests a nullable value, and the checker narrows the value after the test:

```ant
func shout(name: string | null): string {
  if name == null { return "NOBODY" }
  name.toUpper()
}

println(shout("ann"))
println(shout(null))
```

```text
ANN
NOBODY
```

A union value compares with a value of any of its members. A value holding a different member is
unequal, not an error:

```ant
type Circle = { r: int }
type Rect = { w: int, h: int }
type Shape = Circle | Rect
let shape: Shape = Circle{ r: 1 }
shape == Circle{ r: 1 }     // true
shape == Rect{ w: 1, h: 1 } // false
let id: string | int = "7"
id == 7 // false
```

### Named primitive types

A value of a named primitive type equals another of the same type with an equal payload:

```ant
type UserId = string
UserId("u-1") == UserId("u-1") // true
UserId("u-1") == UserId("u-2") // false
```

### Floats

`NaN` is unequal to everything, itself included, also inside a collection:

```ant
let nan = 0.0 / 0.0
nan == nan     // false
nan != nan     // true
[nan] == [nan] // false
-0.0 == 0.0    // true
```

### Functions

Functions compare by identity. Two functions with the same body are different values:

```ant
let double = func(n: int): int { n * 2 }
let twice = func(n: int): int { n * 2 }
double == double // true
double == twice  // false
```

### Other values

`Html` compares nodes, so markup built two ways is one value. Regular expressions compare source and
flags, and type values compare the types they denote:

```ant
let pattern = /ant/i
<b>3</b> == <b>{3}</b>                   // true
pattern == /ant/i                        // true
type(Array[int]) == type(Array[int])     // true
type(int | string) == type(string | int) // true
```

### Cyclic values

A value that contains itself compares in finite time:

```ant
type Node = { name: string, next: Node | null }
let a = Node{ name: "loop", next: null }
a.next = a
let b = Node{ name: "loop", next: null }
b.next = b
a == b // true
```

## Applications

Search an array for a record by value. `contains` compares with `==`:

```ant
type Point = { x: int, y: int }
let visited = [Point{ x: 0, y: 0 }, Point{ x: 1, y: 2 }]
visited.contains(Point{ x: 1, y: 2 }) // true
visited.contains(Point{ x: 5, y: 5 }) // false
```

Use arrays and records as map keys. A key matches any `==` key, not only the same value:

```ant
let routes = Map{ [0, 0]: "origin", [1, 2]: "depot" }
routes[[1, 2]] // "depot"
routes[[9, 9]] // null
```

Skip work when a value has not changed:

```ant
type Config = { port: int, debug: bool }
var current = Config{ port: 8080, debug: false }
func apply(next: Config) {
  if next == current {
    println("unchanged")
    return
  }
  current = next
  println("reloaded on port ${next.port}")
}
apply(Config{ port: 8080, debug: false })
apply(Config{ port: 9090, debug: false })
```

```text
unchanged
reloaded on port 9090
```

## Properties and relations

`a != b` is `!(a == b)`:

```ant
let a = [1, 2]
let b = [1, 3]
assert((a != b) == !(a == b))
```

`==` compares contents; `Array.same`, `Map.same`, `Set.same`, and `Record.same` compare identity:

```ant
let xs = [1, 2]
let ys = [1, 2]
let zs = xs
assert(xs == ys && !Array.same(xs, ys))
assert(Array.same(xs, zs))
```

Equality looks at the current contents, so a change through one name can make two values unequal:

```ant
let xs = [1, 2]
let ys = [1, 2]
assert(xs == ys)
xs.push(3)
assert(xs != ys)
```

## Possible issues

### Values of unrelated types

Neither `int` nor `string` is assignable to the other, so the comparison does not compile. Convert
one side first:

```ant
let input = "42"
let matches = input == 42 // error: string and int cannot be compared with ==: neither is assignable to the other
```

```ant
let input = "42"
input.toInt() == 42 // true
```

### `int` and `float64` are different types

There is no implicit conversion between numeric types:

```ant
let half = 1 == 1.0 // error: int and float64 cannot be compared with ==: neither is assignable to the other
```

```ant
float64(1) == 1.0 // true
```

### Records with the same fields

Two record types with the same fields are still different types:

```ant
type Point = { x: int, y: int }
type Vec = { x: int, y: int }
let same = Point{ x: 1, y: 2 } == Vec{ x: 1, y: 2 }
// error: Point and Vec cannot be compared with ==: neither is assignable to the other
```

### A named primitive and its base

A `UserId` is not a `string`. Tag the other side, or unwrap with `string(...)`:

```ant
type UserId = string
let id = UserId("u-1")
let found = id == "u-1" // error: UserId and string cannot be compared with ==: neither is assignable to the other
```

```ant
type UserId = string
let id = UserId("u-1")
id == UserId("u-1") // true
string(id) == "u-1" // true
```

### Testing a non-nullable value for `null`

A value whose type excludes `null` can never be `null`, and the checker rejects the test:

```ant
let name = "Ann"
let missing = name == null // error: string and null cannot be compared with ==: neither is assignable to the other
```

### Comparing with an empty literal or a generic call

`==` gives no expected type to its other side, so `[]`, `Map{}`, and a call such as `Result.ok(1)`
cannot infer their types there. Use `isEmpty`, or bind the value with a type first:

```ant
let xs: Array[int] = []
let empty = xs == [] // error: cannot infer the element type of an empty array; annotate it
```

```ant
let xs: Array[int] = []
xs.isEmpty() // true
let r: Result[int, Error] = Result.ok(1)
let expected: Result[int, Error] = Result.ok(1)
r == expected // true
```

### Float rounding

Float arithmetic rounds, so a sum can miss the literal it seems to equal by the last bit. Compare
with a tolerance:

```ant
0.1 + 0.2 == 0.3                   // false
Math.abs(0.1 + 0.2 - 0.3) < 1.0e-9 // true
```

### A record literal in an if header

In the header of an `if` or `while`, a record literal must be parenthesized:

```ant
type Point = { x: int, y: int }
let p = Point{ x: 1, y: 2 }
if p == Point{ x: 1, y: 2 } { println("home") } // error: a record or map literal in the if header must be parenthesized
```

```ant
type Point = { x: int, y: int }
let p = Point{ x: 1, y: 2 }
if p == (Point{ x: 1, y: 2 }) {
  println("home")
}
```

```text
home
```

### A pipeline as an operand

`|` binds looser than `==`, so a pipeline used as an operand must be parenthesized:

```ant
let blank = " " | String.trim() == "" // error: '|' binds looser than '==', so parenthesize the pipe, as in (a | f()) == b
```

```ant
(" " | String.trim()) == "" // true
" ".trim() == ""            // true
```

### Arrays are invariant

An `Array[int]` is not an `Array[int | null]`, so the two do not compare even with equal elements:

```ant
let a: Array[int | null] = [1, 2]
let b: Array[int] = [1, 2]
let same = a == b
// error: Array[int | null] and Array[int] cannot be compared with ==: neither is assignable to the other
```

### There is no `===`

Ant has one equality operator. For identity, call `same` on the namespace:

```ant
let a = [1]
let b = a === a // error: unexpected '='
```

## See also

[Comparison operators](comparison.md) · [is](is.md) · [Logical operators](logical.md) ·
[Array.same](../../lib/Array/same.md) · [Record.same](../../lib/Record/same.md) ·
[String.equalFold](../../lib/String/equalFold.md) · [null](../literals/null.md) ·
[Values](../../guide/values.md)
