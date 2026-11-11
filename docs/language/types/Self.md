<!-- status: reviewed -->

# Self

`Self` names the type that satisfies an interface. It is written only in the signatures of an
[`interface`](../keywords/interface.md) declaration, where each signature takes a `Self` as its
receiver.

## Details

- `Self` has no other use. In a function on a type, a record type, or an annotation, it is a compile
  error: write the type's name there.
- The first parameter of every signature has the type `Self`, exactly. `Self | null` there is a
  compile error.
- A type `S` satisfies the interface when each of its functions, with `Self` replaced by `S`, is
  assignable to the interface's signature.
- An interface that mentions `Self` only as the receiver is a type. One that mentions it anywhere
  else, as another parameter, a result, or inside a type such as `Self | null` or `Array[Self]`, is
  only a bound, written `[T: I]`.

## Basic examples

The receiver of each signature is a `Self`. `Circle` satisfies `Shape` because `Circle.area` has the
signature with `Self` replaced by `Circle`:

```ant
interface Shape {
  area(s: Self): float64
}

type Circle = { radius: float64 }

func Circle.area(c: Circle): float64 { 3.0 * c.radius * c.radius }

let s: Shape = Circle{ radius: 2.0 }
println(s.area())
```

```text
12.0
```

A `Self` in the result promises a value of the same type as the receiver. Such an interface is a
bound, and the result keeps the caller's type:

```ant
interface Grow {
  grow(g: Self): Self
}

type Circle = { radius: float64 }

func Circle.grow(c: Circle): Circle { Circle{ radius: c.radius * 2.0 } }

func twice[T: Grow](x: T): T { x.grow().grow() }

let big = twice(Circle{ radius: 1.0 })
println(big.radius)
```

```text
4.0
```

## Scope

### Several parameters of type `Self`

A signature may take more than one `Self`. Both are the same type, so `merge` never mixes a `Range`
with another type:

```ant
interface Merge {
  merge(a: Self, b: Self): Self
}

type Range = { lo: int, hi: int }

func Range.merge(a: Range, b: Range): Range {
  Range{ lo: Math.min(a.lo, b.lo), hi: Math.max(a.hi, b.hi) }
}

func mergeAll[T: Merge](first: T, rest: Array[T]): T {
  rest.reduce(first, func(acc: T, x: T): T { acc.merge(x) })
}

let span = mergeAll(Range{ lo: 3, hi: 4 }, [Range{ lo: 1, hi: 2 }, Range{ lo: 5, hi: 9 }])
println(span)
```

```text
Range{ lo: 1, hi: 9 }
```

### `Self` inside other types

`Self` may appear inside a union, an array, or any other type of a signature:

```ant
interface Linked {
  next(n: Self): Self | null
}

type Step = { name: string, after: Step | null = null }

func Step.next(s: Step): Step | null { s.after }

func walk[T: Linked](start: T): Array[T] {
  let out = [start]
  var cur = start.next()
  while cur != null {
    out.push(cur)
    cur = cur.next()
  }
  out
}

let steps = walk(Step{ name: "fetch", after: Step{ name: "build", after: Step{ name: "test" } } })
println(steps.map({ x.name }))
```

```text
["fetch", "build", "test"]
```

### Built-in types

`Self` stands for a built-in type too. `string` and every `Array` have `slice` and `length`, so both
satisfy `Sliceable`, and `firstHalf` returns the type it was given:

```ant
interface Sliceable {
  slice(s: Self, from: int, to: int): Self
  length(s: Self): int
}

func firstHalf[T: Sliceable](x: T): T { x.slice(0, x.length() / 2) }

println(firstHalf("antelope"))
println(firstHalf([1, 2, 3, 4]))
```

```text
ante
[1, 2]
```

### Generic interfaces

In a generic interface, `Self` is the satisfying type and the type parameters are the rest:

```ant
interface Collect[T] {
  put(c: Self, v: T): Self
  items(c: Self): Array[T]
}

type Bag = { xs: Array[int] }

func Bag.put(b: Bag, v: int): Bag { Bag{ xs: [...b.xs, v] } }
func Bag.items(b: Bag): Array[int] { b.xs }

func fill[C: Collect[int]](c: C, vs: Array[int]): C {
  var out = c
  for v in vs {
    out = out.put(v)
  }
  out
}

println(fill(Bag{ xs: [] }, [1, 2, 3]).items())
```

