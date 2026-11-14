<!-- status: reviewed -->

# Interfaces

An interface names a set of functions. A type that has those functions satisfies the interface, and
nothing declares that it does. Use an interface when code must work with types it does not know in
advance; use a union when the set of variants is fixed.

## Declaring an interface

```ant
interface Shape {
  area(s: Self): float64
  name(s: Self): string
}

interface Container[T] {
  get(c: Self, i: int): T | null
  size(c: Self): int
}
```

The body lists signatures, one per line, written like function types with parameter names. The first
parameter of each is the receiver, and its type is `Self`: the type that satisfies the interface. An
interface may take type parameters, as `Container[T]` does, but its signatures do not declare their
own.

Declare interfaces at the top level of a module. Add `export` to use one from other modules.
Satisfying is checked where the value becomes the interface: a type whose `area` is not exported is
a `Shape` inside its module, but not in a module that imports it. Export the functions an interface
needs.

## Satisfying an interface

A type satisfies an interface when every call the interface promises checks on it. For a record or a
named primitive type, that means a function on the type for each signature:

```ant
interface Shape {
  area(s: Self): float64
  name(s: Self): string
}

type Circle = { radius: float64 }

func Circle.area(c: Circle): float64 { 3.0 * c.radius * c.radius }
func Circle.name(c: Circle): string { "circle" }

type Rect = { width: float64, height: float64 }

func Rect.area(r: Rect): float64 { r.width * r.height }
func Rect.name(r: Rect): string { "rect" }

let shapes: Array[Shape] = [Circle{ radius: 1.0 }, Rect{ width: 2.0, height: 3.0 }]
for s in shapes {
  println("${s.name()}: ${s.area()}")
}
```

```text
circle: 3.0
rect: 6.0
```

`Circle` and `Rect` never mention `Shape`. They declare `area` and `name`, so both are `Shape`s, and
one array holds them side by side. `s.area()` calls the function of whatever type `s` holds.

Built-in types satisfy interfaces by the same rule. A `string`, an `Array`, and `Bytes` each have
`length`, so all three are `Sized`. A generic function on a generic type covers every instantiation:

```ant
interface Sized {
  length(s: Self): int
}

interface Show {
  show(s: Self): string
}

type Box[T] = { value: T }

func Box.show[T](b: Box[T]): string { "Box(${b.value})" }

let items: Array[Sized] = ["ant", [1, 2], "hi".toBytes()]
println(items.map({ x.length() }))

let boxes: Array[Show] = [Box{ value: 1 }, Box{ value: "a" }]
println(boxes.map({ x.show() }))
```

```text
[3, 2, 2]
["Box(1)", "Box(a)"]
```

A generic interface works the same way. `Stack[T]` declares `get` and `size`, so a `Stack[int]` is a
`Container[int]`:

```ant
interface Container[T] {
  get(c: Self, i: int): T | null
  size(c: Self): int
}

type Stack[T] = { items: Array[T] }

func Stack.get[T](s: Stack[T], i: int): T | null { s.items.get(i) }
func Stack.size[T](s: Stack[T]): int { s.items.length() }

func describe(c: Container[int]): string { "${c.size()} items, first ${c.get(0)}" }

println(describe(Stack{ items: [1, 2] }))
```

```text
2 items, first 1
```

When a type misses a function, the compile error names it. A field that holds a function does not
count: the interface needs a function declared on the type.

```ant
interface Shape {
  area(s: Self): float64
}

type Rect = { width: float64, height: float64 }
type Fixed = { area: func(): float64 }

let r: Shape = Rect{ width: 1.0, height: 2.0 }
// error: Rect has no function area of type func(Rect): float64
let f: Shape = Fixed{ area: { 1.0 } }
// error: Fixed has a field area, not a function declared on it; the interface needs func Fixed.area
```

## Interface values

An interface whose signatures mention `Self` only as the receiver is a type. A value of the
interface is the value itself, with no wrapper: `==`, `String.from`, and `Json.stringify` see the
`Circle` or the `Rect` inside.

