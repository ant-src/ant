<!-- status: reviewed -->

# `for`

`for x in e { ... }` runs its block once for each element of `e`: an array, a map, a set, a string,
`Bytes`, or a sequence function. The loop is a statement of type `void`.

## Details

- Each iteration binds a fresh `let`. A closure made in the body keeps that iteration's value, and
  assigning the loop variable is a compile error.
- A map is visited as `Entry` records and a set as its members, both in insertion order. A string
  yields each code point as a one-character string, and `Bytes` yields each `byte`.
- An array is read by index while the index is below its current length, so elements pushed during
  the loop are visited. Entries added to a map or a set during the loop are visited too.
- The left side may be a destructuring instead of a name. An array destructuring panics when an
  element has fewer elements than its fixed positions.
- `break` leaves the loop, `continue` goes on with the next element, and `return` or `try` in the
  body leave the function the loop is in.

## Basic examples

Visit each element of an array, in order:

```ant
let names = ["Ann", "Bob", "Cid"]
for name in names {
  println("Hello, ${name}")
}
```

```text
Hello, Ann
Hello, Bob
Hello, Cid
```

Destructure each `Entry` of a map into its key and value:

```ant
let ages = Map{ "Ann": 31, "Bob": 27 }
for { key, value } in ages {
  println("${key} is ${value}")
}
```

```text
Ann is 31
Bob is 27
```

Count with a range from `Iter.range`, which excludes its end:

```ant
for i in Iter.range(0, 3) {
  println(i)
}
```

```text
0
1
2
```

## Scope

### Collections

Without destructuring, a map yields whole `Entry` records:

```ant
let stock = Map{ tea: 3, rice: 0 }
for entry in stock {
  println(entry)
}
```

```text
Entry{ key: "tea", value: 3 }
Entry{ key: "rice", value: 0 }
```

A set yields its members in the order they were first added; a repeated value is a member once:

```ant
let tags = Set.of(["red", "blue", "red", "green"])
for tag in tags {
  println(tag)
}
```

```text
red
blue
green
```

A string yields one-character strings, one per code point:

```ant
for c in "añb" {
  println("[${c}]")
}
```

```text
[a]
[ñ]
[b]
```

`Bytes` yields each `byte`:

```ant
for b in "hi".toBytes() {
  println(b)
}
```

```text
104
105
```

An empty collection runs the block zero times:

```ant
let none: Array[int] = []
for n in none {
  println(n)
}
println("done")
```

```text
done
```

### Destructuring

A record destructuring binds the named fields; `{ field: name }` renames one, and fields not
mentioned are ignored:

```ant
type Point = { x: int, y: int, label: string }
let points = [Point{ x: 1, y: 2, label: "a" }, Point{ x: 3, y: 4, label: "b" }]
for { x: px, y: py } in points {
  println(px * py)
}
```

```text
2
12
```

An array destructuring binds by position. `_` discards one, and `...rest` binds what is left as a
new array:

```ant
let rows = [["ann", "31", "Oslo"], ["bob", "27", "Rome"]]
for [name, _, city] in rows {
  println("${name}: ${city}")
}
for [name, ...rest] in rows {
  println("${name} has ${rest.length()} more")
}
```

```text
ann: Oslo
bob: Rome
ann has 2 more
bob has 2 more
```

`Array.entries` pairs each element with its index, for a loop that needs both:

```ant
for { index, value } in ["tea", "rice"].entries() {
  println("${index + 1}. ${value}")
}
```

```text
1. tea
2. rice
```

Destructurings nest:

```ant
type Customer = { name: string, city: string }
type Order = { id: int, customer: Customer }
let orders = [Order{ id: 7, customer: Customer{ name: "Ann", city: "Oslo" } }]
for { id, customer: { city } } in orders {
  println("order ${id} ships to ${city}")
}
```

```text
order 7 ships to Oslo
```

### Sequences

A `Seq[T]` is a function that calls `yield` once per element. `for` runs the body at each call, and
`yield` returns `false` after a `break`, so the sequence can stop:

```ant
func countdown(from: int): Seq[int] {
  func(yield: func(int): bool) {
    var n = from
    while n > 0 {
      if !yield(n) {
        println("stopped at ${n}")
        return
      }
      n -= 1
    }
  }
}

for n in countdown(5) {
  if n == 3 { break }
  println(n)
}
```

```text
5
4
stopped at 3
```

The `Iter` functions build sequences lazily, so a pipe in the loop head computes only what the loop
visits:

```ant
for n in Iter.range(1, 1000) | Iter.filter({ x % 7 == 0 }) | Iter.take(3) {
  println(n)
}
```

```text
7
14
21
```

`Channel.values` yields what a channel receives until it is closed:

```ant
let jobs: Channel[string] = Channel.new(capacity: 2)
jobs.send("build")
jobs.send("test")
jobs.close()
for job in jobs.values() {
  println("running ${job}")
}
```

```text
running build
running test
```

### Closures

Each iteration has its own binding, so a closure made in the body keeps its own value:

```ant
let actions: Array[func(): string] = []
for name in ["Ann", "Bob"] {
  actions.push({ "hi ${name}" })
}
println(actions.map({ x() }))
```