```text
[1, 2, 3]
```

### `Self` or the interface's name

A result typed with the interface's name keeps the interface a type, but the caller gets back a
`Shape`, not the type it passed. Write `Self` when the result must keep that type, and the
interface's name when any satisfying value will do:

```ant
interface Shape {
  area(s: Self): float64
  scaled(s: Self, by: float64): Shape
}

type Square = { side: float64 }

func Square.area(s: Square): float64 { s.side * s.side }
func Square.scaled(s: Square, by: float64): Shape { Square{ side: s.side * by } }

let shapes: Array[Shape] = [Square{ side: 1.0 }, Square{ side: 2.0 }]
println(shapes.map({ x.scaled(3.0).area() }))
```

```text
[9.0, 36.0]
```

## Properties and relations

A call through a `Self` bound is the call on the type itself, and its result has that type:

```ant
interface Merge {
  merge(a: Self, b: Self): Self
}

type Range = { lo: int, hi: int }

func Range.merge(a: Range, b: Range): Range {
  Range{ lo: Math.min(a.lo, b.lo), hi: Math.max(a.hi, b.hi) }
}

func both[T: Merge](a: T, b: T): T { a.merge(b) }

let a = Range{ lo: 1, hi: 3 }
let b = Range{ lo: 2, hi: 5 }
assert(both(a, b) == Range.merge(a, b))
assert(both(a, b).hi == 5)
```

## Possible issues

### `Self` outside an interface

A function on a type names its own type. `Self` there is a compile error:

```ant
type Circle = { radius: float64 }

func Circle.grow(c: Self): Circle { c } // error: 'Self' is a reserved word
```

A record cannot refer to itself as `Self` either. Use its name:

```ant
type Node = { next: Self | null } // error: 'Self' is a reserved word
```

```ant
type Node = { value: int, next: Node | null = null }

let list = Node{ value: 1, next: Node{ value: 2 } }
list.next?.value // 2
```

`Self` cannot name a variable:

```ant
let Self = 1 // error: 'Self' cannot be a variable name
```

### The receiver is exactly `Self`

```ant
interface Shape {
  area(s: Self | null): float64 // error: the first parameter of area must have the type Self, the satisfying type
}
```

### An interface with `Self` beyond the receiver is not a type

```ant
interface Merge {
  merge(a: Self, b: Self): Self
}

type Range = { lo: int, hi: int }

func Range.merge(a: Range, b: Range): Range { a }

let m: Merge = Range{ lo: 1, hi: 2 }
// error: Merge mentions Self beyond its receiver, so it is a bound only and not a type
```

Take it as a bound, `[T: Merge]`, as in
[Several parameters of type `Self`](#several-parameters-of-type-self).

### A union does not solve a `Self` bound

Such an interface is not a type, so a union cannot stand in for `T`, even when each member satisfies
it. Narrow first:

```ant
interface Merge {
  merge(a: Self, b: Self): Self
}

type Range = { lo: int, hi: int }
type Span = { lo: int, hi: int }

func Range.merge(a: Range, b: Range): Range { a }
func Span.merge(a: Span, b: Span): Span { a }

func twice[T: Merge](x: T): T { x.merge(x) }

let r: Range | Span = Range{ lo: 1, hi: 2 }
let m = twice(r) // error: T would be Range | Span, which is not assignable to its bound Merge; narrow first
```

### Every `Self` becomes the same type

A function whose second parameter takes another type does not match `merge(a: Self, b: Self)`:

```ant
interface Merge {
  merge(a: Self, b: Self): Self
}

type Range = { lo: int, hi: int }
type Point = { x: int }

func Point.merge(a: Point, b: Range): Point { a }

func twice[T: Merge](x: T): T { x.merge(x) }

let p = twice(Point{ x: 1 }) // error: T would be Point, which is not assignable to its bound Merge
```

## See also

[interface](../keywords/interface.md) · [type](../keywords/type.md) · [func](../keywords/func.md) ·
[calls through `.`](../operators/dot-calls.md) · [Interfaces guide](../../guide/interfaces.md) ·
[Generics guide](../../guide/generics.md) · [Tour: Interfaces](../../tour/16-interfaces.md)