<!-- test: before
interface Shape {
  area(s: Self): float64
}
type Circle = { radius: float64 }
type Square = { side: float64 }
func Circle.area(c: Circle): float64 { 3.0 * c.radius * c.radius }
func Square.area(s: Square): float64 { s.side * s.side }
-->

```ant
let shapes: Array[Shape] = [Circle{ radius: 1.0 }, Square{ side: 3.0 }]
for s in shapes {
  if s is Circle {
    println("circle of radius ${s.radius}")
  } else {
    println("some shape of area ${s.area()}")
  }
}
println(shapes[0] == Circle{ radius: 1.0 })
println(Json.stringify(shapes))
```

```text
circle of radius 1.0
some shape of area 9.0
true
[{"radius":1.0},{"side":3.0}]
```

`is` narrows an interface value to a type that satisfies it. In the `else` branch `s` is still a
`Shape`, not a `Square`: anyone may add another satisfying type later, so the checker never knows
the full list.

Containers do not widen in place. An `Array[Circle]` is not an `Array[Shape]`, because code holding
the wider array could push a `Square` among the circles. Copy the elements into a new array instead:

```ant
interface Shape {
  area(s: Self): float64
}

type Circle = { radius: float64 }

func Circle.area(c: Circle): float64 { 3.0 * c.radius * c.radius }

let circles = [Circle{ radius: 1.0 }]
let shapes: Array[Shape] = circles // error: cannot use Array[Circle] where Array[Shape] is expected
let copied: Array[Shape] = [...circles]
```

A value of one interface is assignable to another whose functions it has all of. A `Shape` with
`area` and `name` passes where `HasArea` is expected:

```ant
interface Shape {
  area(s: Self): float64
  name(s: Self): string
}

interface HasArea {
  area(s: Self): float64
}

type Circle = { radius: float64 }

func Circle.area(c: Circle): float64 { 3.0 * c.radius * c.radius }
func Circle.name(c: Circle): string { "circle" }

func report(x: HasArea): string { "area ${x.area()}" }

let s: Shape = Circle{ radius: 1.0 }
println(report(s))
```

```text
area 3.0
```

## Interfaces as bounds

Every interface can bound a type parameter. One that mentions `Self` beyond the receiver can only be
a bound: an array of it could hold a `Job` and a `Ticket`, and `rank` could not compare the two.

```ant
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
fix
```

`best(jobs)` returns a `Job | null`, so `.name` is there to read. Using `Ranked` as a type is a
compile error:

```ant
interface Ranked {
  rank(a: Self, b: Self): int
}

let queue: Array[Ranked] = []
// error: Ranked mentions Self beyond its receiver, so it is a bound only and not a type
```

### Type or bound

For an interface that is a type, both forms work and run the same. They differ in what the checker
accepts and what the caller gets back:

<!-- test: before
interface Shape {
  area(s: Self): float64
}
type Circle = { radius: float64 }
type Square = { side: float64 }
func Circle.area(c: Circle): float64 { 3.0 * c.radius * c.radius }
func Square.area(s: Square): float64 { s.side * s.side }
-->

```ant
func total(xs: Array[Shape]): float64 {
  xs.map({ x.area() }).sum() ?? 0.0
}

func largest[T: Shape](xs: Array[T]): T | null {
  xs.maxBy({ x.area() })
}

let circles = [Circle{ radius: 1.0 }, Circle{ radius: 2.0 }]
println(total([...circles]))
println(largest(circles)?.radius)

let mixed: Array[Shape] = [Circle{ radius: 1.0 }, Square{ side: 3.0 }]
println(total(mixed))
println(largest(mixed))
```

```text
15.0
2.0
12.0
Square{ side: 3.0 }
```

`total` takes the interface: it accepts a mixed array, and an `Array[Circle]` only as a copy.
`largest` takes a bound: it accepts either, and its result keeps the caller's type, so
`largest(circles)` is a `Circle | null` with a `radius` to read. Take the interface, and reach for a
bound when the result must keep the caller's type.

## Interfaces versus unions

A union is a closed set. You list its members in one place, narrow a value with `is`, and the
checker tells you when a chain of tests covers them all:

