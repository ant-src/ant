<!-- status: reviewed -->

# `interface`

`interface` declares a named set of function signatures. A type satisfies the interface when it has
every function the interface lists; nothing declares that it does.

## Details

- An interface is declared only at the top level of a module. `export interface` exports it.
- Each signature is written like a function type with parameter names. Its first parameter is the
  receiver, of type [`Self`](../types/Self.md): the type that satisfies the interface. A signature
  takes no defaults and no type parameters of its own, and the body lists at least one signature.
- A type satisfies the interface when each call the interface promises, `x.f(…)`, resolves to a
  function on the type whose type, with `Self` as that type, is assignable to the signature. A field
  of function type does not count.
- An interface that mentions `Self` only as the receiver is a type and a bound. One that mentions
  `Self` anywhere else is a bound only.

## Basic examples

Declare an interface, give two types its function, and hold both in one array:

```ant
interface Shape {
  area(s: Self): float64
}

type Circle = { radius: float64 }
type Square = { side: float64 }

func Circle.area(c: Circle): float64 { 3.0 * c.radius * c.radius }
func Square.area(s: Square): float64 { s.side * s.side }

let shapes: Array[Shape] = [Circle{ radius: 1.0 }, Square{ side: 2.0 }]
for s in shapes {
  println(s.area())
}
```

```text
3.0
4.0
```

Take the interface as a parameter. Any satisfying type can be passed:

```ant
interface Named {
  name(n: Self): string
}

type User = { login: string }
type Team = { title: string, size: int }

func User.name(u: User): string { "@${u.login}" }
func Team.name(t: Team): string { "${t.title} (${t.size})" }

func greet(n: Named): string { "Hello, ${n.name()}" }

println(greet(User{ login: "ann" }))
println(greet(Team{ title: "Core", size: 4 }))
```

```text
Hello, @ann
Hello, Core (4)
```

## Scope

### Several signatures and extra parameters

Parameters after the receiver are ordinary parameters:

```ant
interface Scalable {
  scale(s: Self, factor: float64): Self
  describe(s: Self): string
}

type Rect = { width: float64, height: float64 }

func Rect.scale(r: Rect, factor: float64): Rect {
  Rect{ width: r.width * factor, height: r.height * factor }
}
func Rect.describe(r: Rect): string { "${r.width}x${r.height}" }

func doubled[T: Scalable](x: T): string { x.scale(2.0).describe() }

println(doubled(Rect{ width: 1.5, height: 2.0 }))
```

```text
3.0x4.0
```

`scale` returns `Self`, so `Scalable` is a bound only, and `doubled` takes it as `T: Scalable`.

### Built-in types

Built-in types satisfy interfaces by the same rule. `string` and every `Array` have `length`:

```ant
interface Sized {
  length(s: Self): int
}

func longest(items: Array[Sized]): int {
  items.map({ x.length() }).max() ?? 0
}

println(longest(["ant", "wasp", [1, 2, 3, 4, 5]]))
```

```text
5
```

### Generic interfaces

An interface may take type parameters. A call solves them from the argument's functions: `Stack.get`
returns `string | null`, so `lastOf` gets `T` as `string`:

```ant
interface Container[T] {
  get(c: Self, i: int): T | null
  size(c: Self): int
}

type Stack = { items: Array[string] }

func Stack.get(s: Stack, i: int): string | null { s.items.get(i) }
func Stack.size(s: Stack): int { s.items.length() }

func lastOf[T](c: Container[T]): T | null { c.get(c.size() - 1) }

println(lastOf(Stack{ items: ["a", "b", "c"] }))
```

```text
c
```

### Generic types

A generic function on a generic type makes every instantiation satisfy the interface:

```ant
interface Show {
  show(s: Self): string
}

type Box[T] = { value: T }

func Box.show[T](b: Box[T]): string { "Box(${b.value})" }

let items: Array[Show] = [Box{ value: 1 }, Box{ value: "hi" }]
println(items.map({ x.show() }) | String.join(", "))
```

```text
Box(1), Box(hi)
```

### Unions and other interfaces

A union whose every member satisfies the interface is assignable to it, and so is an interface that
has all of its functions:

```ant
interface Named {
  name(n: Self): string
}

interface Labeled {
  name(n: Self): string
  color(n: Self): string
}

type Cat = { nick: string }
type Dog = { nick: string }

func Cat.name(c: Cat): string { c.nick }
func Cat.color(c: Cat): string { "gray" }
func Dog.name(d: Dog): string { d.nick }

let pet: Cat | Dog = Dog{ nick: "Rex" }
let named: Named = pet
println(named.name())

let labeled: Labeled = Cat{ nick: "Tom" }
let plain: Named = labeled
println(plain.name())
```

```text
Rex
Tom
```

### Narrowing with `is`

`is` narrows an interface value to a satisfying type. The `else` branch keeps the interface:

