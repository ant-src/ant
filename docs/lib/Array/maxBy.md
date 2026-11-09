<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Array.maxBy

```ant signature
func Array.maxBy[T, K: Ordered](xs: Array[T], key: func(T): K): T | null
```

Array.maxBy returns the first element of xs with the greatest key, or null when xs is empty. An
element replaces the one held only when its key is strictly greater, so of equal keys the first wins
and a NaN key never wins.

<!-- end generated -->

## Details

- Returns the element, not its key, and `null` when `xs` is empty.
- Calls `key` once per element, from first to last, and never stops early.
- Of equal keys, the first element wins, and a `NaN` key never wins.

## Basic examples

Find the longest word:

```ant
Array.maxBy(["fig", "banana", "kiwi"], { String.length(x) }) // "banana"
```

An empty array has no greatest element, so the result is `null`:

```ant
let empty: Array[string] = []
empty.maxBy({ String.length(x) }) // null
```

Find the oldest person, and read a field of the result with `?.`:

```ant
type Person = { name: string, age: int }
let people = [Person{ name: "Ann", age: 31 }, Person{ name: "Bob", age: 45 }]
people.maxBy({ x.age })?.name // "Bob"
```

## Scope

The result is the element, not its key:

```ant
[3, -7, 5].maxBy({ Math.abs(x) }) // -7
```

Of equal keys, the first element wins:

```ant
["ab", "cd", "e"].maxBy({ String.length(x) }) // "ab"
```

A string key compares by code point, so every uppercase ASCII letter comes before every lowercase
one:

```ant
["apple", "Banana", "cherry"].maxBy({ x }) // "cherry"
["Zed", "apple"].maxBy({ x })              // "apple"
```

A float key works too. To find the least, use `Array.minBy`:

```ant
let temps = [12.5, -3.0, 21.25]
temps.maxBy({ x }) // 21.25
temps.minBy({ x }) // -3.0
```

A `NaN` key never replaces the element held. A `NaN` in the first element is held to the end, since
no key is greater than it:

```ant
[1.0, 0.0 / 0.0, 3.0].maxBy({ x }) // 3.0
[0.0 / 0.0, 1.0, 3.0].maxBy({ x }) // NaN
```

Pass a named function instead of a lambda:

```ant
func vowels(s: string): int {
  String.graphemes(s).countWhere({ "aeiou".contains(x) })
}

["sky", "queue", "area"].maxBy(vowels) // "queue"
```

A key can use the local variables around it. Find the number closest to a target:

```ant
let target = 10
[3, 12, 7, 20].maxBy({ -Math.abs(x - target) }) // 12
```

The element type may include `null`. Give the key a value for it with `??`:

```ant
let scores: Array[int | null] = [null, 70, 85]
scores.maxBy({ x ?? 0 }) // 85
```

Find the greatest grapheme of a string by splitting it first:

```ant
String.graphemes("hello").maxBy({ x }) // "o"
```

Call it through `.` or in a pipe:

```ant
["a", "bbb", "cc"].maxBy({ String.length(x) })         // "bbb"
["a", "bbb", "cc"] | Array.maxBy({ String.length(x) }) // "bbb"
```

## Applications

Find the key of the largest value in a map, through its entries:

```ant
let stock = Map{ "nut": 3, "bolt": 12, "gear": 7 }
stock.entries().maxBy({ x.value })?.key // "bolt"
```

Find the most frequent word, with `Array.groupBy`:

```ant
let words = String.split("the ant and the bee and the ant", " ")
words.groupBy({ x }).entries().maxBy({ x.value.length() })?.key // "the"
```

Find the index of the greatest element, through `Array.entries`:

```ant
[4, 9, 2, 9].entries().maxBy({ x.value })?.index // 1
```

Return early from a function when the array is empty, with `?? return`:

```ant
type Run = { name: string, ms: float64 }

func slowest(runs: Array[Run]): string {
  let run = runs.maxBy({ x.ms }) ?? return "no runs"
  "${run.name} took ${run.ms} ms"
}

println(slowest([Run{ name: "parse", ms: 12.5 }, Run{ name: "check", ms: 30.0 }]))
println(slowest([]))
```

```text
check took 30.0 ms
no runs
```

