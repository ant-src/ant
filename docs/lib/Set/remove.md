<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Set.remove

```ant signature
func Set.remove[T](s: Set[T], x: T): bool
```

Set.remove removes x from s and reports whether it was a member. It mutates s.

<!-- end generated -->

## Details

- Changes `s` in place, so every name that shares the set sees the smaller set.
- Members compare by `==`, so a fresh record or array with the same contents removes the member.
- The other members keep their order. A member removed and added again goes to the end.

## Basic examples

Remove a member, and the result says it was there:

```ant
var tags = Set.of(["red", "green", "blue"])
Set.remove(tags, "green") // true
tags                      // Set.of(["red", "blue"])
```

A value that is not a member gives `false` and leaves the set alone:

```ant
var tags = Set.of(["red", "blue"])
tags.remove("pink") // false
tags                // Set.of(["red", "blue"])
```

Removing the same value twice succeeds only once:

```ant
var online = Set.of(["ada", "linus"])
online.remove("ada") // true
online.remove("ada") // false
```

## Scope

An empty set has nothing to remove. Annotate it so that `T` is known:

```ant
var empty: Set[int] = Set.of([])
empty.remove(1) // false
```

Records compare by tag and fields, so a fresh record with the same fields removes the member:

```ant
type Point = { x: int, y: int }
var visited = Set.of([Point{ x: 1, y: 2 }, Point{ x: 3, y: 4 }])
visited.remove(Point{ x: 1, y: 2 }) // true
visited                             // Set.of([Point{ x: 3, y: 4 }])
```

Arrays and maps compare by contents. Two maps with the same entries are equal in any order:

```ant
var shapes = Set.of([[1, 2], [3]])
shapes.remove([1, 2]) // true
shapes.remove([2, 1]) // false
var configs = Set.of([Map{ a: 1, b: 2 }])
configs.remove(Map{ b: 2, a: 1 }) // true
```

Strings compare by code points, with no normalization:

```ant
var words = Set.of(["café"])
words.remove("cafe\u{301}") // false
words.remove("café")        // true
```

When `T` includes `null`, `null` can be a member and can be removed:

```ant
var answers: Set[string | null] = Set.of(["yes", null])
answers.remove(null) // true
answers              // Set.of(["yes"])
```

Call it through `.`, with a pipe, or with named arguments in any order:

```ant
var s = Set.of([1, 2, 3])
s.remove(1)            // true
s | Set.remove(2)      // true
Set.remove(x: 3, s: s) // true
s                      // Set.of([])
```

A set bound by `let` can still lose members, and every name for it sees the change:

```ant
let a = Set.of(["x", "y"])
var b = a
b.remove("x") // true
a             // Set.of(["y"])
```

## Applications

Toggle a member: remove it when present, add it otherwise:

```ant
func toggle(s: Set[string], x: string) {
  if !s.remove(x) {
    s.add(x)
  }
}

var selected = Set.of(["bold"])
toggle(selected, "bold")
toggle(selected, "italic")
println(selected)
```

```text
Set.of(["italic"])
```

Track work still in flight, and report a task that finishes twice:

```ant
var pending = Set.of(["build", "test", "lint"])

func finish(task: string) {
  if pending.remove(task) {
    println("${task} done, ${pending.size()} left")
  } else {
    println("${task} was not pending")
  }
}

finish("test")
finish("test")
finish("lint")
```

```text
test done, 2 left
test was not pending
lint done, 1 left
```

Move a member to the end, to keep the most recently used one last:

```ant
func touch(recent: Set[string], file: string) {
  recent.remove(file)
  recent.add(file)
}

var recent = Set.of(["a.ant", "b.ant", "c.ant"])
touch(recent, "a.ant")
println(recent.values())
```

```text
["b.ant", "c.ant", "a.ant"]
```

## Properties and relations

The result is what [Set.has](has.md) said before the call, and afterward `x` is not a member:

```ant
var s = Set.of([1, 2, 3])
for x in [2, 9] {
  var had = s.has(x)
  assert(s.remove(x) == had)
  assert(!s.has(x))
}
```

The size drops by one exactly when the result is `true`:

```ant
var s = Set.of(["a", "b"])
for x in ["a", "z"] {
  var size = s.size()
  var removed = s.remove(x)
  assert(s.size() == if removed { size - 1 } else { size })
}
```

`Set.remove` undoes a [Set.add](add.md) that returned `true`:

```ant
var s = Set.of([1, 2])
var before = Set.of(s.values())
assert(s.add(3))
assert(s.remove(3))
assert(s == before)
```

It has the effect of [Set.difference](difference.md) with a one-member set, but in place:

```ant
var s = Set.of([1, 2, 3])
var expected = Set.difference(s, Set.of([2]))
s.remove(2)
assert(s == expected)
```

## Possible issues

The result is a `bool`, not the removed member:

```ant
var s = Set.of([1, 2])
var n: int = s.remove(1) // error: cannot use bool where int is expected
```

The value must have the set's member type. Numeric types never mix:

```ant
var prices = Set.of([1.5, 2.0])
prices.remove(2) // error: cannot use int where float64 is expected
```

A member is hashed when it is added. Changing an array, map, or record after adding it to a set is a
program error, and the set may no longer find it, so `remove` cannot be relied on. Remove the member
first, change it, then add it again.

`NaN` is not `==` to itself, so a `NaN` member cannot be removed. `0.0` and `-0.0` are `==`, so one
removes the other:

```ant
var xs = Set.of([Math.NaN, 0.0])
xs.remove(Math.NaN) // false
xs.remove(-0.0)     // true
xs                  // Set.of([NaN])
```

A member removed during a `for` over the set before the loop reaches it is not visited:

```ant
var s = Set.of([1, 2, 3, 4])
var seen: Array[int] = []
for v in s {
  if v == 1 {
    s.remove(3)
  }
  seen.push(v)
}
seen // [1, 2, 4]
```

`Set.remove` takes one value, not a test. To keep only the members that pass a test, filter
[Set.values](values.md) with [Array.filter](../Array/filter.md) and build a new set:

```ant
var s = Set.of([1, 2, 3, 4])
Set.of(s.values().filter({ x % 2 == 1 })) // Set.of([1, 3])
```

## See also

[Set.add](add.md) · [Set.has](has.md) · [Set.difference](difference.md) · [Set.size](size.md) ·
[Set.values](values.md) · [Set.of](of.md) · [Map.remove](../Map/remove.md) ·
[Array.removeAt](../Array/removeAt.md)