```ant
interface Shape {
  area(s: Self): float64
}

type Circle = { radius: float64 }
type Square = { side: float64 }

func Circle.area(c: Circle): float64 { 3.0 * c.radius * c.radius }
func Square.area(s: Square): float64 { s.side * s.side }

func describe(s: Shape): string {
  if s is Circle { "circle of radius ${s.radius}" } else { "shape of area ${s.area()}" }
}

println(describe(Circle{ radius: 2.0 }))
println(describe(Square{ side: 3.0 }))
```

```text
circle of radius 2.0
shape of area 9.0
```

### Optional interface values

An interface shares a union with `null`, and `?.` calls through it:

```ant
interface Shape {
  area(s: Self): float64
}

type Square = { side: float64 }

func Square.area(s: Square): float64 { s.side * s.side }

let some: Shape | null = Square{ side: 3.0 }
let none: Shape | null = null
println(some?.area())
println(none?.area())
```

```text
9.0
null
```

### Recursive interfaces

A signature may name its own interface. Containers are invariant, so `children` widens the array by
copying it:

```ant
interface Node {
  children(n: Self): Array[Node]
}

type Tree = { kids: Array[Tree] = [] }

func Tree.children(t: Tree): Array[Node] { [...t.kids] }

func count(n: Node): int {
  1 + (n.children().map(count).sum() ?? 0)
}

println(count(Tree{ kids: [Tree{}, Tree{ kids: [Tree{}] }] }))
```

```text
4
```

### Aliases

An alias of an interface stands for it:

```ant
interface Shape {
  area(s: Self): float64
}

type Area = Shape

type Square = { side: float64 }

func Square.area(s: Square): float64 { s.side * s.side }

let a: Area = Square{ side: 2.0 }
println(a.area())
```

```text
4.0
```

### Exported interfaces

`export interface` makes the interface visible to importers:

```ant [geo.ant]
export interface Shape {
  area(s: Self): float64
}

export func total(shapes: Array[Shape]): float64 {
  shapes.map({ x.area() }).sum() ?? 0.0
}
```

```ant
import "./geo.ant" { Shape, total }

type Square = { side: float64 }

func Square.area(s: Square): float64 { s.side * s.side }

let shapes: Array[Shape] = [Square{ side: 1.0 }, Square{ side: 2.0 }]
println(total(shapes))
```

```text
5.0
```

## Applications

Write output to anything with `writeLine`. A `File.Handle` already has it, so standard output and an
in-memory buffer both work:

```ant
interface Sink {
  writeLine(s: Self, line: string): Result[void, Error]
}

type Memory = { lines: Array[string] }

func Memory.writeLine(m: Memory, line: string): Result[void, Error] {
  m.lines.push(line)
  Result.ok(void)
}

func report(out: Sink, names: Array[string]): Result[void, Error] {
  for name in names {
    try out.writeLine("- ${name}")
  }
  Result.ok(void)
}

let memory = Memory{ lines: [] }
let saved = report(memory, ["ant", "bee"])
println(memory.lines)
println(saved)

let printed = report(Os.stdout(), ["wasp"])
println(printed)
```

```text
["- ant", "- bee"]
Ok{ value: void }
- wasp
Ok{ value: void }
```

Sort by an interface that compares two values of one type. It mentions `Self` twice, so it is a
bound:

```ant
interface Compare {
  compare(a: Self, b: Self): int
}

type Job = { name: string, priority: int }

func Job.compare(a: Job, b: Job): int { b.priority - a.priority }

func sorted[T: Compare](xs: Array[T]): Array[T] {
  xs.sortWith(func(a: T, b: T): int { a.compare(b) })
}

let jobs = [Job{ name: "docs", priority: 1 }, Job{ name: "fix", priority: 5 }]
println(sorted(jobs).map({ x.name }))
```

```text
["fix", "docs"]
```

Take a bound instead of the interface when the result must keep the caller's type.
`largest(squares)` is a `Square | null`, so its `side` is there to read:

```ant
interface Shape {
  area(s: Self): float64
}

type Square = { side: float64 }

func Square.area(s: Square): float64 { s.side * s.side }

func largest[T: Shape](xs: Array[T]): T | null {
  xs.maxBy({ x.area() })
}

let squares = [Square{ side: 1.0 }, Square{ side: 3.0 }]
println(largest(squares)?.side)
```

```text
3.0
```

## Properties and relations

An interface value is the value itself, so it compares and prints as the value:

```ant
interface Shape {
  area(s: Self): float64
}

type Square = { side: float64 }

func Square.area(s: Square): float64 { s.side * s.side }

let sq = Square{ side: 2.0 }
let s: Shape = sq
assert(s == sq)
assert(String.from(s) == String.from(sq))
assert(s.area() == sq.area())
```

## Possible issues

A type that lacks one of the functions does not satisfy the interface:

```ant
interface Shape {
  area(s: Self): float64
  name(s: Self): string
}

type Square = { side: float64 }

func Square.area(s: Square): float64 { s.side * s.side }

let s: Shape = Square{ side: 2.0 }
// error: cannot use Square where Shape is expected: Square has no function name of type func(Square): string
```

