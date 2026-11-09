<!-- status: reviewed -->

# Interfaces

An interface names a set of functions. Any type that has them satisfies the interface, with nothing
to declare: if the calls check, the type fits.

```ant
interface Shape {
  area(s: Self): float64
  name(s: Self): string
}

type Circle = { radius: float64 }

func Circle.area(c: Circle): float64 { 3.14 * c.radius * c.radius }
func Circle.name(c: Circle): string { "circle" }

type Rect = { width: float64, height: float64 }

func Rect.area(r: Rect): float64 { r.width * r.height }
func Rect.name(r: Rect): string { "rect" }

let shapes: Array[Shape] = [Circle{ radius: 1.0 }, Rect{ width: 2.0, height: 3.0 }]
for s in shapes {
  println("${s.name()}: ${s.area()}")
}
println(shapes.map({ x.area() }).sum())

interface Sized {
  length(s: Self): int
}

func longest(items: Array[Sized]): int {
  items.map({ x.length() }).max() ?? 0
}

println(longest(["ant", "bee", "wasp"]))
println(longest([[1, 2], "hello", [3]]))

interface Ranked {
  rank(a: Self, b: Self): int
}

type Job = { name: string, priority: int }

func Job.rank(a: Job, b: Job): int { b.priority - a.priority }

func best[T: Ranked](xs: Array[T]): T | null {
  xs.sortWith(func(a: T, b: T): int { a.rank(b) }).first()
}

let jobs = [Job{ name: "deploy", priority: 2 }, Job{ name: "fix", priority: 5 }]
println(best(jobs)?.name)
```

```text
circle: 3.14
rect: 6.0
9.14
4
5
fix
```

`Shape` lists two signatures. The first parameter of each is the receiver, and its type is `Self`:
the type that satisfies the interface. `Circle` and `Rect` each declare `area` and `name` as
functions on the type, so both are `Shape`s, and one `Array[Shape]` holds them side by side. In the
loop, `s.name()` and `s.area()` call the function of whatever type `s` holds. An interface value is
the value itself, with no wrapper around it.

Built-in types satisfy interfaces too. A `string` and an `Array` both have `length`, so `longest`
takes a mix of them as `Sized` values.

`Ranked` mentions `Self` twice: `rank` compares two values of the same type. Such an interface is
not a type, since an array of it could hold a `Job` and something else that `rank` cannot compare.
It is a bound instead. `best[T: Ranked]` takes an array of any one type that has `rank`, and returns
that type, so `best(jobs)` is a `Job | null` and `.name` is there to read.

## Try it

Drop `name` from a type, and it no longer fits `Shape`. An `Array[Circle]` is not an `Array[Shape]`;
spread it into a new array to widen it. Use `Ranked` as a type, and the type checker says it is a
bound only:

```ant
interface Shape {
  area(s: Self): float64
  name(s: Self): string
}

type Square = { side: float64 }

func Square.area(s: Square): float64 { s.side * s.side }

let shapes: Array[Shape] = [Square{ side: 2.0 }] // error: cannot use Square where Shape is expected

type Circle = { radius: float64 }

func Circle.area(c: Circle): float64 { 3.14 * c.radius * c.radius }
func Circle.name(c: Circle): string { "circle" }

let circles = [Circle{ radius: 1.0 }]
let all: Array[Shape] = circles // error: cannot use Array[Circle] where Array[Shape] is expected
let copy: Array[Shape] = [...circles]

interface Ranked {
  rank(a: Self, b: Self): int
}

let queue: Array[Ranked] = []
// error: Ranked mentions Self beyond its receiver, so it is a bound only and not a type
```

A union lists a closed set of types, and `is` tells them apart. An interface is an open set: any
type with the right functions joins it, and you call it through `.`. The
[interfaces chapter](../guide/interfaces.md) of the guide explains when to pick which. The reference
pages are [`interface`](../language/keywords/interface.md), [`Self`](../language/types/Self.md), and
[calls through `.`](../language/operators/dot-calls.md).

Next: [Goroutines](17-goroutines.md).
