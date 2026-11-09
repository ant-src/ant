<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Array.pop

```ant signature
func Array.pop[T](xs: Array[T]): T | null
```

Array.pop removes and returns the last element of xs, or returns null when xs is empty. It mutates
xs.

<!-- end generated -->

## Details

- Removes and returns the last element, changing `xs` in place. Every name that shares the array
  sees the shorter array.
- Returns `null` when `xs` is empty. When `T` includes `null`, that does not tell an empty array
  from a popped `null`; ask with [`Array.isEmpty`](isEmpty.md).

## Basic examples

Pop the last element, and the array is one shorter:

```ant
var xs = [10, 20, 30]
xs.pop() // 30
xs       // [10, 20]
```

An empty array gives `null`. Annotate it so that `T` is known:

```ant
var empty: Array[string] = []
Array.pop(empty) // null
empty            // []
```

Each call removes one more element, until none is left:

```ant
var tasks = ["write", "test"]
tasks.pop() // "test"
tasks.pop() // "write"
tasks.pop() // null
```

## Scope

An array of records gives the last record. Read a field of it with `?.`:

```ant
type Frame = { name: string, line: int }
var trace = [Frame{ name: "main", line: 3 }, Frame{ name: "parse", line: 41 }]
trace.pop()?.name // "parse"
trace             // [Frame{ name: "main", line: 3 }]
```

Elements of a union keep their tag, so `is` tells which member came off:

```ant
type Push = { value: int }
type Clear = {}
type Op = Push | Clear
var ops: Array[Op] = [Push{ value: 1 }, Clear{}]
ops.pop() is Clear // true
ops.pop()          // Push{ value: 1 }
```

A nested array gives its last inner array, which `?.[i]` indexes into:

```ant
var rows = [[1, 2], [3, 4, 5]]
rows.pop()?.[2] // 5
rows            // [[1, 2]]
```

An array of functions gives a function, which `?.(…)` calls:

```ant
var undo: Array[func(int): int] = [{ x + 1 }, { x * 10 }]
undo.pop()?.(4) // 40
undo.length()   // 1
```

An array bound by `let` can still lose elements:

```ant
let queue = ["a", "b"]
queue.pop() // "b"
queue       // ["a"]
```

Every name that shares the array sees the change, including the caller of a function that pops:

```ant
func dropLast(xs: Array[int]) {
  xs.pop()
}

var a = [1, 2, 3]
var b = a
dropLast(b)
println(a)
```

```text
[1, 2]
```

A copy made with spread is a separate array, which `Array.pop` leaves alone:

```ant
var original = [1, 2, 3]
var copy = [...original]
copy.pop() // 3
original   // [1, 2, 3]
```

In a pipe, the array is the first argument:

```ant
var words = "a b c".split(" ")
words | Array.pop() // "c"
words               // ["a", "b"]
```

## Applications

Use an array as a stack: `Array.push` adds to the top, and `Array.pop` takes from it. Check that
brackets match:

```ant
func balanced(s: string): bool {
  let pairs = Map{ ")": "(", "]": "[", "}": "{" }
  var open: Array[string] = []
  for c in s.graphemes() {
    if c == "(" || c == "[" || c == "{" {
      open.push(c)
    } else if pairs.has(c) {
      if open.pop() != pairs.get(c) {
        return false
      }
    }
  }
  open.isEmpty()
}

println(balanced("f(a[1], {b: 2})"))
println(balanced("(]"))
println(balanced("(("))
```

```text
true
false
false
```

Drain a stack with a `while` loop, and stop at `null` with `?? break`:

```ant
var pending = ["a.ant", "b.ant", "c.ant"]
while true {
  var file = pending.pop() ?? break
  println("check ${file}")
}
println(pending.length())
```

```text
check c.ant
check b.ant
check a.ant
0
```

Walk a tree depth first without recursion:

```ant
type Node = { name: string, children: Array[Node] }

var root = Node{
  name: "/",
  children: [Node{ name: "bin", children: [] }, Node{ name: "src", children: [Node{ name: "lib", children: [] }] }],
}

var stack = [root]
while true {
  var node = stack.pop() ?? break
  println(node.name)
  for child in node.children.reverse() {
    stack.push(child)
  }
}
```

