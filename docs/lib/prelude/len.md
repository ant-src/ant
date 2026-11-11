<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# len

```ant signature
func len[T](x: T): int
```

len returns the size of x: an array's elements, a map's entries, a set's members, a string's code
points, the bytes of Bytes, or the values a channel buffers now. x may also be a named primitive
over string, which it unwraps.

<!-- end generated -->

## Details

- `len` of a channel is the number of values buffered at that moment. Another goroutine can change
  it right after; an unbuffered channel always has length 0.
- Any other argument, including a record, a number, or an optional such as `string | null`, is a
  compile error. Check for `null` first.

## Basic examples

Count the elements of an array and the code points of a string:

```ant
len([10, 20, 30]) // 3
len("hello")      // 5
```

Count the entries of a map:

```ant
len(Map{ "apples": 3, "pears": 0 }) // 2
```

## Scope

An empty value has length 0:

```ant
let none: Array[int] = []
let empty: Map[string, int] = Map{}
len(none)  // 0
len("")    // 0
len(empty) // 0
```

A set counts its members, so duplicates count once:

```ant
len(Set.of([1, 2, 2, 3])) // 3
```

A string counts code points, and `Bytes` counts bytes:

```ant
len("héllo")                   // 5
len("héllo".toBytes())         // 6
len(Bytes.fromArray([0, 255])) // 2
```

A named primitive over `string` is unwrapped:

```ant
type UserId = string
len(UserId("u-42")) // 4
```

A union works when `len` takes every member:

```ant
type Text = string | Array[string]

func size(t: Text): int {
  len(t)
}

size("abc")      // 3
size(["a", "b"]) // 2
```

A channel reports the values it buffers now:

```ant
let c: Channel[int] = Channel.new(3)
c.send(1)
c.send(2)
println(len(c))
c.receive()
println(len(c))
```

```text
2
1
```

Pass `len` as a function where the expected type fixes its parameter:

```ant
["a", "bb", "ccc"].map(len) // [1, 2, 3]
```

## Applications

Read the last element of an array:

```ant
let steps = ["plan", "build", "ship"]
steps[len(steps) - 1] // "ship"
```

Sort words by length:

```ant
["pear", "fig", "banana"].sortBy(len) // ["fig", "pear", "banana"]
```

Reject a password that is too short:

```ant
func strong(password: string): bool {
  len(password) >= 8
}

strong("hunter2")       // false
strong("correct horse") // true
```

Count the distinct words of a sentence:

```ant
let words = "the cat saw the dog".split(" ")
len(words)         // 5
len(Set.of(words)) // 4
```

## Properties and relations

`len` agrees with the namespace function of each kind:

```ant
let xs = [3, 1, 4, 1, 5]
let s = "naïve"
assert(len(xs) == Array.length(xs))
assert(len(s) == String.length(s))
assert(len(Set.of(xs)) == Set.size(Set.of(xs)))
```

A set built from an array is never longer than the array, and a string's UTF-8 bytes are never fewer
than its code points:

```ant
for s in ["", "abc", "héllo", "日本"] {
  assert(len(s.toBytes()) >= len(s))
}
let xs = [1, 1, 2]
assert(len(Set.of(xs)) <= len(xs))
```

Joining two strings adds their lengths:

```ant
let a = "foo"
let b = "bär"
assert(len(a + b) == len(a) + len(b))
```

## Possible issues

`len` counts code points, so a character built from several code points counts more than once:

```ant
len("e\u{301}")             // 2
len("e\u{301}".graphemes()) // 1
```

The last valid index is `len(xs) - 1`; indexing at `len(xs)` panics:

```ant
let xs = ["a", "b"]
xs[len(xs)] // panic: index 2 out of range for an array of length 2
```

`len` does not take an optional. Check for `null` first:

```ant
let name: string | null = null
let n = len(name)
// error: len needs an array, a map, a set, a string, Bytes, or a channel, got string | null
```

A record has no length:

```ant
type Point = { x: int, y: int }
let n = len(Point{ x: 1, y: 2 })
// error: len needs an array, a map, a set, a string, Bytes, or a channel, got Point
```

`len` is generic, so it needs an expected function type to be a value:

```ant
let f = len
// error: len is generic; call it, or give the expected function type that fixes its type parameters
```

```ant
let f: func(string): int = len
f("abc") // 3
```

## See also

[Array.length](../Array/length.md), [String.length](../String/length.md),
[String.graphemes](../String/graphemes.md), [Bytes.length](../Bytes/length.md),
[Map.size](../Map/size.md), [Set.size](../Set/size.md), [Channel.length](../Channel/length.md),
[Array.isEmpty](../Array/isEmpty.md), [String.isEmpty](../String/isEmpty.md)