A function with the right name but another type does not satisfy it either:

```ant
interface Shape {
  area(s: Self): float64
}

type Square = { side: int }

func Square.area(s: Square): int { s.side * s.side }

let s: Shape = Square{ side: 2 }
// error: cannot use Square where Shape is expected: Square.area has type func(Square): int, not func(Square): float64
```

A field of function type is not a function on the type:

```ant
interface Shape {
  area(s: Self): float64
}

type Fixed = { area: func(): float64 }

let s: Shape = Fixed{ area: func(): float64 { 1.0 } }
// error: cannot use Fixed where Shape is expected: Fixed has a field area, not a function declared on it
```

Containers are invariant: an `Array[Square]` is not an `Array[Shape]`. Spread it into a new array to
widen it:

```ant
interface Shape {
  area(s: Self): float64
}

type Square = { side: float64 }

func Square.area(s: Square): float64 { s.side * s.side }

let squares = [Square{ side: 1.0 }]
let shapes: Array[Shape] = squares // error: cannot use Array[Square] where Array[Shape] is expected
let copy: Array[Shape] = [...squares]
```

The same invariance applies to return types. A `children` that returns the field as it is does not
satisfy a recursive interface:

```ant
interface Node {
  children(n: Self): Array[Node]
}

type Tree = { kids: Array[Tree] = [] }

func Tree.children(t: Tree): Array[Tree] { t.kids }

let n: Node = Tree{}
// error: cannot use Tree where Node is expected: Tree.children has type func(Tree): Array[Tree], not func(Tree): Array[Node]
```

An interface that mentions `Self` beyond its receiver is not a type:

```ant
interface Compare {
  compare(a: Self, b: Self): int
}

let queue: Array[Compare] = []
// error: Compare mentions Self beyond its receiver, so it is a bound only and not a type
```

An interface value has only the interface's functions. Narrow it with `is` to read a field:

```ant
interface Shape {
  area(s: Self): float64
}

type Circle = { radius: float64 }

func Circle.area(c: Circle): float64 { 3.0 * c.radius * c.radius }

let s: Shape = Circle{ radius: 1.0 }
println(s.radius) // error: Shape is not a record, so it has no field radius
```

An interface is not a namespace, and `is` cannot test it:

```ant
interface Shape {
  area(s: Self): float64
}

type Circle = { radius: float64 }

func Circle.area(c: Circle): float64 { 3.0 * c.radius * c.radius }

let s: Shape = Circle{ radius: 1.0 }
println(Shape.area(s))
// error: Shape is an interface, not a namespace: call area on a value, as x.area(...)
println(s is Shape) // error: is cannot test Shape, an interface: test a type that satisfies it
```

An interface shares a union only with `null` and `void`:

```ant
interface Shape {
  area(s: Self): float64
}

let x: Shape | int = 1
// error: Shape is an interface, which may share a union only with null and void, not with int
```

An interface has no run-time identity, so `type(…)` cannot reach it:

```ant
interface Shape {
  area(s: Self): float64
}

println(type(Shape))
// error: type(...) cannot reach an interface: Shape is one, and an interface has no run-time identity
```

The first parameter of every signature must have the type `Self`, and a parameter takes no default:

```ant
interface Shape {
  area(x: float64): float64 // error: the first parameter of area must have the type Self, the satisfying type
}
```

```ant
interface Shape {
  area(s: Self, scale: float64 = 1.0): float64
  // error: the parameter scale of area takes no default; an interface lists signatures only
}
```

A signature takes no type parameters of its own. Put them on the interface, as `Container[T]` does:

```ant
interface Mapper {
  apply[U](m: Self, x: U): U // error: apply takes no type parameters of its own; an interface declares them on the interface
}
```

An interface lists at least one signature:

```ant
interface Empty {} // error: the interface Empty declares no function; an interface lists at least one
```

An interface is declared only at the top level:

```ant
func setup() {
  interface Shape { // error: 'interface' is only allowed at the top level
    area(s: Self): float64
  }
}
```

A function that is not exported makes its type satisfy the interface only inside its module:

```ant [shapes.ant]
export interface Shape {
  area(s: Self): float64
}

export type Square = { side: float64 }

func Square.area(s: Square): float64 { s.side * s.side }
```

```ant
import "./shapes.ant" { Shape, Square }

let s: Shape = Square{ side: 2.0 }
// error: cannot use Square where Shape is expected: Square.area is not exported
```

## See also

[Self](../types/Self.md) · [type](type.md) · [func](func.md) · [export](export.md) ·
[is](../operators/is.md) · [calls through `.`](../operators/dot-calls.md) ·
[Array.sortWith](../../lib/Array/sortWith.md) · [Interfaces guide](../../guide/interfaces.md) ·
[Generics guide](../../guide/generics.md) · [Tour: Interfaces](../../tour/16-interfaces.md)
