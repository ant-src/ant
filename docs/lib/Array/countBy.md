<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Array.countBy

```ant signature
func Array.countBy[T, K](xs: Array[T], key: func(T): K): Map[K, int]
```

Array.countBy returns a map from each key to how many elements of xs have that key. Keys are in
first-seen order.

<!-- end generated -->

## Details

- Keys are in the order they first appear. An empty `xs` gives an empty map.
- Calls `key` once per element, from first to last.
- Keys compare by `==`, so a record or an array can be a key.
- Gives the same map as `xs.groupBy(key).map({ len(x.value) })`.
- Returns a new map; `xs` is unchanged.

## Basic examples

Count the words of each length. The key `3` comes first because `"ant"` is the first word:

```ant
Array.countBy(["ant", "bee", "wasp"], { x.length() }) // Map{ 3: 2, 4: 1 }
```

Call it through `.` on the array, with `{ x }` as the key, to count each distinct element:

```ant
["the", "ant", "and", "the", "bee"].countBy({ x }) // Map{ "the": 2, "ant": 1, "and": 1, "bee": 1 }
```

## Scope

An empty array gives an empty map. Annotate it so that `T` is known:

```ant
let empty: Array[int] = []
empty.countBy({ x }) // Map{}
```

A named function works as well as a lambda, and so does a function of the prelude:

```ant
func parity(n: int): string {
  if n % 2 == 0 { "even" } else { "odd" }
}

Array.range(1, 6).countBy(parity)                                    // Map{ "odd": 3, "even": 2 }
["ant", "bee", "wasp", "moth", "fly"] | Array.countBy(String.length) // Map{ 3: 3, 4: 2 }
```

Count records by a field:

```ant
type Person = { name: string, team: string }
let people = [
  Person{ name: "Ann", team: "red" },
  Person{ name: "Bob", team: "blue" },
  Person{ name: "Cid", team: "red" },
]
people.countBy({ x.team }) // Map{ "red": 2, "blue": 1 }
```

A record or an array can be the key, since keys compare by `==`:

```ant
[[1, 2], [3], [1, 2]].countBy({ x }) // Map{ [1, 2]: 2, [3]: 1 }
```

`null` is a key like any other, so counting on `x == null` tells missing values from present ones:

```ant
let ages: Array[int | null] = [30, null, 25, null, null]
ages.countBy({ x == null }) // Map{ false: 2, true: 3 }
```

Count the characters of a string by splitting it into graphemes first:

```ant
"🐜a🐜b".graphemes().countBy({ x }) // Map{ "🐜": 2, "a": 1, "b": 1 }
```

## Applications

Count failed logins per address in a log, keeping only lines that end in `401`:

```ant
let log = "203.0.113.9 GET /login 401
192.0.2.77 GET /login 401
203.0.113.9 GET / 200
203.0.113.9 GET /login 401
192.0.2.77 GET /login 401
203.0.113.9 GET /login 401"
log.lines()
  | Array.filter({ x.endsWith(" 401") })
  | Array.countBy({ x.fields()[0] }) // Map{ "203.0.113.9": 3, "192.0.2.77": 2 }
```

Print the most common words first. Sort the entries by their count, descending; words with equal
counts keep their first-seen order, since `Array.sortBy` is stable:

```ant
let text = "the ant and the bee and the wasp"
for { key, value } in text.fields().countBy({ x }).entries().sortBy({ -x.value }) {
  println("${key}: ${value}")
}
```

```text
the: 3
and: 2
ant: 1
bee: 1
wasp: 1
```

Build a histogram of scores in buckets of ten:

```ant
let scores = [72, 85, 91, 78, 64, 88, 95]
let buckets = scores.countBy({ x / 10 * 10 })
for { key, value } in buckets.entries().sortBy({ x.key }) {
  println("${key}: ${"#".repeat(value)}")
}
```

```text
60: #
70: ##
80: ##
90: ##
```

## Properties and relations

The counts add up to the length of `xs`:

```ant
let xs = [5, 3, 8, 1, 3]
let counts = xs.countBy({ x % 2 })
assert((counts.values().sum() ?? 0) == len(xs))
```

Each count is the length of the group `Array.groupBy` builds, and the number `Array.countWhere`
finds for that key:

```ant
let xs = [1, 2, 3, 4, 5, 6, 7]
let key = func(n: int): int { n % 3 }
let counts = xs.countBy(key)
assert(counts == xs.groupBy(key).map({ len(x.value) }))
for { key: k, value } in counts.entries() {
  assert(value == xs.countWhere({ key(x) == k }))
}
```

The keys are the distinct keys of the elements, in first-seen order:

```ant
let xs = ["bee", "ant", "wasp", "fly"]
assert(xs.countBy(String.length).keys() == xs.map(String.length).unique())
```

`key` runs once per element, in the order of `xs`, so a side effect in it sees every element:

```ant
let seen: Array[int] = []
[3, 1, 4, 1].countBy({ seen.push(x); x % 2 }) // Map{ 1: 3, 0: 1 }
seen                                          // [3, 1, 4, 1]
```

## Possible issues

Indexing the result reads `int | null`, because a key may never occur. Supply `0` with `??`:

```ant
let counts = ["a", "b", "a"].countBy({ x })
counts["a"] ?? 0 // 2
counts["z"] ?? 0 // 0
```

The keys are in first-seen order, not by count. Sort the entries when the order matters, as in
Applications above.

`NaN` is not `==` to itself, so every `NaN` key gets a count of its own:

```ant
[1.0, Math.NaN, Math.NaN].countBy({ x }) // Map{ 1.0: 1, NaN: 1, NaN: 1 }
```

A key is hashed when the map gets it. When `key` returns an element itself, such as an array,
mutating that element afterwards mutates the key, and its count may become unreachable. Return a
copy, as in `{ [...x] }`. The map holds no elements, so the copy protects it fully:

```ant
let rows = [[1], [2], [1]]
let counts = rows.countBy({ [...x] })
rows[0].push(9)
counts[[1]] // 2
```

To count the elements that pass one test, `Array.countWhere` returns an `int` and needs no map:

```ant
let xs = [1, 2, 3, 4]
xs.countWhere({ x % 2 == 0 })    // 2
xs.countBy({ x % 2 == 0 })[true] // 2
```

## See also

[Array.groupBy](groupBy.md) · [Array.countWhere](countWhere.md) · [Array.uniqueBy](uniqueBy.md) ·
[Array.sortBy](sortBy.md) · [Map.entries](../Map/entries.md) · [Map.keys](../Map/keys.md)