```text
/
bin
src
lib
```

Undo the last edit, and report when there is nothing to undo:

```ant
type Edit = { text: string }

func undo(history: Array[Edit]): string {
  var last = history.pop() ?? return "nothing to undo"
  "undid ${last.text}"
}

var history = [Edit{ text: "insert a" }, Edit{ text: "delete b" }]
println(undo(history))
println(undo(history))
println(undo(history))
```

```text
undid delete b
undid insert a
nothing to undo
```

## Properties and relations

`Array.pop` returns what `Array.last` returned, and the length drops by one:

```ant
var xs = [4, 5, 6]
var before = xs.length()
var peek = xs.last()
assert(xs.pop() == peek)
assert(xs.length() == before - 1)
```

`Array.pop` undoes `Array.push`:

```ant
var xs = ["a", "b"]
xs.push("c")
assert(xs.pop() == "c")
assert(xs == ["a", "b"])
```

On a non-empty array, it agrees with `Array.removeAt` at the last index:

```ant
var a = [1, 2, 3]
var b = [1, 2, 3]
assert(a.pop() == b.removeAt(b.length() - 1))
assert(a == b)
```

What remains equals the slice without the last element:

```ant
var xs = [7, 8, 9]
var rest = xs.slice(0, xs.length() - 1)
xs.pop()
assert(xs == rest)
```

Popping every element gives them in reverse order:

```ant
var xs = [1, 2, 3]
var out: Array[int] = []
while true {
  out.push(xs.pop() ?? break)
}
assert(out == [3, 2, 1])
assert(xs.isEmpty())
```

## Possible issues

The result is `T | null`, not `T`, so using it as a `T` is a compile error:

```ant
var xs = [3, 1, 2]
println(xs.pop() + 1) // error: + needs both operands of one type, got int | null and int
```

Give a default with `??`, or bind it to a local and narrow that:

```ant
var xs = [3, 1, 2]
(xs.pop() ?? 0) + 1 // 3
```

Testing the result of one call and then popping again removes two elements, not one:

```ant
var xs = [1, 2, 3]
if xs.pop() != null {
  println(xs.pop())
}
println(xs)
```

```text
2
[1]
```

Bind the result to a local, and narrow that. The array loses one element:

```ant
var xs = [1, 2, 3]
var top = xs.pop()
if top != null {
  println(top + 1)
}
println(xs)
```

```text
4
[1, 2]
```

When the element type includes `null`, a `null` result does not tell an empty array from a popped
`null`. Ask [`Array.isEmpty`](isEmpty.md) first:

```ant
var answers: Array[string | null] = ["yes", null]
answers.pop()     // null
answers.isEmpty() // false
```

Popping one shared array from two goroutines without synchronization is a data race, a program error
with no guaranteed result. Guard each pop with a [Mutex](../Mutex/index.md):

```ant
var work = [1, 2, 3, 4]
var done: Array[int] = []
var mu = Mutex.new()
var wg = WaitGroup.new()
for _ in [1, 2] {
  WaitGroup.add(wg, 1)
  go {
    defer { WaitGroup.done(wg) }
    while true {
      var n = Mutex.withLock(mu, { work.pop() }) ?? break
      Mutex.withLock(mu, { done.push(n * n) })
    }
  }
}
WaitGroup.wait(wg)
println(done.sort())
println(work)
```

```text
[1, 4, 9, 16]
[]
```

`Array.pop` takes from the end. To take from the front, as a queue does, use
[`Array.removeAt`](removeAt.md) at index `0`, which panics on an empty array:

```ant
var line = ["first", "second"]
line.removeAt(0) // "first"
line             // ["second"]
```

## See also

[Array.push](push.md) · [Array.last](last.md) · [Array.removeAt](removeAt.md) ·
[Array.insert](insert.md) · [Array.isEmpty](isEmpty.md) · [Array.slice](slice.md) ·
[?? operator](../../language/operators/null-coalescing.md) ·
[?. operator](../../language/operators/optional-chaining.md) · [Null guide](../../guide/null.md)