```text
["hi Ann", "hi Bob"]
```

### Leaving the loop

`continue` skips to the next element and `break` ends the loop:

```ant
for n in [1, -2, 3, 0, 5] {
  if n < 0 { continue }
  if n == 0 { break }
  println(n)
}
```

```text
1
3
```

In nested loops, `break` and `continue` act on the innermost loop. Ant has no labels:

```ant
for row in [[1, 2, 0, 4], [5, 0, 6]] {
  for n in row {
    if n == 0 { break }
    println(n)
  }
  println("next row")
}
```

```text
1
2
next row
5
next row
```

To leave both loops, put them in a function and use `return`.

`return` in the body leaves the enclosing function, and the loop with it:

```ant
type Task = { name: string, done: bool }

func firstOpen(tasks: Array[Task]): string {
  for task in tasks {
    if !task.done { return task.name }
  }
  "all done"
}

println(firstOpen([Task{ name: "a", done: true }, Task{ name: "b", done: false }]))
println(firstOpen([]))
```

```text
b
all done
```

`Array.find` gives the same result; the loop pays off when the body does more than test.

## Applications

Parse every line and stop at the first failure. `try` in the body leaves the function from inside
the loop:

```ant
func parseLine(line: string): Result[int, Error] {
  let n = line.toInt() ?? return Result.fail("not a number: ${line}")
  Result.ok(n)
}

func parseAll(lines: Array[string]): Result[Array[int], Error] {
  let out: Array[int] = []
  for { index, value: line } in lines.entries() {
    out.push(try (parseLine(line) | Result.wrapErr("line ${index + 1}")))
  }
  Result.ok(out)
}

println(parseAll(["1", "2", "3"]))
println(parseAll(["1", "x", "3"]))
```

```text
Ok{ value: [1, 2, 3] }
Err{ error: Error{ message: "line 2: not a number: x" } }
```

Build a map of counts, updating one entry per element:

```ant
let words = "the cat and the hat".split(" ")
let counts: Map[string, int] = Map{}
for word in words {
  counts[word] = (counts.get(word) ?? 0) + 1
}
println(counts)
```

```text
Map{ "the": 2, "cat": 1, "and": 1, "hat": 1 }
```

Process a work queue that grows while the loop runs. Pushed elements are visited:

```ant
let children = Map{ root: ["src", "docs"], src: ["lib"] }
let queue = ["root"]
for name in queue {
  queue.extend(children.get(name) ?? [])
}
println(queue)
```

```text
["root", "src", "docs", "lib"]
```

## Properties and relations

A loop that pushes each element into a new array copies the array, in order:

```ant
let xs = [3, 1, 2]
let copy: Array[int] = []
for x in xs {
  copy.push(x)
}
assert(copy == xs)
```

A loop over a map visits the entries in the order `Map.entries` returns them:

```ant
let m = Map{ b: 2, a: 1, c: 3 }
let seen: Array[Entry[string, int]] = []
for e in m {
  seen.push(e)
}
assert(seen == m.entries())
```

A loop over a string visits as many one-character strings as `String.length` counts:

```ant
let s = "héllo"
var n = 0
for c in s {
  assert(c.length() == 1)
  n += 1
}
assert(n == s.length())
```

## Possible issues

The loop variable is a `let`. Assigning it would not change the array, so the checker rejects it:

```ant
let xs = [1, 2, 3]
for x in xs {
  x = x * 2 // error: cannot assign to x, a for variable; declare a var to hold a value that changes
}
```

To change the elements, use `Array.map` for a new array, or assign by index through `Array.entries`.

An array destructuring panics when an element is too short, as indexing past the end does:

```ant
let pairs = [[1, 2], [3]]
for [a, b] in pairs { // panic: a destructuring needs 2 elements, got 1
  println(a + b)
}
```

```text
3
```

`for` needs the collection itself, not a value that may be `null`. Handle `null` first:

```ant
func total(prices: Array[int] | null): int {
  var sum = 0
  for p in prices { // error: for needs an array, a map, a set, a string, Bytes, or a sequence, got Array[int] | null
    sum += p
  }
  sum
}
```

A guard such as `if prices == null { return 0 }` before the loop narrows `prices` to `Array[int]`.

A string yields code points, not the characters a reader sees. An accented letter written with a
combining mark is two iterations; `String.graphemes` gives what a reader sees as characters:

```ant
let word = "e\u{301}"
var steps = 0
for c in word {
  steps += 1
}
println(steps)
println(word.graphemes().length())
```

```text
2
1
```

Pushing to the array being visited, on every iteration, never ends the loop. Push only while a
condition holds, as the work queue in Applications does, or loop over a copy.

## See also

[while](while.md) · [break](break.md) · [continue](continue.md) · [let](let.md) · [var](var.md) ·
[Iter.range](../../lib/Iter/range.md) · [Seq](../../lib/Iter/types/Seq.md) ·
[Array.entries](../../lib/Array/entries.md) · [Map.entries](../../lib/Map/entries.md) ·
[Channel.values](../../lib/Channel/values.md) · [String.graphemes](../../lib/String/graphemes.md)
