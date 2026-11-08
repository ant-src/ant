<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Array.push

```ant signature
func Array.push[T](xs: Array[T], v: T): void
```

Array.push appends v to the end of xs. It mutates xs.

<!-- end generated -->

## Details

- `Array.push` mutates `xs` and returns `void`
  ([Standard library](https://github.com/ant-src/spec/blob/master/ant-lang.md#11-standard-library)).
- `xs` stays the same array, so every other name for it sees the new element
  ([Values](https://github.com/ant-src/spec/blob/master/ant-lang.md#values)).
- A `let` fixes the binding, not the elements, so `Array.push` works on a `let` array
  ([Variables](https://github.com/ant-src/spec/blob/master/ant-lang.md#3-variables)).
- A `for` loop over `xs` visits elements pushed during the loop
  ([Expressions](https://github.com/ant-src/spec/blob/master/ant-lang.md#expressions)).
- The parameters are named `xs` and `v`, so a call can name its arguments in any order
  ([Calling](https://github.com/ant-src/spec/blob/master/ant-lang.md#calling)).
- Two goroutines that push to one array without a lock are a data race
  ([Data races](https://github.com/ant-src/spec/blob/master/ant-lang.md#data-races)).

## Basic examples

Append a value to the end of an array:

```ant
var xs = [1, 2]
Array.push(xs, 3)
xs // [1, 2, 3]
```

Grow an empty array one element at a time. An empty literal needs its element type:

```ant
var names: Array[string] = []
Array.push(names, "Ada")
Array.push(names, "Linus")
names // ["Ada", "Linus"]
```

## Scope

Records, `null` in an array that allows it, and arrays inside arrays push alike:

```ant
type Point = { x: int, y: int }
var points = [Point{ x: 1, y: 2 }]
Array.push(points, Point{ x: 3, y: 4 })
points // [Point{ x: 1, y: 2 }, Point{ x: 3, y: 4 }]
var maybe: Array[int | null] = [1]
Array.push(maybe, null)
maybe // [1, null]
var rows = [[1, 2]]
Array.push(rows, [3])
rows // [[1, 2], [3]]
```

A `let` array can be pushed to, since the binding stays the same:

```ant
let stack = ["a"]
Array.push(stack, "b")
stack // ["a", "b"]
```

Call it through `.` on an array, or with a pipe, which passes its left side as the first argument:

```ant
var xs = [1]
xs.push(2)
xs | Array.push(3)
xs // [1, 2, 3]
```

Name the arguments, in any order:

```ant
var xs = [1]
Array.push(v: 2, xs: xs)
xs // [1, 2]
```

## Applications

Collect the results of a loop that keeps state, here the running total:

```ant
var totals: Array[int] = []
var sum = 0
for n in [3, 1, 4, 1, 5] {
  sum += n
  Array.push(totals, sum)
}
totals // [3, 4, 8, 9, 14]
```

Use an array as a stack with [Array.pop](pop.md), here to check that brackets balance:

```ant
func balanced(s: string): bool {
  var open: Array[string] = []
  for c in s {
    if c == "(" || c == "[" {
      Array.push(open, c)
    } else if c == ")" || c == "]" {
      var want = if c == ")" { "(" } else { "[" }
      if Array.pop(open) != want {
        return false
      }
    }
  }
  Array.isEmpty(open)
}

balanced("([]())") // true
balanced("([)]")   // false
balanced("((")     // false
```

Visit a tree breadth-first. The `for` loop reaches the nodes pushed while it runs:

```ant
type Node = { name: string, children: Array[Node] }

var tree = Node{
  name: "root",
  children: [
    Node{ name: "a", children: [Node{ name: "a1", children: [] }] },
    Node{ name: "b", children: [] },
  ],
}
var queue = [tree]
for node in queue {
  println(node.name)
  for child in node.children {
    Array.push(queue, child)
  }
}
```

```text
root
a
b
a1
```

Push from several goroutines under a [Mutex](../Mutex/index.md), which keeps the pushes from racing:

```ant
var results: Array[int] = []
var mu = Mutex.new()
var wg = WaitGroup.new()
for n in [1, 2, 3, 4] {
  WaitGroup.add(wg, 1)
  go {
    defer { WaitGroup.done(wg) }
    Mutex.withLock(mu, { Array.push(results, n * n) })
  }
}
WaitGroup.wait(wg)
println(Array.sort(results))
```

```text
[1, 4, 9, 16]
```

## Properties and relations

The length grows by one, and the last element is `v`:

```ant
var xs = [1, 2]
var before = len(xs)
Array.push(xs, 7)
assert(len(xs) == before + 1)
assert(Array.last(xs) == 7)
```

[Array.pop](pop.md) undoes it:

```ant
var xs = [1, 2]
Array.push(xs, 3)
assert(Array.pop(xs) == 3)
assert(xs == [1, 2])
```

It does in place what the spread literal `[...xs, v]` does in a new array:

```ant
var xs = [1, 2]
var copy = [...xs, 3]
Array.push(xs, 3)
assert(xs == copy)
```

It is [Array.insert](insert.md) at the length, and [Array.extend](extend.md) with one element:

```ant
var a = [1]
var b = [1]
var c = [1]
Array.push(a, 2)
Array.insert(b, len(b), 2)
Array.extend(c, [2])
assert(a == b && b == c)
```

Every other name for the array sees the push:

```ant
var xs = [1]
var alias = xs
Array.push(xs, 2)
assert(Array.same(alias, xs))
alias // [1, 2]
```

## Possible issues

`Array.push` returns `void`, not the array, so its result cannot be assigned back:

<!-- test: error cannot use void where Array[int] is expected -->

```ant
var xs = [1]
xs = Array.push(xs, 2)
```

To get a new array and keep the old one, use the spread literal:

```ant
var xs = [1]
var ys = [...xs, 2]
xs // [1]
ys // [1, 2]
```

`v` must have the element type. An array is one value, so pushing it adds one element; to append its
elements, use [Array.extend](extend.md):

<!-- test: error cannot use Array[int] where int is expected -->

```ant
var xs = [1]
Array.push(xs, [2, 3])
```

```ant
var xs = [1]
Array.extend(xs, [2, 3])
xs // [1, 2, 3]
```

An empty literal has no element type to infer, so annotate the variable:

<!-- test: error cannot infer the element type of an empty array -->

```ant
var xs = []
Array.push(xs, 1)
```

The value is pushed, not copied. An array or a record pushed in is the same value in the array, so a
later change to it shows through both:

```ant
var row = [0]
var grid: Array[Array[int]] = []
Array.push(grid, row)
row[0] = 9
grid // [[9]]
```

A `for` loop over the array visits what the body pushes, so a loop that always pushes never ends.
Loop over a copy, or push to another array:

```ant
var xs = [1, 2]
for x in [...xs] {
  Array.push(xs, x * 10)
}
xs // [1, 2, 10, 20]
```

A function that takes an array and a callback reads the elements once, when it is called, while a
`for` loop reads the array as it goes. A callback that pushes to the array it was given does not see
its own pushes
([Standard library](https://github.com/ant-src/spec/blob/master/ant-lang.md#11-standard-library)):

```ant
var a = [1, 2, 3]
Array.map(a, { Array.push(a, x + 10); x * 2 }) // [2, 4, 6]
a                                              // [1, 2, 3, 11, 12, 13]
```

## See also

[Array.pop](pop.md), [Array.extend](extend.md), [Array.insert](insert.md),
[Array.concat](concat.md), [Array.last](last.md), [Array.same](same.md)
