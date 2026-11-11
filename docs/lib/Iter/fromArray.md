<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Iter.fromArray

```ant signature
func Iter.fromArray[T](xs: Array[T]): Seq[T]
```

Iter.fromArray returns a sequence of the elements of xs, in order.

<!-- end generated -->

## Details

- Building the sequence runs nothing. The `Iter` functions you chain on it, such as
  [`Iter.map`](map.md), run only when [`Iter.toArray`](toArray.md) or `for` runs the result, and
  only for as many elements as are taken.
- The sequence reads `xs` each time it runs, not when you build it. It sees elements set or pushed
  before the run, and elements pushed during the run.

## Basic examples

Turn an array into a sequence and back:

```ant
Iter.fromArray([1, 2, 3]).toArray() // [1, 2, 3]
```

Chain lazy steps on it:

```ant
Iter.fromArray(["ant", "bee", "cat"]).map({ x.toUpper() }).toArray() // ["ANT", "BEE", "CAT"]
```

Take the first elements of an array as a sequence:

```ant
Iter.fromArray([5, 6, 7, 8]).take(2).toArray() // [5, 6]
```

## Scope

An empty array gives an empty sequence:

```ant
let empty: Array[string] = []
Iter.fromArray(empty).toArray() // []
```

The elements may be records:

```ant
type Point = { x: int, y: int }
let points = [Point{ x: 1, y: 2 }, Point{ x: 3, y: 4 }]
Iter.fromArray(points).map({ x.x + x.y }).toArray() // [3, 7]
```

`null` elements pass through like any other:

```ant
let marks: Array[int | null] = [90, null, 75]
Iter.fromArray(marks).filter({ x != null }).toArray() // [90, 75]
```

Nested arrays stay arrays; each element of the sequence is one inner array:

```ant
Iter.fromArray([[1, 2], [], [3]]).map({ x.length() }).toArray() // [2, 0, 1]
```

Start a pipe with it:

```ant
[4, 5, 6] | Iter.fromArray() | Iter.filter({ x % 2 == 0 }) | Iter.toArray() // [4, 6]
```

## Applications

Write a function once over `Seq[int]` and feed it an array or a range:

```ant
func total(s: Seq[int]): int {
  var sum = 0
  for n in s {
    sum += n
  }
  sum
}

println(total(Iter.fromArray([10, 20, 30])))
println(total(Iter.range(1, 5)))
```

```text
60
10
```

Find the first few results of an expensive step without running it on every element. The counter
shows `map` ran on two elements, not five:

```ant
var calls = 0
let big = Iter.fromArray([3, 8, 1, 20, 7])
  .map({
    calls += 1
    x * 10
  })
  .filter({ x > 50 })
  .take(1)
  .toArray()
println(big)
println(calls)
```

```text
[80]
2
```

Stop a loop over the sequence early with `break`:

```ant
let lines = ["# title", "text", "", "more"]
for line in Iter.fromArray(lines) {
  if line == "" {
    break
  }
  println(line)
}
```

```text
# title
text
```

## Properties and relations

[`Iter.toArray`](toArray.md) undoes it, giving an array equal to `xs`:

```ant
let xs = [3, 1, 2]
assert(Iter.fromArray(xs).toArray() == xs)
```

Running the same sequence twice gives the same elements:

```ant
let s = Iter.fromArray(["a", "b"])
assert(s.toArray() == s.toArray())
```

Mapping lazily and collecting gives what [`Array.map`](../Array/map.md) gives eagerly:

```ant
let xs = [1, 2, 3]
assert(Iter.fromArray(xs).map({ x * x }).toArray() == xs.map({ x * x }))
```

## Possible issues

An array is not a `Seq`. Passing one where a sequence is expected is a compile error; wrap it with
`Iter.fromArray`:

```ant
let xs = [1, 2, 3]
println(Iter.take(xs, 2)) // error: cannot use Array[int] where Seq[T] is expected
```

An array's `.` reaches `Array` functions, not `Iter` ones, so `xs.fromArray()` is a compile error.
Call `Iter.fromArray(xs)` or pipe into it:

```ant
let xs = [4, 5, 6]
println(xs.fromArray().toArray()) // error: Array[int] has no field or function fromArray
```

The sequence does not copy `xs`. Changes to the array before the run show up in it:

```ant
let xs = [1, 2, 3]
let s = Iter.fromArray(xs)
xs.push(4)
s.toArray() // [1, 2, 3, 4]
```

Elements pushed during the run are visited too, unlike [`Array.map`](../Array/map.md), which visits
only the elements present when it was called:

```ant
let ys = [1, 2]
for n in Iter.fromArray(ys) {
  if n < 3 {
    ys.push(n + 2)
  }
  println(n)
}
```

```text
1
2
3
4
```

`for` iterates over an array directly, so a loop needs no `Iter.fromArray`:

```ant
for n in [1, 2] {
  println(n)
}
```

```text
1
2
```

`.filter` on an array is [`Array.filter`](../Array/filter.md), which runs at once and returns an
array; on the sequence it is [`Iter.filter`](filter.md), which runs nothing until the sequence is
run:

```ant
[1, 2, 3, 4].filter({ x > 2 })                           // [3, 4]
Iter.fromArray([1, 2, 3, 4]).filter({ x > 2 }).toArray() // [3, 4]
```

## See also

[Iter.toArray](toArray.md) · [Iter.range](range.md) · [Iter.map](map.md) · [Iter.filter](filter.md)
· [Iter.take](take.md) · [Iter](index.md) · [for](../../language/keywords/for.md)