## Properties and relations

With the identity as the key, `Array.maxBy` gives what `Array.max` gives on integers and strings:

```ant
let xs = [4, 9, 2, 9]
assert(xs.maxBy({ x }) == xs.max())
let names = ["Ann", "Bob", "Al"]
assert(names.maxBy({ x }) == names.max())
```

Negating a float key, or a signed key that is never the most negative value of its kind, turns it
into `Array.minBy`, and of equal keys both keep the first element. On an unsigned key, or on the
most negative signed value, the negation panics; see Possible issues:

```ant
type Box = { name: string, size: int }
let boxes = [Box{ name: "a", size: 2 }, Box{ name: "b", size: 1 }, Box{ name: "c", size: 1 }]
assert(boxes.maxBy({ -x.size })?.name == "b")
assert(boxes.maxBy({ -x.size }) == boxes.minBy({ x.size }))
```

The element it returns is the one in the array, so a change to it shows in the array:

```ant
type Item = { name: string, stock: int }
let items = [Item{ name: "nut", stock: 3 }, Item{ name: "bolt", stock: 12 }]
let item = items.maxBy({ x.stock })
if item != null {
  item.stock = 0
}
assert(items[1].stock == 0)
```

`key` runs once per element, from the first to the last, even after the greatest key is found:

```ant
let seen: Array[int] = []
[9, 1, 5].maxBy({ Array.push(seen, x); x }) // 9
seen                                        // [9, 1, 5]
```

## Possible issues

The result may be `null`, so reading a field of it without narrowing is a compile error:

```ant
type Person = { name: string, age: int }
let people = [Person{ name: "Ann", age: 31 }]
println(people.maxBy({ x.age }).name)
// error: Person | null has no field name: narrow the value first
```

Use `?.`, `??`, or a test against `null` first, as in the examples above.

The key must be ordered. A `bool` or a record is not, which is a compile error:

```ant
println([1, 2, 3].maxBy({ x > 1 }))
// error: K would be bool, which is not a member of its bound Ordered
```

Of equal keys, `Array.maxBy` keeps the first, while the last element of a stable ascending sort is
the last of them:

```ant
type Run = { name: string, ms: float64 }
let runs = [Run{ name: "a", ms: 12.5 }, Run{ name: "b", ms: 9.0 }, Run{ name: "c", ms: 12.5 }]
runs.maxBy({ x.ms })?.name         // "a"
runs.sortBy({ x.ms }).last()?.name // "c"
```

Negating an unsigned key to find the least panics, since the negation overflows. So does negating
the most negative value of a signed kind:

```ant
[uint8(1), uint8(2)].maxBy({ -x }) // panic: integer overflow
[Math.minInt, 0].maxBy({ -x })     // panic: integer overflow
```

Use `Array.minBy` instead.

When the elements may be `null`, a `null` result does not say whether `xs` was empty or a `null`
element won:

```ant
let a: Array[int | null] = [null]
a.maxBy({ x ?? 0 }) // null
let b: Array[int | null] = []
b.maxBy({ x ?? 0 }) // null
```

Test `Array.isEmpty` first, or find the index through `Array.entries`:

```ant
let a: Array[int | null] = [null]
a.isEmpty()                                     // false
Array.entries(a).maxBy({ x.value ?? 0 })?.index // 0
```

On floats, `maxBy({ x })` and `Array.max` can differ. `Array.max` is `NaN` when any element is `NaN`
and puts `-0.0` below `0.0`, while `Array.maxBy` never lets a `NaN` key win and keeps the first of
equal keys:

```ant
[1.0, 0.0 / 0.0].maxBy({ x }) // 1.0
[1.0, 0.0 / 0.0].max()        // NaN
[-0.0, 0.0].maxBy({ x })      // -0.0
[-0.0, 0.0].max()             // 0.0
```

## See also

[Array.minBy](minBy.md) · [Array.max](max.md) · [Array.min](min.md) · [Array.sortBy](sortBy.md) ·
[Array.find](find.md) · [Array.reduce](reduce.md) · [Array.entries](entries.md) ·
[?. operator](../../language/operators/optional-chaining.md) ·
[?? operator](../../language/operators/null-coalescing.md)