```ant
type Circle = { radius: float64 }
type Rect = { width: float64, height: float64 }
type Shape = Circle | Rect

func area(s: Shape): float64 {
  if s is Circle {
    3.0 * s.radius * s.radius
  } else if s is Rect {
    s.width * s.height
  }
}

println(area(Rect{ width: 2.0, height: 3.0 }))
```

```text
6.0
```

The chain has no final `else` and still returns a `float64`, because after the second test no member
is left. Add a `Triangle` to `Shape`, and `area` fails to check until it handles it.

An interface is an open set. Each type brings its own functions, a new type joins without touching
existing code, and callers reach the functions with `.`. No chain of tests is ever complete, so
there is no exhaustiveness check.

Pick by who adds the variants:

- **The variants are yours and fixed**, such as the states of a connection or the tokens of a
  parser: a union. Code that handles every case gets checked when a case is added.
- **Other code adds the variants**, such as anything a library writes to or sorts: an interface.
  Each type keeps its behavior next to its data.
- **The behavior is configuration captured at construction**, such as a handler and its options: a
  record of closures. A field that holds a function is data, not a function on the type, so it
  satisfies no interface.

The two do not mix freely. A union holds an interface only beside `null` and `void`, so
`Shape | null` is fine. For a value that is a `Shape` or an error, return a `Result[Shape, Error]`:

```ant
interface Shape {
  area(s: Self): float64
}

func parse(text: string): Shape | Error {
  // error: Shape is an interface, which may share a union only with null and void, not with Error
  panic("unreachable")
}
```

A function declared on a union alias does not make the members satisfy an interface. A `.` call on a
`Circle` looks in `Circle`'s namespace, not in `Figure`'s:

```ant
interface Shape {
  area(s: Self): float64
}

type Circle = { radius: float64 }
type Rect = { width: float64, height: float64 }
type Figure = Circle | Rect

func Figure.area(f: Figure): float64 {
  if f is Circle { 3.0 * f.radius * f.radius } else { f.width * f.height }
}

let f: Figure = Circle{ radius: 1.0 }
let s: Shape = f // error: cannot use Figure where Shape is expected
```

Declare `area` on `Circle` and on `Rect` instead. A union whose every member satisfies the interface
is assignable to it.

## Habits from other languages

- **There is no `implements`.** A type satisfies an interface by having the functions. Declaring
  them on the type is the whole job, and a type from another module, or a built-in one, can satisfy
  an interface its author never saw.
- **An interface is not a class or a namespace.** It has no fields, no default functions, and no
  constructor. `Shape.area` is a compile error; call `area` on a value.
- **`is` tests types, not interfaces.** At run time a value carries its tag, not a list of the
  interfaces it satisfies. Test for a type that satisfies the interface instead.

```ant
interface Shape {
  area(s: Self): float64
}

type Circle = { radius: float64 }

func Circle.area(c: Circle): float64 { 3.0 * c.radius * c.radius }

let s: Shape = Circle{ radius: 1.0 }
let f = Shape.area
// error: Shape is an interface, not a namespace: call area on a value, as x.area(...)
let b = s is Shape // error: is cannot test Shape, an interface: test a type that satisfies it
```

- **Calls through an interface are positional.** Each type may name its parameters differently, so
  `s.scale(by: 2.0)` on a `Shape` is a compile error; write `s.scale(2.0)`.

```ant
interface Shape {
  scale(s: Self, by: float64): Shape
}

type Circle = { radius: float64 }

func Circle.scale(c: Circle, by: float64): Circle { Circle{ radius: c.radius * by } }

let s: Shape = Circle{ radius: 1.0 }
let t = s.scale(by: 2.0)
// error: a call through Shape is positional, since each implementation names its parameters as it likes
let u = s.scale(2.0)
```

- **There is no top type.** An interface has at least one signature, so no interface accepts every
  value. A function that takes any value is generic, as `func show[T](v: T)`.

## See also

- [Tour: Interfaces](../tour/16-interfaces.md)
- [`interface`](../language/keywords/interface.md) and [`Self`](../language/types/Self.md)
- [Unions](unions.md)
- [Generics](generics.md)
- [`is`](../language/operators/is.md)
- [Calls through `.`](../language/operators/dot-calls.md)
